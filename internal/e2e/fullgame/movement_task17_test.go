package fullgame

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestTask17FullgamePlansVisibleRouteWithoutCanonicalState(t *testing.T) {
	projection, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "start", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "target", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units: []visibility.UnitProjection{{ID: "settler", OwnerID: "account-1", TileID: "start", Domain: "land", Movement: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	path, err := PlanVisibleRoute(projection, "settler", "target")
	if err != nil {
		t.Fatal(err)
	}
	if len(path) != 1 || path[0] != "target" {
		t.Fatalf("unexpected planned path: %#v", path)
	}
}
