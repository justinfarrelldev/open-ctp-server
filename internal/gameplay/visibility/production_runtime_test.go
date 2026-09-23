package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

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
