package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestCityInfluenceIsAnObservedFactWithFogMemoryAndExplicitEmpty(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "seen", Terrain: engine.TerrainPlains, OwnerID: "a", InfluenceCityID: "old-city", Open: true}, {ID: "empty", X: 1, Terrain: engine.TerrainPlains, Open: true}, {ID: "far", X: 100, Terrain: engine.TerrainPlains, Open: true}, {ID: "never", X: 200, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "seen", Domain: engine.DomainLand}}
	state.Cities = []engine.City{{ID: "old-city", Name: "Known", OwnerID: "a", TileID: "seen", Population: 1}}
	observed, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Project(observed, "a")
	if err != nil {
		t.Fatal(err)
	}
	assert := func(p Projection, id, want string, stale bool) {
		t.Helper()
		for _, tile := range p.Tiles {
			if tile.ID == id {
				if tile.InfluenceCityID == nil || *tile.InfluenceCityID != want || tile.Stale != stale {
					t.Fatalf("tile influence %s = %+v", id, tile)
				}
				return
			}
		}
		t.Fatalf("missing tile %s", id)
	}
	assert(p, "seen", "old-city", false)
	assert(p, "empty", "", false)
	observed.Units[0].TileID = "far"
	observed.Cities[0].ID = "HIDDEN_NEW_CITY"
	observed.Cities[0].Name = "HIDDEN_NEW_NAME"
	observed.Cities[0].OwnerID = "b"
	observed.Tiles[0].OwnerID = "b"
	observed.Tiles[0].InfluenceCityID = "HIDDEN_NEW_CITY"
	observed.Phase.Turn++
	before, _ := engine.CanonicalStateJSON(observed)
	p, err = Project(observed, "a")
	if err != nil {
		t.Fatal(err)
	}
	assert(p, "seen", "old-city", true)
	raw, _ := json.Marshal(p)
	if bytes.Contains(raw, []byte("HIDDEN_NEW")) {
		t.Fatalf("hidden live influence/city leaked: %s", raw)
	}
	for _, tile := range p.Tiles {
		if tile.ID == "never" {
			t.Fatal("unobserved assignment exposed")
		}
	}
	after, _ := engine.CanonicalStateJSON(observed)
	if !bytes.Equal(before, after) {
		t.Fatal("read-only projection mutated influence")
	}
	restored, err := engine.DecodeCanonicalStateJSON(before)
	if err != nil {
		t.Fatal(err)
	}
	p, err = Project(restored, "a")
	if err != nil {
		t.Fatal(err)
	}
	assert(p, "seen", "old-city", true)
	restored.Units[0].TileID = "seen"
	for n := range restored.Tiles {
		if restored.Tiles[n].ID == "seen" {
			restored.Tiles[n].InfluenceCityID = ""
		}
	}
	restored.Cities = nil
	refreshed, err := Update(restored)
	if err != nil {
		t.Fatal(err)
	}
	p, err = Project(refreshed, "a")
	if err != nil {
		t.Fatal(err)
	}
	assert(p, "seen", "", false)
}
