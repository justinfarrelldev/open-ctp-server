package engine

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func task8State() State {
	state := fixtureState()
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		panic(err)
	}
	state.Rules = RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
	state.Tiles = []Tile{
		{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a"},
		{ID: "b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a"},
		{ID: "c", X: 2, Y: 0, Terrain: TerrainOcean},
		{ID: "d", X: 1, Y: 1, Terrain: TerrainPlains, OwnerID: "player-a"},
		{ID: "e", X: 2, Y: 1, Terrain: TerrainPlains, OwnerID: "player-a"},
	}
	state.Units = []Unit{
		{ID: "land", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 4, ZOC: true},
		{ID: "enemy", OwnerID: "player-b", TileID: "d", Domain: DomainLand, Movement: 4, ZOC: true},
		{ID: "ship", OwnerID: "player-a", TileID: "c", Domain: DomainSea, Movement: 4, CargoCapacity: 2, CargoDomains: []string{DomainLand}},
		{ID: "cargo", OwnerID: "player-a", TileID: "c", Domain: DomainLand, Movement: 3},
		{ID: "land-2", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 4},
	}
	return state
}

func task8Command(t *testing.T, kind string, payload any) Command {
	t.Helper()
	bytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return Command{SchemaVersion: CommandSchemaV1, ID: "task8-" + kind, ActorPlayerID: "player-a", Type: kind, Payload: bytes}
}

func applyTask8(t *testing.T, state State, kind string, payload any) (State, []Event, error) {
	t.Helper()
	next, events, _, err := Apply(state, task8Command(t, kind, payload), state.Rules, fixtureSources())
	return next, events, err
}

func TestTask8MoveCapturesUndefendedForeignCity(t *testing.T) {
	state := task8State()
	state.Cities = []City{{ID: "foreign-city", OwnerID: "player-b", TileID: "b"}}

	next, events, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventUnitMoved {
		t.Fatalf("events = %+v", events)
	}
	if next.Cities[0].OwnerID != "player-a" {
		t.Fatalf("city owner = %q", next.Cities[0].OwnerID)
	}
}

func TestTask8MoveTerrainZOCAndVisibilityVectors(t *testing.T) {
	state := task8State()
	next, events, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}})
	if err != nil {
		t.Fatal(err)
	}
	if next.Units[2].ID != "land" || next.Units[2].Movement != 2 || next.Units[2].TileID != "b" {
		t.Fatalf("unexpected move result: %#v", next.Units)
	}
	if len(events) != 1 || events[0].Type != EventUnitMoved || len(next.Fog) != 1 || next.Fog[0].TileID != "b" {
		t.Fatalf("move must publish exactly one visible result: %#v %#v", events, next.Fog)
	}
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b", "e"}}); err == nil || !strings.Contains(err.Error(), ErrObstructed.Error()) {
		t.Fatalf("ZOC path must be a generic obstruction, got %v", err)
	}
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"c"}}); err == nil {
		t.Fatal("land unit entered ocean")
	}
	for index := range state.Units {
		if state.Units[index].ID == "land" {
			state.Units[index].OwnerID = "player-b"
		}
	}
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}}); err == nil {
		t.Fatal("actor moved another player's unit")
	}
}

func TestTask8CargoFortifyDisbandAndArmyVectors(t *testing.T) {
	state := task8State()
	next, _, err := applyTask8(t, state, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Units[0].TransportID != "ship" || next.Units[0].TileID != "" {
		t.Fatalf("cargo was not canonicalized: %#v", next.Units[0])
	}
	next, _, err = applyTask8(t, next, CommandDisembark, map[string]string{"unit_id": "cargo", "tile_id": "e"})
	if err != nil {
		t.Fatal(err)
	}
	if next.Units[0].TransportID != "" || next.Units[0].TileID != "e" {
		t.Fatalf("cargo was not disembarked: %#v", next.Units[0])
	}
	next, _, err = applyTask8(t, next, CommandFortify, map[string]string{"unit_id": "cargo"})
	if err != nil || !next.Units[0].Fortified {
		t.Fatalf("fortify failed: %v %#v", err, next.Units[0])
	}
	next, events, err := applyTask8(t, next, CommandDisband, map[string]string{"unit_id": "cargo"})
	if err != nil || len(next.Units) != 4 || events[0].Type != EventUnitDisbanded {
		t.Fatalf("disband failed: %v %#v %#v", err, next.Units, events)
	}

	state = task8State()
	next, _, err = applyTask8(t, state, CommandArmyCreate, map[string]any{"army_id": "army", "unit_ids": []string{"land", "land-2"}})
	if err != nil || len(next.Armies) != 1 || next.Units[2].ArmyID != "army" {
		t.Fatalf("army create failed: %v %#v", err, next)
	}
	next, _, err = applyTask8(t, next, CommandArmyLeave, map[string]string{"army_id": "army", "unit_id": "land-2"})
	if err != nil || len(next.Armies[0].UnitIDs) != 1 || next.Units[1].ArmyID != "" {
		t.Fatalf("army leave failed: %v %#v", err, next)
	}
	if err := CheckInvariants(next); err != nil {
		t.Fatal(err)
	}
}

func TestTask8CapacityPhaseAndTransportLossVectors(t *testing.T) {
	state := task8State()
	state.Units = append(state.Units, Unit{ID: "cargo-2", OwnerID: "player-a", TileID: "c", Domain: DomainLand, Movement: 3}, Unit{ID: "cargo-3", OwnerID: "player-a", TileID: "c", Domain: DomainLand, Movement: 3})
	next, _, err := applyTask8(t, state, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}
	next, _, err = applyTask8(t, next, CommandEmbark, map[string]string{"unit_id": "cargo-2", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := applyTask8(t, next, CommandEmbark, map[string]string{"unit_id": "cargo-3", "transport_id": "ship"}); err == nil {
		t.Fatal("full transport accepted cargo")
	}
	next, _, err = applyTask8(t, next, CommandDisband, map[string]string{"unit_id": "ship"})
	if err != nil || len(next.Units) != 4 {
		t.Fatalf("transport loss must remove cargo atomically: %v %#v", err, next.Units)
	}
	state = task8State()
	state.Phase.ActivePlayerIndex = 1
	if _, _, err := applyTask8(t, state, CommandFortify, map[string]string{"unit_id": "land"}); err == nil {
		t.Fatal("inactive phase actor accepted action")
	}
}

func TestTask8RejectsUnknownDomainsAndAllowsLoadedTransportMovement(t *testing.T) {
	state := task8State()
	state.Units[0].Domain = "unsupported"
	if err := ValidateState(state); err == nil {
		t.Fatal("unknown unit domain was accepted")
	}

	state = task8State()
	next, _, err := applyTask8(t, state, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}
	next.Tiles = append(next.Tiles, Tile{ID: "f", X: 3, Y: 0, Terrain: TerrainOcean})
	moved, _, err := applyTask8(t, next, CommandMoveUnit, map[string]any{"unit_id": "ship", "path": []string{"f"}})
	if err != nil {
		t.Fatalf("loaded transport should move using transport movement domain, got %v", err)
	}
	if ship, ok := unitByID(moved, "ship"); !ok || ship.TileID != "f" {
		t.Fatalf("loaded transport did not move: %#v", moved.Units)
	}
}

func TestTask8ArmyJoinRemovalAndInvalidVectors(t *testing.T) {
	state := task8State()
	next, _, err := applyTask8(t, state, CommandArmyCreate, map[string]any{"army_id": "army", "unit_ids": []string{"land"}})
	if err != nil {
		t.Fatal(err)
	}
	next, _, err = applyTask8(t, next, CommandArmyJoin, map[string]string{"army_id": "army", "unit_id": "land-2"})
	if err != nil || len(next.Armies[0].UnitIDs) != 2 {
		t.Fatalf("army join failed: %v %#v", err, next.Armies)
	}
	next, _, err = applyTask8(t, next, CommandArmyLeave, map[string]string{"army_id": "army", "unit_id": "land-2"})
	if err != nil {
		t.Fatal(err)
	}
	next, _, err = applyTask8(t, next, CommandArmyLeave, map[string]string{"army_id": "army", "unit_id": "land"})
	if err != nil || len(next.Armies) != 0 {
		t.Fatalf("last army member did not remove army: %v %#v", err, next.Armies)
	}

	state = task8State()
	next, _, err = applyTask8(t, state, CommandArmyCreate, map[string]any{"army_id": "army", "unit_ids": []string{"land", "land-2"}})
	if err != nil {
		t.Fatal(err)
	}
	next, _, err = applyTask8(t, next, CommandDisband, map[string]string{"unit_id": "land"})
	if err != nil || len(next.Armies) != 1 || len(next.Armies[0].UnitIDs) != 1 {
		t.Fatalf("disband did not remove army member atomically: %v %#v", err, next.Armies)
	}
	bad := Command{SchemaVersion: CommandSchemaV1, ID: "bad", ActorPlayerID: "player-a", Type: CommandMoveUnit, Payload: json.RawMessage(`{}`)}
	if _, _, _, err := Apply(task8State(), bad, state.Rules, fixtureSources()); err == nil {
		t.Fatal("empty move payload accepted")
	}
	bad.Type = "unknown-unit-command"
	if _, _, _, err := Apply(task8State(), bad, state.Rules, fixtureSources()); err == nil {
		t.Fatal("unknown command accepted")
	}
}

func TestUnitMovementSpendsFuelAndRejectsEmptyFuel(t *testing.T) {
	state := task8State()
	state.Units = []Unit{{ID: "air", OwnerID: "player-a", TileID: "a", Domain: DomainAir, Movement: 2, Fuel: 1, MaxFuel: 1}}
	state.Tiles = []Tile{{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "b", X: 1, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "d", X: 1, Y: 1, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, OpenAccess: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 1, Domains: []string{DomainAir}}}}}

	if err := moveUnit(&state, "player-a", moveUnitPayload{UnitID: "air", Path: []string{"b"}}, runtime); err != nil {
		t.Fatalf("moveUnit() error = %v", err)
	}
	unit, ok := unitByID(state, "air")
	if !ok || unit.Fuel != 0 || unit.Movement != 1 || unit.TileID != "b" {
		t.Fatalf("unit after fuel move = %+v ok=%v", unit, ok)
	}
	if err := moveUnit(&state, "player-a", moveUnitPayload{UnitID: "air", Path: []string{"d"}}, runtime); err == nil {
		t.Fatal("moveUnit accepted a fueled unit with empty fuel")
	}
}

func TestAirUnitsRefuelAtFriendlyCitiesOnAdvanceRound(t *testing.T) {
	state := task8State()
	state.Units = []Unit{{ID: "air", OwnerID: "player-a", TileID: "a", Domain: DomainAir, Movement: 0, Fuel: 1, MaxFuel: 5}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "a"}}
	runtime := &rules.RuntimeRules{Starter: rules.RuntimeStarter{Movement: 2}}

	advanceRound(&state, runtime)
	unit, ok := unitByID(state, "air")
	if !ok || unit.Fuel != 5 || unit.Movement != 2 {
		t.Fatalf("unit after advance = %+v ok=%v, want refueled and movement refreshed", unit, ok)
	}
}

func TestNoFuelThenCrashRemovesStrandedAirUnitsOnAdvanceRound(t *testing.T) {
	state := task8State()
	state.Units = []Unit{{ID: "air", OwnerID: "player-a", TileID: "a", Domain: DomainAir, Movement: 0, Fuel: 0, MaxFuel: 5, Effects: []Effect{{ID: "crash", Kind: EffectNoFuelThenCrash}}}}
	runtime := &rules.RuntimeRules{Starter: rules.RuntimeStarter{Movement: 2}}

	advanceRound(&state, runtime)

	if _, ok := unitByID(state, "air"); ok {
		t.Fatal("stranded no-fuel air unit survived advanceRound")
	}
}

func TestRetailCanEntrenchRestrictsFortify(t *testing.T) {
	state := task8State()
	state.Units = []Unit{{ID: "special", OwnerID: "player-a", TileID: "a", Effects: []Effect{{ID: "retail", Kind: EffectCantCaptureCity}}}}
	if _, _, err := applyTask8(t, state, CommandFortify, map[string]string{"unit_id": "special"}); err == nil {
		t.Fatal("fortify accepted a retail unit without CanEntrench")
	}
	state.Units[0].Effects = append(state.Units[0].Effects, Effect{ID: "entrench", Kind: EffectCanEntrench})
	if next, _, err := applyTask8(t, state, CommandFortify, map[string]string{"unit_id": "special"}); err != nil || !next.Units[0].Fortified {
		t.Fatalf("fortify with CanEntrench failed: %v %#v", err, next.Units[0])
	}
}

func TestTask8MovementAndCargoFailureVectors(t *testing.T) {
	state := task8State()
	for index := range state.Units {
		if state.Units[index].ID == "land" {
			state.Units[index].Movement = 0
		}
	}
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}}); err == nil {
		t.Fatal("zero-movement unit moved")
	}
	state = task8State()
	for index := range state.Tiles {
		if state.Tiles[index].ID == "b" {
			state.Tiles[index].OwnerID = "player-b"
		}
	}
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}}); err == nil {
		t.Fatal("closed enemy tile entered")
	}
	state = task8State()
	next, _, err := applyTask8(t, state, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := applyTask8(t, next, CommandDisembark, map[string]string{"unit_id": "cargo", "tile_id": "c"}); err == nil {
		t.Fatal("cargo disembarked to water")
	}
	if _, _, err := applyTask8(t, next, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"}); err == nil {
		t.Fatal("already embarked cargo re-embarked")
	}
}

func TestTask8ArmyMovementAndOpenAccessVectors(t *testing.T) {
	state := task8State()
	next, _, err := applyTask8(t, state, CommandArmyCreate, map[string]any{"army_id": "army", "unit_ids": []string{"land", "land-2"}})
	if err != nil {
		t.Fatal(err)
	}
	next, _, err = applyTask8(t, next, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}})
	if err != nil || next.Units[2].TileID != "b" {
		t.Fatalf("army member movement failed: %v %#v", err, next.Units)
	}
	state = task8State()
	for index := range state.Tiles {
		if state.Tiles[index].ID == "b" {
			state.Tiles[index].OwnerID, state.Tiles[index].Open = "player-b", true
		}
	}
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}}); err != nil {
		t.Fatalf("open tile access rejected: %v", err)
	}
}

func TestTask8RejectsSplitArmiesAndLockedResolution(t *testing.T) {
	state := task8State()
	next, _, err := applyTask8(t, state, CommandArmyCreate, map[string]any{"army_id": "army", "unit_ids": []string{"land", "land-2"}})
	if err != nil {
		t.Fatal(err)
	}
	for index := range next.Units {
		if next.Units[index].ID == "land-2" {
			next.Units[index].TileID = "b"
		}
	}
	if err := ValidateState(next); err == nil {
		t.Fatal("split army was accepted")
	}
	state = task8State()
	state.Phase.ResolutionLocked = true
	if _, _, err := applyTask8(t, state, CommandFortify, map[string]string{"unit_id": "land"}); err == nil {
		t.Fatal("locked phase accepted action")
	}
}

func TestTask17MovementRequiresPinnedTypedRuntimeDefinitions(t *testing.T) {
	state := task8State()
	state.Rules.Definitions = nil
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}}); err == nil || !strings.Contains(err.Error(), ErrObstructed.Error()) {
		t.Fatalf("movement without immutable typed definitions must fail closed, got %v", err)
	}
}

func TestTask17RuntimeAccessAndCostVectors(t *testing.T) {
	state := task8State()
	state.Tiles[1].OwnerID = "player-b"
	if _, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}}); err == nil || !strings.Contains(err.Error(), ErrObstructed.Error()) {
		t.Fatalf("foreign closed tile must obstruct movement, got %v", err)
	}
	state.Tiles[1].Open = true
	next, _, err := applyTask8(t, state, CommandMoveUnit, map[string]any{"unit_id": "land", "path": []string{"b"}})
	if err != nil {
		t.Fatal(err)
	}
	moved, ok := unitByID(next, "land")
	if !ok || moved.Movement != 2 {
		t.Fatalf("forest cost must come from the pinned typed terrain model: %#v", moved)
	}
}
