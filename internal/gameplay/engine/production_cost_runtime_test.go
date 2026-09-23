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
