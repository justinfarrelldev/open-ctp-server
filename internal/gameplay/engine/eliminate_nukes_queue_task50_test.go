package engine

import "testing"

func TestRetailEliminateNukesRemovesQueuedNuclearAttackUnits(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Queue: []ProductionItem{
		{ID: "UNIT_NUKE", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "nuke", Kind: EffectNuclearAttack}}},
		{ID: "UNIT_GUARD", Kind: ProductionUnit, Cost: 1},
	}}}

	eliminateNukes(&state)
	if len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].ID != "UNIT_GUARD" {
		t.Fatalf("queue after nuke elimination = %#v", state.Cities[0].Queue)
	}
}
