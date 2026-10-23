package puppet

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestPublicHumanRouteSupportsRushBuyAfterProductionStarts(t *testing.T) {
	route, err := publicRoute(engine.CommandRushBuy)
	if err != nil || route != "rush-buy" {
		t.Fatalf("rush-buy route = %q, %v", route, err)
	}
}

type recordingHumanHTTPClient struct {
	command HumanCommand
	calls   int
}

func (c *recordingHumanHTTPClient) SubmitHumanCommand(_ context.Context, command HumanCommand) (SubmissionResult, error) {
	c.command = command
	c.calls++
	return SubmissionResult{Revision: command.ExpectedRevision + 1}, nil
}

func TestPuppeteerPlansOnlyFromDetachedPublicBytes(t *testing.T) {
	state := engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		Revision:      1,
		Players: []engine.Player{
			{ID: "account-7", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman},
			{ID: "account-8", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
		},
		Nations: []engine.Nation{{ID: "nation-a"}, {ID: "nation-b"}},
		Phase:   engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerOrder: []string{"account-7", "account-8"}},
		Tiles: []engine.Tile{
			{ID: "visible-origin", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true},
			{ID: "HIDDEN_CANARY_TILE", X: 20, Y: 20, Terrain: engine.TerrainForest, Open: true},
		},
		Units: []engine.Unit{
			{ID: "public-scout", OwnerID: "account-7", TileID: "visible-origin", Domain: engine.DomainLand, Movement: 2},
			{ID: "HIDDEN_CANARY_UNIT", OwnerID: "account-8", TileID: "HIDDEN_CANARY_TILE", Domain: engine.DomainLand, Movement: 2},
		},
		RNG: []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
	}
	projection, err := visibility.AIInput(state, "account-7")
	if err != nil {
		t.Fatal(err)
	}
	history, err := json.Marshal([]PublicExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "visible-origin", Sequence: 4}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(projection), "HIDDEN_CANARY") || strings.Contains(string(history), "HIDDEN_CANARY") {
		t.Fatalf("public input fixture leaked hidden canary: projection=%s history=%s", projection, history)
	}
	client := &recordingHumanHTTPClient{}
	var observed RawPlanningInput
	puppeteer := Puppeteer{Client: client, ObservePlanningInput: func(input RawPlanningInput) { observed = input }}
	result, err := puppeteer.SelectAndSubmit(context.Background(), HumanActor{GameID: 44, SessionID: 70}, PlanningInput{
		ActorPlayerID: "account-7", ExpectedRevision: 4, PublicProjectionJSON: projection, PublicHistoryJSON: history,
		TestHarnessPolicyLevel: TestHarnessPolicyLevelPassOnly,
	}, "puppet-pass-4")
	if err != nil {
		t.Fatal(err)
	}
	projection[0] = 'x'
	history[0] = 'x'
	if strings.Contains(string(observed.PublicProjectionJSON), "HIDDEN_CANARY") || strings.Contains(string(observed.PublicHistoryJSON), "HIDDEN_CANARY") {
		t.Fatalf("raw puppet planning input leaked hidden canary: %#v", observed)
	}
	if !json.Valid(observed.PublicProjectionJSON) || !json.Valid(observed.PublicHistoryJSON) {
		t.Fatal("observed public planning bytes alias caller buffers")
	}
	if result.Candidate.Type != engine.CommandPass || result.Revision != 5 || client.calls != 1 {
		t.Fatalf("unexpected puppet result: result=%#v calls=%d", result, client.calls)
	}
	if client.command.Route != "pass" || client.command.SessionID != 70 || client.command.ExpectedRevision != 4 || client.command.IdempotencyKey != "puppet-pass-4" {
		t.Fatalf("unexpected public human command: %#v", client.command)
	}
}

func TestPuppeteerUsesSerializedPublicHistoryForProductionPlanning(t *testing.T) {
	projection, err := json.Marshal(visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}},
			{ID: "visited", X: -1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}},
			{ID: "fresh", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{engine.DomainLand}, Accessible: true}},
		},
		Units:  []visibility.UnitProjection{{ID: "scout", OwnerID: "account-7", TileID: "origin", Domain: engine.DomainLand, Movement: 1, ArmyID: "army"}},
		Cities: []visibility.CityProjection{{ID: "home", OwnerID: "account-7", TileID: "origin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	history, err := json.Marshal([]PublicExplorationStep{{CommandType: engine.CommandMoveUnit, DestinationTileID: "visited", Sequence: 7}})
	if err != nil {
		t.Fatal(err)
	}
	client := &recordingHumanHTTPClient{}
	result, err := (Puppeteer{Client: client}).SelectAndSubmit(context.Background(), HumanActor{GameID: 44, SessionID: 70}, PlanningInput{
		ActorPlayerID: "account-7", ExpectedRevision: 4, PublicProjectionJSON: projection, PublicHistoryJSON: history,
		TestHarnessPolicyLevel: TestHarnessPolicyLevelProductionFirst,
	}, "history-aware-move")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(result.Candidate.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if result.Candidate.Type != engine.CommandMoveUnit || len(payload.Path) != 1 || payload.Path[0] != "fresh" {
		t.Fatalf("history-aware puppet candidate = %#v payload=%#v, want fresh visible step", result.Candidate, payload)
	}
	if client.command.Route != "move" {
		t.Fatalf("history-aware puppet route = %q, want move", client.command.Route)
	}
}

func TestPuppeteerRejectsUnknownHarnessPolicyBeforeSubmission(t *testing.T) {
	client := &recordingHumanHTTPClient{}
	_, err := (Puppeteer{Client: client}).SelectAndSubmit(context.Background(), HumanActor{GameID: 1, SessionID: 1}, PlanningInput{
		ActorPlayerID: "account-1", ExpectedRevision: 1, PublicProjectionJSON: []byte(`{}`), PublicHistoryJSON: []byte(`[]`),
		TestHarnessPolicyLevel: "ctp2-difficulty-hard",
	}, "invalid-policy")
	if err == nil || client.calls != 0 {
		t.Fatalf("unknown/non-harness policy reached submission: err=%v calls=%d", err, client.calls)
	}
}

func TestClientSubmitsEnrichedPublicCommandAndDecodesRevision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/game/44/move-unit" || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected request: method=%s path=%s content-type=%q", r.Method, r.URL.Path, r.Header.Get("Content-Type"))
		}
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		for name, want := range map[string]string{
			"session_id":        "70",
			"expected_revision": "4",
			"idempotency_key":   `"move-4"`,
			"command_id":        `"command-4"`,
			"unit_id":           `"unit-7"`,
		} {
			if string(body[name]) != want {
				t.Errorf("body[%q] = %s, want %s", name, body[name], want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"revision":5}`))
	}))
	defer server.Close()

	result, err := (Client{BaseURL: server.URL, HTTPClient: server.Client()}).SubmitHumanCommand(context.Background(), HumanCommand{
		GameID: 44, SessionID: 70, ExpectedRevision: 4, Route: "move-unit", IdempotencyKey: "move-4", CommandID: "command-4", Payload: []byte(`{"unit_id":"unit-7"}`),
	})
	if err != nil || result.Revision != 5 {
		t.Fatalf("submission result = %#v, err=%v", result, err)
	}
}

func TestClientRejectsInvalidCommandsAndHTTPResponses(t *testing.T) {
	client := Client{}
	if _, err := client.SubmitHumanCommand(context.Background(), HumanCommand{}); err == nil {
		t.Fatal("invalid command was accepted")
	}
	valid := HumanCommand{GameID: 1, SessionID: 2, ExpectedRevision: 3, Route: "pass", IdempotencyKey: "pass-3", CommandID: "command-3"}
	malformed := valid
	malformed.Payload = []byte(`{"bad":`)
	if _, err := client.SubmitHumanCommand(context.Background(), malformed); err == nil || !strings.Contains(err.Error(), "decode public human payload") {
		t.Fatalf("malformed payload error = %v", err)
	}

	forbidden := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden detail must not escape", http.StatusForbidden)
	}))
	valid.Payload = []byte(`{}`)
	_, err := (Client{BaseURL: forbidden.URL, HTTPClient: forbidden.Client()}).SubmitHumanCommand(context.Background(), valid)
	forbidden.Close()
	var httpError *HTTPError
	if !errors.As(err, &httpError) || httpError.Status != http.StatusForbidden || httpError.Error() != "public human command returned status 403" {
		t.Fatalf("forbidden error = %#v", err)
	}

	malformedResponse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"revision":`))
	}))
	_, err = (Client{BaseURL: malformedResponse.URL, HTTPClient: malformedResponse.Client()}).SubmitHumanCommand(context.Background(), valid)
	malformedResponse.Close()
	if err == nil {
		t.Fatal("malformed success response was accepted")
	}
}
