package game

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/world"
)

func TestRetailWorldHandoffKeepsEveryGeneratedTerrainFamily(t *testing.T) {
	runtime := rules.RuntimeRules{Movement: rules.RuntimeMovement{Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_DESERT"}, {ID: "TERRAIN_MOUNTAIN"}, {ID: "TERRAIN_PLAINS"}}}}
	for terrain, want := range map[string]string{"desert": "TERRAIN_DESERT", "mountain": "TERRAIN_MOUNTAIN"} {
		if got := engineTerrainForRuntime(terrain, runtime); got != want {
			t.Errorf("%s flattened to %s, want %s", terrain, got, want)
		}
	}
}

func TestWorldStartPreservesIndependentRiverAndGoods(t *testing.T) {
	var tile world.WorldTile
	if err := json.Unmarshal([]byte(`{"x":0,"y":0,"terrain":"plains","river":true,"good_id":"TERRAIN_PLAINS_GOOD_TWO"}`), &tile); err != nil {
		t.Fatal(err)
	}
	pin, err := serverRuntimePin()
	if err != nil {
		t.Fatal(err)
	}
	snapshot := world.WorldSnapshot{Config: world.WorldConfig{StartingSettlers: 1}, Tiles: []world.WorldTile{tile}, Spawns: []world.PlayerSpawn{{AccountID: 7}}}
	state, err := initialGameplayState(snapshot, "hybrid", pin, []GamePlayer{{AccountID: 7}})
	if err != nil {
		t.Fatal(err)
	}
	if !state.Tiles[0].River || state.Tiles[0].GoodID != "TERRAIN_PLAINS_GOOD_TWO" {
		t.Fatalf("world metadata lost: %+v", state.Tiles[0])
	}
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := engine.DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Tiles[0] != state.Tiles[0] {
		t.Fatal("persistence lost tile metadata")
	}
}
