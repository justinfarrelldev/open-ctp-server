package engine

import (
	"bytes"
	"slices"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func startingAdvanceFixture() (State, rules.RuntimeRules) {
	state := fixtureState()
	state.Nations = []Nation{{ID: "nation-a"}, {ID: "nation-b"}}
	for n := range state.Players {
		if state.Players[n].ID == "player-b" {
			state.Players[n].Controller = ControllerAI
		}
	}
	runtime := rules.RuntimeRules{StartAgeID: "AGE_ONE", EndAgeID: "AGE_TWO", Ages: []rules.RuntimeAgeDefinition{{ID: "AGE_ONE", Index: 0}, {ID: "AGE_TWO", Index: 1}}, Advances: []rules.RuntimeAdvanceDefinition{
		{ID: "a", AgeID: "AGE_ONE"}, {ID: "b", AgeID: "AGE_ONE"}, {ID: "late", AgeID: "AGE_ONE"}, {ID: "research", AgeID: "AGE_TWO", Prerequisites: []string{"tool"}}, {ID: "tool", AgeID: "AGE_ONE"},
	}, StartingAdvances: &rules.RuntimeStartingAdvances{Version: rules.StartingAdvancePolicyVersion, DifficultyID: "medium", Difficulties: []rules.StartingAdvanceDifficulty{{ID: "medium", MaxHumanAdvances: 2, MaxAIAdvances: 0, Chances: []rules.StartingAdvanceChance{
		{AdvanceID: "tool", HumanChance: 100, AIChance: 100},
		{AdvanceID: "b", HumanChance: 99, AIChance: 0},
		{AdvanceID: "a", HumanChance: 99, AIChance: 100},
		{AdvanceID: "late", HumanChance: 120, AIChance: 100},
	}}}}}
	return state, runtime
}

func TestStartingAdvancesPreserveSourceOrderCapsAndControllerDistinctions(t *testing.T) {
	state, runtime := startingAdvanceFixture()
	before, _ := CanonicalStateJSON(state)
	draws := map[string]int{}
	next, err := InitializeStartingAdvances(state, &runtime, func(player string) int { draws[player]++; return 0 })
	if err != nil {
		t.Fatal(err)
	}
	human := nationForID(&next, "nation-a")
	ai := nationForID(&next, "nation-b")
	if !slices.Equal(human.Advances, []string{"b", "late", "tool"}) || !slices.Equal(ai.Advances, []string{"a", "late", "tool"}) {
		t.Fatalf("source/cap semantics lost: human=%+v AI=%+v", human, ai)
	}
	if draws["player-a"] != 1 || draws["player-b"] != 0 {
		t.Fatalf("guaranteed/capped rows consumed random draws: %v", draws)
	}
	for _, nation := range next.Nations {
		if nation.CurrentAgeID != "AGE_ONE" || nation.ResearchTargetID == "" || !CanResearchAdvance(nation, runtime, nation.ResearchTargetID) {
			t.Fatalf("age/research omitted: %+v", nation)
		}
	}
	after, _ := CanonicalStateJSON(state)
	if !bytes.Equal(before, after) {
		t.Fatal("initialization mutated input")
	}
	raw, _ := CanonicalStateJSON(next)
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := InitializeStartingAdvances(restored, &runtime, func(string) int { t.Fatal("restart rerolled grants"); return 0 })
	if err != nil {
		t.Fatal(err)
	}
	replayRaw, _ := CanonicalStateJSON(replayed)
	if !bytes.Equal(raw, replayRaw) {
		t.Fatal("repeat initialization changed canonical state")
	}
	if len(next.AchievedFeatIDs) != 0 || len(next.Notifications) != 0 || next.Revision != state.Revision || next.EventSequence != state.EventSequence {
		t.Fatal("initial grants became earned-research rewards/events")
	}
}

func TestStartingAdvanceChanceThresholdAndZeroChanceDraws(t *testing.T) {
	state, runtime := startingAdvanceFixture()
	runtime.StartingAdvances.Difficulties[0].MaxHumanAdvances = 4
	runtime.StartingAdvances.Difficulties[0].MaxAIAdvances = 4
	draws := map[string]int{}
	next, err := InitializeStartingAdvances(state, &runtime, func(player string) int {
		draws[player]++
		if player == "player-a" {
			return 99
		}
		return 0
	})
	if err != nil {
		t.Fatal(err)
	}
	human := nationForID(&next, "nation-a")
	ai := nationForID(&next, "nation-b")
	if !slices.Equal(human.Advances, []string{"late", "tool"}) || !slices.Equal(ai.Advances, []string{"a", "late", "tool"}) || draws["player-a"] != 2 || draws["player-b"] != 1 {
		t.Fatalf("chance threshold/zero policy lost: human=%+v AI=%+v draws=%v", human, ai, draws)
	}
}

func TestStartingAdvancesLaterAgeAndLegacyBoundaries(t *testing.T) {
	state, runtime := startingAdvanceFixture()
	runtime.StartAgeID = "AGE_TWO"
	next, err := InitializeStartingAdvances(state, &runtime, func(string) int { t.Fatal("later start rolled first-age chance table"); return 0 })
	if err != nil {
		t.Fatal(err)
	}
	for _, nation := range next.Nations {
		if !slices.Equal(nation.Advances, []string{"a", "b", "late", "tool"}) || nation.ResearchTargetID != "research" || nation.CurrentAgeID != "AGE_TWO" {
			t.Fatalf("later-age grants/research: %+v", nation)
		}
	}
	state.Revision = 1
	if _, err := InitializeStartingAdvances(state, &runtime, func(string) int { return 0 }); err == nil {
		t.Fatal("existing save was implicitly initialized")
	}
	runtime.StartingAdvances = nil
	unchanged, err := InitializeStartingAdvances(state, &runtime, nil)
	if err != nil {
		t.Fatal(err)
	}
	original, _ := CanonicalStateJSON(state)
	legacy, _ := CanonicalStateJSON(unchanged)
	if !bytes.Equal(original, legacy) {
		t.Fatal("legacy/synthetic pin was backfilled")
	}
}

func TestStartingAdvanceInvalidRollOrPolicyRollsBack(t *testing.T) {
	for _, invalid := range []int{-1, 100} {
		state, runtime := startingAdvanceFixture()
		before, _ := CanonicalStateJSON(state)
		if _, err := InitializeStartingAdvances(state, &runtime, func(string) int { return invalid }); err == nil {
			t.Fatal("invalid RNG result accepted")
		}
		after, _ := CanonicalStateJSON(state)
		if !bytes.Equal(before, after) {
			t.Fatal("failed grants mutated authority")
		}
	}
	state, runtime := startingAdvanceFixture()
	runtime.StartingAdvances.DifficultyID = "unknown"
	if _, err := InitializeStartingAdvances(state, &runtime, func(string) int { t.Fatal("invalid policy consumed RNG"); return 0 }); err == nil {
		t.Fatal("invalid difficulty accepted")
	}
}
