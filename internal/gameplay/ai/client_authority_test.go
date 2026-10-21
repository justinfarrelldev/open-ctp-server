package ai

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestClientAIProductionNeverInventsCatalogContent(t *testing.T) {
	projection := visibility.Projection{Cities: []visibility.CityProjection{{ID: "home", OwnerID: "a", Population: 1}}, Rules: &visibility.RulesProjection{}}
	if candidates := productionCandidates(projection, "a"); len(candidates) != 0 {
		t.Fatalf("AI invented production instead of waiting for legal content: %+v", candidates)
	}
}

func TestClientAIUsesAtomicQueueInsertForNewlyUnlockedEmergencyDefender(t *testing.T) {
	projection := visibility.Projection{
		Cities:       []visibility.CityProjection{{ID: "home", OwnerID: "a", Population: 2, Queue: []engine.ProductionItem{{ID: "UNIT_SETTLER", EntryID: "home:queue:1", Kind: engine.ProductionUnit, Cost: 100}}}},
		Rules:        &visibility.RulesProjection{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Effects: []rules.RuntimeStarterEffect{{Kind: engine.EffectSettler}}}, {ID: "warrior", Domain: engine.DomainLand, Attack: 1, Defense: 1, ProductionCost: 10}}},
		LegalActions: []visibility.ActionProjection{{Command: engine.CommandEditProduction, Route: "edit-production", EntityID: "home"}},
	}
	raw, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(raw, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if candidate.Type != engine.CommandEditProduction {
			continue
		}
		var payload engine.EditProductionPayload
		if err := json.Unmarshal(candidate.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Entries) != 2 || payload.Entries[0].ItemID != "warrior" || payload.Entries[1].EntryID != "home:queue:1" {
			t.Fatalf("emergency insert lost original order: %+v", payload)
		}
		return
	}
	t.Fatal("newly available defender did not interrupt expensive settler queue")
}

func TestClientUndefendedAIResearchesAnAvailableCheapDefenderUnlock(t *testing.T) {
	projection := visibility.Projection{
		Cities:   []visibility.CityProjection{{ID: "home", OwnerID: "a", Population: 1}},
		Research: &visibility.ResearchProjection{Available: []string{"religion", "tools"}},
		Rules: &visibility.RulesProjection{
			Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Effects: []rules.RuntimeStarterEffect{{Kind: engine.EffectSettler}}}, {ID: "warrior", Domain: engine.DomainLand, Attack: 1, Defense: 1, ProductionCost: 10, Prerequisites: []string{"tools"}}},
			AI:    &rules.RuntimeAI{AdvanceLists: []rules.RuntimeAIList{{Items: []string{"religion"}}}},
		},
	}
	candidates := researchCandidates(projection, "a")
	if len(candidates) != 1 || candidates[0].TargetID != "tools" {
		t.Fatalf("undefended city research ignores available defender: %+v", candidates)
	}
	projection.Units = []visibility.UnitProjection{{ID: "defender", OwnerID: "a"}}
	candidates = researchCandidates(projection, "a")
	if len(candidates) != 1 || candidates[0].TargetID != "religion" {
		t.Fatalf("established military should retain strategic research policy: %+v", candidates)
	}
}
