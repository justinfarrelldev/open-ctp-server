package engine

import "testing"

func TestRetailEliminateNukesWonderGloballyBlocksNuclearAttackUnitProduction(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}
	state.Wonders = []Wonder{{ID: "WONDER_NANITE_DEFUSER", OwnerID: "player-b", CityID: "city-b", Scope: "nation", Effects: []Effect{{ID: "defuse", Kind: EffectEliminateNukes}}}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_NUKE", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "nuke", Kind: EffectNuclearAttack}}}}, nil)
	if err == nil {
		t.Fatal("nuclear attack unit production succeeded after any EliminateNukes wonder existed")
	}
}
