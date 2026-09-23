package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

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
