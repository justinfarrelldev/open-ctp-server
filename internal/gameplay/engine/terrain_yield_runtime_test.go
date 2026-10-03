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

func TestRetailBuildingMaxPopulationCapsFoodGrowth(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Cities[0].Population = 14
	state.Cities[0].StoredFood = cityFoodGrowthThreshold(14)
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_AQUEDUCT", Effects: []Effect{{ID: "aqueduct-max-population", Kind: EffectMaxPopulation, Value: 14}}}}
	state.Phase.Turn = state.RecalculatedTurn + 1
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityFood: 40}}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "settler", Kind: EffectSettler}}}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if next.Cities[0].Population != 14 {
		t.Fatalf("population = %d, want capped at 14", next.Cities[0].Population)
	}
}

func TestRetailOvercrowdingEffectsAccumulateFromBuildings(t *testing.T) {
	state := State{}
	city := City{ID: "city-a", Buildings: []Building{{ID: "IMPROVE_AQUEDUCT", Effects: []Effect{{ID: "aqueduct", Kind: EffectOvercrowdingLevel, Value: 6}}}, {ID: "IMPROVE_ARCOLOGIES", Effects: []Effect{{ID: "arcologies", Kind: EffectOvercrowdingLevel, Value: 8}}}}}
	if got := cityOvercrowdingLevel(state, city); got != 14 {
		t.Fatalf("city overcrowding level = %d, want summed building levels", got)
	}
}

func TestRetailTradeRouteWonderGoldDoesNotInflateSourceCityYield(t *testing.T) {
	state := State{
		Players:     []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}},
		Nations:     []Nation{{ID: "nation-a"}, {ID: "nation-b"}},
		Tiles:       []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS", Open: true}, {ID: "tile-b", Terrain: "TERRAIN_PLAINS", Open: true}},
		Cities:      []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}, {ID: "city-b", OwnerID: "player-b", TileID: "tile-b"}},
		Wonders:     []Wonder{{ID: "WONDER_EAST_INDIA_COMPANY", OwnerID: "player-a", Scope: "nation", Effects: []Effect{{ID: "water", Kind: EffectGoldPerWaterTradeRoute, Value: 15}, {ID: "international", Kind: EffectGoldPerInternationalTradeRoute, Value: 25}}}},
		TradeRoutes: []TradeRoute{{ID: "route-a-b", SourceCityID: "city-a", DestinationCityID: "city-b", CrossesWater: true}},
	}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Gold: 3}}}}
	yield, err := cityYieldWithRuntime(state, state.Cities[0], runtime)
	if err != nil {
		t.Fatal(err)
	}
	if yield.Gold != 3 {
		t.Fatalf("trade route wonder gold city yield = %d, want base 3 only", yield.Gold)
	}
}

func TestRetailWonderGoldPerTelevisionDoesNotInflateCityYield(t *testing.T) {
	state := State{Players: []Player{{ID: "player-a", NationID: "nation-a"}}, Nations: []Nation{{ID: "nation-a"}}, Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS", Open: true}}, Cities: []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_TELEVISION", Effects: []Effect{{ID: "tv", Kind: EffectTelevision}}}}}}, Wonders: []Wonder{{ID: "WONDER_MEDIA", OwnerID: "player-a", Scope: "owner", Effects: []Effect{{ID: "tv-gold", Kind: EffectGoldPerTelevision, Value: 7}}}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Gold: 3}}}}

	yield, err := cityYieldWithRuntime(state, state.Cities[0], runtime)
	if err != nil {
		t.Fatal(err)
	}
	if yield.Gold != 3 {
		t.Fatalf("gold = %d, want base 3 only", yield.Gold)
	}
}

func TestRetailCrimePercentReducesProductionAndGold(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS", Open: true}}, Cities: []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", CrimePercent: 25}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Production: 8, Gold: 12, Food: 4}}}}

	yield, err := cityYieldWithRuntime(state, state.Cities[0], runtime)
	if err != nil {
		t.Fatal(err)
	}
	if yield.Food != 4 || yield.Production != 6 || yield.Gold != 9 {
		t.Fatalf("yield = %+v, want crime to reduce production and gold only", yield)
	}
}

func TestRetailBuildingUpkeepReducesGoldNotProduction(t *testing.T) {
	state := State{Players: []Player{{ID: "player-a", NationID: "nation-a"}}, Nations: []Nation{{ID: "nation-a"}}, Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS", Open: true}}, Cities: []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_FACTORY", Upkeep: 3}}}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Production: 5, Gold: 5}}}}

	yield, err := cityYieldWithRuntime(state, state.Cities[0], runtime)
	if err != nil {
		t.Fatal(err)
	}
	if yield.Production != 5 || yield.Gold != 2 {
		t.Fatalf("yield = %+v, want upkeep to reduce gold only", yield)
	}
}

func TestRetailFoodVatReducesFoodConsumption(t *testing.T) {
	city := City{Population: 4, Yields: Yield{Food: 4}, Effects: []Effect{{ID: "vat", Kind: EffectFoodVatPercent, Value: 50}}}

	advanceCityFood(&city, 0, 0, foodVatPercent(State{}, city))

	if city.Population != 4 || city.StoredFood != 0 {
		t.Fatalf("food vat city = %+v", city)
	}
}

func TestRetailFoodVatReducesFoodConsumptionFromBuildings(t *testing.T) {
	state := State{Players: []Player{{ID: "player-a", NationID: "nation-a"}}, Nations: []Nation{{ID: "nation-a"}}}
	city := City{ID: "city-a", OwnerID: "player-a", Population: 4, Yields: Yield{Food: 4}, Buildings: []Building{{ID: "IMPROVE_FOOD_VAT", Effects: []Effect{{ID: "vat", Kind: EffectFoodVatPercent, Value: 50}}}}}

	advanceCityFood(&city, 0, 0, foodVatPercent(state, city))

	if city.Population != 4 || city.StoredFood != 0 {
		t.Fatalf("building food vat city = %+v", city)
	}
}

func TestRetailStarvationProtectionDelaysPopulationLoss(t *testing.T) {
	city := City{ID: "city-a", Population: 2, Yields: Yield{Food: 0}}
	advanceCityFood(&city, 0, 2, 0)
	if city.Population != 2 || city.StarvationTurns != 1 {
		t.Fatalf("after first starvation city = %+v, want protected with one starvation turn recorded", city)
	}
	advanceCityFood(&city, 0, 2, 0)
	if city.Population != 2 || city.StarvationTurns != 2 {
		t.Fatalf("after second starvation city = %+v, want protection limit reached", city)
	}
	advanceCityFood(&city, 0, 2, 0)
	if city.Population != 1 {
		t.Fatalf("after third starvation population = %d, want 1", city.Population)
	}
}

func TestRetailHappinessEffectsAccumulateFromBuildingsAndWonders(t *testing.T) {
	state := State{Players: []Player{{ID: "player-a", NationID: "nation-a"}}, Nations: []Nation{{ID: "nation-a"}}}
	city := City{ID: "city-a", OwnerID: "player-a", Buildings: []Building{{ID: "IMPROVE_ARENA", Effects: []Effect{{ID: "arena", Kind: EffectHappiness, Value: 2}}}}}
	state.Wonders = []Wonder{{ID: "WONDER_GUTTENBERGS_BIBLE", OwnerID: "player-a", Scope: "owner", Effects: []Effect{{ID: "bible", Kind: EffectHappiness, Value: 2}}}}
	if got := cityHappiness(state, city); got != 79 {
		t.Fatalf("city happiness = %d, want base + building + wonder", got)
	}
}

func TestRetailCrimeEffectsAccumulateFromBuildingsAndWonders(t *testing.T) {
	state := State{Players: []Player{{ID: "player-a", NationID: "nation-a"}}, Nations: []Nation{{ID: "nation-a"}}}
	city := City{ID: "city-a", OwnerID: "player-a", Buildings: []Building{{ID: "IMPROVE_COURTHOUSE", Effects: []Effect{{ID: "courthouse", Kind: EffectCrimePercent, Value: -10}}}}}
	state.Wonders = []Wonder{{ID: "WONDER_ZERO_CRIME_BILL", OwnerID: "player-a", Scope: "owner", Effects: []Effect{{ID: "zero-crime", Kind: EffectCrimePercent, Value: -60}}}}
	if got := cityCrimePercent(state, city); got != -70 {
		t.Fatalf("city crime percent = %d, want building + wonder reducers", got)
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

func TestRuntimeGoldPerCitizenEffectUsesPopulation(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 3, Buildings: []Building{{ID: "IMPROVE_TELEVISION", Effects: []Effect{{ID: "tv-gold", Kind: EffectYieldGoldPerPopulation, Value: 5}}}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Gold: 10}}}}
	yield, err := cityYieldWithRuntime(state, city, runtime)
	if err != nil {
		t.Fatalf("cityYieldWithRuntime() error = %v", err)
	}
	if yield.Gold != 25 {
		t.Fatalf("gold yield = %d, want base 10 + 5*population 3", yield.Gold)
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
