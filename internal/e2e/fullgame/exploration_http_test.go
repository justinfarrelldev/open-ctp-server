package fullgame

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/service"
)

func TestHTTPExplorationTransportRefreshesOnlyFromPublicMoveGateway(t *testing.T) {
	initial := []byte(`{"revision":1,"tiles":[]}`)
	refreshed := []byte(`{"revision":2,"tiles":[{"id":"north"}]}`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/game/9/move" {
			t.Fatalf("public move request = %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["expected_revision"] != float64(1) || body["unit_id"] != "scout" {
			t.Fatalf("gateway authority fields = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(service.Result{Revision: 2, ViewerProjection: refreshed})
	}))
	defer server.Close()
	transport := &HTTPExplorationTransport{BaseURL: server.URL, GameID: 9, SessionID: 3, Revision: 1, InitialProjection: initial}
	projection, err := transport.Move(context.Background(), "scout", []string{"north"})
	if err != nil {
		t.Fatal(err)
	}
	if string(projection) != string(refreshed) || string(transport.InitialProjection) != string(refreshed) || transport.Revision != 2 {
		t.Fatalf("gateway projection refresh = %q / %#v", projection, transport)
	}
}

func TestHTTPExplorationTransportBootstrapsFromPublicSSEResync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/game/9/events" || r.URL.Query().Get("session_id") != "3" || r.Header.Get("Last-Event-ID") != "0" {
			t.Fatalf("SSE bootstrap request = %s %s %q", r.Method, r.URL.String(), r.Header.Get("Last-Event-ID"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("id: 7\nevent: resync\ndata: {\"current_revision\":7,\"snapshot\":{\"revision\":7,\"tiles\":[]}}\n\n"))
	}))
	defer server.Close()
	transport := &HTTPExplorationTransport{BaseURL: server.URL, GameID: 9, SessionID: 3}
	cursor, projection, err := transport.Bootstrap(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if cursor != 7 || transport.Revision != 7 || string(projection) != `{"revision":7,"tiles":[]}` {
		t.Fatalf("SSE bootstrap = cursor %d transport %#v projection %s", cursor, transport, projection)
	}
}

func TestHTTPExplorationTransportRejectsStaleResyncWithoutRegressingCurrentProjection(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.Header.Get("Last-Event-ID"); got != "9" {
			t.Fatalf("resync %d cursor = %q, want 9", requests, got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if requests == 1 {
			_, _ = w.Write([]byte("id: 8\nevent: resync\ndata: {\"current_revision\":8,\"snapshot\":{\"revision\":8,\"tiles\":[]}}\n\n"))
			return
		}
		_, _ = w.Write([]byte("id: 10\nevent: resync\ndata: {\"current_revision\":10,\"snapshot\":{\"revision\":10,\"tiles\":[{\"id\":\"current\"}]}}\n\n"))
	}))
	defer server.Close()

	current := []byte(`{"revision":9,"tiles":[{"id":"known"}]}`)
	transport := &HTTPExplorationTransport{BaseURL: server.URL, GameID: 9, SessionID: 3, Cursor: 9, Revision: 9, InitialProjection: current}
	if _, _, err := transport.Bootstrap(context.Background(), transport.Cursor); err == nil {
		t.Fatal("stale resync unexpectedly replaced current projection")
	}
	if transport.Cursor != 9 || transport.Revision != 9 || string(transport.InitialProjection) != string(current) {
		t.Fatalf("stale resync regressed transport: %#v", transport)
	}
	cursor, projection, err := transport.Bootstrap(context.Background(), transport.Cursor)
	if err != nil {
		t.Fatal(err)
	}
	if cursor != 10 || transport.Cursor != 10 || transport.Revision != 10 || string(projection) != `{"revision":10,"tiles":[{"id":"current"}]}` {
		t.Fatalf("current resync = cursor %d transport %#v projection %s", cursor, transport, projection)
	}
}

func TestHTTPExplorationTransportMoveUsesRevisionFromCurrentResync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if got := r.Header.Get("Last-Event-ID"); got != "4" {
				t.Fatalf("resume cursor = %q, want 4", got)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("id: 6\nevent: resync\ndata: {\"current_revision\":6,\"snapshot\":{\"revision\":6,\"tiles\":[{\"id\":\"next\"}]}}\n\n"))
		case http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["expected_revision"] != float64(6) || body["path"] == nil {
				t.Fatalf("move did not use current public revision: %#v", body)
			}
			_ = json.NewEncoder(w).Encode(service.Result{Revision: 7, ViewerProjection: []byte(`{"revision":7,"tiles":[{"id":"after-move"}]}`)})
		default:
			t.Fatalf("unexpected public request %s", r.Method)
		}
	}))
	defer server.Close()

	transport := &HTTPExplorationTransport{BaseURL: server.URL, GameID: 9, SessionID: 3, Cursor: 4, Revision: 4, InitialProjection: []byte(`{"revision":4,"tiles":[]}`)}
	if _, _, err := transport.Bootstrap(context.Background(), transport.Cursor); err != nil {
		t.Fatal(err)
	}
	if _, err := transport.Move(context.Background(), "scout", []string{"next"}); err != nil {
		t.Fatal(err)
	}
	if transport.Revision != 7 || transport.Cursor != 6 {
		t.Fatalf("move current state = %#v", transport)
	}
}

func TestHTTPExplorationTransportUsesDedicatedPublicCombatEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/game/9/combat" {
			t.Fatalf("public combat request = %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["expected_revision"] != float64(7) || body["attacker_army_id"] != "a" || body["defender_army_id"] != "b" {
			t.Fatalf("combat authority fields = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(service.Result{Revision: 8, ViewerProjection: []byte(`{"revision":8,"tiles":[]}`), Combat: &engine.CombatResult{Outcome: engine.CombatOutcomeAttackerWon, CapturedCityID: "city-b"}})
	}))
	defer server.Close()
	transport := &HTTPExplorationTransport{BaseURL: server.URL, GameID: 9, SessionID: 3, Revision: 7}
	combat, err := transport.Combat(context.Background(), "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if combat.Result.Outcome != engine.CombatOutcomeAttackerWon || combat.Result.CapturedCityID != "city-b" || combat.Terminal {
		t.Fatalf("combat transition = %#v", combat)
	}
	if transport.Revision != 8 || string(combat.Projection) != `{"revision":8,"tiles":[]}` {
		t.Fatalf("combat refresh = %#v %s", transport, combat.Projection)
	}
}

func TestHTTPExplorationTransportCommandReceiptIsReplayStableAndRevisionDistinct(t *testing.T) {
	transport := &HTTPExplorationTransport{SessionID: 3, Revision: 7, Namespace: "scenario-attacker"}
	payload := map[string]json.RawMessage{}
	first, err := transport.commandID("pass", payload)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := transport.commandID("pass", payload)
	if err != nil {
		t.Fatal(err)
	}
	if first != replay {
		t.Fatalf("identical pass replay receipt changed: %q != %q", first, replay)
	}
	transport.Revision++
	later, err := transport.commandID("pass", payload)
	if err != nil {
		t.Fatal(err)
	}
	if later == first {
		t.Fatalf("later pass reused receipt %q", later)
	}
}
