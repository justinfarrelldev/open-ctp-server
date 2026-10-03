package visibility

import (
	"bytes"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRetailBombardActionsIncludeVisibleUndefendedEnemyCities(t *testing.T) {
	projection := Projection{
		Tiles:  []TileProjection{{ID: "origin", X: 0, Y: 0}, {ID: "target", X: 1, Y: 0}},
		Units:  []UnitProjection{{ID: "artillery", OwnerID: "player-a", TileID: "origin", Domain: engine.DomainLand, Movement: 1, Ranged: 10, BombardRange: 1, BombRounds: 1}},
		Cities: []CityProjection{{ID: "city-b", OwnerID: "player-b", TileID: "target", Population: 2}},
	}
	tiles := map[string]TileProjection{"origin": projection.Tiles[0], "target": projection.Tiles[1]}

	actions := bombardActions(projection, tiles, projection.Units[0])
	for _, action := range actions {
		if action.Command == engine.CommandBombard && action.TargetID == "target" && bytes.Contains(action.Payload, []byte(`"target_tile_id":"target"`)) {
			return
		}
	}
	t.Fatalf("city bombard action missing: %#v", actions)
}

func TestRetailBombardActionsSuppressStaleCityTargets(t *testing.T) {
	projection := Projection{
		Tiles:  []TileProjection{{ID: "origin", X: 0, Y: 0}, {ID: "target", X: 1, Y: 0, Stale: true}},
		Units:  []UnitProjection{{ID: "artillery", OwnerID: "player-a", TileID: "origin", Domain: engine.DomainLand, Movement: 1, Ranged: 10, BombardRange: 1, BombRounds: 1}},
		Cities: []CityProjection{{ID: "city-b", OwnerID: "player-b", TileID: "target", Population: 2, Stale: true}},
	}
	tiles := map[string]TileProjection{"origin": projection.Tiles[0], "target": projection.Tiles[1]}

	if actions := bombardActions(projection, tiles, projection.Units[0]); len(actions) != 0 {
		t.Fatalf("stale city bombard action leaked: %#v", actions)
	}
}
