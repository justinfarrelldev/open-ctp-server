package engine

import (
	"encoding/json"
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
