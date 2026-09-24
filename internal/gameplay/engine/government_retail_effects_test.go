package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestGovernmentRuntimeYieldPercentEffects(t *testing.T) {
	nation := Nation{ID: "nation-a", Effects: []Effect{{ID: "manual", Kind: EffectYieldFood, Value: 1}}}
	upsertGovernmentEffects(&nation, rules.RuntimeGovernmentDefinition{ID: "GOVERNMENT_RETAIL", ResearchPercent: 100, FoodPercent: 25, ProductionPercent: 15, GoldPercent: 25, PollutionPercent: -20})
	if !hasEffect(nation.Effects, EffectResearchPercent) || !hasEffect(nation.Effects, EffectYieldFoodPercent) || !hasEffect(nation.Effects, EffectYieldProductionPercent) || !hasEffect(nation.Effects, EffectYieldGoldPercent) || !hasEffect(nation.Effects, EffectPollutionPercent) || !hasEffect(nation.Effects, EffectYieldFood) {
		t.Fatalf("government effects = %#v", nation.Effects)
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
