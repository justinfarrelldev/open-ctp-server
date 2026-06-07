package lobby

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type flushingResponseRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (r *flushingResponseRecorder) Flush() {
	r.flushed = true
}

type plainResponseWriter struct {
	header http.Header
	code   int
	body   strings.Builder
}

func (w *plainResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *plainResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func (w *plainResponseWriter) Write(body []byte) (int, error) {
	if w.code == 0 {
		w.code = http.StatusOK
	}
	return w.body.Write(body)
}

func TestStatusTrackingResponseWriter_WriteSetsStatusAndBody(t *testing.T) {
	base := httptest.NewRecorder()
	tracked := &statusTrackingResponseWriter{ResponseWriter: base}

	n, err := tracked.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len("hello") {
		t.Fatalf("Write() bytes = %d want %d", n, len("hello"))
	}
	if !tracked.wroteHeader || !tracked.wroteBody {
		t.Fatalf("writer flags = %+v want header/body recorded", tracked)
	}
	if tracked.statusCode != http.StatusOK {
		t.Fatalf("statusCode = %d want %d", tracked.statusCode, http.StatusOK)
	}
	if base.Code != http.StatusOK {
		t.Fatalf("recorder status = %d want %d", base.Code, http.StatusOK)
	}
	if body := base.Body.String(); body != "hello" {
		t.Fatalf("body = %q want %q", body, "hello")
	}
}

func TestStatusTrackingResponseWriter_Flush(t *testing.T) {
	t.Run("sets implicit ok before flushing", func(t *testing.T) {
		base := &flushingResponseRecorder{ResponseRecorder: httptest.NewRecorder()}
		tracked := &statusTrackingResponseWriter{ResponseWriter: base}

		tracked.Flush()

		if !base.flushed {
			t.Fatal("expected Flush to be forwarded")
		}
		if !tracked.wroteHeader {
			t.Fatal("expected Flush to mark header written")
		}
		if tracked.statusCode != http.StatusOK {
			t.Fatalf("statusCode = %d want %d", tracked.statusCode, http.StatusOK)
		}
	})

	t.Run("ignores non flusher response writers", func(t *testing.T) {
		base := &plainResponseWriter{}
		tracked := &statusTrackingResponseWriter{ResponseWriter: base}

		tracked.Flush()

		if tracked.wroteHeader {
			t.Fatalf("expected non-flusher path to leave header untouched, got %+v", tracked)
		}
		if base.code != 0 {
			t.Fatalf("response writer code = %d want 0", base.code)
		}
	})
}

func TestHandleLobbyHandlerError(t *testing.T) {
	t.Run("writes internal server error when header not written", func(t *testing.T) {
		base := httptest.NewRecorder()
		tracked := &statusTrackingResponseWriter{ResponseWriter: base}

		handleLobbyHandlerError(tracked, errors.New("boom"))

		if base.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d want %d", base.Code, http.StatusInternalServerError)
		}
		if body := base.Body.String(); !strings.Contains(body, "boom") {
			t.Fatalf("body = %q want boom", body)
		}
	})

	t.Run("appends error body when status already written", func(t *testing.T) {
		base := httptest.NewRecorder()
		tracked := &statusTrackingResponseWriter{ResponseWriter: base}
		tracked.WriteHeader(http.StatusBadRequest)

		handleLobbyHandlerError(tracked, errors.New("bad request detail"))

		if base.Code != http.StatusBadRequest {
			t.Fatalf("status = %d want %d", base.Code, http.StatusBadRequest)
		}
		if body := base.Body.String(); body != "bad request detail" {
			t.Fatalf("body = %q want %q", body, "bad request detail")
		}
	})

	t.Run("does nothing when body already written", func(t *testing.T) {
		base := httptest.NewRecorder()
		tracked := &statusTrackingResponseWriter{ResponseWriter: base, wroteBody: true}

		handleLobbyHandlerError(tracked, errors.New("ignored"))

		if body := base.Body.String(); body != "" {
			t.Fatalf("body = %q want empty", body)
		}
	})
}

func TestLobbyHandlers_WrapperCoverage(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	tests := []struct {
		name        string
		handler     func(http.ResponseWriter, *http.Request, *sqlx.DB, *auth.SessionStore)
		method      string
		target      string
		body        string
		wantStatus  int
		wantBodySub string
	}{
		{
			name:        "get lobby bad request",
			handler:     GetLobbyHandler,
			method:      http.MethodGet,
			target:      "/lobby/get_lobby",
			body:        `{}`,
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "lobby_id must be specified",
		},
		{
			name:        "update lobby bad request",
			handler:     UpdateLobbyHandler,
			method:      http.MethodPut,
			target:      "/lobby/update_lobby",
			body:        `{"lobby":{},"session_id":1}`,
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "lobby_id must be specified",
		},
		{
			name:        "delete lobby bad request",
			handler:     DeleteLobbyHandler,
			method:      http.MethodDelete,
			target:      "/lobby/delete_lobby",
			body:        `{"session_id":1}`,
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "lobby_id must be specified",
		},
		{
			name:        "leave lobby unauthorized",
			handler:     LeaveLobbyHandler,
			method:      http.MethodPost,
			target:      "/lobby/leave_lobby",
			body:        `{"session_id":0}`,
			wantStatus:  http.StatusUnauthorized,
			wantBodySub: auth.ErrInvalidBodySessionID.Error(),
		},
		{
			name:        "connect lobby bad request",
			handler:     ConnectLobbyHandler,
			method:      http.MethodGet,
			target:      "/lobby/connect_lobby",
			body:        "",
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "lobby_id must be specified",
		},
		{
			name:        "send chat unauthorized",
			handler:     SendChatHandler,
			method:      http.MethodPost,
			target:      "/lobby/send_chat",
			body:        `{"message":"hello","session_id":0}`,
			wantStatus:  http.StatusUnauthorized,
			wantBodySub: auth.ErrInvalidBodySessionID.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			tt.handler(rr, req, sqlxDB, store)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d want %d", rr.Code, tt.wantStatus)
			}
			if body := rr.Body.String(); !strings.Contains(body, tt.wantBodySub) {
				t.Fatalf("body = %q want substring %q", body, tt.wantBodySub)
			}
		})
	}
}
