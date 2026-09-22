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
	if len(runtime.Advances) != 2 || runtime.Advances[0].ID != "ADVANCE_SYNTHETIC_ROOT" || runtime.Advances[1].Prerequisites[0] != "ADVANCE_SYNTHETIC_ROOT" {
		t.Fatalf("advances = %#v, want deterministic synthetic prerequisite tree", runtime.Advances)
	}
	if got, want := runtime.ProductionIDs(), "IMPROVE_SYNTHETIC,UNIT_SYNTHETIC,WONDER_SYNTHETIC"; got != want {
		t.Fatalf("production IDs = %q, want %q", got, want)
	}
	if got, want := runtime.EffectIDs(), "IMPROVE_SYNTHETIC,WONDER_SYNTHETIC"; got != want {
		t.Fatalf("effect IDs = %q, want %q", got, want)
	}
}

func TestCompileRuntimePreservesSyntheticUnitCapabilities(t *testing.T) {
	t.Parallel()

	ruleset, err := ServerOwnedSyntheticRuleset("score_2300")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := CompileRuntime(ruleset.CanonicalJSON, ruleset.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	unit, ok := runtime.Unit("UNIT_SYNTHETIC")
	if !ok {
		t.Fatal("starter unit definition missing")
	}
	if unit.Domain != "land" || unit.Movement != 2 || unit.Attack != 4 || unit.Defense != 4 || unit.Strength != 4 || unit.Armor != 2 || unit.Health != 10 || !unit.ZOC || unit.VisibilityRange != 2 || len(unit.Effects) != 6 {
		t.Fatalf("unit capabilities = %+v, want starter land combat profile", unit)
	}
	if runtime.Starter.Domain != unit.Domain || runtime.Starter.Movement != unit.Movement || runtime.Starter.Attack != unit.Attack || runtime.Starter.Defense != unit.Defense || runtime.Starter.Strength != unit.Strength || runtime.Starter.Armor != unit.Armor || runtime.Starter.Health != unit.Health || runtime.Starter.ZOC != unit.ZOC || runtime.Starter.VisibilityRange != unit.VisibilityRange || len(runtime.Starter.Effects) != len(unit.Effects) {
		t.Fatalf("starter = %+v, want derived from %+v", runtime.Starter, unit)
	}
}

func TestCompileRuntimeRejectsMalformedUnitCapabilities(t *testing.T) {
	t.Parallel()

	ruleset, err := ServerOwnedSyntheticRuleset("score_2300")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{name: "domain", old: `"Domain":"land"`, new: `"Domain":"orbital"`, want: "invalid Domain"},
		{name: "cargo without capacity", old: `"ZOC":"true"`, new: `"CargoDomains":"land","ZOC":"true"`, want: "CargoDomains without TransportCapacity"},
		{name: "duplicate prerequisites", old: `"Role":"starter"`, new: `"Prerequisites":"ADVANCE_A,ADVANCE_A","Role":"starter"`, want: "duplicate Prerequisites"},
		{name: "invalid effect", old: `"Effects":"settler,special_embassy,special_steal:100,special_capture:0,special_terraform:1,special_terraform_turns:1"`, new: `"Effects":"teleport"`, want: "invalid Effects"},
		{name: "missing advance prerequisite", old: `"PrerequisiteAdvanceIDs":"ADVANCE_SYNTHETIC_ROOT"`, new: `"PrerequisiteAdvanceIDs":"ADVANCE_SYNTHETIC_MISSING"`, want: "missing prerequisite"},
		{name: "cyclic advance prerequisite", old: `"PrerequisiteAdvanceIDs":"ADVANCE_SYNTHETIC_ROOT"`, new: `"PrerequisiteAdvanceIDs":"ADVANCE_SYNTHETIC_UNLOCK"`, want: "cyclic prerequisites"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			definitions := []byte(strings.Replace(string(ruleset.CanonicalJSON), tc.old, tc.new, 1))
			canonical, err := CanonicalDefinitions(definitions)
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(canonical)
			_, err = CompileRuntime(canonical, hex.EncodeToString(digest[:]))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("CompileRuntime() error = %v, want %q", err, tc.want)
			}
		})
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
		name        string
		definitions []byte
		hash        string
		want        string
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
