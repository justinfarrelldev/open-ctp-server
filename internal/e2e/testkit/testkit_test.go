package testkit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScenarioProvidesIsolatedDeterministicSources(t *testing.T) {
	t.Parallel()

	first := NewScenario("same scenario", 42)
	second := NewScenario("same scenario", 42)
	other := NewScenario("other scenario", 42)
	if first.Namespace == other.Namespace {
		t.Fatalf("namespaces must be disjoint: %q", first.Namespace)
	}
	if first.Clock.Now() != second.Clock.Now() || first.Clock.Now() != time.Unix(0, 0).UTC() {
		t.Fatalf("fake clocks must begin at the deterministic epoch")
	}
	first.Clock.Advance(time.Second)
	if second.Clock.Now() != time.Unix(0, 0).UTC() {
		t.Fatal("scenarios must not share clocks")
	}
	if first.IDs.Next() != second.IDs.Next() || first.RNG.Int63() != second.RNG.Int63() {
		t.Fatal("same scenario seed must reproduce IDs and RNG output")
	}
	if _, ok := first.Context.Deadline(); !ok {
		t.Fatal("scenario requires a deadline")
	} else if remaining := time.Until(first.Deadline()); remaining > ScenarioBudget || remaining < ScenarioBudget-time.Second {
		t.Fatalf("scenario deadline must be bounded to %s, got %s", ScenarioBudget, remaining)
	}
}

func TestConcurrentScenariosUseDisjointNamespacesEventsAndClocks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	type result struct {
		namespace string
		evidence  FailureEvidence
		clock     time.Time
	}
	results := make(chan result, 8)
	for index := 0; index < 8; index++ {
		index := index
		go func() {
			scenario := NewScenario(fmt.Sprintf("concurrent-%d", index), int64(index))
			defer scenario.Close()
			scenario.Clock.Advance(time.Duration(index) * time.Second)
			evidence, err := scenario.CaptureFailure(root, index, []any{index}, []any{scenario.ID}, []any{scenario.Namespace})
			if err != nil {
				t.Errorf("capture %d: %v", index, err)
				return
			}
			results <- result{namespace: scenario.Namespace, evidence: evidence, clock: scenario.Clock.Now()}
		}()
	}
	seenNamespaces := make(map[string]struct{})
	seenEvents := make(map[string]struct{})
	seenClocks := make(map[time.Time]struct{})
	for range 8 {
		result := <-results
		if _, exists := seenNamespaces[result.namespace]; exists {
			t.Fatalf("namespace reused: %q", result.namespace)
		}
		if _, exists := seenEvents[result.evidence.EventHash]; exists {
			t.Fatalf("event evidence reused: %q", result.evidence.EventHash)
		}
		if _, exists := seenClocks[result.clock]; exists {
			t.Fatalf("clock reused: %s", result.clock)
		}
		seenNamespaces[result.namespace] = struct{}{}
		seenEvents[result.evidence.EventHash] = struct{}{}
		seenClocks[result.clock] = struct{}{}
	}
}

func TestBarrierCoordinatesWithoutSleep(t *testing.T) {
	t.Parallel()
	barrier := NewBarrier(2)
	context, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	arrived := make(chan struct{})
	firstResult := make(chan error, 1)
	go func() {
		close(arrived)
		firstResult <- barrier.Wait(context)
	}()
	<-arrived
	if err := barrier.Wait(context); err != nil {
		t.Fatalf("second barrier wait: %v", err)
	}
	if err := <-firstResult; err != nil {
		t.Fatalf("first barrier wait: %v", err)
	}
}

func TestFailureEvidenceIncludesSafeDeterministicPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	scenario := NewScenario("failure fixture", 99)
	evidence, err := scenario.CaptureFailure(root, 3, []any{"command"}, []any{"event"}, []any{"projection"})
	if err != nil {
		t.Fatalf("capture failure: %v", err)
	}
	if evidence.ScenarioID != "failure fixture" || evidence.Seed != 99 || evidence.CommandIndex != 3 {
		t.Fatalf("evidence identity = %#v", evidence)
	}
	for _, path := range []string{evidence.TranscriptPath, evidence.EventPath, evidence.ProjectionPath} {
		relative, err := filepath.Rel(root, path)
		if err != nil || strings.HasPrefix(relative, "..") || filepath.Base(path) == "." {
			t.Fatalf("unsafe evidence path %q", path)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing evidence %q: %v", path, err)
		}
	}
}

func TestFaultPlanFailsAtConfiguredCommand(t *testing.T) {
	t.Parallel()
	plan := FaultPlan{AtCommand: 2, Err: errors.New("injected failure")}
	if err := plan.Check(1); err != nil {
		t.Fatalf("unexpected earlier fault: %v", err)
	}
	if got := plan.Check(2); !errors.Is(got, plan.Err) {
		t.Fatalf("fault = %v, want injected error", got)
	}
}

func TestTimeoutPolicyRejectsMissingInnerOrOuterLimits(t *testing.T) {
	t.Parallel()
	valid := "GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./internal/gameplay/..."
	if err := ValidateTestCommand(valid); err != nil {
		t.Fatalf("valid command rejected: %v", err)
	}
	for _, command := range []string{
		"go test -timeout=8m -p=2 ./...",
		"timeout --signal=TERM --kill-after=30s 10m go test -p=2 ./...",
		"GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m ./...",
	} {
		if err := ValidateTestCommand(command); err == nil {
			t.Fatalf("invalid command accepted: %q", command)
		}
	}
}

func TestCoverageThresholdValidation(t *testing.T) {
	t.Parallel()
	if err := ValidateCoverage("internal/gameplay/engine", 90, 90); err != nil {
		t.Fatalf("threshold boundary rejected: %v", err)
	}
	if err := ValidateCoverage("internal/gameplay/engine", 89.9, 90); err == nil {
		t.Fatal("below-threshold coverage accepted")
	}
}

func TestRepositoryTimeoutPolicy(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	policy := filepath.Join(directory, "policy.yml")
	valid := "GOMAXPROCS=4 timeout --signal=TERM --kill-after=30s 10m go test -timeout=8m -p=2 ./...\n"
	if err := os.WriteFile(policy, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRepositoryTimeoutPolicy(policy); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(policy, []byte("go test ./...\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRepositoryTimeoutPolicy(policy); err == nil {
		t.Fatal("repository timeout policy accepted an unbounded test command")
	}
}
