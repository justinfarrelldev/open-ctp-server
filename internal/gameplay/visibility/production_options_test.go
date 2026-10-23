package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func productionOptionsState(t *testing.T) engine.State {
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{
		{ID: "first", Domain: engine.DomainLand, Movement: 1, Health: 1, ProductionCost: 10},
		{ID: "second", Domain: engine.DomainLand, Movement: 1, Health: 1, ProductionCost: 20},
		{ID: "special", Domain: engine.DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{Kind: engine.EffectHasPopAndCanBuild}}},
		{ID: "park", Domain: engine.DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{Kind: engine.EffectSpecialCreatePark}}},
	}, Production: []rules.RuntimeProduction{{ID: "building", Kind: engine.ProductionBuilding, Cost: 30}}}
	return projectionStateWithRuntime(t, runtime)
}

func TestProductionOptionsCompleteAppendAndReadOnly(t *testing.T) {
	state := productionOptionsState(t)
	for _, nonempty := range []bool{false, true} {
		if nonempty {
			state.Cities[0].Queue = []engine.ProductionItem{{ID: "first", Kind: engine.ProductionUnit, Cost: 10}}
			state.Cities[0].StoredProduction = 7
		}
		before := mustStateJSON(t, state)
		p, err := Project(state, "a")
		if err != nil {
			t.Fatal(err)
		}
		wire := metadataWire(t, p.Cities[0])
		options, ok := wire["production_options"].(map[string]any)
		if !ok {
			t.Fatalf("missing complete options: %v", wire)
		}
		if options["status"] != "complete" || options["revision"] != float64(state.Revision) || options["city_id"] != "city-a" || options["content_hash"] != state.Rules.ContentHash || options["operation"] != "append" {
			t.Fatalf("unbound options: %v", options)
		}
		items := options["items"].([]any)
		if len(items) != 3 {
			t.Fatalf("incomplete or ineligible options: %v", items)
		}
		for _, raw := range items {
			item := raw.(map[string]any)
			id := item["id"].(string)
			if !productionActionTarget(p, engine.CommandQueueProduction, id) {
				t.Errorf("eligible action missing %s", id)
			}
			entries := []engine.QueueSelection{}
			for _, q := range p.Cities[0].Queue {
				entries = append(entries, engine.QueueSelection{EntryID: q.EntryID})
			}
			entries = append(entries, engine.QueueSelection{ItemID: id})
			body, _ := json.Marshal(engine.EditProductionPayload{CityID: "city-a", Entries: entries})
			next, _, _, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: "append", ActorPlayerID: "a", Type: engine.CommandEditProduction, Payload: body}, state.Rules, engine.Sources{Clock: topologyClock{}, IDs: &topologyIDs{}, RNG: topologyRNG{}})
			if err != nil {
				t.Fatalf("advertised append %s rejected: %v", id, err)
			}
			if nonempty && next.Cities[0].StoredProduction != 7 {
				t.Fatal("append lost head funding")
			}
		}
		if !bytes.Equal(before, mustStateJSON(t, state)) {
			t.Fatal("inspection mutated state")
		}
	}
}

func TestProductionOptionsRespectPlanningQueuedAndClosedStates(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		for _, status := range []string{"open", "queued", "closed"} {
			state := productionOptionsState(t)
			state.Phase.Mode = mode
			state.Policy.Stage = "planning"
			if status == "queued" {
				state.Policy.Queue = []engine.TurnPolicyCommand{{Kind: "action", ActorID: "a", CommandID: "queued", IdempotencyKey: "key", Payload: []byte(`{}`)}}
			}
			if status == "closed" {
				state.Policy.EarlyEnds = []string{"a"}
			}
			p, err := Project(state, "a")
			if err != nil {
				t.Fatal(err)
			}
			options := p.Cities[0].ProductionOptions
			want := "unavailable"
			if status == "open" {
				want = "complete"
			}
			if options == nil || options.Status != want {
				t.Fatalf("%s/%s: %+v", mode, status, options)
			}
		}
	}
}

func TestProductionOptionsUnavailableIsNotCompleteEmpty(t *testing.T) {
	state := productionOptionsState(t)
	state.Phase.ActivePlayerIndex = 1
	p, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	options := metadataWire(t, p.Cities[0])["production_options"].(map[string]any)
	if options["status"] != "unavailable" || len(options["items"].([]any)) != 0 {
		t.Fatalf("phase restricted options: %v", options)
	}
	p, err = Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	for _, city := range p.Cities {
		if _, ok := metadataWire(t, city)["production_options"]; ok {
			t.Fatal("foreign eligibility leaked")
		}
	}
}
