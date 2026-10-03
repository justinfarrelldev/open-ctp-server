package rules

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestRetailFieldInventorySnapshot(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}

	snapshots := []struct {
		catalog   string
		fields    []string
		wantCount int
		mustHave  []string
	}{
		{catalog: "unit", fields: retailUnitFieldInventory(db.Units), wantCount: 194, mustHave: []string{"Attack", "CargoData.MaxCargo", "MaxFuel", "NoFuelThenCrash", "PlantNuke.EscapeChance", "UpgradeTo"}},
		{catalog: "advance", fields: retailAdvanceFieldInventory(db.Advances), wantCount: 19, mustHave: []string{"Cost", "Prerequisites", "PollutionProductionModifier", "TriggerFeat"}},
		{catalog: "feat", fields: retailFeatFieldInventory(db.Feats), wantCount: 13, mustHave: []string{"Duration", "EffectIncreaseProduction", "EffectReduceCityWalls", "Building.Building"}},
		{catalog: "wonder", fields: retailWonderFieldInventory(db.Wonders), wantCount: 39, mustHave: []string{"ProductionCost", "StartGaiaController", "PollutersToParks", "IncreaseScientists"}},
		{catalog: "building", fields: retailBuildingFieldInventory(db.Buildings), wantCount: 42, mustHave: []string{"ProductionCost", "PopulationPollutionPercent", "ProductionPollutionPercent", "SciencePercent"}},
		{catalog: "terrain_improvement", fields: retailTerrainImprovementFieldInventory(db.TerrainImprovements), wantCount: 36, mustHave: []string{"CantBuildOn", "Excludes", "TerrainEffect.BonusProduction", "TerrainEffect.ProductionTime"}},
		{catalog: "terrain", fields: retailTerrainFieldInventory(db.Terrains), wantCount: 43, mustHave: []string{"EnvBase.Food", "EnvCity.Gold", "Resources", "TransformAdd.Materials"}},
		{catalog: "government", fields: retailGovernmentFieldInventory(db.Governments), wantCount: 69, mustHave: []string{"MaxScienceRate", "PollutionCoef", "ReadyWarCoef", "WarDiscontentPerUnit"}},
		{catalog: "pollution", fields: retailPollutionFieldInventory(db.PollutionTriggerLists), wantCount: 1, mustHave: []string{"POLLUTION_TRIGGER"}},
	}

	for _, snapshot := range snapshots {
		if got := len(snapshot.fields); got != snapshot.wantCount {
			t.Fatalf("%s field count = %d, want %d; fields=%#v", snapshot.catalog, got, snapshot.wantCount, snapshot.fields)
		}
		for _, field := range snapshot.mustHave {
			if !retailFieldInventoryContains(snapshot.fields, field) {
				t.Fatalf("%s field inventory missing %q in %#v", snapshot.catalog, field, snapshot.fields)
			}
		}
	}
}

func TestRetailFieldCoverageLedgerMentionsEveryParsedField(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	ledgerPath := localRetailFieldCoverageLedger(t)
	ledger, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read retail field coverage ledger %s: %v", ledgerPath, err)
	}
	ledgerText := string(ledger)
	inventories := map[string][]string{
		"unit":                retailUnitFieldInventory(db.Units),
		"advance":             retailAdvanceFieldInventory(db.Advances),
		"feat":                retailFeatFieldInventory(db.Feats),
		"wonder":              retailWonderFieldInventory(db.Wonders),
		"building":            retailBuildingFieldInventory(db.Buildings),
		"terrain_improvement": retailTerrainImprovementFieldInventory(db.TerrainImprovements),
		"terrain":             retailTerrainFieldInventory(db.Terrains),
		"government":          retailGovernmentFieldInventory(db.Governments),
		"pollution":           retailPollutionFieldInventory(db.PollutionTriggerLists),
	}
	for catalog, fields := range inventories {
		for _, field := range fields {
			if !strings.Contains(ledgerText, field) {
				t.Fatalf("retail field coverage ledger missing %s field %q", catalog, field)
			}
		}
	}
}

func localRetailFieldCoverageLedger(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, ".sisyphus", "evidence", "ctp2-full-retail-task-22-field-coverage-inventory.md")
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	t.Skip("retail field coverage ledger is not present")
	return ""
}

func retailUnitFieldInventory(values []RetailUnit) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailAdvanceFieldInventory(values []RetailAdvance) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailFeatFieldInventory(values []RetailFeat) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailWonderFieldInventory(values []RetailWonder) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailBuildingFieldInventory(values []RetailBuilding) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailTerrainImprovementFieldInventory(values []RetailTerrainImprovement) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailTerrainFieldInventory(values []RetailTerrain) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailGovernmentFieldInventory(values []RetailGovernment) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailPollutionFieldInventory(values []RetailPollutionTriggerList) []string {
	fields := make([][]RetailRawField, 0, len(values))
	for _, value := range values {
		fields = append(fields, value.RawFields)
	}
	return retailFieldInventory(fields)
}

func retailFieldInventory(records [][]RetailRawField) []string {
	set := make(map[string]bool)
	for _, fields := range records {
		collectRetailFieldInventory(fields, "", set)
	}
	result := make([]string, 0, len(set))
	for field := range set {
		result = append(result, field)
	}
	sort.Strings(result)
	return result
}

func collectRetailFieldInventory(fields []RetailRawField, prefix string, out map[string]bool) {
	for _, field := range fields {
		key := field.Key
		if prefix != "" {
			key = prefix + "." + key
		}
		out[key] = true
		collectRetailFieldInventory(field.Fields, key, out)
	}
}

func retailFieldInventoryContains(fields []string, want string) bool {
	index := sort.SearchStrings(fields, want)
	return index < len(fields) && fields[index] == want
}
