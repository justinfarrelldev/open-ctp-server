package engine

import "testing"

func TestRetailEliminateNukesWonderGloballyBlocksNuclearPlantBuildings(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}
	state.Wonders = []Wonder{{ID: "WONDER_NANITE_DEFUSER", OwnerID: "player-b", CityID: "city-b", Scope: "nation", Effects: []Effect{{ID: "defuse", Kind: EffectEliminateNukes}}}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_NUCLEAR_PLANT", Kind: ProductionBuilding, Cost: 1, Effects: []Effect{{ID: "plant", Kind: EffectNuclearPlant}}}}, nil)
	if err == nil {
		t.Fatal("nuclear plant production succeeded after any EliminateNukes wonder existed")
	}
}

func TestRetailNuclearPlantBuildingsAllowedBeforeEliminateNukesWonder(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "IMPROVE_NUCLEAR_PLANT", Kind: ProductionBuilding, Cost: 1, Effects: []Effect{{ID: "plant", Kind: EffectNuclearPlant}}}}, nil)
	if err != nil {
		t.Fatalf("nuclear plant production before EliminateNukes wonder: %v", err)
	}
}
