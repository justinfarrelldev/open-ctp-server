package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestAdvanceInfrastructureProductionConvertsShieldsToPublicWorks(t *testing.T) {
	state := advanceProductionState(t, rules.RuntimeGovernmentDefinition{ID: "GOVERNMENT_TEST", InfrastructureCoefficient: 0.5})
	state.Nations[0].Advances = []string{"ADVANCE_MASS_PRODUCTION"}
	state.Cities[0].Queue = []ProductionItem{{ID: "INFRASTRUCTURE", Kind: ProductionInfrastructure, Cost: 1, Prerequisite: "ADVANCE_MASS_PRODUCTION"}}

	next := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if next.Nations[0].PublicWorks != 5 {
		t.Fatalf("public works after infrastructure = %d, want ceil(9*0.5)=5", next.Nations[0].PublicWorks)
	}
	if next.Cities[0].StoredProduction != 0 || len(next.Cities[0].Queue) != 1 || next.Cities[0].Queue[0].Kind != ProductionInfrastructure {
		t.Fatalf("infrastructure queue/storage = queue %#v stored %d, want persistent queue with zero stored production", next.Cities[0].Queue, next.Cities[0].StoredProduction)
	}
}

func TestAdvanceCapitalizationProductionConvertsShieldsToGold(t *testing.T) {
	state := advanceProductionState(t, rules.RuntimeGovernmentDefinition{ID: "GOVERNMENT_TEST", CapitalizationCoefficient: 0.25})
	state.Nations[0].Advances = []string{"ADVANCE_GLOBAL_ECONOMICS"}
	state.Cities[0].Queue = []ProductionItem{{ID: "CAPITALIZATION", Kind: ProductionCapitalization, Cost: 1, Prerequisite: "ADVANCE_GLOBAL_ECONOMICS"}}

	next := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if next.Nations[0].Gold != 3 {
		t.Fatalf("gold after capitalization = %d, want ceil(9*0.25)=3", next.Nations[0].Gold)
	}
	if next.Cities[0].StoredProduction != 0 || len(next.Cities[0].Queue) != 1 || next.Cities[0].Queue[0].Kind != ProductionCapitalization {
		t.Fatalf("capitalization queue/storage = queue %#v stored %d, want persistent queue with zero stored production", next.Cities[0].Queue, next.Cities[0].StoredProduction)
	}
}

func TestAdvanceProductionRequiresUnlockedAdvance(t *testing.T) {
	state := advanceProductionState(t, rules.RuntimeGovernmentDefinition{ID: "GOVERNMENT_TEST", InfrastructureCoefficient: 0.5})
	err := runtimeApplyExpectError(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "INFRASTRUCTURE", Kind: ProductionInfrastructure, Cost: 1}})
	if err == nil {
		t.Fatal("queueing infrastructure before its enabling advance succeeded")
	}
}

func advanceProductionState(t *testing.T, government rules.RuntimeGovernmentDefinition) State {
	t.Helper()
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Movement:             rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityProduction: 9}}},
		Advances: []rules.RuntimeAdvanceDefinition{
			{ID: "ADVANCE_GLOBAL_ECONOMICS", Cost: 1, EnablesCapitalization: true},
			{ID: "ADVANCE_MASS_PRODUCTION", Cost: 1, EnablesInfrastructure: true},
		},
		Production: []rules.RuntimeProduction{
			{ID: "CAPITALIZATION", Kind: ProductionCapitalization, Cost: 1, Prerequisites: []string{"ADVANCE_GLOBAL_ECONOMICS"}},
			{ID: "INFRASTRUCTURE", Kind: ProductionInfrastructure, Cost: 1, Prerequisites: []string{"ADVANCE_MASS_PRODUCTION"}},
			{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1},
		},
		Governments: []rules.RuntimeGovernmentDefinition{government},
		Units:       []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
	}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := runtimeStartedState(t)
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}
	state.Nations[0].GovernmentID = government.ID
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 1}}
	return state
}
