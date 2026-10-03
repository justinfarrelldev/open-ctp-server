package visibility

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestProjectionSuppressesResearchBeyondConfiguredEndAge(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Starter:              rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1},
		Production:           []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: engine.ProductionUnit}},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: engine.DomainLand, Movement: 1, Health: 1}},
		Ages:                 []rules.RuntimeAgeDefinition{{ID: "AGE_ONE", Index: 0}, {ID: "AGE_TWO", Index: 1}, {ID: "AGE_THREE", Index: 2}},
		EndAgeID:             "AGE_TWO",
		Advances: []rules.RuntimeAdvanceDefinition{
			{ID: "ADVANCE_ROOT", Cost: 1, AgeID: "AGE_ONE"},
			{ID: "ADVANCE_ALLOWED", Cost: 1, AgeID: "AGE_TWO", Prerequisites: []string{"ADVANCE_ROOT"}},
			{ID: "ADVANCE_FUTURE", Cost: 1, AgeID: "AGE_THREE", Prerequisites: []string{"ADVANCE_ROOT"}},
		},
	}
	state := projectionStateWithRuntime(t, runtime)
	state.Nations = []engine.Nation{{ID: "a", Advances: []string{"ADVANCE_ROOT"}}, {ID: "b"}}
	state.Players = []engine.Player{{ID: "a", NationID: "a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projection.Research == nil || !containsString(projection.Research.Available, "ADVANCE_ALLOWED") || containsString(projection.Research.Available, "ADVANCE_FUTURE") {
		t.Fatalf("available research = %+v, want allowed age and no future age", projection.Research)
	}
}
