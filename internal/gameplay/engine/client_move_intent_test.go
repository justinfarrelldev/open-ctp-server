package engine

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestClientRejectedCoordinateIntentCannotPoisonHybridPlanning(t *testing.T) {
	state := task8State()
	state.Phase.Mode = "hybrid"
	command := task8Command(t, CommandMoveUnit, map[string]any{"unit_id": "land", "target_x": 99, "target_y": 99})
	if _, _, _, err := Apply(state, command, state.Rules, fixtureSources()); err == nil {
		t.Fatal("known-invalid coordinate intent was stored until it could deadlock resolution")
	}
	if len(state.Policy.Queue) != 0 {
		t.Fatal("rejected coordinate intent changed planning state")
	}
}

func TestClientCoordinateMoveDoesNotRequireHiddenTileIdentity(t *testing.T) {
	state := task8State()
	// The player may address an adjacent public coordinate, without receiving the
	// canonical target ID/terrain. Normal movement still decides the outcome.
	command := task8Command(t, CommandMoveUnit, map[string]any{"unit_id": "land", "target_x": 1, "target_y": 0})
	next, _, _, err := Apply(state, command, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	moved, ok := unitByID(next, "land")
	if !ok || moved.TileID != "b" || moved.Movement != 2 {
		t.Fatalf("coordinate intent bypassed/lost execution validation: %+v", moved)
	}
	for _, payload := range []string{
		`{"unit_id":"land","target_x":99,"target_y":99}`,
		`{"unit_id":"land","target_x":1}`,
		`{"unit_id":"land","target_x":1,"target_y":0,"path":["b"]}`,
		`{"unit_id":"land","target_x":null,"target_y":null,"path":["b"]}`,
	} {
		before, err := CanonicalStateJSON(state)
		if err != nil {
			t.Fatal(err)
		}
		command.Payload = json.RawMessage(payload)
		if _, _, _, err := Apply(state, command, state.Rules, fixtureSources()); err == nil {
			t.Fatal("invalid coordinate move accepted")
		}
		after, err := CanonicalStateJSON(state)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Fatal("rejected move changed canonical position")
		}
	}
}
