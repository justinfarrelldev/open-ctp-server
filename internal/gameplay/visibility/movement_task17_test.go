package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestTask17ProjectionExposesPlanningOnlyForCurrentlyVisibleTiles(t *testing.T) {
	state := fixtureState()
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state.Rules.Definitions = ruleset.CanonicalJSON
	state.Rules.ContentHash = ruleset.ContentHash
	state.Units[0].Domain = "land"
	state.Units[0].Movement = 2

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, tile := range projection.Tiles {
		if tile.Stale && tile.Planning != nil {
			t.Fatalf("stale tile %q leaked current planning data", tile.ID)
		}
		if !tile.Stale && tile.Planning == nil {
			t.Fatalf("visible tile %q omitted authoritative planning data", tile.ID)
		}
	}
	for _, unit := range projection.Units {
		if unit.OwnerID == "a" && (unit.Domain == "" || unit.Movement < 0) {
			t.Fatalf("owned unit %q omitted movement planner inputs: %#v", unit.ID, unit)
		}
		if unit.OwnerID != "a" && (unit.Domain != "" || unit.Movement != 0) {
			t.Fatalf("foreign unit %q leaked movement capability: %#v", unit.ID, unit)
		}
	}
	raw, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Tiles []map[string]any `json:"tiles"`
		Units []map[string]any `json:"units"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, tile := range decoded.Tiles {
		if stale, _ := tile["stale"].(bool); stale && tile["planning"] != nil {
			t.Fatalf("raw projection leaked stale planning: %s", raw)
		}
	}
	for _, unit := range decoded.Units {
		if unit["owner_id"] != "a" && (unit["domain"] != nil || unit["movement"] != nil) {
			t.Fatalf("raw projection leaked foreign planning capability: %s", raw)
		}
	}
}

func TestTask17PlanUsesOnlyVisibleProjectionData(t *testing.T) {
	projection := Projection{
		Tiles: []TileProjection{
			{ID: "a", X: 0, Y: 0, Terrain: "plains", Planning: &TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "b", X: 1, Y: 0, Terrain: "forest", Planning: &TilePlanning{Cost: 2, Domains: []string{"land"}, Accessible: true}},
			{ID: "c", X: 2, Y: 0, Terrain: "plains", Planning: &TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true, EnemyZOC: true}},
		},
		Units: []UnitProjection{{ID: "unit-a", OwnerID: "player-a", TileID: "a", Domain: "land", Movement: 3}},
	}
	path, err := Plan(projection, "unit-a", "c")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(path), 2; got != want || path[0] != "b" || path[1] != "c" {
		t.Fatalf("unexpected deterministic legal path: %#v", path)
	}
	projection.Tiles[1].Stale = true
	projection.Tiles[1].Planning = nil
	if _, err := Plan(projection, "unit-a", "c"); err == nil {
		t.Fatal("planner must reject a route requiring stale terrain")
	}
}
