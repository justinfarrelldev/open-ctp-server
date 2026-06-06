package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/didip/tollbooth/v7"
)

func TestLimiterClassesHaveExpectedRates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		limiter func() interface {
			GetMax() float64
			GetMessage() string
			GetMessageContentType() string
		}
		expected float64
	}{
		{name: "auth write", limiter: func() interface {
			GetMax() float64
			GetMessage() string
			GetMessageContentType() string
		} { return NewAuthWriteLimiter() }, expected: 10.0 / 60.0},
		{name: "account", limiter: func() interface {
			GetMax() float64
			GetMessage() string
			GetMessageContentType() string
		} { return NewAccountLimiter() }, expected: 1.0},
		{name: "lobby game", limiter: func() interface {
			GetMax() float64
			GetMessage() string
			GetMessageContentType() string
		} { return NewLobbyGameLimiter() }, expected: 2.0},
		{name: "lobby sse", limiter: func() interface {
			GetMax() float64
			GetMessage() string
			GetMessageContentType() string
		} { return NewLobbySSELimiter() }, expected: 20.0 / 60.0},
		{name: "health", limiter: func() interface {
			GetMax() float64
			GetMessage() string
			GetMessageContentType() string
		} { return NewHealthLimiter() }, expected: 5.0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lmt := tt.limiter()
			if got := lmt.GetMax(); got != tt.expected {
				t.Fatalf("unexpected limiter max: got %v want %v", got, tt.expected)
			}
			if got := lmt.GetMessageContentType(); got != defaultMessageContentType {
				t.Fatalf("unexpected content type: got %q want %q", got, defaultMessageContentType)
			}
			if got := lmt.GetMessage(); got != defaultMessageJSON() {
				t.Fatalf("unexpected message: got %q want %q", got, defaultMessageJSON())
			}
		})
	}
}

func TestLimitFuncHandlerReturnsJSON429(t *testing.T) {
	limiter := tollbooth.NewLimiter(1, nil)
	limiter.SetMessageContentType(defaultMessageContentType)
	limiter.SetMessage(defaultMessageJSON())
	limiter.SetTokenBucketExpirationTTL(time.Hour)
	limiter.SetBasicAuthExpirationTTL(time.Hour)
	limiter.SetHeaderEntryExpirationTTL(time.Hour)

	handler := tollbooth.LimitFuncHandler(limiter, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	first := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	handler.ServeHTTP(first, req1)

	if first.Code != http.StatusOK {
		t.Fatalf("first request should pass: got %d want %d", first.Code, http.StatusOK)
	}

	second := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	handler.ServeHTTP(second, req2)

	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should be limited: got %d want %d", second.Code, http.StatusTooManyRequests)
	}

	if got := second.Header().Get("Content-Type"); !strings.HasPrefix(got, defaultMessageContentType) {
		t.Fatalf("unexpected content type: got %q want prefix %q", got, defaultMessageContentType)
	}

	expectedBody := defaultMessageJSON()
	if got := strings.TrimSpace(second.Body.String()); got != expectedBody {
		t.Fatalf("unexpected body: got %q want %q", got, expectedBody)
	}
}
