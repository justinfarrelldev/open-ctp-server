package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestGovernmentRuntimeYieldPercentEffects(t *testing.T) {
	nation := Nation{ID: "nation-a", Effects: []Effect{{ID: "manual", Kind: EffectYieldFood, Value: 1}}}
	upsertGovernmentEffects(&nation, rules.RuntimeGovernmentDefinition{ID: "GOVERNMENT_RETAIL", ResearchPercent: 100, FoodPercent: 25, ProductionPercent: 15, GoldPercent: 25, PollutionPercent: -20, PollutionUnhappyCoef: 0.04, CrimeCoef: 1, CrimeOffset: 95, TooManyCitiesThreshold: 2, TooManyCitiesCoefficient: 1.5, EmpireDistanceScale: 0.001, MinEmpireDistance: 500, MaxEmpireDistance: 2000, PositiveWorkdayCoef: 1, NegativeWorkdayCoef: 1, PositiveWagesCoef: 1, NegativeWagesCoef: 1, PositiveRationsCoef: 1, NegativeRationsCoef: 1})
	if !hasEffect(nation.Effects, EffectResearchPercent) || !hasEffect(nation.Effects, EffectYieldFoodPercent) || !hasEffect(nation.Effects, EffectYieldProductionPercent) || !hasEffect(nation.Effects, EffectYieldGoldPercent) || !hasEffect(nation.Effects, EffectPollutionPercent) || !hasEffect(nation.Effects, EffectPollutionUnhappyCoef) || !hasEffect(nation.Effects, EffectCrimeCoef) || !hasEffect(nation.Effects, EffectCrimeOffset) || !hasEffect(nation.Effects, EffectTooManyCitiesThreshold) || !hasEffect(nation.Effects, EffectTooManyCitiesCoefficient) || !hasEffect(nation.Effects, EffectEmpireDistanceScale) || !hasEffect(nation.Effects, EffectMinEmpireDistance) || !hasEffect(nation.Effects, EffectMaxEmpireDistance) || !hasEffect(nation.Effects, EffectPositiveWorkdayCoef) || !hasEffect(nation.Effects, EffectNegativeWagesCoef) || !hasEffect(nation.Effects, EffectPositiveRationsCoef) || !hasEffect(nation.Effects, EffectYieldFood) {
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
