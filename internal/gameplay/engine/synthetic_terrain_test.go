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
		state := task10State()
		state.Tiles[0].Terrain = terrain.ID
		if _, err := terrainYieldWithRuntime(state, "a", &runtime); err != nil {
			t.Errorf("generated terrain %s can't be worked: %v", terrain.ID, err)
		}
	}
}
