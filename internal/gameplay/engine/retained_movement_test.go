package engine

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
)

func journeyState(t *testing.T) State {
	t.Helper()
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	state.Tiles = nil
	for x := 0; x < 8; x++ {
		state.Tiles = append(state.Tiles, Tile{ID: fmt.Sprintf("t%d", x), X: x, Terrain: TerrainPlains, Open: true})
	}
	state.Tiles = append(state.Tiles, Tile{ID: "far", X: 50, Y: 50, Terrain: TerrainPlains, Open: true})
	state.Cities = []City{{ID: "foreign", OwnerID: "player-b", TileID: "far", Population: 1}}
	state.Units = []Unit{{ID: "traveller", OwnerID: "player-a", TileID: "t0", Domain: DomainLand, Movement: 1, MaxMovement: 1}}
	return state
}

func retainedCount(t *testing.T, state State) int {
	t.Helper()
	raw, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Orders []json.RawMessage `json:"movement_orders"`
	}
	if err = json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	return len(decoded.Orders)
}

func TestRetainedMovementContinuesAfterRestartAndCompletes(t *testing.T) {
	state := journeyState(t)
	state = runtimeApply(t, state, "player-a", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0})
	unit, _ := unitByID(state, "traveller")
	if unit.TileID != "t1" || retainedCount(t, state) != 1 {
		t.Fatalf("first prefix: %+v", state)
	}
	raw, _ := CanonicalStateJSON(state)
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	a := runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	b := runtimeApply(t, restored, "player-a", CommandRoundComplete, map[string]any{})
	if canonicalStateHashForTest(t, a) != canonicalStateHashForTest(t, b) {
		t.Fatal("restart diverged")
	}
	for i := 0; i < 10 && retainedCount(t, a) > 0; i++ {
		a = runtimeApply(t, a, "player-a", CommandRoundComplete, map[string]any{})
	}
	unit, _ = unitByID(a, "traveller")
	if unit.TileID != "t7" || retainedCount(t, a) != 0 {
		t.Fatalf("journey did not complete: %+v", a)
	}
}

func TestRetainedMovementCancellationReplacementAndManualOverride(t *testing.T) {
	for _, kind := range []string{"cancel-movement-order", CommandSleep, CommandFortify, CommandMoveUnit} {
		t.Run(kind, func(t *testing.T) {
			state := journeyState(t)
			state = runtimeApply(t, state, "player-a", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0})
			payload := map[string]any{"unit_id": "traveller"}
			if kind == CommandMoveUnit {
				state.Units[0].Movement = 1
				payload["path"] = []string{"t0"}
			}
			state = runtimeApply(t, state, "player-a", kind, payload)
			if retainedCount(t, state) != 0 {
				t.Fatal("manual command retained old destination")
			}
		})
	}
	state := journeyState(t)
	state = runtimeApply(t, state, "player-a", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0})
	state = runtimeApply(t, state, "player-a", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 0, "target_y": 0})
	state = runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	unit, _ := unitByID(state, "traveller")
	if unit.TileID != "t0" || retainedCount(t, state) != 0 {
		t.Fatal("replacement did not take effect")
	}
}

func TestRetainedMovementStopsWithoutAttackingOrSkippingObstacle(t *testing.T) {
	for _, obstacle := range []string{"terrain", "enemy", "zoc"} {
		t.Run(obstacle, func(t *testing.T) {
			state := journeyState(t)
			state = runtimeApply(t, state, "player-a", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0})
			switch obstacle {
			case "terrain":
				for i := range state.Tiles {
					if state.Tiles[i].ID == "t2" {
						state.Tiles[i].Terrain = TerrainOcean
					}
				}
			case "enemy":
				state.Units = append(state.Units, Unit{ID: "enemy", OwnerID: "player-b", TileID: "t2", Domain: DomainLand})
			case "zoc":
				state.Units = append(state.Units, Unit{ID: "enemy", OwnerID: "player-b", TileID: "t3", Domain: DomainLand, ZOC: true})
			}
			state = runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
			unit, _ := unitByID(state, "traveller")
			want := "t1"
			if obstacle == "zoc" {
				want = "t2"
			}
			if unit.TileID != want || retainedCount(t, state) != 0 {
				t.Fatalf("unsafe obstacle behavior: %+v", state)
			}
		})
	}
}

func TestRetainedMovementArmyUsesSlowestMemberAndCancelsMembershipChanges(t *testing.T) {
	state := journeyState(t)
	state.Units = append(state.Units, Unit{ID: "slow", OwnerID: "player-a", TileID: "t0", Domain: DomainLand, Movement: 0, MaxMovement: 1})
	state = runtimeApply(t, state, "player-a", CommandArmyCreate, map[string]any{"army_id": "army", "unit_ids": []string{"traveller", "slow"}})
	state = runtimeApply(t, state, "player-a", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0})
	for _, unit := range state.Units {
		if unit.TileID != "t0" {
			t.Fatal("split army")
		}
	}
	state = runtimeApply(t, state, "player-a", CommandArmyLeave, map[string]any{"army_id": "army", "unit_id": "slow"})
	if retainedCount(t, state) != 0 {
		t.Fatal("changed army retained journey")
	}
}

func TestRetainedMovementRejectsForeignInvalidAndWrongPhaseWithoutMutation(t *testing.T) {
	state := journeyState(t)
	before := canonicalStateHashForTest(t, state)
	for _, payload := range []map[string]any{{"unit_id": "unknown", "target_x": 7, "target_y": 0}, {"unit_id": "traveller", "target_x": 999, "target_y": 0}, {"unit_id": "traveller", "target_x": 7}, {"unit_id": "traveller", "target_x": nil, "target_y": 0}} {
		if err := runtimeApplyExpectError(t, state, "player-a", "set-movement-order", payload); err == nil {
			t.Fatalf("accepted %+v", payload)
		}
	}
	if err := runtimeApplyExpectError(t, state, "player-b", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0}); err == nil {
		t.Fatal("foreign/wrong-phase accepted")
	}
	if canonicalStateHashForTest(t, state) != before {
		t.Fatal("failed orders mutated input")
	}
}

func TestRetainedMovementQueuedOnlyExecutesOnResolution(t *testing.T) {
	for _, mode := range []string{turns.ModeSimultaneous, turns.ModeHybrid, turns.ModeLongTurn} {
		t.Run(mode, func(t *testing.T) {
			state := journeyState(t)
			state.Phase.Mode = mode
			state = runtimeApply(t, state, "player-a", "set-movement-order", map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0})
			if mode != turns.ModeLongTurn {
				unit, _ := unitByID(state, "traveller")
				if unit.TileID != "t0" {
					t.Fatal("planning executed movement")
				}
			}
			end := CommandRoundComplete
			if mode == turns.ModeLongTurn {
				end = CommandPass
			}
			state = runtimeApply(t, state, "player-a", end, map[string]any{})
			state = runtimeApply(t, state, "player-b", end, map[string]any{})
			unit, _ := unitByID(state, "traveller")
			if unit.TileID == "t0" {
				t.Fatal("resolved journey did not advance")
			}
			if retainedCount(t, state) != 1 {
				t.Fatal("resolution lost destination")
			}
		})
	}
}
