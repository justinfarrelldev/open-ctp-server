package combat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/service"
)

type recordingExecutor struct {
	request service.CommandRequest
	err     error
}

func (e *recordingExecutor) ExecuteCommand(_ context.Context, _ *auth.Session, request service.CommandRequest) (service.Result, error) {
	e.request = request
	if e.err != nil {
		return service.Result{}, e.err
	}
	return service.Result{Revision: 3, ReceiptID: "receipt-1", EventCursor: 7, Combat: &engine.CombatResult{Outcome: engine.CombatOutcomeAttackerWon, AttackerCasualties: 0, DefenderCasualties: 1, AttackerArmyID: "attackers", DefenderArmyID: "defenders"}, ViewerProjection: json.RawMessage(`{"revision":3}`)}, nil
}

func TestCombatHandlerDelegatesOnlyCanonicalCombatCommand(t *testing.T) {
	executor := &recordingExecutor{}
	handler := &Handler{executor: executor, sessions: &auth.SessionStore{}, authenticate: func(http.ResponseWriter, *http.Request, request) (*auth.Session, error) {
		return &auth.Session{AccountID: 7, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
	body := `{"session_id":1,"expected_revision":2,"idempotency_key":"idempotent","command_id":"combat-1","attacker_army_id":"attackers","defender_army_id":"defenders","roll":0,"outcome":"attacker_won"}`
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/combat", strings.NewReader(body)))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("client authority fields were accepted: %d %s", recorder.Code, recorder.Body.String())
	}

	body = `{"session_id":1,"expected_revision":2,"idempotency_key":"idempotent","command_id":"combat-1","attacker_army_id":"attackers","defender_army_id":"defenders"}`
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/combat", strings.NewReader(body)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
	}
	if executor.request.GameID != 9 || executor.request.Type != engine.CommandCombat || strings.Contains(string(executor.request.Payload), "roll") || strings.Contains(string(executor.request.Payload), "outcome") {
		t.Fatalf("adapter did not construct a canonical combat command: %#v", executor.request)
	}
	if strings.Contains(recorder.Body.String(), "defender-melee") || !strings.Contains(recorder.Body.String(), "event_cursor") || !strings.Contains(recorder.Body.String(), "receipt_id") {
		t.Fatalf("response was not viewer-safe or complete: %s", recorder.Body.String())
	}
}

func TestCombatHandlerHidesGatewayFailureDetails(t *testing.T) {
	executor := &recordingExecutor{err: &service.Error{Status: http.StatusUnprocessableEntity, Message: "hidden defender army 99"}}
	handler := &Handler{executor: executor, sessions: &auth.SessionStore{}, authenticate: func(http.ResponseWriter, *http.Request, request) (*auth.Session, error) {
		return &auth.Session{AccountID: 7, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/combat", strings.NewReader(`{"session_id":1,"expected_revision":2,"idempotency_key":"stale","command_id":"combat-1","attacker_army_id":"attackers","defender_army_id":"defenders"}`)))
	if recorder.Code != http.StatusUnprocessableEntity || strings.Contains(recorder.Body.String(), "defender army") {
		t.Fatalf("gateway error leaked target data: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestCombatHandlerRejectsNonPOSTAndMalformedPath(t *testing.T) {
	handler := &Handler{}
	for _, request := range []*http.Request{httptest.NewRequest(http.MethodGet, "/game/1/combat", nil), httptest.NewRequest(http.MethodPost, "/game/invalid/combat", nil)} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusMethodNotAllowed && recorder.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status %d", recorder.Code)
		}
	}
}
