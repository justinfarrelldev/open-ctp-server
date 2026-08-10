package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestCompileRuntimeBuildsClosedDeterministicModel(t *testing.T) {
	t.Parallel()

	ruleset, err := Import(ImportConfig{Directory: fixtureDirectory(t)})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	runtime, err := CompileRuntime(ruleset.CanonicalJSON, ruleset.ContentHash)
	if err != nil {
		t.Fatalf("CompileRuntime() error = %v", err)
	}
	if got, want := strings.Join(runtime.VictoryEvaluatorIDs(), ","), "bloodlust_conquest,diplomatic_world_union,science_gaia_controller,score_2300"; got != want {
		t.Fatalf("victory evaluators = %q, want %q", got, want)
	}
	if runtime.Calendar.FinalTurn != 48 {
		t.Fatalf("calendar final turn = %d, want 48", runtime.Calendar.FinalTurn)
	}
	if runtime.Gaia.ActivationTurns != 1 {
		t.Fatalf("Gaia activation turns = %d, want 1", runtime.Gaia.ActivationTurns)
	}
	if got, want := runtime.ProductionIDs(), "IMPROVE_SYNTHETIC,UNIT_SYNTHETIC,WONDER_SYNTHETIC"; got != want {
		t.Fatalf("production IDs = %q, want %q", got, want)
	}
	if got, want := runtime.EffectIDs(), "IMPROVE_SYNTHETIC,WONDER_SYNTHETIC"; got != want {
		t.Fatalf("effect IDs = %q, want %q", got, want)
	}
}

func TestServerOwnedSyntheticRulesetSelectsConfiguredVictoryEvaluators(t *testing.T) {
	ruleset, err := ServerOwnedSyntheticRuleset("score_2300", "bloodlust_conquest")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := CompileRuntime(ruleset.CanonicalJSON, ruleset.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(runtime.VictoryEvaluatorIDs(), ","), "bloodlust_conquest,score_2300"; got != want {
		t.Fatalf("selected victory evaluators = %q, want %q", got, want)
	}
	if _, err := ServerOwnedSyntheticRuleset("unknown"); err == nil {
		t.Fatal("unsupported victory evaluator was accepted")
	}
}

func TestCompileRuntimeRejectsInvalidOrTamperedDefinitions(t *testing.T) {
	t.Parallel()

	ruleset, err := Import(ImportConfig{Directory: fixtureDirectory(t)})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	unknownKind := []byte(strings.Replace(string(ruleset.CanonicalJSON), `"kind":"unit"`, `"kind":"unsupported"`, 1))
	unknownKindHash := sha256.Sum256(unknownKind)
	cases := []struct {
		name       string
		definitions []byte
		hash       string
		want       string
	}{
		{name: "hash", definitions: ruleset.CanonicalJSON, hash: strings.Repeat("0", 64), want: "rules content hash mismatch"},
		{name: "malformed", definitions: []byte("{"), hash: ruleset.ContentHash, want: "rules definitions are malformed"},
		{name: "unknown record kind", definitions: unknownKind, hash: hex.EncodeToString(unknownKindHash[:]), want: "unknown kind"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := CompileRuntime(tc.definitions, tc.hash)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("CompileRuntime() error = %v, want %q", err, tc.want)
			}
		})
	}
}
