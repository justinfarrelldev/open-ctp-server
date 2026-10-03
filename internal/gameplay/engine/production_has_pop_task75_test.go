package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailHasPopAndCanBuildUnitsAreNotCityProduction(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 3}}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_LAND_CITY", Domain: DomainLand, Movement: 1, Health: 10, ProductionCost: 1, Effects: []rules.RuntimeStarterEffect{{ID: "city-unit", Kind: EffectHasPopAndCanBuild}}}}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_LAND_CITY", Kind: ProductionUnit, Cost: 1}}, runtime)
	if err == nil {
		t.Fatalf("HasPopAndCanBuild city-unit was accepted into normal production queue: %+v", state.Cities[0].Queue)
	}
}
