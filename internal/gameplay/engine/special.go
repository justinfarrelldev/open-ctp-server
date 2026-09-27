package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	retailUnconvertedIndulgenceGold      = 3
	retailUnconvertedIndulgenceHappiness = 2

	EffectSpecialEmbassy                       = "special_embassy"
	EffectSpecialSoothsaying                   = "special_soothsaying"
	EffectSpecialSteal                         = "special_steal"
	EffectSpecialStealSpecificChance           = "special_steal_specific_chance"
	EffectSpecialCapture                       = "special_capture"
	EffectSpecialTerraform                     = "special_terraform"
	EffectSpecialTerraformTurns                = "special_terraform_turns"
	EffectSpecialConvertCity                   = "special_convert_city"
	EffectSpecialCreateFranchise               = "special_create_franchise"
	EffectFranchise                            = "franchise"
	EffectSpecialInvestigateCity               = "special_investigate_city"
	EffectSpecialPlantNuke                     = "special_plant_nuke"
	EffectSpecialCauseUnhappiness              = "special_cause_unhappiness"
	EffectSpecialCauseUnhappinessTimer         = "special_cause_unhappiness_timer"
	EffectSpecialCauseUnhappinessAmount        = "special_cause_unhappiness_amount"
	EffectSpecialIndulgenceSales               = "special_indulgence_sales"
	EffectSpecialInciteRevolution              = "special_incite_revolution"
	EffectSpecialInciteRevolutionEliteChance   = "special_incite_revolution_elite_chance"
	EffectSpecialInciteRevolutionDeathChance   = "special_incite_revolution_death_chance"
	EffectSpecialSlaveUprising                 = "special_slave_uprising"
	EffectSpecialBioTerror                     = "special_bio_terror"
	EffectSpecialNanoTerror                    = "special_nano_terror"
	EffectSpecialPlague                        = "special_plague"
	EffectSpecialAssassinateRuler              = "special_assassinate_ruler"
	EffectRulerAssassinated                    = "ruler_assassinated"
	EffectSpecialSlaveRaid                     = "special_slave_raid"
	EffectSpecialSettlerSlaveRaid              = "special_settler_slave_raid"
	EffectSpecialUndergroundRailway            = "special_underground_railway"
	EffectFreedPopulation                      = "freed_population"
	EffectSpecialInjoin                        = "special_injoin"
	EffectInjunction                           = "injunction"
	EffectSpecialReform                        = "special_reform"
	EffectSpecialCreatePark                    = "special_create_park"
	EffectSpecialExpel                         = "special_expel"
	EffectSpecialSue                           = "special_sue"
	EffectDefendAgainstSpies                   = "defend_against_spies"
	EffectSpecialConvertCityDeathChance        = EffectSpecialConvertCity + "_death_chance"
	EffectSpecialCreateFranchiseDeathChance    = EffectSpecialCreateFranchise + "_death_chance"
	EffectSpecialInvestigateCityDeathChance    = EffectSpecialInvestigateCity + "_death_chance"
	EffectSpecialPlantNukeDeathChance          = EffectSpecialPlantNuke + "_death_chance"
	EffectSpecialPlantNukeEscapeChance         = EffectSpecialPlantNuke + "_escape_chance"
	EffectSpecialBioTerrorDeathChance          = EffectSpecialBioTerror + "_death_chance"
	EffectSpecialNanoTerrorDeathChance         = EffectSpecialNanoTerror + "_death_chance"
	EffectSpecialPlagueDeathChance             = EffectSpecialPlague + "_death_chance"
	EffectSpecialAssassinateRulerDeathChance   = EffectSpecialAssassinateRuler + "_death_chance"
	EffectSpecialSlaveRaidDeathChance          = EffectSpecialSlaveRaid + "_death_chance"
	EffectSpecialSettlerSlaveRaidDeathChance   = EffectSpecialSettlerSlaveRaid + "_death_chance"
	EffectSpecialUndergroundRailwayDeathChance = EffectSpecialUndergroundRailway + "_death_chance"
	EffectLawsuit                              = "lawsuit"
	EffectWatchful                             = "watchful"
	EffectSoothsaying                          = "soothsaying"
)

type SpecialActionPayload struct {
	UnitID       string `json:"unit_id"`
	CityID       string `json:"city_id"`
	TargetUnitID string `json:"target_unit_id,omitempty"`
}

type PillagePayload struct {
	UnitID string `json:"unit_id"`
}

type TerraformPayload struct {
	UnitID        string `json:"unit_id"`
	TileID        string `json:"tile_id"`
	TargetTerrain string `json:"target_terrain"`
}

type BuildImprovementPayload struct {
	UnitID        string `json:"unit_id"`
	TileID        string `json:"tile_id"`
	ImprovementID string `json:"improvement_id"`
}

type CancelTerraformPayload struct {
	TileID string `json:"tile_id"`
}

type CreateParkPayload struct {
	UnitID string `json:"unit_id"`
	TileID string `json:"tile_id"`
}

type ExpelPayload struct {
	UnitID       string `json:"unit_id"`
	TargetUnitID string `json:"target_unit_id"`
}

func isSpecialCommand(kind string) bool {
	switch kind {
	case CommandEstablishEmbassy, CommandSoothsaying, CommandCauseUnhappiness, CommandInciteRevolution, CommandIndulgenceSales, CommandSlaveUprising, CommandStealTechnology, CommandConvertCity, CommandBioTerror, CommandNanoTerror, CommandPlague, CommandSlaveRaid, CommandSettlerSlaveRaid, CommandCreateFranchise, CommandInvestigateCity, CommandPlantNuke, CommandAssassinateRuler, CommandInjoin, CommandReform, CommandUndergroundRailway, CommandSue, CommandCreatePark, CommandExpel, CommandPillage, CommandTerraform, CommandBuildImprovement, CommandCancelTerraform:
		return true
	default:
		return false
	}
}

func applySpecialCommand(state *State, command Command, sources Sources, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	if err := activeActor(*state, command.ActorPlayerID); err != nil {
		return "", nil, err
	}
	switch command.Type {
	case CommandEstablishEmbassy, CommandSoothsaying, CommandCauseUnhappiness, CommandInciteRevolution, CommandIndulgenceSales, CommandSlaveUprising, CommandStealTechnology, CommandConvertCity, CommandBioTerror, CommandNanoTerror, CommandPlague, CommandSlaveRaid, CommandSettlerSlaveRaid, CommandCreateFranchise, CommandInvestigateCity, CommandPlantNuke, CommandAssassinateRuler, CommandInjoin, CommandReform, CommandUndergroundRailway, CommandSue:
		var payload SpecialActionPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", nil, err
		}
		if command.Type == CommandSue && payload.TargetUnitID != "" {
			return sueUnits(state, command.ActorPlayerID, payload)
		}
		unit, city, actorNation, targetNation, err := specialTarget(state, command.ActorPlayerID, payload, command.Type == CommandSue)
		if err != nil {
			return "", nil, err
		}
		if eventType, payload, stopped := stopSpyByDefenders(state, unit, city, command.Type, sources); stopped {
			return eventType, payload, nil
		}
		switch command.Type {
		case CommandEstablishEmbassy:
			if !hasEffect(unit.Effects, EffectSpecialEmbassy) || HasActiveWonderEffect(*state, city.OwnerID, EffectCloseEmbassies) || HasEffectiveEmbassyWith(*state, command.ActorPlayerID, city.OwnerID) {
				return "", nil, errors.New("special action is not authorized")
			}
			if err := spendSpecialMove(unit); err != nil {
				return "", nil, err
			}
			actorNation.Embassies = append(actorNation.Embassies, targetNation.ID)
			relation := ensureRelation(state, command.ActorPlayerID, city.OwnerID)
			relation.Contact = true
			if relation.EstablishedAt == 0 {
				relation.EstablishedAt = state.Phase.Turn
			}
			return EventEmbassyEstablished, json.RawMessage(`{"disclosed":true}`), nil
		case CommandSoothsaying:
			if !hasEffect(unit.Effects, EffectSpecialSoothsaying) {
				return "", nil, errors.New("special action is not authorized")
			}
			if err := spendSpecialMove(unit); err != nil {
				return "", nil, err
			}
			city.Effects = append(city.Effects, Effect{ID: "soothsaying:" + unit.ID, Kind: EffectSoothsaying})
			return EventSoothsayingApplied, json.RawMessage(`{"disclosed":true}`), nil
		case CommandCauseUnhappiness:
			return causeUnhappiness(state, unit, city, sources)
		case CommandInciteRevolution:
			return inciteRevolution(state, unit, city, sources)
		case CommandIndulgenceSales:
			return indulgenceSales(state, unit, city, actorNation, targetNation, runtime)
		case CommandSlaveUprising:
			return slaveUprising(unit, city)
		case CommandStealTechnology:
			return stealTechnology(state, unit, city, actorNation, targetNation, sources)
		case CommandConvertCity:
			return convertCity(state, unit, city, sources)
		case CommandCreateFranchise:
			return createFranchise(state, unit, city, sources)
		case CommandInvestigateCity:
			return investigateCity(state, unit, city, actorNation, state.Phase.Turn, sources)
		case CommandPlantNuke:
			return plantNuke(state, unit, city, actorNation, targetNation, sources, runtime)
		case CommandAssassinateRuler:
			return assassinateRuler(state, unit, city, targetNation, sources)
		case CommandInjoin:
			return injoinCity(unit, city)
		case CommandReform:
			return reformCity(unit, city)
		case CommandUndergroundRailway:
			return undergroundRailway(state, unit, city, actorNation, sources)
		case CommandSue:
			return sueCity(unit, city, command.ActorPlayerID)
		case CommandBioTerror, CommandNanoTerror, CommandPlague:
			return cityTerror(state, unit, city, command.Type, sources)
		default:
			return slaveRaid(state, unit, city, command.Type, sources)
		}
	case CommandPillage:
		var payload PillagePayload
		if err := decodePayload(command.Payload, &payload); err != nil || payload.UnitID == "" {
			return "", nil, errors.New("pillage payload is invalid")
		}
		return pillage(state, command.ActorPlayerID, payload)
	case CommandCreatePark:
		var payload CreateParkPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", nil, err
		}
		return createPark(state, command.ActorPlayerID, payload)
	case CommandExpel:
		var payload ExpelPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", nil, err
		}
		return expelUnit(state, command.ActorPlayerID, payload)
	case CommandTerraform:
		var payload TerraformPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", nil, err
		}
		unit, err := ownedUnit(state, command.ActorPlayerID, payload.UnitID)
		if err != nil {
			return "", nil, errors.New("special action is not authorized")
		}
		tile, ok := tileByID(*state, payload.TileID)
		unitTile, unitOK := tileByID(*state, unit.TileID)
		if !ok || !unitOK || !adjacent(tile, unitTile) || !hasEffect(unit.Effects, EffectSpecialTerraform) || !terraformLegal(tile.Terrain, payload.TargetTerrain) || overlayForTile(*state, tile.ID) != nil {
			return "", nil, errors.New("terrain action is not eligible")
		}
		actorNation := nationForID(state, nationForPlayer(*state, command.ActorPlayerID))
		cost, turns := effectValue(unit.Effects, EffectSpecialTerraform), effectValue(unit.Effects, EffectSpecialTerraformTurns)
		if actorNation == nil || cost <= 0 || turns <= 0 || actorNation.PublicWorks < cost {
			return "", nil, errors.New("terrain action is not eligible")
		}
		if err := spendSpecialMove(unit); err != nil {
			return "", nil, err
		}
		actorNation.PublicWorks -= cost
		state.Overlays = append(state.Overlays, TileOverlay{ID: "terraform:" + tile.ID, TileID: tile.ID, Kind: "terraform", OwnerID: command.ActorPlayerID, Remaining: turns, TargetTerrain: payload.TargetTerrain})
		return EventTerraformQueued, json.RawMessage(`{"disclosed":false}`), nil
	case CommandBuildImprovement:
		var payload BuildImprovementPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", nil, err
		}
		return buildTerrainImprovement(state, command.ActorPlayerID, payload, runtime)
	case CommandCancelTerraform:
		var payload CancelTerraformPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", nil, err
		}
		overlay := overlayForTile(*state, payload.TileID)
		if overlay == nil || overlay.OwnerID != command.ActorPlayerID || overlay.Remaining == 0 {
			return "", nil, errors.New("terrain action is not eligible")
		}
		state.Overlays = filterOverlays(state.Overlays, overlay.ID)
		return EventTerraformCancelled, json.RawMessage(`{"disclosed":false}`), nil
	}
	return "", nil, fmt.Errorf("unsupported special command %q", command.Type)
}

func stopSpyByDefenders(state *State, unit *Unit, city *City, commandType string, sources Sources) (string, json.RawMessage, bool) {
	if !spyDefenseAppliesToCommand(unit, commandType) {
		return "", nil, false
	}
	for _, defender := range state.Units {
		if defender.OwnerID != city.OwnerID || defender.TileID != city.TileID || defender.TransportID != "" || defender.ID == unit.ID {
			continue
		}
		chance := effectValue(defender.Effects, EffectDefendAgainstSpies)
		if chance <= 0 {
			continue
		}
		if chance > 100 {
			chance = 100
		}
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) < chance {
			removeUnitAndCargo(state, unit.ID)
			return spyDefenseFailureEvent(commandType), json.RawMessage(`{"disclosed":true,"stopped_by_spies":true}`), true
		}
	}
	return "", nil, false
}

func spyDefenseAppliesToCommand(unit *Unit, commandType string) bool {
	switch commandType {
	case CommandCauseUnhappiness:
		return hasEffect(unit.Effects, EffectSpecialCauseUnhappiness)
	case CommandInciteRevolution:
		return hasEffect(unit.Effects, EffectSpecialInciteRevolution)
	case CommandIndulgenceSales:
		return hasEffect(unit.Effects, EffectSpecialIndulgenceSales)
	case CommandSlaveUprising:
		return hasEffect(unit.Effects, EffectSpecialSlaveUprising)
	case CommandStealTechnology:
		return hasEffect(unit.Effects, EffectSpecialSteal)
	case CommandConvertCity:
		return hasEffect(unit.Effects, EffectSpecialConvertCity)
	case CommandInvestigateCity:
		return hasEffect(unit.Effects, EffectSpecialInvestigateCity)
	case CommandPlantNuke:
		return hasEffect(unit.Effects, EffectSpecialPlantNuke)
	case CommandAssassinateRuler:
		return hasEffect(unit.Effects, EffectSpecialAssassinateRuler)
	default:
		return false
	}
}

func spyDefenseFailureEvent(commandType string) string {
	switch commandType {
	case CommandStealTechnology:
		return EventTechnologyFailed
	case CommandConvertCity:
		return EventCityConversionFailed
	case CommandPlantNuke:
		return EventNukeFailed
	case CommandAssassinateRuler:
		return EventRulerAssassinationFailed
	default:
		return EventCityInvestigationFailed
	}
}

func buildTerrainImprovement(state *State, actorID string, payload BuildImprovementPayload, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	if runtime == nil {
		return "", nil, errors.New("terrain action is not eligible")
	}
	definition, ok := runtime.TerrainImprovement(payload.ImprovementID)
	if !ok {
		return "", nil, errors.New("terrain action is not eligible")
	}
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || !hasEffect(unit.Effects, EffectSpecialTerraform) {
		return "", nil, errors.New("terrain action is not eligible")
	}
	tile, ok := tileByID(*state, payload.TileID)
	unitTile, unitOK := tileByID(*state, unit.TileID)
	if !ok || !unitOK || !adjacent(tile, unitTile) || pendingOverlayForTile(*state, tile.ID) != nil || !terrainImprovementAllowed(*state, runtime, definition, tile.ID) {
		return "", nil, errors.New("terrain action is not eligible")
	}
	effective := effectiveTerrainImprovement(*state, runtime, definition, tile.ID)
	actorNation := nationForID(state, nationForPlayer(*state, actorID))
	if actorNation == nil || effective.PublicWorksCost < 1 || effective.BuildTurns < 1 || actorNation.PublicWorks < effective.PublicWorksCost {
		return "", nil, errors.New("terrain action is not eligible")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	actorNation.PublicWorks -= effective.PublicWorksCost
	state.Overlays = append(state.Overlays, TileOverlay{ID: "improvement:" + tile.ID + ":" + definition.ID, TileID: tile.ID, Kind: "improvement", OwnerID: actorID, Remaining: effective.BuildTurns, TargetTerrain: effective.TargetTerrain, ImprovementID: definition.ID, MovementCostDelta: effective.MovementCostDelta, YieldFood: effective.YieldFood, YieldProduction: effective.YieldProduction, YieldGold: effective.YieldGold, PollutionDelta: effective.PollutionDelta})
	return EventImprovementQueued, json.RawMessage(`{"disclosed":false}`), nil
}

func effectiveTerrainImprovement(state State, runtime *rules.RuntimeRules, definition rules.RuntimeTerrainImprovementDefinition, tileID string) rules.RuntimeTerrainImprovementDefinition {
	if runtime == nil || definition.TargetTerrain == "" {
		return definition
	}
	from, fromOK := runtime.Movement.Terrain(EffectiveTerrain(state, tileID))
	to, toOK := runtime.Movement.Terrain(definition.TargetTerrain)
	if !fromOK || !toOK || from.TransformRemoveTime < 1 || from.TransformRemoveMaterials < 1 || to.TransformAddTime < 1 || to.TransformAddMaterials < 1 {
		return definition
	}
	effective := definition
	effective.BuildTurns = from.TransformRemoveTime + to.TransformAddTime
	effective.PublicWorksCost = from.TransformRemoveMaterials + to.TransformAddMaterials
	return effective
}

func terrainImprovementAllowed(state State, runtime *rules.RuntimeRules, definition rules.RuntimeTerrainImprovementDefinition, tileID string) bool {
	terrain := EffectiveTerrain(state, tileID)
	if containsString(definition.CantBuildOn, terrain) {
		return false
	}
	for _, overlay := range state.Overlays {
		if overlay.TileID != tileID || overlay.Kind != "improvement" || overlay.Remaining != 0 {
			continue
		}
		if overlay.ImprovementID == definition.ID || containsString(definition.Excludes, overlay.ImprovementID) {
			return false
		}
		existing, ok := runtime.TerrainImprovement(overlay.ImprovementID)
		if !ok {
			continue
		}
		if (existing.Class != "" && containsString(definition.Excludes, existing.Class)) || (definition.Class != "" && containsString(existing.Excludes, definition.Class)) || containsString(existing.Excludes, definition.ID) {
			return false
		}
	}
	return true
}

func specialTarget(state *State, actorID string, payload SpecialActionPayload, allowOwnCity bool) (*Unit, *City, *Nation, *Nation, error) {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil {
		return nil, nil, nil, nil, errors.New("special action is not authorized")
	}
	var city *City
	for i := range state.Cities {
		if state.Cities[i].ID == payload.CityID {
			city = &state.Cities[i]
			break
		}
	}
	if city == nil || (!allowOwnCity && city.OwnerID == actorID) {
		return nil, nil, nil, nil, errors.New("special action target is unavailable")
	}
	unitTile, ok := tileByID(*state, unit.TileID)
	cityTile, cityOK := tileByID(*state, city.TileID)
	if !ok || !cityOK || !adjacent(unitTile, cityTile) {
		return nil, nil, nil, nil, errors.New("special action target is unavailable")
	}
	actorNation := nationForID(state, nationForPlayer(*state, actorID))
	targetNation := nationForID(state, nationForPlayer(*state, city.OwnerID))
	if actorNation == nil || targetNation == nil {
		return nil, nil, nil, nil, errors.New("special action target is unavailable")
	}
	return unit, city, actorNation, targetNation, nil
}

func sueCity(unit *Unit, city *City, actorID string) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialSue) {
		return "", nil, errors.New("special action is not authorized")
	}
	if city.OwnerID == actorID {
		if !hasExternalFranchise(*city, actorID) {
			return "", nil, errors.New("special action is not authorized")
		}
		if err := spendSpecialMove(unit); err != nil {
			return "", nil, err
		}
		city.Effects = removeExternalFranchises(city.Effects, actorID)
		return EventCitySued, json.RawMessage(`{"disclosed":true}`), nil
	}
	if hasEffect(city.Effects, EffectLawsuit) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	city.StoredProduction = 0
	city.Effects = append(city.Effects, Effect{ID: "lawsuit", Kind: EffectLawsuit, Value: 1})
	return EventCitySued, json.RawMessage(`{"disclosed":true}`), nil
}

func sueUnits(state *State, actorID string, payload SpecialActionPayload) (string, json.RawMessage, error) {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TileID == "" || !hasEffect(unit.Effects, EffectSpecialSue) {
		return "", nil, errors.New("sue target is unavailable")
	}
	target := unitPointer(state, payload.TargetUnitID)
	unitTile, unitOK := tileByID(*state, unit.TileID)
	if target == nil || target.OwnerID == actorID || target.TileID == "" || target.TransportID != "" || !hasEffect(target.Effects, EffectCanBeSued) || !unitOK {
		return "", nil, errors.New("sue target is unavailable")
	}
	targetTile, ok := tileByID(*state, target.TileID)
	if !ok || (target.TileID != unit.TileID && !adjacent(unitTile, targetTile)) {
		return "", nil, errors.New("sue target is unavailable")
	}
	victimIDs := make([]string, 0)
	for _, candidate := range state.Units {
		if candidate.OwnerID != actorID && candidate.TileID == target.TileID && candidate.TransportID == "" && hasEffect(candidate.Effects, EffectCanBeSued) {
			victimIDs = append(victimIDs, candidate.ID)
		}
	}
	if len(victimIDs) == 0 {
		return "", nil, errors.New("sue target is unavailable")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	for _, victimID := range victimIDs {
		removeUnitAndCargo(state, victimID)
	}
	return EventUnitsSued, json.RawMessage(`{"disclosed":true}`), nil
}

func hasExternalFranchise(city City, ownerID string) bool {
	for _, effect := range city.Effects {
		if effect.Kind == EffectFranchise && effect.ID != "franchise:"+ownerID {
			return true
		}
	}
	return false
}

func removeExternalFranchises(effects []Effect, ownerID string) []Effect {
	kept := effects[:0]
	for _, effect := range effects {
		if effect.Kind == EffectFranchise && effect.ID != "franchise:"+ownerID {
			continue
		}
		kept = append(kept, effect)
	}
	return kept
}

func undergroundRailway(state *State, unit *Unit, city *City, actorNation *Nation, sources Sources) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialUndergroundRailway) || city.Population <= 1 {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialUndergroundRailway)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialUndergroundRailway, sources)
			return EventUndergroundRailwayFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	city.Population--
	upsertNationEffect(actorNation, EffectFreedPopulation, 1)
	maybeResolveSpecialDeath(state, unit.ID, EffectSpecialUndergroundRailway, sources)
	return EventUndergroundRailway, json.RawMessage(`{"disclosed":true}`), nil
}

func upsertNationEffect(nation *Nation, kind string, delta int) {
	for index := range nation.Effects {
		if nation.Effects[index].Kind == kind {
			nation.Effects[index].Value += delta
			return
		}
	}
	nation.Effects = append(nation.Effects, Effect{ID: kind, Kind: kind, Value: delta})
}

func reformCity(unit *Unit, city *City) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialReform) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	city.Effects = filterEffectsByKind(city.Effects, EffectInjunction)
	city.CrimePercent = 0
	city.Happiness++
	return EventCityReformed, json.RawMessage(`{"disclosed":true}`), nil
}

func filterEffectsByKind(effects []Effect, kind string) []Effect {
	kept := effects[:0]
	for _, effect := range effects {
		if effect.Kind != kind {
			kept = append(kept, effect)
		}
	}
	return kept
}

func injoinCity(unit *Unit, city *City) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialInjoin) || hasEffect(city.Effects, EffectInjunction) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	city.Effects = append(city.Effects, Effect{ID: "injunction", Kind: EffectInjunction, Value: 1})
	return EventCityInjoined, json.RawMessage(`{"disclosed":true}`), nil
}

func expelUnit(state *State, actorID string, payload ExpelPayload) (string, json.RawMessage, error) {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TileID == "" || !hasEffect(unit.Effects, EffectSpecialExpel) {
		return "", nil, errors.New("expel target is unavailable")
	}
	target := unitPointer(state, payload.TargetUnitID)
	unitTile, unitOK := tileByID(*state, unit.TileID)
	if target == nil || target.OwnerID == actorID || target.TileID == "" || target.TransportID != "" || !unitOK || !hasEffect(target.Effects, EffectCanBeExpelled) {
		return "", nil, errors.New("expel target is unavailable")
	}
	targetTile, ok := tileByID(*state, target.TileID)
	if !ok || !adjacent(unitTile, targetTile) {
		return "", nil, errors.New("expel target is unavailable")
	}
	destination := expelDestination(*state, *target, targetTile)
	if destination == "" {
		return "", nil, errors.New("expel target is unavailable")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	target.TileID = destination
	target.Fortified = false
	return EventUnitExpelled, json.RawMessage(`{"disclosed":true}`), nil
}

func unitPointer(state *State, unitID string) *Unit {
	for index := range state.Units {
		if state.Units[index].ID == unitID {
			return &state.Units[index]
		}
	}
	return nil
}

func expelDestination(state State, target Unit, origin Tile) string {
	candidates := make([]Tile, 0)
	for _, tile := range state.Tiles {
		if tile.ID == origin.ID || !tile.Open || !adjacent(tile, origin) || tileOccupied(state, tile.ID) || cityOnTileState(state, tile.ID) {
			continue
		}
		candidates = append(candidates, tile)
	}
	sort.Slice(candidates, func(i, j int) bool {
		leftOwned, rightOwned := candidates[i].OwnerID == target.OwnerID, candidates[j].OwnerID == target.OwnerID
		if leftOwned != rightOwned {
			return leftOwned
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0].ID
}

func tileOccupied(state State, tileID string) bool {
	for _, unit := range state.Units {
		if unit.TileID == tileID && unit.TransportID == "" {
			return true
		}
	}
	return false
}

func cityOnTileState(state State, tileID string) bool {
	for _, city := range state.Cities {
		if city.TileID == tileID {
			return true
		}
	}
	return false
}

func assassinateRuler(state *State, unit *Unit, city *City, targetNation *Nation, sources Sources) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialAssassinateRuler) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialAssassinateRuler)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialAssassinateRuler, sources)
			return EventRulerAssassinationFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	targetNation.GovernmentID = ""
	targetNation.ResearchProgress = 0
	if !hasEffect(targetNation.Effects, EffectRulerAssassinated) {
		targetNation.Effects = append(targetNation.Effects, Effect{ID: "assassinated-ruler", Kind: EffectRulerAssassinated, Value: 1})
	}
	maybeResolveSpecialDeath(state, unit.ID, EffectSpecialAssassinateRuler, sources)
	return EventRulerAssassinated, json.RawMessage(`{"disclosed":true}`), nil
}

func createPark(state *State, actorID string, payload CreateParkPayload) (string, json.RawMessage, error) {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TileID == "" || unit.TransportID != "" || !hasEffect(unit.Effects, EffectSpecialCreatePark) {
		return "", nil, errors.New("park target is unavailable")
	}
	tile := tilePointer(state, payload.TileID)
	unitTile, ok := tileByID(*state, unit.TileID)
	if tile == nil || !ok || (!adjacent(*tile, unitTile) && tile.ID != unitTile.ID) || parkForTile(*state, tile.ID) != nil {
		return "", nil, errors.New("park target is unavailable")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	tile.Pollution = 0
	state.Overlays = append(state.Overlays, TileOverlay{ID: "park:" + actorID + ":" + tile.ID, TileID: tile.ID, Kind: "park", OwnerID: actorID})
	return EventParkCreated, json.RawMessage(`{"disclosed":false}`), nil
}

func parkForTile(state State, tileID string) *TileOverlay {
	for i := range state.Overlays {
		if state.Overlays[i].TileID == tileID && state.Overlays[i].Kind == "park" {
			return &state.Overlays[i]
		}
	}
	return nil
}

func nationCanPlantNuke(nation Nation, runtime *rules.RuntimeRules) bool {
	if runtime == nil {
		return true
	}
	nukeAdvanceID := ""
	for _, advance := range runtime.Advances {
		if advance.AllowsPlantNuke {
			nukeAdvanceID = advance.ID
			break
		}
	}
	return nukeAdvanceID == "" || containsString(nation.Advances, nukeAdvanceID)
}

func plantNuke(state *State, unit *Unit, city *City, actorNation *Nation, targetNation *Nation, sources Sources, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialPlantNuke) || !nationCanPlantNuke(*actorNation, runtime) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	cityEffects := cityEffects(*state, *city)
	if hasEffect(cityEffects, EffectProtectFromNukes) || hasEffect(cityEffects, EffectForceField) || hasEffect(targetNation.Effects, EffectEliminateNukes) {
		return EventNukeFailed, json.RawMessage(`{"disclosed":true}`), nil
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialPlantNuke)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialPlantNuke, sources)
			return EventNukeFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	if city.Population > 1 {
		city.Population /= 2
		if city.Population < 1 {
			city.Population = 1
		}
	}
	if tile := tilePointer(state, city.TileID); tile != nil {
		tile.Pollution += 10
	}
	destroyedUnits := resolveNukeBlastUnits(state, unit.ID, city.TileID)
	resolvePlantNukeEscape(state, unit, sources)
	maybeResolveSpecialDeath(state, unit.ID, EffectSpecialPlantNuke, sources)
	payload, _ := json.Marshal(struct {
		Disclosed      bool `json:"disclosed"`
		DestroyedUnits int  `json:"destroyed_units,omitempty"`
	}{Disclosed: true, DestroyedUnits: destroyedUnits})
	return EventNukePlanted, payload, nil
}

func resolveNukeBlastUnits(state *State, actingUnitID, cityTileID string) int {
	origin, ok := tileByID(*state, cityTileID)
	if !ok {
		return 0
	}
	removed := make([]string, 0)
	for _, unit := range state.Units {
		if unit.ID == actingUnitID || unit.TileID == "" || unit.TransportID != "" {
			continue
		}
		tile, ok := tileByID(*state, unit.TileID)
		if !ok || tileDistance(origin, tile) > 1 {
			continue
		}
		if hasEffect(unit.Effects, EffectDeathEffectsHappy) {
			recordUnitLossForPeaceMovement(state, unit)
		}
		removed = append(removed, unit.ID)
	}
	for _, unitID := range removed {
		removeUnitAndCargo(state, unitID)
	}
	return len(removed)
}

func resolvePlantNukeEscape(state *State, unit *Unit, sources Sources) {
	escapeChance := effectValue(unit.Effects, EffectSpecialPlantNukeEscapeChance)
	if escapeChance <= 0 {
		return
	}
	if escapeChance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= escapeChance {
			return
		}
	}
	tileID := nearestFriendlyCityTile(*state, unit.OwnerID, unit.TileID)
	if tileID == "" {
		return
	}
	if unit.ArmyID != "" {
		moveArmyToTile(state, unit.ArmyID, tileID)
		return
	}
	unit.TileID = tileID
}

func nearestFriendlyCityTile(state State, ownerID, fromTileID string) string {
	from, ok := tileByID(state, fromTileID)
	if !ok {
		return ""
	}
	bestTileID, bestDistance, bestCityID := "", 0, ""
	for _, city := range state.Cities {
		if city.OwnerID != ownerID || city.TileID == "" {
			continue
		}
		tile, ok := tileByID(state, city.TileID)
		if !ok {
			continue
		}
		distance := tileDistance(from, tile)
		if bestTileID == "" || distance < bestDistance || distance == bestDistance && city.ID < bestCityID {
			bestTileID, bestDistance, bestCityID = city.TileID, distance, city.ID
		}
	}
	return bestTileID
}

func tilePointer(state *State, tileID string) *Tile {
	for index := range state.Tiles {
		if state.Tiles[index].ID == tileID {
			return &state.Tiles[index]
		}
	}
	return nil
}

func investigateCity(state *State, unit *Unit, city *City, actorNation *Nation, turn int, sources Sources) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialInvestigateCity) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialInvestigateCity)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			setCityWatchful(city)
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialInvestigateCity, sources)
			return EventCityInvestigationFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	report := CityReport{CityID: city.ID, OwnerID: city.OwnerID, TileID: city.TileID, Turn: turn, Population: city.Population, Effects: append([]Effect(nil), city.Effects...)}
	for _, building := range city.Buildings {
		report.Buildings = append(report.Buildings, building.ID)
	}
	for index := range actorNation.CityReports {
		if actorNation.CityReports[index].CityID == report.CityID {
			actorNation.CityReports[index] = report
			setCityWatchful(city)
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialInvestigateCity, sources)
			return EventCityInvestigated, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	actorNation.CityReports = append(actorNation.CityReports, report)
	setCityWatchful(city)
	maybeResolveSpecialDeath(state, unit.ID, EffectSpecialInvestigateCity, sources)
	return EventCityInvestigated, json.RawMessage(`{"disclosed":true}`), nil
}

func createFranchise(state *State, unit *Unit, city *City, sources Sources) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialCreateFranchise) || hasCityFranchise(*city, unit.OwnerID) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialCreateFranchise)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialCreateFranchise, sources)
			return EventFranchiseFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	city.Effects = append(city.Effects, Effect{ID: "franchise:" + unit.OwnerID, Kind: EffectFranchise})
	maybeResolveSpecialDeath(state, unit.ID, EffectSpecialCreateFranchise, sources)
	return EventFranchiseCreated, json.RawMessage(`{"disclosed":true}`), nil
}

func hasCityFranchise(city City, ownerID string) bool {
	for _, effect := range city.Effects {
		if effect.Kind == EffectFranchise && effect.ID == "franchise:"+ownerID {
			return true
		}
	}
	return false
}

func citySpecialActionChance(city City, unitEffects []Effect, actionKind string) int {
	return specialActionChance(city, unitEffects, actionKind, false)
}

func unitCitySpecialActionChance(city City, unit Unit, actionKind string) int {
	return specialActionChance(city, unit.Effects, actionKind, unit.Veteran)
}

func specialActionChance(city City, unitEffects []Effect, actionKind string, veteran bool) int {
	chance := effectValue(unitEffects, actionKind)
	if veteran {
		if eliteChance := effectValue(unitEffects, actionKind+"_elite_chance"); eliteChance > 0 {
			chance = eliteChance
		}
	}
	if chance <= 0 {
		chance = 100
	}
	if effectValue(city.Effects, EffectWatchful) > 0 {
		chance = chance * 50 / 100
	}
	return chance
}

func setCityWatchful(city *City) {
	upsertCityEffect(city, Effect{ID: "watchful", Kind: EffectWatchful, Value: 7})
}

func slaveRaid(state *State, unit *Unit, city *City, commandType string, sources Sources) (string, json.RawMessage, error) {
	requiredEffect := EffectSpecialSlaveRaid
	if commandType == CommandSettlerSlaveRaid {
		requiredEffect = EffectSpecialSettlerSlaveRaid
	} else if commandType != CommandSlaveRaid {
		return "", nil, errors.New("special action is not authorized")
	}
	if !hasEffect(unit.Effects, requiredEffect) || HasActiveWonderEffect(*state, unit.OwnerID, EffectProhibitSlavers) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	if protectedByCityEffect(state, city, EffectPreventSlavery, sources) {
		return EventSlaveRaidFailed, json.RawMessage(`{"disclosed":true}`), nil
	}
	chance := unitCitySpecialActionChance(*city, *unit, requiredEffect)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, requiredEffect, sources)
			return EventSlaveRaidFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	if city.Population > 1 {
		city.Population--
	}
	maybeResolveSpecialDeath(state, unit.ID, requiredEffect, sources)
	return EventSlaveRaidApplied, json.RawMessage(`{"disclosed":true}`), nil
}

func cityTerror(state *State, unit *Unit, city *City, commandType string, sources Sources) (string, json.RawMessage, error) {
	requiredEffect, protectionEffect, successEvent := "", "", ""
	switch commandType {
	case CommandBioTerror:
		requiredEffect, protectionEffect, successEvent = EffectSpecialBioTerror, EffectProtectFromBioAgents, EventBioTerrorApplied
	case CommandNanoTerror:
		requiredEffect, protectionEffect, successEvent = EffectSpecialNanoTerror, EffectProtectFromNanoVirus, EventNanoTerrorApplied
	case CommandPlague:
		requiredEffect, successEvent = EffectSpecialPlague, EventPlagueApplied
	default:
		return "", nil, errors.New("special action is not authorized")
	}
	if !hasEffect(unit.Effects, requiredEffect) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	if protectionEffect != "" && protectedByCityEffect(state, city, protectionEffect, sources) {
		return EventTerrorFailed, json.RawMessage(`{"disclosed":true}`), nil
	}
	chance := unitCitySpecialActionChance(*city, *unit, requiredEffect)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, requiredEffect, sources)
			return EventTerrorFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	if city.Population > 1 {
		city.Population--
	}
	maybeResolveSpecialDeath(state, unit.ID, requiredEffect, sources)
	return successEvent, json.RawMessage(`{"disclosed":true}`), nil
}

func protectedByCityEffect(state *State, city *City, effectKind string, sources Sources) bool {
	protection := effectValue(cityEffects(*state, *city), effectKind)
	if protection <= 0 {
		return false
	}
	if protection >= 100 {
		next, _ := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		return true
	}
	next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
	setRNG(state, RNGDomainSpecialAction, next)
	return int(roll%100) < protection
}

func convertCity(state *State, unit *Unit, city *City, sources Sources) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialConvertCity) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	if HasActiveWonderEffect(*state, city.OwnerID, EffectPreventConversionEmpire) {
		return EventCityConversionFailed, json.RawMessage(`{"disclosed":true}`), nil
	}
	protection := effectValue(cityEffects(*state, *city), EffectPreventConversion)
	if protection > 0 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) < protection {
			return EventCityConversionFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialConvertCity)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialConvertCity, sources)
			return EventCityConversionFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	city.OwnerID = unit.OwnerID
	maybeResolveSpecialDeath(state, unit.ID, EffectSpecialConvertCity, sources)
	return EventCityConverted, json.RawMessage(`{"disclosed":true}`), nil
}

func indulgenceSaleAmounts(runtime *rules.RuntimeRules) (int, int) {
	gold := retailUnconvertedIndulgenceGold
	happiness := retailUnconvertedIndulgenceHappiness
	if runtime != nil {
		if runtime.Constants.UnconvertedIndulgenceGold > 0 {
			gold = runtime.Constants.UnconvertedIndulgenceGold
		}
		if runtime.Constants.UnconvertedIndulgenceHappiness > 0 {
			happiness = runtime.Constants.UnconvertedIndulgenceHappiness
		}
	}
	return gold, happiness
}

func slaveUprising(unit *Unit, city *City) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialSlaveUprising) {
		return "", nil, errors.New("special action is not authorized")
	}
	if city.Specialists["POP_SLAVE"] <= 0 {
		return "", nil, errors.New("special action target is unavailable")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	city.Specialists["POP_SLAVE"]--
	if city.Specialists["POP_SLAVE"] <= 0 {
		delete(city.Specialists, "POP_SLAVE")
	}
	city.Rioting = true
	return EventSlaveUprising, json.RawMessage(`{"disclosed":true}`), nil
}

func indulgenceSales(state *State, unit *Unit, city *City, actorNation *Nation, targetNation *Nation, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialIndulgenceSales) {
		return "", nil, errors.New("special action is not authorized")
	}
	gold, happiness := indulgenceSaleAmounts(runtime)
	if actorNation == nil || targetNation == nil || targetNation.Gold < gold {
		return "", nil, errors.New("special action target is unavailable")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	targetNation.Gold -= gold
	actorNation.Gold += gold
	city.Effects = append(city.Effects, Effect{ID: "indulgence:" + unit.ID, Kind: EffectHappiness, Value: happiness, Turns: 1})
	return EventIndulgenceSold, json.RawMessage(`{"disclosed":true}`), nil
}

func inciteRevolution(state *State, unit *Unit, city *City, sources Sources) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialInciteRevolution) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialInciteRevolution)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			maybeResolveSpecialDeath(state, unit.ID, EffectSpecialInciteRevolution, sources)
			return EventRevolutionFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	city.OwnerID = unit.OwnerID
	maybeResolveSpecialDeath(state, unit.ID, EffectSpecialInciteRevolution, sources)
	return EventRevolutionIncited, json.RawMessage(`{"disclosed":true}`), nil
}

func causeUnhappiness(state *State, unit *Unit, city *City, sources Sources) (string, json.RawMessage, error) {
	if !hasEffect(unit.Effects, EffectSpecialCauseUnhappiness) {
		return "", nil, errors.New("special action is not authorized")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	chance := unitCitySpecialActionChance(*city, *unit, EffectSpecialCauseUnhappiness)
	setCityWatchful(city)
	if chance < 100 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) >= chance {
			return EventCauseUnhappinessFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	amount := effectValue(unit.Effects, EffectSpecialCauseUnhappinessAmount)
	if amount <= 0 {
		amount = 1
	}
	turns := effectValue(unit.Effects, EffectSpecialCauseUnhappinessTimer)
	if turns <= 0 {
		turns = 1
	}
	city.Effects = append(city.Effects, Effect{ID: "happiness-attack:" + unit.ID, Kind: EffectHappiness, Value: -amount, Turns: turns})
	return EventCauseUnhappinessApplied, json.RawMessage(`{"disclosed":true}`), nil
}

func stoppedBySpiesEverywhere(state *State, unit *Unit, city *City, sources Sources) bool {
	if !HasActiveWonderEffect(*state, city.OwnerID, EffectSpiesEverywhere) {
		return false
	}
	next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
	setRNG(state, RNGDomainSpecialAction, next)
	if int(roll%100) >= 50 {
		return false
	}
	removeUnitAndCargo(state, unit.ID)
	return true
}

func stealTechnology(state *State, unit *Unit, city *City, actor, target *Nation, sources Sources) (string, json.RawMessage, error) {
	chance := effectValue(unit.Effects, EffectSpecialSteal)
	if chance < 0 || chance > 100 {
		return "", nil, errors.New("special action is not authorized")
	}
	eligible := make([]string, 0, len(target.Advances))
	for _, advance := range target.Advances {
		if !containsString(actor.Advances, advance) {
			eligible = append(eligible, advance)
		}
	}
	sort.Strings(eligible)
	if len(eligible) == 0 {
		return "", nil, errors.New("special action target is unavailable")
	}
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	if stoppedBySpiesEverywhere(state, unit, city, sources) {
		return EventTechnologyFailed, json.RawMessage(`{"disclosed":true}`), nil
	}
	next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
	setRNG(state, RNGDomainSpecialAction, next)
	if int(roll%100) >= chance {
		next, capture := sources.RNG.Next(RNGDomainSpecialAction, next)
		setRNG(state, RNGDomainSpecialAction, next)
		if int(capture%100) < effectValue(unit.Effects, EffectSpecialCapture) {
			removeUnitAndCargo(state, unit.ID)
			return EventTechnologyCapture, json.RawMessage(`{"disclosed":true}`), nil
		}
		return EventTechnologyFailed, json.RawMessage(`{"disclosed":true}`), nil
	}
	next, choice := sources.RNG.Next(RNGDomainSpecialAction, next)
	setRNG(state, RNGDomainSpecialAction, next)
	actor.Advances = append(actor.Advances, eligible[int(choice%uint64(len(eligible)))])
	return EventTechnologyStolen, json.RawMessage(`{"disclosed":true}`), nil
}

func EffectiveTerrain(state State, tileID string) string {
	tile, ok := tileByID(state, tileID)
	if !ok {
		return ""
	}
	if overlay := overlayForTile(state, tileID); overlay != nil && overlay.Kind == "terraform" && overlay.Remaining == 0 {
		return overlay.TargetTerrain
	}
	return tile.Terrain
}

func advanceTerraform(state *State) {
	for i := range state.Overlays {
		if state.Overlays[i].Remaining <= 0 {
			continue
		}
		before := state.Overlays[i].Remaining
		state.Overlays[i].Remaining--
		if state.Overlays[i].Kind == "improvement" && before == 1 {
			applyImprovementCompletion(state, &state.Overlays[i])
		}
	}
	state.Overlays = filterCompletedOneShotImprovements(state.Overlays)
}

func applyImprovementCompletion(state *State, overlay *TileOverlay) {
	if overlay.PollutionDelta == 0 && overlay.TargetTerrain == "" {
		return
	}
	for i := range state.Tiles {
		if state.Tiles[i].ID != overlay.TileID {
			continue
		}
		if overlay.TargetTerrain != "" {
			state.Tiles[i].Terrain = overlay.TargetTerrain
		}
		state.Tiles[i].Pollution += overlay.PollutionDelta
		if state.Tiles[i].Pollution < 0 {
			state.Tiles[i].Pollution = 0
		}
		return
	}
}

func filterCompletedOneShotImprovements(overlays []TileOverlay) []TileOverlay {
	kept := overlays[:0]
	for _, overlay := range overlays {
		if overlay.Kind == "improvement" && overlay.Remaining == 0 && overlay.TargetTerrain != "" {
			continue
		}
		if overlay.Kind == "improvement" && overlay.Remaining == 0 && overlay.MovementCostDelta == 0 && overlay.YieldFood == 0 && overlay.YieldProduction == 0 && overlay.YieldGold == 0 {
			continue
		}
		kept = append(kept, overlay)
	}
	return kept
}

func validateSpecialState(state State) error {
	for _, nation := range state.Nations {
		if nation.ID == "" || nation.PublicWorks < 0 || nation.Gold < 0 || uniqueStrings(nation.Advances) || uniqueStrings(nation.Embassies) || invalidCityReports(nation.CityReports) {
			return errors.New("nation special state is invalid")
		}
	}
	for _, overlay := range state.Overlays {
		switch overlay.Kind {
		case "terraform":
			if overlay.TileID == "" || overlay.OwnerID == "" || overlay.Remaining < 0 || !terraformLegal(terrainOf(state, overlay.TileID), overlay.TargetTerrain) {
				return errors.New("terrain overlay is invalid")
			}
		case "improvement":
			if overlay.TileID == "" || overlay.OwnerID == "" || overlay.Remaining < 0 || overlay.ImprovementID == "" {
				return errors.New("terrain overlay is invalid")
			}
		case "park":
			if overlay.TileID == "" || overlay.OwnerID == "" || overlay.Remaining != 0 {
				return errors.New("terrain overlay is invalid")
			}
		default:
			return errors.New("terrain overlay is invalid")
		}
	}
	return nil
}

func pillage(state *State, actorID string, payload PillagePayload) (string, json.RawMessage, error) {
	unit, err := ownedUnit(state, actorID, payload.UnitID)
	if err != nil || unit.TileID == "" || unit.TransportID != "" || !hasEffect(unit.Effects, EffectCanPillage) {
		return "", nil, errors.New("pillage target is unavailable")
	}
	tile, ok := tileByID(*state, unit.TileID)
	if !ok || tile.OwnerID == actorID {
		return "", nil, errors.New("pillage target is unavailable")
	}
	kept := state.Overlays[:0]
	removed := 0
	for _, overlay := range state.Overlays {
		if overlay.TileID == unit.TileID && overlay.Kind == "improvement" && overlay.Remaining == 0 {
			removed++
			continue
		}
		kept = append(kept, overlay)
	}
	if removed == 0 {
		return "", nil, errors.New("pillage target is unavailable")
	}
	state.Overlays = kept
	if err := spendSpecialMove(unit); err != nil {
		return "", nil, err
	}
	payloadJSON, _ := json.Marshal(struct {
		UnitID  string `json:"unit_id"`
		TileID  string `json:"tile_id"`
		Removed int    `json:"removed"`
	}{UnitID: unit.ID, TileID: unit.TileID, Removed: removed})
	return EventPillaged, payloadJSON, nil
}

func terraformLegal(from, to string) bool { return from == TerrainForest && to == TerrainPlains }
func invalidCityReports(reports []CityReport) bool {
	seen := map[string]bool{}
	for _, report := range reports {
		if report.CityID == "" || report.OwnerID == "" || report.TileID == "" || report.Turn < 0 || report.Population < 0 || uniqueStrings(report.Buildings) {
			return true
		}
		if seen[report.CityID] {
			return true
		}
		seen[report.CityID] = true
	}
	return false
}

func terrainOf(state State, tileID string) string {
	tile, _ := tileByID(state, tileID)
	return tile.Terrain
}
func nationForID(state *State, id string) *Nation {
	for i := range state.Nations {
		if state.Nations[i].ID == id {
			return &state.Nations[i]
		}
	}
	return nil
}
func effectValue(effects []Effect, kind string) int {
	for _, effect := range effects {
		if effect.Kind == kind {
			return effect.Value
		}
	}
	return 0
}
func maybeResolveSpecialDeath(state *State, unitID, actionKind string, sources Sources) {
	unit, ok := unitByID(*state, unitID)
	if !ok {
		return
	}
	chance := effectValue(unit.Effects, actionKind+"_death_chance")
	if chance <= 0 {
		return
	}
	next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
	setRNG(state, RNGDomainSpecialAction, next)
	if int(roll%100) < chance {
		removeUnitAndCargo(state, unitID)
	}
}

func spendSpecialMove(unit *Unit) error {
	if unit.Movement < 1 {
		return errors.New("special action is not authorized")
	}
	unit.Movement--
	return nil
}
func rngValue(state State, domain string) uint64 {
	entry, _ := state.RNGState(domain)
	return entry.State
}
func setRNG(state *State, domain string, value uint64) {
	for i := range state.RNG {
		if state.RNG[i].Domain == domain {
			state.RNG[i].State = value
			return
		}
	}
}
func overlayForTile(state State, tileID string) *TileOverlay {
	for i := range state.Overlays {
		if state.Overlays[i].TileID == tileID {
			return &state.Overlays[i]
		}
	}
	return nil
}

func pendingOverlayForTile(state State, tileID string) *TileOverlay {
	for i := range state.Overlays {
		if state.Overlays[i].TileID == tileID && state.Overlays[i].Remaining > 0 {
			return &state.Overlays[i]
		}
	}
	return nil
}
func filterOverlays(overlays []TileOverlay, id string) []TileOverlay {
	kept := overlays[:0]
	for _, overlay := range overlays {
		if overlay.ID != id {
			kept = append(kept, overlay)
		}
	}
	return kept
}
func uniqueStrings(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}
