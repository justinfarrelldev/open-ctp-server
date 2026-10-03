package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailReadinessSupportFloorsAfterSummingFractionalCosts(t *testing.T) {
	runtime := &rules.RuntimeRules{Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_TEST", SupportCoef: 1, ReadyPeaceCoef: 0.6, ReadyWarCoef: 1}}}
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", GovernmentID: "GOVERNMENT_TEST", Readiness: "peace"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Production: 2}}},
		Units:   []Unit{{ID: "unit-a", OwnerID: "player-a", ShieldHunger: 1}, {ID: "unit-b", OwnerID: "player-a", ShieldHunger: 1}},
	}

	applyReadinessSupportTax(&state, runtime)

	if got, want := state.Cities[0].Yields.Production, 1; got != want {
		t.Fatalf("production after fractional support = %d, want %d", got, want)
	}
}
