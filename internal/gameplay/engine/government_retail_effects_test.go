package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestGovernmentRuntimeYieldPercentEffects(t *testing.T) {
	nation := Nation{ID: "nation-a", Effects: []Effect{{ID: "manual", Kind: EffectYieldFood, Value: 1}}}
	upsertGovernmentEffects(&nation, rules.RuntimeGovernmentDefinition{ID: "GOVERNMENT_RETAIL", ResearchBonus: 1, ProductionPercent: 15, GoldPercent: 25})
	if !hasEffect(nation.Effects, EffectResearch) || !hasEffect(nation.Effects, EffectYieldProductionPercent) || !hasEffect(nation.Effects, EffectYieldGoldPercent) || !hasEffect(nation.Effects, EffectYieldFood) {
		t.Fatalf("government effects = %#v", nation.Effects)
	}
}
