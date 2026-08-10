// Package visibility derives deterministic player-visible state and persistent fog memory.
package visibility

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const defaultSightRadius = 2

type TileProjection struct {
	ID       string        `json:"id"`
	X        int           `json:"x"`
	Y        int           `json:"y"`
	Terrain  string        `json:"terrain"`
	Stale    bool          `json:"stale"`
	Observed int           `json:"observed_turn"`
	Planning *TilePlanning `json:"planning,omitempty"`
}

type TilePlanning struct {
	Cost       int      `json:"cost"`
	Domains    []string `json:"domains"`
	Accessible bool     `json:"accessible"`
	EnemyZOC   bool     `json:"enemy_zoc"`
}

type UnitProjection struct {
	ID        string          `json:"id"`
	OwnerID   string          `json:"owner_id"`
	TileID    string          `json:"tile_id"`
	ArmyID    string          `json:"army_id,omitempty"`
	Fortified bool            `json:"fortified"`
	Domain    string          `json:"domain,omitempty"`
	Movement  int             `json:"movement,omitempty"`
	Effects   []engine.Effect `json:"effects,omitempty"`
}

type CityProjection struct {
	ID       string          `json:"id"`
	OwnerID  string          `json:"owner_id"`
	TileID   string          `json:"tile_id"`
	Effects  []engine.Effect `json:"effects,omitempty"`
	Stale    bool            `json:"stale"`
	Observed int             `json:"observed_turn"`
}

type Projection struct {
	Revision         int64            `json:"revision"`
	Turn             int              `json:"turn"`
	Tiles            []TileProjection `json:"tiles"`
	Units            []UnitProjection `json:"units"`
	Cities           []CityProjection `json:"cities"`
	EmbassyPlayerIDs []string         `json:"embassy_player_ids,omitempty"`
}

// Update records only what each player can currently observe. The returned state
// remains the single canonical authority and is safe to persist/reload unchanged.
func Update(state engine.State) (engine.State, error) {
	if err := engine.ValidateState(state); err != nil {
		return engine.State{}, err
	}
	visible := visibleTiles(state)
	memory := make(map[string]engine.FogObservation, len(state.Fog))
	for _, observation := range state.Fog {
		memory[observation.PlayerID+"\x00"+observation.TileID] = observation
	}
	cities := citiesByTile(state.Cities)
	for playerID, tiles := range visible {
		for tileID := range tiles {
			tile, ok := tileByID(state.Tiles, tileID)
			if !ok {
				continue
			}
			observation := engine.FogObservation{PlayerID: playerID, TileID: tile.ID, Turn: state.Phase.Turn, Terrain: tile.Terrain}
			if city, ok := cities[tile.ID]; ok {
				observation.CityID = city.ID
				observation.CityOwnerID = city.OwnerID
				observation.CityEffects = append([]engine.Effect(nil), city.Effects...)
			}
			memory[playerID+"\x00"+tileID] = observation
		}
	}
	state.Fog = state.Fog[:0]
	for _, observation := range memory {
		state.Fog = append(state.Fog, observation)
	}
	sort.Slice(state.Fog, func(i, j int) bool {
		if state.Fog[i].PlayerID == state.Fog[j].PlayerID {
			return state.Fog[i].TileID < state.Fog[j].TileID
		}
		return state.Fog[i].PlayerID < state.Fog[j].PlayerID
	})
	return state, nil
}

func Project(state engine.State, playerID string) (Projection, error) {
	if playerID == "" {
		return Projection{}, errors.New("viewer player is required")
	}
	if err := engine.ValidateState(state); err != nil {
		return Projection{}, err
	}
	visible := visibleTiles(state)[playerID]
	memory := memoryByTile(state.Fog, playerID)
	planning, err := planningTiles(state, playerID, visible)
	if err != nil {
		return Projection{}, err
	}
	result := Projection{Revision: state.Revision, Turn: state.Phase.Turn}
	for _, player := range state.Players {
		if player.ID != playerID {
			continue
		}
		for _, nation := range state.Nations {
			if nation.ID != player.NationID {
				continue
			}
			for _, embassyNationID := range nation.Embassies {
				for _, embassyPlayer := range state.Players {
					if embassyPlayer.NationID == embassyNationID {
						result.EmbassyPlayerIDs = append(result.EmbassyPlayerIDs, embassyPlayer.ID)
					}
				}
			}
		}
	}
	for _, tile := range state.Tiles {
		if visible[tile.ID] {
			result.Tiles = append(result.Tiles, TileProjection{ID: tile.ID, X: tile.X, Y: tile.Y, Terrain: tile.Terrain, Observed: state.Phase.Turn, Planning: planning[tile.ID]})
			continue
		}
		if observation, ok := memory[tile.ID]; ok {
			result.Tiles = append(result.Tiles, TileProjection{ID: tile.ID, X: tile.X, Y: tile.Y, Terrain: observation.Terrain, Stale: true, Observed: observation.Turn})
			if observation.CityID != "" {
				result.Cities = append(result.Cities, CityProjection{ID: observation.CityID, OwnerID: observation.CityOwnerID, TileID: tile.ID, Effects: append([]engine.Effect(nil), observation.CityEffects...), Stale: true, Observed: observation.Turn})
			}
		}
	}
	for _, unit := range state.Units {
		if unit.OwnerID == playerID || visible[unit.TileID] {
			projected := UnitProjection{ID: unit.ID, OwnerID: unit.OwnerID, TileID: unit.TileID, ArmyID: unit.ArmyID, Fortified: unit.Fortified}
			if unit.OwnerID == playerID {
				projected.Domain, projected.Movement = unit.Domain, unit.Movement
				projected.Effects = append([]engine.Effect(nil), unit.Effects...)
			}
			result.Units = append(result.Units, projected)
		}
	}
	for _, city := range state.Cities {
		if visible[city.TileID] {
			result.Cities = append(result.Cities, CityProjection{ID: city.ID, OwnerID: city.OwnerID, TileID: city.TileID, Effects: append([]engine.Effect(nil), city.Effects...), Observed: state.Phase.Turn})
		}
	}
	sort.Slice(result.Tiles, func(i, j int) bool { return result.Tiles[i].ID < result.Tiles[j].ID })
	sort.Slice(result.Units, func(i, j int) bool { return result.Units[i].ID < result.Units[j].ID })
	sort.Slice(result.Cities, func(i, j int) bool { return result.Cities[i].ID < result.Cities[j].ID })
	sort.Strings(result.EmbassyPlayerIDs)
	return result, nil
}

func planningTiles(state engine.State, playerID string, visible map[string]bool) (map[string]*TilePlanning, error) {
	result := make(map[string]*TilePlanning)
	if len(state.Rules.Definitions) == 0 || string(state.Rules.Definitions) == "null" {
		return result, nil
	}
	var definition struct {
		CompatibilityVersion string `json:"compatibility_version"`
	}
	if err := json.Unmarshal(state.Rules.Definitions, &definition); err != nil {
		return nil, err
	}
	if definition.CompatibilityVersion != rules.CTP2CompatibilityVersion {
		return result, nil
	}
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		return nil, err
	}
	if !runtime.Movement.Configured {
		return result, nil
	}
	for _, tile := range state.Tiles {
		if !visible[tile.ID] {
			continue
		}
		terrain, ok := runtime.Movement.Terrain(engine.EffectiveTerrain(state, tile.ID))
		if !ok {
			return nil, errors.New("visible tile terrain is unavailable to the runtime movement model")
		}
		result[tile.ID] = &TilePlanning{
			Cost:       terrain.Cost,
			Domains:    append([]string(nil), terrain.Domains...),
			Accessible: tile.OwnerID == "" || tile.OwnerID == playerID || (runtime.Movement.OpenAccess && tile.Open),
		}
	}
	if !runtime.Movement.ZOCStopsPath {
		return result, nil
	}
	for _, unit := range state.Units {
		if unit.OwnerID == playerID || !unit.ZOC || unit.TransportID == "" && !visible[unit.TileID] {
			continue
		}
		for tileID, planning := range result {
			tile, _ := tileByID(state.Tiles, tileID)
			source, ok := tileByID(state.Tiles, unit.TileID)
			if ok && adjacentEngineTiles(source, tile) {
				planning.EnemyZOC = true
			}
		}
	}
	return result, nil
}

func Plan(projection Projection, unitID, targetTileID string) ([]string, error) {
	var mover UnitProjection
	foundMover := false
	occupied := make(map[string]bool)
	for _, unit := range projection.Units {
		if unit.ID == unitID {
			mover, foundMover = unit, true
		}
	}
	if !foundMover || mover.Domain == "" || mover.TileID == "" || mover.Movement < 0 || targetTileID == "" {
		return nil, errors.New("planner inputs are unavailable")
	}
	for _, unit := range projection.Units {
		if unit.OwnerID != mover.OwnerID && unit.ID != unitID {
			occupied[unit.TileID] = true
		}
	}
	tiles := make(map[string]TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || tile.Planning.Cost < 1 || !contains(tile.Planning.Domains, mover.Domain) {
			continue
		}
		tiles[tile.ID] = tile
	}
	if _, ok := tiles[mover.TileID]; !ok {
		return nil, errors.New("planner origin is not currently visible")
	}
	if _, ok := tiles[targetTileID]; !ok {
		return nil, errors.New("planner target is not currently visible")
	}
	type node struct {
		tileID string
		spent  int
		path   []string
	}
	queue := []node{{tileID: mover.TileID}}
	best := map[string]int{mover.TileID: 0}
	for len(queue) > 0 {
		sort.Slice(queue, func(i, j int) bool {
			if queue[i].spent == queue[j].spent {
				return queue[i].tileID < queue[j].tileID
			}
			return queue[i].spent < queue[j].spent
		})
		current := queue[0]
		queue = queue[1:]
		if current.tileID == targetTileID {
			return current.path, nil
		}
		from := tiles[current.tileID]
		for _, to := range tiles {
			if !adjacentProjection(from, to) || !to.Planning.Accessible || occupied[to.ID] {
				continue
			}
			spent := current.spent + to.Planning.Cost
			if spent > mover.Movement || (to.Planning.EnemyZOC && to.ID != targetTileID) {
				continue
			}
			if previous, seen := best[to.ID]; seen && previous <= spent {
				continue
			}
			best[to.ID] = spent
			path := append(append([]string(nil), current.path...), to.ID)
			queue = append(queue, node{tileID: to.ID, spent: spent, path: path})
		}
	}
	return nil, errors.New("no legal visible route")
}

// PlanKnown permits routing through tiles already retained in a player's projection.
// It derives only coarse navigation metadata from the remembered public terrain.
func PlanKnown(projection Projection, unitID, targetTileID string) ([]string, error) {
	known := projection
	known.Tiles = append([]TileProjection(nil), projection.Tiles...)
	for index := range known.Tiles {
		tile := &known.Tiles[index]
		if !tile.Stale || tile.Planning != nil {
			continue
		}
		planning := TilePlanning{Cost: 1, Domains: []string{"land", "amphibious"}, Accessible: true}
		if tile.Terrain == "forest" {
			planning.Cost = 2
		}
		if tile.Terrain == "ocean" {
			planning.Domains = []string{"sea", "amphibious"}
		}
		tile.Stale = false
		tile.Planning = &planning
	}
	return Plan(known, unitID, targetTileID)
}

func contains(items []string, value string) bool {
	index := sort.SearchStrings(items, value)
	return index < len(items) && items[index] == value
}

func adjacentProjection(left, right TileProjection) bool {
	dx, dy := abs(left.X-right.X), abs(left.Y-right.Y)
	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}

func adjacentEngineTiles(left, right engine.Tile) bool {
	dx, dy := abs(left.X-right.X), abs(left.Y-right.Y)
	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}

// AIInput is deliberately byte-identical to the controller's HTTP projection.
func AIInput(state engine.State, playerID string) ([]byte, error) {
	projection, err := Project(state, playerID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(projection)
}

// SafeError prevents existence-specific details from crossing a viewer boundary.
func SafeError(error) []byte { return []byte(`{"error":"gameplay command unavailable"}`) }

type Projector struct{ PlayerID string }

func NewProjector(playerID ...string) Projector {
	if len(playerID) == 0 {
		return Projector{}
	}
	return Projector{PlayerID: playerID[0]}
}
func (p Projector) ProjectSnapshot(viewer events.Viewer, canonical []byte) ([]byte, error) {
	state, err := engine.DecodeCanonicalStateJSON(canonical)
	if err != nil {
		return nil, err
	}
	projection, err := Project(state, p.viewerPlayerID(viewer))
	if err != nil {
		return nil, err
	}
	if !state.Terminal.Ended {
		return json.Marshal(projection)
	}
	return json.Marshal(struct {
		Projection
		Terminal struct {
			WinningPlayerID string `json:"winning_player_id"`
			Reason          string `json:"reason"`
		} `json:"terminal"`
	}{Projection: projection, Terminal: struct {
		WinningPlayerID string `json:"winning_player_id"`
		Reason          string `json:"reason"`
	}{WinningPlayerID: state.Terminal.WinningPlayerID, Reason: state.Terminal.Reason}})
}
func (p Projector) ProjectEvent(_ events.Viewer, event events.Event) ([]byte, error) {
	return json.Marshal(struct {
		Type     string `json:"type"`
		Sequence int64  `json:"sequence"`
	}{Type: event.Type, Sequence: event.Sequence})
}
func (p Projector) viewerPlayerID(viewer events.Viewer) string {
	if p.PlayerID != "" {
		return p.PlayerID
	}
	if viewer.AccountID > 0 {
		return "account-" + strconv.FormatInt(viewer.AccountID, 10)
	}
	return strconv.FormatInt(viewer.PlayerID, 10)
}

func visibleTiles(state engine.State) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	for _, player := range state.Players {
		result[player.ID] = make(map[string]bool)
	}
	for _, unit := range state.Units {
		addSight(result[unit.OwnerID], state, unit.TileID, defaultSightRadius)
	}
	for _, city := range state.Cities {
		addSight(result[city.OwnerID], state, city.TileID, defaultSightRadius)
	}
	return result
}
func addSight(visible map[string]bool, state engine.State, origin string, radius int) {
	from, ok := tileByID(state.Tiles, origin)
	if !ok {
		return
	}
	for _, tile := range state.Tiles {
		distance := abs(tile.X-from.X) + abs(tile.Y-from.Y)
		if distance > radius || !hasLOS(state, from, tile) {
			continue
		}
		visible[tile.ID] = true
	}
}
func hasLOS(state engine.State, from, to engine.Tile) bool {
	x, y := from.X, from.Y
	dx, dy := sign(to.X-from.X), sign(to.Y-from.Y)
	for x != to.X || y != to.Y {
		x += dx
		y += dy
		if x == to.X && y == to.Y {
			return true
		}
		tile, ok := tileAt(state.Tiles, x, y)
		if !ok || engine.EffectiveTerrain(state, tile.ID) == engine.TerrainForest {
			return false
		}
	}
	return true
}
func tileByID(tiles []engine.Tile, id string) (engine.Tile, bool) {
	for _, tile := range tiles {
		if tile.ID == id {
			return tile, true
		}
	}
	return engine.Tile{}, false
}
func tileAt(tiles []engine.Tile, x, y int) (engine.Tile, bool) {
	for _, tile := range tiles {
		if tile.X == x && tile.Y == y {
			return tile, true
		}
	}
	return engine.Tile{}, false
}
func citiesByTile(cities []engine.City) map[string]engine.City {
	result := make(map[string]engine.City, len(cities))
	for _, city := range cities {
		result[city.TileID] = city
	}
	return result
}
func memoryByTile(fog []engine.FogObservation, playerID string) map[string]engine.FogObservation {
	result := make(map[string]engine.FogObservation)
	for _, observation := range fog {
		if observation.PlayerID == playerID {
			result[observation.TileID] = observation
		}
	}
	return result
}
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}
