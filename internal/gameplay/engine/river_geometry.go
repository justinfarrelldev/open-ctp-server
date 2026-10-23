package engine

// RiverGeometry is an observed artwork fact of ONE tile, as in original
// TiledMap::PostProcessTile/TryRiver and UnseenCell's stored TileInfo. It is not
// permission to inspect any neighbor. Edges are SW,NW,NE,SE (bits 0..3), mapping
// to server coordinate deltas (0,+1),(-1,0),(0,-1),(+1,0). Zero is a real isolated
// river/source, not unknown connectivity. Nil is unavailable or not a river.
type RiverGeometry struct {
	Version          string `json:"version"`
	Edges            uint8  `json:"edges"`
	Beach            bool   `json:"beach"`
	ObservedRevision int64  `json:"observed_revision"`
}

func (i *Inspector) riverGeometry(tile Tile) *RiverGeometry {
	if !tile.River {
		return nil
	}
	result := &RiverGeometry{Version: "ctp2-river-edges-v1", Beach: riverBeach(EffectiveTerrain(i.state, tile.ID)), ObservedRevision: i.state.Revision}
	for bit, delta := range [][2]int{{0, 1}, {-1, 0}, {0, -1}, {1, 0}} {
		// Ignore an off-map edge when coordinate addition would overflow.
		x, y := tile.X+delta[0], tile.Y+delta[1]
		if delta[0] > 0 && x < tile.X || delta[0] < 0 && x > tile.X || delta[1] > 0 && y < tile.Y || delta[1] < 0 && y > tile.Y {
			continue
		}
		neighbor, ok := i.coordinates[[2]int{x, y}]
		if ok && neighbor.River && !(result.Beach && riverBeach(EffectiveTerrain(i.state, neighbor.ID))) {
			result.Edges |= 1 << bit
		}
	}
	return result
}

func riverBeach(terrain string) bool { return terrain == "beach" || terrain == "TERRAIN_WATER_BEACH" }
