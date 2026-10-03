package engine

import "testing"

func TestRetailEnableParkRangersWonderGloballyUnlocksParkRangerProduction(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}
	state.Wonders = []Wonder{{ID: "WONDER_THE_EDEN_PROJECT", OwnerID: "player-b", CityID: "city-b", Scope: "nation", Effects: []Effect{{ID: "park-rangers", Kind: EffectEnableParkRangers}}}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_ECO_RANGER", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "create-park", Kind: EffectSpecialCreatePark}}}}, nil)
	if err != nil {
		t.Fatalf("queue park ranger with globally built Eden Project: %v", err)
	}
}

func TestRetailCreateParkProductionRequiresAnyEnableParkRangersWonder(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_ECO_RANGER", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "create-park", Kind: EffectSpecialCreatePark}}}}, nil)
	if err == nil {
		t.Fatal("park ranger production succeeded before any EnableParkRangers wonder existed")
	}
}
