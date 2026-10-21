package engine

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// researchGoalNext uses deterministic prerequisite-ID ordering and validates the
// complete remaining graph. A recursion-stack guard fixes the original recursive
// goal selector's potential nontermination on malformed/self-referential data.
func researchGoalNext(nation Nation, runtime rules.RuntimeRules, goal string) (string, error) {
	visiting := map[string]bool{}
	memo := map[string]string{}
	var visit func(string) (string, error)
	visit = func(id string) (string, error) {
		if containsString(nation.Advances, id) {
			return "", nil
		}
		if visiting[id] {
			return "", errors.New("cyclic research goal")
		}
		if next, ok := memo[id]; ok {
			return next, nil
		}
		advance, ok := runtime.Advance(id)
		if !ok || advanceBeyondEndAge(advance, runtime) {
			return "", errors.New("research goal is unavailable")
		}
		visiting[id] = true
		prerequisites := append([]string(nil), advance.Prerequisites...)
		sort.Strings(prerequisites)
		next := ""
		for _, prerequisite := range prerequisites {
			candidate, err := visit(prerequisite)
			if err != nil {
				return "", err
			}
			if next == "" {
				next = candidate
			}
		}
		delete(visiting, id)
		if next == "" && canResearchAdvance(nation, runtime, id) {
			next = id
		}
		if next == "" {
			return "", errors.New("research goal has no legal prerequisite")
		}
		memo[id] = next
		return next, nil
	}
	return visit(goal)
}

// CanSetResearchGoal advertises only owned, unreached goals with a legal path
// under the pinned end-age/prerequisite rules; static rules reveal no foreign work.
func CanSetResearchGoal(nation Nation, runtime rules.RuntimeRules, goal string) bool {
	if goal == "" || containsString(nation.Advances, goal) {
		return false
	}
	next, err := researchGoalNext(nation, runtime, goal)
	return err == nil && next != ""
}

func setResearchGoal(state *State, actor string, payload setResearchPayload, runtime *rules.RuntimeRules) (string, json.RawMessage, error) {
	if err := activeActor(*state, actor); err != nil {
		return "", nil, err
	}
	nation := nationForID(state, nationForPlayer(*state, actor))
	if nation == nil || runtime == nil {
		return "", nil, errors.New("research rules are unavailable")
	}
	if payload.AdvanceID != "" {
		if !CanSetResearchGoal(*nation, *runtime, payload.AdvanceID) {
			return "", nil, errors.New("research goal is unavailable")
		}
		next, err := researchGoalNext(*nation, *runtime, payload.AdvanceID)
		if err != nil {
			return "", nil, err
		}
		if next != nation.ResearchTargetID {
			nation.ResearchTargetID = next
			nation.ResearchProgress = 0
		}
	}
	// Clearing a goal cancels future routing, not the currently accumulated work.
	nation.ResearchGoalID = payload.AdvanceID
	raw, _ := json.Marshal(payload)
	return "research.goal_changed", raw, nil
}

func clearCompletedResearchGoal(nation *Nation) {
	if nation.ResearchGoalID != "" && containsString(nation.Advances, nation.ResearchGoalID) {
		nation.ResearchGoalID = ""
	}
}
