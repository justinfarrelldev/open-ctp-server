package rules

import "testing"

func TestLoadRetailRulesFromDirectoryCompilesTypedCatalogs(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	if got, want := len(db.Units), 74; got != want {
		t.Fatalf("len(Units) = %d, want %d", got, want)
	}
	if got, want := len(db.Advances), 107; got != want {
		t.Fatalf("len(Advances) = %d, want %d", got, want)
	}
	if got, want := len(db.Wonders), 30; got != want {
		t.Fatalf("len(Wonders) = %d, want %d", got, want)
	}
	if got, want := len(db.Buildings), 52; got != want {
		t.Fatalf("len(Buildings) = %d, want %d", got, want)
	}
	if got, want := len(db.TerrainImprovements), 40; got != want {
		t.Fatalf("len(TerrainImprovements) = %d, want %d", got, want)
	}
	if got, want := len(db.Terrains), 26; got != want {
		t.Fatalf("len(Terrains) = %d, want %d", got, want)
	}
	if got, want := len(db.Governments), 12; got != want {
		t.Fatalf("len(Governments) = %d, want %d", got, want)
	}
	if got, want := len(db.PollutionTriggerLists), 4; got != want {
		t.Fatalf("len(PollutionTriggerLists) = %d, want %d", got, want)
	}
	if db.PollutionModel.PopulationStart != 16 || db.PollutionModel.PopulationRatio != 15 || db.PollutionModel.ProductionStart != 300 || db.PollutionModel.ProductionRatio != 1 {
		t.Fatalf("pollution model = %+v", db.PollutionModel)
	}
	if got, want := len(db.AI.UnitBuildLists), 27; got != want {
		t.Fatalf("len(AI.UnitBuildLists) = %d, want %d", got, want)
	}
	if got, want := len(db.AI.BuildingBuildLists), 13; got != want {
		t.Fatalf("len(AI.BuildingBuildLists) = %d, want %d", got, want)
	}
	if got, want := len(db.AI.WonderBuildLists), 7; got != want {
		t.Fatalf("len(AI.WonderBuildLists) = %d, want %d", got, want)
	}
	if got, want := len(db.AI.AdvanceLists), 7; got != want {
		t.Fatalf("len(AI.AdvanceLists) = %d, want %d", got, want)
	}
	if got, want := len(db.AI.ImprovementLists), 4; got != want {
		t.Fatalf("len(AI.ImprovementLists) = %d, want %d", got, want)
	}
}

func TestLoadRetailRulesFromDirectoryPreservesTypedRetailValues(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}

	carrier, ok := db.Unit("UNIT_AIRCRAFT_CARRIER")
	if !ok {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER not found")
	}
	if carrier.Attack != 25 || carrier.Defense != 30 || carrier.Firepower != 1 || carrier.Armor != 1 || carrier.MaxHP != 10 {
		t.Fatalf("carrier combat stats = %+v", carrier)
	}
	if carrier.ShieldCost != 3000 || carrier.GoldHunger != 150 || carrier.MaxMovePoints != 400 || carrier.VisionRange != 3 {
		t.Fatalf("carrier economic/movement stats = %+v", carrier)
	}
	if got, want := carrier.MovementTypes, []string{"Sea", "ShallowWater"}; !sameStringSlices(got, want) {
		t.Fatalf("carrier MovementTypes = %#v, want %#v", got, want)
	}
	if got, want := carrier.CanCarry, []string{"MedAir", "SmallAir"}; !sameStringSlices(got, want) {
		t.Fatalf("carrier CanCarry = %#v, want %#v", got, want)
	}
	if carrier.CargoMax != 5 {
		t.Fatalf("carrier CargoMax = %d, want 5", carrier.CargoMax)
	}
	if !containsString(carrier.Flags, "CantCaptureCity") || !containsString(carrier.Flags, "CanPillage") {
		t.Fatalf("carrier flags do not include expected retail bits: %#v", carrier.Flags)
	}

	railroad, ok := db.Advance("ADVANCE_RAILROAD")
	if !ok {
		t.Fatalf("ADVANCE_RAILROAD not found")
	}
	if railroad.Cost != 10912 {
		t.Fatalf("ADVANCE_RAILROAD Cost = %d, want 10912", railroad.Cost)
	}
	if got, want := railroad.Prerequisites, []string{"ADVANCE_BANKING", "ADVANCE_MODERN_METALLURGY"}; !sameStringSlices(got, want) {
		t.Fatalf("ADVANCE_RAILROAD Prerequisites = %#v, want %#v", got, want)
	}
	if !containsString(railroad.Flags, "Railroads") {
		t.Fatalf("ADVANCE_RAILROAD flags = %#v, want Railroads", railroad.Flags)
	}

	zeroCrime, ok := db.Wonder("WONDER_ZERO_CRIME_BILL")
	if !ok {
		t.Fatalf("WONDER_ZERO_CRIME_BILL not found")
	}
	if zeroCrime.ProductionCost != 8600 || zeroCrime.EnableAdvance != "ADVANCE_CRIMINAL_CODE" {
		t.Fatalf("zero crime bill values = %+v", zeroCrime)
	}
	if got := zeroCrime.IntValues["DecCrimePercent"]; got != 60 {
		t.Fatalf("WONDER_ZERO_CRIME_BILL DecCrimePercent = %d, want 60", got)
	}

	factory, ok := db.Building("IMPROVE_FACTORY")
	if !ok {
		t.Fatalf("IMPROVE_FACTORY not found")
	}
	if factory.ProductionCost != 2025 || factory.Upkeep != 8 || factory.EnableAdvance != "ADVANCE_INDUSTRIAL_REVOLUTION" {
		t.Fatalf("factory values = %+v", factory)
	}
	if got := factory.FloatValues["ProductionPercent"]; got != 0.15 {
		t.Fatalf("IMPROVE_FACTORY ProductionPercent = %g, want 0.15", got)
	}
	if !containsString(factory.Flags, "AllowGrunts") {
		t.Fatalf("factory flags = %#v, want AllowGrunts", factory.Flags)
	}

	farms, ok := db.TerrainImprovement("TILEIMP_FARMS")
	if !ok {
		t.Fatalf("TILEIMP_FARMS not found")
	}
	if farms.Class != "Farm" || farms.Level != 1 {
		t.Fatalf("TILEIMP_FARMS class/level = %+v", farms)
	}
	if got, want := farms.ConstructionTiles, []int{3, 4, 4}; !sameInts(got, want) {
		t.Fatalf("TILEIMP_FARMS ConstructionTiles = %#v, want %#v", got, want)
	}
	if len(farms.Effects) != 1 {
		t.Fatalf("TILEIMP_FARMS Effects len = %d, want 1", len(farms.Effects))
	}
	if got, want := farms.Effects[0].Terrains, []string{"TERRAIN_DESERT", "TERRAIN_GRASSLAND", "TERRAIN_PLAINS"}; !sameStringSlices(got, want) {
		t.Fatalf("TILEIMP_FARMS effect terrains = %#v, want %#v", got, want)
	}
	if farms.Effects[0].BonusFood != 10 || farms.Effects[0].ProductionCost != 200 || farms.Effects[0].ProductionTime != 2 {
		t.Fatalf("TILEIMP_FARMS effect values = %+v", farms.Effects[0])
	}
	terraformGrassland, ok := db.TerrainImprovement("TILEIMP_TERRAFORM_GRASSLAND")
	if !ok || terraformGrassland.TerraformTerrain != "TERRAIN_GRASSLAND" {
		t.Fatalf("TILEIMP_TERRAFORM_GRASSLAND = %+v ok=%v", terraformGrassland, ok)
	}

	plains, ok := db.Terrain("TERRAIN_PLAINS")
	if !ok {
		t.Fatalf("TERRAIN_PLAINS not found")
	}
	if plains.InternalType != "Plains" || plains.EnvBase.Movement != 100 || plains.EnvBase.Food != 10 || plains.EnvBase.Shield != 10 || plains.EnvBase.Gold != 5 {
		t.Fatalf("TERRAIN_PLAINS values = %+v", plains)
	}
	if plains.EnvCity.Movement != 50 || plains.EnvCity.Food != 10 || plains.EnvCity.Shield != 15 || plains.EnvCity.Gold != 15 {
		t.Fatalf("TERRAIN_PLAINS EnvCity = %+v", plains.EnvCity)
	}
	if got, want := plains.MovementTypes, []string{"Air", "Land"}; !sameStringSlices(got, want) {
		t.Fatalf("TERRAIN_PLAINS movement types = %#v, want %#v", got, want)
	}

	democracy, ok := db.Government("GOVERNMENT_DEMOCRACY")
	if !ok {
		t.Fatalf("GOVERNMENT_DEMOCRACY not found")
	}
	if democracy.EnableAdvance != "ADVANCE_DEMOCRACY" || democracy.Rank != 2 || democracy.MaxScienceRate != 0.8 || democracy.PollutionCoef != 0.8 {
		t.Fatalf("GOVERNMENT_DEMOCRACY values = %+v", democracy)
	}

	smallMap, ok := db.PollutionTriggerList("POLLUTION_TRIGGER_LIST_SMALL_MAP")
	if !ok {
		t.Fatalf("POLLUTION_TRIGGER_LIST_SMALL_MAP not found")
	}
	if got, want := smallMap.TriggerLevels, []int64{4000000, 6000000, 8000000, 10000000, 15000000, 9000000000}; !sameInt64s(got, want) {
		t.Fatalf("small-map pollution levels = %#v, want %#v", got, want)
	}
	if got, want := smallMap.DisasterTypes, []string{"POLLUTION_FLOOD_DISASTER", "POLLUTION_OZONE_DISASTER", "POLLUTION_OZONE_DISASTER", "POLLUTION_FLOOD_DISASTER", "POLLUTION_OZONE_DISASTER", "POLLUTION_FLOOD_DISASTER"}; !sameStringSlices(got, want) {
		t.Fatalf("small-map pollution disasters = %#v, want %#v", got, want)
	}
}

func sameInts(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func sameInt64s(got, want []int64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
