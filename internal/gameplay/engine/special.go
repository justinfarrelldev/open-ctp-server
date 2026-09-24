package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	EffectSpecialEmbassy            = "special_embassy"
	EffectSpecialSoothsaying        = "special_soothsaying"
	EffectSpecialSteal              = "special_steal"
	EffectSpecialCapture            = "special_capture"
	EffectSpecialTerraform          = "special_terraform"
	EffectSpecialTerraformTurns     = "special_terraform_turns"
	EffectSpecialConvertCity        = "special_convert_city"
	EffectSpecialCreateFranchise    = "special_create_franchise"
	EffectSpecialInvestigateCity    = "special_investigate_city"
	EffectSpecialPlantNuke          = "special_plant_nuke"
	EffectSpecialBioTerror          = "special_bio_terror"
	EffectSpecialNanoTerror         = "special_nano_terror"
	EffectSpecialPlague             = "special_plague"
	EffectSpecialAssassinateRuler   = "special_assassinate_ruler"
	EffectSpecialSlaveRaid          = "special_slave_raid"
	EffectSpecialSettlerSlaveRaid   = "special_settler_slave_raid"
	EffectSpecialUndergroundRailway = "special_underground_railway"
	EffectSpecialInjoin             = "special_injoin"
	EffectSpecialReform             = "special_reform"
	EffectSpecialCreatePark         = "special_create_park"
	EffectSpecialExpel              = "special_expel"
	EffectSpecialSue                = "special_sue"
	EffectSoothsaying               = "soothsaying"
)

type SpecialActionPayload struct {
	UnitID string `json:"unit_id"`
	CityID string `json:"city_id"`
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

func isSpecialCommand(kind string) bool {
	switch kind {
	case CommandEstablishEmbassy, CommandSoothsaying, CommandStealTechnology, CommandConvertCity, CommandBioTerror, CommandNanoTerror, CommandPlague, CommandPillage, CommandTerraform, CommandBuildImprovement, CommandCancelTerraform:
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
	case CommandEstablishEmbassy, CommandSoothsaying, CommandStealTechnology, CommandConvertCity, CommandBioTerror, CommandNanoTerror, CommandPlague:
		var payload SpecialActionPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", nil, err
		}
		unit, city, actorNation, targetNation, err := specialTarget(state, command.ActorPlayerID, payload)
		if err != nil {
			return "", nil, err
		}
		switch command.Type {
		case CommandEstablishEmbassy:
			if !hasEffect(unit.Effects, EffectSpecialEmbassy) || containsString(actorNation.Embassies, targetNation.ID) {
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
		case CommandStealTechnology:
			return stealTechnology(state, unit, actorNation, targetNation, sources)
		case CommandConvertCity:
			return convertCity(state, unit, city, sources)
		default:
			return cityTerror(state, unit, city, command.Type, sources)
		}
	case CommandPillage:
		var payload PillagePayload
		if err := decodePayload(command.Payload, &payload); err != nil || payload.UnitID == "" {
			return "", nil, errors.New("pillage payload is invalid")
		}
		return pillage(state, command.ActorPlayerID, payload)
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

func specialTarget(state *State, actorID string, payload SpecialActionPayload) (*Unit, *City, *Nation, *Nation, error) {
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
	if city == nil || city.OwnerID == actorID {
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
	if city.Population > 1 {
		city.Population--
	}
	return successEvent, json.RawMessage(`{"disclosed":true}`), nil
}

func protectedByCityEffect(state *State, city *City, effectKind string, sources Sources) bool {
	protection := effectValue(city.Effects, effectKind)
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
	protection := effectValue(city.Effects, EffectPreventConversion)
	if protection > 0 {
		next, roll := sources.RNG.Next(RNGDomainSpecialAction, rngValue(*state, RNGDomainSpecialAction))
		setRNG(state, RNGDomainSpecialAction, next)
		if int(roll%100) < protection {
			return EventCityConversionFailed, json.RawMessage(`{"disclosed":true}`), nil
		}
	}
	city.OwnerID = unit.OwnerID
	return EventCityConverted, json.RawMessage(`{"disclosed":true}`), nil
}

func stealTechnology(state *State, unit *Unit, actor, target *Nation, sources Sources) (string, json.RawMessage, error) {
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
		if nation.ID == "" || nation.PublicWorks < 0 || uniqueStrings(nation.Advances) || uniqueStrings(nation.Embassies) {
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
