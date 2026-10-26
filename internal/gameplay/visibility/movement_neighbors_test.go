package visibility

import (
	"reflect"
	"sort"
	"testing"
)

func TestIndexedMovementNeighborsMatchFullOrderedScan(t *testing.T) {
	// Duplicate coordinates, negative coordinates and deliberately shuffled IDs
	// retain the old scanner's exact ID tie ordering, including diagonal steps.
	tiles := []TileProjection{{ID: "z", X: 0, Y: 0}, {ID: "a", X: 1, Y: 1}, {ID: "b", X: 1, Y: 1}, {ID: "d", X: -1, Y: 0}, {ID: "c", X: 2, Y: 0}, {ID: "e", X: 0, Y: -1}, {ID: "f", X: 3, Y: 3}}
	sort.Slice(tiles, func(i, j int) bool { return tiles[i].ID < tiles[j].ID })
	positions := movementPositionIndex(tiles)
	for _, from := range tiles {
		var expected []int
		for index, to := range tiles {
			if adjacentProjection(from, to) {
				expected = append(expected, index)
			}
		}
		got := movementNeighborIndices(from, positions)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("%s: neighbors %v != scan %v", from.ID, got, expected)
		}
	}
}
