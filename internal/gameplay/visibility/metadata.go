package visibility

import (
	"math"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

// ContractProjection versions the projection independently of persisted state.
// Features describe supported contracts, not permission to bypass action/phase
// validation. False feature decisions must not be emulated by local mutations.
type ContractProjection struct {
	Version  int             `json:"version"`
	Features map[string]bool `json:"features"`
}

func contractProjection(projection Projection) *ContractProjection {
	movement := projection.Rules != nil && projection.Rules.Movement != nil && projection.Rules.Movement.Configured
	governor := projection.Rules != nil && len(projection.Rules.GovernorPriorities) > 0
	return &ContractProjection{Version: 2, Features: map[string]bool{
		"fog_metadata": true, "tile_inspection": true, "atomic_production_queue": true,
		"city_orders": true, "retained_sleep": true, "research_goals": true,
		"coordinate_move_intents": movement, "calendar": projection.Calendar != nil,
		"production_notifications": true, "own_statistics_history": true,
		"retained_movement_orders": movement, "governor": governor, "policy_optimizer": false,
		"city_culture_style": false, "ruin_rewards": movement, "foreign_statistics_history": false,
	}}
}

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
	width, widthOK := worldDimension(minX, maxX)
	height, heightOK := worldDimension(minY, maxY)
	if !widthOK || !heightOK {
		return nil
	}
	return &WorldProjection{MinX: minX, MinY: minY, Width: width, Height: height}
}

func worldDimension(low, high int) (int, bool) {
	if high < low {
		return 0, false
	}
	// Unsigned subtraction represents the full nonnegative signed-coordinate
	// distance without overflowing. Check before converting/adding the endpoint.
	span := uint64(high) - uint64(low)
	if span >= uint64(math.MaxInt) {
		return 0, false
	}
	return int(span) + 1, true
}

func ruinPresence(tile engine.Tile) *bool {
	present := tile.Ruin != nil
	return &present
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
