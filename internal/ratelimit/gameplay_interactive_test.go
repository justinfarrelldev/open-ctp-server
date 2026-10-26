package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/didip/tollbooth/v7"
)

func TestGameplayLimiterAllowsInteractiveTurnsAndStillBoundsFloods(t *testing.T) {
	limiter := NewGameplayCommandLimiter()
	handler := tollbooth.LimitFuncHandler(limiter, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	send := func() int {
		request := httptest.NewRequest(http.MethodPost, "/game/44/pass", nil)
		request.RemoteAddr = "127.0.0.1:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Code
	}
	for range 12 {
		if status := send(); status != http.StatusOK {
			t.Fatalf("interactive end-turn sequence hit legacy two-per-second budget: %d", status)
		}
	}
	limited := false
	for range 100 {
		if send() == http.StatusTooManyRequests {
			limited = true
		}
	}
	if !limited {
		t.Fatal("interactive admission must still reject floods")
	}
	if limiter.GetMax() != 30 {
		t.Fatalf("unexpected interactive command budget: %v", limiter.GetMax())
	}
}
