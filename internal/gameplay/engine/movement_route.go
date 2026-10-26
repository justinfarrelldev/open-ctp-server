package engine

import (
	"container/heap"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// MovementSight is the same geometric sight used by the viewer projection.
// Routing reads terrain/occupants only inside this set, never fogged live cells.
func MovementSight(state State, playerID string, globalRadar bool) map[string]bool {
	visible := map[string]bool{}
	add := func(originID string, radius int) {
		origin, ok := tileByID(state, originID)
		if !ok {
			return
		}
		for _, tile := range state.Tiles {
			if tileDistance(origin, tile) <= radius && combatHasLOS(state, origin, tile) {
				visible[tile.ID] = true
			}
		}
	}
	for _, unit := range state.Units {
		if unit.OwnerID == playerID {
			add(unit.TileID, max(2, unit.VisibilityRange))
		}
	}
	for _, city := range state.Cities {
		if city.OwnerID == playerID {
			add(city.TileID, 2)
		}
	}
	for _, overlay := range state.Overlays {
		if overlay.OwnerID != playerID || overlay.Remaining > 0 || overlay.Kind != "improvement" {
			continue
		}
		radius := max(overlay.VisionRange, overlay.RadarRange)
		if overlay.ListeningPost {
			radius = max(radius, 2)
		}
		if radius > 0 {
			add(overlay.TileID, radius)
		}
	}
	if globalRadar && HasActiveWonderEffect(state, playerID, EffectGlobalRadar) {
		for _, tile := range state.Tiles {
			visible[tile.ID] = true
		}
	}
	return visible
}

type routeNode struct {
	id       string
	cost     int // g: entry costs already paid, scaled by 100
	estimate int // f = g + admissible remaining cost
}
type routeQueue []routeNode

func (q routeQueue) Len() int { return len(q) }
func (q routeQueue) Less(i, j int) bool {
	if q[i].estimate != q[j].estimate {
		return q[i].estimate < q[j].estimate
	}
	if q[i].cost == q[j].cost {
		return q[i].id < q[j].id
	}
	return q[i].cost < q[j].cost
}
func (q routeQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *routeQueue) Push(x any)   { *q = append(*q, x.(routeNode)) }
func (q *routeQueue) Pop() any     { old := *q; n := old[len(old)-1]; *q = old[:len(old)-1]; return n }

// nextMovementStep uses released UnitAstar-style A* over a visible route, or explores a reachable
// visible frontier toward an unseen destination. Stable ID ties make replay
// independent of map/slice iteration. Replanning after every step incorporates
// newly revealed terrain and changing obstacles without persisting hidden paths.
func nextMovementStep(state State, order MovementOrder, runtime *rules.RuntimeRules) (Tile, bool) {
	visible := MovementSight(state, order.PlayerID, true)
	// Keep execution's full state separate: routing sees only detected units.
	known := state
	known.Units = nil
	for _, unit := range state.Units {
		if MovementUnitVisible(state, order.PlayerID, visible, unit) {
			known.Units = append(known.Units, unit)
		}
	}
	tiles := map[string]Tile{}
	coordinates := map[[2]int]string{}
	targetID := ""
	for _, tile := range state.Tiles {
		coordinates[[2]int{tile.X, tile.Y}] = tile.ID
		if visible[tile.ID] {
			tiles[tile.ID] = tile
			if tile.X == order.TargetX && tile.Y == order.TargetY {
				targetID = tile.ID
			}
		}
	}
	leader, ok := unitByID(state, order.UnitID)
	if !ok {
		return Tile{}, false
	}
	movers := make([]*Unit, 0, len(order.UnitIDs))
	for _, id := range order.UnitIDs {
		member, ok := unitByID(state, id)
		if !ok {
			return Tile{}, false
		}
		movers = append(movers, member)
	}
	domains := movementDomains(state, *leader)
	armyMax := int(^uint(0) >> 1)
	allAir := true
	for _, member := range movers {
		allowance := member.MaxMovement
		if definition, ok := runtime.Unit(member.TypeID); ok {
			allowance = definition.Movement
		}
		if allowance < 1 {
			allowance = runtime.Starter.Movement
		}
		armyMax = min(armyMax, max(1, allowance))
		allAir = allAir && member.Domain == DomainAir
	}
	player, _ := playerByIDValue(state, order.PlayerID)
	pretty := player.Controller != ControllerAI
	// Unlike the original nearby-world heuristic, this lower bound reads only
	// visible costs and cannot overestimate after capped/pretty entry costs.
	minimum := int(^uint(0) >> 1)
	for id := range tiles {
		if cost, ok := terrainCost(state, runtime, id); ok {
			minimum = min(minimum, MovementRouteCost(Tile{}, Tile{}, cost, armyMax, allAir, pretty))
		}
	}
	if minimum == int(^uint(0)>>1) {
		return Tile{}, false
	}
	if allAir {
		minimum = min(minimum, 9000)
	} // UnitAstar::EstimateFutureCost: 90 points.
	target := Tile{X: order.TargetX, Y: order.TargetY}
	best := map[string]int{order.TileID: 0}
	previous := map[string]string{}
	queue := &routeQueue{{id: order.TileID, estimate: MovementRouteEstimate(tiles[order.TileID], target, minimum)}}
	heap.Init(queue)
	frontier := ""
	frontierDistance := int(^uint(0) >> 1)
	for queue.Len() > 0 {
		current := heap.Pop(queue).(routeNode)
		if best[current.id] != current.cost {
			continue
		}
		from := tiles[current.id]
		if current.id == targetID {
			return firstRouteTile(tiles, previous, order.TileID, current.id)
		}
		neighbors := []string{}
		boundary := false
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				id, exists := coordinates[[2]int{from.X + dx, from.Y + dy}]
				if !exists {
					continue
				}
				if !visible[id] {
					boundary = true
					continue
				}
				neighbors = append(neighbors, id)
			}
		}
		zoc := current.id != order.TileID && !moversIgnoreZOC(movers) && inEnemyZOC(known, runtime, order.PlayerID, current.id, domains)
		if targetID == "" && (boundary || zoc) && current.id != order.TileID {
			distance := max(absInt(from.X-order.TargetX), absInt(from.Y-order.TargetY))
			if distance < frontierDistance || (distance == frontierDistance && (frontier == "" || current.cost < best[frontier] || (current.cost == best[frontier] && current.id < frontier))) {
				frontier, frontierDistance = current.id, distance
			}
		}
		if zoc {
			continue
		}
		sort.Strings(neighbors)
		for _, id := range neighbors {
			to := tiles[id]
			if !canEnterMovementTile(known, runtime, order.PlayerID, movers, domains, to) {
				continue
			}
			cost, ok := terrainCost(state, runtime, id)
			if !ok {
				continue
			}
			possible := true
			for _, member := range movers {
				cost := UnitStepCost(runtime.Movement.RetailMovementVersion, member.Domain, cost)
				if member.MaxFuel > 0 && (member.MaxFuel < cost || (member.Fuel < cost && !unitCanRefuel(state, *member))) {
					possible = false
				}
				if runtime.Movement.RetailMovementVersion != 1 && refreshedUnitMovement(state, *member, runtime.Starter.Movement) < cost {
					possible = false
				}
			}
			cost = MovementRouteCost(from, to, cost, armyMax, allAir, pretty)
			if !possible || current.cost > int(^uint(0)>>1)-cost {
				continue
			}
			total := current.cost + cost
			if old, seen := best[id]; seen && old <= total {
				continue
			}
			best[id], previous[id] = total, current.id
			heap.Push(queue, routeNode{id: id, cost: total, estimate: total + MovementRouteEstimate(to, target, minimum)})
		}
	}
	if frontier != "" {
		return firstRouteTile(tiles, previous, order.TileID, frontier)
	}
	return Tile{}, false
}

func firstRouteTile(tiles map[string]Tile, previous map[string]string, origin, target string) (Tile, bool) {
	if target == origin {
		return Tile{}, false
	}
	for previous[target] != "" && previous[target] != origin {
		target = previous[target]
	}
	tile, ok := tiles[target]
	return tile, ok && previous[target] == origin
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
