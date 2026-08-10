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

func applyLifecycleCommand(state *State, command Command, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	switch command.Type {
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
	threshold := 0
	if runtime != nil {
		threshold = runtime.Gaia.ActivationTurns
	}
	for _, nation := range state.Nations {
		if threshold > 0 && len(nation.Advances) >= threshold {
			return playerForNation(state, nation.ID)
		}
		if runtime == nil {
			for _, effect := range nation.Effects {
				if effect.Kind == EffectVictory && effect.Value > 0 {
					return playerForNation(state, nation.ID)
				}
			}
		}
	}
	return ""
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
