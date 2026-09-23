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
	if carrier.Attack != 25 || carrier.Defense != 30 || carrier.Movement != 400 || carrier.TransportCapacity != 5 || carrier.ProductionCost != 3000 || carrier.Upkeep != 150 {
		t.Fatalf("carrier runtime = %+v", carrier)
	}
	if got, want := carrier.CargoDomains, []string{"air"}; !sameStringSlices(got, want) {
		t.Fatalf("carrier cargo domains = %#v, want %#v", got, want)
	}

	railroad, ok := runtime.Advance("ADVANCE_RAILROAD")
	if !ok || railroad.Cost != 10912 || !sameStringSlices(railroad.Prerequisites, []string{"ADVANCE_BANKING", "ADVANCE_MODERN_METALLURGY"}) {
		t.Fatalf("railroad runtime = %+v ok=%v", railroad, ok)
	}

	plains, ok := runtime.Movement.Terrain("TERRAIN_PLAINS")
	if !ok || plains.Cost != 100 || !sameStringSlices(plains.Domains, []string{"air", "land"}) {
		t.Fatalf("plains runtime terrain = %+v ok=%v", plains, ok)
	}

	farms, ok := runtime.TerrainImprovement("TILEIMP_FARMS")
	if !ok || farms.BuildTurns != 2 || farms.PublicWorksCost != 200 || farms.YieldFood != 10 {
		t.Fatalf("farms runtime = %+v ok=%v", farms, ok)
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
