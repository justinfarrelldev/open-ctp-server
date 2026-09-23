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

func TestRetailRuntimeTerrainYieldsDriveCityYield(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "tile-a", Terrain: "TERRAIN_PLAINS"}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_PLAINS", Cost: 100, Domains: []string{DomainLand}, Food: 10, Production: 10, Gold: 5}}}}
	yield, err := cityYieldWithRuntime(state, city, runtime)
	if err != nil {
		t.Fatalf("cityYieldWithRuntime() error = %v", err)
	}
	if yield.Food != 10 || yield.Production != 10 || yield.Gold != 5 {
		t.Fatalf("retail terrain yield = %+v, want 10/10/5", yield)
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
