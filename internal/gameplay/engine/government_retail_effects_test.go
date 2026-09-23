package engine

import (
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
