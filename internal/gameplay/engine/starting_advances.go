package engine

import (
	"errors"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// StartingAdvanceRoll is an injected new-game-only [0,100) draw. It must use a
// deterministic server-owned stream per player, never client or ambient RNG.
type StartingAdvanceRoll func(playerID string) int

// InitializeStartingAdvances implements original Player::InitPlayer's bootstrap,
// not earned research completion. It allocates no events, command identities or
// rewards, and commits its isolated nation changes only on complete success.
// Loading a legacy pin/save never implicitly calls or enables this policy.
func InitializeStartingAdvances(state State, runtime *rules.RuntimeRules, roll StartingAdvanceRoll) (State, error) {
	if runtime == nil || runtime.StartingAdvances == nil || state.StartingAdvancesInitialized {
		return state, nil
	}
	if err := rules.ValidateStartingAdvances(*runtime); err != nil {
		return State{}, err
	}
	if err := ValidateState(state); err != nil {
		return State{}, err
	}
	if state.Revision != 0 || state.EventSequence != 0 || state.Phase.Turn != 1 || len(state.Cities) != 0 || len(state.Nations) != len(state.Players) {
		return State{}, errors.New("starting advances require a fresh new-game state")
	}
	for _, nation := range state.Nations {
		if len(nation.Advances) != 0 || nation.CurrentAgeID != "" || nation.ResearchTargetID != "" || nation.ResearchGoalID != "" || nation.ResearchProgress != 0 {
			return State{}, errors.New("starting advances cannot replace existing nation progress")
		}
	}
	seen := map[string]bool{}
	for _, player := range state.Players {
		if nationForID(&state, player.NationID) == nil || seen[player.NationID] || player.Lifecycle != PlayerActive || player.Controller != ControllerHuman && player.Controller != ControllerAI {
			return State{}, errors.New("new-game starting nation/controller is invalid")
		}
		seen[player.NationID] = true
	}
	profile, _ := runtime.StartingAdvances.Difficulty()
	start, _ := runtime.Age(runtime.StartAgeID)
	first, _ := runtime.FirstAge()
	candidate := state
	candidate.Nations = append([]Nation(nil), state.Nations...)
	for n := range candidate.Nations {
		candidate.Nations[n].Advances = nil
	}
	players := append([]Player(nil), state.Players...)
	sort.Slice(players, func(a, b int) bool { return players[a].ID < players[b].ID })
	for _, player := range players {
		nation := nationForID(&candidate, player.NationID)
		if start.Index != first.Index {
			continue
		}
		cap := profile.MaxHumanAdvances
		if player.Controller == ControllerAI {
			cap = profile.MaxAIAdvances
		}
		granted := 0
		for _, choice := range profile.Chances {
			chance := choice.HumanChance
			if player.Controller == ControllerAI {
				chance = choice.AIChance
			}
			give := chance >= 100
			if !give && granted < cap {
				if roll == nil {
					return State{}, errors.New("starting advance RNG is required")
				}
				draw := roll(player.ID)
				if draw < 0 || draw >= 100 {
					return State{}, errors.New("starting advance RNG result is outside [0,100)")
				}
				give = draw < chance
			}
			if give {
				nation.Advances = append(nation.Advances, choice.AdvanceID)
				granted++
			}
		}
	}
	// This existing helper supplies strictly earlier-age discoveries for later
	// starts, sets CurrentAgeID, and selects legal research AFTER all initial grants.
	// Runtime order is canonical ID order, our deterministic research tie-break.
	GrantStartingAgeAdvances(&candidate, runtime, start.ID)
	candidate.StartingAdvancesInitialized = true
	return candidate, nil
}
