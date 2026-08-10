package visibility

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestTask11CompletedTerraformChangesLOSWithoutLeakingQueuedOverlay(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "origin", X: 0, Y: 0, Terrain: engine.TerrainPlains}, {ID: "hidden-terraform-token", X: 1, Y: 0, Terrain: engine.TerrainForest}, {ID: "beyond", X: 2, Y: 0, Terrain: engine.TerrainPlains}}
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "origin"}, {ID: "SECRET_ACTOR", OwnerID: "b", TileID: "beyond"}}
	state.Cities = nil
	state.Overlays = []engine.TileOverlay{{ID: "terraform:hidden-terraform-token", TileID: "hidden-terraform-token", Kind: "terraform", OwnerID: "b", Remaining: 1, TargetTerrain: engine.TerrainPlains}}
	queued, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, tile := range queued.Tiles {
		if tile.ID == "beyond" {
			t.Fatal("queued overlay changed LOS early")
		}
	}
	if raw := string(mustProjectionJSON(t, queued)); strings.Contains(raw, "SECRET_ACTOR") {
		t.Fatalf("queued projection leaked: %s", raw)
	}

	state.Overlays[0].Remaining = 0
	complete, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tile := range complete.Tiles {
		if tile.ID == "beyond" {
			found = true
		}
	}
	if !found {
		t.Fatal("completed overlay did not update LOS")
	}
	ai, err := AIInput(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if string(ai) != string(mustProjectionJSON(t, complete)) {
		t.Fatal("AI projection differs from player projection")
	}
}

func mustProjectionJSON(t *testing.T, projection Projection) []byte {
	t.Helper()
	projected, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	return projected
}
