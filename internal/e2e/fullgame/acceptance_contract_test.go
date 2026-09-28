package fullgame

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
)

type recordingAcceptancePolicy struct {
	name   string
	called *string
}

func (p recordingAcceptancePolicy) Name() string { return p.name }

func (recordingAcceptancePolicy) validateRestart(AcceptanceRestartPlan, AcceptanceBudgets) error {
	return nil
}
func (p recordingAcceptancePolicy) execute(_ context.Context, _ AcceptanceConfig, _ *testkit.Scenario) (Artifacts, error) {
	*p.called = p.name
	return Artifacts{}, nil
}

type blockingAcceptancePolicy struct{}

func (blockingAcceptancePolicy) Name() string { return "blocking-test-policy" }

func (blockingAcceptancePolicy) validateRestart(AcceptanceRestartPlan, AcceptanceBudgets) error {
	return nil
}
func (blockingAcceptancePolicy) execute(ctx context.Context, _ AcceptanceConfig, _ *testkit.Scenario) (Artifacts, error) {
	<-ctx.Done()
	return Artifacts{}, ctx.Err()
}
func TestTaskFiveExecutionSemanticsRemainGreen(t *testing.T) {
	config := RowOneConquestAcceptance()
	restartEvents := make([]AcceptanceRestartEvent, 0, 1)
	config.ObserveRestart = func(event AcceptanceRestartEvent) {
		restartEvents = append(restartEvents, event)
	}
	execute := func(namespace string, restart AcceptanceRestartPlan) Artifacts {
		run := config
		run.Restart = restart
		kit := testkit.NewScenario(namespace, int64(run.Seed))
		defer kit.Close()
		artifacts, err := (PublicExecutor{}).Execute(kit.Context, run, kit)
		if err != nil {
			t.Fatal(err)
		}
		return artifacts
	}
	disabled := AcceptanceRestartPlan{Point: AcceptanceRestartDisabled}
	first := execute("TestPublicExecutorCompletesMinimalScenarioThroughProductionRuntime-clean-first", disabled)
	second := execute("TestPublicExecutorCompletesMinimalScenarioThroughProductionRuntime-clean-second", disabled)
	restarted := execute("TestPublicExecutorCompletesMinimalScenarioThroughProductionRuntime-restart", config.Restart)
	if first != second || first != restarted {
		t.Fatalf("Task 5 execution artifacts differ: first=%+v second=%+v restarted=%+v", first, second, restarted)
	}
	if len(restartEvents) != 1 || restartEvents[0] != (AcceptanceRestartEvent{Point: AcceptanceRestartAfterDurableRound}) {
		t.Fatalf("conquest restart events = %+v, want one durable-round restart", restartEvents)
	}
}

func TestPublicExecutorDispatchesOnlyExplicitHumanPolicy(t *testing.T) {
	called := ""
	config := RowOneConquestAcceptance()
	config.HumanPolicy = recordingAcceptancePolicy{name: "explicit-policy", called: &called}
	kit := testkit.NewScenario(t.Name(), int64(config.Seed))
	defer kit.Close()
	if _, err := (PublicExecutor{}).Execute(context.Background(), config, kit); err != nil {
		t.Fatal(err)
	}
	if called != "explicit-policy" {
		t.Fatalf("explicit policy call = %q", called)
	}
}
func TestAcceptanceConfigRejectsMissingExplicitPoliciesAndBudgets(t *testing.T) {
	tests := map[string]func(*AcceptanceConfig){"human policy": func(c *AcceptanceConfig) { c.HumanPolicy = nil }, "autonomous AI": func(c *AcceptanceConfig) { c.AutonomousAIPolicy = "" }, "command budget": func(c *AcceptanceConfig) { c.Budgets.Commands = 0 }, "turn budget": func(c *AcceptanceConfig) { c.Budgets.Turns = 0 }, "time budget": func(c *AcceptanceConfig) { c.Budgets.Deadline = 0 }}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			config := RowOneConquestAcceptance()
			mutate(&config)
			if err := config.validate(); err == nil {
				t.Fatal("invalid acceptance config was accepted")
			}
		})
	}
}

func TestAcceptanceConfigValidatesTypedRestartPlanByPolicy(t *testing.T) {
	validConquest := RowOneConquestAcceptance()
	validDiplomacy := RowTwoDiplomacyAcceptance()
	for name, config := range map[string]AcceptanceConfig{
		"conquest durable round": validConquest,
		"conquest disabled": func() AcceptanceConfig {
			config := validConquest
			config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartDisabled}
			return config
		}(),
		"diplomacy completed turn": validDiplomacy,
		"diplomacy disabled": func() AcceptanceConfig {
			config := validDiplomacy
			config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartDisabled}
			return config
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			if err := config.validate(); err != nil {
				t.Fatal(err)
			}
		})
	}

	for name, config := range map[string]AcceptanceConfig{
		"conquest after turn": func() AcceptanceConfig {
			config := validConquest
			config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartAfterCompletedTurn, CompletedTurn: 12}
			return config
		}(),
		"conquest durable round with turn": func() AcceptanceConfig {
			config := validConquest
			config.Restart.CompletedTurn = 1
			return config
		}(),
		"diplomacy durable round": func() AcceptanceConfig {
			config := validDiplomacy
			config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartAfterDurableRound}
			return config
		}(),
		"diplomacy zero turn": func() AcceptanceConfig {
			config := validDiplomacy
			config.Restart.CompletedTurn = 0
			return config
		}(),
		"diplomacy turn at budget": func() AcceptanceConfig {
			config := validDiplomacy
			config.Restart.CompletedTurn = config.Budgets.Turns
			return config
		}(),
		"unknown point": func() AcceptanceConfig {
			config := validConquest
			config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartPoint("unknown")}
			return config
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			if err := config.validate(); err == nil {
				t.Fatal("invalid typed restart plan was accepted")
			}
		})
	}
}
func TestPublicExecutorTimeBudgetIsBoundedAndDiagnostic(t *testing.T) {
	config := RowOneConquestAcceptance()
	config.HumanPolicy = blockingAcceptancePolicy{}
	config.Budgets.Deadline = 10 * time.Millisecond
	kit := testkit.NewScenario(t.Name(), int64(config.Seed))
	defer kit.Close()
	started := time.Now()
	_, err := (PublicExecutor{}).Execute(context.Background(), config, kit)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "time budget exhausted") {
		t.Fatalf("bounded deadline error = %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("bounded deadline took %s", elapsed)
	}
}

func TestPublicExecutorCommandAndTurnBudgetsFailWithPublicDiagnostics(t *testing.T) {
	for name, mutate := range map[string]func(*AcceptanceConfig){
		"command": func(config *AcceptanceConfig) { config.Budgets.Commands = 1 },
		"turn":    func(config *AcceptanceConfig) { config.Budgets.Turns = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			config := RowOneConquestAcceptance()
			config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartDisabled}
			mutate(&config)
			kit := testkit.NewScenario(t.Name(), int64(config.Seed))
			defer kit.Close()
			started := time.Now()
			_, err := (PublicExecutor{}).Execute(context.Background(), config, kit)
			if err == nil || !strings.Contains(err.Error(), "budget") {
				t.Fatalf("bounded %s budget error = %v", name, err)
			}
			if strings.Contains(strings.ToLower(err.Error()), "canonical") {
				t.Fatalf("%s budget diagnostic exposed canonical state: %v", name, err)
			}
			if elapsed := time.Since(started); elapsed > 30*time.Second {
				t.Fatalf("%s budget failure took %s", name, elapsed)
			}
		})
	}
}
func TestAcceptanceFixturesAreCatalogIndependentAndPinned(t *testing.T) {
	one, two := RowOneConquestAcceptance(), RowTwoDiplomacyAcceptance()
	if one.Name == two.Name || one.HumanPolicy.Name() == two.HumanPolicy.Name() {
		t.Fatal("acceptance fixtures are not independently pinned")
	}
	if one.ExpectedStateHash != "21407d478695aa8d92f297d6dec402697ad30baa7fac0d6ce7c605abb7448490" || one.ExpectedEventHash != "9d01453176673740a3658ff4711d131155df42e94bf1c58325088730e9e114db" {
		t.Fatal("row one accepted artifacts changed")
	}
	for _, config := range []AcceptanceConfig{one, two} {
		if err := config.validate(); err != nil {
			t.Fatalf("%s: %v", config.Name, err)
		}
	}
}

func TestAcceptanceRulesFixtureDrivesPublicSetupAndProductionPin(t *testing.T) {
	config := RowOneConquestAcceptance()
	if got, want := acceptanceRulesJSON(config.Rules), `["victory:bloodlust_conquest"]`; got != want {
		t.Fatalf("public rules setup = %s, want %s", got, want)
	}
	for name, mutate := range map[string]func(*AcceptanceRulesFixture){
		"hash":    func(r *AcceptanceRulesFixture) { r.Hash = strings.Repeat("0", 64) },
		"victory": func(r *AcceptanceRulesFixture) { r.Victory = "diplomatic_world_union" },
	} {
		t.Run(name, func(t *testing.T) {
			invalid := config
			mutate(&invalid.Rules)
			if err := invalid.validate(); err == nil {
				t.Fatal("invalid operational rules fixture was accepted")
			}
		})
	}
}

func TestAITranscriptRequiresProductionGatewayProvenance(t *testing.T) {
	entry := aiTranscriptEntry{GameID: 41, ActorPlayerID: "account-2", IdempotencyKey: "ai-1", CommandID: "ai-1", Type: "pass", Payload: json.RawMessage(`{}`)}
	if err := validateAITranscript([]aiTranscriptEntry{entry}, 41, []string{"account-2"}, 1); err != nil {
		t.Fatal(err)
	}
	if err := validateAITranscript(nil, 41, []string{"account-2"}, 1); err == nil {
		t.Fatal("missing AI observer entry was accepted")
	}
	if err := validateAITranscript([]aiTranscriptEntry{entry, entry}, 41, []string{"account-2"}, 1); err == nil {
		t.Fatal("extra unmatched AI observer entry was accepted")
	}
	invalid := entry
	invalid.ActorPlayerID = "account-1"
	if err := validateAITranscript([]aiTranscriptEntry{invalid}, 41, []string{"account-2"}, 1); err == nil {
		t.Fatal("human actor was accepted as production AI provenance")
	}
	invalid = entry
	invalid.IdempotencyKey = ""
	if err := validateAITranscript([]aiTranscriptEntry{invalid}, 41, []string{"account-2"}, 1); err == nil {
		t.Fatal("AI transcript without gateway idempotency identity was accepted")
	}
}

func TestAIObservationRequiresExactDriverCommandCount(t *testing.T) {
	if err := validateAIObservation(1, 1); err != nil {
		t.Fatal(err)
	}
	if err := validateAIObservation(1, 0); err == nil {
		t.Fatal("missing AI observer submission was accepted")
	}
	if err := validateAIObservation(1, 2); err == nil {
		t.Fatal("extra AI observer submission was accepted")
	}
	if err := validateAIObservation(0, 0); err != nil {
		t.Fatalf("zero-command AI drive should be accepted when no observer submission occurs: %v", err)
	}
}
