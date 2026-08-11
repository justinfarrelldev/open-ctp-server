package gameplayhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/service"
)

type recordingExecutor struct {
	request service.CommandRequest
	err     error
	calls   int
}

func (e *recordingExecutor) ExecuteCommand(_ context.Context, _ *auth.Session, request service.CommandRequest) (service.Result, error) {
	e.calls++
	e.request = request
	if e.err != nil {
		return service.Result{}, e.err
	}
	return service.Result{Revision: 3, ReceiptID: "receipt-1", EventCursor: 7, ViewerProjection: json.RawMessage(`{"tiles":[{"id":"visible"}]}`)}, nil
}

func authenticatedHandler(executor *recordingExecutor) *Handler {
	return &Handler{executor: executor, authenticate: func(http.ResponseWriter, *http.Request, request) (*auth.Session, error) {
		return &auth.Session{AccountID: 7, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
}

func TestHandlerDelegatesCommandWithServerDerivedIdentity(t *testing.T) {
	executor := &recordingExecutor{}
	handler := authenticatedHandler(executor)
	body := `{"session_id":1,"expected_revision":2,"idempotency_key":"replay-key","command_id":"move-1","unit_id":"unit-a","path":["tile-b"]}`
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/move", strings.NewReader(body)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d %s", recorder.Code, recorder.Body.String())
	}
	if executor.request.GameID != 9 || executor.request.Type != engine.CommandMoveUnit || strings.Contains(string(executor.request.Payload), "actor") {
		t.Fatalf("handler did not issue the canonical command: %#v", executor.request)
	}
	if !strings.Contains(recorder.Body.String(), "receipt_id") || strings.Contains(recorder.Body.String(), "hidden") {
		t.Fatalf("response was incomplete or leaked canonical data: %s", recorder.Body.String())
	}
}

func TestHandlerRejectsUnknownAuthorityAndHidesGatewayDetails(t *testing.T) {
	executor := &recordingExecutor{err: &service.Error{Status: http.StatusUnprocessableEntity, Message: "hidden foreign unit target"}}
	handler := authenticatedHandler(executor)
	for _, body := range []string{
		`{"session_id":1,"expected_revision":2,"idempotency_key":"key","command_id":"move","unit_id":"unit-a","path":["tile-b"],"actor_player_id":"victim"}`,
		`{"session_id":1,"expected_revision":2,"idempotency_key":"key","command_id":"move","unit_id":"unit-a","path":["tile-b"]}`,
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/move", strings.NewReader(body)))
		if strings.Contains(body, "actor_player_id") {
			if recorder.Code != http.StatusBadRequest || executor.calls != 0 {
				t.Fatalf("authority field was accepted: %d %s", recorder.Code, recorder.Body.String())
			}
			continue
		}
		if recorder.Code != http.StatusUnprocessableEntity || strings.Contains(recorder.Body.String(), "foreign unit") {
			t.Fatalf("gateway detail leaked: %d %s", recorder.Code, recorder.Body.String())
		}
	}
}

func TestHandlerPreservesStaleAndIdempotencyConflictsWithoutDetails(t *testing.T) {
	for _, message := range []string{"gameplay revision is stale", "idempotency key was already used with a different payload"} {
		executor := &recordingExecutor{err: &service.Error{Status: http.StatusConflict, Message: message}}
		recorder := httptest.NewRecorder()
		authenticatedHandler(executor).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/move", strings.NewReader(`{"session_id":1,"expected_revision":2,"idempotency_key":"key","command_id":"move","unit_id":"unit-a","path":["tile-b"]}`)))
		if recorder.Code != http.StatusConflict || strings.Contains(recorder.Body.String(), "revision") || strings.Contains(recorder.Body.String(), "idempotency") {
			t.Fatalf("conflict was not safely projected: %d %s", recorder.Code, recorder.Body.String())
		}
	}
}

func TestHandlerRateLimitRejectsBeforeGateway(t *testing.T) {
	executor := &recordingExecutor{}
	limiter := tollbooth.NewLimiter(1, nil)
	limited := tollbooth.LimitFuncHandler(limiter, authenticatedHandler(executor).ServeHTTP)
	for attempt := 0; attempt < 2; attempt++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/game/9/move", strings.NewReader(`{"session_id":1,"expected_revision":2,"idempotency_key":"key","command_id":"move","unit_id":"unit-a","path":["tile-b"]}`))
		request.RemoteAddr = "127.0.0.1:1234"
		limited.ServeHTTP(recorder, request)
		if attempt == 0 && recorder.Code != http.StatusOK {
			t.Fatalf("threshold request failed: %d", recorder.Code)
		}
		if attempt == 1 && (recorder.Code != http.StatusTooManyRequests || executor.calls != 1) {
			t.Fatalf("limited request reached gateway: %d calls=%d", recorder.Code, executor.calls)
		}
	}
}

func TestHandlerRejectsUnauthenticatedAndMalformedRequests(t *testing.T) {
	handler := &Handler{executor: &recordingExecutor{}, authenticate: func(w http.ResponseWriter, _ *http.Request, _ request) (*auth.Session, error) {
		writeError(w, http.StatusUnauthorized)
		return nil, context.Canceled
	}}
	for _, body := range []string{"{", `{"session_id":1,"expected_revision":2,"idempotency_key":"key","command_id":"move","unit_id":"unit-a","path":["tile-b"]}`} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/move", strings.NewReader(body)))
		if recorder.Code != http.StatusBadRequest && recorder.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected status: %d %s", recorder.Code, recorder.Body.String())
		}
	}
}
