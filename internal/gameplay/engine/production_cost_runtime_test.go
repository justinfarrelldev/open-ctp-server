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
