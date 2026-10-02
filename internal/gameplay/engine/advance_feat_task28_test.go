package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestAdvanceTriggerFeatAddsTimedNationEffects(t *testing.T) {
	state := advanceFeatState(t)
	state.Nations[0].ResearchTargetID = "ADVANCE_CONCRETE"

	next := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if !containsString(next.Nations[0].Advances, "ADVANCE_CONCRETE") {
		t.Fatalf("advances after research = %v, want ADVANCE_CONCRETE", next.Nations[0].Advances)
	}
	if !hasEffectValueTurns(next.Nations[0].Effects, EffectYieldProductionPercent, 5, 15) {
		t.Fatalf("nation effects after advance feat = %#v, want 15-turn +5 production percent", next.Nations[0].Effects)
	}

	next = runtimeApply(t, next, "player-a", CommandRoundComplete, struct{}{})
	if !hasEffectValueTurns(next.Nations[0].Effects, EffectYieldProductionPercent, 5, 14) {
		t.Fatalf("nation effects after feat countdown = %#v, want 14 turns remaining", next.Nations[0].Effects)
	}
}

func TestAdvanceFeatReduceCityWallsReducesDefenderModifier(t *testing.T) {
	state := advanceFeatState(t)
	state.Nations[0].Effects = []Effect{{ID: "feat:gunpowder", Kind: EffectReduceCityWalls, Value: -5, Turns: 25}}
	state.Tiles[1].Terrain = TerrainPlains
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "tile-b", Buildings: []Building{{ID: "walls", Effects: []Effect{{ID: "walls-defense", Kind: EffectCombatDefense, Value: 15}, {ID: "walls", Kind: EffectCityWalls, Value: 15}}}}}}
	attacker := Unit{ID: "attacker", OwnerID: "player-a", Attack: 10}
	defender := Unit{ID: "defender", OwnerID: "player-b", TileID: "tile-b", Defense: 10}

	if got := combatDefenseModifier(state, defender, attacker); got != 10 {
		t.Fatalf("defense modifier with wall-reduction feat = %d, want 10", got)
	}
}

func advanceFeatState(t *testing.T) State {
	t.Helper()
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Advances:             []rules.RuntimeAdvanceDefinition{{ID: "ADVANCE_CONCRETE", Cost: 1, TriggerFeatID: "FEAT_ADVANCE_CONCRETE"}},
		Feats:                []rules.RuntimeFeatDefinition{{ID: "FEAT_ADVANCE_CONCRETE", Duration: 15, Effects: []rules.RuntimeStarterEffect{{ID: "feat-production", Kind: EffectYieldProductionPercent, Value: 5}}}},
		Production:           []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
	}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := runtimeStartedState(t)
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}
	return state
}

func hasEffectValueTurns(effects []Effect, kind string, value, turns int) bool {
	for _, effect := range effects {
		if effect.Kind == kind && effect.Value == value && effect.Turns == turns {
			return true
		}
	}
	return false
}
