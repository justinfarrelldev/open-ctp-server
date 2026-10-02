package rules

import (
	"sort"
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
