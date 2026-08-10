package engine

import "testing"

func TestTask17RoundRefreshesMovementFromPinnedRuntime(t *testing.T) {
	state := runtimeStartedState(t)
	state.Units = []Unit{{ID: "scout", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 0}}
	next := runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	unit, ok := unitByID(next, "scout")
	if !ok || unit.Movement != 2 {
		t.Fatalf("round movement refresh = %#v, want pinned runtime allowance 2", unit)
	}
}
