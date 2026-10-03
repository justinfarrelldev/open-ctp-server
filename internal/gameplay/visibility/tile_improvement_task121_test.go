package visibility

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRetailTileImprovementCanSeeUnderwaterDetectsSubmarine(t *testing.T) {
	state := fixtureState()
	state.Units = []engine.Unit{{ID: "sub", OwnerID: "b", TileID: "1,0", Effects: []engine.Effect{{ID: "sub", Kind: engine.EffectSubmarine}}}}
	state.Overlays = []engine.TileOverlay{{ID: "sonar", TileID: "0,0", Kind: "improvement", OwnerID: "a", Remaining: 0, ImprovementID: "TILEIMP_SONAR", VisionRange: 2, CanSee: []string{"Underwater"}}}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, unit := range projection.Units {
		if unit.ID == "sub" {
			return
		}
	}
	t.Fatalf("submarine hidden from underwater detector overlay: %#v", projection.Units)
}
