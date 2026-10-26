package runtime

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAIProgressDoesNotWaitForGlobalOutboxDispatch(t *testing.T) {
	runtime := New(nil, Config{})
	if runtime.AIWorker.AfterCommit != nil {
		t.Fatal("AI drive still waits for unrelated durable outbox dispatch instead of the immediate commit hint")
	}
	if runtime.Dispatcher == nil {
		t.Fatal("background durable retry dispatcher must remain configured")
	}
}

func TestSnapshotSubscriptionBudgetIsIndependentOfBootstrapAndLegacyReads(t *testing.T) {
	runtime := New(nil, Config{})
	read := func(snapshot bool) int {
		req := httptest.NewRequest(http.MethodGet, "/game/44/events", nil) // missing session rejects before DB access
		if snapshot {
			req.Header.Set("X-Gameplay-Stream", "snapshots-v1")
		}
		response := httptest.NewRecorder()
		runtime.Handler.ServeHTTP(response, req)
		return response.Code
	}
	if got := read(false); got == http.StatusTooManyRequests {
		t.Fatal("first legacy request throttled")
	}
	if got := read(false); got != http.StatusTooManyRequests {
		t.Fatalf("legacy connection budget changed: %d", got)
	}
	if got := read(true); got == http.StatusTooManyRequests {
		t.Fatal("bootstrap consumed the initial snapshot subscription token")
	}
	if got := read(true); got != http.StatusTooManyRequests {
		t.Fatalf("snapshot reconnect budget missing: %d", got)
	}
}
