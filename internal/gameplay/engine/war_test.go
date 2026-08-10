package engine

import (
	"encoding/json"
	"testing"
)

func TestSetWarPersistsNormalizedCanonicalEdge(t *testing.T) {
	state := fixtureState()
	state.Phase.Mode = "hybrid"
	state.Players = append(state.Players, Player{ID: "player-c", NationID: "nation-c", Lifecycle: PlayerActive, Controller: ControllerHuman})
	state.Phase.ActivePlayerOrder = append(state.Phase.ActivePlayerOrder, "player-c")
	command := Command{SchemaVersion: CommandSchemaV1, ID: "war-a-b", ActorPlayerID: "player-b", Type: CommandSetWar, Payload: json.RawMessage(`{"target_player_id":"player-a","at_war":true}`)}
	next, events, _, err := Apply(state, command, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Policy.Wars) != 1 || next.Policy.Wars[0] != (TurnPolicyWar{A: "player-a", B: "player-b"}) || len(events) != 1 || events[0].Type != EventWarStateChanged {
		t.Fatalf("war edge/event = %#v %#v", next.Policy.Wars, events)
	}
	command.ID = "peace-a-b"
	command.Payload = json.RawMessage(`{"target_player_id":"player-a","at_war":false}`)
	next, _, _, err = Apply(next, command, next.Rules, fixtureSources())
	if err != nil || len(next.Policy.Wars) != 0 {
		t.Fatalf("war removal = %#v %v", next.Policy.Wars, err)
	}
}

func TestSetWarRejectsInvalidTarget(t *testing.T) {
	state := fixtureState()
	_, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bad-war", ActorPlayerID: "player-a", Type: CommandSetWar, Payload: json.RawMessage(`{"target_player_id":"missing","at_war":true}`)}, state.Rules, fixtureSources())
	if err == nil {
		t.Fatal("invalid war target accepted")
	}
}
