package engine

import "testing"

func TestRiverGeometryAllOriginalEdgePatterns(t *testing.T) {
	for mask := 0; mask < 16; mask++ {
		state := task10State()
		state.Tiles = []Tile{{ID: "center", Terrain: TerrainPlains, River: true}}
		for bit, d := range [][2]int{{0, 1}, {-1, 0}, {0, -1}, {1, 0}} {
			state.Tiles = append(state.Tiles, Tile{ID: string(rune('a' + bit)), X: d[0], Y: d[1], Terrain: TerrainPlains, River: mask&(1<<bit) != 0})
		}
		inspector, err := NewInspector(state)
		if err != nil {
			t.Fatal(err)
		}
		geometry := inspector.Tile("center").RiverGeometry
		if geometry == nil || int(geometry.Edges) != mask || geometry.Beach {
			t.Fatalf("pattern %d: %+v", mask, geometry)
		}
	}
}

func TestRiverGeometryBeachMouthSuppressesBeachToBeachEdges(t *testing.T) {
	state := task10State()
	state.Tiles = []Tile{{ID: "mouth", Terrain: "beach", River: true}, {ID: "land", X: -1, Terrain: TerrainPlains, River: true}, {ID: "beach", X: 1, Terrain: "beach", River: true}}
	inspector, err := NewInspector(state)
	if err != nil {
		t.Fatal(err)
	}
	geometry := inspector.Tile("mouth").RiverGeometry
	if geometry == nil || !geometry.Beach || geometry.Edges != 2 {
		t.Fatalf("mouth: %+v", geometry)
	}
}
