package visibility

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRetailCargoSizeMismatchSuppressesEmbarkAction(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "tile-a", Terrain: engine.TerrainPlains, Open: true}}
	state.Cities = nil
	state.Units = []engine.Unit{
		{ID: "carrier", OwnerID: "a", TileID: "tile-a", Domain: engine.DomainSea, CargoCapacity: 1, CargoDomains: []string{engine.DomainLand}, CargoClasses: []string{"SmallLand"}},
		{ID: "tank", OwnerID: "a", TileID: "tile-a", Domain: engine.DomainLand, Size: "Large"},
	}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandEmbark && action.EntityID == "tank" && action.TargetID == "carrier" {
			t.Fatalf("size-mismatched embark action leaked: %#v", action)
		}
	}
}
