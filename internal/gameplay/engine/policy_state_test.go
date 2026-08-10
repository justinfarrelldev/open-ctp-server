package engine

import "testing"

func TestCanonicalStatePersistsTurnPolicyState(t *testing.T) {
	state := fixtureState()
	state.Policy = TurnPolicyState{
		Stage:          "planning",
		EarlyEnds:      []string{"player-a"},
		Queue:          []TurnPolicyCommand{{Kind: "action", ActorID: "player-b", UnitOrArmyID: "unit-b", ReceiptSequence: 2, CommandID: "command-b", IdempotencyKey: "receipt-b", Payload: []byte(`{"unit_id":"unit-b"}`)}},
		Wars:           []TurnPolicyWar{{A: "player-a", B: "player-b"}},
		Deadline:       &TurnPolicyDeadline{PlayerID: "player-a", AtUnix: 60},
		Components:     [][]string{{"player-a", "player-b"}},
		ComponentIndex: 0,
	}

	canonical, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCanonicalStateJSON(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Policy.Queue) != 1 || decoded.Policy.Queue[0].ActorID != "player-b" || decoded.Policy.Queue[0].CommandID != "command-b" || decoded.Policy.Queue[0].IdempotencyKey != "receipt-b" || string(decoded.Policy.Queue[0].Payload) != `{"unit_id":"unit-b"}` || decoded.Policy.Deadline == nil || decoded.Policy.Deadline.PlayerID != "player-a" || len(decoded.Policy.Components) != 1 || len(decoded.Policy.Wars) != 1 {
		t.Fatalf("policy state did not round-trip: %#v", decoded.Policy)
	}
	resolver := decoded.AsTurnPolicyState()
	if len(resolver.Queue) != 1 || resolver.Queue[0].ActorID != "player-b" || resolver.Queue[0].CommandID != "command-b" || resolver.Queue[0].IdempotencyKey != "receipt-b" || string(resolver.Queue[0].Payload) != `{"unit_id":"unit-b"}` || resolver.Deadline == nil || resolver.Deadline.PlayerID != "player-a" || len(resolver.Components) != 1 || len(resolver.Wars) != 1 {
		t.Fatalf("canonical policy state did not rebuild resolver state: %#v", resolver)
	}
}

func TestCanonicalStateRejectsInvalidTurnPolicyState(t *testing.T) {
	state := fixtureState()
	state.Policy.Queue = []TurnPolicyCommand{{Kind: "action", ActorID: "missing"}}
	if _, err := CanonicalStateJSON(state); err == nil {
		t.Fatal("canonical state accepted queued action from an unknown player")
	}

	state = fixtureState()
	state.Policy.Components = [][]string{{"player-a", "missing"}}
	if _, err := CanonicalStateJSON(state); err == nil {
		t.Fatal("canonical state accepted a conflict component with an unknown player")
	}

	state = fixtureState()
	state.Policy.Wars = []TurnPolicyWar{{A: "player-a", B: "missing"}}
	if _, err := CanonicalStateJSON(state); err == nil {
		t.Fatal("canonical state accepted a war with an unknown player")
	}
}
