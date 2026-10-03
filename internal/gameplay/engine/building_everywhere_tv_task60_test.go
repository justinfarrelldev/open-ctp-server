package engine

import "testing"

func TestRetailTelevisionGoldCountsWonderEffectiveBuildings(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}},
		Nations: []Nation{{ID: "nation-a"}, {ID: "nation-b"}},
		Cities:  []City{{ID: "foreign", OwnerID: "player-b", TileID: "foreign-tile", Population: 4}},
		Wonders: []Wonder{
			{ID: "WONDER_HOLLYWOOD", OwnerID: "player-a", Effects: []Effect{{ID: "tv-gold", Kind: EffectGoldPerTelevision, Value: 5}}},
			{ID: "WONDER_EFFECTIVE_TV", OwnerID: "player-b", Effects: []Effect{{ID: "IMPROVE_TV", Kind: EffectBuildingEverywhere}, {ID: "tv", Kind: EffectTelevision}}},
		},
	}
	if got := wonderGoldForPlayer(state, "player-a"); got != 20 {
		t.Fatalf("television wonder gold = %d, want 20 from foreign effective television building", got)
	}
}
