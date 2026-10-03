package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailEffectiveBuildingBrokerageAmplifiesCommerce(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Starter:              rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
		Production: []rules.RuntimeProduction{
			{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1},
			{ID: "IMPROVE_BROKERAGE", Kind: ProductionBuilding, Cost: 1, Effects: []rules.RuntimeStarterEffect{{ID: "commerce", Kind: EffectYieldGoldPercent, Value: 20}, {ID: "brokerage", Kind: EffectBrokerage}}},
		},
	}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}}
	state.Wonders = []Wonder{
		{ID: "WONDER_EFFECTIVE_BROKERAGE", OwnerID: "player-a", Effects: []Effect{{ID: "IMPROVE_BROKERAGE", Kind: EffectBuildingEverywhere}}},
		{ID: "WONDER_ELECTRONIC_COMMERCE_ACT", OwnerID: "player-a", Effects: []Effect{{ID: "brokerages", Kind: EffectIncreaseBrokerages, Value: 50}}},
	}

	yield, err := cityYield(state, state.Cities[0])
	if err != nil {
		t.Fatal(err)
	}
	if yield.Gold != 3 {
		t.Fatalf("effective brokerage commerce gold = %d, want 3", yield.Gold)
	}
}
