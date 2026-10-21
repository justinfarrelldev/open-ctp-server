package engine

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func withTestRuin(t *testing.T, state State, tileID string, roll, value int) State {
	t.Helper()
	raw, _ := json.Marshal(state)
	var wire map[string]any
	_ = json.Unmarshal(raw, &wire)
	for _, entry := range wire["tiles"].([]any) {
		tile := entry.(map[string]any)
		if tile["id"] == tileID {
			tile["ruin"] = map[string]any{"profile": "villages-only-v1", "type_roll": roll, "value_roll": value}
		}
	}
	raw, _ = json.Marshal(wire)
	_ = json.Unmarshal(raw, &state)
	return state
}

func TestRuinGoldAuthoritativeOnceAndRestartSafe(t *testing.T) {
	state := withTestRuin(t, journeyState(t), "t1", 0, 5000)
	before := canonicalStateHashForTest(t, state)
	next := runtimeApply(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
	nation := nationForID(&next, "nation-a")
	if nation.Gold != 225 {
		t.Fatalf("gold=%d, want released villages-only 150 + 150*5000/10000", nation.Gold)
	}
	if canonicalStateHashForTest(t, state) != before {
		t.Fatal("reward mutated input")
	}
	if len(next.Notifications) != 1 || next.Notifications[0].Kind != "ruin.explored" {
		t.Fatalf("no authoritative reward notice: %+v", next.Notifications)
	}
	replay := runtimeApply(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
	if canonicalStateHashForTest(t, next) != canonicalStateHashForTest(t, replay) {
		t.Fatal("reward rerolled")
	}
	raw, _ := CanonicalStateJSON(next)
	next, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	unit, _ := unitByID(next, "traveller")
	unit.Movement = 2
	next = runtimeApply(t, next, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t0", "t1"}})
	if nationForID(&next, "nation-a").Gold != 225 || len(next.Notifications) != 1 {
		t.Fatal("ruin paid twice")
	}
}

func TestRuinEmptyAndCityAndIntermediateTiles(t *testing.T) {
	for _, tc := range []struct {
		name string
		roll int
		city bool
	}{{"empty", 99, false}, {"city", 70, true}} {
		t.Run(tc.name, func(t *testing.T) {
			state := withTestRuin(t, journeyState(t), "t1", tc.roll, 0)
			state.Units[0].Movement = 3
			state = runtimeApply(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1", "t2"}})
			city, exists := cityByTile(state, "t1")
			if exists != tc.city {
				t.Fatalf("city reward existence=%v", exists)
			}
			if exists && (city.OwnerID != "player-a" || city.Population < 1) {
				t.Fatal("invalid reward city")
			}
			if len(state.Notifications) != 1 {
				t.Fatal("intermediate ruin not explored")
			}
		})
	}
}

func TestRuinUnitsUsePinnedDefinitionsAndFullCreationMovement(t *testing.T) {
	for _, roll := range []int{45, 75} {
		t.Run(fmt.Sprint(roll), func(t *testing.T) {
			state := journeyState(t)
			runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
			if err != nil {
				t.Fatal(err)
			}
			runtime.Units = []rules.RuntimeUnitDefinition{
				{ID: "mercenary", Domain: DomainLand, Movement: 4, Attack: 2, Defense: 3, Health: 9},
				{ID: "settler", Domain: DomainLand, Movement: 2, Health: 5, Effects: []rules.RuntimeStarterEffect{{ID: "settler", Kind: EffectSettler}}},
			}
			definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
			if err != nil {
				t.Fatal(err)
			}
			state.Rules = RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: definitions}
			state = withTestRuin(t, state, "t1", roll, 0)
			state = runtimeApply(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
			unit, ok := unitByID(state, "ruin-unit:t1")
			if !ok || unit.OwnerID != "player-a" || unit.Movement != unit.MaxMovement || unit.Health != unit.MaxHealth {
				t.Fatalf("invalid created reward unit: %+v", unit)
			}
			if roll == 45 && (unit.TypeID != "mercenary" || unit.Attack != 2) {
				t.Fatal("mercenary not pinned")
			}
			if roll == 75 && !hasEffect(unit.Effects, EffectSettler) {
				t.Fatal("nomad not a settler")
			}
		})
	}
}

func TestRuinAdvanceExecutesFeatAndResearchLifecycle(t *testing.T) {
	state := journeyState(t)
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Governments = nil
	runtime.TerrainImprovements = nil
	runtime.Advances = []rules.RuntimeAdvanceDefinition{{ID: "ROOT", Cost: 10, TriggerFeatID: "FEAT"}, {ID: "NEXT", Cost: 20, Prerequisites: []string{"ROOT"}}}
	runtime.Feats = []rules.RuntimeFeatDefinition{{ID: "FEAT", Duration: 4, Effects: []rules.RuntimeStarterEffect{{ID: "happy", Kind: EffectHappiness, Value: 5}}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	nation := nationForID(&state, "nation-a")
	nation.ResearchTargetID = "ROOT"
	nation.ResearchGoalID = "ROOT"
	nation.ResearchProgress = 3
	state = withTestRuin(t, state, "t1", 30, 0)
	state = runtimeApply(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
	nation = nationForID(&state, "nation-a")
	if !containsString(nation.Advances, "ROOT") || nation.ResearchTargetID != "NEXT" || nation.ResearchProgress != 0 || nation.ResearchGoalID != "" || !containsString(state.AchievedFeatIDs, "FEAT") {
		t.Fatalf("grant skipped advance lifecycle: %+v", nation)
	}
}

func TestRuinRejectedPathCannotConsumeOrReward(t *testing.T) {
	state := withTestRuin(t, journeyState(t), "t1", 0, 0)
	before := canonicalStateHashForTest(t, state)
	if err := runtimeApplyExpectError(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1", "t7"}}); err == nil {
		t.Fatal("invalid path accepted")
	}
	if canonicalStateHashForTest(t, state) != before {
		t.Fatal("rejected movement consumed ruin")
	}
}
