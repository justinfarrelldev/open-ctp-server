package gameplayhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/service"
)

func TestCommandRejectionHasSafeCodeAndCorrelation(t *testing.T) {
	executor := &recordingExecutor{err: &service.Error{Status: 422, Message: "SECRET hidden wonder / session credential"}}
	recorder := httptest.NewRecorder()
	authenticatedHandler(executor).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/game/9/edit-production", strings.NewReader(`{"session_id":123,"command_id":"secret-input","idempotency_key":"key","expected_revision":2,"city_id":"home","entries":[]}`)))
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "command_invalid" || body["correlation_id"] == "" || body["message"] == "" || body["error"] != "gameplay request rejected" {
		t.Fatalf("opaque rejection: %v", body)
	}
	if recorder.Header().Get("X-Gameplay-Correlation-ID") != body["correlation_id"] {
		t.Fatal("missing correlation header")
	}
	if strings.Contains(recorder.Body.String(), "SECRET") || strings.Contains(recorder.Body.String(), "secret-input") {
		t.Fatal("private details leaked")
	}
}
