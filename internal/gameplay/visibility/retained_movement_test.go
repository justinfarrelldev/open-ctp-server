package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRetainedMovementProjectionPrivateAndAvailableWithoutMovement(t *testing.T) {
	state := inspectionState(t)
	unit := state.Units[0]
	state.Units[0].Domain = engine.DomainLand
	state.Units[0].Movement = 0
	state.MovementOrders = []engine.MovementOrder{{UnitID: unit.ID, UnitIDs: []string{unit.ID}, PlayerID: "a", TileID: unit.TileID, TargetX: 10, CommandID: "PRIVATE_JOURNEY"}}
	own, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	wire, _ := json.Marshal(own)
	if !bytes.Contains(wire, []byte("PRIVATE_JOURNEY")) {
		t.Fatal("reconnect lost retained order")
	}
	found, cancel := false, false
	for _, intent := range own.ActionIntents {
		if intent.Command == engine.CommandSetMovementOrder {
			found = true
		}
	}
	for _, action := range own.LegalActions {
		if action.Command == engine.CommandCancelMovementOrder {
			cancel = true
		}
	}
	if !found || !cancel {
		t.Fatal("missing retained set/cancel authorization")
	}
	foreign, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	wire, _ = json.Marshal(foreign)
	if bytes.Contains(wire, []byte("PRIVATE_JOURNEY")) {
		t.Fatal("foreign viewer saw journey")
	}
}
