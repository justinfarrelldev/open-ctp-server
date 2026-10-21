package engine

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"testing"
)

func TestClientPersistentResearchGoalSelectsPrerequisitesAndCompletes(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state = runtimeApply(t, state, "player-a", CommandSetResearchGoal, setResearchPayload{AdvanceID: "ADVANCE_SYNTHETIC_UNLOCK"})
	nation := nationForID(&state, "nation-a")
	if nation.ResearchGoalID != "ADVANCE_SYNTHETIC_UNLOCK" || nation.ResearchTargetID != "ADVANCE_SYNTHETIC_ROOT" {
		t.Fatalf("goal bypassed prerequisites: %+v", nation)
	}
	raw, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	state, err = DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	for rounds := 0; rounds < 20 && !containsString(nationForID(&state, "nation-a").Advances, "ADVANCE_SYNTHETIC_UNLOCK"); rounds++ {
		state = runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	}
	nation = nationForID(&state, "nation-a")
	if !containsString(nation.Advances, "ADVANCE_SYNTHETIC_UNLOCK") || nation.ResearchGoalID != "" {
		t.Fatalf("retained goal failed to progress/clear: %+v", nation)
	}
}

func TestClientResearchGoalChangeAndClearDoNotResetUnchangedCurrentResearch(t *testing.T) {
	state := runtimeStartedState(t)
	state = runtimeApply(t, state, "player-a", CommandSetResearchGoal, setResearchPayload{AdvanceID: "ADVANCE_SYNTHETIC_UNLOCK"})
	nationForID(&state, "nation-a").ResearchProgress = 1
	state = runtimeApply(t, state, "player-a", CommandSetResearchGoal, setResearchPayload{AdvanceID: "ADVANCE_SYNTHETIC_ROOT"})
	if nationForID(&state, "nation-a").ResearchProgress != 1 {
		t.Fatal("goal change reset unchanged current research")
	}
	state = runtimeApply(t, state, "player-a", CommandSetResearchGoal, setResearchPayload{})
	nation := nationForID(&state, "nation-a")
	if nation.ResearchGoalID != "" || nation.ResearchTargetID != "ADVANCE_SYNTHETIC_ROOT" || nation.ResearchProgress != 1 {
		t.Fatal("clearing goal cancelled current work")
	}
	if err := runtimeApplyExpectError(t, state, "player-a", CommandSetResearchGoal, setResearchPayload{AdvanceID: "missing"}); err == nil {
		t.Fatal("unknown goal accepted")
	}
}

func TestClientResearchGoalCycleFailsClosedInsteadOfRecursingForever(t *testing.T) {
	runtime := rules.RuntimeRules{Advances: []rules.RuntimeAdvanceDefinition{{ID: "a", Prerequisites: []string{"b"}}, {ID: "b", Prerequisites: []string{"a"}}}}
	if CanSetResearchGoal(Nation{}, runtime, "a") {
		t.Fatal("cyclic goal was advertised")
	}
}
