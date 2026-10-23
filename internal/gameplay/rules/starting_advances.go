package rules

import (
	"fmt"
	"slices"
)

const StartingAdvancePolicyVersion = "ctp2-starting-advances-v1"
const DefaultStartingDifficulty = "medium"

// The unnamed DiffDB blocks have this fixed original source order. Local profile
// preferences are not authoritative server configuration.
var startingDifficultyIDs = []string{"beginner", "easy", "medium", "hard", "very-hard", "impossible"}

// StartingAdvanceChance order is semantic: earlier successes consume the cap.
// Chances >=100 are guaranteed even beyond the cap, as Player::InitPlayer does.
type StartingAdvanceChance struct {
	AdvanceID   string `json:"advance_id"`
	HumanChance int    `json:"human_chance"`
	AIChance    int    `json:"ai_chance"`
}

type StartingAdvanceDifficulty struct {
	ID               string                  `json:"id"`
	MaxHumanAdvances int                     `json:"max_human_advances"`
	MaxAIAdvances    int                     `json:"max_ai_advances"`
	Chances          []StartingAdvanceChance `json:"chances"`
}

// RuntimeStartingAdvances is a NEW-GAME policy pinned in canonical definitions.
// Nil legacy/synthetic policies are never inferred or silently backfilled.
type RuntimeStartingAdvances struct {
	Version      string                      `json:"version"`
	DifficultyID string                      `json:"difficulty_id"`
	Difficulties []StartingAdvanceDifficulty `json:"difficulties"`
}

func (p RuntimeStartingAdvances) Difficulty() (StartingAdvanceDifficulty, bool) {
	for _, difficulty := range p.Difficulties {
		if difficulty.ID == p.DifficultyID {
			return difficulty, true
		}
	}
	return StartingAdvanceDifficulty{}, false
}

// ValidateStartingAdvances is shared by import, pin compilation and bootstrap.
// Neither profile nor chance order is sorted: source traversal determines grants.
func ValidateStartingAdvances(runtime RuntimeRules) error {
	policy := runtime.StartingAdvances
	if policy == nil {
		return nil
	}
	if policy.Version != StartingAdvancePolicyVersion || len(policy.Difficulties) == 0 || len(policy.Difficulties) > len(startingDifficultyIDs) {
		return fmt.Errorf("starting advance policy is unsupported")
	}
	start, startOK := runtime.Age(runtime.StartAgeID)
	end, endOK := runtime.Age(runtime.EndAgeID)
	if !startOK || !endOK || start.Index > end.Index {
		return fmt.Errorf("starting advance policy requires a valid start/end age interval")
	}
	advances := make(map[string]bool, len(runtime.Advances))
	for _, advance := range runtime.Advances {
		advances[advance.ID] = true
	}
	seen := map[string]bool{}
	for _, profile := range policy.Difficulties {
		if !slices.Contains(startingDifficultyIDs, profile.ID) || seen[profile.ID] || profile.MaxHumanAdvances < 0 || profile.MaxAIAdvances < 0 || len(profile.Chances) == 0 {
			return fmt.Errorf("starting advance difficulty is malformed or duplicated")
		}
		seen[profile.ID] = true
		choices := map[string]bool{}
		for _, choice := range profile.Chances {
			if !advances[choice.AdvanceID] || choices[choice.AdvanceID] || choice.HumanChance < 0 || choice.AIChance < 0 {
				return fmt.Errorf("starting advance chance is malformed, duplicated or references unknown content")
			}
			choices[choice.AdvanceID] = true
		}
	}
	if !seen[policy.DifficultyID] {
		return fmt.Errorf("selected starting advance difficulty is unavailable")
	}
	return nil
}

func projectStartingAdvances(profiles []StartingAdvanceDifficulty) *RuntimeStartingAdvances {
	if len(profiles) == 0 {
		return nil
	}
	policy := &RuntimeStartingAdvances{Version: StartingAdvancePolicyVersion, DifficultyID: DefaultStartingDifficulty, Difficulties: make([]StartingAdvanceDifficulty, len(profiles))}
	for n, profile := range profiles {
		policy.Difficulties[n] = profile
		policy.Difficulties[n].Chances = append([]StartingAdvanceChance(nil), profile.Chances...)
	}
	return policy
}
