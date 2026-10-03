package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailEffectiveBuildingProtectionBlocksNuclearAttack(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Starter:              rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
		Production: []rules.RuntimeProduction{
			{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1},
			{ID: "IMPROVE_FORCE_FIELD", Kind: ProductionBuilding, Cost: 1, Effects: []rules.RuntimeStarterEffect{{ID: "force-field", Kind: EffectForceField}}},
		},
	}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := State{
		Rules:   RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions},
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}},
		Wonders: []Wonder{{ID: "WONDER_FORCE_FIELDS", OwnerID: "player-a", Effects: []Effect{{ID: "IMPROVE_FORCE_FIELD", Kind: EffectBuildingEverywhere}}}},
	}

	if !cityProtectedFromNuclearAttack(state, state.Cities[0]) {
		t.Fatalf("effective force-field building did not protect city from nuclear attack")
	}
}
