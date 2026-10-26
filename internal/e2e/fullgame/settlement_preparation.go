package fullgame

import (
	"encoding/json"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

// Prepare the scripted three-city opener from public facts. Distinct tiles are
// not enough: the original game also prohibits adjacent city centers.
func prepareContinuityCitySites(public visibility.Projection, units []visibility.UnitProjection) ([]ScriptCommand, bool, error) {
	tiles := make(map[string]visibility.TileProjection, len(public.Tiles))
	candidates := make([]visibility.TileProjection, 0, len(public.Tiles))
	for _, tile := range public.Tiles {
		if !tile.Stale {
			tiles[tile.ID] = tile
			if tile.Planning != nil && tile.Planning.Accessible {
				candidates = append(candidates, tile)
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	used := make(map[string]bool, len(units))
	for _, unit := range units {
		used[unit.TileID] = true
	}
	var centers []engine.Tile
	var commands []ScriptCommand
	preparing := false
	for _, unit := range units {
		current := continuityCenter(tiles[unit.TileID])
		if visibility.FoundingSiteAvailable(public, unit.TileID) && engine.CitySiteAvailable(current, centers) {
			centers = append(centers, current)
			continue
		}
		preparing = true
		var chosen []string
		var destination engine.Tile
		bestTransitDistance := -1
		for _, tile := range candidates {
			if used[tile.ID] {
				continue
			}
			path, err := visibility.Plan(public, unit.ID, tile.ID)
			if err != nil || len(path) == 0 {
				continue
			}
			center := continuityCenter(tile)
			if visibility.FoundingSiteAvailable(public, tile.ID) && engine.CitySiteAvailable(center, centers) {
				chosen, destination = path, center
				break
			}
			// A first step may be needed to reveal a site beyond initial sight.
			// Prefer transit that spreads the settlers, not a direct unseen hop.
			if distance := continuitySiteSeparation(center, centers); distance > bestTransitDistance {
				chosen, destination, bestTransitDistance = path, center, distance
			}
		}
		if len(chosen) == 0 {
			// Exhausted movers wait for a real turn refresh.
			continue
		}
		payload, err := json.Marshal(map[string]any{"unit_id": unit.ID, "path": chosen})
		if err != nil {
			return nil, false, err
		}
		commands = append(commands, ScriptCommand{Command: "move", ActorRole: "human-1", Payload: payload})
		used[destination.ID] = true
		// Reserve even a transit endpoint so the other mover does not converge.
		centers = append(centers, destination)
	}
	if preparing && len(commands) == 0 {
		closeout := "pass"
		if public.PlanningStatus == visibility.PlanningStatusOpen || public.PlanningStatus == visibility.PlanningStatusQueued {
			closeout = "round-complete"
		}
		commands = []ScriptCommand{{Command: closeout, ActorRole: "human-1", Payload: json.RawMessage(`{}`)}}
	}
	return commands, preparing, nil
}

func continuityCenter(tile visibility.TileProjection) engine.Tile {
	return engine.Tile{ID: tile.ID, X: tile.X, Y: tile.Y}
}

func continuitySiteSeparation(site engine.Tile, centers []engine.Tile) int {
	distance := int(^uint(0) >> 1)
	for _, center := range centers {
		dx, dy := site.X-center.X, site.Y-center.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		distance = min(distance, max(dx, dy))
	}
	return distance
}
