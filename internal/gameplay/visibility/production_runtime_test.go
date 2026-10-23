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

	actions := canonicalProductionActions(t, projection, "player-a")
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

func TestProductionActionsHideUnitsNeedingAbsentCityGood(t *testing.T) {
	projection := Projection{
		Cities: []CityProjection{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}},
		Rules:  &RulesProjection{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_COFFEE_EXPORTER", Domain: engine.DomainLand, Movement: 1, Health: 1, ProductionCost: 10, NeedsCityGoodID: "GOOD_COFFEE"}}},
	}
	if hasProductionAction(canonicalProductionActions(t, projection, "player-a"), "UNIT_COFFEE_EXPORTER") {
		t.Fatal("city-good-gated unit was projected without the required good")
	}
	projection.Cities[0].Goods = []string{"GOOD_COFFEE"}
	if !hasProductionAction(canonicalProductionActions(t, projection, "player-a"), "UNIT_COFFEE_EXPORTER") {
		t.Fatal("city-good-gated unit was not projected after the required good was present")
	}
}

// Legacy action fixtures now enter through canonical rules and city state rather
// than testing a second, projection-only approximation of eligibility.
func canonicalProductionActions(t *testing.T, projection Projection, actor string) []ActionProjection {
	t.Helper()
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Units: projection.Rules.Units, Production: projection.Rules.Production}
	if len(runtime.Units) == 0 {
		runtime.Units = []rules.RuntimeUnitDefinition{{ID: "dummy", Domain: engine.DomainLand, Movement: 1, Health: 1}}
	}
	if len(runtime.Production) == 0 {
		runtime.Production = []rules.RuntimeProduction{{ID: runtime.Units[0].ID, Kind: engine.ProductionUnit, Cost: 1}}
	}
	state := projectionStateWithRuntime(t, runtime)
	state.Players[0].ID = actor
	state.Units[0].OwnerID = actor
	state.Phase.ActivePlayerOrder[0] = actor
	state.Cities[0].OwnerID = actor
	state.Cities[0].Goods = projection.Cities[0].Goods
	if projection.Research != nil {
		state.Nations = []engine.Nation{{ID: state.Players[0].NationID, Advances: projection.Research.Advances}}
	}
	p, err := Project(state, actor)
	if err != nil {
		t.Fatal(err)
	}
	return p.LegalActions
}

func hasProductionAction(actions []ActionProjection, targetID string) bool {
	for _, action := range actions {
		if action.Command == engine.CommandQueueProduction && action.TargetID == targetID {
			return true
		}
	}
	return false
}
