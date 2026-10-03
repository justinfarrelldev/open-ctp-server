package engine

import "testing"

func TestRetailFreeSlavesWonderGloballyBlocksSlaveryUnitProduction(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}
	state.Wonders = []Wonder{{ID: "WONDER_EMANCIPATION", OwnerID: "player-b", CityID: "city-b", Scope: "nation", Effects: []Effect{{ID: "free", Kind: EffectFreeSlaves}}}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_SLAVER", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "raid", Kind: EffectSpecialSlaveRaid}}}}, nil)
	if err == nil {
		t.Fatal("slaver production succeeded after any FreeSlaves wonder existed")
	}
}

func TestRetailFreeSlavesWonderGloballyBlocksSlaveUprisingProduction(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}
	state.Wonders = []Wonder{{ID: "WONDER_EMANCIPATION", OwnerID: "player-b", CityID: "city-b", Scope: "nation", Effects: []Effect{{ID: "free", Kind: EffectFreeSlaves}}}}

	err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_AGITATOR", Kind: ProductionUnit, Cost: 1, Effects: []Effect{{ID: "uprising", Kind: EffectSpecialSlaveUprising}}}}, nil)
	if err == nil {
		t.Fatal("slave-uprising production succeeded after any FreeSlaves wonder existed")
	}
}
