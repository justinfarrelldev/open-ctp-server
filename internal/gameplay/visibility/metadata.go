package visibility

import "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"

// WorldProjection describes the public playable coordinate graph, not hidden
// terrain or generator inputs. Current engine adjacency does not wrap either axis.
type WorldProjection struct {
	MinX   int  `json:"min_x"`
	MinY   int  `json:"min_y"`
	Width  int  `json:"width"`
	Height int  `json:"height"`
	WrapX  bool `json:"wrap_x"`
	WrapY  bool `json:"wrap_y"`
}

func worldProjection(state engine.State) *WorldProjection {
	if len(state.Tiles) == 0 {
		return nil
	}
	minX, maxX, minY, maxY := state.Tiles[0].X, state.Tiles[0].X, state.Tiles[0].Y, state.Tiles[0].Y
	for _, tile := range state.Tiles {
		if tile.X < minX {
			minX = tile.X
		}
		if tile.X > maxX {
			maxX = tile.X
		}
		if tile.Y < minY {
			minY = tile.Y
		}
		if tile.Y > maxY {
			maxY = tile.Y
		}
	}
	return &WorldProjection{MinX: minX, MinY: minY, Width: maxX - minX + 1, Height: maxY - minY + 1}
}

// applyInspection never reads authority: stale callers pass stored observations.
func applyInspection(tile *TileProjection, facts *engine.TileInspection) {
	if facts == nil {
		return
	}
	tile.Yields = copyValue(facts.Yields)
	tile.YieldContext = facts.YieldContext
	tile.MovementCost = copyValue(facts.MovementCost)
	tile.Pollution = facts.Pollution
	for _, overlay := range facts.Overlays {
		tile.Overlays = append(tile.Overlays, TileOverlayProjection{Kind: overlay.Kind, ImprovementID: overlay.ImprovementID, Remaining: overlay.Remaining})
	}
}

// copyValue preserves unknown/known distinctions while preventing projected
// pointers from aliasing persistent observations or canonical state.
func copyValue[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

// Population zero is omitted by the existing wire schema; nil legacy observations
// therefore remain unavailable instead of being reconstructed from hidden cities.
func knownPopulation(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
