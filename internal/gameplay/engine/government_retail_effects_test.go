package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestReadinessSupportDisbandsHighestCostUnitsOverProductionBudget(t *testing.T) {
	runtime := &rules.RuntimeRules{Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_TEST", SupportCoef: 1, ReadyWarCoef: 1}}}
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", GovernmentID: "GOVERNMENT_TEST"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Production: 2}}},
		Armies:  []Army{{ID: "army-a", OwnerID: "player-a", UnitIDs: []string{"unit-a", "unit-b"}}},
		Units: []Unit{
			{ID: "unit-a", OwnerID: "player-a", ShieldHunger: 5, ArmyID: "army-a"},
			{ID: "unit-b", OwnerID: "player-a", ShieldHunger: 1, ArmyID: "army-a"},
		},
	}
	applyReadinessSupportTax(&state, runtime)
	if len(state.Units) != 1 || state.Units[0].ID != "unit-b" {
		t.Fatalf("surviving units = %#v, want only unit-b", state.Units)
	}
	if got, want := state.Cities[0].Yields.Production, 1; got != want {
		t.Fatalf("city production after surviving support = %d, want %d", got, want)
	}
}

func TestReadinessSupportIgnoresProfessionalUnits(t *testing.T) {
	runtime := &rules.RuntimeRules{Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_TEST", SupportCoef: 1, ReadyWarCoef: 1}}}
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", GovernmentID: "GOVERNMENT_TEST", Readiness: "war"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", Yields: Yield{Production: 4}}},
		Units:   []Unit{{ID: "unit-a", OwnerID: "player-a", ShieldHunger: 3, IsProfessional: true}},
	}
	applyReadinessSupportTax(&state, runtime)
	if got, want := state.Cities[0].Yields.Production, 4; got != want {
		t.Fatalf("production = %d, want professional unit to be support-free with %d", got, want)
	}
}

func TestReadinessSupportReducesCityProduction(t *testing.T) {
	runtime := &rules.RuntimeRules{Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_TEST", SupportCoef: 1, ReadyWarCoef: 1}}}
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", GovernmentID: "GOVERNMENT_TEST"}},
		Cities: []City{
			{ID: "city-a", OwnerID: "player-a", Yields: Yield{Production: 10}},
			{ID: "city-b", OwnerID: "player-a", Yields: Yield{Production: 5}},
		},
		Units: []Unit{
			{ID: "unit-a", OwnerID: "player-a", ShieldHunger: 3},
			{ID: "unit-b", OwnerID: "player-a", ShieldHunger: 50, NeedsNoSupport: true},
		},
	}
	applyReadinessSupportTax(&state, runtime)
	if got, want := state.Cities[0].Yields.Production, 8; got != want {
		t.Fatalf("city-a production = %d, want %d", got, want)
	}
	if got, want := state.Cities[1].Yields.Production, 4; got != want {
		t.Fatalf("city-b production = %d, want %d", got, want)
	}
}

func TestSetReadinessCommandChangesNationReadiness(t *testing.T) {
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations = []Nation{{ID: "nation-a", Readiness: "peace"}}
	body, _ := json.Marshal(map[string]string{"readiness": "war"})
	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "ready", ActorPlayerID: "player-a", Type: CommandSetReadiness, Payload: body}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got, want := next.Nations[0].Readiness, "war"; got != want {
		t.Fatalf("readiness = %q, want %q", got, want)
	}
	if len(events) != 1 || events[0].Type != EventReadinessChanged {
		t.Fatalf("events = %#v, want readiness changed", events)
	}
}

func TestSetReadinessRejectsUnsupportedLevel(t *testing.T) {
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations = []Nation{{ID: "nation-a"}}
	_, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bad-ready", ActorPlayerID: "player-a", Type: CommandSetReadiness, Payload: json.RawMessage(`{"readiness":"panic"}`)}, state.Rules, fixtureSources())
	if err == nil {
		t.Fatal("unsupported readiness should fail")
	}
}

func TestSocialPolicyChangesProductionGoldAndFoodConsumption(t *testing.T) {
	workday, wages, rations := -1, 1, 2
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a"}}
	state.Nations = []Nation{{ID: "nation-a", Workday: &workday, Wages: &wages, Rations: &rations}}
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2, Effects: []Effect{{ID: "production", Kind: EffectYieldProduction, Value: 97}, {ID: "gold", Kind: EffectYieldGold, Value: 19}}}}
	yield, err := cityYield(state, state.Cities[0])
	if err != nil {
		t.Fatal(err)
	}
	if got, want := yield.Production, 115; got != want {
		t.Fatalf("workday production = %d, want %d", got, want)
	}
	if got, want := yield.Gold, 10; got != want {
		t.Fatalf("wage-adjusted gold = %d, want %d", got, want)
	}
	if got, want := cityFoodConsumptionForNation(state, state.Cities[0]), 6; got != want {
		t.Fatalf("rations food consumption = %d, want %d", got, want)
	}
}

func TestGovernmentRuntimeYieldPercentEffects(t *testing.T) {
	nation := Nation{ID: "nation-a", Effects: []Effect{{ID: "manual", Kind: EffectYieldFood, Value: 1}}}
	upsertGovernmentEffects(&nation, rules.RuntimeGovernmentDefinition{ID: "GOVERNMENT_RETAIL", ResearchPercent: 100, FoodPercent: 25, ProductionPercent: 15, GoldPercent: 25, PollutionPercent: -20, PollutionUnhappyCoef: 0.04, CrimeCoef: 1, CrimeOffset: 95, TooManyCitiesThreshold: 2, TooManyCitiesCoefficient: 1.5, EmpireDistanceScale: 0.001, MinEmpireDistance: 500, MaxEmpireDistance: 2000, AtHomeRadius: 2, OverseasCoef: 0.5, PositiveWorkdayCoef: 1, NegativeWorkdayCoef: 1, PositiveWagesCoef: 1, NegativeWagesCoef: 1, PositiveRationsCoef: 1, NegativeRationsCoef: 1})
	if !hasEffect(nation.Effects, EffectResearchPercent) || !hasEffect(nation.Effects, EffectYieldFoodPercent) || !hasEffect(nation.Effects, EffectYieldProductionPercent) || !hasEffect(nation.Effects, EffectYieldGoldPercent) || !hasEffect(nation.Effects, EffectPollutionPercent) || !hasEffect(nation.Effects, EffectPollutionUnhappyCoef) || !hasEffect(nation.Effects, EffectCrimeCoef) || !hasEffect(nation.Effects, EffectCrimeOffset) || !hasEffect(nation.Effects, EffectTooManyCitiesThreshold) || !hasEffect(nation.Effects, EffectTooManyCitiesCoefficient) || !hasEffect(nation.Effects, EffectEmpireDistanceScale) || !hasEffect(nation.Effects, EffectMinEmpireDistance) || !hasEffect(nation.Effects, EffectMaxEmpireDistance) || !hasEffect(nation.Effects, EffectAtHomeRadius) || !hasEffect(nation.Effects, EffectOverseasCoef) || !hasEffect(nation.Effects, EffectPositiveWorkdayCoef) || !hasEffect(nation.Effects, EffectNegativeWagesCoef) || !hasEffect(nation.Effects, EffectPositiveRationsCoef) || !hasEffect(nation.Effects, EffectYieldFood) {
		t.Fatalf("government effects = %#v", nation.Effects)
	}
}

func TestRetailSetSocialPolicyCommandPersistsSliders(t *testing.T) {
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations = []Nation{{ID: "nation-a"}}
	body, _ := json.Marshal(map[string]int{"workday": 1, "wages": -1, "rations": 0})
	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "social", ActorPlayerID: "player-a", Type: CommandSetSocialPolicy, Payload: body}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if next.Nations[0].Workday == nil || *next.Nations[0].Workday != 1 || next.Nations[0].Wages == nil || *next.Nations[0].Wages != -1 || next.Nations[0].Rations == nil || *next.Nations[0].Rations != 0 || len(events) != 1 || events[0].Type != EventSocialPolicyChanged {
		t.Fatalf("social policy command failed: nation=%+v events=%+v", next.Nations[0], events)
	}
}

func TestRetailPublicWorksRateDivertsProductionToMaterials(t *testing.T) {
	state := fixtureState()
	state.Nations = []Nation{{ID: "nation-a"}, {ID: "nation-b"}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Effects: []Effect{{ID: "workshop", Kind: EffectYieldProduction, Value: 7}}, Queue: []ProductionItem{{ID: "UNIT_WARRIOR", Kind: ProductionUnit, Cost: 100}}}}
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	body := json.RawMessage(`{"public_works_percent":30}`)
	withRate, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "public-works-rate", ActorPlayerID: "player-a", Type: CommandSetPublicWorksRate, Payload: body}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("set public works rate: %v", err)
	}
	if withRate.Nations[0].PublicWorksPercent == nil || *withRate.Nations[0].PublicWorksPercent != 30 || len(events) != 1 || events[0].Type != EventPublicWorksRateChanged {
		t.Fatalf("public works rate command = nation:%+v events:%+v", withRate.Nations[0], events)
	}
	next := runtimeApply(t, withRate, "player-a", CommandRoundComplete, struct{}{})
	if next.Nations[0].PublicWorks != 3 || next.Cities[0].StoredProduction != 7 {
		t.Fatalf("public works/production = %d/%d, want 3/7", next.Nations[0].PublicWorks, next.Cities[0].StoredProduction)
	}
}

func TestRetailSetScienceRateCommandCapsAtGovernmentMaximum(t *testing.T) {
	state := fixtureState()
	state.Nations = []Nation{{ID: "nation-a", Effects: []Effect{{ID: "max-science", Kind: EffectMaxSciencePercent, Value: 60}}}, {ID: "nation-b"}}
	body := json.RawMessage(`{"science_percent":100}`)
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "science-max", ActorPlayerID: "player-a", Type: CommandSetScienceRate, Payload: body}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("set science rate: %v", err)
	}
	if next.Nations[0].SciencePercent == nil || *next.Nations[0].SciencePercent != 60 {
		t.Fatalf("science percent = %v, want capped government max 60", next.Nations[0].SciencePercent)
	}
}

func TestRetailSetScienceRateCommandPersistsExplicitZero(t *testing.T) {
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations = []Nation{{ID: "nation-a"}}
	body, _ := json.Marshal(map[string]int{"science_percent": 0})
	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "science-zero", ActorPlayerID: "player-a", Type: CommandSetScienceRate, Payload: body}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if next.Nations[0].SciencePercent == nil || *next.Nations[0].SciencePercent != 0 || len(events) != 1 || events[0].Type != EventScienceRateChanged {
		t.Fatalf("science rate command failed: nation=%+v events=%+v", next.Nations[0], events)
	}
}

func TestRetailReadinessHPModifierChangesCombatDamage(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-b", NationID: "nation-b"}},
		Nations: []Nation{{ID: "nation-b", Readiness: "peace", Effects: []Effect{{ID: "peace-hp", Kind: EffectReadyPeaceHP, Value: 50}}}},
	}
	target := Unit{ID: "target", OwnerID: "player-b"}
	if got, want := readinessAdjustedCombatDamage(state, target, 1), 2; got != want {
		t.Fatalf("readiness-adjusted damage = %d, want %d", got, want)
	}
}

func TestPeaceMovementDefeatStateAdjustsHappiness(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Nations: []Nation{{ID: "nation-a", PeaceMovement: &PeaceMovementState{RawOverseasDefeat: 250, RawHomeDefeat: 100}}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 73; got != want {
		t.Fatalf("happiness with overseas/home defeat = %d, want %d", got, want)
	}
}

func TestCombatLossesClassifyPeaceMovementHomeAndOverseas(t *testing.T) {
	state := State{
		Tiles:   []Tile{{ID: "home", X: 0, Y: 0}, {ID: "front", X: 4, Y: 0}},
		Players: []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "radius", Kind: EffectAtHomeRadius, Value: 1}}}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", TileID: "home"}},
	}
	recordUnitLossForPeaceMovement(&state, Unit{ID: "home-loss", OwnerID: "player-a", TileID: "home"})
	recordUnitLossForPeaceMovement(&state, Unit{ID: "overseas-loss", OwnerID: "player-a", TileID: "front"})
	got := state.Nations[0].PeaceMovement
	if got == nil || got.UnitsLostHome != 1 || got.UnitsLostOverseas != 1 {
		t.Fatalf("loss counters = %#v, want one home and one overseas", got)
	}
}

func TestAdvancePeaceMovementAppliesDefeatDecayAndResetsLossCounters(t *testing.T) {
	state := State{
		Nations: []Nation{{ID: "nation-a", GovernmentID: "GOVERNMENT_TEST", PeaceMovement: &PeaceMovementState{RawOverseasDefeat: 100, RawHomeDefeat: 100, UnitsLostOverseas: 2, UnitsLostHome: 1}}},
	}
	runtime := &rules.RuntimeRules{Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_TEST", OverseasDefeatDecay: 0.25, OverseasDefeatCoef: 1, HomeDefeatDecay: 0.5, HomeDefeatCoef: 1}}}
	advancePeaceMovement(&state, runtime)
	got := state.Nations[0].PeaceMovement
	if got == nil || got.RawOverseasDefeat != 275 || got.RawHomeDefeat != 150 || got.UnitsLostOverseas != 0 || got.UnitsLostHome != 0 {
		t.Fatalf("peace movement = %#v, want decayed defeat state with reset losses", got)
	}
}

func TestRetailCityCaptureClearsConversionWhenConvertedOwnerCaptures(t *testing.T) {
	state := richState()
	state.Cities = []City{{ID: "city", OwnerID: "player-b", TileID: "city-tile", ConvertedToPlayerID: "player-a", ConvertedBy: "cleric"}}
	state.Tiles = append(state.Tiles, Tile{ID: "city-tile", X: 2, Y: 2, Terrain: TerrainPlains, OwnerID: "player-b", Open: true})

	if got := captureCityAt(&state, "player-a", "city-tile"); got != "city" {
		t.Fatalf("captureCityAt() = %q", got)
	}
	if state.Cities[0].ConvertedToPlayerID != "" || state.Cities[0].ConvertedBy != "" {
		t.Fatalf("captured city conversion was not cleared: %+v", state.Cities[0])
	}
}

func TestRetailConquestDistressAppliesToCapturedCityAndDecays(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "distress", Kind: EffectConquestDistress, Value: -5}, {ID: "decay", Kind: EffectConquestDistressDecay, Value: 20}}}, {ID: "nation-b"}},
		Cities:  []City{{ID: "city", OwnerID: "player-b", TileID: "city-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}}},
	}
	if got := captureCityAt(&state, "player-a", "city-tile"); got != "city" {
		t.Fatalf("captured city = %q", got)
	}
	if got, want := state.Cities[0].ConquestDistress, -5; got != want {
		t.Fatalf("conquest distress = %d, want %d", got, want)
	}
	if got, want := cityHappiness(state, state.Cities[0]), 75; got != want {
		t.Fatalf("happiness = %d, want conquest distress penalty to %d", got, want)
	}
	advanceConquestDistress(&state.Cities[0], state.Nations[0])
	if got, want := state.Cities[0].ConquestDistress, -4; got != want {
		t.Fatalf("decayed conquest distress = %d, want %d", got, want)
	}
}

func TestRetailCityCaptureRemovesTradeRoutesTouchingCapturedCity(t *testing.T) {
	state := State{
		Cities: []City{{ID: "captured", OwnerID: "player-b", TileID: "city-tile"}, {ID: "ally", OwnerID: "player-a", TileID: "ally-tile"}, {ID: "third", OwnerID: "player-c", TileID: "third-tile"}},
		TradeRoutes: []TradeRoute{
			{ID: "captured-ally", SourceCityID: "captured", DestinationCityID: "ally"},
			{ID: "third-ally", SourceCityID: "third", DestinationCityID: "ally"},
		},
	}
	if got := captureCityAt(&state, "player-a", "city-tile"); got != "captured" {
		t.Fatalf("captured city = %q", got)
	}
	if len(state.TradeRoutes) != 1 || state.TradeRoutes[0].ID != "third-ally" {
		t.Fatalf("capture did not remove captured-city trade routes: %#v", state.TradeRoutes)
	}
}

func TestRetailWarDiscontentReducesCityHappiness(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "max", Kind: EffectWarDiscontentMaxUnits, Value: 1}, {ID: "per", Kind: EffectWarDiscontentPerUnit, Value: 100}}}},
		Tiles:   []Tile{{ID: "city-tile"}, {ID: "field-a"}, {ID: "field-b"}, {ID: "field-c"}},
		Cities:  []City{{ID: "city", OwnerID: "player-a", TileID: "city-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}}},
		Units: []Unit{
			{ID: "unit-a", OwnerID: "player-a", TileID: "field-a", Attack: 1},
			{ID: "unit-b", OwnerID: "player-a", TileID: "field-b", Attack: 1},
			{ID: "unit-c", OwnerID: "player-a", TileID: "field-c", Attack: 1},
		},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 78; got != want {
		t.Fatalf("happiness = %d, want war discontent penalty to %d", got, want)
	}
}

func TestRetailOverseasUnitsReduceHappinessOutsideAtHomeRadius(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "radius", Kind: EffectAtHomeRadius, Value: 2}, {ID: "overseas", Kind: EffectOverseasCoef, Value: 100}}}},
		Tiles:   []Tile{{ID: "city-tile", X: 0, Y: 0}, {ID: "distant", X: 4, Y: 0}},
		Cities:  []City{{ID: "city", OwnerID: "player-a", TileID: "city-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}}},
		Units:   []Unit{{ID: "unit-a", OwnerID: "player-a", TileID: "distant"}},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 78; got != want {
		t.Fatalf("happiness = %d, want overseas movement penalty to %d", got, want)
	}
}

func TestRetailLowerPeaceMovementReducesWarDiscontentPenalty(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "max", Kind: EffectWarDiscontentMaxUnits, Value: 1}, {ID: "per", Kind: EffectWarDiscontentPerUnit, Value: 100}}}},
		Tiles:   []Tile{{ID: "city-tile"}, {ID: "field-a"}, {ID: "field-b"}, {ID: "field-c"}},
		Cities:  []City{{ID: "city", OwnerID: "player-a", TileID: "city-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}, {ID: "movie-palace", Kind: EffectLowerPeaceMovementPercent, Value: -50}}}},
		Units: []Unit{
			{ID: "unit-a", OwnerID: "player-a", TileID: "field-a", Attack: 1},
			{ID: "unit-b", OwnerID: "player-a", TileID: "field-b", Attack: 1},
			{ID: "unit-c", OwnerID: "player-a", TileID: "field-c", Attack: 1},
		},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 79; got != want {
		t.Fatalf("happiness = %d, want lower peace movement penalty to %d", got, want)
	}
}

func TestRetailWarDiscontentExemptsOwnedFortifiedTiles(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "max", Kind: EffectWarDiscontentMaxUnits, Value: 1}, {ID: "per", Kind: EffectWarDiscontentPerUnit, Value: 100}}}},
		Tiles:   []Tile{{ID: "city-tile", OwnerID: "player-a"}, {ID: "fort-a", OwnerID: "player-a"}, {ID: "field-b"}, {ID: "field-c"}},
		Cities:  []City{{ID: "city", OwnerID: "player-a", TileID: "city-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}}},
		Overlays: []TileOverlay{
			{ID: "fort", TileID: "fort-a", Kind: "improvement", OwnerID: "player-a", ImprovementID: "TILEIMP_FORTIFICATIONS"},
		},
		Units: []Unit{
			{ID: "unit-a", OwnerID: "player-a", TileID: "fort-a", Attack: 1},
			{ID: "unit-b", OwnerID: "player-a", TileID: "field-b", Attack: 1},
			{ID: "unit-c", OwnerID: "player-a", TileID: "field-c", Attack: 1},
		},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 79; got != want {
		t.Fatalf("happiness = %d, want owned fort exemption penalty to %d", got, want)
	}
}

func TestRetailNoPollutionUnhappinessWonderSuppressesPenalty(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "pollution-unhappy", Kind: EffectPollutionUnhappyCoef, Value: 4}}}},
		Tiles:   []Tile{{ID: "city-tile", Pollution: 50}},
		Cities:  []City{{ID: "city", OwnerID: "player-a", TileID: "city-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}}},
		Wonders: []Wonder{{ID: "wonder", OwnerID: "player-a", CityID: "city", Effects: []Effect{{ID: "no-pollution", Kind: EffectNoPollutionUnhappiness}}}},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 80; got != want {
		t.Fatalf("happiness = %d, want no pollution unhappiness penalty = %d", got, want)
	}
}

func TestRetailPollutionUnhappyCoefReducesHappiness(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "pollution-unhappy", Kind: EffectPollutionUnhappyCoef, Value: 4}}}},
		Tiles:   []Tile{{ID: "city-tile", Pollution: 50}},
		Cities:  []City{{ID: "city", OwnerID: "player-a", TileID: "city-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}}},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 78; got != want {
		t.Fatalf("happiness = %d, want base minus 50*0.04 = %d", got, want)
	}
}

func TestRetailGovernmentCrimeCoefComputesCrimeFromHappiness(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "offset", Kind: EffectCrimeOffset, Value: 95}, {ID: "coef", Kind: EffectCrimeCoef, Value: 100}}}},
		Cities:  []City{{ID: "city", OwnerID: "player-a", Happiness: 80}},
	}
	if got, want := cityCrimePercent(state, state.Cities[0]), 15; got != want {
		t.Fatalf("crime = %d, want (95-80)*1 = %d", got, want)
	}
}

func TestRetailSocialSliderGovernmentCoefficientsAffectHappiness(t *testing.T) {
	workday := 2
	wages := -1
	rations := 1
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Workday: &workday, Wages: &wages, Rations: &rations, Effects: []Effect{
			{ID: "workday-exp", Kind: EffectWorkdayExpectation, Value: 0}, {ID: "workday-pos", Kind: EffectPositiveWorkdayCoef, Value: 100}, {ID: "workday-neg", Kind: EffectNegativeWorkdayCoef, Value: 100},
			{ID: "wages-exp", Kind: EffectWagesExpectation, Value: 0}, {ID: "wages-pos", Kind: EffectPositiveWagesCoef, Value: 100}, {ID: "wages-neg", Kind: EffectNegativeWagesCoef, Value: 200},
			{ID: "rations-exp", Kind: EffectRationsExpectation, Value: 0}, {ID: "rations-pos", Kind: EffectPositiveRationsCoef, Value: 50}, {ID: "rations-neg", Kind: EffectNegativeRationsCoef, Value: 100},
		}}},
		Cities: []City{{ID: "city", OwnerID: "player-a", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}}},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 81; got != want {
		t.Fatalf("happiness = %d, want base + workday 2 - wages 2 + rations 1 = %d", got, want)
	}
}

func TestRetailEmpireDistanceUsesTerrainPathCosts(t *testing.T) {
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}}, {ID: "TERRAIN_MOUNTAIN", Cost: 900, Domains: []string{DomainLand}}}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := State{
		Rules:   RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions},
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "scale", Kind: EffectEmpireDistanceScale, Value: 10000}, {ID: "max", Kind: EffectMaxEmpireDistance, Value: 5000}}}},
		Tiles: []Tile{
			{ID: "capital-tile", X: 0, Y: 0, Terrain: TerrainPlains},
			{ID: "mountain", X: 1, Y: 0, Terrain: "TERRAIN_MOUNTAIN"},
			{ID: "remote-tile", X: 2, Y: 0, Terrain: TerrainPlains},
			{ID: "south-a", X: 0, Y: 1, Terrain: TerrainPlains},
			{ID: "south-b", X: 1, Y: 1, Terrain: TerrainPlains},
			{ID: "south-c", X: 2, Y: 1, Terrain: TerrainPlains},
		},
		Cities: []City{
			{ID: "capital", OwnerID: "player-a", TileID: "capital-tile", Buildings: []Building{{ID: "palace", Effects: []Effect{{ID: "capital", Kind: EffectCapitol}}}}, Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}},
			{ID: "remote", OwnerID: "player-a", TileID: "remote-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}},
		},
	}
	if got, want := cityHappiness(state, state.Cities[1]), 75; got != want {
		t.Fatalf("happiness = %d, want path-cost empire-distance penalty to %d", got, want)
	}
}

func TestRetailEmpireDistanceGovernmentPenaltyReducesHappiness(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "scale", Kind: EffectEmpireDistanceScale, Value: 10000}, {ID: "max", Kind: EffectMaxEmpireDistance, Value: 1000}}}},
		Tiles:   []Tile{{ID: "capital-tile", X: 0, Y: 0}, {ID: "remote-tile", X: 10, Y: 0}},
		Cities: []City{
			{ID: "capital", OwnerID: "player-a", TileID: "capital-tile", Buildings: []Building{{ID: "palace", Effects: []Effect{{ID: "capital", Kind: EffectCapitol}}}}, Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}},
			{ID: "remote", OwnerID: "player-a", TileID: "remote-tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}},
		},
	}
	if got, want := cityHappiness(state, state.Cities[1]), 70; got != want {
		t.Fatalf("happiness = %d, want base minus capped distance*scale penalty = %d", got, want)
	}
}

func TestRetailTooManyCitiesGovernmentPenaltyReducesHappiness(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Effects: []Effect{{ID: "threshold", Kind: EffectTooManyCitiesThreshold, Value: 1}, {ID: "coefficient", Kind: EffectTooManyCitiesCoefficient, Value: 200}}}},
		Cities: []City{
			{ID: "a", OwnerID: "player-a", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}},
			{ID: "b", OwnerID: "player-a", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}},
			{ID: "c", OwnerID: "player-a", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 80}}},
		},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 76; got != want {
		t.Fatalf("happiness = %d, want base minus 2*(3-1) penalty = %d", got, want)
	}
}

func TestRetailMartialLawHappinessUsesGovernmentCapAndUnitMarkers(t *testing.T) {
	state := fixtureState()
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive}}
	state.Nations = []Nation{{ID: "nation-a", Effects: []Effect{{ID: "ml-max", Kind: EffectMartialLawMaxUnits, Value: 2}, {ID: "ml-effect", Kind: EffectMartialLawEffect, Value: 3}, {ID: "ml-threshold", Kind: EffectMartialLawThreshold, Value: 80}}}}
	state.Cities = []City{{ID: "city", OwnerID: "player-a", TileID: "tile", Effects: []Effect{{ID: "base", Kind: EffectHappiness, Value: 75}}}}
	state.Units = []Unit{
		{ID: "garrison-a", OwnerID: "player-a", TileID: "tile", Effects: []Effect{{ID: "martial", Kind: EffectExertsMartialLaw}}},
		{ID: "garrison-b", OwnerID: "player-a", TileID: "tile", Effects: []Effect{{ID: "martial", Kind: EffectExertsMartialLaw}}},
		{ID: "garrison-c", OwnerID: "player-a", TileID: "tile", Effects: []Effect{{ID: "martial", Kind: EffectExertsMartialLaw}}},
	}
	if got, want := cityHappiness(state, state.Cities[0]), 80; got != want {
		t.Fatalf("martial law happiness = %d, want capped %d", got, want)
	}
}

func TestSignedYieldPercentEffectsCanReduceCityYield(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Cities[0].Effects = []Effect{
		{ID: "a-food", Kind: EffectYieldFood, Value: 18},
		{ID: "a-production", Kind: EffectYieldProduction, Value: 17},
		{ID: "a-gold", Kind: EffectYieldGold, Value: 19},
		{ID: "z-food-government", Kind: EffectYieldFoodPercent, Value: -25},
		{ID: "z-production-government", Kind: EffectYieldProductionPercent, Value: -25},
		{ID: "z-gold-government", Kind: EffectYieldGoldPercent, Value: -25},
	}
	yield, err := cityYield(state, state.Cities[0])
	if err != nil {
		t.Fatal(err)
	}
	if yield.Food != 15 || yield.Production != 15 || yield.Gold != 15 {
		t.Fatalf("yield = %+v, want signed 25%% reductions with retail-style ceil", yield)
	}
}
