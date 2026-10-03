package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailEffectiveBuildingEverywhereAppliesBuildingEconomyEffects(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Starter:              rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1},
		Movement:             rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityProduction: 10}}},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
		Production: []rules.RuntimeProduction{
			{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1},
			{ID: "IMPROVE_FACTORY", Kind: ProductionBuilding, Cost: 1, Effects: []rules.RuntimeStarterEffect{
				{ID: "factory-production", Kind: EffectYieldProductionPercent, Value: 50},
				{ID: "factory-science", Kind: EffectResearchPercent, Value: 50},
			}},
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
		Tiles:   []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Yields: Yield{Production: 10, Gold: 10}}},
		Wonders: []Wonder{{ID: "WONDER_INTERNET", OwnerID: "player-a", Effects: []Effect{{ID: "IMPROVE_FACTORY", Kind: EffectBuildingEverywhere}}}},
	}

	yield, err := cityYieldWithRuntime(state, state.Cities[0], &runtime)
	if err != nil {
		t.Fatal(err)
	}
	if yield.Production != 15 {
		t.Fatalf("effective-building production yield = %d, want 15", yield.Production)
	}
	if got := researchRate(state, state.Nations[0]); got != 16 {
		t.Fatalf("effective-building research rate = %d, want 16", got)
	}

	state.Wonders[0].Effects = append(state.Wonders[0].Effects,
		Effect{ID: "WONDER_INTERNET:factory-production", Kind: EffectYieldProductionPercent, Value: 50},
		Effect{ID: "WONDER_INTERNET:factory-science", Kind: EffectResearchPercent, Value: 50},
	)
	yield, err = cityYieldWithRuntime(state, state.Cities[0], &runtime)
	if err != nil {
		t.Fatal(err)
	}
	if yield.Production != 15 || researchRate(state, state.Nations[0]) != 16 {
		t.Fatalf("direct + effective building effects doubled: yield=%#v research=%d", yield, researchRate(state, state.Nations[0]))
	}
}
