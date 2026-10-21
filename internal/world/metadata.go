package world

import (
	"sort"
	"strings"
)

// RiverGoodsGeneratorVersion introduces independent river/resource placement.
// Older generator versions deliberately retain their exact terrain-only output.
// This is a server generator, not a reproduction of the retail RNG stream.
const RiverGoodsGeneratorVersion = "ctp2-rivers-goods-v2"

// generateRiverGoods follows the retail separation of terrain, rivers and goods
// (wldgen.cpp NewGenerateRivers/GenerateGoods). River drainage strictly descends
// distance-to-water: finite paths cannot cycle, dead-end inland, or mark oceans.
// The current gameplay graph is bounded/cardinal, even though legacy spawn
// scoring uses wrapped distances. No river connection crosses that boundary.
func generateRiverGoods(tiles []WorldTile, profile mapProfile, wetness int, rng SplitMix64RNG) {
	distance := make([]int, len(tiles))
	queue := make([]int, 0, len(tiles))
	for i, tile := range tiles {
		distance[i] = -1
		if !isLandTerrain(tile.Terrain) {
			distance[i] = 0
			queue = append(queue, i)
		}
		// Retail terrain Resources use canonical semantic IDs; placement is server
		// state, never inferred from client artwork. Separate RNG domains prevent
		// changing goods from perturbing drainage or the spawn-selection stream.
		value := coordinateNoise(rng.Split("goods").Seed(), tile.X, tile.Y)
		if value < 1.0/49.0 {
			terrain := strings.ToUpper(tile.Terrain)
			if tile.Terrain == "ocean" {
				terrain = "WATER_DEEP"
			}
			if tile.Terrain == "shelf" {
				terrain = "WATER_SHELF"
			}
			suffixes := []string{"ONE", "TWO", "THREE", "FOUR"}
			variant := int(coordinateNoise(rng.Split("good-kind").Seed(), tile.X, tile.Y) * 4)
			tiles[i].GoodID = "TERRAIN_" + terrain + "_GOOD_" + suffixes[variant]
		}
	}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		for _, next := range cardinalNeighbors(current, profile.width, profile.height) {
			if distance[next] < 0 {
				distance[next] = distance[current] + 1
				queue = append(queue, next)
			}
		}
	}
	sources := make([]int, 0, len(tiles))
	for i := range tiles {
		if distance[i] > 1 {
			sources = append(sources, i)
		}
	}
	sort.Slice(sources, func(i, j int) bool {
		a, b := sources[i], sources[j]
		if distance[a] != distance[b] {
			return distance[a] > distance[b]
		}
		an := coordinateNoise(rng.Seed(), tiles[a].X, tiles[a].Y)
		bn := coordinateNoise(rng.Seed(), tiles[b].X, tiles[b].Y)
		if an != bn {
			return an < bn
		}
		return a < b
	})
	count := len(tiles) * wetness / 12000
	if wetness > 0 && count == 0 {
		count = 1
	}
	for _, source := range sources {
		if count == 0 {
			break
		}
		if tiles[source].River {
			continue
		}
		count--
		current := source
		for distance[current] > 0 && !tiles[current].River {
			tiles[current].River = true
			for _, next := range cardinalNeighbors(current, profile.width, profile.height) {
				if distance[next] == distance[current]-1 {
					current = next
					break
				}
			}
		}
	}
}

// cardinalNeighbors returns a stable ordered subset of the bounded grid graph.
func cardinalNeighbors(index, width, height int) []int {
	neighbors := make([]int, 0, 4)
	x, y := index%width, index/width
	if x > 0 {
		neighbors = append(neighbors, index-1)
	}
	if y > 0 {
		neighbors = append(neighbors, index-width)
	}
	if x+1 < width {
		neighbors = append(neighbors, index+1)
	}
	if y+1 < height {
		neighbors = append(neighbors, index+width)
	}
	return neighbors
}
