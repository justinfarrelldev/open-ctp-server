package visibility

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRetailInvisibleForGlobeSatUnitHiddenFromGlobalRadarOnly(t *testing.T) {
	state := fixtureState()
	state.Cities = append(state.Cities, engine.City{ID: "radar-city", OwnerID: "a", TileID: "0,0"})
	state.Wonders = []engine.Wonder{{ID: "WONDER_RADAR", OwnerID: "a", CityID: "radar-city", Scope: "nation", Effects: []engine.Effect{{ID: "radar", Kind: engine.EffectGlobalRadar}}}}
	state.Tiles = append(state.Tiles, engine.Tile{ID: "3,0", X: 3, Y: 0, Terrain: engine.TerrainPlains, Open: true})
	state.Units = append(state.Units, engine.Unit{ID: "GLOBE_HIDDEN", OwnerID: "b", TileID: "3,0", Effects: []engine.Effect{{ID: "hidden", Kind: engine.EffectInvisibleForGlobeSat}}})

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projectionHasUnit(projection, "GLOBE_HIDDEN") {
		t.Fatalf("global radar revealed invisible-for-globe-sat unit: %+v", projection.Units)
	}
}

func TestRetailInvisibleForGlobeSatUnitVisibleToNormalSight(t *testing.T) {
	state := fixtureState()
	state.Units = append(state.Units, engine.Unit{ID: "GLOBE_HIDDEN", OwnerID: "b", TileID: "1,0", Effects: []engine.Effect{{ID: "hidden", Kind: engine.EffectInvisibleForGlobeSat}}})

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !projectionHasUnit(projection, "GLOBE_HIDDEN") {
		t.Fatalf("normal sight did not reveal invisible-for-globe-sat unit: %+v", projection.Units)
	}
}
