package rules

import "testing"

func TestProjectRetailRuntimeIncludesCompleteReleasedCatalog(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	runtime := ProjectRetailRuntime(db)
	if got, want := len(runtime.Units), 74; got != want {
		t.Fatalf("runtime units = %d, want %d", got, want)
	}
	if got, want := len(runtime.Advances), 107; got != want {
		t.Fatalf("runtime advances = %d, want %d", got, want)
	}
	if got, want := len(runtime.Production), 156; got != want {
		t.Fatalf("runtime production = %d, want %d", got, want)
	}
	if got, want := len(runtime.Movement.Terrains), 26; got != want {
		t.Fatalf("runtime terrains = %d, want %d", got, want)
	}
	if got, want := len(runtime.TerrainImprovements), 40; got != want {
		t.Fatalf("runtime terrain improvements = %d, want %d", got, want)
	}
	if got, want := len(runtime.PollutionTriggers), 4; got != want {
		t.Fatalf("runtime pollution triggers = %d, want %d", got, want)
	}
	if runtime.PollutionModel == nil || runtime.PollutionModel.PopulationStart != 16 || runtime.PollutionModel.PopulationRatio != 15 || runtime.PollutionModel.ProductionStart != 300 || runtime.PollutionModel.ProductionRatio != 1 {
		t.Fatalf("runtime pollution model = %+v", runtime.PollutionModel)
	}
	if runtime.Gaia.ActivationTurns != 10 || runtime.Gaia.BuildingID != "IMPROVE_GAIA_COMPUTER" || runtime.Gaia.WonderID != "WONDER_THE_SOLARIS_PROJECT" {
		t.Fatalf("runtime Gaia = %+v", runtime.Gaia)
	}
	if runtime.AI == nil || len(runtime.AI.UnitBuildLists) != 27 || len(runtime.AI.BuildingBuildLists) != 13 || len(runtime.AI.WonderBuildLists) != 7 || len(runtime.AI.AdvanceLists) != 7 || len(runtime.AI.ImprovementLists) != 4 {
		t.Fatalf("runtime AI lists = %+v", runtime.AI)
	}
}

func TestProjectRetailRuntimeMapsSelectedRetailValues(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	runtime := ProjectRetailRuntime(db)

	if runtime.AI == nil || runtime.AI.UnitBuildLists[0].ID != "UNIT_BUILD_LIST_SPECIAL" || !sameStringSlices(runtime.AI.UnitBuildLists[0].Items[:3], []string{"UNIT_DIPLOMAT", "UNIT_SLAVER", "UNIT_ABOLITIONIST"}) {
		t.Fatalf("runtime AI unit list = %+v", runtime.AI)
	}
	if runtime.AI.BuildingBuildLists[1].ID != "BUILDING_BUILD_LIST_GROWTH" || !containsString(runtime.AI.BuildingBuildLists[1].Items, "IMPROVE_GRANARY") {
		t.Fatalf("runtime AI building lists = %+v", runtime.AI.BuildingBuildLists)
	}

	settler, ok := runtime.Unit("UNIT_SETTLER")
	if !ok {
		t.Fatalf("UNIT_SETTLER missing")
	}
	if settler.Movement != 100 || settler.Attack != 0 || settler.Defense != 10 || settler.Health != 10 || settler.ProductionCost != 740 {
		t.Fatalf("settler runtime = %+v", settler)
	}
	if !runtimeUnitHasEffect(settler, "settler") || !runtimeUnitHasEffectValue(settler, "settle_size", 1) {
		t.Fatalf("settler effects missing: %+v", settler.Effects)
	}
	seaEngineer, ok := runtime.Unit("UNIT_SEA_ENGINEER")
	if !ok || !runtimeUnitHasEffectID(seaEngineer, "settle_building", "IMPROVE_GRANARY") {
		t.Fatalf("sea engineer runtime = %+v ok=%v", seaEngineer, ok)
	}
	if runtime.Starter.Movement != 100 || runtime.Starter.Health != 10 || !runtimeStarterHasEffect(runtime.Starter, "settler") {
		t.Fatalf("starter = %+v, want retail settler", runtime.Starter)
	}

	carrier, ok := runtime.Unit("UNIT_AIRCRAFT_CARRIER")
	if !ok {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER missing")
	}
	if carrier.Attack != 25 || carrier.Defense != 30 || carrier.Ranged != 25 || carrier.BombardRange != 1 || carrier.BombRounds != 10 || carrier.Movement != 400 || carrier.TransportCapacity != 5 || carrier.ProductionCost != 3000 || carrier.Upkeep != 150 {
		t.Fatalf("carrier runtime = %+v", carrier)
	}
	if got, want := carrier.CargoDomains, []string{"air"}; !sameStringSlices(got, want) {
		t.Fatalf("carrier cargo domains = %#v, want %#v", got, want)
	}
	fighter, ok := runtime.Unit("UNIT_FIGHTER")
	if !ok || fighter.Domain != "air" || fighter.MaxFuel != 2400 {
		t.Fatalf("fighter runtime = %+v ok=%v", fighter, ok)
	}

	railroad, ok := runtime.Advance("ADVANCE_RAILROAD")
	if !ok || railroad.Cost != 10912 || !sameStringSlices(railroad.Prerequisites, []string{"ADVANCE_BANKING", "ADVANCE_MODERN_METALLURGY"}) {
		t.Fatalf("railroad runtime = %+v ok=%v", railroad, ok)
	}

	plains, ok := runtime.Movement.Terrain("TERRAIN_PLAINS")
	if !ok || plains.Cost != 100 || plains.Food != 10 || plains.Production != 10 || plains.Gold != 5 || plains.CityFood != 10 || plains.CityProduction != 15 || plains.CityGold != 15 || !sameStringSlices(plains.Domains, []string{"air", "land"}) {
		t.Fatalf("plains runtime terrain = %+v ok=%v", plains, ok)
	}

	farms, ok := runtime.TerrainImprovement("TILEIMP_FARMS")
	if !ok || farms.BuildTurns != 2 || farms.PublicWorksCost != 200 || farms.YieldFood != 10 || farms.Class != "Farm" || !retailContainsString(farms.CantBuildOn, "TERRAIN_WATER_SHELF") || !retailContainsString(farms.Excludes, "Mine") {
		t.Fatalf("farms runtime = %+v ok=%v", farms, ok)
	}

	granary, ok := runtime.ProductionDefinition("IMPROVE_GRANARY")
	if !ok || granary.Kind != "building" || granary.Cost != 300 || granary.Upkeep != 1 || !sameStringSlices(granary.Prerequisites, []string{"ADVANCE_AGRICULTURE"}) {
		t.Fatalf("granary production = %+v ok=%v", granary, ok)
	}
	democracy, ok := runtime.Government("GOVERNMENT_DEMOCRACY")
	if !ok || democracy.EnableAdvanceID != "ADVANCE_DEMOCRACY" || democracy.ResearchBonus != 1 || democracy.FoodPercent != 25 || democracy.ProductionPercent != 15 || democracy.GoldPercent != 25 || democracy.PollutionPercent != -20 {
		t.Fatalf("democracy government = %+v ok=%v", democracy, ok)
	}
	anarchy, ok := runtime.Government("GOVERNMENT_ANARCHY")
	if !ok || anarchy.FoodPercent != -25 || anarchy.ProductionPercent != -25 || anarchy.GoldPercent != -25 {
		t.Fatalf("anarchy government = %+v ok=%v", anarchy, ok)
	}

	hagia, ok := runtime.ProductionDefinition("WONDER_HAGIA_SOPHIA")
	if !ok || hagia.Kind != "wonder" || hagia.Cost != 3240 || !sameStringSlices(hagia.Prerequisites, []string{"ADVANCE_THEOLOGY"}) {
		t.Fatalf("hagia production = %+v ok=%v", hagia, ok)
	}
	factory, ok := runtime.ProductionDefinition("IMPROVE_FACTORY")
	if !ok || len(factory.Effects) != 1 || factory.Effects[0].Kind != "yield_production_percent" || factory.Effects[0].Value != 15 {
		t.Fatalf("factory production effects = %+v ok=%v", factory, ok)
	}
	bazaar, ok := runtime.ProductionDefinition("IMPROVE_BAZAAR")
	if !ok || len(bazaar.Effects) != 1 || bazaar.Effects[0].Kind != "yield_gold_percent" || bazaar.Effects[0].Value != 10 {
		t.Fatalf("bazaar production effects = %+v ok=%v", bazaar, ok)
	}
	ecoTransit, ok := runtime.ProductionDefinition("IMPROVE_ECO_TRANSIT")
	if !ok || !runtimeProductionHasEffectValue(ecoTransit, "pollution_population_percent", -50) {
		t.Fatalf("eco transit production effects = %+v ok=%v", ecoTransit, ok)
	}
	matter, ok := runtime.ProductionDefinition("IMPROVE_MATTER_DECOMPILER")
	if !ok || !runtimeProductionHasEffectValue(matter, "pollution_production_percent", -50) {
		t.Fatalf("matter decomposer production effects = %+v ok=%v", matter, ok)
	}

	television, ok := runtime.ProductionDefinition("IMPROVE_TELEVISION")
	if !ok || !runtimeProductionHasEffectValue(television, "yield_gold_per_population", 5) {
		t.Fatalf("television production effects = %+v ok=%v", television, ok)
	}

	academy, ok := runtime.ProductionDefinition("IMPROVE_ACADEMY")
	if !ok || len(academy.Effects) != 1 || academy.Effects[0].Kind != "research" || academy.Effects[0].Value != 10 {
		t.Fatalf("academy production effects = %+v ok=%v", academy, ok)
	}
	granaryEffects, ok := runtime.ProductionDefinition("IMPROVE_GRANARY")
	if !ok || len(granaryEffects.Effects) != 1 || granaryEffects.Effects[0].Kind != "yield_food_percent" || granaryEffects.Effects[0].Value != 10 {
		t.Fatalf("granary production effects = %+v ok=%v", granaryEffects, ok)
	}
	academyWonder, ok := runtime.ProductionDefinition("WONDER_ARISTOTLES_LYCEUM")
	if !ok || len(academyWonder.Effects) != 1 || academyWonder.Effects[0].Kind != "research" || academyWonder.Effects[0].Value != 20 {
		t.Fatalf("Aristotle's Lyceum effects = %+v ok=%v", academyWonder, ok)
	}
	pyramids, ok := runtime.ProductionDefinition("WONDER_PYRAMIDS")
	if !ok || len(pyramids.Effects) != 1 || pyramids.Effects[0].Kind != "yield_gold" || pyramids.Effects[0].Value != 100 {
		t.Fatalf("Pyramids effects = %+v ok=%v", pyramids, ok)
	}
}

func runtimeProductionHasEffectValue(production RuntimeProduction, kind string, value int) bool {
	for _, effect := range production.Effects {
		if effect.Kind == kind && effect.Value == value {
			return true
		}
	}
	return false
}

func runtimeUnitHasEffect(unit RuntimeUnitDefinition, kind string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func runtimeUnitHasEffectValue(unit RuntimeUnitDefinition, kind string, value int) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind && effect.Value == value {
			return true
		}
	}
	return false
}

func runtimeUnitHasEffectID(unit RuntimeUnitDefinition, kind string, id string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind && effect.ID == id {
			return true
		}
	}
	return false
}

func runtimeStarterHasEffect(starter RuntimeStarter, kind string) bool {
	for _, effect := range starter.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}
