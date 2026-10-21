package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func inspectionState(t *testing.T) engine.State {
	t.Helper()
	state := clientMetadataState()
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{TypeID: "settler", Domain: engine.DomainLand, Movement: 100, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "settler", Domain: engine.DomainLand, Movement: 100, Health: 1}}, Production: []rules.RuntimeProduction{{ID: "settler", Kind: engine.ProductionUnit, Cost: 10}}, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: engine.TerrainPlains, Cost: 100, Domains: []string{engine.DomainLand}, Food: 10, Production: 10, Gold: 5, CityFood: 20, CityProduction: 20, CityGold: 10, CityCost: 25, RiverFood: 5, RiverProduction: 5, RiverGold: 5, RiverCost: 50}}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = engine.RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Overlays = []engine.TileOverlay{{ID: "farm", TileID: "t1", OwnerID: "b", Kind: "improvement", ImprovementID: "TILEIMP_FARM", YieldFood: 7, MovementCostDelta: -5}, {ID: "unfinished", TileID: "t1", OwnerID: "b", Kind: "improvement", ImprovementID: "TILEIMP_MINE", Remaining: 2, YieldFood: 90}}
	return state
}

func TestClientInspectionEffectiveFactsSurviveFogWithoutRecomputation(t *testing.T) {
	state := inspectionState(t)
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	assert := func(projection Projection) {
		t.Helper()
		tile := metadataTile(t, projection, "t1")
		wire := metadataWire(t, tile)
		if tile.Planning != nil && tile.Planning.Cost != 20 {
			t.Errorf("planning and execution movement costs diverge: %+v", tile.Planning)
		}
		yields, ok := wire["yields"].(map[string]any)
		if !ok || yields["food"] != float64(32) || yields["production"] != float64(25) || yields["gold"] != float64(15) {
			t.Errorf("not effective engine city-center yields: %v", wire)
		}
		if wire["movement_cost"] != float64(20) || wire["yield_context"] != "city-center" {
			t.Errorf("missing display movement/context: %v", wire)
		}
		overlays, ok := wire["overlays"].([]any)
		if !ok || len(overlays) != 2 {
			t.Errorf("observation lost overlay identity")
		}
	}
	assert(projection)
	state.Units[0].TileID = "t10"
	state.Overlays = nil
	state.Tiles[1].River = false
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	state, err = engine.DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	assert(projection)
}

func TestClientRushQuoteIsSeparateAndRevisionBound(t *testing.T) {
	state := clientMetadataState()
	state.Nations = []engine.Nation{{ID: "a", Gold: 200}}
	state.Cities = []engine.City{{ID: "home", OwnerID: "a", TileID: "t0", Population: 2, StoredProduction: 40, Queue: []engine.ProductionItem{{ID: "warrior", Kind: engine.ProductionUnit, Cost: 100}}}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	wire := metadataWire(t, projection.Cities[0])
	quote, ok := wire["rush_buy_quote"].(map[string]any)
	if !ok || quote["cost"] != float64(96) || quote["revision"] != float64(state.Revision) || quote["status"] != "available" {
		t.Errorf("missing exact revision-bound quote: %v", wire)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandRushBuy {
			var payload map[string]any
			if err := json.Unmarshal(action.Payload, &payload); err != nil {
				t.Fatal(err)
			}
			if len(payload) != 1 || payload["city_id"] != "home" {
				t.Errorf("display quote polluted executable payload: %v", payload)
			}
		}
	}
	foreign, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	for _, city := range foreign.Cities {
		if _, ok := metadataWire(t, city)["rush_buy_quote"]; ok {
			t.Fatal("foreign quote leaked")
		}
	}
}

func TestClientPublicBoundsDoNotGrowWithExploration(t *testing.T) {
	state := clientMetadataState()
	first, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	state.Units[0].TileID = "t10"
	second, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, projection := range []Projection{first, second} {
		wire := metadataWire(t, projection)
		world, ok := wire["world"].(map[string]any)
		if !ok || world["width"] != float64(12) || world["height"] != float64(1) || world["wrap_x"] != false || world["wrap_y"] != false {
			t.Errorf("bad public world bounds: %v", world)
		}
		if _, exists := world["seed"]; exists {
			t.Fatal("public bounds exposed seed")
		}
	}
}
