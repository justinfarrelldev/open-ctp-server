package ai

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type recordingGateway struct {
	mu               sync.Mutex
	requests         []Submission
	revision         int64
	terminal         bool
	terminalOnSubmit bool
}

func (g *recordingGateway) Submit(_ context.Context, request Submission) (SubmissionResult, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.terminal {
		return SubmissionResult{}, ErrTerminal
	}
	for _, prior := range g.requests {
		if prior.IdempotencyKey == request.IdempotencyKey {
			return SubmissionResult{Revision: g.revision, Replayed: true}, nil
		}
	}
	g.requests = append(g.requests, request)
	g.revision++
	return SubmissionResult{Revision: g.revision, Terminal: g.terminalOnSubmit}, nil
}

type staticProjection struct{ bytes []byte }

func (p staticProjection) Projection(context.Context, int64, string) ([]byte, error) {
	return append([]byte(nil), p.bytes...), nil
}

type staticHistory struct{ steps []ExplorationStep }

func (h staticHistory) ExplorationHistory(context.Context, int64, string) ([]ExplorationStep, error) {
	return append([]ExplorationStep(nil), h.steps...), nil
}

func projectionJSON(t *testing.T) []byte {
	t.Helper()
	bytes, err := json.Marshal(visibility.Projection{
		Revision: 4,
		Turn:     2,
		Tiles: []visibility.TileProjection{
			{ID: "a", X: 0, Y: 0, Terrain: engine.TerrainPlains},
			{ID: "b", X: 1, Y: 0, Terrain: engine.TerrainForest},
		},
		Units: []visibility.UnitProjection{
			{ID: "own", OwnerID: "7", TileID: "a"},
			{ID: "enemy", OwnerID: "9", TileID: "b"},
		},
		Cities: []visibility.CityProjection{{ID: "city", OwnerID: "7", TileID: "a"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestPlannerUsesOnlyProjectionAndStableCandidateTuples(t *testing.T) {
	input := projectionJSON(t)
	first, err := Plan(input, "7")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Plan(input, "7")
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("planning is not deterministic: %s != %s", firstJSON, secondJSON)
	}
	if len(first) == 0 || first[0].Type != engine.CommandArmyCreate {
		t.Fatalf("visible army candidate missing: %#v", first)
	}
	foundPass := false
	for _, candidate := range first {
		foundPass = foundPass || candidate.Type == engine.CommandPass
	}
	if !foundPass {
		t.Fatalf("fallback pass candidate missing: %#v", first)
	}
	for index := 1; index < len(first); index++ {
		if Compare(first[index-1], first[index]) > 0 {
			t.Fatalf("candidates are not sorted: %#v", first)
		}
	}
	for _, candidate := range first {
		payload, _ := json.Marshal(candidate.Payload)
		if string(payload) == "" || contains(string(payload), "enemy") || contains(string(payload), "SECRET") {
			t.Fatalf("candidate reads hidden data: %#v", candidate)
		}
	}
}

func TestPlannerUsesProjectedScienceRateAction(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		LegalActions: []visibility.ActionProjection{{Command: engine.CommandSetScienceRate, Route: "set-science-rate", EntityID: "ai", TargetID: "100", Payload: json.RawMessage(`{"science_percent":100}`)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandSetScienceRate && candidate.EntityID == "ai" && candidate.TargetID == "100" {
			return
		}
	}
	t.Fatalf("science-rate candidate missing: %#v", candidates)
}

func TestPlannerUsesProjectedReadinessWarAction(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		LegalActions: []visibility.ActionProjection{{Command: engine.CommandSetReadiness, Route: "set-readiness", EntityID: "ai", TargetID: "war", Payload: json.RawMessage(`{"readiness":"war"}`)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandSetReadiness && candidate.EntityID == "ai" && candidate.TargetID == "war" {
			return
		}
	}
	t.Fatalf("readiness candidate missing: %#v", candidates)
}

func TestPlannerUsesProjectedCitySpecialActions(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		LegalActions: []visibility.ActionProjection{{Command: engine.CommandCreateFranchise, Route: "create-franchise", EntityID: "branch", TargetID: "city", Payload: json.RawMessage(`{"unit_id":"branch","city_id":"city"}`)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandCreateFranchise && candidate.EntityID == "branch" && candidate.TargetID == "city" {
			return
		}
	}
	t.Fatalf("city special candidate missing: %#v", candidates)
}

func TestPlannerUsesProjectedUnitSpecialTargetActions(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		LegalActions: []visibility.ActionProjection{
			{Command: engine.CommandExpel, Route: "expel", EntityID: "lawyer", TargetID: "unit-b", Payload: json.RawMessage(`{"unit_id":"lawyer","target_unit_id":"unit-b"}`)},
			{Command: engine.CommandSue, Route: "sue", EntityID: "lawyer", TargetID: "unit-c", Payload: json.RawMessage(`{"unit_id":"lawyer","target_unit_id":"unit-c"}`)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		seen[candidate.Type] = true
	}
	if !seen[engine.CommandExpel] || !seen[engine.CommandSue] {
		t.Fatalf("unit special target candidates missing: %#v", candidates)
	}
}

func TestPlannerUsesProjectedEstablishContactAction(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		LegalActions: []visibility.ActionProjection{{Command: engine.CommandEstablishContact, Route: "establish-contact", EntityID: "ai", TargetID: "foreign", Payload: json.RawMessage(`{"target_player_id":"foreign"}`)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandEstablishContact && candidate.TargetID == "foreign" {
			return
		}
	}
	t.Fatalf("establish-contact candidate missing: %#v", candidates)
}

func TestPlannerFoundsHomeCityWithSettlersOnly(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{{ID: "origin", X: 0, Y: 0, Terrain: engine.TerrainPlains}},
		Units: []visibility.UnitProjection{
			{ID: "warrior", OwnerID: "ai", TileID: "origin"},
			{ID: "settler", OwnerID: "ai", TileID: "origin", Effects: []engine.Effect{{ID: "settler", Kind: engine.EffectSettler}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := PlanWithHistory(input, "ai", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandFoundCity || candidates[0].EntityID != "settler" {
		t.Fatalf("no-city founding candidates = %#v, want settler-only found-city first", candidates)
	}
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandFoundCity && candidate.EntityID != "settler" {
			t.Fatalf("non-settler found-city candidate leaked: %#v", candidate)
		}
	}
}

func TestPlannerSkipsArmyBoundSettlerWhenChoosingFoundCity(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{{ID: "origin", X: 0, Y: 0, Terrain: engine.TerrainPlains}},
		Units: []visibility.UnitProjection{
			{ID: "settler-army", OwnerID: "ai", TileID: "origin", ArmyID: "army-1", Effects: []engine.Effect{{ID: "settler", Kind: engine.EffectSettler}}},
			{ID: "settler-free", OwnerID: "ai", TileID: "origin", Effects: []engine.Effect{{ID: "settler", Kind: engine.EffectSettler}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := PlanWithHistory(input, "ai", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandFoundCity || candidates[0].EntityID != "settler-free" {
		t.Fatalf("army-bound founding candidates = %#v, want unassigned settler-free found-city first", candidates)
	}
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandFoundCity && candidate.EntityID == "settler-army" {
			t.Fatalf("army-bound settler leaked into found-city candidates: %#v", candidate)
		}
	}
}

func TestPlannerSettlesExpansionSettlerAwayFromExistingCities(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		PlanningStatus: visibility.PlanningStatusOpen,
		Cities:         []visibility.CityProjection{{ID: "capital", OwnerID: "ai", TileID: "capital-tile"}},
		Units:          []visibility.UnitProjection{{ID: "settler", OwnerID: "ai", TileID: "frontier", Domain: engine.DomainLand, Effects: []engine.Effect{{ID: "settle", Kind: engine.EffectSettler}}}},
		Tiles:          []visibility.TileProjection{{ID: "frontier", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}}},
		Rules:          &visibility.RulesProjection{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandFoundCity || candidates[0].EntityID != "settler" {
		t.Fatalf("top candidate = %#v, want settler found-city", candidates)
	}
}

func TestPlannerMovesUnescortedSettlerOutOfCityBeforeArmyCreation(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		PlanningStatus: visibility.PlanningStatusOpen,
		Cities:         []visibility.CityProjection{{ID: "capital", OwnerID: "ai", TileID: "capital-tile"}},
		Units:          []visibility.UnitProjection{{ID: "settler", OwnerID: "ai", TileID: "capital-tile", Domain: engine.DomainLand, Movement: 1, Effects: []engine.Effect{{ID: "settle", Kind: engine.EffectSettler}}}},
		Tiles:          []visibility.TileProjection{{ID: "capital-tile", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}}, {ID: "frontier", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}}},
		Rules:          &visibility.RulesProjection{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandMoveUnit || candidates[0].EntityID != "settler" {
		t.Fatalf("top candidate = %#v, want settler move before army creation", candidates)
	}
}

func TestPlannerSelectsVisibleLegalEmbassyWithoutRunnerInstruction(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Tiles:  []visibility.TileProjection{{ID: "a", X: 0, Y: 0}, {ID: "b", X: 1, Y: 0}},
		Units:  []visibility.UnitProjection{{ID: "envoy", OwnerID: "7", TileID: "a", Movement: 1, Effects: []engine.Effect{{ID: "embassy", Kind: engine.EffectSpecialEmbassy}}}},
		Cities: []visibility.CityProjection{{ID: "foreign-city", OwnerID: "9", TileID: "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "7")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandEstablishEmbassy {
		t.Fatalf("visible embassy candidate = %#v", candidates)
	}
}

func TestPlannerUsesHistoryToChooseNovelVisibleEmbassyApproach(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "visited-approach", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "fresh-step", X: 0, Y: 1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "fresh-middle", X: 1, Y: 1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "fresh-approach", X: 2, Y: 1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "foreign-city-tile", X: 2, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units:  []visibility.UnitProjection{{ID: "envoy", OwnerID: "ai", TileID: "origin", Movement: 3, Domain: "land", Effects: []engine.Effect{{ID: "embassy", Kind: engine.EffectSpecialEmbassy}}}},
		Cities: []visibility.CityProjection{{ID: "foreign-city", OwnerID: "foreign", TileID: "foreign-city-tile"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := PlanWithHistory(input, "ai", []ExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "visited-approach", Sequence: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandMoveUnit {
		t.Fatalf("embassy approach candidates = %#v", candidates)
	}
	var payload struct {
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(candidates[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Path) != 1 || payload.Path[0] == "visited-approach" {
		t.Fatalf("history-aware embassy approach = %#v, want a route to a novel visible endpoint", candidates[0])
	}
}

func TestPlannerExploresVisibleFrontierWithoutReversingBounce(t *testing.T) {
	projection := visibility.Projection{
		Turn: 4,
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "reverse", X: 0, Y: -1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "reverse-north", X: 0, Y: -2},
			{ID: "reverse-west", X: -1, Y: -1},
			{ID: "reverse-east", X: 1, Y: -1},
			{ID: "frontier", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "7", TileID: "origin", Domain: "land", Movement: 1}},
	}
	input, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Plan(input, "7")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Plan(input, "7")
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("frontier planning changed for identical public projection: %s != %s", firstJSON, secondJSON)
	}
	for _, candidate := range first {
		if candidate.Type != engine.CommandMoveUnit {
			continue
		}
		var payload struct {
			Path []string `json:"path"`
		}
		if err := json.Unmarshal(candidate.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Path) != 1 || payload.Path[0] != "frontier" {
			t.Fatalf("frontier move = %#v, want one non-reversing step to frontier", candidate)
		}
		return
	}
	t.Fatalf("missing visible frontier move: %#v", first)
}

func TestPlannerExcludesPersistedDestinationsOnlyAmongVisibleLegalRoutes(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "visited", X: -1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "fresh", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units:  []visibility.UnitProjection{{ID: "scout", OwnerID: "ai", TileID: "origin", Domain: "land", Movement: 1, ArmyID: "army"}},
		Cities: []visibility.CityProjection{{ID: "home", OwnerID: "ai", TileID: "origin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	steps := []ExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "visited", Sequence: 7}}
	candidates, err := PlanWithHistory(input, "ai", steps)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandMoveUnit {
		t.Fatalf("history-aware exploration candidates = %#v", candidates)
	}
	var payload struct {
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(candidates[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Path) != 1 || payload.Path[0] != "fresh" {
		t.Fatalf("history-aware route = %#v, want visible unvisited path step", candidates[0])
	}
	allVisited, err := PlanWithHistory(input, "ai", append(steps, ExplorationStep{CommandType: engine.CommandMoveUnit, DestinationTileID: "fresh", Sequence: 1}))
	if err != nil {
		t.Fatal(err)
	}
	if len(allVisited) == 0 || allVisited[0].Type != engine.CommandMoveUnit {
		t.Fatalf("planner discarded all legal visible routes after exhausting history: %#v", allVisited)
	}

}

func TestPlannerExploresAwayFromResolvedVisibleRelationship(t *testing.T) {
	projection := visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "west", X: 4, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "origin", X: 5, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "east", X: 6, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "ai", TileID: "origin", ArmyID: "army", Domain: "land", Movement: 1}},
		Cities: []visibility.CityProjection{
			{ID: "own-city", OwnerID: "ai", TileID: "origin"},
			{ID: "known-city", OwnerID: "known", TileID: "west"},
		},
		EmbassyPlayerIDs: []string{"known"},
	}
	input, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := Plan(input, "ai")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandMoveUnit {
		t.Fatalf("coverage candidate = %#v", candidates)
	}
	var payload struct {
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(candidates[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Path) != 1 || payload.Path[0] != "east" {
		t.Fatalf("resolved-relationship coverage move = %#v, want east frontier", candidates[0])
	}
}

func planningStatusProjectionJSON(t *testing.T, status string) []byte {
	t.Helper()
	projection := map[string]any{
		"revision": 4,
		"turn":     2,
		"tiles": []map[string]any{
			{"id": "origin", "x": 0, "y": 0, "terrain": engine.TerrainPlains},
		},
		"units": []map[string]any{
			{"id": "settler", "owner_id": "ai", "tile_id": "origin"},
		},
		"cities": []map[string]any{},
	}
	if status != "" {
		projection["planning_status"] = status
	}
	bytes, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func openFallbackProjectionJSON(t *testing.T) []byte {
	t.Helper()
	bytes, err := json.Marshal(visibility.Projection{
		Revision:       4,
		Turn:           2,
		PlanningStatus: "open",
		Tiles:          []visibility.TileProjection{{ID: "origin", X: 0, Y: 0, Terrain: engine.TerrainPlains}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestDriverClosesQueuedPlanningWithOrdinaryRoundComplete(t *testing.T) {
	gateway := &recordingGateway{revision: 4}
	driver := NewDriver(staticProjection{bytes: planningStatusProjectionJSON(t, "queued")}, gateway, Config{ActionBudget: 1})
	result, err := driver.Drive(context.Background(), 55, "ai", 4)
	if err != nil {
		t.Fatal(err)
	}
	if result.Commands != 1 || result.Revision != 5 {
		t.Fatalf("unexpected queued-closeout result: %#v", result)
	}
	if len(gateway.requests) != 1 {
		t.Fatalf("queued-closeout requests = %#v, want one round-complete", gateway.requests)
	}
	if got := gateway.requests[0].Type; got != engine.CommandRoundComplete {
		t.Fatalf("queued planning command = %q, want %q", got, engine.CommandRoundComplete)
	}
	if payload := string(gateway.requests[0].Payload); payload != `{}` {
		t.Fatalf("queued planning payload = %s, want {}", payload)
	}
}

func TestDriverRebuildsQueuedCloseoutDecisionAfterRestart(t *testing.T) {
	firstGateway := &recordingGateway{revision: 4}
	secondGateway := &recordingGateway{revision: 4}
	input := planningStatusProjectionJSON(t, "queued")
	first := NewDriver(staticProjection{bytes: input}, firstGateway, Config{ActionBudget: 1})
	second := NewDriver(staticProjection{bytes: input}, secondGateway, Config{ActionBudget: 1})
	if _, err := first.Drive(context.Background(), 55, "ai", 4); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Drive(context.Background(), 55, "ai", 4); err != nil {
		t.Fatal(err)
	}
	if len(firstGateway.requests) != 1 || len(secondGateway.requests) != 1 {
		t.Fatalf("restart queued-closeout requests differ: first=%#v second=%#v", firstGateway.requests, secondGateway.requests)
	}
	if firstGateway.requests[0].Type != engine.CommandRoundComplete || secondGateway.requests[0].Type != engine.CommandRoundComplete {
		t.Fatalf("restart queued-closeout types differ: first=%#v second=%#v", firstGateway.requests, secondGateway.requests)
	}
	if firstGateway.requests[0].CommandID != secondGateway.requests[0].CommandID || firstGateway.requests[0].IdempotencyKey != secondGateway.requests[0].IdempotencyKey {
		t.Fatalf("restart queued-closeout receipt identity differs: first=%#v second=%#v", firstGateway.requests, secondGateway.requests)
	}
}

func TestDriverReturnsZeroCommandsWhenPlanningAlreadyClosed(t *testing.T) {
	gateway := &recordingGateway{revision: 4}
	driver := NewDriver(staticProjection{bytes: planningStatusProjectionJSON(t, "closed")}, gateway, Config{ActionBudget: 1})
	result, err := driver.Drive(context.Background(), 55, "ai", 4)
	if err != nil {
		t.Fatal(err)
	}
	if result.Commands != 0 || result.Revision != 4 || result.Replayed {
		t.Fatalf("closed planning result = %#v, want zero-command no-op at same revision", result)
	}
	if len(gateway.requests) != 0 {
		t.Fatalf("closed planning submitted commands: %#v", gateway.requests)
	}
}

func TestPlannerUsesRoundCompleteFallbackOnlyForOpenPlanning(t *testing.T) {
	candidates, err := PlanWithHistory(openFallbackProjectionJSON(t), "ai", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 {
		t.Fatal("planner returned no candidates")
	}
	if got := candidates[0].Type; got != engine.CommandRoundComplete {
		t.Fatalf("open planning fallback = %q, want %q; candidates=%#v", got, engine.CommandRoundComplete, candidates)
	}
}

func TestDriverClosesOpenPlanningThroughOrdinaryGatewayRoundComplete(t *testing.T) {
	gateway := &recordingGateway{revision: 4}
	observed := []Submission{}
	driver := NewDriver(staticProjection{bytes: openFallbackProjectionJSON(t)}, gateway, Config{ActionBudget: 1, Observe: func(submission Submission) {
		observed = append(observed, submission)
	}})
	result, err := driver.Drive(context.Background(), 55, "ai", 4)
	if err != nil {
		t.Fatal(err)
	}
	if result.Commands != 1 || result.Revision != 5 || result.Replayed {
		t.Fatalf("open planning result = %#v, want one accepted ordinary closeout", result)
	}
	if len(gateway.requests) != 1 || len(observed) != 1 {
		t.Fatalf("open planning gateway=%#v observed=%#v, want one ordinary submission", gateway.requests, observed)
	}
	if got := gateway.requests[0].Type; got != engine.CommandRoundComplete {
		t.Fatalf("open planning command = %q, want %q", got, engine.CommandRoundComplete)
	}
	if observed[0].CommandID != gateway.requests[0].CommandID || observed[0].IdempotencyKey != gateway.requests[0].IdempotencyKey {
		t.Fatalf("open planning observed submission = %#v, want %#v", observed[0], gateway.requests[0])
	}
}

func TestDriverRebuildsOpenPlanningCloseoutDecisionAfterRestart(t *testing.T) {
	firstGateway := &recordingGateway{revision: 4}
	secondGateway := &recordingGateway{revision: 4}
	input := openFallbackProjectionJSON(t)
	first := NewDriver(staticProjection{bytes: input}, firstGateway, Config{ActionBudget: 1})
	second := NewDriver(staticProjection{bytes: input}, secondGateway, Config{ActionBudget: 1})
	if _, err := first.Drive(context.Background(), 55, "ai", 4); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Drive(context.Background(), 55, "ai", 4); err != nil {
		t.Fatal(err)
	}
	if len(firstGateway.requests) != 1 || len(secondGateway.requests) != 1 {
		t.Fatalf("restart open-closeout requests differ: first=%#v second=%#v", firstGateway.requests, secondGateway.requests)
	}
	if firstGateway.requests[0].Type != engine.CommandRoundComplete || secondGateway.requests[0].Type != engine.CommandRoundComplete {
		t.Fatalf("restart open-closeout types differ: first=%#v second=%#v", firstGateway.requests, secondGateway.requests)
	}
	if firstGateway.requests[0].CommandID != secondGateway.requests[0].CommandID || firstGateway.requests[0].IdempotencyKey != secondGateway.requests[0].IdempotencyKey {
		t.Fatalf("restart open-closeout receipt identity differs: first=%#v second=%#v", firstGateway.requests, secondGateway.requests)
	}
}

func TestDriverIsDeterministicIdempotentAndBounded(t *testing.T) {
	gateway := &recordingGateway{revision: 4}
	driver := NewDriver(staticProjection{bytes: projectionJSON(t)}, gateway, Config{ActionBudget: 1})
	result, err := driver.Drive(context.Background(), 55, "7", 4)
	if err != nil {
		t.Fatal(err)
	}
	if result.Commands != 1 || result.Revision != 5 {
		t.Fatalf("unexpected drive result: %#v", result)
	}
	replay, err := driver.Drive(context.Background(), 55, "7", 4)
	if err != nil {
		t.Fatal(err)
	}
	if replay.Commands != 1 || !replay.Replayed {
		t.Fatalf("restart must replay the same ordinary receipt: %#v", replay)
	}
	gateway.mu.Lock()
	defer gateway.mu.Unlock()
	if len(gateway.requests) != 1 || gateway.requests[0].ActorPlayerID != "7" || gateway.requests[0].IdempotencyKey == "" {
		t.Fatalf("driver did not submit one stable ordinary command: %#v", gateway.requests)
	}
}

func TestDriverRebuildsIdenticalHistoryAwareTranscriptAfterRestart(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		PlanningStatus: visibility.PlanningStatusOpen,
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "visited", X: -1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "fresh", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units:  []visibility.UnitProjection{{ID: "scout", OwnerID: "ai", TileID: "origin", Domain: "land", Movement: 1, ArmyID: "army"}},
		Cities: []visibility.CityProjection{{ID: "home", OwnerID: "ai", TileID: "origin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	history := staticHistory{steps: []ExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "visited", Sequence: 7}}}
	firstGateway := &recordingGateway{revision: 4}
	secondGateway := &recordingGateway{revision: 4}
	first := NewDriver(staticProjection{bytes: input}, firstGateway, Config{ActionBudget: 1, History: history})
	second := NewDriver(staticProjection{bytes: input}, secondGateway, Config{ActionBudget: 1, History: history})
	if _, err := first.Drive(context.Background(), 55, "ai", 4); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Drive(context.Background(), 55, "ai", 4); err != nil {
		t.Fatal(err)
	}
	if len(firstGateway.requests) != 1 || len(secondGateway.requests) != 1 || firstGateway.requests[0].CommandID != secondGateway.requests[0].CommandID || string(firstGateway.requests[0].Payload) != string(secondGateway.requests[0].Payload) {
		t.Fatalf("restart transcript differs: first=%#v second=%#v", firstGateway.requests, secondGateway.requests)
	}
	if !contains(string(firstGateway.requests[0].Payload), "fresh") {
		t.Fatalf("history-aware driver did not exclude the recorded step: %#v", firstGateway.requests[0])
	}
}

func TestDriverObservesOnlySubmittedAutonomousCommands(t *testing.T) {
	gateway := &recordingGateway{}
	input := projectionJSON(t)
	observed := []Submission{}
	driver := NewDriver(staticProjection{bytes: input}, gateway, Config{ActionBudget: 1, Observe: func(submission Submission) {
		observed = append(observed, submission)
	}})
	result, err := driver.Drive(context.Background(), 9, "7", 4)
	if err != nil {
		t.Fatal(err)
	}
	if result.Commands != 1 || len(observed) != 1 || len(gateway.requests) != 1 {
		t.Fatalf("commands=%d observed=%#v submitted=%#v", result.Commands, observed, gateway.requests)
	}
	if got, want := observed[0], gateway.requests[0]; got.CommandID != want.CommandID || got.Type != want.Type || string(got.Payload) != string(want.Payload) {
		t.Fatalf("observed submission = %#v, want %#v", got, want)
	}
}

func TestDriverCancelsTerminalAndConcurrentDuplicateJobs(t *testing.T) {
	gateway := &recordingGateway{revision: 4}
	driver := NewDriver(staticProjection{bytes: projectionJSON(t)}, gateway, Config{ActionBudget: 1})
	var group sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := driver.Drive(context.Background(), 55, "7", 4)
			errs <- err
		}()
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	gateway.mu.Lock()
	if len(gateway.requests) != 1 {
		t.Fatalf("duplicate jobs submitted %d commands", len(gateway.requests))
	}
	gateway.terminal = true
	gateway.mu.Unlock()
	if _, err := driver.Drive(context.Background(), 55, "7", 5); !errors.Is(err, ErrTerminal) {
		t.Fatalf("terminal game must stop driver: %v", err)
	}
	acceptedTerminal := NewDriver(staticProjection{bytes: projectionJSON(t)}, &recordingGateway{revision: 4, terminalOnSubmit: true}, Config{ActionBudget: 1})
	if result, err := acceptedTerminal.Drive(context.Background(), 56, "7", 4); !errors.Is(err, ErrTerminal) || result.Commands != 1 || result.Revision != 5 {
		t.Fatalf("accepted terminal command = %#v, %v", result, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := driver.Drive(ctx, 55, "7", 5); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled job must not submit: %v", err)
	}
}

func TestPlannerPursuesStaleForeignCityWithoutEmbassyCommand(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "west", X: -1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "east", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "foreign-city-tile", X: 2, Y: 0, Stale: true},
		},
		Units:  []visibility.UnitProjection{{ID: "envoy", OwnerID: "ai", TileID: "origin", ArmyID: "army", Domain: "land", Movement: 1, Effects: []engine.Effect{{ID: "embassy", Kind: engine.EffectSpecialEmbassy}}}},
		Cities: []visibility.CityProjection{{ID: "home", OwnerID: "ai", TileID: "origin"}, {ID: "foreign", OwnerID: "foreign", TileID: "foreign-city-tile", Stale: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := PlanWithHistory(input, "ai", []ExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "west", Sequence: 1}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := PlanWithHistory(input, "ai", []ExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "west", Sequence: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 || first[0].Type != engine.CommandMoveUnit || string(first[0].Payload) != string(second[0].Payload) {
		t.Fatalf("stale pursuit candidates differ or do not move: first=%#v second=%#v", first, second)
	}
	if contains(string(first[0].Payload), "foreign-city-tile") || !contains(string(first[0].Payload), "east") {
		t.Fatalf("stale pursuit move = %#v, want legal visible east step", first[0])
	}
	for _, candidate := range first {
		if candidate.Type == engine.CommandEstablishEmbassy {
			t.Fatalf("stale city produced embassy command: %#v", candidate)
		}
	}
}

func TestPlannerRanksTheSubmittedExplorationStepBeforeAReachableEndpoint(t *testing.T) {
	input, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "revisited-hop", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "distant-frontier", X: 2, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "fresh-hop", X: 0, Y: 1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "fresh-west", X: -1, Y: 1},
			{ID: "fresh-east", X: 1, Y: 1},
			{ID: "fresh-south", X: 0, Y: 2},
		},
		Units:  []visibility.UnitProjection{{ID: "scout", OwnerID: "ai", TileID: "origin", ArmyID: "army", Domain: "land", Movement: 2}},
		Cities: []visibility.CityProjection{{ID: "home", OwnerID: "ai", TileID: "origin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := PlanWithHistory(input, "ai", []ExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "revisited-hop", Sequence: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].Type != engine.CommandMoveUnit {
		t.Fatalf("exploration candidates = %#v", candidates)
	}
	var payload struct {
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(candidates[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Path) != 1 || payload.Path[0] != "fresh-hop" {
		t.Fatalf("submitted exploration hop = %#v, want current visible unvisited hop", candidates[0])
	}
}

func contains(value, fragment string) bool {
	return len(fragment) > 0 && len(value) >= len(fragment) && (value == fragment || containsAt(value, fragment))
}

func containsAt(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
