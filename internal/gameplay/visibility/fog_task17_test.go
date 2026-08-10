package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestTask17VisibilityRefreshesOnlyAfterLegalSightSourceMoves(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{
		{ID: "t0", X: 0, Y: 0, Terrain: engine.TerrainPlains},
		{ID: "t1", X: 1, Y: 0, Terrain: engine.TerrainPlains},
		{ID: "t2", X: 2, Y: 0, Terrain: engine.TerrainPlains},
		{ID: "t3", X: 3, Y: 0, Terrain: engine.TerrainPlains},
		{ID: "t4", X: 4, Y: 0, Terrain: engine.TerrainPlains},
	}
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "t0"}, {ID: "hidden-target", OwnerID: "b", TileID: "t4"}}
	state.Cities = nil
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	before, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	beforeRaw, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(beforeRaw, []byte("hidden-target")) {
		t.Fatalf("raw projection revealed an unseen target: %s", beforeRaw)
	}
	state.Units[0].TileID = "t2"
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	after, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, unit := range after.Units {
		if unit.ID == "hidden-target" {
			found = true
		}
	}
	if !found {
		t.Fatalf("legally visible target was absent after moving sight source: %#v", after)
	}
}
