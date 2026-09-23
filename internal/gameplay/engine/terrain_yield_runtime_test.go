package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

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
