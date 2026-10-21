package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func governorState(t *testing.T) State {
	t.Helper()
	state := journeyState(t)
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.AI = &rules.RuntimeAI{BuildingBuildLists: []rules.RuntimeAIList{
		{ID: "BUILDING_BUILD_LIST_SCIENCE", Items: []string{"LOCKED", "LIBRARY", "UNIVERSITY"}},
		{ID: "BUILDING_BUILD_LIST_PRODUCTION", Items: []string{"FACTORY"}},
	}}
	runtime.Production = append(runtime.Production,
		rules.RuntimeProduction{ID: "LOCKED", Kind: ProductionBuilding, Cost: 10, Prerequisites: []string{"ADVANCE_SYNTHETIC_CHILD"}},
		rules.RuntimeProduction{ID: "LIBRARY", Kind: ProductionBuilding, Cost: 10},
		rules.RuntimeProduction{ID: "UNIVERSITY", Kind: ProductionBuilding, Cost: 20},
		rules.RuntimeProduction{ID: "FACTORY", Kind: ProductionBuilding, Cost: 30})
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Cities = append(state.Cities, City{ID: "home", OwnerID: "player-a", TileID: "t0", Population: 3})
	return state
}

func governorWire(t *testing.T, city City) map[string]any {
	t.Helper()
	raw, _ := json.Marshal(city)
	var wire map[string]any
	_ = json.Unmarshal(raw, &wire)
	result, _ := wire["governor"].(map[string]any)
	return result
}

func TestGovernorFillsFromPinnedPriorityAndContinuesAfterRestart(t *testing.T) {
	state := governorState(t)
	state = runtimeApply(t, state, "player-a", "set-governor", map[string]any{"city_id": "home", "enabled": true, "priority": "science"})
	city, _ := cityByID(state, "home")
	if governorWire(t, *city)["enabled"] != true || len(city.Queue) != 1 || city.Queue[0].ID != "LIBRARY" || city.Queue[0].Cost != 10 {
		t.Fatalf("invalid governor decision: %+v", city)
	}
	city.StoredProduction = 10
	raw, _ := CanonicalStateJSON(state)
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	a := runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	b := runtimeApply(t, restored, "player-a", CommandRoundComplete, map[string]any{})
	if canonicalStateHashForTest(t, a) != canonicalStateHashForTest(t, b) {
		t.Fatal("governor restart diverged")
	}
	city, _ = cityByID(a, "home")
	if len(city.Buildings) != 1 || city.Buildings[0].ID != "LIBRARY" || len(city.Queue) != 1 || city.Queue[0].ID != "UNIVERSITY" {
		t.Fatalf("governor stalled: %+v", city)
	}
}

func TestGovernorPreservesManualQueueAndDisableStopsAutomation(t *testing.T) {
	state := governorState(t)
	state = runtimeApply(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "home", Item: ProductionItem{ID: "FACTORY", Kind: ProductionBuilding, Cost: 30}})
	city, _ := cityByID(state, "home")
	city.StoredProduction = 9
	entry := city.Queue[0].EntryID
	state = runtimeApply(t, state, "player-a", "set-governor", map[string]any{"city_id": "home", "enabled": true, "priority": "science"})
	city, _ = cityByID(state, "home")
	if city.Queue[0].EntryID != entry || city.StoredProduction != 9 {
		t.Fatal("governor replaced manual queue/progress")
	}
	state = runtimeApply(t, state, "player-a", "set-governor", map[string]any{"city_id": "home", "enabled": false})
	state = runtimeApply(t, state, "player-a", CommandEditProduction, EditProductionPayload{CityID: "home", Entries: []QueueSelection{}})
	state = runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	city, _ = cityByID(state, "home")
	if len(city.Queue) != 0 || governorWire(t, *city)["enabled"] != false {
		t.Fatal("disabled governor continued")
	}
}

func TestGovernorRejectsMissingUnsupportedForeignAndWrongPhase(t *testing.T) {
	state := governorState(t)
	before := canonicalStateHashForTest(t, state)
	for _, payload := range []map[string]any{
		{"city_id": "home", "enabled": true, "priority": "cheat"},
		{"city_id": "home", "priority": "science"},
		{"city_id": "home", "enabled": nil, "priority": "science"},
		{"city_id": "foreign", "enabled": true, "priority": "science"},
	} {
		if err := runtimeApplyExpectError(t, state, "player-a", "set-governor", payload); err == nil {
			t.Fatalf("accepted %+v", payload)
		}
	}
	if err := runtimeApplyExpectError(t, state, "player-b", "set-governor", map[string]any{"city_id": "home", "enabled": true, "priority": "science"}); err == nil {
		t.Fatal("wrong phase/owner accepted")
	}
	if canonicalStateHashForTest(t, state) != before {
		t.Fatal("invalid governor mutated state")
	}
}

func TestGovernorPriorityReplacementAndCaptureReset(t *testing.T) {
	state := governorState(t)
	state = runtimeApply(t, state, "player-a", "set-governor", map[string]any{"city_id": "home", "enabled": true, "priority": "science"})
	state = runtimeApply(t, state, "player-a", "set-governor", map[string]any{"city_id": "home", "enabled": true, "priority": "production"})
	city, _ := cityByID(state, "home")
	if city.Queue[0].ID != "LIBRARY" {
		t.Fatal("priority change destroyed queue")
	}
	state = runtimeApply(t, state, "player-a", CommandEditProduction, EditProductionPayload{CityID: "home", Entries: []QueueSelection{}})
	state = runtimeApply(t, state, "player-a", CommandRoundComplete, map[string]any{})
	city, _ = cityByID(state, "home")
	if len(city.Queue) != 1 || city.Queue[0].ID != "FACTORY" {
		t.Fatal("new priority ignored")
	}
	captureCityAt(&state, "player-b", "t0")
	city, _ = cityByID(state, "home")
	if settings := governorWire(t, *city); settings != nil && settings["enabled"] == true {
		t.Fatal("captured governor kept former owner's orders")
	}
}
