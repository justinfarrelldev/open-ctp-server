package engine

import (
	"encoding/json"
	"errors"
	"sort"

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
	if runtime == nil || runtime.Starter.Movement < 1 {
		return
	}
	for index := range state.Units {
		if state.Units[index].TransportID == "" {
			state.Units[index].Movement = runtime.Starter.Movement
		}
	}
	advanceResearch(state, runtime)
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
			nation.ResearchTargetID = nextResearchAdvance(*nation, *runtime)
			if nation.ResearchTargetID == "" {
				nation.ResearchProgress = 0
			}
		}
	}
}

func researchRate(state State, nation Nation) int {
	rate := 1
	for _, effect := range nation.Effects {
		if effect.Kind == EffectResearch {
			rate += effect.Value
		}
	}
	for _, city := range state.Cities {
		if nationForPlayer(state, city.OwnerID) != nation.ID {
			continue
		}
		for _, effect := range cityEffects(state, city) {
			if effect.Kind == EffectResearch {
				rate += effect.Value
			}
		}
	}
	return rate
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
	for _, nationID := range activeNations {
		nation := nationForID(&state, nationID)
		if nation == nil || !containsAll(nation.Embassies, activeNations, nationID) {
			return ""
		}
	}
	return playerForNation(state, activeNations[0])
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
