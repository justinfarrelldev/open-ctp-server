package visibility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func clientMetadataState() engine.State {
	state := fixtureState()
	state.Tiles = nil
	for x := 0; x < 12; x++ {
		state.Tiles = append(state.Tiles, engine.Tile{ID: fmt.Sprintf("t%d", x), X: x, Terrain: engine.TerrainPlains, Open: true})
	}
	state.Tiles[1].River = true
	state.Tiles[1].GoodID = "TERRAIN_PLAINS_GOOD_TWO"
	state.Tiles[1].OwnerID = "b"
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "t0"}}
	state.Cities = []engine.City{{ID: "city", Name: "Memphis", TypeID: "CITY_GENERIC", OwnerID: "b", TileID: "t1", Population: 5}}
	return state
}

func metadataWire(t *testing.T, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	return wire
}

func metadataTile(t *testing.T, projection Projection, id string) TileProjection {
	t.Helper()
	for _, tile := range projection.Tiles {
		if tile.ID == id {
			return tile
		}
	}
	t.Fatalf("tile %s absent: %+v", id, projection.Tiles)
	return TileProjection{}
}

func TestClientMetadataVisibleRememberedHiddenChangeAndReobservation(t *testing.T) {
	state := clientMetadataState()
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"river": true, "good_id": "TERRAIN_PLAINS_GOOD_TWO", "owner_id": "b"}
	assertFacts := func(projection Projection, stale bool, name string, population int) {
		t.Helper()
		wire := metadataWire(t, metadataTile(t, projection, "t1"))
		for key, value := range want {
			if wire[key] != value {
				t.Errorf("%s=%v want %v (stale=%v)", key, wire[key], value, stale)
			}
		}
		if wire["stale"] != stale {
			t.Errorf("stale=%v want=%v", wire["stale"], stale)
		}
		if len(projection.Cities) != 1 {
			t.Fatalf("city count=%d", len(projection.Cities))
		}
		city := metadataWire(t, projection.Cities[0])
		if city["name"] != name || city["type_id"] != "CITY_GENERIC" || city["population"] != float64(population) {
			t.Errorf("city metadata missing: %v", city)
		}
		for _, private := range []string{"queue", "stored_production", "specialists", "yields"} {
			if _, ok := city[private]; ok {
				t.Errorf("foreign city leaked %s", private)
			}
		}
	}
	assertFacts(projection, false, "Memphis", 5)
	plain := metadataWire(t, metadataTile(t, projection, "t0"))
	if plain["river"] != false || plain["good_id"] != "" || plain["owner_id"] != "" {
		t.Errorf("known negatives not explicit: %v", plain)
	}
	state.Units[0].TileID = "t10"
	state.Phase.Turn++
	state.Tiles[1].River = false
	state.Tiles[1].GoodID = "secret-good"
	state.Tiles[1].OwnerID = "a"
	state.Cities[0].Name = "SECRET_RENAME"
	state.Cities[0].Population = 9
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	state, err = engine.DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	assertFacts(projection, true, "Memphis", 5)
	if tile := metadataTile(t, projection, "t1"); tile.Observed != 4 || tile.Planning != nil {
		t.Errorf("stale tile freshness/planning leak: %+v", tile)
	}
	state.Units[0].TileID = "t0"
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	want = map[string]any{"river": false, "good_id": "secret-good", "owner_id": "a"}
	assertFacts(projection, false, "SECRET_RENAME", 9)
}

func TestClientMetadataNeverDiscoveredNeighborAndPlayerIsolation(t *testing.T) {
	state := clientMetadataState()
	state.Cities = nil
	state.Tiles[3].River = true
	state.Tiles[3].GoodID = "UNDISCOVERED_GOOD"
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	for _, viewer := range []string{"a", "b"} {
		projection, err := Project(state, viewer)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(projection)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte("UNDISCOVERED_GOOD")) {
			t.Fatalf("never-discovered neighbor exposed to %s", viewer)
		}
		if viewer == "b" && len(projection.Tiles) != 0 {
			t.Fatalf("other player's observations leaked: %s", raw)
		}
	}
}

func TestClientMetadataLegacyObservationsStayUnknown(t *testing.T) {
	state := clientMetadataState()
	state.Units[0].TileID = "t10"
	state.Fog = []engine.FogObservation{{PlayerID: "a", TileID: "t1", Turn: 1, Terrain: engine.TerrainPlains, CityID: "city", CityOwnerID: "b"}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	wire := metadataWire(t, metadataTile(t, projection, "t1"))
	for _, key := range []string{"river", "good_id", "owner_id", "yields", "movement_cost", "overlays"} {
		if value, exists := wire[key]; exists {
			t.Errorf("legacy %s backfilled from hidden state: %v", key, value)
		}
	}
	city := metadataWire(t, projection.Cities[0])
	for _, key := range []string{"name", "type_id", "population"} {
		if value, exists := city[key]; exists {
			t.Errorf("legacy city %s backfilled: %v", key, value)
		}
	}
}

func TestClientTradeAccountingOnlyEndpointOwners(t *testing.T) {
	state := clientMetadataState()
	state.Units = append(state.Units, engine.Unit{ID: "observer", OwnerID: "b", TileID: "t0"})
	state.Cities = []engine.City{{ID: "source", OwnerID: "a", TileID: "t0"}, {ID: "destination", OwnerID: "a", TileID: "t1"}}
	state.TradeRoutes = []engine.TradeRoute{{ID: "route", SourceCityID: "source", DestinationCityID: "destination", GoodID: "elephant", Gold: 0}}
	owner, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(owner.TradeRoutes) != 1 {
		t.Fatal("owner route absent")
	}
	wire := metadataWire(t, owner.TradeRoutes[0])
	if wire["good_id"] != "elephant" || wire["gold"] != float64(0) {
		t.Errorf("missing route accounting/explicit zero: %v", wire)
	}
	observer, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range observer.TradeRoutes {
		wire := metadataWire(t, route)
		for _, key := range []string{"good_id", "gold"} {
			if _, exists := wire[key]; exists {
				t.Errorf("observer saw private accounting %v", wire)
			}
		}
	}
	state.Cities[1].TileID = "t10"
	hidden, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(hidden.TradeRoutes, []TradeRouteProjection{}) {
		t.Errorf("hidden endpoints exposed: %+v", hidden.TradeRoutes)
	}
}
