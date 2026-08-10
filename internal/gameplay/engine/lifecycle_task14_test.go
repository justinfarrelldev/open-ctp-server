package engine

import (
	"testing"
)

func TestLifecycleVictoryEvaluatorsEndOnceInManifestOrder(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*State)
		want    string
	}{
		{
			name: "bloodlust conquest",
			prepare: func(state *State) {
				state.Players[1].Lifecycle = PlayerLeft
			},
			want: "bloodlust_conquest",
		},
		{
			name: "world union",
			prepare: func(state *State) {
				state.Nations = []Nation{{ID: "nation-a", Embassies: []string{"nation-b"}}, {ID: "nation-b", Embassies: []string{"nation-a"}}}
			},
			want: "diplomatic_world_union",
		},
		{
			name: "score 2300",
			prepare: func(state *State) {
				state.Phase.Turn = 2300
				state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Effects: []Effect{{ID: "score-a", Kind: EffectVictory, Value: 2}}}, {ID: "city-b", OwnerID: "player-b", TileID: "tile-b", Effects: []Effect{{ID: "score-b", Kind: EffectVictory, Value: 1}}}}
				state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}, {ID: "tile-b", Terrain: TerrainPlains, Open: true}}
			},
			want: "score_2300",
		},
		{
			name: "science gaia controller",
			prepare: func(state *State) {
				state.Nations = []Nation{{ID: "nation-a", Effects: []Effect{{ID: "gaia", Kind: EffectVictory, Value: 1}}}, {ID: "nation-b"}}
			},
			want: "science_gaia_controller",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := fixtureState()
			tt.prepare(&state)
			next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "round", ActorPlayerID: "player-a", Type: "round-complete"}, state.Rules, fixtureSources())
			if err != nil {
				t.Fatal(err)
			}
			if !next.Terminal.Ended || next.Terminal.Reason != tt.want || next.Victory.Reason != tt.want {
				t.Fatalf("terminal victory = %#v / %#v, want %q", next.Terminal, next.Victory, tt.want)
			}
			if len(events) != 1 || events[0].Type != "terminal" {
				t.Fatalf("terminal event = %#v", events)
			}
			if _, retryEvents, _, err := Apply(next, fixtureCommand(), next.Rules, fixtureSources()); err == nil || retryEvents != nil {
				t.Fatal("terminal game accepted a post-terminal command")
			}
		})
	}
}

func TestOwnerEarlyEndAndLeaveTakeoverLifecycle(t *testing.T) {
	state := fixtureState()
	state.Phase.ResolutionLocked = true

	left, leaveEvents, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "leave", ActorPlayerID: "player-a", Type: "leave-game"}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	leftPlayer, ok := playerByID(&left, "player-a")
	if !ok || leftPlayer.Controller != ControllerHuman || !leftPlayer.TakeoverPending || len(leaveEvents) != 1 || leaveEvents[0].Type != "player.disconnected" {
		t.Fatalf("leave must schedule, not apply, takeover: %#v %#v", leftPlayer, leaveEvents)
	}
	if _, _, _, err := Apply(left, Command{SchemaVersion: CommandSchemaV1, ID: "locked", ActorPlayerID: "player-a", Type: "takeover-ai"}, left.Rules, fixtureSources()); err == nil {
		t.Fatal("takeover was allowed before the policy boundary")
	}

	left.Phase.ResolutionLocked = false
	takenOver, takeoverEvents, _, err := Apply(left, Command{SchemaVersion: CommandSchemaV1, ID: "takeover", ActorPlayerID: "player-a", Type: "takeover-ai"}, left.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	takenOverPlayer, ok := playerByID(&takenOver, "player-a")
	if !ok || takenOverPlayer.Controller != ControllerAI || takenOverPlayer.TakeoverPending || len(takeoverEvents) != 1 || takeoverEvents[0].Type != "player.controller_changed" {
		t.Fatalf("takeover did not change the persisted controller: %#v %#v", takenOverPlayer, takeoverEvents)
	}

	ended, events, _, err := Apply(takenOver, Command{SchemaVersion: CommandSchemaV1, ID: "end", ActorPlayerID: "player-a", Type: "end-game"}, takenOver.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if !ended.Terminal.Ended || ended.Terminal.Reason != "owner_ended" || ended.Terminal.WinningPlayerID != "player-a" || len(events) != 1 || events[0].Type != "terminal" {
		t.Fatalf("owner early end was not singular and durable: %#v %#v", ended.Terminal, events)
	}
	if ended.Victory.Reason != "" || ended.Victory.WinnerNationID != "" {
		t.Fatalf("owner early end must not fabricate a natural victory: %#v", ended.Victory)
	}
}

func TestAllDeparturesScheduleOneTakeoverAndRemainReplaySafe(t *testing.T) {
	state := fixtureState()
	state.Phase.ResolutionLocked = true
	for _, playerID := range []string{"player-a", "player-b"} {
		next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "leave-" + playerID, ActorPlayerID: playerID, Type: CommandLeaveGame}, state.Rules, fixtureSources())
		if err != nil || len(events) != 1 || events[0].Type != EventPlayerDisconnected {
			t.Fatalf("leave %q = %#v, %#v, %v", playerID, next, events, err)
		}
		state = next
	}
	if _, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "leave-again", ActorPlayerID: "player-a", Type: CommandLeaveGame}, state.Rules, fixtureSources()); err == nil || events != nil {
		t.Fatal("repeated leave must not create another cleanup event")
	}

	state.Phase.ResolutionLocked = false
	for _, playerID := range []string{"player-a", "player-b"} {
		next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "takeover-" + playerID, ActorPlayerID: playerID, Type: CommandTakeoverAI}, state.Rules, fixtureSources())
		if err != nil || len(events) != 1 || events[0].Type != EventControllerChanged {
			t.Fatalf("takeover %q = %#v, %#v, %v", playerID, next, events, err)
		}
		state = next
	}
	for _, player := range state.Players {
		if player.Controller != ControllerAI || player.TakeoverPending {
			t.Fatalf("all departed players must be AI-controlled exactly once: %#v", state.Players)
		}
	}
	if _, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "takeover-again", ActorPlayerID: "player-a", Type: CommandTakeoverAI}, state.Rules, fixtureSources()); err == nil || events != nil {
		t.Fatal("repeated takeover must not create another controller event")
	}
}
