package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRunRequiresExplicitDataDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run(nil, &stdout, &stderr); got != 2 {
		t.Fatalf("run() exit code = %d, want 2", got)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "--data-dir is required") {
		t.Fatalf("stdout=%q stderr=%q, want required flag error only", stdout.String(), stderr.String())
	}
}

func TestRunRejectsUnknownDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--data-dir", filepath.Join(t.TempDir(), "missing")}, &stdout, &stderr); got != 1 {
		t.Fatalf("run() exit code = %d, want 1", got)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "invalid_directory") {
		t.Fatalf("stdout=%q stderr=%q, want source-located importer error", stdout.String(), stderr.String())
	}
}

func TestRunEmitsDeterministicRulesetIdentity(t *testing.T) {
	dataDir, err := filepath.Abs(filepath.Join("..", "..", "internal", "gameplay", "rules", "testdata", "synthetic-ctp2"))
	if err != nil {
		t.Fatalf("Abs() error = %v", err)
	}
	first := execute(t, dataDir)
	second := execute(t, dataDir)
	if first != second {
		t.Fatalf("repeat output differs:\nfirst=%q\nsecond=%q", first, second)
	}
	var identity struct {
		CompatibilityVersion string `json:"compatibility_version"`
		ContentHash          string `json:"content_hash"`
	}
	if err := json.Unmarshal([]byte(first), &identity); err != nil {
		t.Fatalf("JSON output error = %v output=%q", err, first)
	}
	if identity.CompatibilityVersion == "" || len(identity.ContentHash) != 64 || strings.Contains(first, "UNIT_SYNTHETIC") {
		t.Fatalf("identity = %+v output=%q, want version/hash only", identity, first)
	}
}

func TestRunEmitsRetailProjectedRuntimeIdentity(t *testing.T) {
	dataDir := localRetailGameDataRoot(t)
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--retail", "--data-dir", dataDir}, &stdout, &stderr); got != 0 {
		t.Fatalf("run() exit code = %d stderr=%q", got, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
	var identity struct {
		CompatibilityVersion string `json:"compatibility_version"`
		ContentHash          string `json:"content_hash"`
		RulesetType          string `json:"ruleset_type"`
		RuntimeCounts        struct {
			Units      int `json:"units"`
			Advances   int `json:"advances"`
			Production int `json:"production"`
		} `json:"runtime_counts"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &identity); err != nil {
		t.Fatalf("JSON output error = %v output=%q", err, stdout.String())
	}
	if identity.RulesetType != "retail-runtime-projection" || len(identity.ContentHash) != 64 || identity.RuntimeCounts.Units != 74 || identity.RuntimeCounts.Advances != 107 || identity.RuntimeCounts.Production != 156 {
		t.Fatalf("identity = %+v output=%q", identity, stdout.String())
	}
}

func TestRunEmitsRetailDefinitionsThatCompileWithIdentityHash(t *testing.T) {
	dataDir := localRetailGameDataRoot(t)
	identityJSON := executeArgs(t, "--retail", "--data-dir", dataDir)
	var identity struct {
		ContentHash string `json:"content_hash"`
	}
	if err := json.Unmarshal([]byte(identityJSON), &identity); err != nil {
		t.Fatalf("identity JSON error = %v output=%q", err, identityJSON)
	}
	definitions := executeArgs(t, "--retail", "--data-dir", dataDir, "--format", "definitions")
	runtime, err := rules.CompileRuntime([]byte(definitions), identity.ContentHash)
	if err != nil {
		t.Fatalf("CompileRuntime(retail definitions) error = %v", err)
	}
	if _, ok := runtime.Unit("UNIT_SETTLER"); !ok || len(runtime.Units) != 74 {
		t.Fatalf("compiled runtime = %+v", runtime)
	}
}

func TestRunRejectsUnknownFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--data-dir", filepath.Join(t.TempDir(), "missing"), "--format", "xml"}, &stdout, &stderr); got != 2 {
		t.Fatalf("run() exit code = %d, want 2", got)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "--format must be identity or definitions") {
		t.Fatalf("stdout=%q stderr=%q, want format error", stdout.String(), stderr.String())
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := run([]string{"--installation-dir", "anywhere"}, &stdout, &stderr); got != 2 {
		t.Fatalf("run() exit code = %d, want 2", got)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("stdout=%q stderr=%q, want unknown flag error", stdout.String(), stderr.String())
	}
}

func execute(t *testing.T, dataDir string) string {
	t.Helper()
	return executeArgs(t, "--data-dir", dataDir)
}

func executeArgs(t *testing.T, args ...string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if got := run(args, &stdout, &stderr); got != 0 {
		t.Fatalf("run() exit code = %d stderr=%q", got, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
	return stdout.String()
}

func localRetailGameDataRoot(t *testing.T) string {
	t.Helper()
	candidates := []string{
		filepath.Join("..", "..", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"),
		filepath.Join("..", "..", "open-ctp-server", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"),
		"/home/justi/Projects/open-ctp-server/ctp2/ctp2_released_game/ctp2_data/default/gamedata",
	}
	for _, candidate := range candidates {
		if stat, err := os.Stat(filepath.Join(candidate, "Units.txt")); err == nil && !stat.IsDir() {
			return candidate
		}
	}
	t.Skip("local CTP2 released-game data is not present")
	return ""
}
