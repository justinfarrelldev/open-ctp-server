package ai

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestTask11PlannerUsesProjectionForResearchGovernmentDiplomacy(t *testing.T) {
	projection := visibility.Projection{
		Research:  &visibility.ResearchProjection{Available: []string{"ADVANCE_B", "ADVANCE_A"}, Advances: []string{"ADVANCE_SYNTHETIC_ROOT"}},
		Diplomacy: &visibility.DiplomacyProjection{Proposals: []visibility.DiplomacyProposalProjection{{ID: "p1", FromPlayerID: "other", ToPlayerID: "ai", Kind: "trade"}}, Relations: []visibility.DiplomacyRelationProjection{{TargetPlayerID: "other", Treaty: "peace"}, {TargetPlayerID: "third", Treaty: "peace"}}},
	}
	input, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	assertCandidate(t, candidates, engine.CommandSetResearch)
	assertCandidate(t, candidates, engine.CommandSetGovernment)
	assertCandidate(t, candidates, engine.CommandRespondAgreement)
	assertCandidate(t, candidates, engine.CommandProposeAgreement)
	if candidates[0].Type != engine.CommandSetResearch {
		t.Fatalf("top candidate = %#v, want research first", candidates[0])
	}
}

func TestTask11PlannerUsesProjectionForProductionTerrainTransportAndCombat(t *testing.T) {
	projection := visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "home", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}},
			{ID: "work", X: 1, Y: 0, Pollution: 1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}},
			{ID: "enemy", X: 0, Y: 1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}},
		},
		Units: []visibility.UnitProjection{
			{ID: "worker", OwnerID: "ai", TileID: "home", Movement: 1, Domain: engine.DomainLand, Effects: []engine.Effect{{ID: "terraform", Kind: engine.EffectSpecialTerraform}}},
			{ID: "cargo", OwnerID: "ai", TileID: "home", Movement: 1, Domain: engine.DomainLand},
			{ID: "ship", OwnerID: "ai", TileID: "home", Movement: 1, Domain: engine.DomainSea, CargoCapacity: 1, CargoDomains: []string{engine.DomainLand}},
			{ID: "attacker", OwnerID: "ai", TileID: "home", ArmyID: "army-a", Movement: 1, Domain: engine.DomainLand, Attack: 1},
			{ID: "defender", OwnerID: "foe", TileID: "enemy", ArmyID: "army-b", Domain: engine.DomainLand, Attack: 1},
		},
		Rules:     &visibility.RulesProjection{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_WARRIOR", Domain: engine.DomainLand, Movement: 1, Health: 1, ProductionCost: 1}}},
		Cities:    []visibility.CityProjection{{ID: "city", OwnerID: "ai", TileID: "home"}},
		Diplomacy: &visibility.DiplomacyProjection{PublicWorks: 1},
	}
	input, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	assertCandidate(t, candidates, engine.CommandQueueProduction)
	assertCandidate(t, candidates, engine.CommandBuildImprovement)
	assertCandidate(t, candidates, engine.CommandEmbark)
	assertCandidate(t, candidates, engine.CommandCombat)
}

func assertCandidate(t *testing.T, candidates []Candidate, kind string) {
	t.Helper()
	for _, candidate := range candidates {
		if candidate.Type == kind {
			return
		}
	}
	t.Fatalf("candidate %q missing from %#v", kind, candidates)
}
