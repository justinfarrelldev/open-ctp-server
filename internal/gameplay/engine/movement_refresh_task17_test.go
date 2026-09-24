package engine

import "testing"

func TestRetailObsoleteBoatMovementWonderDoesNotRefreshSeaUnitMovement(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].Advances = []string{"ADVANCE_COUNTER"}
	state.Units = []Unit{{ID: "ship", OwnerID: "player-a", TileID: "tile-a", Domain: DomainSea, Movement: 0, MaxMovement: 3}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}}
	state.Wonders = []Wonder{{ID: "WONDER_LIGHTHOUSE", OwnerID: "player-a", CityID: "city-a", Scope: "nation", ObsoletedBy: "ADVANCE_COUNTER", Effects: []Effect{{ID: "boat", Kind: EffectIncreaseBoatMovement, Value: 2}}}}
	next := runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	unit, ok := unitByID(next, "ship")
	if !ok || unit.Movement != 3 {
		t.Fatalf("obsolete sea movement wonder refresh = %#v, want max movement only", unit)
	}
}

func TestRetailBoatMovementWonderRefreshesSeaUnitMovement(t *testing.T) {
	state := runtimeStartedState(t)
	state.Units = []Unit{{ID: "ship", OwnerID: "player-a", TileID: "tile-a", Domain: DomainSea, Movement: 0, MaxMovement: 3}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}}
	state.Wonders = []Wonder{{ID: "WONDER_LIGHTHOUSE", OwnerID: "player-a", CityID: "city-a", Scope: "nation", Effects: []Effect{{ID: "boat", Kind: EffectIncreaseBoatMovement, Value: 2}}}}
	next := runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	unit, ok := unitByID(next, "ship")
	if !ok || unit.Movement != 5 {
		t.Fatalf("sea unit movement refresh = %#v, want max movement plus wonder bonus", unit)
	}
}

func TestTask17RoundRefreshesMovementFromPinnedRuntime(t *testing.T) {
	state := runtimeStartedState(t)
	state.Units = []Unit{{ID: "scout", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 0}}
	next := runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	unit, ok := unitByID(next, "scout")
	if !ok || unit.Movement != 2 {
		t.Fatalf("round movement refresh = %#v, want pinned runtime allowance 2", unit)
	}
}
