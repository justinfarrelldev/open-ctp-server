package engine

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestClientPlanningHistoryLabelsTheCompletedRound(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			state := task10State()
			state.Phase.Mode = mode
			state.Nations = []Nation{{ID: "nation-a"}, {ID: "nation-b"}}
			next, _, err := resolvePolicyReplay(t, state)
			if err != nil {
				t.Fatal(err)
			}
			if next.Phase.Turn != 2 || len(next.Nations[0].Statistics) != 1 || next.Nations[0].Statistics[0].Turn != 1 {
				t.Fatal("completed-round sample was mislabeled as next round")
			}
		})
	}
}

func TestClientPlanningTerminalStopsRemainingOrdersAndEconomy(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			state := task8State()
			state.Phase.Mode = mode
			state.Nations = []Nation{{ID: "nation-a"}, {ID: "nation-b"}}
			for _, command := range []Command{{SchemaVersion: CommandSchemaV1, ID: "finish", ActorPlayerID: "player-a", Type: CommandEndGame, Payload: json.RawMessage(`{}`)}, {SchemaVersion: CommandSchemaV1, ID: "sleep", ActorPlayerID: "player-a", Type: CommandSleep, Payload: json.RawMessage(`{"unit_id":"land"}`)}} {
				state, _ = applyPolicyTestCommand(t, state, command)
			}
			next, _, err := resolvePolicyReplay(t, state)
			if err != nil {
				t.Fatal(err)
			}
			unit, _ := unitByID(next, "land")
			if !next.Terminal.Ended || !next.Phase.ResolutionLocked || unit.Asleep || next.RecalculatedTurn != 0 || len(next.Nations[0].Statistics) != 0 {
				t.Fatal("terminal resolution continued executing or recalculating")
			}
		})
	}
}

func TestClientPlanningRejectsInvalidAtomicEditsBeforePersistence(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			state := task10State()
			state.Phase.Mode = mode
			state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2}}
			before, _ := CanonicalStateJSON(state)
			command := Command{SchemaVersion: CommandSchemaV1, ID: "bad-edit", ActorPlayerID: "player-a", Type: CommandEditProduction, Payload: json.RawMessage(`{"city_id":"home","entries":[{"entry_id":"absent"}]}`)}
			if _, _, _, err := Apply(state, command, state.Rules, fixtureSources()); err == nil {
				t.Fatal("invalid edit entered the authoritative planning queue")
			}
			after, _ := CanonicalStateJSON(state)
			if !bytes.Equal(before, after) {
				t.Fatal("rejection mutated caller state")
			}
		})
	}
}

func TestClientSecondPlannerCanSubmitCoordinateMovement(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			state := task8State()
			state.Phase.Mode = mode
			for index := range state.Tiles {
				state.Tiles[index].OwnerID = ""
				state.Tiles[index].Open = true
			}
			command := Command{SchemaVersion: CommandSchemaV1, ID: "second-move", ActorPlayerID: "player-b", Type: CommandMoveUnit, Payload: json.RawMessage(`{"unit_id":"enemy","target_x":2,"target_y":1}`)}
			next, _, _, err := Apply(state, command, state.Rules, fixtureSources())
			if err != nil {
				t.Fatal(err)
			}
			unit, _ := unitByID(next, "enemy")
			if unit.TileID != "d" || len(next.Policy.Queue) != 1 {
				t.Fatal("planning must validate and queue, not execute movement")
			}
		})
	}
}

func TestClientValidPlanningOrdersInvalidatedByResolutionRejectPrivatelyWithoutStalling(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			state := task8State()
			state.Phase.Mode = mode
			for _, kind := range []string{CommandDisband, CommandSleep} {
				command := Command{SchemaVersion: CommandSchemaV1, ID: kind, ActorPlayerID: "player-a", Type: kind, Payload: json.RawMessage(`{"unit_id":"land"}`)}
				state, _ = applyPolicyTestCommand(t, state, command)
			}
			state, _ = applyPolicyTestCommand(t, state, Command{SchemaVersion: CommandSchemaV1, ID: "end-a", ActorPlayerID: "player-a", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)})
			restarted, err := canonicalClone(state)
			if err != nil {
				t.Fatal(err)
			}
			resolve := func(state State) State {
				next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "end-b", ActorPlayerID: "player-b", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)}, state.Rules, fixtureSources())
				if err != nil {
					t.Fatal(err)
				}
				return next
			}
			resolved := resolve(state)
			again := resolve(restarted)
			left, _ := CanonicalStateJSON(resolved)
			right, _ := CanonicalStateJSON(again)
			if !bytes.Equal(left, right) {
				t.Fatal("resolution changed after restart")
			}
			_, present := unitByID(resolved, "land")
			if present || resolved.Phase.Turn != 2 || len(resolved.Policy.Queue) != 0 {
				t.Fatal("stale order stalled or corrupted resolution")
			}
			if len(resolved.Notifications) != 1 || resolved.Notifications[0].Kind != "order.rejected" || resolved.Notifications[0].PlayerID != "player-a" {
				t.Fatalf("missing private rejection result: %+v", resolved.Notifications)
			}
		})
	}
}
