package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRuntimeBuildingAndWonderProductionUsesPinnedCostAndPrerequisite(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	runtime := &rules.RuntimeRules{Production: []rules.RuntimeProduction{
		{ID: "IMPROVE_COSTED", Kind: "building", Cost: 11, Upkeep: 3, Prerequisites: []string{"ADVANCE_COSTED"}},
		{ID: "WONDER_COSTED", Kind: "wonder", Cost: 17, Prerequisites: []string{"ADVANCE_COSTED"}},
	}}

	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_COSTED", Kind: ProductionBuilding, Cost: 1}}, runtime); err == nil {
		t.Fatal("queueProduction accepted a runtime building before its advance prerequisite")
	}
	state.Nations = []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_COSTED"}}}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_COSTED", Kind: ProductionBuilding, Cost: 1}}, runtime); err != nil {
		t.Fatalf("queueProduction(building) error = %v", err)
	}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "WONDER_COSTED", Kind: ProductionWonder, Cost: 1}}, runtime); err != nil {
		t.Fatalf("queueProduction(wonder) error = %v", err)
	}
	city, ok := cityByID(state, "city-a")
	if !ok || len(city.Queue) != 2 {
		t.Fatalf("city queue = %#v", city)
	}
	if got := city.Queue[0]; got.Cost != 11 || got.Upkeep != 3 || got.Prerequisite != "ADVANCE_COSTED" {
		t.Fatalf("queued building = %+v", got)
	}
	if got := city.Queue[1]; got.Cost != 17 || got.Prerequisite != "ADVANCE_COSTED" {
		t.Fatalf("queued wonder = %+v", got)
	}
}

func TestRuntimePopulationRemovingUnitProductionRequiresSparePopulation(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1, ProductionCost: 1, Effects: []rules.RuntimeStarterEffect{{ID: "pop", Kind: EffectBuildingRemovesPopulation}}}}}

	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, runtime); err == nil {
		t.Fatal("queueProduction accepted a population-removing unit at population one")
	}
	state.Cities[0].Population = 2
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, runtime); err != nil {
		t.Fatalf("queueProduction with spare population error = %v", err)
	}
}

func TestRuntimeCantBuildUnitCannotBeQueued(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_CITY", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "cant-build", Kind: EffectCantBuild}}}}}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_CITY", Kind: ProductionUnit, Cost: 1}}, runtime); err == nil {
		t.Fatal("CantBuild unit production succeeded")
	}
}

func TestRuntimeNoSlavesAndSlaverProductionExcludeEachOther(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{
		{ID: "UNIT_ABOLITIONIST", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "no-slaves", Kind: EffectNoSlaves}}},
		{ID: "UNIT_SLAVER", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "slave", Kind: EffectSpecialSlaveRaid}}},
	}}
	state.Units = append(state.Units, Unit{ID: "abolitionist", OwnerID: "player-a", TileID: "a", Effects: []Effect{{ID: "no-slaves", Kind: EffectNoSlaves}}})
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_SLAVER", Kind: ProductionUnit, Cost: 1}}, runtime); err == nil {
		t.Fatal("slaver production succeeded while no-slaves unit exists")
	}
	state.Units = []Unit{{ID: "slaver", OwnerID: "player-a", TileID: "a", Effects: []Effect{{ID: "slave", Kind: EffectSpecialSlaveRaid}}}}
	state.Cities[0].Queue = nil
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_ABOLITIONIST", Kind: ProductionUnit, Cost: 1}}, runtime); err == nil {
		t.Fatal("no-slaves production succeeded while slaver unit exists")
	}
}

func TestRuntimeUnitProductionRepeatsUnlessOnlyBuildOne(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Cities[0].Yields.Production = 2
	state.Cities[0].Queue = []ProductionItem{{ID: "UNIT_REPEAT", Kind: ProductionUnit, Cost: 1}}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_ONE", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "only-one", Kind: EffectOnlyBuildOne}}}, {ID: "UNIT_REPEAT", Domain: DomainLand, Movement: 1, Health: 1}}}
	if err := completeProduction(&state, &state.Cities[0], runtime); err != nil {
		t.Fatalf("completeProduction(repeat) error = %v", err)
	}
	if len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].ID != "UNIT_REPEAT" || len(state.Units) != 2 {
		t.Fatalf("repeat unit production should keep sole queue item: city=%+v units=%+v", state.Cities[0], state.Units)
	}
	state.Cities[0].Queue = []ProductionItem{{ID: "UNIT_ONE", Kind: ProductionUnit, Cost: 1}}
	if err := completeProduction(&state, &state.Cities[0], runtime); err != nil {
		t.Fatalf("completeProduction(only-build-one) error = %v", err)
	}
	if len(state.Cities[0].Queue) != 0 || len(state.Units) != 3 {
		t.Fatalf("only-build-one unit production should consume queue item: city=%+v units=%+v", state.Cities[0], state.Units)
	}
}

func TestRuntimeBuildingProductionRequiresPrerequisiteBuildings(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	runtime := &rules.RuntimeRules{Production: []rules.RuntimeProduction{{ID: "IMPROVE_BANK", Kind: "building", Cost: 1, RequiredBuildings: []string{"IMPROVE_BAZAAR"}}}}

	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_BANK", Kind: ProductionBuilding, Cost: 1}}, runtime); err == nil {
		t.Fatal("queueProduction accepted a building before its prerequisite building")
	}
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_BAZAAR"}}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_BANK", Kind: ProductionBuilding, Cost: 1}}, runtime); err != nil {
		t.Fatalf("queueProduction() with prerequisite building error = %v", err)
	}
}

func TestRuntimeBuildingProductionHonorsSeaAndCoastalPlacement(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Tiles = []Tile{{ID: state.Cities[0].TileID, X: 0, Y: 0, Terrain: TerrainOcean}, {ID: "land", X: 1, Y: 0, Terrain: TerrainPlains}, {ID: "inland", X: 4, Y: 4, Terrain: TerrainPlains}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainOcean, Domains: []string{DomainSea}}, {ID: TerrainPlains, Domains: []string{DomainLand}}}}, Production: []rules.RuntimeProduction{{ID: "IMPROVE_BATTLEMENTS", Kind: ProductionBuilding, Cost: 1, CoastalOnly: true}, {ID: "IMPROVE_GRANARY", Kind: ProductionBuilding, Cost: 1, CantBuildInSea: true}}}

	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_GRANARY", Kind: ProductionBuilding, Cost: 1}}, runtime); err == nil {
		t.Fatal("queueProduction accepted a cant-build-in-sea building in a sea city")
	}
	state.Cities[0].TileID = "inland"
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_BATTLEMENTS", Kind: ProductionBuilding, Cost: 1}}, runtime); err == nil {
		t.Fatal("queueProduction accepted a coastal building in an inland city")
	}
	state.Cities[0].TileID = "land"
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_BATTLEMENTS", Kind: ProductionBuilding, Cost: 1}}, runtime); err != nil {
		t.Fatalf("queueProduction coastal city error = %v", err)
	}
}

func TestRuntimeBuildingProductionRejectsAlreadyBuiltCityImprovement(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_GRANARY"}}
	runtime := &rules.RuntimeRules{Production: []rules.RuntimeProduction{{ID: "IMPROVE_GRANARY", Kind: "building", Cost: 1}}}

	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_GRANARY", Kind: ProductionBuilding, Cost: 1}}, runtime); err == nil {
		t.Fatal("queueProduction accepted an already-built city improvement")
	}
}

func TestRuntimeProductionEffectsApplyToCompletedBuildingYields(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Nations = []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_FACTORY"}}}
	runtime := &rules.RuntimeRules{Production: []rules.RuntimeProduction{{ID: "IMPROVE_FACTORY", Kind: "building", Cost: 1, Prerequisites: []string{"ADVANCE_FACTORY"}, Effects: []rules.RuntimeStarterEffect{{ID: "factory-production", Kind: EffectYieldProductionPercent, Value: 50}}}}}

	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_FACTORY", Kind: ProductionBuilding, Cost: 1}}, runtime); err != nil {
		t.Fatalf("queueProduction() error = %v", err)
	}
	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatalf("RecalculateRound() error = %v", err)
	}
	if len(next.Cities[0].Buildings) != 1 || len(next.Cities[0].Buildings[0].Effects) != 1 {
		t.Fatalf("completed building = %#v", next.Cities[0].Buildings)
	}
	if got, want := next.Cities[0].Yields.Production, 5; got != want { // plains production 3 plus ceil(50%) 2, no upkeep.
		t.Fatalf("production yield = %d, want %d", got, want)
	}
}

func TestRuntimeUnitProductionUsesPinnedShieldCost(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_COSTED", ProductionCost: 7}}}

	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_COSTED", Kind: ProductionUnit, Cost: 1}}, runtime); err != nil {
		t.Fatalf("queueProduction() error = %v", err)
	}
	city, ok := cityByID(state, "city-a")
	if !ok || len(city.Queue) != 1 {
		t.Fatalf("city queue = %#v", city)
	}
	if got := city.Queue[0].Cost; got != 7 {
		t.Fatalf("queued runtime unit cost = %d, want pinned cost 7", got)
	}
}
