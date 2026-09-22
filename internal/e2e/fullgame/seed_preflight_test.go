package fullgame

import (
	"fmt"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/world"
)

func TestMinimalPublicConquestSeedIsLandReachable(t *testing.T) {
	wetness := 50
	seed := int64(seedBase + 3)
	snapshot, _, err := world.GenerateWorld(world.GenerationInput{GameID: 1, PlayerOrder: []int64{1, 3}, GeneratorVersion: "ctp2-first-v1", HostParams: world.GenerationHostParams{Style: "sequential", RulesJSON: "[]", ExclusionsJSON: "{}", StartingSettlers: 2, StartingGold: 100, Seed: &seed, Wetness: &wetness, MapSize: "duel", Climate: "temperate", GeneratorVersion: "ctp2-first-v1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Spawns) != 2 {
		t.Fatalf("spawns = %d, want 2", len(snapshot.Spawns))
	}
	if !landConnected(snapshot) {
		t.Fatalf("seed=%d is not land-connected", seed)
	}
}

func landConnected(snapshot world.WorldSnapshot) bool {
	tiles := make(map[string]world.WorldTile, len(snapshot.Tiles))
	for _, tile := range snapshot.Tiles {
		if tile.Terrain != "" && tile.Terrain != "ocean" && tile.Terrain != "shelf" {
			tiles[fmt.Sprintf("%d,%d", tile.X, tile.Y)] = tile
		}
	}
	start := fmt.Sprintf("%d,%d", snapshot.Spawns[0].X, snapshot.Spawns[0].Y)
	target := fmt.Sprintf("%d,%d", snapshot.Spawns[1].X, snapshot.Spawns[1].Y)
	queue := []string{start}
	seen := map[string]bool{start: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == target {
			return true
		}
		tile, ok := tiles[current]
		if !ok {
			continue
		}
		for _, point := range [][2]int{{tile.X + 1, tile.Y}, {tile.X - 1, tile.Y}, {tile.X, tile.Y + 1}, {tile.X, tile.Y - 1}} {
			id := fmt.Sprintf("%d,%d", point[0], point[1])
			if _, ok := tiles[id]; ok && !seen[id] {
				seen[id] = true
				queue = append(queue, id)
			}
		}
	}
	return false
}
