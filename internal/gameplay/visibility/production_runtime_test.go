package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRulesProjectionIncludesRetailAILists(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Production:           []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: engine.ProductionUnit}},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: engine.DomainLand, Movement: 1, Health: 1}},
		Starter:              rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1},
		AI:                   &rules.RuntimeAI{UnitBuildLists: []rules.RuntimeAIList{{ID: "UNIT_BUILD_LIST_LAND_SETTLER", Items: []string{"UNIT_SETTLER"}}}},
	}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	projection := rulesProjection(engine.State{Rules: engine.RulesPin{Definitions: definitions, ContentHash: contentHash}})
	if projection == nil || projection.AI == nil || len(projection.AI.UnitBuildLists) != 1 || projection.AI.UnitBuildLists[0].Items[0] != "UNIT_SETTLER" {
		t.Fatalf("rules projection AI = %+v", projection)
	}
}

func TestProductionActionsIncludeUnlockedRuntimeBuildings(t *testing.T) {
	projection := Projection{
		Cities:   []CityProjection{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}},
		Research: &ResearchProjection{Advances: []string{"ADVANCE_AGRICULTURE"}},
		Rules: &RulesProjection{Production: []rules.RuntimeProduction{
			{ID: "IMPROVE_GRANARY", Kind: engine.ProductionBuilding, Cost: 300, Prerequisites: []string{"ADVANCE_AGRICULTURE"}, Effects: []rules.RuntimeStarterEffect{{ID: "granary-food", Kind: engine.EffectYieldFoodPercent, Value: 10}}},
		}},
	}

	actions := productionActions(projection, "player-a")
	for _, action := range actions {
		if action.Command != engine.CommandQueueProduction || action.TargetID != "IMPROVE_GRANARY" {
			continue
		}
		var payload struct {
			Item engine.ProductionItem `json:"item"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Item.Kind != engine.ProductionBuilding || payload.Item.Cost != 300 || len(payload.Item.Effects) != 1 {
			t.Fatalf("building action item = %+v", payload.Item)
		}
		return
	}
	t.Fatalf("missing runtime building production action: %#v", actions)
}
