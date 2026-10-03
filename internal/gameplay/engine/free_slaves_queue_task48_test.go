package engine

import "testing"

func TestRetailFreeSlavesRemovesQueuedSlaveryUnitsEverywhere(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{
		{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Queue: []ProductionItem{{ID: "UNIT_SLAVER", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "raid", Kind: EffectSpecialSlaveRaid}}}, {ID: "UNIT_GUARD", Kind: ProductionUnit, Cost: 1}}},
		{ID: "city-b", OwnerID: "player-b", TileID: "tile-b", Queue: []ProductionItem{{ID: "UNIT_UPRISING", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "uprising", Kind: EffectSpecialSlaveUprising}}}}},
	}

	emancipateSlaves(&state, "player-a", nil)
	if len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].ID != "UNIT_GUARD" {
		t.Fatalf("city-a queue after emancipation = %#v", state.Cities[0].Queue)
	}
	if len(state.Cities[1].Queue) != 0 {
		t.Fatalf("city-b queue after emancipation = %#v", state.Cities[1].Queue)
	}
}
