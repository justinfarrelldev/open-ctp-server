package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

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

func TestRetailUnitUpkeepDebitsTreasuryNotResearch(t *testing.T) {
	science := 100
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}},
		Nations: []Nation{{ID: "nation-a", Gold: 5, SciencePercent: &science}, {ID: "nation-b"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 10}}},
		Units:   []Unit{{ID: "unit-a", OwnerID: "player-a", Upkeep: 3}, {ID: "unit-b", OwnerID: "player-b", Upkeep: 5}},
	}
	if got, want := researchRate(state, state.Nations[0]), 11; got != want {
		t.Fatalf("research rate = %d, want base 1 + science gold 10 without unit-upkeep subtraction = %d", got, want)
	}
	advanceTreasury(&state)
	if got, want := state.Nations[0].Gold, 2; got != want {
		t.Fatalf("treasury gold = %d, want old 5 - unit upkeep 3 = %d", got, want)
	}
}

func TestRetailBuildingMaintenanceDeficitDebitsTreasury(t *testing.T) {
	science := 100
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations = []Nation{{ID: "nation-a", Gold: 5, SciencePercent: &science}}
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "expensive", Upkeep: 8}}, Effects: []Effect{{ID: "gold", Kind: EffectYieldGold, Value: 2}}}}
	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	advanceTreasury(&next)
	if got, want := next.Nations[0].Gold, 0; got != want {
		t.Fatalf("treasury gold = %d, want old 5 - maintenance deficit 5 = %d", got, want)
	}
}

func TestRetailTreasuryShortfallSellsHighestUpkeepBuilding(t *testing.T) {
	science := 100
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations = []Nation{{ID: "nation-a", Gold: 2, SciencePercent: &science}}
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "cheap", Upkeep: 1}, {ID: "expensive", Upkeep: 8}}, Effects: []Effect{{ID: "gold", Kind: EffectYieldGold, Value: 2}}}}
	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	advanceTreasury(&next)
	if got, want := next.Nations[0].Gold, 0; got != want {
		t.Fatalf("treasury gold = %d, want %d", got, want)
	}
	if len(next.Cities[0].Buildings) != 1 || next.Cities[0].Buildings[0].ID != "cheap" {
		t.Fatalf("buildings after shortfall = %#v, want only cheap", next.Cities[0].Buildings)
	}
}

func TestRetailTreasuryAccruesNonScienceCityGold(t *testing.T) {
	science := 40
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Gold: 5, SciencePercent: &science}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 10}}},
	}
	advanceTreasury(&state)
	if got, want := state.Nations[0].Gold, 11; got != want {
		t.Fatalf("nation gold = %d, want old 5 + non-science 60%% of 10 = %d", got, want)
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

func TestRetailConversionSiphonsGoldToConvertedOwner(t *testing.T) {
	zeroScience := 0
	state := State{
		Nations: []Nation{{ID: "nation-a", SciencePercent: &zeroScience}, {ID: "nation-b", SciencePercent: &zeroScience}},
		Players: []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}, {ID: "player-b", NationID: "nation-b", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Cities:  []City{{ID: "converted", OwnerID: "player-b", TileID: "tile-b", ConvertedToPlayerID: "player-a", ConvertedBy: "cleric", Yields: Yield{Gold: 10}}},
	}
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{ClericConversionPercent: 20, TelevangelistConversionPercent: 40}}

	advanceTreasury(&state, runtime)
	if state.Nations[0].Gold != 2 || state.Nations[1].Gold != 8 {
		t.Fatalf("conversion gold nations = %+v", state.Nations)
	}
}
