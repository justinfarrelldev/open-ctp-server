package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestClientMovementIntentsDoNotDescribeHiddenPassability(t *testing.T) {
	state := inspectionState(t)
	state.Units[0].Domain = engine.DomainLand
	state.Units[0].Movement = 200
	first, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	firstWire := metadataWire(t, first)
	intents, ok := firstWire["action_intents"].([]any)
	if !ok || len(intents) != 2 {
		t.Fatalf("missing fog-safe movement input: %v", firstWire["action_intents"])
	}
	state.Tiles[11].Terrain = engine.TerrainOcean
	state.Tiles[11].OwnerID = "b"
	state.Tiles[11].River = true
	second, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(firstWire["action_intents"])
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(metadataWire(t, second)["action_intents"])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || bytes.Contains(after, []byte("t11")) {
		t.Fatalf("intent leaks hidden legality/IDs: %s", after)
	}
	contract, ok := firstWire["contract"].(map[string]any)
	if !ok || contract["version"] != float64(2) {
		t.Fatalf("missing versioned contract: %v", contract)
	}
	features, ok := contract["features"].(map[string]any)
	if !ok || features["coordinate_move_intents"] != true || features["retained_movement_orders"] != true || features["policy_optimizer"] != false || features["governor"] != false {
		t.Fatalf("unsupported features were not explicit: %v", features)
	}
}
