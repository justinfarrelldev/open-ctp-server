package engine

import "testing"

func TestRetailResearchRateUsesCityTradeYield(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "gov-research", Kind: EffectResearch, Value: 2}}}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 5}}},
	}
	if got, want := researchRate(state, state.Nations[0]), 8; got != want {
		t.Fatalf("research rate = %d, want base 1 + city trade 5 + effect 2 = %d", got, want)
	}
}
