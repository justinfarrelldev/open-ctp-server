package visibility

import (
	"math"
	"sort"
)

// Build coordinates once per search rather than copying/scanning every exposed
// tile for every expanded A* node. Only candidate enumeration changes: original
// ID ordering, heuristic, route cost, affordability, ZOC and ties stay identical.
// Slices retain duplicate-coordinate IDs so even such projections preserve the
// old scanner's semantics. This is viewer data only, never canonical geography.
func movementPositionIndex(ordered []TileProjection) map[[2]int][]int {
	positions := make(map[[2]int][]int, len(ordered))
	for index, tile := range ordered {
		key := [2]int{tile.X, tile.Y}
		positions[key] = append(positions[key], index)
	}
	return positions
}
func movementNeighborIndices(from TileProjection, positions map[[2]int][]int) []int {
	var neighbors []int
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if dx < 0 && from.X == math.MinInt || dx > 0 && from.X == math.MaxInt || dy < 0 && from.Y == math.MinInt || dy > 0 && from.Y == math.MaxInt {
				continue
			}
			neighbors = append(neighbors, positions[[2]int{from.X + dx, from.Y + dy}]...)
		}
	}
	// The positions index refers to the original ID-sorted tile array, so integer
	// ordering exactly reproduces the old full scan's candidate tie order.
	sort.Ints(neighbors)
	return neighbors
}
