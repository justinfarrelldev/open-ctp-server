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
		if unit.Attack < 0 || unit.Defense < 0 || unit.Ranged < 0 || unit.Strength <= 0 || unit.Armor <= 0 || unit.Health <= 0 || unit.MaxHealth < unit.Health {
			return nil, nil, errors.New("combat unit is invalid")
		}
	}
	if !canArmyFight(*state, attackerUnits, defenderUnits) {
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
			removeUnitAndCargo(state, unit.ID)
			result.AttackerCasualties++
		}
	}
	for _, unit := range combatArmyUnits(*state, *defenders) {
		if unit.Health <= 0 {
			removeUnitAndCargo(state, unit.ID)
			result.DefenderCasualties++
		}
	}
	if _, ok := armyByID(state, attackers.ID); !ok {
		result.Outcome = CombatOutcomeDefenderHeld
	} else if _, ok := armyByID(state, defenders.ID); !ok {
		result.Outcome = CombatOutcomeAttackerWon
		if armyCanCaptureTargetCity(*state, *attackers, targetTileID) {
			moveArmyToTile(state, attackers.ID, targetTileID)
			result.CapturedCityID = captureCityAt(state, command.ActorPlayerID, targetTileID)
		}
	} else {
		result.Outcome = CombatOutcomeRetreated
	}
	eventPayload, err := json.Marshal(result)
	if err != nil {
		return nil, nil, err
	}
	return result, eventPayload, nil
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
		if defenderCity {
			return true
		}
		for _, defender := range defenders {
			if attacker.Domain == "" || defender.Domain == "" || attacker.Domain == defender.Domain || attacker.Domain == DomainAmphibious || defender.Domain == DomainAmphibious {
				return true
			}
		}
	}
	return false
}

func combatStrike(state *State, sources Sources, attackers, defenders []*Unit, ranged bool, result *CombatResult) {
	for _, attacker := range livingCombatUnits(attackers) {
		if (ranged && attacker.Ranged <= 0) || (!ranged && attacker.Attack <= 0) {
			continue
		}
		targets := livingCombatUnits(defenders)
		if len(targets) == 0 {
			return
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
			damage := attacker.Strength / target.Armor
			if damage < 1 {
				damage = 1
			}
			target.Health -= damage
		}
	}
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
