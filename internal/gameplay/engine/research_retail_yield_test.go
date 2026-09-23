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

func TestResearchRateSkipsObsoleteRetailBuildingScience(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_OBSOLETE"}}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 5}, Buildings: []Building{{ID: "IMPROVE_OLD_ACADEMY", ObsoletedBy: "ADVANCE_OBSOLETE", Effects: []Effect{{ID: "academy-science", Kind: EffectResearch, Value: 10}}}}}},
	}
	if got := researchRate(state, state.Nations[0]); got != 6 {
		t.Fatalf("research rate = %d, want base + gold only", got)
	}
}

func TestRetailResearchRateUsesCityScienceBuildingEffect(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 5}, Buildings: []Building{{ID: "IMPROVE_ACADEMY", Effects: []Effect{{ID: "academy-science", Kind: EffectResearch, Value: 10}}}}}},
	}
	if got, want := researchRate(state, state.Nations[0]), 16; got != want {
		t.Fatalf("research rate = %d, want base 1 + city trade 5 + academy science 10 = %d", got, want)
	}
}
