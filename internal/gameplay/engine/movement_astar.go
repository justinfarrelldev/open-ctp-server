package engine

// MovementRouteCost ports UnitAstar::ComputeValidMovCost / CheckMoveIntersection
// and Astar::DecayOrtho from robot/pathing/{unitastar,astar}.cpp. Routing cost is
// not movement expenditure: terrain is capped at the army's slowest *maximum*,
// air costs 100, and human "pretty" routes discount orthogonal entries by 5%.
// Hundredths keep the released 0.95 preference deterministic without floats.
func MovementRouteCost(from, to Tile, terrainCost, armyMax int, allAir, pretty bool) int {
	cost := min(terrainCost, armyMax)
	if allAir {
		cost = 100
	}
	scale := 100
	if pretty && (from.X == to.X || (from.X < to.X && from.Y-1 == to.Y) || (from.X > to.X && from.Y+1 == to.Y)) {
		scale = 95
	}
	return max(1, cost) * scale
}

// MovementRouteEstimate adapts normalized distance to the server's bounded,
// eight-neighbor coordinate grid. minimum is a viewer-known lower bound, not a
// query of hidden terrain (the released nearby-cost heuristic reads the world).
func MovementRouteEstimate(from, to Tile, minimum int) int {
	return max(absInt(from.X-to.X), absInt(from.Y-to.Y)) * minimum
}
