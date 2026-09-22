package fullgame

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

var ErrNoVisibleFrontier = errors.New("no visible legal frontier")
var ErrExplorationBudget = errors.New("visible exploration action budget exhausted")

type FrontierExplorer struct {
	ActionBudget int
	Visited      map[string]bool
}

type ExplorationDiscovery struct {
	EnemyUnitID string
	EnemyTileID string
	Moves       int
	Projection  []byte
}

type ExplorationTransport interface {
	Projection(context.Context) ([]byte, error)
	Move(context.Context, string, []string) ([]byte, error)
}

func FixedFrontierExplorer(actionBudget int) FrontierExplorer {
	return FrontierExplorer{ActionBudget: actionBudget, Visited: make(map[string]bool)}
}

func (explorer FrontierExplorer) Explore(ctx context.Context, transport ExplorationTransport, unitID string) (ExplorationDiscovery, error) {
	if transport == nil || unitID == "" || explorer.ActionBudget < 1 {
		return ExplorationDiscovery{}, ErrNoVisibleFrontier
	}
	projection, err := transport.Projection(ctx)
	if err != nil {
		return ExplorationDiscovery{}, err
	}
	visited := explorer.Visited
	if visited == nil {
		visited = make(map[string]bool)
	}
	for action := 0; action < explorer.ActionBudget; action++ {
		if err := ctx.Err(); err != nil {
			return ExplorationDiscovery{}, err
		}
		decoded, mover, err := explorationInputs(projection, unitID)
		if err != nil {
			return ExplorationDiscovery{}, err
		}
		visited[mover.TileID] = true
		if enemy, ok := visibleEnemy(decoded, mover); ok {
			return ExplorationDiscovery{EnemyUnitID: enemy.ID, EnemyTileID: enemy.TileID, Moves: action, Projection: projection}, nil
		}
		_, path, ok := nextReachableFrontier(decoded, mover, visited, unitID)
		if !ok {
			return ExplorationDiscovery{}, ErrNoVisibleFrontier
		}
		projection, err = transport.Move(ctx, unitID, path[:1])
		if err != nil {
			return ExplorationDiscovery{}, err
		}
	}
	return ExplorationDiscovery{}, ErrExplorationBudget
}

func explorationInputs(raw []byte, unitID string) (visibility.Projection, visibility.UnitProjection, error) {
	var projection visibility.Projection
	if err := json.Unmarshal(raw, &projection); err != nil {
		return visibility.Projection{}, visibility.UnitProjection{}, err
	}
	for _, unit := range projection.Units {
		if unit.ID == unitID && unit.Domain != "" && unit.TileID != "" {
			return projection, unit, nil
		}
	}
	return visibility.Projection{}, visibility.UnitProjection{}, ErrNoVisibleFrontier
}

func visibleEnemy(projection visibility.Projection, mover visibility.UnitProjection) (visibility.UnitProjection, bool) {
	units := append([]visibility.UnitProjection(nil), projection.Units...)
	sort.Slice(units, func(i, j int) bool { return units[i].ID < units[j].ID })
	for _, unit := range units {
		if unit.OwnerID != "" && unit.OwnerID != mover.OwnerID && adjacentUnitTile(projection.Tiles, mover.TileID, unit.TileID) {
			return unit, true
		}
	}
	return visibility.UnitProjection{}, false
}

func adjacentUnitTile(tiles []visibility.TileProjection, leftID, rightID string) bool {
	var left visibility.TileProjection
	foundLeft, foundRight := false, false
	for _, tile := range tiles {
		if tile.ID == leftID {
			left, foundLeft = tile, true
		}
		if tile.ID == rightID {
			foundRight = true
		}
	}
	return foundLeft && foundRight && adjacent(left, rightID, tiles)
}

func nextReachableFrontier(projection visibility.Projection, mover visibility.UnitProjection, visited map[string]bool, unitID string) (visibility.TileProjection, []string, bool) {
	candidates := make([]visibility.TileProjection, 0, len(projection.Tiles))
	occupied := make(map[string]bool)
	for _, unit := range projection.Units {
		if unit.OwnerID != mover.OwnerID {
			occupied[unit.TileID] = true
		}
	}
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || !tile.Planning.Accessible || visited[tile.ID] || occupied[tile.ID] || !allows(tile.Planning.Domains, mover.Domain) {
			continue
		}
		candidates = append(candidates, tile)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Y != candidates[j].Y {
			return candidates[i].Y < candidates[j].Y
		}
		if candidates[i].X != candidates[j].X {
			return candidates[i].X < candidates[j].X
		}
		return candidates[i].ID < candidates[j].ID
	})
	for _, candidate := range candidates {
		path, err := visibility.Plan(projection, unitID, candidate.ID)
		if err == nil && len(path) > 0 {
			return candidate, path, true
		}
	}
	return visibility.TileProjection{}, nil, false
}

func nextReachableFrontierToward(projection visibility.Projection, mover visibility.UnitProjection, visited map[string]bool, unitID string, target visibility.TileProjection) (visibility.TileProjection, []string, bool) {
	candidates := make([]visibility.TileProjection, 0, len(projection.Tiles))
	occupied := make(map[string]bool)
	for _, unit := range projection.Units {
		if unit.OwnerID != mover.OwnerID {
			occupied[unit.TileID] = true
		}
	}
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || !tile.Planning.Accessible || visited[tile.ID] || occupied[tile.ID] || !allows(tile.Planning.Domains, mover.Domain) {
			continue
		}
		candidates = append(candidates, tile)
	}
	sort.Slice(candidates, func(i, j int) bool {
		leftDistance := abs(candidates[i].X-target.X) + abs(candidates[i].Y-target.Y)
		rightDistance := abs(candidates[j].X-target.X) + abs(candidates[j].Y-target.Y)
		if leftDistance != rightDistance {
			return leftDistance < rightDistance
		}
		if candidates[i].Y != candidates[j].Y {
			return candidates[i].Y < candidates[j].Y
		}
		if candidates[i].X != candidates[j].X {
			return candidates[i].X < candidates[j].X
		}
		return candidates[i].ID < candidates[j].ID
	})
	for _, candidate := range candidates {
		path, err := visibility.Plan(projection, unitID, candidate.ID)
		if err == nil && len(path) > 0 {
			return candidate, path, true
		}
	}
	return visibility.TileProjection{}, nil, false
}

func nextFrontier(projection visibility.Projection, mover visibility.UnitProjection, visited map[string]bool) (visibility.TileProjection, bool) {
	occupied := make(map[string]bool)
	for _, unit := range projection.Units {
		if unit.OwnerID != mover.OwnerID {
			occupied[unit.TileID] = true
		}
	}
	candidates := make([]visibility.TileProjection, 0, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || !tile.Planning.Accessible || visited[tile.ID] || occupied[tile.ID] || !allows(tile.Planning.Domains, mover.Domain) {
			continue
		}
		candidates = append(candidates, tile)
	}
	if len(candidates) == 0 {
		return visibility.TileProjection{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Y != candidates[j].Y {
			return candidates[i].Y < candidates[j].Y
		}
		if candidates[i].X != candidates[j].X {
			return candidates[i].X < candidates[j].X
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0], true
}

func adjacent(candidate visibility.TileProjection, moverTileID string, tiles []visibility.TileProjection) bool {
	for _, tile := range tiles {
		if tile.ID != moverTileID {
			continue
		}
		dx, dy := candidate.X-tile.X, candidate.Y-tile.Y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
	}
	return false
}

func allows(domains []string, domain string) bool {
	for _, candidate := range domains {
		if candidate == domain {
			return true
		}
	}
	return false
}
