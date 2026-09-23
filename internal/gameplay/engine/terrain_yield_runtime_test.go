package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailRuntimeTerrainYieldsAccumulateProduction(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Queue: []ProductionItem{{ID: "UNIT_SLOW", Kind: ProductionUnit, Cost: 99}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Food: 10, Production: 10, Gold: 5}}}}
	if err := completeProduction(&state, &city, runtime); err != nil {
		t.Fatalf("completeProduction() error = %v", err)
	}
	if city.StoredProduction != 10 {
		t.Fatalf("stored production = %d, want retail terrain production 10", city.StoredProduction)
	}
}

func TestRetailCityProductionUsesExistingCityYieldEffects(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_FACTORY", Effects: []Effect{{ID: "factory-production", Kind: EffectYieldProductionPercent, Value: 50}}}}, Queue: []ProductionItem{{ID: "UNIT_SLOW", Kind: ProductionUnit, Cost: 99}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Production: 10}}}}
	if err := completeProduction(&state, &city, runtime); err != nil {
		t.Fatalf("completeProduction() error = %v", err)
	}
	if city.StoredProduction != 15 {
		t.Fatalf("stored production = %d, want base 10 + factory 50%%", city.StoredProduction)
	}
}

func TestRetailCityFoodSurplusGrowsPopulation(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Cities[0].Population = 1
	state.Cities[0].StoredFood = 95
	state.Phase.Turn = state.RecalculatedTurn + 1
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityFood: 10}}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "settler", Kind: EffectSettler}}}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Population != 2 || next.Cities[0].StoredFood != 3 {
		t.Fatalf("city population/stored food = %d/%d, want 2/3", next.Cities[0].Population, next.Cities[0].StoredFood)
	}
}

func TestRetailRuntimeTerrainYieldsDriveCityYield(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Food: 10, Production: 10, Gold: 5, CityFood: 10, CityProduction: 15, CityGold: 15}}}}
	yield, err := cityYieldWithRuntime(state, city, runtime)
	if err != nil {
		t.Fatalf("cityYieldWithRuntime() error = %v", err)
	}
	if yield.Food != 10 || yield.Production != 15 || yield.Gold != 15 {
		t.Fatalf("retail city terrain yield = %+v, want EnvCity 10/15/15", yield)
	}
}

func TestRetailRuntimeBuildingFoodPercentDrivesCityYield(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_GRANARY", Effects: []Effect{{ID: "granary-food", Kind: EffectYieldFoodPercent, Value: 10}}}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Food: 10}}}}
	yield, err := cityYieldWithRuntime(state, city, runtime)
	if err != nil {
		t.Fatalf("cityYieldWithRuntime() error = %v", err)
	}
	if yield.Food != 11 {
		t.Fatalf("food yield = %d, want base 10 + 10%% bonus rounded up", yield.Food)
	}
}

func TestRuntimeBuildingEffectsObsoleteWhenNationHasObsoleteAdvance(t *testing.T) {
	state := State{
		Tiles:   []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}},
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_OBSOLETE"}}},
	}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_OLD", ObsoletedBy: "ADVANCE_OBSOLETE", Effects: []Effect{{ID: "old-production", Kind: EffectYieldProduction, Value: 99}}}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Production: 10}}}}
	yield, err := cityYieldWithRuntime(state, city, runtime)
	if err != nil {
		t.Fatalf("cityYieldWithRuntime() error = %v", err)
	}
	if yield.Production != 10 {
		t.Fatalf("production yield = %d, want obsolete building ignored", yield.Production)
	}
}

func TestRetailRuntimeWonderEffectsDriveOwnedCityYields(t *testing.T) {
	state := State{
		Tiles:   []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}, {ID: "tile-b", Terrain: "TERRAIN_PLAINS"}},
		Wonders: []Wonder{{ID: "WONDER_PYRAMIDS", OwnerID: "player-a", CityID: "city-a", Effects: []Effect{{ID: "pyramids-gold", Kind: EffectYieldGold, Value: 100}}}, {ID: "WONDER_GENOME_PROJECT", OwnerID: "player-a", CityID: "city-a", Effects: []Effect{{ID: "genome-production", Kind: EffectYieldProductionPercent, Value: 10}}}},
	}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Production: 10, Gold: 5}}}}
	yield, err := cityYieldWithRuntime(state, City{ID: "city-b", OwnerID: "player-a", TileID: "tile-b"}, runtime)
	if err != nil {
		t.Fatalf("cityYieldWithRuntime() error = %v", err)
	}
	if yield.Production != 11 || yield.Gold != 105 {
		t.Fatalf("wonder-modified yield = %+v, want production 11 and gold 105", yield)
	}
}
