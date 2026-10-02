package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func mustPayload(t *testing.T, payload any) json.RawMessage {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestSpaceLaunchLeavesMapAndReentersOwnedCity(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "origin", OwnerID: "player-a", TileID: "tile-a"}, {ID: "destination", OwnerID: "player-a", TileID: "tile-b"}}
	state.Units = []Unit{{ID: "space-plane", OwnerID: "player-a", TileID: "tile-a", Domain: DomainAir, Movement: 100, MaxMovement: 100, Health: 1, MaxHealth: 1, Effects: []Effect{{ID: "space", Kind: EffectSpaceLaunchTurns, Value: 2}}}}

	eventType, err := applyUnitCommand(&state, Command{Type: CommandSpaceLaunch, ActorPlayerID: "player-a", Payload: mustPayload(t, spaceLaunchPayload{UnitID: "space-plane", TargetTileID: "tile-b"})}, nil)
	if err != nil {
		t.Fatalf("space launch: %v", err)
	}
	if eventType != EventUnitLaunched {
		t.Fatalf("event type = %q, want %q", eventType, EventUnitLaunched)
	}
	unit := state.Units[0]
	if !unit.InSpace || unit.TileID != "" || unit.ReentryTileID != "tile-b" || unit.ReentryTurns != 2 {
		t.Fatalf("unit after launch = %#v, want in-space reentry to tile-b in 2 turns", unit)
	}

	runtime := &rules.RuntimeRules{Starter: rules.RuntimeStarter{Movement: 100}}
	advanceRound(&state, runtime)
	if unit := state.Units[0]; !unit.InSpace || unit.ReentryTurns != 1 || unit.TileID != "" {
		t.Fatalf("unit after first reentry tick = %#v, want still in space with 1 turn", unit)
	}
	advanceRound(&state, runtime)
	if unit := state.Units[0]; unit.InSpace || unit.TileID != "tile-b" || unit.ReentryTileID != "" || unit.ReentryTurns != 0 {
		t.Fatalf("unit after reentry = %#v, want returned to owned destination city", unit)
	}
}

func TestSpaceLaunchKillsUnitWhenDestinationCityIsLost(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "origin", OwnerID: "player-a", TileID: "tile-a"}, {ID: "destination", OwnerID: "player-a", TileID: "tile-b"}}
	state.Units = []Unit{{ID: "space-plane", OwnerID: "player-a", TileID: "tile-a", Domain: DomainAir, Movement: 100, MaxMovement: 100, Health: 1, MaxHealth: 1, Effects: []Effect{{ID: "space", Kind: EffectSpaceLaunchTurns, Value: 1}}}}

	if _, err := applyUnitCommand(&state, Command{Type: CommandSpaceLaunch, ActorPlayerID: "player-a", Payload: mustPayload(t, spaceLaunchPayload{UnitID: "space-plane", TargetTileID: "tile-b"})}, nil); err != nil {
		t.Fatalf("space launch: %v", err)
	}
	state.Cities[1].OwnerID = "player-b"
	advanceRound(&state, &rules.RuntimeRules{Starter: rules.RuntimeStarter{Movement: 100}})
	if _, ok := unitByID(state, "space-plane"); ok {
		t.Fatalf("space unit survived after destination city was lost: %#v", state.Units)
	}
}

func TestSpaceLaunchRequiresOwnedDestinationCity(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "origin", OwnerID: "player-a", TileID: "tile-a"}, {ID: "foreign", OwnerID: "player-b", TileID: "tile-b"}}
	state.Units = []Unit{{ID: "space-plane", OwnerID: "player-a", TileID: "tile-a", Domain: DomainAir, Movement: 100, MaxMovement: 100, Health: 1, MaxHealth: 1, Effects: []Effect{{ID: "space", Kind: EffectSpaceLaunchTurns, Value: 2}}}}

	if _, err := applyUnitCommand(&state, Command{Type: CommandSpaceLaunch, ActorPlayerID: "player-a", Payload: mustPayload(t, spaceLaunchPayload{UnitID: "space-plane", TargetTileID: "tile-b"})}, nil); err == nil {
		t.Fatal("space launch succeeded without an owned destination city")
	}
}
