package engine

import (
	"errors"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// Territory uses the server's bounded cardinal tile graph, not an invented
// nearest-city/Voronoi partition. Original CTP2 uses the equivalent flood-fill
// barriers on its wrapped/skew map. National claims and influence are distinct.
type territoryGrid map[[2]int]int

func newTerritoryGrid(state *State) territoryGrid {
	grid := make(territoryGrid, len(state.Tiles))
	for n, tile := range state.Tiles {
		grid[[2]int{tile.X, tile.Y}] = n
	}
	return grid
}

func inCityRadius(center, tile Tile, intRadius, squaredRadius int) bool {
	dx, dy := tile.X-center.X, tile.Y-center.Y
	return abs(dx) <= intRadius && abs(dy) <= intRadius && dx*dx+dy*dy <= squaredRadius
}

func floodCityRadius(state *State, grid territoryGrid, center Tile, intRadius, squaredRadius int, accept func(*Tile) bool) {
	start, ok := grid[[2]int{center.X, center.Y}]
	if !ok {
		return
	}
	queue := []int{start}
	seen := map[int]bool{start: true}
	for head := 0; head < len(queue); head++ {
		tile := &state.Tiles[queue[head]]
		if !inCityRadius(center, *tile, intRadius, squaredRadius) || !accept(tile) {
			continue
		}
		for _, delta := range [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
			if next, ok := grid[[2]int{tile.X + delta[0], tile.Y + delta[1]}]; ok && !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
}

func initializeCityTerritory(state *State, city City, runtime *rules.RuntimeRules) {
	// Even a legacy pin can claim its newly founded center without guessing
	// unsupported radii or retroactively repairing other existing cities.
	for n := range state.Tiles {
		if state.Tiles[n].ID == city.TileID {
			state.Tiles[n].OwnerID = city.OwnerID
			state.Tiles[n].InfluenceCityID = city.ID
		}
	}
	if runtime == nil || runtime.CityTerritory == nil {
		return
	}
	refreshCityInfluence(state, runtime)
	claimCityNationalBorders(state, city, runtime)
}

func claimCityNationalBorders(state *State, city City, runtime *rules.RuntimeRules) {
	if runtime == nil || runtime.CityTerritory == nil {
		return
	}
	center, ok := tileByID(*state, city.TileID)
	if !ok {
		return
	}
	policy := runtime.CityTerritory
	floodCityRadius(state, newTerritoryGrid(state), center, policy.BorderIntRadius, policy.BorderSquaredRadius, func(tile *Tile) bool {
		if tile.OwnerID != "" && tile.OwnerID != city.OwnerID {
			return false
		}
		tile.OwnerID = city.OwnerID
		return true
	})
}

func refreshCityInfluence(state *State, runtime *rules.RuntimeRules) {
	cities := make(map[string]City, len(state.Cities))
	for _, city := range state.Cities {
		cities[city.ID] = city
	}
	// Loss of a city always clears its assignment, including legacy center-only
	// claims. National ownership intentionally survives, as in PrepareToRemove.
	for n := range state.Tiles {
		tile := &state.Tiles[n]
		if tile.InfluenceCityID != "" {
			if _, exists := cities[tile.InfluenceCityID]; !exists {
				tile.InfluenceCityID = ""
			}
		}
	}
	if runtime == nil || runtime.CityTerritory == nil {
		return
	}
	policy := runtime.CityTerritory
	grid := newTerritoryGrid(state)
	cityAt := make(map[string]string, len(cities))
	for _, city := range state.Cities {
		cityAt[city.TileID] = city.ID
	}
	centers := make(map[string]Tile, len(cities))
	for _, tile := range state.Tiles {
		if id, ok := cityAt[tile.ID]; ok {
			centers[id] = tile
		}
	}
	for n := range state.Tiles {
		tile := &state.Tiles[n]
		city, ok := cities[tile.InfluenceCityID]
		if !ok {
			continue
		}
		center, ok := centers[city.ID]
		size := policy.SizeForPopulation(city.Population)
		if !ok || !inCityRadius(center, *tile, size.IntRadius, size.SquaredRadius) {
			tile.InfluenceCityID = ""
		}
	}
	ordered := append([]City(nil), state.Cities...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	// Reserve centers before any expansion; never let a neighbor occupy a center.
	for _, city := range ordered {
		if center, ok := centers[city.ID]; ok {
			n := grid[[2]int{center.X, center.Y}]
			state.Tiles[n].InfluenceCityID = city.ID
			state.Tiles[n].OwnerID = city.OwnerID
		}
	}
	for _, city := range ordered {
		center, ok := centers[city.ID]
		if !ok {
			continue
		}
		size := policy.SizeForPopulation(city.Population)
		floodCityRadius(state, grid, center, size.IntRadius, size.SquaredRadius, func(tile *Tile) bool {
			if tile.InfluenceCityID != "" && tile.InfluenceCityID != city.ID {
				return false
			}
			tile.InfluenceCityID = city.ID
			tile.OwnerID = city.OwnerID
			return true
		})
	}
}

func validateCityTerritory(state State) error {
	cities := make(map[string]City, len(state.Cities))
	for _, city := range state.Cities {
		cities[city.ID] = city
	}
	for _, tile := range state.Tiles {
		if tile.InfluenceCityID != "" {
			city, ok := cities[tile.InfluenceCityID]
			if !ok || city.OwnerID != tile.OwnerID {
				return errors.New("tile city influence is dangling or has inconsistent national ownership")
			}
		}
	}
	return nil
}

func transferCityTerritory(state *State, cityID, newOwner string, runtime *rules.RuntimeRules) {
	for n := range state.Tiles {
		if state.Tiles[n].InfluenceCityID == cityID {
			state.Tiles[n].OwnerID = newOwner
		}
	}
	if city, ok := cityByID(*state, cityID); ok {
		initializeCityTerritory(state, *city, runtime)
	}
}
