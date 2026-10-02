package engine

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

var supportedVictoryEvaluators = []string{
	"bloodlust_conquest",
	"diplomatic_world_union",
	"score_2300",
	"science_gaia_controller",
}

type setResearchPayload struct {
	AdvanceID string `json:"advance_id"`
}

func applyLifecycleCommand(state *State, command Command, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	switch command.Type {
	case CommandSetResearch:
		var payload setResearchPayload
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return "", nil, errors.New("research payload is invalid")
		}
		return setResearch(state, command.ActorPlayerID, payload, runtime)
	case CommandEndGame:
		if _, ok := playerByID(state, command.ActorPlayerID); !ok {
			return "", nil, errors.New("owner is not a game player")
		}
		endOwnerGame(state, command.ActorPlayerID)
		return EventTerminal, terminalPayload(state), nil
	case CommandLeaveGame:
		player, ok := playerByID(state, command.ActorPlayerID)
		if !ok || player.Lifecycle != PlayerActive {
			return "", nil, errors.New("player is not active")
		}
		if player.TakeoverPending {
			return "", nil, errors.New("player departure is already scheduled")
		}
		player.TakeoverPending = true
		return EventPlayerDisconnected, lifecyclePayload(command.ActorPlayerID, player.Controller), nil
	case CommandTakeoverAI:
		if state.Phase.ResolutionLocked {
			return "", nil, errors.New("takeover is allowed only at a policy boundary")
		}
		player, ok := playerByID(state, command.ActorPlayerID)
		if !ok || !player.TakeoverPending {
			return "", nil, errors.New("player has no scheduled takeover")
		}
		player.Controller = ControllerAI
		player.TakeoverPending = false
		return EventControllerChanged, lifecyclePayload(command.ActorPlayerID, ControllerAI), nil
	case CommandRoundComplete:
		next, err := RecalculateRound(*state)
		if err != nil {
			return "", nil, err
		}
		*state = next
		updateGaiaCountdown(state, runtime, state.Phase.Turn)
		advanceRound(state, runtime)
		if winner, reason := evaluateVictory(*state, runtime); winner != "" {
			endGame(state, winner, reason)
			return EventTerminal, terminalPayload(state), nil
		}
		return "round.completed", json.RawMessage(`{}`), nil
	default:
		return "", nil, errors.New("unsupported lifecycle command")
	}
}

func endGame(state *State, winner, reason string) {
	state.Victory = VictoryState{WinnerNationID: nationForPlayer(*state, winner), Reason: reason}
	setTerminal(state, winner, reason)
}

func setResearch(state *State, actorID string, payload setResearchPayload, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	player, ok := playerByID(state, actorID)
	if !ok {
		return "", nil, errors.New("owner is not a game player")
	}
	nation := nationForID(state, player.NationID)
	if nation == nil {
		return "", nil, errors.New("actor nation does not exist")
	}
	if runtime == nil {
		return "", nil, errors.New("research rules are unavailable")
	}
	if _, ok := runtime.Advance(payload.AdvanceID); !ok {
		return "", nil, errors.New("research advance is unsupported")
	}
	if !canResearchAdvance(*nation, *runtime, payload.AdvanceID) {
		return "", nil, errors.New("research prerequisite is not met")
	}
	nation.ResearchTargetID = payload.AdvanceID
	nation.ResearchProgress = 0
	payloadJSON, _ := json.Marshal(struct {
		AdvanceID string `json:"advance_id"`
	}{AdvanceID: payload.AdvanceID})
	return "research.selected", payloadJSON, nil
}

func advanceRound(state *State, runtime *rules.RuntimeRules) {
	state.Phase.Turn++
	if len(state.Phase.ActivePlayerOrder) > 0 {
		state.Phase.ActivePlayerIndex = (state.Phase.ActivePlayerIndex + 1) % len(state.Phase.ActivePlayerOrder)
	}
	if runtime != nil {
		advancePollution(state, runtime)
	}
	advanceTemporaryNationEffects(state)
	if runtime == nil || runtime.Starter.Movement < 1 {
		return
	}
	crashedUnits := make([]string, 0)
	for index := range state.Units {
		if state.Units[index].TransportID == "" {
			healUnitFromTerrainImprovements(state, &state.Units[index])
			state.Units[index].Movement = refreshedUnitMovement(*state, state.Units[index], runtime.Starter.Movement)
		}
		if state.Units[index].MaxFuel > 0 && unitCanRefuel(*state, state.Units[index]) {
			state.Units[index].Fuel = state.Units[index].MaxFuel
		}
		if state.Units[index].MaxFuel > 0 && state.Units[index].Fuel == 0 && hasEffect(state.Units[index].Effects, EffectNoFuelThenCrash) {
			crashedUnits = append(crashedUnits, state.Units[index].ID)
		}
	}
	for _, unitID := range crashedUnits {
		removeUnitAndCargo(state, unitID)
	}
	advanceTreasury(state, runtime)
	advanceResearch(state, runtime)
}

func advanceTemporaryNationEffects(state *State) {
	for nationIndex := range state.Nations {
		effects := state.Nations[nationIndex].Effects[:0]
		for _, effect := range state.Nations[nationIndex].Effects {
			if effect.Turns > 0 {
				effect.Turns--
				if effect.Turns == 0 {
					continue
				}
			}
			effects = append(effects, effect)
		}
		state.Nations[nationIndex].Effects = effects
	}
}

func healUnitFromTerrainImprovements(state *State, unit *Unit) {
	if unit == nil || unit.TileID == "" || unit.Health <= 0 || unit.MaxHealth <= 0 || unit.Health >= unit.MaxHealth {
		return
	}
	bestPercent := 0
	for _, overlay := range state.Overlays {
		if overlay.TileID != unit.TileID || overlay.Kind != "improvement" || overlay.Remaining > 0 || overlay.HealRatePercent <= bestPercent {
			continue
		}
		if overlay.OwnerID != "" && overlay.OwnerID != unit.OwnerID {
			continue
		}
		bestPercent = overlay.HealRatePercent
	}
	if bestPercent <= 0 {
		return
	}
	heal := unit.MaxHealth * bestPercent / 100
	if heal < 1 {
		heal = 1
	}
	unit.Health += heal
	if unit.Health > unit.MaxHealth {
		unit.Health = unit.MaxHealth
	}
}

func unitCanRefuel(state State, unit Unit) bool {
	if unit.TransportID != "" {
		return true
	}
	for _, city := range state.Cities {
		if city.TileID == unit.TileID && city.OwnerID == unit.OwnerID {
			return true
		}
	}
	for _, overlay := range state.Overlays {
		if overlay.TileID == unit.TileID && overlay.Kind == "improvement" && overlay.Remaining <= 0 && overlay.Airport && (overlay.OwnerID == "" || overlay.OwnerID == unit.OwnerID) {
			return true
		}
	}
	return false
}

func advancePollution(state *State, runtime *rules.RuntimeRules) {
	if runtime == nil || len(runtime.PollutionTriggers) == 0 {
		return
	}
	pollutionByNation := playerPollutionTotals(*state)
	if state.Pollution == nil && len(pollutionByNation) == 0 {
		return
	}
	if state.Pollution == nil {
		state.Pollution = &PollutionState{}
	}
	trigger := runtimePollutionTrigger(runtime.PollutionTriggers)
	retailState := rules.RetailPollutionState{Phase: state.Pollution.Phase, GWPhase: state.Pollution.GWPhase, History: append([]int64(nil), state.Pollution.History...), Trend: state.Pollution.Trend}
	retailTrigger := rules.RetailPollutionTriggerList{ID: trigger.ID, TriggerLevels: append([]int64(nil), trigger.TriggerLevels...), DisasterTypes: append([]string(nil), trigger.DisasterTypes...)}
	next, event := rules.RetailPollutionBeginTurn(retailState, retailTrigger, pollutionByNation, false)
	next = rules.RetailPollutionEndRound(next, pollutionByNation)
	state.Pollution.Phase = next.Phase
	state.Pollution.GWPhase = next.GWPhase
	state.Pollution.History = append([]int64(nil), next.History...)
	state.Pollution.Trend = next.Trend
	state.Pollution.LastPollutionLevel = event.PollutionLevel
	if event.Triggered {
		state.Pollution.LastDisaster = event.DisasterType
		state.Pollution.LastTriggerLevel = event.TriggerLevel
		applyPollutionDisaster(state, runtime, event.DisasterType)
		if event.DisasterType == "POLLUTION_FLOOD_DISASTER" {
			state.Pollution.GWPhase++
		}
	}
}

func applyPollutionDisaster(state *State, runtime *rules.RuntimeRules, disasterType string) {
	switch disasterType {
	case "POLLUTION_FLOOD_DISASTER":
		floodPollutionTile(state, runtime)
	case "POLLUTION_OZONE_DISASTER":
		ozonePollutionTile(state, runtime)
	}
}

func floodPollutionTile(state *State, runtime *rules.RuntimeRules) {
	index, ok := highestPollutedLandTile(*state, runtime)
	if !ok {
		return
	}
	tileID := state.Tiles[index].ID
	state.Tiles[index].Terrain = shallowWaterTerrain(runtime)
	state.Tiles[index].OwnerID = ""
	state.Tiles[index].Pollution = 0
	removeFloodedOccupants(state, tileID)
	removeOverlaysOnTile(state, tileID)
}

func ozonePollutionTile(state *State, runtime *rules.RuntimeRules) {
	index, ok := highestPollutedLandTile(*state, runtime)
	if !ok {
		return
	}
	tileID := state.Tiles[index].ID
	state.Tiles[index].Terrain = deadTerrain(runtime)
	state.Tiles[index].Pollution = 0
	removeOverlaysOnTile(state, tileID)
}

func highestPollutedLandTile(state State, runtime *rules.RuntimeRules) (int, bool) {
	bestIndex := -1
	bestPollution := 0
	for index, tile := range state.Tiles {
		if tile.ID == "" || waterTerrain(runtime, tile.Terrain) {
			continue
		}
		if bestIndex == -1 || tile.Pollution > bestPollution || tile.Pollution == bestPollution && tile.ID < state.Tiles[bestIndex].ID {
			bestIndex = index
			bestPollution = tile.Pollution
		}
	}
	return bestIndex, bestIndex >= 0
}

func waterTerrain(runtime *rules.RuntimeRules, terrainID string) bool {
	if runtime != nil {
		if terrain, ok := runtime.Movement.Terrain(terrainID); ok {
			return terrain.Allows(DomainSea) && !terrain.Allows(DomainLand)
		}
	}
	if terrainID == TerrainOcean || strings.HasPrefix(terrainID, "TERRAIN_WATER_") {
		return true
	}
	return false
}

func shallowWaterTerrain(runtime *rules.RuntimeRules) string {
	if runtime != nil {
		for _, id := range []string{"TERRAIN_WATER_SHALLOW", TerrainOcean} {
			if _, ok := runtime.Movement.Terrain(id); ok {
				return id
			}
		}
	}
	return TerrainOcean
}

func deadTerrain(runtime *rules.RuntimeRules) string {
	if runtime != nil {
		if _, ok := runtime.Movement.Terrain("TERRAIN_DEAD"); ok {
			return "TERRAIN_DEAD"
		}
	}
	return "TERRAIN_DEAD"
}

func removeFloodedOccupants(state *State, tileID string) {
	cities := state.Cities[:0]
	for _, city := range state.Cities {
		if city.TileID != tileID {
			cities = append(cities, city)
		}
	}
	state.Cities = cities

	units := state.Units[:0]
	for _, unit := range state.Units {
		if unit.TileID != tileID || unit.Domain == DomainSea || unit.Domain == DomainAmphibious || unit.Domain == DomainAir {
			units = append(units, unit)
		}
	}
	state.Units = units
	pruneArmiesToExistingUnits(state)
}

func pruneArmiesToExistingUnits(state *State) {
	exists := make(map[string]bool, len(state.Units))
	for _, unit := range state.Units {
		exists[unit.ID] = true
	}
	armies := state.Armies[:0]
	for _, army := range state.Armies {
		members := army.UnitIDs[:0]
		for _, unitID := range army.UnitIDs {
			if exists[unitID] {
				members = append(members, unitID)
			}
		}
		if len(members) == 0 {
			continue
		}
		army.UnitIDs = members
		armies = append(armies, army)
	}
	state.Armies = armies
}

func removeOverlaysOnTile(state *State, tileID string) {
	overlays := state.Overlays[:0]
	for _, overlay := range state.Overlays {
		if overlay.TileID != tileID {
			overlays = append(overlays, overlay)
		}
	}
	state.Overlays = overlays
}

func runtimePollutionTrigger(triggers []rules.RuntimePollutionTriggerDefinition) rules.RuntimePollutionTriggerDefinition {
	for _, trigger := range triggers {
		if trigger.ID == "POLLUTION_TRIGGER_LIST_SMALL_MAP" {
			return trigger
		}
	}
	return triggers[0]
}

func playerPollutionTotals(state State) []int64 {
	var total int64
	for _, tile := range state.Tiles {
		total += int64(tile.Pollution)
	}
	if total == 0 {
		return nil
	}
	return []int64{total}
}

func refreshedUnitMovement(state State, unit Unit, fallback int) int {
	movement := fallback
	if unit.MaxMovement > 0 {
		movement = unit.MaxMovement
	}
	if !hasEffect(unit.Effects, EffectLossMoveToDmgNone) && unit.Health > 0 && unit.MaxHealth > 0 && unit.Health < unit.MaxHealth {
		movement = movement * unit.Health / unit.MaxHealth
	}
	if unit.Domain == DomainSea || unit.Domain == DomainAmphibious {
		movement += boatMovementBonus(state, unit.OwnerID)
	}
	if movement < 0 {
		return 0
	}
	return movement
}

func boatMovementBonus(state State, ownerID string) int {
	bonus := ownerEffectSum(state, ownerID, EffectIncreaseBoatMovement)
	for _, wonder := range state.Wonders {
		if !wonderAppliesToPlayer(state, wonder, ownerID) {
			continue
		}
		for _, effect := range wonder.Effects {
			if effect.Kind == EffectIncreaseBoatMovement {
				bonus += effect.Value
			}
		}
	}
	return bonus
}

func advanceResearch(state *State, runtime *rules.RuntimeRules) {
	if runtime == nil || len(runtime.Advances) == 0 {
		return
	}
	for index := range state.Nations {
		nation := &state.Nations[index]
		if nation.ResearchTargetID == "" {
			continue
		}
		rate := researchRate(*state, *nation)
		if rate < 1 {
			rate = 1
		}
		nation.ResearchProgress += rate
		for nation.ResearchTargetID != "" {
			advance, ok := runtime.Advance(nation.ResearchTargetID)
			if !ok || !canResearchAdvance(*nation, *runtime, advance.ID) {
				nation.ResearchTargetID = ""
				nation.ResearchProgress = 0
				break
			}
			if nation.ResearchProgress < advance.Cost {
				break
			}
			nation.ResearchProgress -= advance.Cost
			nation.Advances = append(nation.Advances, advance.ID)
			sort.Strings(nation.Advances)
			applyAdvanceTriggerFeat(nation, advance, runtime)
			nation.ResearchTargetID = nextResearchAdvance(*nation, *runtime)
			if nation.ResearchTargetID == "" {
				nation.ResearchProgress = 0
			}
		}
	}
}

func applyAdvanceTriggerFeat(nation *Nation, advance rules.RuntimeAdvanceDefinition, runtime *rules.RuntimeRules) {
	if nation == nil || runtime == nil || advance.TriggerFeatID == "" {
		return
	}
	feat, ok := runtime.Feat(advance.TriggerFeatID)
	if !ok {
		return
	}
	applyRuntimeFeat(nation, feat)
}

func applyRuntimeFeat(nation *Nation, feat rules.RuntimeFeatDefinition) {
	if nation == nil || feat.Duration < 1 || len(feat.Effects) == 0 {
		return
	}
	for _, effect := range feat.Effects {
		if effect.Kind == "" {
			continue
		}
		effectID := "feat:" + feat.ID + ":" + effect.Kind
		refreshed := false
		for index := range nation.Effects {
			if nation.Effects[index].ID == effectID {
				nation.Effects[index].Kind = effect.Kind
				nation.Effects[index].Value = effect.Value
				nation.Effects[index].Turns = feat.Duration
				refreshed = true
				break
			}
		}
		if !refreshed {
			nation.Effects = append(nation.Effects, Effect{ID: effectID, Kind: effect.Kind, Value: effect.Value, Turns: feat.Duration})
		}
	}
	normalizeEffects(nation.Effects)
}

func advanceTreasury(state *State, runtimes ...*rules.RuntimeRules) {
	var runtime *rules.RuntimeRules
	if len(runtimes) > 0 {
		runtime = runtimes[0]
	}
	for index := range state.Nations {
		nation := &state.Nations[index]
		sciencePercent := nationSciencePercent(*nation, maxSciencePercentForNation(*nation))
		income := 0
		for _, city := range state.Cities {
			if nationForPlayer(*state, city.OwnerID) != nation.ID {
				continue
			}
			cityGold := city.Yields.Gold
			if cityGold > 0 {
				cityGold -= payConversionGold(state, city, runtime)
				if cityGold < 0 {
					cityGold = 0
				}
				income += cityGold - cityGold*sciencePercent/100
			}
			income -= cityBuildingMaintenanceDeficit(*state, city)
		}
		income -= unitUpkeepForNation(*state, nation.ID)
		nation.Gold += income
		if nation.Gold < 0 {
			sellHighestUpkeepBuildingsForShortfall(state, nation.ID, -nation.Gold)
			nation.Gold = 0
		}
	}
}

func payConversionGold(state *State, city City, runtime *rules.RuntimeRules) int {
	if state == nil || city.ConvertedToPlayerID == "" || city.ConvertedToPlayerID == city.OwnerID || city.Yields.Gold <= 0 {
		return 0
	}
	convertedNationID := nationForPlayer(*state, city.ConvertedToPlayerID)
	if convertedNationID == "" {
		return 0
	}
	percent := conversionGoldPercent(city, runtime)
	if percent <= 0 {
		return 0
	}
	amount := city.Yields.Gold * percent / 100
	if amount < 1 {
		amount = 1
	}
	if amount > city.Yields.Gold {
		amount = city.Yields.Gold
	}
	if nation := nationForID(state, convertedNationID); nation != nil {
		nation.Gold += amount
		return amount
	}
	return 0
}

func conversionGoldPercent(city City, runtime *rules.RuntimeRules) int {
	if city.ConvertedBy == "televangelist" {
		if runtime != nil && runtime.Constants.TelevangelistConversionPercent > 0 {
			return runtime.Constants.TelevangelistConversionPercent
		}
		return 40
	}
	if runtime != nil && runtime.Constants.ClericConversionPercent > 0 {
		return runtime.Constants.ClericConversionPercent
	}
	return 20
}

type buildingUpkeepCandidate struct {
	cityID     string
	buildingID string
	upkeep     int
}

func sellHighestUpkeepBuildingsForShortfall(state *State, nationID string, shortfall int) {
	if state == nil || shortfall <= 0 {
		return
	}
	candidates := make([]buildingUpkeepCandidate, 0)
	for _, city := range state.Cities {
		if nationForPlayer(*state, city.OwnerID) != nationID {
			continue
		}
		for _, building := range city.Buildings {
			if building.Upkeep <= 0 || isObsolete(city.Obsolete, building.ObsoletedBy) {
				continue
			}
			candidates = append(candidates, buildingUpkeepCandidate{cityID: city.ID, buildingID: building.ID, upkeep: building.Upkeep})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].upkeep == candidates[j].upkeep {
			if candidates[i].cityID == candidates[j].cityID {
				return candidates[i].buildingID < candidates[j].buildingID
			}
			return candidates[i].cityID < candidates[j].cityID
		}
		return candidates[i].upkeep > candidates[j].upkeep
	})
	for _, candidate := range candidates {
		if shortfall <= 0 {
			return
		}
		city, ok := cityByID(*state, candidate.cityID)
		if !ok {
			continue
		}
		if removeBuildingByID(city, candidate.buildingID) {
			shortfall -= candidate.upkeep
		}
	}
}

func removeBuildingByID(city *City, buildingID string) bool {
	for index := range city.Buildings {
		if city.Buildings[index].ID != buildingID {
			continue
		}
		city.Buildings = append(city.Buildings[:index], city.Buildings[index+1:]...)
		return true
	}
	return false
}

func maxSciencePercentForNation(nation Nation) int {
	maxSciencePercent := -1
	for _, effect := range nation.Effects {
		if effect.Kind == EffectMaxSciencePercent {
			maxSciencePercent = effect.Value
		}
	}
	return maxSciencePercent
}

func researchRate(state State, nation Nation) int {
	rate := 1
	nationResearchFlat := 0
	nationResearchPercent := 0
	maxSciencePercent := maxSciencePercentForNation(nation)
	for _, effect := range nation.Effects {
		switch effect.Kind {
		case EffectResearch:
			nationResearchFlat += effect.Value
		case EffectResearchPercent:
			nationResearchPercent += effect.Value
		}
	}
	sciencePercent := nationSciencePercent(nation, maxSciencePercent)
	cityResearch := 0
	for _, city := range state.Cities {
		if nationForPlayer(state, city.OwnerID) != nation.ID {
			continue
		}
		cityGoldResearch := city.Yields.Gold * sciencePercent / 100
		cityRate := cityGoldResearch + city.Yields.Science + ownerWonderEffectSum(state, city.OwnerID, EffectIncreaseScientists)
		cityResearchFlat := 0
		cityResearchPercent := 0
		for _, effect := range cityResearchEffects(state, city) {
			switch effect.Kind {
			case EffectResearch:
				cityResearchFlat += effect.Value
			case EffectResearchPercent:
				cityResearchPercent += effect.Value
			}
		}
		cityRate += signedYieldPercentDelta(cityRate, cityResearchPercent)
		cityRate += cityResearchFlat
		if cityRate > 0 {
			cityResearch += cityRate
		}
	}
	rate += cityResearch + signedYieldPercentDelta(cityResearch, nationResearchPercent) + nationResearchFlat
	if rate < 0 {
		return 0
	}
	return rate
}

func nationSciencePercent(nation Nation, maxSciencePercent int) int {
	percent := 100
	if nation.SciencePercent != nil {
		percent = *nation.SciencePercent
	}
	if percent > 100 {
		percent = 100
	}
	if maxSciencePercent >= 0 && percent > maxSciencePercent {
		percent = maxSciencePercent
	}
	if percent < 0 {
		return 0
	}
	return percent
}

func unitUpkeepForNation(state State, nationID string) int {
	upkeep := 0
	for _, unit := range state.Units {
		if nationForPlayer(state, unit.OwnerID) == nationID {
			upkeep += unit.Upkeep
		}
	}
	return upkeep
}

func cityResearchEffects(state State, city City) []Effect {
	effects := append([]Effect(nil), city.Effects...)
	for _, building := range city.Buildings {
		if !isObsoleteForPlayer(state, city.OwnerID, city.Obsolete, building.ObsoletedBy) {
			effects = append(effects, building.Effects...)
		}
	}
	for _, wonder := range state.Wonders {
		if !wonder.Destroyed && !isObsoleteForPlayer(state, wonder.OwnerID, city.Obsolete, wonder.ObsoletedBy) && (wonder.Scope == "global" || wonder.OwnerID == city.OwnerID) {
			effects = append(effects, wonder.Effects...)
		}
	}
	return effects
}

func nextResearchAdvance(nation Nation, runtime rules.RuntimeRules) string {
	for _, advance := range runtime.Advances {
		if canResearchAdvance(nation, runtime, advance.ID) {
			return advance.ID
		}
	}
	return ""
}

func canResearchAdvance(nation Nation, runtime rules.RuntimeRules, advanceID string) bool {
	advance, ok := runtime.Advance(advanceID)
	if !ok || containsString(nation.Advances, advanceID) {
		return false
	}
	for _, prerequisite := range advance.Prerequisites {
		if !containsString(nation.Advances, prerequisite) {
			return false
		}
	}
	return true
}

func endOwnerGame(state *State, owner string) {
	setTerminal(state, owner, "owner_ended")
}

func setTerminal(state *State, winner, reason string) {
	state.Terminal = TerminalState{Ended: true, Reason: reason, WinningPlayerID: winner}
	state.Phase.ResolutionLocked = true
	state.Deadlines = nil
	state.Phase.DeadlineID = ""
}

func evaluateVictory(state State, runtime *rules.RuntimeRules) (string, string) {
	evaluators := supportedVictoryEvaluators
	if runtime != nil {
		evaluators = runtime.VictoryEvaluatorIDs()
	}
	for _, evaluator := range evaluators {
		var winner string
		switch evaluator {
		case "bloodlust_conquest":
			winner = conquestWinner(state, runtime)
		case "diplomatic_world_union":
			winner = unionWinner(state)
		case "score_2300":
			winner = scoreWinner(state, runtime)
		case "science_gaia_controller":
			winner = gaiaWinner(state, runtime)
		}
		if winner != "" {
			return winner, evaluator
		}
	}
	return "", ""
}

func conquestWinner(state State, runtime *rules.RuntimeRules) string {
	if runtime == nil {
		return soleActiveNationPlayer(state)
	}
	nations := survivingNationIDs(state)
	if len(nations) != 1 {
		return ""
	}
	return playerForNation(state, nations[0])
}

func unionWinner(state State) string {
	activeNations := activeNationIDs(state)
	if len(activeNations) < 2 {
		return ""
	}
	activePlayers := activePlayerIDs(state)
	for _, nationID := range activeNations {
		nation := nationForID(&state, nationID)
		if nation == nil || !containsAll(nation.Embassies, activeNations, nationID) {
			if !hasUnionTreaties(state, activePlayers) {
				return ""
			}
			break
		}
	}
	return playerForNation(state, activeNations[0])
}

func hasUnionTreaties(state State, players []string) bool {
	if len(players) < 2 {
		return false
	}
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			relation := relationForPlayers(state, players[i], players[j])
			if relation == nil || relation.Treaty != "union" {
				return false
			}
		}
	}
	return true
}

func activePlayerIDs(state State) []string {
	ids := make([]string, 0, len(state.Players))
	for _, player := range state.Players {
		if player.Lifecycle == PlayerActive {
			ids = append(ids, player.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func scoreWinner(state State, runtime *rules.RuntimeRules) string {
	finalTurn := 2300
	if runtime != nil {
		finalTurn = runtime.Calendar.FinalTurn
	}
	if state.Phase.Turn < finalTurn {
		return ""
	}
	scores := make(map[string]int)
	for _, city := range state.Cities {
		if player, ok := playerByIDValue(state, city.OwnerID); ok && player.Lifecycle == PlayerActive {
			scores[city.OwnerID] += city.Yields.Victory + city.Yields.Food + city.Yields.Production + city.Yields.Gold
		}
	}
	winner, best, tied := "", 0, false
	for _, playerID := range sortedKeys(scores) {
		score := scores[playerID]
		if winner == "" || score > best {
			winner, best, tied = playerID, score, false
		} else if score == best {
			tied = true
		}
	}
	if winner == "" || best <= 0 || tied {
		return ""
	}
	return winner
}

func gaiaWinner(state State, runtime *rules.RuntimeRules) string {
	if runtime == nil {
		for _, nation := range state.Nations {
			for _, effect := range nation.Effects {
				if effect.Kind == EffectVictory && effect.Value > 0 {
					return playerForNation(state, nation.ID)
				}
			}
		}
		return ""
	}
	buildingID, wonderID, ok := gaiaPrerequisiteIDs(runtime)
	if !ok {
		return ""
	}
	for _, nation := range state.Nations {
		if nation.GaiaCompletionTurn > 0 && state.Phase.Turn >= nation.GaiaCompletionTurn && nationHasGaiaPrerequisites(state, nation.ID, buildingID, wonderID) {
			return playerForNation(state, nation.ID)
		}
	}
	return ""
}

func updateGaiaCountdown(state *State, runtime *rules.RuntimeRules, resolvedTurn int) {
	if runtime == nil {
		return
	}
	buildingID, wonderID, ok := gaiaPrerequisiteIDs(runtime)
	for index := range state.Nations {
		nation := &state.Nations[index]
		if !ok || !nationHasGaiaPrerequisites(*state, nation.ID, buildingID, wonderID) {
			nation.GaiaCompletionTurn = 0
			continue
		}
		if nation.GaiaCompletionTurn == 0 {
			nation.GaiaCompletionTurn = resolvedTurn + runtime.Gaia.ActivationTurns
		}
	}
}

func gaiaPrerequisiteIDs(runtime *rules.RuntimeRules) (string, string, bool) {
	if runtime == nil {
		return "", "", false
	}
	if runtime.Gaia.BuildingID != "" && runtime.Gaia.WonderID != "" {
		return runtime.Gaia.BuildingID, runtime.Gaia.WonderID, true
	}
	buildingID, wonderID := "", ""
	for _, production := range runtime.Production {
		switch {
		case production.Kind == ProductionBuilding && production.ID == "IMPROVE_SYNTHETIC":
			buildingID = production.ID
		case production.Kind == ProductionWonder && production.ID == "WONDER_SYNTHETIC":
			wonderID = production.ID
		}
	}
	return buildingID, wonderID, buildingID != "" && wonderID != ""
}

func nationHasGaiaPrerequisites(state State, nationID, buildingID, wonderID string) bool {
	owners := make(map[string]bool)
	for _, player := range state.Players {
		if player.NationID == nationID {
			owners[player.ID] = true
		}
	}
	if len(owners) == 0 {
		return false
	}
	return nationHasGaiaBuilding(state, owners, buildingID) && nationHasGaiaWonder(state, owners, wonderID)
}

func nationHasGaiaBuilding(state State, owners map[string]bool, buildingID string) bool {
	for _, city := range state.Cities {
		if !owners[city.OwnerID] {
			continue
		}
		for _, building := range city.Buildings {
			if building.ID == buildingID && !isObsolete(city.Obsolete, building.ObsoletedBy) {
				return true
			}
		}
	}
	return false
}

func nationHasGaiaWonder(state State, owners map[string]bool, wonderID string) bool {
	for _, wonder := range state.Wonders {
		if wonder.ID != wonderID || wonder.Destroyed || !owners[wonder.OwnerID] {
			continue
		}
		city, ok := cityByID(state, wonder.CityID)
		if !ok || isObsolete(city.Obsolete, wonder.ObsoletedBy) {
			continue
		}
		return true
	}
	return false
}

func soleActiveNationPlayer(state State) string {
	nations := activeNationIDs(state)
	if len(nations) != 1 {
		return ""
	}
	return playerForNation(state, nations[0])
}

func activeNationIDs(state State) []string {
	seen := make(map[string]bool)
	for _, player := range state.Players {
		if player.Lifecycle == PlayerActive && player.NationID != "" {
			seen[player.NationID] = true
		}
	}
	return sortedKeys(seen)
}

func survivingNationIDs(state State) []string {
	seen := make(map[string]bool)
	for _, player := range state.Players {
		if player.Lifecycle == PlayerActive && player.NationID != "" && nationHasAssets(state, player) {
			seen[player.NationID] = true
		}
	}
	return sortedKeys(seen)
}

func nationHasAssets(state State, player Player) bool {
	for _, unit := range state.Units {
		if unit.OwnerID == player.ID {
			return true
		}
	}
	for _, city := range state.Cities {
		if city.OwnerID == player.ID {
			return true
		}
	}
	return false
}

func playerForNation(state State, nationID string) string {
	var players []string
	for _, player := range state.Players {
		if player.NationID == nationID && player.Lifecycle == PlayerActive {
			players = append(players, player.ID)
		}
	}
	sort.Strings(players)
	if len(players) == 0 {
		return ""
	}
	return players[0]
}

func playerByID(state *State, id string) (*Player, bool) {
	for index := range state.Players {
		if state.Players[index].ID == id {
			return &state.Players[index], true
		}
	}
	return nil, false
}

func playerByIDValue(state State, id string) (Player, bool) {
	for _, player := range state.Players {
		if player.ID == id {
			return player, true
		}
	}
	return Player{}, false
}

func terminalPayload(state *State) json.RawMessage {
	return lifecyclePayload(state.Terminal.WinningPlayerID, state.Terminal.Reason)
}

func lifecyclePayload(playerID, value string) json.RawMessage {
	body, _ := json.Marshal(struct {
		PlayerID string `json:"player_id"`
		Value    string `json:"value"`
	}{PlayerID: playerID, Value: value})
	return body
}

func containsAll(values, required []string, self string) bool {
	for _, value := range required {
		if value != self && !containsLifecycleString(values, value) {
			return false
		}
	}
	return true
}

func containsLifecycleString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
