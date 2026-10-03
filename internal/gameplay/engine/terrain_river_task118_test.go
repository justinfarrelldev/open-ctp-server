package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailRiverTerrainAddsYieldAndLowersMovementCost(t *testing.T) {
	state := State{Tiles: []Tile{{ID: "river", Terrain: TerrainPlains, River: true, Open: true}}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, Food: 10, Production: 10, Gold: 5, RiverFood: 5, RiverProduction: 5, RiverGold: 5, RiverCost: 50}}}}

	yield, err := terrainYieldWithRuntime(state, "river", runtime)
	if err != nil {
		t.Fatal(err)
	}
	if yield != (Yield{Food: 15, Production: 15, Gold: 10}) {
		t.Fatalf("river terrain yield = %+v, want base plus river 15/15/10", yield)
	}
	cost, ok := terrainCost(state, runtime, "river")
	if !ok || cost != 50 {
		t.Fatalf("river movement cost = %d ok=%v, want 50", cost, ok)
	}
}
