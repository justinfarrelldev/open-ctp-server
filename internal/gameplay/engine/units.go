package engine

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const maxArmySize = 12

var ErrObstructed = errors.New("movement obstructed")

type moveUnitPayload struct {
	UnitID string   `json:"unit_id"`
	Path   []string `json:"path"`
}

type unitPayload struct {
	UnitID string `json:"unit_id"`
}

type embarkPayload struct {
	UnitID      string `json:"unit_id"`
	TransportID string `json:"transport_id"`
}

type disembarkPayload struct {
	UnitID string `json:"unit_id"`
	TileID string `json:"tile_id"`
}

type armyCreatePayload struct {
	ArmyID  string   `json:"army_id"`
	UnitIDs []string `json:"unit_ids"`
}

type armyUnitPayload struct {
	ArmyID string `json:"army_id"`
	UnitID string `json:"unit_id"`
}

func applyUnitCommand(state *State, command Command, runtime *rules.RuntimeRules) (string, error) {
	if err := activeActor(*state, command.ActorPlayerID); err != nil {
		return "", err
	}
	switch command.Type {
	case CommandMoveUnit:
		var payload moveUnitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return EventUnitMoved, moveUnit(state, command.ActorPlayerID, payload, runtime)
	case CommandFortify:
		var payload unitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		unit, err := ownedUnit(state, command.ActorPlayerID, payload.UnitID)
		if err != nil || unit.TransportID != "" {
			return "", invalidUnitAction(err)
		}
		unit.Fortified = true
		return "unit.fortified", nil
	case CommandDisband:
		var payload unitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		if _, err := ownedUnit(state, command.ActorPlayerID, payload.UnitID); err != nil {
			return "", err
		}
		removeUnitAndCargo(state, payload.UnitID)
		return EventUnitDisbanded, nil
	case CommandEmbark:
		var payload embarkPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "unit.embarked", embark(state, command.ActorPlayerID, payload)
	case CommandDisembark:
		var payload disembarkPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "unit.disembarked", disembark(state, command.ActorPlayerID, payload, runtime)
	case CommandArmyCreate:
		var payload armyCreatePayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "army.created", createArmy(state, command.ActorPlayerID, payload)
	case CommandArmyJoin:
		var payload armyUnitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "army.joined", joinArmy(state, command.ActorPlayerID, payload)
	case CommandArmyLeave:
		var payload armyUnitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "army.left", leaveArmy(state, command.ActorPlayerID, payload)
	default:
		return "", fmt.Errorf("unsupported command type %q", command.Type)
	}
}

func decodePayload(raw json.RawMessage, target any) error {
	if len(raw) == 0 || json.Unmarshal(raw, target) != nil {
		return errors.New("invalid unit command payload")
	}
	return nil
}

func activeActor(state State, actorID string) error {
	if (state.Phase.Mode == "simultaneous" || state.Phase.Mode == "hybrid") && state.Phase.ResolutionLocked {
		return nil
	}
	if state.Phase.ResolutionLocked || state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex] != actorID {
		return errors.New("action is not eligible in this phase")
	}
	return nil
}

func ownedUnit(state *State, actorID, unitID string) (*Unit, error) {
	for i := range state.Units {
		if state.Units[i].ID == unitID && state.Units[i].OwnerID == actorID {
			return &state.Units[i], nil
		}
	}
	return nil, ErrObstructed
}

func invalidUnitAction(err error) error {
	if err != nil {
		return err
	}
	return errors.New("unit action is invalid")
}

func moveUnit(state *State, actorID string, payload moveUnitPayload, runtime *rules.RuntimeRules) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TransportID != "" || len(payload.Path) == 0 {
		return invalidUnitAction(err)
	}
	from, ok := tileByID(*state, unit.TileID)
	if !ok {
		return errors.New("unit location is invalid")
	}
	domains := movementDomains(*state, *unit)
	if len(domains) == 0 {
		return ErrObstructed
	}
	movers := []*Unit{unit}
	if unit.ArmyID != "" {
		army, ok := armyByID(state, unit.ArmyID)
		if !ok {
			return ErrObstructed
		}
		movers = movers[:0]
		for _, id := range army.UnitIDs {
			member, ok := unitByID(*state, id)
			if !ok || member.TileID != unit.TileID {
				return ErrObstructed
			}
			movers = append(movers, member)
		}
	}
	remaining := unit.Movement
	for index, tileID := range payload.Path {
		to, ok := tileByID(*state, tileID)
		if !ok || !adjacent(from, to) || !canAccessTile(runtime, actorID, to) || !domainCanEnter(runtime, domains, EffectiveTerrain(*state, to.ID)) || occupiedByEnemy(*state, actorID, to.ID) {
			return ErrObstructed
		}
		cost, ok := terrainCost(*state, runtime, to.ID)
		if !ok {
			return ErrObstructed
		}
		if remaining < cost {
			return ErrObstructed
		}
		remaining -= cost
		if inEnemyZOC(*state, runtime, actorID, to.ID, domains) && index != len(payload.Path)-1 {
			return ErrObstructed
		}
		from = to
	}
	spent := unit.Movement - remaining
	for _, mover := range movers {
		if mover.Movement < spent {
			return ErrObstructed
		}
	}
	for _, mover := range movers {
		mover.TileID, mover.Movement, mover.Fortified = from.ID, mover.Movement-spent, false
	}
	captureCityAt(state, actorID, from.ID)
	state.Fog = append(state.Fog, FogObservation{PlayerID: actorID, TileID: from.ID, Turn: state.Phase.Turn})
	return nil
}

func captureCityAt(state *State, actorID, tileID string) string {
	for i := range state.Cities {
		if state.Cities[i].TileID == tileID && state.Cities[i].OwnerID != actorID {
			state.Cities[i].OwnerID = actorID
			return state.Cities[i].ID
		}
	}
	return ""
}

func embark(state *State, actorID string, payload embarkPayload) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TransportID != "" || unit.ArmyID != "" {
		return invalidUnitAction(err)
	}
	transport, err := ownedUnit(state, actorID, payload.TransportID)
	if err != nil || transport.ID == unit.ID || transport.TransportID != "" || unit.TileID != transport.TileID || transport.CargoCapacity <= cargoCount(*state, transport.ID) || !containsString(transport.CargoDomains, unit.Domain) {
		return ErrObstructed
	}
	unit.TransportID, unit.TileID, unit.Fortified = transport.ID, "", false
	return nil
}

func disembark(state *State, actorID string, payload disembarkPayload, runtime *rules.RuntimeRules) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TransportID == "" {
		return invalidUnitAction(err)
	}
	transport, ok := unitByID(*state, unit.TransportID)
	to, found := tileByID(*state, payload.TileID)
	from, originFound := tileByID(*state, transport.TileID)
	if !ok || !found || !originFound || !adjacent(from, to) || !canAccessTile(runtime, actorID, to) || !domainCanEnter(runtime, []string{unit.Domain}, EffectiveTerrain(*state, to.ID)) || occupiedByEnemy(*state, actorID, to.ID) || stackCount(*state, to.ID) >= maxArmySize {
		return ErrObstructed
	}
	cost, ok := terrainCost(*state, runtime, to.ID)
	if !ok {
		return ErrObstructed
	}
	if unit.Movement < cost {
		return ErrObstructed
	}
	unit.TransportID, unit.TileID, unit.Movement = "", to.ID, unit.Movement-cost
	return nil
}

func createArmy(state *State, actorID string, payload armyCreatePayload) error {
	if payload.ArmyID == "" || len(payload.UnitIDs) == 0 || len(payload.UnitIDs) > maxArmySize || armyExists(*state, payload.ArmyID) {
		return errors.New("army composition is invalid")
	}
	members, err := armyMembers(state, actorID, payload.UnitIDs)
	if err != nil {
		return err
	}
	army := Army{ID: payload.ArmyID, OwnerID: actorID, UnitIDs: append([]string(nil), payload.UnitIDs...)}
	for _, member := range members {
		member.ArmyID = army.ID
	}
	state.Armies = append(state.Armies, army)
	return nil
}

func joinArmy(state *State, actorID string, payload armyUnitPayload) error {
	army, ok := armyByID(state, payload.ArmyID)
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if !ok || err != nil || army.OwnerID != actorID || len(army.UnitIDs) >= maxArmySize || unit.ArmyID != "" || unit.TransportID != "" {
		return ErrObstructed
	}
	first, ok := unitByID(*state, army.UnitIDs[0])
	if !ok || first.TileID != unit.TileID {
		return ErrObstructed
	}
	army.UnitIDs = append(army.UnitIDs, unit.ID)
	unit.ArmyID = army.ID
	return nil
}

func leaveArmy(state *State, actorID string, payload armyUnitPayload) error {
	army, ok := armyByID(state, payload.ArmyID)
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if !ok || err != nil || army.OwnerID != actorID || unit.ArmyID != army.ID {
		return ErrObstructed
	}
	army.UnitIDs = removeString(army.UnitIDs, unit.ID)
	unit.ArmyID = ""
	if len(army.UnitIDs) == 0 {
		state.Armies = removeArmy(state.Armies, army.ID)
	}
	return nil
}

func armyMembers(state *State, actorID string, ids []string) ([]*Unit, error) {
	seen := map[string]bool{}
	members := make([]*Unit, 0, len(ids))
	for _, id := range ids {
		unit, err := ownedUnit(state, actorID, id)
		if err != nil || seen[id] || unit.ArmyID != "" || unit.TransportID != "" {
			return nil, ErrObstructed
		}
		seen[id] = true
		members = append(members, unit)
	}
	for _, member := range members[1:] {
		if member.TileID != members[0].TileID {
			return nil, ErrObstructed
		}
	}
	return members, nil
}

func removeUnitAndCargo(state *State, unitID string) {
	removed := map[string]bool{unitID: true}
	for changed := true; changed; {
		changed = false
		for _, unit := range state.Units {
			if removed[unit.TransportID] && !removed[unit.ID] {
				removed[unit.ID], changed = true, true
			}
		}
	}
	units := state.Units[:0]
	for _, unit := range state.Units {
		if !removed[unit.ID] {
			units = append(units, unit)
		}
	}
	state.Units = units
	armies := state.Armies[:0]
	for _, army := range state.Armies {
		army.UnitIDs = filterStrings(army.UnitIDs, removed)
		if len(army.UnitIDs) > 0 {
			armies = append(armies, army)
		}
	}
	state.Armies = armies
}

func validateUnitState(state State) error {
	if err := uniqueIDs("tile", tileIDs(state.Tiles)); err != nil {
		return err
	}
	for _, tile := range state.Tiles {
		if tile.ID == "" || !supportedTerrain(tile.Terrain) {
			return errors.New("tile terrain is unsupported")
		}
	}
	for _, unit := range state.Units {
		if unit.ID == "" || unit.OwnerID == "" || (unit.Domain != "" && !supportedDomain(unit.Domain)) || unit.Movement < 0 || unit.CargoCapacity < 0 || unit.VisibilityRange < 0 {
			return errors.New("unit capability is invalid")
		}
		if unit.TransportID != "" {
			if unit.TileID != "" || unit.ArmyID != "" {
				return errors.New("cargo must not have a tile or army")
			}
			if _, ok := unitByID(state, unit.TransportID); !ok {
				return errors.New("cargo transport does not exist")
			}
		} else if unit.TileID != "" && len(state.Tiles) > 0 {
			if _, ok := tileByID(state, unit.TileID); !ok {
				return errors.New("unit tile does not exist")
			}
		}
		for _, domain := range unit.CargoDomains {
			if !supportedDomain(domain) {
				return errors.New("cargo domain is unsupported")
			}
		}
	}
	// Pre-Task-8 v1 states did not carry canonical terrain or reciprocal army
	// references. They remain readable but cannot satisfy a unit command's tile checks.
	if len(state.Tiles) == 0 {
		return nil
	}
	seenMembers := map[string]string{}
	for _, army := range state.Armies {
		if army.OwnerID == "" || len(army.UnitIDs) == 0 || len(army.UnitIDs) > maxArmySize {
			return errors.New("army composition is invalid")
		}
		armyTileID := ""
		for _, id := range army.UnitIDs {
			unit, ok := unitByID(state, id)
			if !ok || unit.OwnerID != army.OwnerID || unit.ArmyID != army.ID || unit.TransportID != "" || seenMembers[id] != "" || (armyTileID != "" && unit.TileID != armyTileID) {
				return errors.New("army membership is invalid")
			}
			armyTileID = unit.TileID
			seenMembers[id] = army.ID
		}
	}
	for _, unit := range state.Units {
		if unit.ArmyID != "" && seenMembers[unit.ID] != unit.ArmyID {
			return errors.New("unit army reference is invalid")
		}
	}
	return nil
}

func tileIDs(items []Tile) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
func tileByID(state State, id string) (Tile, bool) {
	for _, tile := range state.Tiles {
		if tile.ID == id {
			return tile, true
		}
	}
	return Tile{}, false
}
func unitByID(state State, id string) (*Unit, bool) {
	for i := range state.Units {
		if state.Units[i].ID == id {
			return &state.Units[i], true
		}
	}
	return nil, false
}
func armyByID(state *State, id string) (*Army, bool) {
	for i := range state.Armies {
		if state.Armies[i].ID == id {
			return &state.Armies[i], true
		}
	}
	return nil, false
}
func armyExists(state State, id string) bool { _, ok := armyByID(&state, id); return ok }
func adjacent(a, b Tile) bool {
	dx, dy := a.X-b.X, a.Y-b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}
func canAccessTile(runtime *rules.RuntimeRules, actorID string, tile Tile) bool {
	return runtime != nil && runtime.Movement.Configured && (tile.OwnerID == "" || tile.OwnerID == actorID || (runtime.Movement.OpenAccess && tile.Open))
}
func supportedDomain(domain string) bool {
	return domain == DomainLand || domain == DomainSea || domain == DomainAmphibious
}
func supportedTerrain(terrain string) bool {
	switch terrain {
	case TerrainPlains, TerrainForest, TerrainOcean, "shelf", "mountain", "tundra", "desert", "swamp", "grassland":
		return true
	}
	return false
}
func domainCanEnter(runtime *rules.RuntimeRules, domains []string, terrainID string) bool {
	if runtime == nil || !runtime.Movement.Configured {
		return false
	}
	terrain, ok := runtime.Movement.Terrain(terrainID)
	if !ok {
		return false
	}
	for _, domain := range domains {
		if terrain.Allows(domain) {
			return true
		}
	}
	return false
}

func terrainCost(state State, runtime *rules.RuntimeRules, tileID string) (int, bool) {
	if runtime == nil || !runtime.Movement.Configured {
		return 0, false
	}
	terrain, ok := runtime.Movement.Terrain(EffectiveTerrain(state, tileID))
	if !ok {
		return 0, false
	}
	cost := terrain.Cost
	for _, overlay := range state.Overlays {
		if overlay.TileID == tileID && overlay.Kind == "improvement" && overlay.Remaining == 0 {
			cost += overlay.MovementCostDelta
		}
	}
	if cost < 1 {
		cost = 1
	}
	return cost, true
}
func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
func cargoCount(state State, transportID string) int {
	total := 0
	for _, unit := range state.Units {
		if unit.TransportID == transportID {
			total++
		}
	}
	return total
}
func stackCount(state State, tileID string) int {
	total := 0
	for _, unit := range state.Units {
		if unit.TileID == tileID && unit.TransportID == "" {
			total++
		}
	}
	return total
}
func removeString(items []string, value string) []string {
	result := items[:0]
	for _, item := range items {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}
func removeArmy(items []Army, id string) []Army {
	result := items[:0]
	for _, item := range items {
		if item.ID != id {
			result = append(result, item)
		}
	}
	return result
}
func filterStrings(items []string, removed map[string]bool) []string {
	result := items[:0]
	for _, item := range items {
		if !removed[item] {
			result = append(result, item)
		}
	}
	return result
}

func movementDomains(state State, unit Unit) []string {
	domains := []string{unit.Domain}
	if unit.ArmyID != "" {
		if army, ok := armyByID(&state, unit.ArmyID); ok {
			for _, id := range army.UnitIDs {
				if other, ok := unitByID(state, id); ok {
					domains = intersectDomains(domains, []string{other.Domain})
				}
			}
		}
	}
	return domains
}
func intersectDomains(left, right []string) []string {
	result := make([]string, 0, len(left))
	for _, candidate := range left {
		if containsString(right, candidate) {
			result = append(result, candidate)
		}
	}
	return result
}
func inEnemyZOC(state State, runtime *rules.RuntimeRules, actorID, tileID string, domains []string) bool {
	tile, ok := tileByID(state, tileID)
	if !ok || runtime == nil || !runtime.Movement.Configured || !runtime.Movement.ZOCStopsPath {
		return false
	}
	for _, unit := range state.Units {
		if !unit.ZOC || unit.OwnerID == actorID || unit.TransportID != "" {
			continue
		}
		source, ok := tileByID(state, unit.TileID)
		if ok && adjacent(source, tile) && domainCanEnter(runtime, domains, EffectiveTerrain(state, tile.ID)) {
			return true
		}
	}
	return false
}

func occupiedByEnemy(state State, actorID, tileID string) bool {
	for _, unit := range state.Units {
		if unit.TileID == tileID && unit.OwnerID != actorID {
			return true
		}
	}
	return false
}
