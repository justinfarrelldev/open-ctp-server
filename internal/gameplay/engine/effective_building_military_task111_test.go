package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailEffectiveBuildingMilitaryEffectsModifyCombat(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Starter:              rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
		Production: []rules.RuntimeProduction{
			{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1},
			{ID: "IMPROVE_BALLISTA_TOWERS", Kind: ProductionBuilding, Cost: 1, Effects: []rules.RuntimeStarterEffect{{ID: "ballista", Kind: EffectCombatAttackLand, Value: 20}}},
			{ID: "IMPROVE_CITY_WALLS", Kind: ProductionBuilding, Cost: 1, Effects: []rules.RuntimeStarterEffect{{ID: "walls-defense", Kind: EffectCombatDefense, Value: 15}, {ID: "walls", Kind: EffectCityWalls, Value: 15}}},
		},
	}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := combatState()
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Cities = append(state.Cities, City{ID: "city-a", OwnerID: "player-a", TileID: "a"})
	state.Wonders = []Wonder{
		{ID: "WONDER_ATTACKS", OwnerID: "player-a", Effects: []Effect{{ID: "IMPROVE_BALLISTA_TOWERS", Kind: EffectBuildingEverywhere}}},
		{ID: "WONDER_WALLS", OwnerID: "player-b", Effects: []Effect{{ID: "IMPROVE_CITY_WALLS", Kind: EffectBuildingEverywhere}}},
	}

	attacker := state.Units[0]
	defender := state.Units[2]
	defender.Domain = DomainLand
	if got := combatAttackModifier(state, attacker, defender); got != 21 { // nation banner + effective Ballista Towers.
		t.Fatalf("effective-building land attack modifier = %d, want 21", got)
	}

	attacker.Effects = []Effect{{ID: "ignore-walls", Kind: EffectIgnoreCityWalls}}
	if got := combatDefenseModifier(state, defender, attacker); got != 3 { // fortified + forest + direct city marker; effective wall defense is ignored.
		t.Fatalf("effective-building ignored city-wall defense modifier = %d, want 3", got)
	}
}
