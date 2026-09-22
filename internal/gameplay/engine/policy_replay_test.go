package engine

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func policyCommand(id, actor string) Command {
	return Command{SchemaVersion: CommandSchemaV1, ID: id, ActorPlayerID: actor, Type: CommandNoop, Payload: json.RawMessage(`{"nested":{"b":2,"a":1}}`)}
}

func applyPolicyTestCommand(t *testing.T, state State, command Command) (State, []Event) {
	t.Helper()
	next, events, _, err := Apply(state, command, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Apply(%s): %v", command.ID, err)
	}
	return next, events
}

func movePolicyCommand(t *testing.T, id, actor, unitID string, path ...string) Command {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"unit_id": unitID, "path": path})
	if err != nil {
		t.Fatal(err)
	}
	return Command{SchemaVersion: CommandSchemaV1, ID: id, ActorPlayerID: actor, Type: CommandMoveUnit, Payload: payload}
}

func resolvePolicyReplay(t *testing.T, state State) (State, []Event, error) {
	t.Helper()
	state, _ = applyPolicyTestCommand(t, state, Command{SchemaVersion: CommandSchemaV1, ID: "early-a", ActorPlayerID: "player-a", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)})
	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "early-b", ActorPlayerID: "player-b", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)}, state.Rules, fixtureSources())
	return next, events, err
}

func applyImmediateMoveProof(t *testing.T, state State, command Command) error {
	t.Helper()
	baseline, err := canonicalClone(state)
	if err != nil {
		t.Fatal(err)
	}
	baseline.Phase.Mode = "sequential"
	baseline.Phase.ResolutionLocked = false
	for index, actorID := range baseline.Phase.ActivePlayerOrder {
		if actorID == command.ActorPlayerID {
			baseline.Phase.ActivePlayerIndex = index
			_, _, _, err := Apply(baseline, command, baseline.Rules, fixtureSources())
			return err
		}
	}
	t.Fatalf("actor %q not found in active order %#v", command.ActorPlayerID, baseline.Phase.ActivePlayerOrder)
	return nil
}

func TestSimultaneousPolicyReplaysCanonicalQueuedEnvelopesInStableOrder(t *testing.T) {
	state := fixtureState()
	state.Phase.Mode = "simultaneous"
	state.Phase.BaseRevision = state.Revision

	state, _ = applyPolicyTestCommand(t, state, policyCommand("queued-b", "player-b"))
	state, _ = applyPolicyTestCommand(t, state, policyCommand("queued-a", "player-a"))
	if len(state.Policy.Queue) != 2 {
		t.Fatalf("queued commands = %#v", state.Policy.Queue)
	}
	var stored Command
	if err := json.Unmarshal(state.Policy.Queue[0].Payload, &stored); err != nil {
		t.Fatal(err)
	}
	expectedStored := policyCommand("queued-b", "player-b")
	expectedStored.IdempotencyKey = expectedStored.ID
	expectedStored.Payload = canonicalPayload(expectedStored.Payload)
	if !reflect.DeepEqual(stored, expectedStored) {
		t.Fatalf("stored envelope = %#v", stored)
	}

	state, _ = applyPolicyTestCommand(t, state, Command{SchemaVersion: CommandSchemaV1, ID: "early-a", ActorPlayerID: "player-a", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)})
	resolved, events := applyPolicyTestCommand(t, state, Command{SchemaVersion: CommandSchemaV1, ID: "early-b", ActorPlayerID: "player-b", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)})

	if resolved.Revision != 4 || resolved.EventSequence != 4 || resolved.Phase.Turn != 2 || resolved.Policy.Recalculations != 1 || len(resolved.Policy.Queue) != 0 {
		t.Fatalf("resolved policy state = %#v", resolved)
	}
	if len(events) != 1 || events[0].CommandID != "early-b" || events[0].Sequence != 4 {
		t.Fatalf("resolution event = %#v", events)
	}
}

func TestSimultaneousPolicyReplaySurvivesCanonicalRestart(t *testing.T) {
	queue := func(t *testing.T) State {
		t.Helper()
		state := fixtureState()
		state.Phase.Mode = "simultaneous"
		state, _ = applyPolicyTestCommand(t, state, policyCommand("queued-b", "player-b"))
		state, _ = applyPolicyTestCommand(t, state, policyCommand("queued-a", "player-a"))
		return state
	}
	resolve := func(t *testing.T, state State) (State, []Event) {
		t.Helper()
		state, _ = applyPolicyTestCommand(t, state, Command{SchemaVersion: CommandSchemaV1, ID: "early-a", ActorPlayerID: "player-a", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)})
		return applyPolicyTestCommand(t, state, Command{SchemaVersion: CommandSchemaV1, ID: "early-b", ActorPlayerID: "player-b", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)})
	}

	uninterrupted, uninterruptedEvents := resolve(t, queue(t))
	encoded, err := CanonicalStateJSON(queue(t))
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := DecodeCanonicalStateJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	afterRestart, restartEvents := resolve(t, restarted)
	if got, want := artifactHash(t, afterRestart, restartEvents), artifactHash(t, uninterrupted, uninterruptedEvents); got != want {
		t.Fatalf("restart replay artifact = %s; want %s", got, want)
	}
}

func TestSimultaneousPolicyResolutionSuppressesDeterministicSameDestinationCollision(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			state := task8State()
			state.Phase.Mode = mode
			state.Phase.BaseRevision = state.Revision
			for index := range state.Tiles {
				if state.Tiles[index].ID == "d" {
					state.Tiles[index].OwnerID = ""
					state.Tiles[index].Open = true
				}
			}
			units := make([]Unit, 0, len(state.Units))
			for _, unit := range state.Units {
				if unit.ID == "enemy" || unit.ID == "cargo" || unit.ID == "ship" {
					continue
				}
				if unit.ID == "land-2" {
					unit.OwnerID = "player-b"
					unit.TileID = "e"
				}
				units = append(units, unit)
			}
			state.Units = units

			if err := applyImmediateMoveProof(t, state, movePolicyCommand(t, "live-b", "player-b", "land-2", "d")); err != nil {
				t.Fatalf("player-b baseline move: %v", err)
			}
			if err := applyImmediateMoveProof(t, state, movePolicyCommand(t, "live-a", "player-a", "land", "d")); err != nil {
				t.Fatalf("player-a baseline move: %v", err)
			}

			state, _ = applyPolicyTestCommand(t, state, movePolicyCommand(t, "queued-b", "player-b", "land-2", "d"))
			state, _ = applyPolicyTestCommand(t, state, movePolicyCommand(t, "queued-a", "player-a", "land", "d"))

			resolved, events, err := resolvePolicyReplay(t, state)
			if err != nil {
				t.Fatalf("resolve replay: %v", err)
			}
			if len(events) != 1 || events[0].Type != "turn.resolved" {
				t.Fatalf("resolution event = %#v", events)
			}
			var playerATile, playerBTile string
			for _, unit := range resolved.Units {
				switch unit.ID {
				case "land":
					playerATile = unit.TileID
				case "land-2":
					playerBTile = unit.TileID
				}
			}
			if playerATile != "d" {
				t.Fatalf("player-a tile = %q, want d", playerATile)
			}
			if playerBTile != "e" {
				t.Fatalf("player-b tile = %q, want e", playerBTile)
			}
		})
	}
}

func TestSimultaneousPolicyResolutionKeepsBaselineInvalidMoveFatal(t *testing.T) {
	state := task8State()
	state.Phase.Mode = "simultaneous"
	state.Phase.BaseRevision = state.Revision
	for index := range state.Units {
		if state.Units[index].ID == "land-2" {
			state.Units[index].OwnerID = "player-b"
			state.Units[index].TileID = "e"
		}
	}

	if err := applyImmediateMoveProof(t, state, movePolicyCommand(t, "baseline-invalid", "player-b", "land-2", "d")); !errors.Is(err, ErrObstructed) {
		t.Fatalf("baseline invalid move error = %v, want %v", err, ErrObstructed)
	}

	state, _ = applyPolicyTestCommand(t, state, policyCommand("queued-a", "player-a"))
	state, _ = applyPolicyTestCommand(t, state, movePolicyCommand(t, "queued-b", "player-b", "land-2", "d"))

	_, _, err := resolvePolicyReplay(t, state)
	if !errors.Is(err, ErrObstructed) {
		t.Fatalf("resolve error = %v, want %v", err, ErrObstructed)
	}
}

func TestSimultaneousPolicyResolutionKeepsNonMovementReplayErrorsFatal(t *testing.T) {
	state := task8State()
	state.Phase.Mode = "simultaneous"
	state.Phase.BaseRevision = state.Revision

	payload := json.RawMessage(`"not-an-object"`)
	state, _ = applyPolicyTestCommand(t, state, Command{SchemaVersion: CommandSchemaV1, ID: "queued-a", ActorPlayerID: "player-a", Type: CommandFoundCity, Payload: payload})
	state, _ = applyPolicyTestCommand(t, state, policyCommand("queued-b", "player-b"))

	_, _, err := resolvePolicyReplay(t, state)
	if err == nil {
		t.Fatal("resolution accepted malformed non-movement payload")
	}
}

func TestSimultaneousPolicyResolutionRollsBackWhenQueuedEnvelopeIsInvalid(t *testing.T) {
	state := fixtureState()
	state.Phase.Mode = "simultaneous"
	state.Policy.Queue = []TurnPolicyCommand{{
		Kind:            "action",
		ActorID:         "player-a",
		ActionClass:     "non_hostile",
		ReceiptSequence: 1,
		CommandID:       "bad-queued",
		IdempotencyKey:  "receipt-bad",
		Payload:         json.RawMessage(`{"schema_version":1,"id":"bad-queued","actor_player_id":"player-a","type":"unknown"}`),
	}}
	state.Policy.EarlyEnds = []string{"player-a"}

	_, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "early-b", ActorPlayerID: "player-b", Type: CommandRoundComplete, Payload: json.RawMessage(`{}`)}, state.Rules, fixtureSources())
	if err == nil {
		t.Fatal("resolution accepted an invalid queued envelope")
	}
	if state.Revision != 0 || len(state.Policy.Queue) != 1 || state.Policy.Queue[0].CommandID != "bad-queued" {
		t.Fatalf("resolution mutated caller state on failure: %#v", state)
	}
}
