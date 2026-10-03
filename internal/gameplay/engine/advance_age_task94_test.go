package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailStartingAgeGrantsEarlierAgeAdvances(t *testing.T) {
	runtime := rules.RuntimeRules{
		Ages:       []rules.RuntimeAgeDefinition{{ID: "AGE_ONE", Index: 0}, {ID: "AGE_TWO", Index: 1}, {ID: "AGE_THREE", Index: 2}},
		StartAgeID: "AGE_THREE",
		Advances: []rules.RuntimeAdvanceDefinition{
			{ID: "ADVANCE_CURRENT", Cost: 1, AgeID: "AGE_THREE", Prerequisites: []string{"ADVANCE_MIDDLE"}},
			{ID: "ADVANCE_MIDDLE", Cost: 1, AgeID: "AGE_TWO", Prerequisites: []string{"ADVANCE_ROOT"}},
			{ID: "ADVANCE_ROOT", Cost: 1, AgeID: "AGE_ONE"},
		},
	}
	state := State{Nations: []Nation{{ID: "nation-a"}}}

	GrantStartingAgeAdvances(&state, &runtime, "")

	if got, want := state.Nations[0].CurrentAgeID, "AGE_THREE"; got != want {
		t.Fatalf("current age = %q, want %q", got, want)
	}
	if !containsString(state.Nations[0].Advances, "ADVANCE_ROOT") || !containsString(state.Nations[0].Advances, "ADVANCE_MIDDLE") || containsString(state.Nations[0].Advances, "ADVANCE_CURRENT") {
		t.Fatalf("starting advances = %v, want earlier ages only", state.Nations[0].Advances)
	}
	if got, want := state.Nations[0].ResearchTargetID, "ADVANCE_CURRENT"; got != want {
		t.Fatalf("research target = %q, want %q", got, want)
	}
}

func TestRetailAdvanceCompletionEntersNewAgeAfterNeedAdvances(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Starter:              rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
		Production:           []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}},
		Ages:                 []rules.RuntimeAgeDefinition{{ID: "AGE_ONE", Index: 0}, {ID: "AGE_TWO", Index: 1, NeedAdvances: 2}},
		StartAgeID:           "AGE_ONE",
		EndAgeID:             "AGE_TWO",
		Advances: []rules.RuntimeAdvanceDefinition{
			{ID: "ADVANCE_ROOT", Cost: 1, AgeID: "AGE_ONE"},
			{ID: "ADVANCE_AGE_TWO_A", Cost: 1, AgeID: "AGE_TWO", Prerequisites: []string{"ADVANCE_ROOT"}},
			{ID: "ADVANCE_AGE_TWO_B", Cost: 1, AgeID: "AGE_TWO", Prerequisites: []string{"ADVANCE_ROOT"}},
		},
	}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := State{
		SchemaVersion: StateSchemaV1,
		Rules:         RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: definitions},
		Phase:         PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", ActivePlayerOrder: []string{"player-a"}},
		Players:       []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Nations:       []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_ROOT", "ADVANCE_AGE_TWO_A"}, CurrentAgeID: "AGE_ONE", ResearchTargetID: "ADVANCE_AGE_TWO_B"}},
		RNG:           []RNGDomainState{{Domain: RNGDomainCombat}, {Domain: RNGDomainSpecialAction}, {Domain: RNGDomainAI}, {Domain: RNGDomainRiot}},
	}
	payload, _ := json.Marshal(struct{}{})

	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "round", ActorPlayerID: "player-a", Type: CommandRoundComplete, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := next.Nations[0].CurrentAgeID, "AGE_TWO"; got != want {
		t.Fatalf("current age = %q, want %q", got, want)
	}
}
