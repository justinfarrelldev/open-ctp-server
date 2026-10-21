package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRuinProjectionRemembersPresenceNotHiddenRollsOrConsumption(t *testing.T) {
	state := inspectionState(t)
	state.Tiles[0].Ruin = &engine.Ruin{Profile: engine.VillageRuinProfile, TypeRoll: 17, ValueRoll: 4321}
	state.Tiles[11].Ruin = &engine.Ruin{Profile: engine.VillageRuinProfile, TypeRoll: 12, ValueRoll: 9876}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	own, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	tile := metadataWire(t, metadataTile(t, own, "t0"))
	if tile["ruin"] != true {
		t.Fatal("visible ruin omitted")
	}
	raw, _ := json.Marshal(own)
	for _, secret := range []string{"type_roll", "value_roll", "villages-only-v1", "9876"} {
		if bytes.Contains(raw, []byte(secret)) {
			t.Fatal("ruin roll leaked")
		}
	}
	state.Units[0].TileID = "t11"
	state.Tiles[0].Ruin = nil
	own, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	tile = metadataWire(t, metadataTile(t, own, "t0"))
	if tile["ruin"] != true || tile["stale"] != true {
		t.Fatal("hidden ruin consumption leaked into memory")
	}
}
