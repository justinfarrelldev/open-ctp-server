package rules

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestParseCTP2StyleSyntheticCatalog(t *testing.T) {
	t.Parallel()

	ruleset, err := Import(ImportConfig{Directory: fixtureDirectory(t)})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if got, want := recordIDs(ruleset.Records), "ADVANCE_SYNTHETIC_ROOT,ADVANCE_SYNTHETIC_UNLOCK,ENDGAME_GAIA_VECTOR,ENDGAME_UNION_VECTOR,GOVERNMENT_SYNTHETIC_REPUBLIC,IMPROVE_SYNTHETIC,POLLUTION_SOURCE_SYNTHETIC_FACTORY,TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP,TERRAIN_IMPROVEMENT_SYNTHETIC_FARM,TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD,TERRAIN_SYNTHETIC,UNIT_SYNTHETIC,WONDER_SYNTHETIC"; got != want {
		t.Fatalf("record IDs = %q, want %q", got, want)
	}
	if ruleset.ContentHash == "" || len(ruleset.CanonicalJSON) == 0 {
		t.Fatal("expected canonical catalog bytes and content hash")
	}
}

func TestPlayableCivilizationMetadataFailsClosed(t *testing.T) {
	playable := "CIVILIZATION_TEST_01 { Playable true Special false Barbarian false }"
	ruleset, err := Import(ImportConfig{Directory: writeCatalog(t, map[string]string{"Civilisation.ctp2synthetic": playable})})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	ids, err := ruleset.PlayableCivilizationIDs()
	if err != nil || strings.Join(ids, ",") != "CIVILIZATION_TEST_01" {
		t.Fatalf("PlayableCivilizationIDs() = %v, %v", ids, err)
	}

	for name, record := range map[string]string{
		"special":      "CIVILIZATION_TEST_02 { Playable true Special true Barbarian false }",
		"barbarian":    "CIVILIZATION_TEST_03 { Playable true Special false Barbarian true }",
		"not playable": "CIVILIZATION_TEST_04 { Playable false Special false Barbarian false }",
	} {
		t.Run(name, func(t *testing.T) {
			imported, err := Import(ImportConfig{Directory: writeCatalog(t, map[string]string{"Civilisation.ctp2synthetic": record})})
			if err != nil {
				t.Fatalf("Import() error = %v", err)
			}
			if _, err := imported.PlayableCivilizationIDs(); err == nil {
				t.Fatal("unselectable civilization was accepted")
			}
		})
	}

	for name, record := range map[string]string{
		"unknown field":  "CIVILIZATION_TEST_05 { Playable true Special false Barbarian false Hidden true }",
		"malformed bool": "CIVILIZATION_TEST_06 { Playable yes Special false Barbarian false }",
		"missing field":  "CIVILIZATION_TEST_07 { Playable true Special false }",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Import(ImportConfig{Directory: writeCatalog(t, map[string]string{"Civilisation.ctp2synthetic": record})}); err == nil {
				t.Fatal("invalid civilization metadata was accepted")
			}
		})
	}
}

func TestParseRejectsCustomAndUnsupportedCTP2ConstructsWithLocation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		files map[string]string
		code  string
		id    string
	}{
		{name: "legacy custom grammar", files: map[string]string{"Units.txt": "[unit:scout]\nname=Scout\n"}, code: ErrorMalformedValue},
		{name: "duplicate ID", files: map[string]string{"Units.txt": "UNIT_SYNTHETIC { Description A Role starter MaxMovePoints 1 }\n", "units.ctp2synthetic": "UNIT_SYNTHETIC { Description B Role starter MaxMovePoints 1 }\n"}, code: ErrorDuplicateID, id: "UNIT_SYNTHETIC"},
		{name: "malformed value", files: map[string]string{"Units.txt": "UNIT_SYNTHETIC { Description A Role starter MaxMovePoints fast }\n"}, code: ErrorMalformedValue, id: "UNIT_SYNTHETIC"},
		{name: "unsupported behavior", files: map[string]string{"Units.txt": "UNIT_SYNTHETIC { Description A Role starter MaxMovePoints 1 TeleportAnywhere }\n"}, code: ErrorUnsupportedRequiredTag, id: "UNIT_SYNTHETIC"},
		{name: "nested behavior", files: map[string]string{"Units.txt": "UNIT_SYNTHETIC { Description A Role starter MaxMovePoints 1 CargoData { MaxCargo 1 } }\n"}, code: ErrorUnsupportedRequiredTag, id: "UNIT_SYNTHETIC"},
		{name: "unsupported file semantics", files: map[string]string{"Const.txt": "CONST_SYNTHETIC { Value 1 }\n"}, code: ErrorUnsupportedMechanic, id: "CONST_SYNTHETIC"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Import(ImportConfig{Directory: writeCatalog(t, tc.files)})
			var importErr *ImportError
			if !errors.As(err, &importErr) {
				t.Fatalf("Import() error = %v, want ImportError", err)
			}
			if importErr.Code != tc.code || (tc.id != "" && importErr.RecordID != tc.id) || importErr.Filename == "" || importErr.Line == 0 {
				t.Fatalf("ImportError = %+v, want code=%q id=%q and source location", importErr, tc.code, tc.id)
			}
		})
	}
}

func TestManifestInventoryMapsEveryBaselineEvaluatorToOriginalVector(t *testing.T) {
	t.Parallel()

	manifest := SupportedManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	got := make([]string, 0, len(manifest.VictoryEvaluators))
	for _, evaluator := range manifest.VictoryEvaluators {
		got = append(got, evaluator.ID)
	}
	sort.Strings(got)
	if want := "bloodlust_conquest,diplomatic_world_union,science_gaia_controller,score_2300"; strings.Join(got, ",") != want {
		t.Fatalf("victory evaluator inventory = %q, want %q", strings.Join(got, ","), want)
	}
	manifest.VictoryVectors = manifest.VictoryVectors[:len(manifest.VictoryVectors)-1]
	if err := manifest.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing evaluator vector rejection")
	}
}

func TestCanonicalHashIsIndependentOfFileAndRecordOrder(t *testing.T) {
	t.Parallel()

	first, err := Import(ImportConfig{Directory: fixtureDirectory(t)})
	if err != nil {
		t.Fatalf("first Import() error = %v", err)
	}
	second, err := Import(ImportConfig{Directory: writeCatalog(t, map[string]string{
		"Wonder.ctp2synthetic":             "WONDER_SYNTHETIC { Description WONDER_SYNTHETIC }\n",
		"Advance.ctp2synthetic":            "ADVANCE_SYNTHETIC_UNLOCK { PrerequisiteAdvanceIDs ADVANCE_SYNTHETIC_ROOT Description ADVANCE_SYNTHETIC_UNLOCK Cost 2 } ADVANCE_SYNTHETIC_ROOT { Cost 2 Description ADVANCE_SYNTHETIC_ROOT }\n",
		"Units.ctp2synthetic":              "UNIT_SYNTHETIC { ZOC true VisibilityRange 2 Strength 4 Role starter MaxMovePoints 2 Health 10 Effects \"settler,special_embassy,special_steal:100,special_capture:0,special_terraform:1,special_terraform_turns:1\" Domain land Description UNIT_SYNTHETIC Defense 4 Attack 4 Armor 2 }\n",
		"terrain.ctp2synthetic":            "TERRAIN_SYNTHETIC { Description TERRAIN_SYNTHETIC }\n",
		"TerrainImprovement.ctp2synthetic": "TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD { Description TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD BuildTurns 1 PublicWorksCost 1 MovementCostDelta -1 } TERRAIN_IMPROVEMENT_SYNTHETIC_FARM { Description TERRAIN_IMPROVEMENT_SYNTHETIC_FARM BuildTurns 1 PublicWorksCost 1 YieldFood 1 } TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP { Description TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP BuildTurns 1 PublicWorksCost 1 PollutionDelta -2 }\n",
		"PollutionSource.ctp2synthetic":    "POLLUTION_SOURCE_SYNTHETIC_FACTORY { Description POLLUTION_SOURCE_SYNTHETIC_FACTORY PollutionPoints 2 EffectScope city }\n",
		"Government.ctp2synthetic":         "GOVERNMENT_SYNTHETIC_REPUBLIC { Description GOVERNMENT_SYNTHETIC_REPUBLIC EnableAdvanceID ADVANCE_SYNTHETIC_ROOT ResearchBonus 1 }\n",
		"buildings.ctp2synthetic":          "IMPROVE_SYNTHETIC { Description IMPROVE_SYNTHETIC }\n",
		"EndGameObjects.ctp2synthetic":     "ENDGAME_UNION_VECTOR { TurnsToActivate 0 } ENDGAME_GAIA_VECTOR { TurnsToActivate 1 }\n",
	})})
	if err != nil {
		t.Fatalf("second Import() error = %v", err)
	}
	if first.ContentHash != second.ContentHash || string(first.CanonicalJSON) != string(second.CanonicalJSON) {
		t.Fatalf("canonical imports differ:\nfirst=%s %s\nsecond=%s %s", first.ContentHash, first.CanonicalJSON, second.ContentHash, second.CanonicalJSON)
	}
}

func TestProtectedAssetGuard(t *testing.T) {
	t.Parallel()

	if err := CheckProtectedAssets([]AssetFile{{Path: "internal/gameplay/rules/testdata/synthetic-ctp2/Units.ctp2synthetic", Content: []byte("UNIT_SYNTHETIC { Description Unit Role starter MaxMovePoints 1 }")}}); err != nil {
		t.Fatalf("synthetic fixture rejected: %v", err)
	}
	if err := CheckProtectedAssets([]AssetFile{{Path: "data/Units.txt"}}); err == nil {
		t.Fatal("retail filename was accepted")
	}
	if err := CheckTrackedAssets(repositoryRoot(t)); err != nil {
		t.Fatalf("CheckTrackedAssets() error = %v", err)
	}
}

func TestProtectedAssetGuardRejectsTrackedContentMarker(t *testing.T) {
	dir := t.TempDir()
	const filename = "fixture.txt"
	marker := strings.Join([]string{"CTP2", "RETAIL", "ASSET"}, "_")
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(marker), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := checkTrackedFiles(dir, []string{filename}); err == nil {
		t.Fatal("checkTrackedFiles() error = nil, want protected content rejection")
	}
}

func fixtureDirectory(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "synthetic-ctp2")
}

func writeCatalog(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile(%q): %v", name, err)
		}
	}
	return dir
}

func recordIDs(records []Record) string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	return strings.Join(ids, ",")
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("Abs repository root: %v", err)
	}
	return root
}
