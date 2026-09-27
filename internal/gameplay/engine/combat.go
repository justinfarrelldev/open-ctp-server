package engine

import (
	"encoding/json"
	"errors"
	"sort"
)

const (
	EffectCombatRanged = "combat_ranged"

	CombatOutcomeAttackerWon  = "attacker_won"
	CombatOutcomeDefenderHeld = "defender_held"
	CombatOutcomeRetreated    = "retreated"
)

type combatPayload struct {
	AttackerArmyID string `json:"attacker_army_id"`
	DefenderArmyID string `json:"defender_army_id"`
}

type bombardPayload struct {
	UnitID       string `json:"unit_id"`
	TargetTileID string `json:"target_tile_id"`
}

// CombatResult is deliberately aggregate-only: it is sufficient to animate a
// visible battle without disclosing either army's composition or RNG draws.
type CombatResult struct {
	Outcome            string `json:"outcome"`
	AttackerCasualties int    `json:"attacker_casualties"`
	DefenderCasualties int    `json:"defender_casualties"`
	AttackerArmyID     string `json:"attacker_army_id"`
	DefenderArmyID     string `json:"defender_army_id"`
	CapturedCityID     string `json:"captured_city_id,omitempty"`
}

func applyCombat(state *State, command Command, sources Sources) (*CombatResult, json.RawMessage, error) {
	if err := activeActor(*state, command.ActorPlayerID); err != nil {
		return nil, nil, err
	}
	var payload combatPayload
	if err := decodePayload(command.Payload, &payload); err != nil || payload.AttackerArmyID == "" || payload.DefenderArmyID == "" || payload.AttackerArmyID == payload.DefenderArmyID {
		return nil, nil, errors.New("combat target is unavailable")
	}
	attackers, ok := armyByID(state, payload.AttackerArmyID)
	if !ok || attackers.OwnerID != command.ActorPlayerID {
		return nil, nil, errors.New("combat target is unavailable")
	}
	defenders, ok := armyByID(state, payload.DefenderArmyID)
	if !ok || defenders.OwnerID == command.ActorPlayerID {
		return nil, nil, errors.New("combat target is unavailable")
	}
	attackerUnits := combatArmyUnits(*state, *attackers)
	defenderUnits := combatArmyUnits(*state, *defenders)
	if len(attackerUnits) == 0 || len(defenderUnits) == 0 || attackerUnits[0].TileID == "" || defenderUnits[0].TileID == "" || !adjacentTiles(*state, attackerUnits[0].TileID, defenderUnits[0].TileID) || !combatVisible(*state, command.ActorPlayerID, defenderUnits[0].TileID) {
		return nil, nil, errors.New("combat target is unavailable")
	}
	targetTileID := defenderUnits[0].TileID
	for _, unit := range append(append([]*Unit(nil), attackerUnits...), defenderUnits...) {
		if unit.Attack < 0 || unit.Defense < 0 || unit.Ranged < 0 || unit.Strength <= 0 || unit.Armor <= 0 || unit.Firepower < 0 || unit.Health <= 0 || unit.MaxHealth < unit.Health {
			return nil, nil, errors.New("combat unit is invalid")
		}
	}
	if !canArmyFight(*state, attackerUnits, defenderUnits) || barbarianAttackBlocked(*state, attackers.OwnerID, defenders.OwnerID) {
		return nil, nil, errors.New("combat target is unavailable")
	}

	result := &CombatResult{AttackerArmyID: attackers.ID, DefenderArmyID: defenders.ID}
	// The fixed CTP2-derived sequence is attacker ranged, defender ranged,
	// attacker melee, defender melee, with IDs ordering each line. No rejected
	// command reaches this point, so it cannot advance combat RNG.
	for rounds := 0; rounds < 64 && len(attackerUnits) > 0 && len(defenderUnits) > 0; rounds++ {
		combatStrike(state, sources, attackerUnits, defenderUnits, true, result)
		defenderUnits = livingCombatUnits(defenderUnits)
		if len(defenderUnits) == 0 {
			break
		}
		combatStrike(state, sources, defenderUnits, attackerUnits, true, result)
		attackerUnits = livingCombatUnits(attackerUnits)
		if len(attackerUnits) == 0 {
			break
		}
		combatStrike(state, sources, attackerUnits, defenderUnits, false, result)
		defenderUnits = livingCombatUnits(defenderUnits)
		if len(defenderUnits) == 0 {
			break
		}
		combatStrike(state, sources, defenderUnits, attackerUnits, false, result)
		attackerUnits = livingCombatUnits(attackerUnits)
	}
	for _, unit := range combatArmyUnits(*state, *attackers) {
		if unit.Health <= 0 {
			recordUnitLossForPeaceMovement(state, *unit)
			removeUnitAndCargo(state, unit.ID)
			result.AttackerCasualties++
		}
	}
	for _, unit := range combatArmyUnits(*state, *defenders) {
		if unit.Health <= 0 {
			recordUnitLossForPeaceMovement(state, *unit)
			removeUnitAndCargo(state, unit.ID)
			result.DefenderCasualties++
		}
	}
	for _, unit := range combatArmyUnits(*state, *attackers) {
		if unit.Health > 0 && hasEffect(unit.Effects, EffectSingleUse) {
			removeUnitAndCargo(state, unit.ID)
			result.AttackerCasualties++
		}
	}
	if _, ok := armyByID(state, attackers.ID); !ok {
		result.Outcome = CombatOutcomeDefenderHeld
		promoteArmyVeterans(state, defenders.ID)
	} else if _, ok := armyByID(state, defenders.ID); !ok {
		result.Outcome = CombatOutcomeAttackerWon
		if armyCanCaptureTargetCity(*state, *attackers, targetTileID) {
			moveArmyToTile(state, attackers.ID, targetTileID)
			result.CapturedCityID = captureCityAt(state, command.ActorPlayerID, targetTileID)
		}
		promoteArmyVeterans(state, attackers.ID)
	} else {
		result.Outcome = CombatOutcomeRetreated
	}
	eventPayload, err := json.Marshal(result)
	if err != nil {
		return nil, nil, err
	}
	return result, eventPayload, nil
}

func recordUnitLossForPeaceMovement(state *State, unit Unit) {
	if state == nil || unit.OwnerID == "" {
		return
	}
	nation := nationForID(state, nationForPlayer(*state, unit.OwnerID))
	if nation == nil {
		return
	}
	if nation.PeaceMovement == nil {
		nation.PeaceMovement = &PeaceMovementState{}
	}
	if unitLostAtHome(*state, unit, *nation) {
		nation.PeaceMovement.UnitsLostHome++
	} else {
		nation.PeaceMovement.UnitsLostOverseas++
	}
}

func unitLostAtHome(state State, unit Unit, nation Nation) bool {
	unitTile, ok := tileByID(state, unit.TileID)
	if !ok {
		return true
	}
	radius := atHomeRadiusForNation(nation)
	nearest := -1
	for _, city := range state.Cities {
		if city.OwnerID != unit.OwnerID {
			continue
		}
		cityTile, ok := tileByID(state, city.TileID)
		if !ok {
			continue
		}
		distance := tileDistance(unitTile, cityTile)
		if nearest < 0 || distance < nearest {
			nearest = distance
		}
	}
	if nearest < 0 {
		return true
	}
	return nearest <= radius
}

func atHomeRadiusForNation(nation Nation) int {
	radius := 0
	for _, effect := range nation.Effects {
		if effect.Kind == EffectAtHomeRadius && effect.Value > radius {
			radius = effect.Value
		}
	}
	return radius
}

func promoteArmyVeterans(state *State, armyID string) {
	army, ok := armyByID(state, armyID)
	if !ok {
		return
	}
	unitIDs := make(map[string]bool, len(army.UnitIDs))
	for _, unitID := range army.UnitIDs {
		unitIDs[unitID] = true
	}
	for index := range state.Units {
		if unitIDs[state.Units[index].ID] && state.Units[index].Health > 0 {
			state.Units[index].Veteran = true
		}
	}
}

func barbarianAttackBlocked(state State, attackerID, defenderID string) bool {
	return isBarbarianPlayer(state, attackerID) && HasActiveWonderEffect(state, defenderID, EffectProtectFromBarbarians)
}

func isBarbarianPlayer(state State, playerID string) bool {
	if playerID == "barbarian" || playerID == "vandals" || playerID == "PLAYER_INDEX_VANDALS" {
		return true
	}
	player, ok := playerByIDValue(state, playerID)
	if !ok {
		return false
	}
	return player.NationID == "CIVILIZATION_BARBARIAN" || player.NationID == "nation-barbarian" || player.NationID == "barbarian" || player.NationID == "vandals"
}

func armyCanCaptureTargetCity(state State, attackers Army, targetTileID string) bool {
	hasCity := false
	for _, city := range state.Cities {
		if city.TileID == targetTileID {
			hasCity = true
			break
		}
	}
	if !hasCity {
		return true
	}
	for _, unit := range combatArmyUnits(state, attackers) {
		if !hasEffect(unit.Effects, EffectCantCaptureCity) {
			return true
		}
	}
	return false
}

func applyBombard(state *State, command Command, sources Sources) (string, json.RawMessage, error) {
	if err := activeActor(*state, command.ActorPlayerID); err != nil {
		return "", nil, err
	}
	var payload bombardPayload
	if err := decodePayload(command.Payload, &payload); err != nil || payload.UnitID == "" || payload.TargetTileID == "" {
		return "", nil, errors.New("bombard target is unavailable")
	}
	unit, err := ownedUnit(state, command.ActorPlayerID, payload.UnitID)
	if err != nil || unit.TileID == "" || unit.TransportID != "" || unit.Movement < 1 || unit.Ranged <= 0 || unit.BombardRange < 1 || unit.BombRounds < 1 {
		return "", nil, errors.New("bombard target is unavailable")
	}
	origin, ok := tileByID(*state, unit.TileID)
	targetTile, targetOK := tileByID(*state, payload.TargetTileID)
	if !ok || !targetOK || !combatVisible(*state, command.ActorPlayerID, payload.TargetTileID) || tileDistance(origin, targetTile) > unit.BombardRange {
		return "", nil, errors.New("bombard target is unavailable")
	}
	targets := bombardEligibleTargets(*unit, bombardTargets(*state, command.ActorPlayerID, payload.TargetTileID))
	if len(targets) == 0 {
		return "", nil, errors.New("bombard target is unavailable")
	}
	activeDefenseHits := 0
	interceptors := activeDefenseInterceptors(*state, command.ActorPlayerID, targets[0].OwnerID, *unit, targetTile)
	for _, interceptor := range interceptors {
		defense := unit.Defense + combatDefenseModifier(*state, *unit, *interceptor)
		if defense < 0 {
			defense = 0
		}
		roll := combatRoll(state, sources)
		if int(roll%1000) < interceptor.Ranged*1000/(interceptor.Ranged+defense) {
			unit.Health -= combatDamage(*interceptor, *unit)
			activeDefenseHits++
		}
		if unit.Health <= 0 {
			break
		}
	}
	if unit.Health <= 0 {
		unit.Movement--
		removeUnitAndCargo(state, unit.ID)
		payloadJSON, _ := json.Marshal(struct {
			UnitID            string `json:"unit_id"`
			TargetTileID      string `json:"target_tile_id"`
			Hits              int    `json:"hits"`
			Destroyed         int    `json:"destroyed"`
			ActiveDefenseHits int    `json:"active_defense_hits,omitempty"`
		}{UnitID: payload.UnitID, TargetTileID: payload.TargetTileID, ActiveDefenseHits: activeDefenseHits})
		return EventBombarded, payloadJSON, nil
	}
	hits := 0
	for round := 0; round < unit.BombRounds && len(targets) > 0; round++ {
		target := targets[0]
		defense := target.Defense + combatDefenseModifier(*state, *target, *unit)
		if defense < 0 {
			defense = 0
		}
		roll := combatRoll(state, sources)
		if int(roll%1000) < unit.Ranged*1000/(unit.Ranged+defense) {
			target.Health -= combatDamage(*unit, *target)
			hits++
		}
		targets = livingCombatUnits(targets)
	}
	counterBombardHits := counterBombard(state, sources, unit, livingCombatUnits(targets))
	unit.Movement--
	destroyed := 0
	for _, target := range bombardTargets(*state, command.ActorPlayerID, payload.TargetTileID) {
		if target.Health <= 0 {
			removeUnitAndCargo(state, target.ID)
			destroyed++
		}
	}
	if unit.Health <= 0 || hasEffect(unit.Effects, EffectSingleUse) {
		removeUnitAndCargo(state, unit.ID)
	}
	payloadJSON, _ := json.Marshal(struct {
		UnitID             string `json:"unit_id"`
		TargetTileID       string `json:"target_tile_id"`
		Hits               int    `json:"hits"`
		Destroyed          int    `json:"destroyed"`
		ActiveDefenseHits  int    `json:"active_defense_hits,omitempty"`
		CounterBombardHits int    `json:"counter_bombard_hits,omitempty"`
	}{UnitID: unit.ID, TargetTileID: payload.TargetTileID, Hits: hits, Destroyed: destroyed, ActiveDefenseHits: activeDefenseHits, CounterBombardHits: counterBombardHits})
	return EventBombarded, payloadJSON, nil
}

func counterBombard(state *State, sources Sources, attacker *Unit, defenders []*Unit) int {
	hits := 0
	for _, defender := range defenders {
		if attacker.Health <= 0 {
			break
		}
		if !hasEffect(defender.Effects, EffectCanCounterBombard) || defender.Ranged <= 0 || defender.BombRounds < 1 || len(bombardEligibleTargets(*defender, []*Unit{attacker})) == 0 {
			continue
		}
		for round := 0; round < defender.BombRounds && attacker.Health > 0; round++ {
			defense := attacker.Defense + combatDefenseModifier(*state, *attacker, *defender)
			if defense < 0 {
				defense = 0
			}
			roll := combatRoll(state, sources)
			if int(roll%1000) < defender.Ranged*1000/(defender.Ranged+defense) {
				attacker.Health -= combatDamage(*defender, *attacker)
				hits++
			}
		}
	}
	return hits
}

func tileDistance(left, right Tile) int {
	dx := left.X - right.X
	if dx < 0 {
		dx = -dx
	}
	dy := left.Y - right.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func activeDefenseInterceptors(state State, actorID, protectedOwnerID string, attacker Unit, targetTile Tile) []*Unit {
	interceptors := make([]*Unit, 0)
	for i := range state.Units {
		unit := &state.Units[i]
		if unit.OwnerID == actorID || unit.OwnerID != protectedOwnerID || unit.TransportID != "" || unit.ActiveDefenseRange < 1 || unit.Ranged < 1 || unit.Health <= 0 {
			continue
		}
		tile, ok := tileByID(state, unit.TileID)
		if !ok || tileDistance(tile, targetTile) > unit.ActiveDefenseRange || !activeDefenseCanDefend(*unit, attacker) || !activeDefenseHasRequiredEnabler(state, *unit) {
			continue
		}
		interceptors = append(interceptors, unit)
	}
	sort.Slice(interceptors, func(i, j int) bool { return interceptors[i].ID < interceptors[j].ID })
	return interceptors
}

func activeDefenseHasRequiredEnabler(state State, interceptor Unit) bool {
	if !hasEffect(interceptor.Effects, EffectActiveDefenseRequiresEnabler) {
		return true
	}
	for _, cargo := range state.Units {
		if cargo.TransportID == interceptor.ID && cargo.OwnerID == interceptor.OwnerID && hasEffect(cargo.Effects, EffectEnableCarrierDefenses) {
			return true
		}
	}
	return false
}

func activeDefenseCanDefend(interceptor, attacker Unit) bool {
	if !interceptorHasDefendMetadata(interceptor) {
		return true
	}
	switch attacker.Domain {
	case DomainLand:
		return hasEffect(interceptor.Effects, EffectDefendLand)
	case DomainSea:
		return hasEffect(interceptor.Effects, EffectDefendSea)
	case DomainAir:
		return hasEffect(interceptor.Effects, EffectDefendAir)
	case DomainSpace:
		return hasEffect(interceptor.Effects, EffectDefendSpace)
	case DomainAmphibious:
		return hasEffect(interceptor.Effects, EffectDefendLand) || hasEffect(interceptor.Effects, EffectDefendSea)
	default:
		return true
	}
}

func interceptorHasDefendMetadata(interceptor Unit) bool {
	return hasEffect(interceptor.Effects, EffectDefendLand) || hasEffect(interceptor.Effects, EffectDefendSea) || hasEffect(interceptor.Effects, EffectDefendAir) || hasEffect(interceptor.Effects, EffectDefendSpace) || hasEffect(interceptor.Effects, EffectDefendUnderwater)
}

func bombardTargets(state State, actorID, tileID string) []*Unit {
	targets := make([]*Unit, 0)
	for i := range state.Units {
		if state.Units[i].OwnerID != actorID && state.Units[i].TileID == tileID && state.Units[i].TransportID == "" {
			targets = append(targets, &state.Units[i])
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].ID < targets[j].ID })
	return targets
}

func canArmyFight(state State, attackers, defenders []*Unit) bool {
	defenderCity := false
	if len(defenders) > 0 {
		for _, city := range state.Cities {
			if city.TileID == defenders[0].TileID {
				defenderCity = true
				break
			}
		}
	}
	for _, attacker := range attackers {
		if attacker.Attack <= 0 && attacker.Ranged <= 0 {
			continue
		}
		for _, defender := range defenders {
			if attackerHasCanAttackMetadata(*attacker) {
				if attackerCanAttackTarget(*attacker, *defender) {
					return true
				}
				continue
			}
			if defenderCity || attacker.Domain == "" || defender.Domain == "" || attacker.Domain == defender.Domain || attacker.Domain == DomainAmphibious || defender.Domain == DomainAmphibious {
				return true
			}
		}
	}
	return false
}

func attackerHasCanAttackMetadata(attacker Unit) bool {
	return hasEffect(attacker.Effects, EffectCanAttackLand) || hasEffect(attacker.Effects, EffectCanAttackSea) || hasEffect(attacker.Effects, EffectCanAttackAir) || hasEffect(attacker.Effects, EffectCanAttackSpace) || hasEffect(attacker.Effects, EffectCanAttackUnderwater)
}

func attackerCanAttackTarget(attacker, target Unit) bool {
	if hasEffect(target.Effects, EffectSubmarine) {
		return hasEffect(attacker.Effects, EffectCanAttackUnderwater)
	}
	switch target.Domain {
	case DomainLand:
		return hasEffect(attacker.Effects, EffectCanAttackLand)
	case DomainSea:
		return hasEffect(attacker.Effects, EffectCanAttackSea)
	case DomainAir:
		return hasEffect(attacker.Effects, EffectCanAttackAir)
	case DomainSpace:
		return hasEffect(attacker.Effects, EffectCanAttackSpace)
	case DomainAmphibious:
		return hasEffect(attacker.Effects, EffectCanAttackLand) || hasEffect(attacker.Effects, EffectCanAttackSea)
	default:
		return true
	}
}

func bombardEligibleTargets(attacker Unit, targets []*Unit) []*Unit {
	eligible := targets[:0]
	for _, target := range targets {
		if unitCanBombardTarget(attacker, *target) {
			eligible = append(eligible, target)
		}
	}
	return eligible
}

func unitCanBombardTarget(attacker, target Unit) bool {
	if bombarderHasCanBombardMetadata(attacker) {
		return attackerCanBombardTarget(attacker, target)
	}
	return unitCanTarget(attacker, target)
}

func bombarderHasCanBombardMetadata(attacker Unit) bool {
	return hasEffect(attacker.Effects, EffectCanBombardLand) || hasEffect(attacker.Effects, EffectCanBombardSea) || hasEffect(attacker.Effects, EffectCanBombardAir) || hasEffect(attacker.Effects, EffectCanBombardSpace) || hasEffect(attacker.Effects, EffectCanBombardUnderwater)
}

func attackerCanBombardTarget(attacker, target Unit) bool {
	if hasEffect(target.Effects, EffectSubmarine) {
		return hasEffect(attacker.Effects, EffectCanBombardUnderwater)
	}
	switch target.Domain {
	case DomainLand:
		return hasEffect(attacker.Effects, EffectCanBombardLand)
	case DomainSea:
		return hasEffect(attacker.Effects, EffectCanBombardSea)
	case DomainAir:
		return hasEffect(attacker.Effects, EffectCanBombardAir)
	case DomainSpace:
		return hasEffect(attacker.Effects, EffectCanBombardSpace)
	case DomainAmphibious:
		return hasEffect(attacker.Effects, EffectCanBombardLand) || hasEffect(attacker.Effects, EffectCanBombardSea)
	default:
		return true
	}
}

func combatEligibleTargets(attacker Unit, targets []*Unit) []*Unit {
	eligible := targets[:0]
	for _, target := range targets {
		if unitCanTarget(attacker, *target) {
			eligible = append(eligible, target)
		}
	}
	return eligible
}

func unitCanTarget(attacker, target Unit) bool {
	if attackerHasCanAttackMetadata(attacker) {
		return attackerCanAttackTarget(attacker, target)
	}
	return attacker.Domain == "" || target.Domain == "" || attacker.Domain == target.Domain || attacker.Domain == DomainAmphibious || target.Domain == DomainAmphibious
}

func combatStrike(state *State, sources Sources, attackers, defenders []*Unit, ranged bool, result *CombatResult) {
	for _, attacker := range livingCombatUnits(attackers) {
		if (ranged && attacker.Ranged <= 0) || (!ranged && attacker.Attack <= 0) {
			continue
		}
		targets := combatEligibleTargets(*attacker, livingCombatUnits(defenders))
		if len(targets) == 0 {
			continue
		}
		target := targets[0]
		attack := attacker.Attack
		if ranged {
			attack = attacker.Ranged
		}
		attack += combatAttackModifier(*state, *attacker, *target)
		defense := target.Defense + combatDefenseModifier(*state, *target, *attacker)
		if attack <= 0 || defense < 0 {
			continue
		}
		roll := combatRoll(state, sources)
		if int(roll%1000) < attack*1000/(attack+defense) {
			damage := combatDamage(*attacker, *target)
			target.Health -= readinessAdjustedCombatDamage(*state, *target, damage)
		}
	}
}

func combatDamage(attacker, target Unit) int {
	armor := target.Armor
	if armor < 1 {
		armor = 1
	}
	damage := attacker.Strength / armor
	if damage < 1 {
		damage = 1
	}
	firepower := attacker.Firepower
	if firepower < 1 {
		firepower = 1
	}
	return damage * firepower
}

func readinessAdjustedCombatDamage(state State, target Unit, baseDamage int) int {
	if baseDamage < 1 {
		baseDamage = 1
	}
	modifierCentis := readinessHPModifierCentis(state, target.OwnerID)
	if modifierCentis <= 0 {
		modifierCentis = 100
	}
	damage := (baseDamage*100 + modifierCentis - 1) / modifierCentis
	if damage < 1 {
		return 1
	}
	return damage
}

func readinessHPModifierCentis(state State, ownerID string) int {
	nation := nationForID(&state, nationForPlayer(state, ownerID))
	if nation == nil {
		return 100
	}
	kind := EffectReadyWarHP
	switch nation.Readiness {
	case "peace":
		kind = EffectReadyPeaceHP
	case "alert":
		kind = EffectReadyAlertHP
	case "war", "":
		kind = EffectReadyWarHP
	}
	modifier := 0
	for _, effect := range nation.Effects {
		if effect.Kind == kind && effect.Value > modifier {
			modifier = effect.Value
		}
	}
	if modifier <= 0 {
		return 100
	}
	return modifier
}

func combatRoll(state *State, sources Sources) uint64 {
	for i := range state.RNG {
		if state.RNG[i].Domain == RNGDomainCombat {
			next, value := sources.RNG.Next(RNGDomainCombat, state.RNG[i].State)
			state.RNG[i].State = next
			return value
		}
	}
	return 0
}

func combatArmyUnits(state State, army Army) []*Unit {
	units := make([]*Unit, 0, len(army.UnitIDs))
	for _, id := range army.UnitIDs {
		if unit, ok := unitByID(state, id); ok {
			units = append(units, unit)
		}
	}
	sort.Slice(units, func(i, j int) bool { return units[i].ID < units[j].ID })
	return units
}

func livingCombatUnits(units []*Unit) []*Unit {
	living := units[:0]
	for _, unit := range units {
		if unit.Health > 0 {
			living = append(living, unit)
		}
	}
	return living
}

func combatDefenseModifier(state State, unit Unit, attacker Unit) int {
	modifier := combatModifier(state, unit, false)
	if hasEffect(attacker.Effects, EffectIgnoreCityWalls) {
		modifier -= cityWallDefenseAt(state, unit.TileID)
	}
	if hasEffect(attacker.Effects, EffectMounted) {
		modifier += unit.Defense * effectValue(unit.Effects, EffectCombatDefenseVsMountedPercent) / 100
	}
	if attacker.Domain == DomainAir {
		modifier += unit.Defense * effectValue(unit.Effects, EffectCombatDefenseVsAirPercent) / 100
	}
	return modifier
}

func cityWallDefenseAt(state State, tileID string) int {
	for _, city := range state.Cities {
		if city.TileID == tileID {
			return effectValue(cityLocalEffects(state, city), EffectCityWalls)
		}
	}
	return 0
}

func combatAttackModifier(state State, unit Unit, target Unit) int {
	modifier := combatModifier(state, unit, true)
	if hasEffect(target.Effects, EffectWoodenShip) {
		modifier += unit.Attack * effectValue(unit.Effects, EffectCombatAttackVsWoodenShipPercent) / 100
	}
	for _, city := range state.Cities {
		if city.TileID == unit.TileID {
			modifier += effectsModifierAgainst(cityLocalEffects(state, city), true, target.Domain)
		}
	}
	return modifier
}

func combatModifier(state State, unit Unit, attack bool) int {
	modifier := effectsModifierAgainst(unit.Effects, attack, "")
	if !attack && unit.Fortified {
		modifier++
	}
	if !attack {
		if tile, ok := tileByID(state, unit.TileID); ok && tile.Terrain == TerrainForest {
			modifier++
		}
		for _, city := range state.Cities {
			if city.TileID == unit.TileID {
				modifier += effectsModifierAgainst(cityLocalEffects(state, city), false, "")
			}
		}
	}
	for _, player := range state.Players {
		if player.ID == unit.OwnerID {
			for _, nation := range state.Nations {
				if nation.ID == player.NationID {
					modifier += effectsModifierAgainst(nation.Effects, attack, "")
				}
			}
		}
	}
	return modifier
}

func effectsModifierAgainst(effects []Effect, attack bool, targetDomain string) int {
	kind := EffectCombatDefense
	if attack {
		kind = EffectCombatAttack
	}
	value := 0
	for _, effect := range effects {
		if effect.Kind == kind || effect.Kind == EffectCombat || attack && combatAttackDomainEffectApplies(effect.Kind, targetDomain) {
			value += effect.Value
		}
	}
	return value
}

func combatAttackDomainEffectApplies(kind, targetDomain string) bool {
	switch kind {
	case EffectCombatAttackLand:
		return targetDomain == DomainLand || targetDomain == DomainAmphibious
	case EffectCombatAttackSea:
		return targetDomain == DomainSea || targetDomain == DomainAmphibious
	case EffectCombatAttackAir:
		return targetDomain == DomainAir
	default:
		return false
	}
}

func combatVisible(state State, playerID, tileID string) bool {
	for _, seen := range state.Fog {
		if seen.PlayerID == playerID && seen.TileID == tileID {
			return true
		}
	}
	return false
}
func adjacentTiles(state State, fromID, toID string) bool {
	from, a := tileByID(state, fromID)
	to, b := tileByID(state, toID)
	return a && b && adjacent(from, to)
}
func moveArmyToTile(state *State, armyID, tileID string) {
	for i := range state.Units {
		if state.Units[i].ArmyID == armyID {
			state.Units[i].TileID = tileID
			state.Units[i].Fortified = false
		}
	}
}
