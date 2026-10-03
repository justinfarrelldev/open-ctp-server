package engine

import "testing"

func TestRetailCityCaptureDestroysCapturedCapitolBuilding(t *testing.T) {
	state := fixtureState()
	state.Cities = []City{{ID: "capital", OwnerID: "player-b", TileID: "tile-b", Population: 4, Buildings: []Building{
		{ID: "IMPROVE_CAPITOL", Effects: []Effect{{ID: "capitol", Kind: EffectCapitol}}},
		{ID: "IMPROVE_GRANARY", Effects: []Effect{{ID: "food", Kind: EffectYieldFood, Value: 1}}},
	}}}

	captureCityAt(&state, "player-a", "tile-b")
	if cityHasBuilding(state.Cities[0], "IMPROVE_CAPITOL") {
		t.Fatalf("captured city retained capitol building: %+v", state.Cities[0].Buildings)
	}
	if !cityHasBuilding(state.Cities[0], "IMPROVE_GRANARY") {
		t.Fatalf("capture removed non-capitol building: %+v", state.Cities[0].Buildings)
	}
}
