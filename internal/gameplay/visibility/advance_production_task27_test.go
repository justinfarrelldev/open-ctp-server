package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestAdvanceProductionActionsRequireProjectedAdvance(t *testing.T) {
	projection := Projection{
		Cities: []CityProjection{{ID: "city-a", OwnerID: "player-a"}},
		Rules: &RulesProjection{Production: []rules.RuntimeProduction{
			{ID: "CAPITALIZATION", Kind: engine.ProductionCapitalization, Cost: 1, Prerequisites: []string{"ADVANCE_GLOBAL_ECONOMICS"}},
			{ID: "INFRASTRUCTURE", Kind: engine.ProductionInfrastructure, Cost: 1, Prerequisites: []string{"ADVANCE_MASS_PRODUCTION"}},
		}},
		Research: &ResearchProjection{Advances: []string{"ADVANCE_GLOBAL_ECONOMICS"}},
	}

	actions := canonicalProductionActions(t, projection, "player-a")
	if !hasProductionAction(actions, "CAPITALIZATION") {
		t.Fatalf("capitalization action missing after enabling advance: %#v", actions)
	}
	if hasProductionAction(actions, "INFRASTRUCTURE") {
		t.Fatalf("infrastructure action projected before enabling advance: %#v", actions)
	}
	payload := productionActionPayload(t, actions, "CAPITALIZATION")
	if payload.Item.Kind != engine.ProductionCapitalization || payload.Item.Prerequisite != "ADVANCE_GLOBAL_ECONOMICS" {
		t.Fatalf("capitalization payload = %+v", payload.Item)
	}
}

func productionActionPayload(t *testing.T, actions []ActionProjection, targetID string) struct {
	CityID string                `json:"city_id"`
	Item   engine.ProductionItem `json:"item"`
} {
	t.Helper()
	for _, action := range actions {
		if action.Command != engine.CommandQueueProduction || action.TargetID != targetID {
			continue
		}
		var payload struct {
			CityID string                `json:"city_id"`
			Item   engine.ProductionItem `json:"item"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		return payload
	}
	t.Fatalf("production action %q not found in %#v", targetID, actions)
	return struct {
		CityID string                `json:"city_id"`
		Item   engine.ProductionItem `json:"item"`
	}{}
}
