package engine

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	CommandSetMovementOrder    = "set-movement-order"
	CommandCancelMovementOrder = "cancel-movement-order"
)

// MovementOrder binds intent to an exact stack and expected location. Coordinates
// are public inputs, never hidden tile identifiers or an omniscient path preview.
type MovementOrder struct {
	UnitID    string   `json:"unit_id"`
	PlayerID  string   `json:"player_id"`
	ArmyID    string   `json:"army_id,omitempty"`
	UnitIDs   []string `json:"unit_ids"`
	TileID    string   `json:"tile_id"`
	TargetX   int      `json:"target_x"`
	TargetY   int      `json:"target_y"`
	CommandID string   `json:"command_id"`
}

type SetMovementOrderPayload struct {
	UnitID  string `json:"unit_id"`
	TargetX *int   `json:"target_x"`
	TargetY *int   `json:"target_y"`
}

func applyMovementOrder(state *State, command Command, runtime *rules.RuntimeRules) (string, error) {
	if err := activeActor(*state, command.ActorPlayerID); err != nil {
		return "", err
	}
	var payload SetMovementOrderPayload
	if err := decodePayload(command.Payload, &payload); err != nil {
		return "", err
	}
	unit, err := ownedUnit(state, command.ActorPlayerID, payload.UnitID)
	if err != nil || unit.TransportID != "" || unit.InSpace {
		return "", ErrObstructed
	}
	if command.Type == CommandCancelMovementOrder {
		cancelMovementForUnit(state, unit.ID)
		return "movement_order.cancelled", nil
	}
	if runtime == nil || !runtime.Movement.Configured || payload.TargetX == nil || payload.TargetY == nil {
		return "", ErrObstructed
	}
	found := false
	for _, tile := range state.Tiles {
		if tile.X == *payload.TargetX && tile.Y == *payload.TargetY {
			found = true
			break
		}
	}
	if !found || len(movementDomains(*state, *unit)) == 0 {
		return "", ErrObstructed
	}
	order := MovementOrder{UnitID: unit.ID, PlayerID: unit.OwnerID, ArmyID: unit.ArmyID, TileID: unit.TileID, TargetX: *payload.TargetX, TargetY: *payload.TargetY, CommandID: command.ID, UnitIDs: []string{unit.ID}}
	if unit.ArmyID != "" {
		army, _ := armyByID(state, unit.ArmyID)
		order.UnitIDs = append([]string(nil), army.UnitIDs...)
	}
	sort.Strings(order.UnitIDs)
	for _, id := range order.UnitIDs {
		member, _ := unitByID(*state, id)
		member.Asleep = false
		member.Fortified = false
	}
	cancelMovementForUnit(state, unit.ID)
	if advanceMovementOrder(state, &order, runtime) {
		state.MovementOrders = append(state.MovementOrders, order)
	}
	return "movement_order.set", nil
}

func cancelMovementForUnit(state *State, id string) {
	kept := state.MovementOrders[:0]
	for _, order := range state.MovementOrders {
		if !containsString(order.UnitIDs, id) {
			kept = append(kept, order)
		}
	}
	state.MovementOrders = kept
}

// Manual commands replace intent even when their successful result stays on the
// same tile (sleep/fortify/upgrade/attack). Failed trials are discarded by Apply.
func cancelManualMovementOrders(state *State, command Command) {
	if command.Type == CommandSetMovementOrder || command.Type == CommandCancelMovementOrder {
		return
	}
	var target struct {
		UnitID         string `json:"unit_id"`
		ArmyID         string `json:"army_id"`
		AttackerArmyID string `json:"attacker_army_id"`
	}
	_ = json.Unmarshal(command.Payload, &target)
	if target.UnitID != "" {
		cancelMovementForUnit(state, target.UnitID)
	}
	for _, id := range []string{target.ArmyID, target.AttackerArmyID} {
		if army, ok := armyByID(state, id); ok {
			for _, unitID := range army.UnitIDs {
				cancelMovementForUnit(state, unitID)
			}
		}
	}
}

func movementOrderValid(state State, order MovementOrder) bool {
	player, ok := playerByIDValue(state, order.PlayerID)
	if !ok || player.Lifecycle != PlayerActive {
		return false
	}
	unit, ok := unitByID(state, order.UnitID)
	if !ok || unit.ArmyID != order.ArmyID {
		return false
	}
	ids := []string{unit.ID}
	if order.ArmyID != "" {
		army, ok := armyByID(&state, order.ArmyID)
		if !ok {
			return false
		}
		ids = army.UnitIDs
	}
	if len(ids) != len(order.UnitIDs) {
		return false
	}
	for _, id := range ids {
		member, ok := unitByID(state, id)
		if !ok || !containsString(order.UnitIDs, id) || member.OwnerID != order.PlayerID || member.TileID != order.TileID || member.TransportID != "" || member.InSpace || member.Asleep || member.Fortified {
			return false
		}
	}
	return true
}

func pruneMovementOrders(state *State) {
	kept := state.MovementOrders[:0]
	for _, order := range state.MovementOrders {
		if movementOrderValid(*state, order) {
			kept = append(kept, order)
		} else {
			movementOrderNotice(state, order, "stopped")
		}
	}
	state.MovementOrders = kept
}

// Automatic progress runs exactly once at the new-round boundary, after refresh,
// including for disconnected players. It never calls Apply or manufactures a
// second transaction; the enclosing accepted command owns the whole transition.
func advanceMovementOrders(state *State, runtime *rules.RuntimeRules) {
	if state.Terminal.Ended {
		return
	}
	orders := state.MovementOrders
	state.MovementOrders = nil
	sort.Slice(orders, func(i, j int) bool { return orders[i].UnitID < orders[j].UnitID })
	for _, order := range orders {
		if state.Terminal.Ended {
			break
		}
		if advanceMovementOrder(state, &order, runtime) {
			state.MovementOrders = append(state.MovementOrders, order)
		}
		if winner, reason := evaluateVictory(*state, runtime); winner != "" {
			endGame(state, winner, reason)
		}
	}
}

func advanceMovementOrder(state *State, order *MovementOrder, runtime *rules.RuntimeRules) bool {
	if !movementOrderValid(*state, *order) {
		movementOrderNotice(state, *order, "stopped")
		return false
	}
	// At most one step per map tile, independent of client distance/movement input.
	for step := 0; step <= len(state.Tiles); step++ {
		from, ok := tileByID(*state, order.TileID)
		if !ok {
			break
		}
		if from.X == order.TargetX && from.Y == order.TargetY {
			movementOrderNotice(state, *order, "completed")
			return false
		}
		x, y := from.X+direction(from.X, order.TargetX), from.Y+direction(from.Y, order.TargetY)
		var to Tile
		found := false
		for _, tile := range state.Tiles {
			if tile.X == x && tile.Y == y {
				to = tile
				found = true
				break
			}
		}
		unit, _ := unitByID(*state, order.UnitID)
		if !found || !canAccessTile(runtime, order.PlayerID, to) || !domainCanEnter(runtime, movementDomains(*state, *unit), EffectiveTerrain(*state, to.ID)) || occupiedByEnemy(*state, order.PlayerID, to.ID) {
			break
		}
		cost, ok := terrainCost(*state, runtime, to.ID)
		if !ok {
			break
		}
		canAfford, impossible := true, false
		for _, id := range order.UnitIDs {
			member, _ := unitByID(*state, id)
			if member.Movement < cost {
				canAfford = false
				if refreshedUnitMovement(*state, *member, runtime.Starter.Movement) < cost {
					impossible = true
				}
			}
			if member.MaxFuel > 0 && member.Fuel < cost {
				canAfford = false
				if member.MaxFuel < cost || !unitCanRefuel(*state, *member) {
					impossible = true
				}
			}
		}
		if impossible {
			break
		}
		if !canAfford {
			return true
		}
		if err := moveUnit(state, order.PlayerID, moveUnitPayload{UnitID: order.UnitID, Path: []string{to.ID}}, runtime); err != nil {
			break
		}
		order.TileID = to.ID
		if to.X == order.TargetX && to.Y == order.TargetY {
			movementOrderNotice(state, *order, "completed")
			return false
		}
		movers := make([]*Unit, 0, len(order.UnitIDs))
		for _, id := range order.UnitIDs {
			unit, _ := unitByID(*state, id)
			movers = append(movers, unit)
		}
		if !moversIgnoreZOC(movers) && inEnemyZOC(*state, runtime, order.PlayerID, to.ID, movementDomains(*state, *movers[0])) {
			break
		}
	}
	movementOrderNotice(state, *order, "stopped")
	return false
}

func direction(from, to int) int {
	if from < to {
		return 1
	}
	if from > to {
		return -1
	}
	return 0
}

func movementOrderNotice(state *State, order MovementOrder, status string) {
	raw, _ := json.Marshal(struct {
		Revision                  int64
		UnitID, CommandID, Status string
	}{state.Revision + 1, order.UnitID, order.CommandID, status})
	appendNotification(state, Notification{ID: "movement:" + hashBytes(raw), PlayerID: order.PlayerID, Kind: "movement_order." + status, Turn: state.Phase.Turn, UnitID: order.UnitID, CommandID: order.CommandID})
}

func validateMovementOrders(state State) error {
	seen := map[string]bool{}
	for _, order := range state.MovementOrders {
		if order.UnitID == "" || order.PlayerID == "" || order.CommandID == "" || order.TileID == "" || len(order.UnitIDs) == 0 || len(order.UnitIDs) > maxArmySize || !containsString(order.UnitIDs, order.UnitID) {
			return errors.New("invalid movement order")
		}
		for _, id := range order.UnitIDs {
			if id == "" || seen[id] {
				return errors.New("overlapping movement orders")
			}
			seen[id] = true
		}
	}
	return nil
}
