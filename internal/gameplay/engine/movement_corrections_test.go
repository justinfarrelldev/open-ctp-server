package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func correctedMovementState(t *testing.T) State {
	t.Helper()
	s := journeyState(t)
	base, err := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	base.Movement.RetailMovementVersion = 1
	raw, hash, err := rules.CanonicalRuntimeDefinitions(base)
	if err != nil {
		t.Fatal(err)
	}
	s.Rules.Definitions, s.Rules.ContentHash = raw, hash
	return s
}

func setFirstMove(t *testing.T, unit *Unit, ready bool) {
	t.Helper()
	unit.FirstMove = &ready
}

func TestCorrectedForeignTerrainAndProtectedCity(t *testing.T) {
	for _, obstacle := range []string{"none", "city", "enemy", "full"} {
		t.Run(obstacle, func(t *testing.T) {
			s := correctedMovementState(t)
			s.Tiles[1].OwnerID, s.Tiles[1].Open = "player-b", false
			switch obstacle {
			case "city":
				s.Cities[0].TileID = "t1"
			case "enemy":
				s.Units = append(s.Units, Unit{ID: "enemy", OwnerID: "player-b", TileID: "t1"})
			case "full":
				for i := 0; i < maxArmySize; i++ {
					s.Units = append(s.Units, Unit{ID: string(rune('a' + i)), OwnerID: "player-a", TileID: "t1"})
				}
			}
			err := runtimeApplyExpectError(t, s, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
			if (err == nil) != (obstacle == "none") {
				t.Fatalf("obstacle %s: %v", obstacle, err)
			}
		})
	}
}

func TestCorrectedFirstMoveAndRestart(t *testing.T) {
	s := correctedMovementState(t)
	s.Units[0].Movement, s.Units[0].MaxMovement = 100, 100
	setFirstMove(t, &s.Units[0], true)
	// Synthetic forest cost 2 is scaled here to the reported retail 200.
	runtime, _ := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
	for i := range runtime.Movement.Terrains {
		if runtime.Movement.Terrains[i].ID == TerrainForest {
			runtime.Movement.Terrains[i].Cost = 200
		}
	}
	raw, hash, _ := rules.CanonicalRuntimeDefinitions(runtime)
	s.Rules.Definitions, s.Rules.ContentHash = raw, hash
	s.Tiles[1].Terrain = TerrainForest
	bytes, _ := CanonicalStateJSON(s)
	restored, err := DecodeCanonicalStateJSON(bytes)
	if err != nil {
		t.Fatal(err)
	}
	var expectedHash string
	for _, input := range []State{s, restored} {
		next := runtimeApply(t, input, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
		hash := canonicalStateHashForTest(t, next)
		if expectedHash != "" && hash != expectedHash {
			t.Fatal("serialized first-move state changed execution")
		}
		expectedHash = hash
		if next.Units[0].TileID != "t1" || next.Units[0].Movement != 0 {
			t.Fatalf("expensive first step: %+v", next.Units[0])
		}
		if err := runtimeApplyExpectError(t, next, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t2"}}); err == nil {
			t.Fatal("second step granted allowance")
		}
	}
}

func TestCorrectedRetainedWaitsThenTakesExpensiveFirstStep(t *testing.T) {
	s := correctedMovementState(t)
	setFirstMove(t, &s.Units[0], false)
	s.Units[0].Movement = 0
	s.Tiles[1].Terrain = TerrainForest
	s = runtimeApply(t, s, "player-a", CommandSetMovementOrder, map[string]any{"unit_id": "traveller", "target_x": 2, "target_y": 0})
	if len(s.MovementOrders) != 1 {
		t.Fatal("legal expensive step discarded instead of deferred")
	}
	s = runtimeApply(t, s, "player-a", CommandRoundComplete, map[string]any{})
	if s.Units[0].TileID != "t1" || s.Units[0].Movement != 0 || len(s.MovementOrders) != 1 {
		t.Fatalf("fresh step failed: %+v", s.Units[0])
	}
	raw, _ := CanonicalStateJSON(s)
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	a := runtimeApply(t, s, "player-a", CommandRoundComplete, map[string]any{})
	b := runtimeApply(t, restored, "player-a", CommandRoundComplete, map[string]any{})
	if canonicalStateHashForTest(t, a) != canonicalStateHashForTest(t, b) || a.Units[0].TileID != "t2" {
		t.Fatal("retained movement diverged after restart")
	}
}

func TestCorrectedPerMemberFirstMoveAndFuel(t *testing.T) {
	for _, kind := range []string{"fresh", "spent", "unknown", "fuel", "cheap-first"} {
		t.Run(kind, func(t *testing.T) {
			s := correctedMovementState(t)
			s.Tiles[1].Terrain = TerrainForest
			s.Units[0].Movement = 1
			setFirstMove(t, &s.Units[0], true)
			slow := s.Units[0]
			slow.ID = "slow"
			switch kind {
			case "spent":
				setFirstMove(t, &slow, false)
			case "unknown":
				slow.FirstMove = nil
			case "fuel":
				slow.MaxFuel = 10
				slow.Fuel = 1
			case "cheap-first":
				s.Tiles[1].Terrain = TerrainPlains
				s.Tiles[2].Terrain = TerrainForest
				s.Units[0].Movement = 2
				slow.Movement = 2
			}
			s.Units = append(s.Units, slow)
			s = runtimeApply(t, s, "player-a", CommandArmyCreate, map[string]any{"army_id": "army", "unit_ids": []string{"traveller", "slow"}})
			path := []string{"t1"}
			if kind == "cheap-first" {
				path = append(path, "t2")
			}
			before := canonicalStateHashForTest(t, s)
			err := runtimeApplyExpectError(t, s, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": path})
			if (err == nil) != (kind == "fresh") {
				t.Fatalf("%s: %v", kind, err)
			}
			if canonicalStateHashForTest(t, s) != before {
				t.Fatal("trial mutated original")
			}
		})
	}
}

func TestCorrectedCargoReceivesPerUnitRefresh(t *testing.T) {
	for _, refresh := range []func(*State, *rules.RuntimeRules){refreshStarterMovement, advanceRound} {
		s := correctedMovementState(t)
		s.Units[0].TransportID = "ship"
		s.Units[0].Movement = 0
		setFirstMove(t, &s.Units[0], false)
		s.Units = append(s.Units, Unit{ID: "ship", OwnerID: "player-a", TileID: "t0", Domain: DomainSea, MaxMovement: 1})
		runtime, _ := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
		refresh(&s, &runtime)
		cargo, _ := unitByID(s, "traveller")
		if cargo.Movement != 1 || cargo.FirstMove == nil || !*cargo.FirstMove {
			t.Fatalf("cargo allowance not refreshed: %+v", cargo)
		}
	}
}

func TestCorrectedDisembarkFirstMoveAndProtectedCity(t *testing.T) {
	for _, city := range []bool{false, true} {
		s := correctedMovementState(t)
		s.Tiles[1].Terrain = TerrainForest
		s.Units[0].TransportID = "ship"
		setFirstMove(t, &s.Units[0], true)
		s.Units = append(s.Units, Unit{ID: "ship", OwnerID: "player-a", TileID: "t0", Domain: DomainSea, CargoCapacity: 1})
		if city {
			s.Cities[0].TileID = "t1"
		}
		runtime, _ := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
		err := disembark(&s, "player-a", disembarkPayload{UnitID: "traveller", TileID: "t1"}, &runtime)
		if (err == nil) == city {
			t.Fatalf("city=%v disembark: %v", city, err)
		}
		if !city && (s.Units[0].Movement != 0 || *s.Units[0].FirstMove) {
			t.Fatal("disembark did not spend allowance")
		}
	}
}

func TestCorrectedMixedDomainArmy(t *testing.T) {
	for _, domain := range []string{DomainAmphibious, DomainAir, DomainSea} {
		s := correctedMovementState(t)
		slow := s.Units[0]
		slow.ID = "other"
		slow.Domain = domain
		slow.Movement = 100
		s.Units = append(s.Units, slow)
		s = runtimeApply(t, s, "player-a", CommandArmyCreate, map[string]any{"army_id": "mixed", "unit_ids": []string{"traveller", "other"}})
		// The synthetic land tile needs explicit air eligibility, as retail land has.
		runtime, _ := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
		for i := range runtime.Movement.Terrains {
			if runtime.Movement.Terrains[i].ID == TerrainPlains {
				runtime.Movement.Terrains[i].Domains = []string{DomainAir, DomainAmphibious, DomainLand}
			}
		}
		err := moveUnit(&s, "player-a", moveUnitPayload{UnitID: "traveller", Path: []string{"t1"}}, &runtime)
		if (err == nil) != (domain != DomainSea) {
			t.Fatalf("mixed %s army: %v", domain, err)
		}
	}
}

func TestCorrectedCityWarAndCaptureCapability(t *testing.T) {
	for _, cantCapture := range []bool{false, true} {
		s := correctedMovementState(t)
		s.Cities[0].TileID = "t1"
		setWarState(&s, "player-a", "player-b", true)
		if cantCapture {
			s.Units[0].Effects = []Effect{{ID: "civilian", Kind: EffectCantCaptureCity}}
		}
		err := runtimeApplyExpectError(t, s, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
		if (err == nil) == cantCapture {
			t.Fatalf("cant_capture=%v: %v", cantCapture, err)
		}
	}
}

func TestCorrectedRoutingDoesNotConsultHiddenUnits(t *testing.T) {
	s := correctedMovementState(t)
	s.Tiles = append(s.Tiles, Tile{ID: "detour", X: 1, Y: 1, Terrain: TerrainPlains, Open: true})
	runtime, _ := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
	order := MovementOrder{UnitID: "traveller", UnitIDs: []string{"traveller"}, PlayerID: "player-a", TileID: "t0", TargetX: 2, TargetY: 0}
	baseline, ok := nextMovementStep(s, order, &runtime)
	if !ok {
		t.Fatal("baseline route missing")
	}
	s.Units = append(s.Units, Unit{ID: "hidden", OwnerID: "player-b", TileID: baseline.ID, Domain: DomainLand, ZOC: true, Effects: []Effect{{ID: "stealth", Kind: EffectStealth}}})
	next, ok := nextMovementStep(s, order, &runtime)
	if !ok || next.ID != baseline.ID {
		t.Fatalf("hidden unit changed route %s -> %s", baseline.ID, next.ID)
	}
}

func TestCorrectedRoutingIgnoresUnseenTerrainAndReplansObstacle(t *testing.T) {
	s := correctedMovementState(t)
	runtime, _ := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
	order := MovementOrder{UnitID: "traveller", UnitIDs: []string{"traveller"}, PlayerID: "player-a", TileID: "t0", TargetX: 7, TargetY: 0}
	baseline, ok := nextMovementStep(s, order, &runtime)
	if !ok {
		t.Fatal("frontier missing")
	}
	s.Tiles[4].Terrain = TerrainOcean
	next, ok := nextMovementStep(s, order, &runtime)
	if !ok || next.ID != baseline.ID {
		t.Fatal("unseen terrain influenced route")
	}
	s.Tiles[1].Terrain = TerrainOcean
	s.Tiles = append(s.Tiles, Tile{ID: "detour", X: 1, Y: 1, Terrain: TerrainPlains, Open: true})
	next, ok = nextMovementStep(s, order, &runtime)
	if !ok || next.ID != "detour" {
		t.Fatalf("changed obstacle not replanned: %+v", next)
	}
}

func TestReleasedAstarCappedCostsAndPrettyPath(t *testing.T) {
	s := correctedMovementState(t)
	s.Tiles = []Tile{{ID: "t0", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "forest", X: 0, Y: 1, Terrain: TerrainForest, Open: true}, {ID: "target", X: 0, Y: 2, Terrain: TerrainPlains, Open: true}, {ID: "detour", X: 1, Y: 1, Terrain: TerrainPlains, Open: true}}
	runtime, _ := rules.CompileRuntime(s.Rules.Definitions, s.Rules.ContentHash)
	order := MovementOrder{UnitID: "traveller", UnitIDs: []string{"traveller"}, PlayerID: "player-a", TileID: "t0", TargetX: 0, TargetY: 2}
	next, ok := nextMovementStep(s, order, &runtime)
	// UnitAstar::ComputeValidMovCost caps the forest to the slowest max (1).
	// UnitAstar::InitPoint / Astar::DecayOrtho then prefer 0.95+0.95 over 1+0.95.
	if !ok || next.ID != "forest" {
		t.Fatalf("not released capped/pretty routing: %+v", next)
	}
}

func TestCorrectedRetainedDetours(t *testing.T) {
	s := correctedMovementState(t)
	s.Tiles[1].Terrain = TerrainOcean
	s.Tiles = append(s.Tiles, Tile{ID: "detour", X: 1, Y: 1, Terrain: TerrainPlains, Open: true})
	s.Units[0].Movement = 2
	s = runtimeApply(t, s, "player-a", CommandSetMovementOrder, map[string]any{"unit_id": "traveller", "target_x": 2, "target_y": 0})
	if s.Units[0].TileID != "t2" {
		t.Fatalf("did not route around water: %+v", s.Units[0])
	}
}
