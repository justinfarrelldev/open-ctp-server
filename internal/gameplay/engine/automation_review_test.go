package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetainedMovementDoesNotWaitForeverForImpossibleTerrainCost(t *testing.T) {
	state := journeyState(t)
	for i := range state.Tiles {
		if state.Tiles[i].ID == "t1" {
			state.Tiles[i].Terrain = TerrainForest
		}
	}
	state = runtimeApply(t, state, "player-a", CommandSetMovementOrder, map[string]any{"unit_id": "traveller", "target_x": 7, "target_y": 0})
	if len(state.MovementOrders) != 0 {
		t.Fatal("retained permanently unaffordable step")
	}
}

func TestRetainedMovementArrivalInZOCIsCompletion(t *testing.T) {
	state := journeyState(t)
	state.Units = append(state.Units, Unit{ID: "enemy", OwnerID: "player-b", TileID: "t2", Domain: DomainLand, ZOC: true})
	state = runtimeApply(t, state, "player-a", CommandSetMovementOrder, map[string]any{"unit_id": "traveller", "target_x": 1, "target_y": 0})
	if len(state.MovementOrders) != 0 || len(state.Notifications) != 1 || state.Notifications[0].Kind != "movement_order.completed" {
		t.Fatal("arrival reported as obstacle stop")
	}
}

func TestRuinNomadsUseStarterInsteadOfAlphabeticallyEarlierFutureSettler(t *testing.T) {
	state := journeyState(t)
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Units = []rules.RuntimeUnitDefinition{
		{ID: "ENGINEER", Domain: DomainLand, Movement: 4, Health: 20, Effects: []rules.RuntimeStarterEffect{{ID: "settle", Kind: EffectSettler}}},
		{ID: "SETTLER", Domain: DomainLand, Movement: 2, Health: 5, Effects: []rules.RuntimeStarterEffect{{ID: "settle", Kind: EffectSettler}}},
	}
	runtime.Starter.TypeID = "SETTLER"
	defs, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: defs}
	nationForID(&state, "nation-a").Effects = []Effect{{ID: "hp", Kind: EffectIncreaseHP, Value: 3}}
	state = withTestRuin(t, state, "t1", 75, 0)
	state = runtimeApply(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
	unit, _ := unitByID(state, "ruin-unit:t1")
	if unit == nil || unit.TypeID != "SETTLER" || unit.Health != 8 || unit.MaxHealth != 8 {
		t.Fatalf("invalid nomads: %+v", unit)
	}
}

func TestFoundingCityRemovesUnopenedRuinWithoutAwardingReward(t *testing.T) {
	state := journeyState(t)
	state = withTestRuin(t, state, "t0", 0, 0)
	state.Units[0].Effects = []Effect{{ID: "settle", Kind: EffectSettler}}
	state = runtimeApply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "traveller", CityID: "home", Name: "Home"})
	tile, _ := tileByID(state, "t0")
	if tile.Ruin != nil || nationForID(&state, "nation-a").Gold != 0 {
		t.Fatal("city did not remove ruin without reward")
	}
}
