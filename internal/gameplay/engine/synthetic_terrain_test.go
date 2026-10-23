package engine

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"testing"
)

func TestNewSyntheticTerrainFamiliesSupportEconomyWithoutFlattening(t *testing.T) {
	pin, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := rules.CompileRuntime(pin.CanonicalJSON, pin.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	for _, terrain := range runtime.Movement.Terrains {
		wantCost := 1
		if terrain.ID == TerrainForest {
			wantCost = 2
		}
		if terrain.Cost != wantCost {
			t.Errorf("synthetic %s movement drifted from pre-handoff fixture policy: %d", terrain.ID, terrain.Cost)
		}
		state := task10State()
		state.Tiles[0].Terrain = terrain.ID
		if _, err := terrainYieldWithRuntime(state, "a", &runtime); err != nil {
			t.Errorf("generated terrain %s can't be worked: %v", terrain.ID, err)
		}
	}
}
