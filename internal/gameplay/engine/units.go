package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const maxArmySize = 12

var ErrObstructed = errors.New("movement obstructed")

type moveUnitPayload struct {
	UnitID  string   `json:"unit_id"`
	Path    []string `json:"path,omitempty"`
	TargetX *int     `json:"target_x,omitempty"`
	TargetY *int     `json:"target_y,omitempty"`
}

type unitPayload struct {
	UnitID string `json:"unit_id"`
}

type upgradeUnitPayload struct {
	UnitID    string `json:"unit_id"`
	UpgradeTo string `json:"upgrade_to"`
}

type embarkPayload struct {
	UnitID      string `json:"unit_id"`
	TransportID string `json:"transport_id"`
}

type disembarkPayload struct {
	UnitID string `json:"unit_id"`
	TileID string `json:"tile_id"`
}

type giftUnitPayload struct {
	UnitID         string `json:"unit_id"`
	TargetPlayerID string `json:"target_player_id"`
}

type pirateTradePayload struct {
	UnitID  string `json:"unit_id"`
	RouteID string `json:"route_id"`
}

type spaceLaunchPayload struct {
	UnitID       string `json:"unit_id"`
	TargetTileID string `json:"target_tile_id"`
}

type nuclearAttackPayload struct {
	UnitID       string `json:"unit_id"`
	TargetTileID string `json:"target_tile_id"`
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
		payload, err := decodeMovePayload(command.Payload)
		if err != nil {
			return "", err
		}
		return EventUnitMoved, moveUnit(state, command.ActorPlayerID, payload, runtime)
	case CommandFortify:
		var payload unitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		unit, err := ownedUnit(state, command.ActorPlayerID, payload.UnitID)
		if err != nil || unit.TransportID != "" || !unitCanFortify(*unit) {
			return "", invalidUnitAction(err)
		}
		unit.Fortified, unit.Asleep = true, false
		return "unit.fortified", nil
	case CommandSleep, CommandWake:
		var payload unitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		unit, err := ownedUnit(state, command.ActorPlayerID, payload.UnitID)
		if err != nil || unit.InSpace || unit.TransportID != "" {
			return "", invalidUnitAction(err)
		}
		unit.Asleep = command.Type == CommandSleep
		if unit.Asleep {
			return "unit.slept", nil
		}
		return "unit.woke", nil
	case CommandUpgradeUnit:
		var payload upgradeUnitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return EventUnitUpgraded, upgradeUnit(state, command.ActorPlayerID, payload, runtime)
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
	case CommandGiftUnit:
		var payload giftUnitPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return EventUnitGifted, giftUnit(state, command.ActorPlayerID, payload)
	case CommandPirateTrade:
		var payload pirateTradePayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return EventTradeRoutePirated, pirateTrade(state, command.ActorPlayerID, payload, runtime)
	case CommandSpaceLaunch:
		var payload spaceLaunchPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return EventUnitLaunched, spaceLaunch(state, command.ActorPlayerID, payload)
	case CommandNuclearAttack:
		var payload nuclearAttackPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return nuclearAttack(state, command.ActorPlayerID, payload, runtime)
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

// ErrPhaseRestricted is safe to classify without disclosing mechanical causes.
var ErrPhaseRestricted = errors.New("action is not eligible in this phase")

func activeActor(state State, actorID string) error {
	if (state.Phase.Mode == "simultaneous" || state.Phase.Mode == "hybrid") && state.Phase.ResolutionLocked {
		return nil
	}
	if state.Phase.ResolutionLocked || state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex] != actorID {
		return ErrPhaseRestricted
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

func upgradeUnit(state *State, actorID string, payload upgradeUnitPayload, runtime *rules.RuntimeRules) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || payload.UpgradeTo == "" || runtime == nil || unit.TransportID != "" {
		return invalidUnitAction(err)
	}
	currentID := unit.TypeID
	if currentID == "" {
		currentID = producedUnitTypeID(unit.ID)
	}
	current, ok := runtime.Unit(currentID)
	if currentID == "" || !ok || !containsString(current.UpgradeTo, payload.UpgradeTo) {
		return errors.New("unit upgrade is invalid")
	}
	target, ok := runtime.Unit(payload.UpgradeTo)
	if !ok || !playerHasAllPrerequisites(*state, actorID, target.Prerequisites) || !playerMatchesUnitGovernment(*state, actorID, target) {
		return errors.New("unit upgrade is invalid")
	}
	applyUnitDefinition(unit, target)
	return nil
}

func producedUnitTypeID(unitID string) string {
	parts := strings.Split(unitID, ":")
	if len(parts) >= 4 && parts[0] == "unit" {
		return parts[2]
	}
	return ""
}

func playerMatchesUnitGovernment(state State, playerID string, unit rules.RuntimeUnitDefinition) bool {
	requiredGovernmentIDs := unitRequiredGovernmentIDs(unit)
	if len(requiredGovernmentIDs) == 0 {
		return true
	}
	nation := nationForID(&state, nationForPlayer(state, playerID))
	return nation != nil && containsString(requiredGovernmentIDs, nation.GovernmentID)
}

func playerMatchesRequiredGovernment(state State, playerID, requiredGovernmentID string) bool {
	return playerMatchesUnitGovernment(state, playerID, rules.RuntimeUnitDefinition{RequiredGovernmentID: requiredGovernmentID})
}

func unitRequiredGovernmentIDs(unit rules.RuntimeUnitDefinition) []string {
	if len(unit.RequiredGovernmentIDs) > 0 {
		return unit.RequiredGovernmentIDs
	}
	if unit.RequiredGovernmentID != "" {
		return []string{unit.RequiredGovernmentID}
	}
	return nil
}

func playerHasAllPrerequisites(state State, playerID string, prerequisites []string) bool {
	if len(prerequisites) == 0 {
		return true
	}
	nationID := nationForPlayer(state, playerID)
	for _, nation := range state.Nations {
		if nation.ID != nationID {
			continue
		}
		for _, prerequisite := range prerequisites {
			if !containsString(nation.Advances, prerequisite) && !wonderExists(state, prerequisite) {
				return false
			}
		}
		return true
	}
	return false
}

func applyUnitDefinition(unit *Unit, definition rules.RuntimeUnitDefinition) {
	oldHealth, oldMaxHealth := unit.Health, unit.MaxHealth
	unit.TypeID = definition.ID
	unit.Domain = definition.Domain
	unit.Movement = 0
	unit.MaxMovement = definition.Movement
	unit.Fuel = definition.MaxFuel
	unit.MaxFuel = definition.MaxFuel
	unit.CargoCapacity = definition.TransportCapacity
	unit.CargoDomains = append([]string(nil), definition.CargoDomains...)
	unit.CargoClasses = append([]string(nil), definition.CargoClasses...)
	unit.Size = definition.Size
	unit.Upkeep = definition.Upkeep
	unit.ShieldHunger = definition.ShieldHunger
	unit.FoodHunger = definition.FoodHunger
	unit.NeedsNoSupport = definition.NeedsNoSupport
	unit.IsSpecialForces = definition.IsSpecialForces
	unit.ZOC = definition.ZOC
	unit.Effects = runtimeUnitEffects(definition)
	unit.Attack = definition.Attack
	unit.Defense = definition.Defense
	unit.Ranged = definition.Ranged
	unit.BombardRange = definition.BombardRange
	unit.BombRounds = definition.BombRounds
	unit.ActiveDefenseRange = definition.ActiveDefenseRange
	unit.Strength = definition.Strength
	unit.Armor = definition.Armor
	unit.Firepower = definition.Firepower
	unit.MaxHealth = definition.Health
	if unit.MaxHealth < 1 {
		unit.MaxHealth = 1
	}
	if oldMaxHealth > 0 && oldHealth > 0 {
		unit.Health = oldHealth * unit.MaxHealth / oldMaxHealth
		if unit.Health < 1 {
			unit.Health = 1
		}
	} else {
		unit.Health = unit.MaxHealth
	}
	unit.VisibilityRange = definition.VisibilityRange
	unit.SettleCityTypeID = definition.SettleCityTypeID
	unit.Fortified = false
}

func invalidUnitAction(err error) error {
	if err != nil {
		return err
	}
	return errors.New("unit action is invalid")
}

func unitCanFortify(unit Unit) bool {
	if hasEffect(unit.Effects, EffectCanEntrench) {
		return true
	}
	return !unitHasRetailCapabilityMetadata(unit)
}

func spaceLaunch(state *State, actorID string, payload spaceLaunchPayload) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	turns := unitSpaceLaunchTurns(unit)
	if err != nil || payload.TargetTileID == "" || unit.TransportID != "" || unit.InSpace || turns < 1 {
		return invalidUnitAction(err)
	}
	if unit.ArmyID != "" {
		leaveUnitArmy(state, unit.ID)
	}
	city, ok := cityByTile(*state, payload.TargetTileID)
	if !ok || city.OwnerID != actorID {
		return errors.New("space launch destination must be an owned city")
	}
	unit.TileID = ""
	unit.Movement = 0
	unit.Fortified = false
	unit.InSpace = true
	unit.ReentryTileID = payload.TargetTileID
	unit.ReentryTurns = turns
	return nil
}

func nuclearAttack(state *State, actorID string, payload nuclearAttackPayload, runtime *rules.RuntimeRules) (string, error) {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || payload.TargetTileID == "" || unit.TransportID != "" || unit.InSpace || !hasEffect(unit.Effects, EffectNuclearAttack) {
		return "", invalidUnitAction(err)
	}
	city, ok := cityByTile(*state, payload.TargetTileID)
	if !ok || city.OwnerID == actorID {
		return "", errors.New("nuclear attack target must be a foreign city")
	}
	if cityProtectedFromNuclearAttack(*state, city) {
		removeUnitAndCargo(state, unit.ID)
		return EventNuclearAttackFailed, nil
	}
	detonateNukeAtCity(state, city.ID, unit.ID, runtime)
	removeUnitAndCargo(state, unit.ID)
	return EventNuclearAttack, nil
}

func cityProtectedFromNuclearAttack(state State, city City) bool {
	effects := cityEffects(state, city)
	return hasEffect(effects, EffectProtectFromNukes) || hasEffect(effects, EffectForceField)
}

func detonateNukeAtCity(state *State, cityID, sourceUnitID string, runtime *rules.RuntimeRules) {
	city, ok := cityByID(*state, cityID)
	if !ok {
		return
	}
	cityTileID := city.TileID
	if city.Population >= 3 {
		killed := city.Population * nukePopulationPercent(runtime) / 100
		if killed < 1 {
			killed = 1
		}
		city.Population -= killed
		if city.Population < 1 {
			city.Population = 1
		}
	} else {
		removeCityByID(state, cityID)
	}
	if tile := tilePointer(state, cityTileID); tile != nil {
		tile.Pollution += nukePollution(runtime)
	}
	addNukeGlobalPollution(state, runtime)
	resolveNukeBlastUnits(state, sourceUnitID, cityTileID)
	resolveNukeBlastImprovements(state, cityTileID)
}

func unitSpaceLaunchTurns(unit *Unit) int {
	if unit == nil {
		return 0
	}
	turns := 0
	for _, effect := range unit.Effects {
		if effect.Kind == EffectSpaceLaunchTurns && effect.Value > turns {
			turns = effect.Value
		}
	}
	return turns
}

func cityByTile(state State, tileID string) (City, bool) {
	for _, city := range state.Cities {
		if city.TileID == tileID {
			return city, true
		}
	}
	return City{}, false
}

func unitHasRetailCapabilityMetadata(unit Unit) bool {
	for _, effect := range unit.Effects {
		switch effect.Kind {
		case EffectCantCaptureCity, EffectNoFuelThenCrash, EffectCanPillage, EffectBuildingRemovesPopulation, EffectIgnoreZOC, EffectCanAttackLand, EffectCanAttackSea, EffectCanAttackAir, EffectCanAttackSpace, EffectCanAttackUnderwater, EffectSubmarine, EffectWoodenShip,
			EffectSpecialEmbassy, EffectSpecialSoothsaying, EffectSpecialSteal, EffectSpecialStealSpecificChance, EffectSpecialCapture, EffectSpecialTerraform, EffectSpecialTerraformTurns,
			EffectSpecialConvertCity, EffectSpecialCreateFranchise, EffectSpecialInvestigateCity, EffectSpecialPlantNuke, EffectSpecialCauseUnhappiness, EffectSpecialIndulgenceSales, EffectSpecialInciteRevolution, EffectSpecialSlaveUprising, EffectSpecialBioTerror, EffectSpecialNanoTerror, EffectSpecialPlague,
			EffectSpecialAssassinateRuler, EffectSpecialSlaveRaid, EffectSpecialSettlerSlaveRaid, EffectSpecialUndergroundRailway, EffectSpecialInjoin, EffectSpecialReform,
			EffectSpecialCreatePark, EffectSpecialExpel, EffectSpecialSue:
			return true
		}
	}
	return false
}

func moveUnit(state *State, actorID string, payload moveUnitPayload, runtime *rules.RuntimeRules) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TransportID != "" {
		return invalidUnitAction(err)
	}
	if payload.TargetX != nil || payload.TargetY != nil {
		if payload.TargetX == nil || payload.TargetY == nil || len(payload.Path) != 0 {
			return ErrObstructed
		}
		for _, tile := range state.Tiles {
			if tile.X == *payload.TargetX && tile.Y == *payload.TargetY {
				payload.Path = []string{tile.ID}
				break
			}
		}
	}
	if len(payload.Path) == 0 {
		return ErrObstructed
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
	visitedColumns := []int{from.X}
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
		visitedColumns = append(visitedColumns, to.X)
		if !moversIgnoreZOC(movers) && inEnemyZOC(*state, runtime, actorID, to.ID, domains) && index != len(payload.Path)-1 {
			return ErrObstructed
		}
		from = to
	}
	spent := unit.Movement - remaining
	for _, mover := range movers {
		if mover.Movement < spent || (mover.MaxFuel > 0 && mover.Fuel < spent) {
			return ErrObstructed
		}
	}
	for _, mover := range movers {
		mover.TileID, mover.Movement, mover.Fortified = from.ID, mover.Movement-spent, false
		mover.Asleep = false
		if mover.MaxFuel > 0 {
			mover.Fuel -= spent
		}
		recordCircumnavigationColumns(state, mover, visitedColumns, runtime)
	}
	captureCityAt(state, actorID, from.ID, runtime)
	state.Fog = append(state.Fog, FogObservation{PlayerID: actorID, TileID: from.ID, Turn: state.Phase.Turn})
	for _, tileID := range payload.Path {
		exploreRuin(state, actorID, tileID, runtime)
	}
	return nil
}

// decodeMovePayload accepts either a path or a complete coordinate pair. Wire
// shape validation is shared by immediate execution, planning and replay.
func decodeMovePayload(raw json.RawMessage) (moveUnitPayload, error) {
	var payload moveUnitPayload
	if err := decodePayload(raw, &payload); err != nil {
		return payload, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return payload, ErrObstructed
	}
	_, x := fields["target_x"]
	_, y := fields["target_y"]
	_, path := fields["path"]
	// Presence, not length, selects the wire shape. Even []/null cannot
	// silently combine a path field with coordinate input.
	if (x || y) && (payload.TargetX == nil || payload.TargetY == nil || path) {
		return payload, ErrObstructed
	}
	return payload, nil
}

func recordCircumnavigationColumns(state *State, unit *Unit, columns []int, runtime *rules.RuntimeRules) {
	if state == nil || unit == nil || (unit.Domain != DomainSea && unit.Domain != DomainAmphibious) || len(columns) == 0 {
		return
	}
	for _, column := range columns {
		if !containsInt(unit.CircumnavigatedColumns, column) {
			unit.CircumnavigatedColumns = append(unit.CircumnavigatedColumns, column)
		}
	}
	sort.Ints(unit.CircumnavigatedColumns)
	if runtime == nil || !hasEveryMapColumn(*state, unit.CircumnavigatedColumns) {
		return
	}
	nation := nationForID(state, nationForPlayer(*state, unit.OwnerID))
	if nation == nil {
		return
	}
	feat, ok := runtime.Feat("FEAT_SAILED_AROUND_WORLD")
	if !ok {
		return
	}
	applyRuntimeFeat(state, nation, feat)
}

func hasEveryMapColumn(state State, columns []int) bool {
	if len(state.Tiles) == 0 || len(columns) == 0 {
		return false
	}
	required := map[int]bool{}
	for _, tile := range state.Tiles {
		required[tile.X] = true
	}
	for _, column := range columns {
		delete(required, column)
	}
	return len(required) == 0
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func captureCityAt(state *State, actorID, tileID string, runtimes ...*rules.RuntimeRules) string {
	var runtime *rules.RuntimeRules
	if len(runtimes) > 0 {
		runtime = runtimes[0]
	}
	for i := range state.Cities {
		if state.Cities[i].TileID == tileID && state.Cities[i].OwnerID != actorID {
			cityID := state.Cities[i].ID
			defeatedOwnerID := state.Cities[i].OwnerID
			defeatedCityCount := ownerCityCount(*state, defeatedOwnerID)
			recapturedByFounder := state.Cities[i].FounderID != "" && state.Cities[i].FounderID == actorID
			removeCapturedCapitolBuildings(&state.Cities[i])
			resetCapturedCityEffects(&state.Cities[i])
			state.Cities[i].OwnerID = actorID
			state.Cities[i].Governor = nil
			normalizeCityConversionOwner(&state.Cities[i])
			state.Cities[i].ConquestDistress = conquestDistressForOwner(*state, actorID)
			state.TradeRoutes = removeTradeRoutesTouchingCity(state.TradeRoutes, cityID)
			transferCityWondersToOwner(state, cityID, actorID, runtime)
			applyConqueredByForceFeat(state, actorID, defeatedOwnerID, defeatedCityCount, runtime)
			if recapturedByFounder {
				applyCityRecapturedFeat(state, actorID, runtime)
			}
			transferCityTerritory(state, cityID, actorID, runtime)
			return cityID
		}
	}
	return ""
}

func resetCapturedCityEffects(city *City) {
	kept := city.Effects[:0]
	for _, effect := range city.Effects {
		if effect.Turns > 0 {
			continue
		}
		switch effect.Kind {
		case EffectFranchise, EffectInjunction, EffectWatchful:
			continue
		default:
			kept = append(kept, effect)
		}
	}
	city.Effects = kept
}

func removeCapturedCapitolBuildings(city *City) {
	kept := city.Buildings[:0]
	for _, building := range city.Buildings {
		if hasEffect(building.Effects, EffectCapitol) {
			continue
		}
		kept = append(kept, building)
	}
	city.Buildings = kept
}

func transferCityWondersToOwner(state *State, cityID, newOwnerID string, runtime *rules.RuntimeRules) {
	for index := range state.Wonders {
		wonder := &state.Wonders[index]
		if wonder.CityID != cityID || wonder.Destroyed || wonder.OwnerID == newOwnerID {
			continue
		}
		wonder.OwnerID = newOwnerID
		applyWonderCompletionEffects(state, *wonder, runtime)
	}
}

func ownerCityCount(state State, ownerID string) int {
	count := 0
	for _, city := range state.Cities {
		if city.OwnerID == ownerID {
			count++
		}
	}
	return count
}

func applyConqueredByForceFeat(state *State, conquerorPlayerID, defeatedPlayerID string, defeatedCityCount int, runtime *rules.RuntimeRules) {
	if state == nil || runtime == nil || defeatedCityCount != 1 {
		return
	}
	conquerorNation := nationForID(state, nationForPlayer(*state, conquerorPlayerID))
	defeatedNation := nationForID(state, nationForPlayer(*state, defeatedPlayerID))
	if conquerorNation == nil || defeatedNation == nil {
		return
	}
	maxCityCount := defeatedNation.MaxCityCount
	if defeatedCityCount > maxCityCount {
		maxCityCount = defeatedCityCount
	}
	for _, feat := range runtime.Feats {
		if feat.ID != "FEAT_CONQUERED_BY_FORCE" || feat.MinimumCivSize < 1 || maxCityCount < feat.MinimumCivSize {
			continue
		}
		applyRuntimeFeat(state, conquerorNation, feat)
		return
	}
}

func applyCityRecapturedFeat(state *State, founderPlayerID string, runtime *rules.RuntimeRules) {
	if state == nil || runtime == nil {
		return
	}
	founderNation := nationForID(state, nationForPlayer(*state, founderPlayerID))
	if founderNation == nil {
		return
	}
	for _, feat := range runtime.Feats {
		if feat.ID != "FEAT_CITY_RECAPTURED" {
			continue
		}
		applyRuntimeFeat(state, founderNation, feat)
		return
	}
}

func removeTradeRoutesTouchingCity(routes []TradeRoute, cityID string) []TradeRoute {
	kept := routes[:0]
	for _, route := range routes {
		if route.SourceCityID == cityID || route.DestinationCityID == cityID {
			continue
		}
		kept = append(kept, route)
	}
	return kept
}

func conquestDistressForOwner(state State, ownerID string) int {
	nation := nationForID(&state, nationForPlayer(state, ownerID))
	if nation == nil {
		return 0
	}
	distress := 0
	for _, effect := range nation.Effects {
		if effect.Kind == EffectConquestDistress {
			distress += effect.Value
		}
	}
	return distress
}

func embark(state *State, actorID string, payload embarkPayload) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TransportID != "" || unit.ArmyID != "" {
		return invalidUnitAction(err)
	}
	transport, err := ownedUnit(state, actorID, payload.TransportID)
	if err != nil || transport.ID == unit.ID || transport.TransportID != "" || unit.TileID != transport.TileID || transport.CargoCapacity <= cargoCount(*state, transport.ID) || !transportCanCarryUnit(*transport, *unit) {
		return ErrObstructed
	}
	unit.TransportID, unit.TileID, unit.Fortified = transport.ID, "", false
	return nil
}

func transportCanCarryUnit(transport Unit, unit Unit) bool {
	if len(transport.CargoClasses) > 0 {
		for _, cargoClass := range unitCargoClasses(unit) {
			if containsString(transport.CargoClasses, cargoClass) {
				return true
			}
		}
		return false
	}
	if containsString(transport.CargoDomains, unit.Domain) {
		return true
	}
	if unit.Domain == DomainAmphibious {
		return containsString(transport.CargoDomains, DomainLand) || containsString(transport.CargoDomains, DomainSea)
	}
	return false
}

func unitCargoClasses(unit Unit) []string {
	size := unit.Size
	if size == "" {
		size = "Small"
	}
	size = strings.TrimSpace(size)
	if strings.EqualFold(size, "Medium") {
		size = "Med"
	}
	switch strings.ToLower(unit.Domain) {
	case DomainLand:
		return []string{size + "Land"}
	case DomainAir:
		return []string{size + "Air"}
	case DomainSea:
		return []string{size + "Water"}
	case DomainAmphibious:
		return []string{size + "Land", size + "Water"}
	case DomainSpace:
		return []string{size + "Space"}
	default:
		return []string{size + unit.Domain}
	}
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
	remainingMovement := unit.Movement - cost
	if !cargoKeepsMovementAfterDisembark(*unit, *transport) {
		remainingMovement = 0
	}
	unit.TransportID, unit.TileID, unit.Movement = "", to.ID, remainingMovement
	return nil
}

func cargoKeepsMovementAfterDisembark(unit Unit, transport Unit) bool {
	if unit.Domain == DomainAir {
		return true
	}
	if hasEffect(unit.Effects, EffectParatrooper) && hasEffect(transport.Effects, EffectParatrooperTransport) {
		return true
	}
	return hasEffect(unit.Effects, EffectCanBeachAssault) && transport.Domain == DomainSea
}

func giftUnit(state *State, actorID string, payload giftUnitPayload) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	target, ok := playerByID(state, payload.TargetPlayerID)
	capital, hasCapital := capitalCity(*state, actorID)
	if err != nil || !ok || target.Lifecycle != PlayerActive || payload.TargetPlayerID == actorID || unit.TileID == "" || unit.TransportID != "" || unit.InSpace || !hasEffect(unit.Effects, EffectCanBeGifted) || !hasCapital {
		return invalidUnitAction(err)
	}
	if cargoCount(*state, unit.ID) > 0 {
		return ErrObstructed
	}
	leaveUnitArmy(state, unit.ID)
	unit.OwnerID = payload.TargetPlayerID
	unit.TileID = capital.TileID
	unit.TransportID = ""
	unit.Fortified = false
	return nil
}

func pirateTrade(state *State, actorID string, payload pirateTradePayload, runtime *rules.RuntimeRules) error {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || payload.RouteID == "" || unit.TileID == "" || unit.TransportID != "" || unit.InSpace || unit.ArmyID == "" || !hasEffect(unit.Effects, EffectCanPirate) {
		return errors.New("piracy action is invalid")
	}
	army, ok := armyByID(state, unit.ArmyID)
	if !ok || army.OwnerID != actorID {
		return errors.New("piracy action is invalid")
	}
	route, ok := tradeRouteByID(state, payload.RouteID)
	if !ok || !tradeRouteTouchesTile(*state, *route, unit.TileID) || tradeRouteSourceOwner(*state, *route) == actorID {
		return errors.New("piracy action is invalid")
	}
	if err := spendSpecialMove(unit, runtime); err != nil {
		return err
	}
	route.PiratingArmyID = unit.ArmyID
	return nil
}

func tradeRouteByID(state *State, routeID string) (*TradeRoute, bool) {
	for index := range state.TradeRoutes {
		if state.TradeRoutes[index].ID == routeID {
			return &state.TradeRoutes[index], true
		}
	}
	return nil, false
}

func tradeRouteTouchesTile(state State, route TradeRoute, tileID string) bool {
	source, sourceOK := cityByID(state, route.SourceCityID)
	destination, destinationOK := cityByID(state, route.DestinationCityID)
	return (sourceOK && source.TileID == tileID) || (destinationOK && destination.TileID == tileID)
}

func tradeRouteSourceOwner(state State, route TradeRoute) string {
	city, ok := cityByID(state, route.SourceCityID)
	if !ok {
		return ""
	}
	return city.OwnerID
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

func leaveUnitArmy(state *State, unitID string) {
	unit, ok := unitByID(*state, unitID)
	if !ok || unit.ArmyID == "" {
		return
	}
	army, ok := armyByID(state, unit.ArmyID)
	if !ok {
		unit.ArmyID = ""
		return
	}
	army.UnitIDs = removeString(army.UnitIDs, unit.ID)
	unit.ArmyID = ""
	if len(army.UnitIDs) == 0 {
		state.Armies = removeArmy(state.Armies, army.ID)
	}
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
		if unit.ID == "" || unit.OwnerID == "" || (unit.Domain != "" && !supportedDomain(unit.Domain)) || unit.Movement < 0 || unit.MaxMovement < 0 || unit.Fuel < 0 || unit.MaxFuel < 0 || (unit.MaxFuel == 0 && unit.Fuel != 0) || (unit.MaxFuel > 0 && unit.Fuel > unit.MaxFuel) || unit.CargoCapacity < 0 || unit.Upkeep < 0 || unit.ShieldHunger < 0 || unit.FoodHunger < 0 || unit.BombardRange < 0 || unit.BombRounds < 0 || unit.ActiveDefenseRange < 0 || unit.VisibilityRange < 0 || unit.Firepower < 0 || unit.ReentryTurns < 0 {
			return errors.New("unit capability is invalid")
		}
		if unit.InSpace && (unit.TileID != "" || unit.TransportID != "" || unit.ReentryTileID == "") {
			return errors.New("space unit reentry state is invalid")
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
	return domain == DomainLand || domain == DomainSea || domain == DomainAmphibious || domain == DomainAir || domain == DomainSpace
}
func supportedTerrain(terrain string) bool {
	if strings.HasPrefix(terrain, "TERRAIN_") {
		return true
	}
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
	if tile, ok := tileByID(state, tileID); ok {
		if tile.River && terrain.RiverCost > 0 && (cost == 0 || terrain.RiverCost < cost) {
			cost = terrain.RiverCost
		}
		if _, cityOnTile := cityByTile(state, tileID); cityOnTile && terrain.CityCost > 0 && (cost == 0 || terrain.CityCost < cost) {
			cost = terrain.CityCost
		}
	}
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
func moversIgnoreZOC(movers []*Unit) bool {
	if len(movers) == 0 {
		return false
	}
	for _, mover := range movers {
		if !hasEffect(mover.Effects, EffectIgnoreZOC) {
			return false
		}
	}
	return true
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
