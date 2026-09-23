package ai

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestPlannerUsesProjectedAdvanceCostsForResearchChoice(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Research: &visibility.ResearchProjection{Available: []string{"ADVANCE_EXPENSIVE", "ADVANCE_CHEAP"}},
		Rules: &visibility.RulesProjection{Advances: []rules.RuntimeAdvanceDefinition{
			{ID: "ADVANCE_EXPENSIVE", Cost: 900},
			{ID: "ADVANCE_CHEAP", Cost: 100},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandSetResearch || candidates[0].TargetID != "ADVANCE_CHEAP" {
		t.Fatalf("research candidates = %#v, want cheapest projected advance first", candidates)
	}
}

func TestPlannerUsesProjectedGovernmentUnlocksAndBonuses(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Research:  &visibility.ResearchProjection{Advances: []string{"ADVANCE_UNLOCK"}},
		Diplomacy: &visibility.DiplomacyProjection{},
		Rules: &visibility.RulesProjection{Governments: []rules.RuntimeGovernmentDefinition{
			{ID: "GOVERNMENT_LOW", ResearchBonus: 1},
			{ID: "GOVERNMENT_HIGH", EnableAdvanceID: "ADVANCE_UNLOCK", ResearchBonus: 5, PublicWorksBonus: 2},
			{ID: "GOVERNMENT_LOCKED", EnableAdvanceID: "ADVANCE_LOCKED", ResearchBonus: 99},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandSetGovernment || candidates[0].TargetID != "GOVERNMENT_HIGH" {
		t.Fatalf("government candidates = %#v, want best unlocked projected government first", candidates)
	}
}

func TestPlannerQueuesProjectedRetailLikeUnitProduction(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Cities:   []visibility.CityProjection{{ID: "city", OwnerID: "ai", TileID: "tile"}},
		Research: &visibility.ResearchProjection{Advances: []string{"ADVANCE_BRONZE"}},
		Rules: &visibility.RulesProjection{Units: []rules.RuntimeUnitDefinition{
			{ID: "UNIT_SCOUT", Domain: engine.DomainLand, Movement: 2, Attack: 1, Defense: 1, Health: 10, ProductionCost: 20},
			{ID: "UNIT_LEGION", Domain: engine.DomainLand, Movement: 1, Attack: 5, Defense: 4, Health: 10, ProductionCost: 30, Prerequisites: []string{"ADVANCE_BRONZE"}},
			{ID: "UNIT_TANK", Domain: engine.DomainLand, Movement: 3, Attack: 20, Defense: 10, Health: 10, ProductionCost: 50, Prerequisites: []string{"ADVANCE_COMBUSTION"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if candidate.Type != engine.CommandQueueProduction {
			continue
		}
		var payload struct {
			Item engine.ProductionItem `json:"item"`
		}
		if err := json.Unmarshal(candidate.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Item.ID != "UNIT_LEGION" || payload.Item.Cost != 30 || payload.Item.Kind != engine.ProductionUnit {
			t.Fatalf("production item = %+v, want unlocked strongest legion", payload.Item)
		}
		return
	}
	t.Fatalf("missing production candidate: %#v", candidates)
}
