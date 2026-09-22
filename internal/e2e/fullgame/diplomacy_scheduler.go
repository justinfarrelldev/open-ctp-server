package fullgame

import (
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type diplomaticActionKind string

const (
	diplomaticMove    diplomaticActionKind = "move"
	diplomaticEmbassy diplomaticActionKind = "establish-embassy"
)

type diplomaticAction struct {
	kind   diplomaticActionKind
	unitID string
	cityID string
	fromID string
	path   []string
}

type humanDiplomacyScheduler struct {
	maxActions     int
	actions        int
	visitedByUnit  map[string]map[string]bool
	previousByUnit map[string]string
}

func newHumanDiplomacyScheduler(maxActions int) *humanDiplomacyScheduler {
	return &humanDiplomacyScheduler{maxActions: maxActions, visitedByUnit: make(map[string]map[string]bool), previousByUnit: make(map[string]string)}
}

func (s *humanDiplomacyScheduler) remaining() bool {
	return s != nil && s.actions < s.maxActions
}

func (s *humanDiplomacyScheduler) recordSuccess(action diplomaticAction) {
	if !s.remaining() {
		return
	}
	s.actions++
	if action.kind != diplomaticMove || action.unitID == "" || len(action.path) == 0 {
		return
	}
	visited := s.visitedByUnit[action.unitID]
	if visited == nil {
		visited = make(map[string]bool)
		s.visitedByUnit[action.unitID] = visited
	}
	for _, tileID := range action.path {
		visited[tileID] = true
	}
	s.previousByUnit[action.unitID] = action.fromID
}

func (s *humanDiplomacyScheduler) nextAction(projection visibility.Projection, humanPlayer string) (diplomaticAction, bool) {
	if !s.remaining() || humanPlayer == "" {
		return diplomaticAction{}, false
	}
	units := eligibleDiplomacyUnits(projection.Units, humanPlayer)
	knownEmbassies := make(map[string]bool, len(projection.EmbassyPlayerIDs))
	for _, playerID := range projection.EmbassyPlayerIDs {
		knownEmbassies[playerID] = true
	}
	cities := append([]visibility.CityProjection(nil), projection.Cities...)
	sort.Slice(cities, func(i, j int) bool { return cities[i].ID < cities[j].ID })
	for _, unit := range units {
		if !hasPublicEffect(unit.Effects, engine.EffectSpecialEmbassy) {
			continue
		}
		for _, city := range cities {
			if city.Stale || city.OwnerID == humanPlayer || knownEmbassies[city.OwnerID] {
				continue
			}
			if adjacentUnitTile(projection.Tiles, unit.TileID, city.TileID) {
				return diplomaticAction{kind: diplomaticEmbassy, unitID: unit.ID, cityID: city.ID}, true
			}
			if path, ok := embassyApproach(projection, unit, city); ok {
				return diplomaticAction{kind: diplomaticMove, unitID: unit.ID, fromID: unit.TileID, path: path[:1]}, true
			}
		}
	}
	for _, unit := range units {
		visited := s.visitedByUnit[unit.ID]
		if visited == nil {
			visited = make(map[string]bool)
			s.visitedByUnit[unit.ID] = visited
		}
		visited[unit.TileID] = true
		if path, ok := nextDiplomaticFrontier(projection, unit, humanPlayer, visited); ok {
			return diplomaticAction{kind: diplomaticMove, unitID: unit.ID, fromID: unit.TileID, path: path[:1]}, true
		}
		if path, ok := s.revisitRoute(projection, unit); ok {
			return diplomaticAction{kind: diplomaticMove, unitID: unit.ID, fromID: unit.TileID, path: path[:1]}, true
		}
	}
	return diplomaticAction{}, false
}

func eligibleDiplomacyUnits(units []visibility.UnitProjection, humanPlayer string) []visibility.UnitProjection {
	eligible := make([]visibility.UnitProjection, 0, len(units))
	for _, unit := range units {
		if unit.OwnerID == humanPlayer && unit.Movement > 0 && unit.ID != "" && unit.TileID != "" && unit.Domain != "" {
			eligible = append(eligible, unit)
		}
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].ID < eligible[j].ID })
	return eligible
}

func embassyApproach(projection visibility.Projection, unit visibility.UnitProjection, city visibility.CityProjection) ([]string, bool) {
	candidates := make([]visibility.TileProjection, 0)
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || !tile.Planning.Accessible || !adjacentUnitTile(projection.Tiles, tile.ID, city.TileID) {
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
		path, err := visibility.Plan(projection, unit.ID, candidate.ID)
		if err == nil && len(path) > 0 {
			return path, true
		}
	}
	return nil, false
}

func (s *humanDiplomacyScheduler) revisitRoute(projection visibility.Projection, unit visibility.UnitProjection) ([]string, bool) {
	blocked := make(map[string]bool)
	for {
		candidate, path, ok := nextReachableFrontier(projection, unit, blocked, unit.ID)
		if !ok {
			return nil, false
		}
		blocked[candidate.ID] = true
		if len(path) > 0 && path[0] != s.previousByUnit[unit.ID] {
			return path, true
		}
	}
}

func nextDiplomaticFrontier(projection visibility.Projection, unit visibility.UnitProjection, humanPlayer string, visited map[string]bool) ([]string, bool) {
	cities := make([]visibility.TileProjection, 0)
	tiles := make(map[string]visibility.TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		tiles[tile.ID] = tile
	}
	for _, city := range projection.Cities {
		if city.OwnerID == humanPlayer && !city.Stale {
			if tile, ok := tiles[city.TileID]; ok {
				cities = append(cities, tile)
			}
		}
	}
	type route struct {
		tile     visibility.TileProjection
		path     []string
		distance int
	}
	routes := make([]route, 0)
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || !tile.Planning.Accessible || visited[tile.ID] || !allows(tile.Planning.Domains, unit.Domain) {
			continue
		}
		path, err := visibility.Plan(projection, unit.ID, tile.ID)
		if err != nil || len(path) == 0 {
			continue
		}
		distance := 0
		if len(cities) > 0 {
			distance = abs(tile.X-cities[0].X) + abs(tile.Y-cities[0].Y)
			for _, city := range cities[1:] {
				candidate := abs(tile.X-city.X) + abs(tile.Y-city.Y)
				if candidate < distance {
					distance = candidate
				}
			}
		}
		routes = append(routes, route{tile: tile, path: path, distance: distance})
	}
	if len(routes) == 0 {
		return nil, false
	}
	sort.Slice(routes, func(i, j int) bool {
		if len(cities) > 0 && routes[i].tile.X != routes[j].tile.X {
			return routes[i].tile.X > routes[j].tile.X
		}
		if routes[i].distance != routes[j].distance {
			return routes[i].distance < routes[j].distance
		}
		if routes[i].tile.Y != routes[j].tile.Y {
			return routes[i].tile.Y < routes[j].tile.Y
		}
		if routes[i].tile.X != routes[j].tile.X {
			return routes[i].tile.X > routes[j].tile.X
		}
		return routes[i].tile.ID < routes[j].tile.ID
	})
	return routes[0].path, true
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
