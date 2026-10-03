package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailBuildingEverywhereSatisfiesPrerequisiteBuilding(t *testing.T) {
	state := fixtureState()
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 3}}
	state.Wonders = []Wonder{{ID: "WONDER_HAGIA_SOPHIA", OwnerID: "player-a", Effects: []Effect{{ID: "IMPROVE_SHRINE", Kind: EffectBuildingEverywhere}}}}
	runtime := &rules.RuntimeRules{Production: []rules.RuntimeProduction{{ID: "IMPROVE_TEMPLE", Kind: ProductionBuilding, Cost: 1, RequiredBuildings: []string{"IMPROVE_SHRINE"}}}}

	if err := requireProductionUnlock(state, state.Cities[0], ProductionItem{Kind: ProductionBuilding, ID: "IMPROVE_TEMPLE"}, runtime); err != nil {
		t.Fatalf("building-everywhere wonder did not satisfy prerequisite building: %v", err)
	}
}

func TestRetailBuildingEverywhereRequiresActiveOwnerWonder(t *testing.T) {
	state := fixtureState()
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 3}}
	state.Wonders = []Wonder{{ID: "WONDER_HAGIA_SOPHIA", OwnerID: "player-b", Effects: []Effect{{ID: "IMPROVE_SHRINE", Kind: EffectBuildingEverywhere}}}}
	runtime := &rules.RuntimeRules{Production: []rules.RuntimeProduction{{ID: "IMPROVE_TEMPLE", Kind: ProductionBuilding, Cost: 1, RequiredBuildings: []string{"IMPROVE_SHRINE"}}}}

	if err := requireProductionUnlock(state, state.Cities[0], ProductionItem{Kind: ProductionBuilding, ID: "IMPROVE_TEMPLE"}, runtime); err == nil {
		t.Fatal("foreign building-everywhere wonder satisfied local prerequisite building")
	}
	state.Wonders[0].OwnerID = "player-a"
	state.Wonders[0].Destroyed = true
	if err := requireProductionUnlock(state, state.Cities[0], ProductionItem{Kind: ProductionBuilding, ID: "IMPROVE_TEMPLE"}, runtime); err == nil {
		t.Fatal("destroyed building-everywhere wonder satisfied local prerequisite building")
	}
}
