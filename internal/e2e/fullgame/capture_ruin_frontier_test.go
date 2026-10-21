package fullgame

import (
	"fmt"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestCaptureDiscoveryBoundaryAvoidsCirclingKnownMapEdge(t *testing.T) {
	projection := visibility.Projection{World: &visibility.WorldProjection{Width: 4, Height: 4}, Units: []visibility.UnitProjection{{ID: "reserve", OwnerID: "human", TileID: "0-0", Domain: "land", Movement: 1}}}
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			projection.Tiles = append(projection.Tiles, visibility.TileProjection{ID: fmt.Sprintf("%d-%d", x, y), X: x, Y: y, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}})
		}
	}
	projection.LegalActions = []visibility.ActionProjection{
		{Command: engine.CommandMoveUnit, EntityID: "reserve", Payload: []byte(`{"unit_id":"reserve","path":["0-1"]}`)},
		{Command: engine.CommandMoveUnit, EntityID: "reserve", Payload: []byte(`{"unit_id":"reserve","path":["1-0"]}`)},
	}
	path, ok := captureUnseenFrontierStep(projection, projection.Units[0])
	if !ok || len(path) != 1 || path[0] != "0-1" {
		t.Fatalf("discovery step=%v, want legal step toward unknown southern boundary", path)
	}
	projection.World.Height = 2
	if _, ok := captureUnseenFrontierStep(projection, projection.Units[0]); ok {
		t.Fatal("fully discovered bounds invented a frontier")
	}
}
