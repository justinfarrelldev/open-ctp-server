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
}

func TestProjectRetailRuntimeMapsSelectedRetailValues(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	runtime := ProjectRetailRuntime(db)

	settler, ok := runtime.Unit("UNIT_SETTLER")
	if !ok {
		t.Fatalf("UNIT_SETTLER missing")
	}
	if settler.Movement != 100 || settler.Attack != 0 || settler.Defense != 10 || settler.Health != 10 || settler.ProductionCost != 740 {
		t.Fatalf("settler runtime = %+v", settler)
	}
	if !runtimeUnitHasEffect(settler, "settler") {
		t.Fatalf("settler effect missing: %+v", settler.Effects)
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
	if !ok || plains.Cost != 100 || plains.Food != 10 || plains.Production != 10 || plains.Gold != 5 || !sameStringSlices(plains.Domains, []string{"air", "land"}) {
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
	if !ok || democracy.EnableAdvanceID != "ADVANCE_DEMOCRACY" || democracy.ResearchBonus != 1 || democracy.ProductionPercent != 15 || democracy.GoldPercent != 25 {
		t.Fatalf("democracy government = %+v ok=%v", democracy, ok)
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
	academy, ok := runtime.ProductionDefinition("IMPROVE_ACADEMY")
	if !ok || len(academy.Effects) != 1 || academy.Effects[0].Kind != "research" || academy.Effects[0].Value != 10 {
		t.Fatalf("academy production effects = %+v ok=%v", academy, ok)
	}
	granaryEffects, ok := runtime.ProductionDefinition("IMPROVE_GRANARY")
	if !ok || len(granaryEffects.Effects) != 1 || granaryEffects.Effects[0].Kind != "yield_food_percent" || granaryEffects.Effects[0].Value != 10 {
		t.Fatalf("granary production effects = %+v ok=%v", granaryEffects, ok)
	}
}

func runtimeUnitHasEffect(unit RuntimeUnitDefinition, kind string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind {
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
