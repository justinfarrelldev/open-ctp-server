package visibility

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestAdvertiseLegalActionUsesRetailAdvertiseRoute(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "home", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "foreign-tile", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Cities = []engine.City{{ID: "foreign", OwnerID: "b", TileID: "foreign-tile"}}
	state.Units = []engine.Unit{{ID: "corp", OwnerID: "a", TileID: "home", Movement: 1, Effects: []engine.Effect{{ID: "advertise", Kind: engine.EffectAdvertise}, {ID: "cause", Kind: engine.EffectSpecialCauseUnhappiness}}}}
	var err error
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !hasAction(projection.LegalActions, engine.CommandAdvertise) {
		t.Fatalf("advertise action missing from %#v", projection.LegalActions)
	}
	if hasAction(projection.LegalActions, engine.CommandCauseUnhappiness) {
		t.Fatalf("retail Advertise unit should project advertise route, not generic cause-unhappiness: %#v", projection.LegalActions)
	}
}
