package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestCorrectedPlanningFirstStepAndArmy(t *testing.T) {
	for _, status := range []string{"true", "false", "null"} {
		t.Run(status, func(t *testing.T) {
			var unit UnitProjection
			if err := json.Unmarshal([]byte(`{"id":"u","owner_id":"a","tile_id":"origin","domain":"land","movement":100,"max_movement":100,"first_move":`+status+`}`), &unit); err != nil {
				t.Fatal(err)
			}
			p := Projection{Units: []UnitProjection{unit}, Tiles: []TileProjection{{ID: "origin", X: 0, Planning: &TilePlanning{Cost: 100, Accessible: true, Domains: []string{"land"}}}, {ID: "forest", X: 1, Planning: &TilePlanning{Cost: 200, Accessible: true, Domains: []string{"land"}}}}}
			_, err := Plan(p, "u", "forest")
			if (err == nil) != (status == "true") {
				t.Fatalf("first_move=%s route: %v", status, err)
			}
			actions := moveActions(p, map[string]TileProjection{"origin": p.Tiles[0]}, unit)
			if (len(actions) > 0) != (status == "true") {
				t.Fatalf("first_move=%s actions: %+v", status, actions)
			}
		})
	}
}

func TestCorrectedPlanningMixedArmyAndAir(t *testing.T) {
	ready, spent := true, false
	for _, kind := range []string{"slow-spent", "slow-fresh", "air", "amphibious", "civilian-city", "full-stack", "stationary"} {
		t.Run(kind, func(t *testing.T) {
			leader := UnitProjection{ID: "u", OwnerID: "a", TileID: "origin", Domain: "land", Movement: 100, FirstMove: &ready}
			p := Projection{Units: []UnitProjection{leader}, Rules: &RulesProjection{Movement: &rules.RuntimeMovement{RetailMovementVersion: 1}}, Tiles: []TileProjection{{ID: "origin", Planning: &TilePlanning{Cost: 100, Accessible: true, Domains: []string{"air", "land"}}}, {ID: "target", X: 1, Planning: &TilePlanning{Cost: 200, Accessible: true, Domains: []string{"air", "land"}}}}}
			want := true
			switch kind {
			case "slow-spent", "slow-fresh":
				p.Units[0].ArmyID = "army"
				slow := p.Units[0]
				slow.ID = "slow"
				slow.FirstMove = &spent
				if kind == "slow-fresh" {
					slow.FirstMove = &ready
				} else {
					want = false
				}
				p.Units = append(p.Units, slow)
			case "air":
				p.Units[0].Domain = "air"
				p.Units[0].FirstMove = &spent
			case "amphibious":
				p.Units[0].Domain = "amphibious"
			case "civilian-city":
				p.Cities = []CityProjection{{ID: "city", OwnerID: "b", TileID: "target"}}
				p.Units[0].Effects = []engine.Effect{{Kind: engine.EffectCantCaptureCity}}
				want = false
			case "stationary":
				p.Units[0].TypeID = "UNIT_CITY"
				p.Rules.Units = []rules.RuntimeUnitDefinition{{ID: "UNIT_CITY", Movement: 0}}
				want = false
			case "full-stack":
				for i := 0; i < 12; i++ {
					p.Units = append(p.Units, UnitProjection{ID: string(rune('a' + i)), OwnerID: "a", TileID: "target"})
				}
				want = false
			}
			_, err := Plan(p, "u", "target")
			if (err == nil) != want {
				t.Fatalf("%s: %v", kind, err)
			}
		})
	}
}

func TestCorrectedDisembarkPlanningFirstMove(t *testing.T) {
	for _, ready := range []bool{false, true} {
		cargo := UnitProjection{ID: "cargo", OwnerID: "a", TransportID: "ship", Domain: "land", FirstMove: &ready}
		ship := UnitProjection{ID: "ship", OwnerID: "a", TileID: "origin", Domain: "sea"}
		origin := TileProjection{ID: "origin"}
		target := TileProjection{ID: "forest", X: 1, Planning: &TilePlanning{Cost: 200, Accessible: true, Domains: []string{"land"}}}
		p := Projection{Units: []UnitProjection{cargo, ship}, Tiles: []TileProjection{origin, target}}
		actions := disembarkActions(p, map[string]TileProjection{"origin": origin}, map[string]UnitProjection{"ship": ship}, cargo)
		if (len(actions) == 1) != ready {
			t.Fatalf("cargo first_move=%v: %+v", ready, actions)
		}
	}
}

func TestReleasedAstarProjectionRoutePreferences(t *testing.T) {
	p := Projection{Rules: &RulesProjection{Movement: &rules.RuntimeMovement{RetailMovementVersion: 1}}, Units: []UnitProjection{{ID: "u", OwnerID: "a", TileID: "origin", Domain: "land", Movement: 3, MaxMovement: 1}}, Tiles: []TileProjection{{ID: "origin", Planning: &TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}}, {ID: "forest", Y: 1, Planning: &TilePlanning{Cost: 2, Accessible: true, Domains: []string{"land"}}}, {ID: "target", Y: 2, Planning: &TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}}, {ID: "detour", X: 1, Y: 1, Planning: &TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}}}}
	path, err := Plan(p, "u", "target")
	if err != nil || len(path) != 2 || path[0] != "forest" {
		t.Fatalf("human capped/pretty route: %v %v", path, err)
	}
	aiPath, err := PlanForAI(p, "u", "target")
	if err != nil || len(aiPath) != 2 || aiPath[0] != "detour" {
		t.Fatalf("AI route should omit human preference: %v %v", aiPath, err)
	}
	p.Tiles[1], p.Tiles[3] = p.Tiles[3], p.Tiles[1]
	again, err := Plan(p, "u", "target")
	if err != nil || len(again) != 2 || again[0] != path[0] {
		t.Fatalf("route changed with slice order: %v %v", again, err)
	}
}

func TestCorrectedPlanningAccessAndPrivateFirstMove(t *testing.T) {
	s := inspectionState(t)
	runtime, err := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Movement.RetailMovementVersion = 1
	defs, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	s.Rules.Definitions, s.Rules.ContentHash = defs, hash
	s.Cities = nil
	s.Tiles[1].OwnerID = "b"
	s.Tiles[1].Open = false
	s.Units[0].FirstMove = engine.NewFirstMove(&runtime)
	p, err := Project(s, "a")
	if err != nil {
		t.Fatal(err)
	}
	tile := metadataTile(t, p, "t1")
	if tile.Planning == nil || !tile.Planning.Accessible {
		t.Fatal("ordinary foreign terrain still closed")
	}
	raw, _ := json.Marshal(p.Units[0])
	var wire map[string]any
	_ = json.Unmarshal(raw, &wire)
	if wire["first_move"] != true {
		t.Fatal("owner did not receive explicit first-move fact")
	}
	s.Units = append(s.Units, engine.Unit{ID: "observer", OwnerID: "b", TileID: "t1"})
	p, err = Project(s, "b")
	if err != nil {
		t.Fatal(err)
	}
	foundEnemy := false
	for _, unit := range p.Units {
		if unit.OwnerID == "a" {
			foundEnemy = true
			raw, _ = json.Marshal(unit)
			var enemy map[string]any
			_ = json.Unmarshal(raw, &enemy)
			if _, ok := enemy["first_move"]; ok {
				t.Fatal("foreign movement state leaked")
			}
		}
	}
	if !foundEnemy {
		t.Fatal("privacy assertion had no visible foreign unit")
	}
}
