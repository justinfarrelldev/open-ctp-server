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

func TestRetailResearchRateUsesCitySciencePercentEffect(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 10}, Buildings: []Building{{ID: "IMPROVE_ACADEMY", Effects: []Effect{{ID: "academy-science", Kind: EffectResearchPercent, Value: 10}}}}}},
	}
	if got, want := researchRate(state, state.Nations[0]), 12; got != want {
		t.Fatalf("research rate = %d, want base 1 + city trade 10 + academy 10%% = %d", got, want)
	}
}

func TestRetailResearchRateSubtractsUnitUpkeep(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}},
		Nations: []Nation{{ID: "nation-a"}, {ID: "nation-b"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 10}}},
		Units:   []Unit{{ID: "unit-a", OwnerID: "player-a", Upkeep: 3}, {ID: "unit-b", OwnerID: "player-b", Upkeep: 5}},
	}
	if got, want := researchRate(state, state.Nations[0]), 8; got != want {
		t.Fatalf("research rate = %d, want base 1 + city gold 10 - unit upkeep 3", got)
	}
}

func TestRetailResearchRateUsesNationSciencePercent(t *testing.T) {
	science := 50
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", SciencePercent: &science}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 12}}},
	}
	if got, want := researchRate(state, state.Nations[0]), 7; got != want {
		t.Fatalf("research rate = %d, want science slider half gold plus base = %d", got, want)
	}
	zero := 0
	state.Nations[0].SciencePercent = &zero
	if got, want := researchRate(state, state.Nations[0]), 1; got != want {
		t.Fatalf("research rate = %d, want explicit zero science to leave only base = %d", got, want)
	}
}

func TestRetailResearchRateHonorsGovernmentMaxScienceRate(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "government:democracy:max-science", Kind: EffectMaxSciencePercent, Value: 80}}}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 10}}},
	}
	if got, want := researchRate(state, state.Nations[0]), 9; got != want {
		t.Fatalf("research rate = %d, want base 1 + city trade capped to 8 by government = %d", got, want)
	}

	state.Nations[0].Effects[0].Value = 0
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_ACADEMY", Effects: []Effect{{ID: "academy-science", Kind: EffectResearch, Value: 3}}}}
	if got, want := researchRate(state, state.Nations[0]), 4; got != want {
		t.Fatalf("research rate = %d, want max science cap to leave flat city research uncapped = %d", got, want)
	}
}

func TestRetailResearchRateUsesGovernmentKnowledgeCoefficient(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "government:democracy:research", Kind: EffectResearchPercent, Value: 100}}}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 5}}},
	}
	if got, want := researchRate(state, state.Nations[0]), 11; got != want {
		t.Fatalf("research rate = %d, want base 1 + city trade 5 doubled by government = %d", got, want)
	}
}
