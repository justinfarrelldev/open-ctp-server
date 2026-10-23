package ai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func idleProductionProjection() visibility.Projection {
	item := engine.ProductionItem{ID: "UNIT_WARRIOR", Kind: engine.ProductionUnit, Cost: 20}
	return visibility.Projection{Revision: 4, Cities: []visibility.CityProjection{{ID: "home", OwnerID: "ai", Population: 1, ProductionOptions: &engine.ProductionOptions{Status: "complete", Items: []engine.ProductionItem{item}}}}, Rules: &visibility.RulesProjection{Units: []rules.RuntimeUnitDefinition{{ID: item.ID, Domain: engine.DomainLand, Movement: 1, Health: 10, Attack: 5, Defense: 2, ProductionCost: 20}}}}
}

func TestIdleProductionBeatsCloseoutWithActualOneActionBudget(t *testing.T) {
	for _, planning := range []string{"", visibility.PlanningStatusOpen} {
		t.Run("planning="+planning, func(t *testing.T) {
			p := idleProductionProjection()
			p.PlanningStatus = planning
			raw, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			candidates, err := Plan(raw, "ai")
			if err != nil {
				t.Fatal(err)
			}
			if candidates[0].Type != engine.CommandQueueProduction {
				t.Fatalf("idle production lost to %s (scores: %+v)", candidates[0].Type, candidates)
			}
			gateway := &recordingGateway{revision: p.Revision}
			driver := NewDriver(staticProjection{bytes: raw}, gateway, Config{ActionBudget: 1, CloseoutAfterAction: true})
			result, err := driver.Drive(context.Background(), 1, "ai", p.Revision)
			if err != nil {
				t.Fatal(err)
			}
			if result.Commands != 2 || len(gateway.requests) != 2 || gateway.requests[0].Type != engine.CommandQueueProduction {
				t.Fatalf("one-action driver failed to queue before closing: %+v %+v", result, gateway.requests)
			}
			expected := engine.CommandPass
			if planning != "" {
				expected = engine.CommandRoundComplete
			}
			if gateway.requests[1].Type != expected {
				t.Fatalf("closeout = %s, want %s", gateway.requests[1].Type, expected)
			}
		})
	}
}

func TestIdleProductionOutranksRoutineExploration(t *testing.T) {
	p := idleProductionProjection()
	p.PlanningStatus = visibility.PlanningStatusOpen
	p.Cities[0].TileID = "home-tile"
	p.Units = []visibility.UnitProjection{{ID: "settler", OwnerID: "ai", TileID: "home-tile", Domain: engine.DomainLand, Movement: 1, Effects: []engine.Effect{{ID: "settle", Kind: engine.EffectSettler}}}}
	p.Tiles = []visibility.TileProjection{{ID: "home-tile", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}}, {ID: "frontier", X: 1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}}}
	raw, _ := json.Marshal(p)
	candidates, err := Plan(raw, "ai")
	if err != nil {
		t.Fatal(err)
	}
	hasMove := false
	for _, c := range candidates {
		hasMove = hasMove || c.Type == engine.CommandMoveUnit
	}
	if !hasMove || candidates[0].Type != engine.CommandQueueProduction {
		t.Fatalf("idle city starved by routine exploration: %+v", candidates)
	}
}

func TestProductionPriorityDoesNotInventIneligibleItems(t *testing.T) {
	for _, mode := range []string{"empty", "unavailable", "foreign", "stale", "queued", "only-legal-unit"} {
		t.Run(mode, func(t *testing.T) {
			p := idleProductionProjection()
			switch mode {
			case "empty":
				p.Cities[0].ProductionOptions.Items = nil
			case "unavailable":
				p.Cities[0].ProductionOptions.Status = "unavailable"
			case "foreign":
				p.Cities[0].OwnerID = "other"
			case "stale":
				p.Cities[0].Stale = true
			case "queued":
				p.Cities[0].Queue = []engine.ProductionItem{{ID: "existing"}}
			case "only-legal-unit":
				p.Rules.Units = append(p.Rules.Units, rules.RuntimeUnitDefinition{ID: "UNIT_INELIGIBLE", Domain: engine.DomainLand, Movement: 1, Health: 100, Attack: 999, Defense: 999, ProductionCost: 30})
			}
			raw, _ := json.Marshal(p)
			candidates, err := Plan(raw, "ai")
			if err != nil {
				t.Fatal(err)
			}
			if mode == "only-legal-unit" {
				if candidates[0].Type != engine.CommandQueueProduction || candidates[0].TargetID != "UNIT_WARRIOR" {
					t.Fatalf("planner ignored authoritative legal items: %+v", candidates)
				}
				return
			}
			for _, c := range candidates {
				if c.Type == engine.CommandQueueProduction {
					t.Fatalf("invented ineligible production in %s: %+v", mode, c)
				}
			}
		})
	}
}
