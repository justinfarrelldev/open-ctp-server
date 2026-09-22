package events

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type streamRecorder struct{ *httptest.ResponseRecorder }

type nonStreamingWriter struct {
	header http.Header
	status int
}

func (w *nonStreamingWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *nonStreamingWriter) Write(payload []byte) (int, error) { return len(payload), nil }
func (w *nonStreamingWriter) WriteHeader(status int)            { w.status = status }

func TestHandlerWritesProjectedTerminalOnce(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	sessions := auth.NewSessionStore(xdb)
	handler := NewHandler(xdb, sessions, ProjectorFuncs{Snapshot: func(_ Viewer, _ []byte) ([]byte, error) {
		return []byte(`{"terminal":"complete"}`), nil
	}}, nil)

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(int64(22), 4, now, now.Add(time.Hour)))
	mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(int64(42), int64(4)))
	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(nil))
	mock.ExpectQuery("SELECT game_id, sequence, revision, ordinal, event_type, payload_json").WithArgs(int64(9), int64(0)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "sequence", "revision", "ordinal", "event_type", "payload_json"}))
	mock.ExpectQuery("SELECT last_event_sequence, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(int64(7), `{"terminal":{"ended":true,"winning_player_id":"hidden-opponent"}}`))

	req := httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil)
	writer := streamRecorder{httptest.NewRecorder()}
	handler.ServeHTTP(writer, req)
	body := writer.Body.String()
	if strings.Count(body, "event: terminal") != 1 || strings.Contains(body, "hidden-opponent") || !strings.Contains(body, `{"terminal":"complete"}`) {
		t.Fatalf("terminal SSE must be projected exactly once: %q", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func (streamRecorder) Flush() {}

func TestHandlerWritesOnlyProjectedSSEBytes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	sessions := auth.NewSessionStore(xdb)
	handler := NewHandler(xdb, sessions, ProjectorFuncs{
		Event:    func(_ Viewer, _ Event) ([]byte, error) { return []byte(`{"visible":"movement"}`), nil },
		Snapshot: func(_ Viewer, _ []byte) ([]byte, error) { return []byte(`{"visible":"snapshot"}`), nil },
	}, nil)

	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(int64(22), 4, now, now.Add(time.Hour)))
	mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(int64(42), int64(4)))
	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(int64(1)))
	mock.ExpectQuery(regexp.QuoteMeta(snapshotForResyncQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "canonical_state_json"}).AddRow(int64(1), `{"terminal":{"ended":false}}`))
	mock.ExpectQuery("SELECT last_event_sequence, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(int64(1), `{"terminal":{"ended":false}}`))

	req := httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil)
	writer := streamRecorder{httptest.NewRecorder()}
	handler.ServeHTTP(writer, req)
	body := writer.Body.String()
	if writer.Code != http.StatusOK || writer.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("status=%d headers=%v", writer.Code, writer.Header())
	}
	if !strings.Contains(body, "id: 1\nevent: resync\ndata: {\"current_revision\":1,\"snapshot\":{\"visible\":\"snapshot\"}}\n\n") {
		t.Fatalf("unexpected SSE body: %q", body)
	}
	if strings.Contains(body, "enemy-99") || strings.Contains(body, "coordinates") {
		t.Fatalf("SSE leaked canonical bytes: %q", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBrokerScopesWakeupsAndUnsubscribesIdempotently(t *testing.T) {
	var nilBroker *Broker
	if wakeups, unsubscribe := nilBroker.Subscribe(1); wakeups != nil {
		t.Fatal("nil broker returned a wakeup channel")
	} else {
		unsubscribe()
		nilBroker.Publish(1)
	}

	broker := NewBroker()
	first, unsubscribeFirst := broker.Subscribe(7)
	second, unsubscribeSecond := broker.Subscribe(8)
	if broker.subscriberCount(7) != 1 || broker.subscriberCount(8) != 1 {
		t.Fatalf("subscriber counts = %d/%d, want 1/1", broker.subscriberCount(7), broker.subscriberCount(8))
	}
	broker.Publish(7)
	broker.Publish(7)
	select {
	case <-first:
	default:
		t.Fatal("matching subscriber did not receive a wakeup")
	}
	select {
	case <-second:
		t.Fatal("different game subscriber received a wakeup")
	default:
	}

	unsubscribeFirst()
	unsubscribeFirst()
	if broker.subscriberCount(7) != 0 {
		t.Fatalf("game 7 subscriber count = %d, want 0", broker.subscriberCount(7))
	}
	if _, open := <-first; open {
		t.Fatal("unsubscribed wakeup channel remained open")
	}
	unsubscribeSecond()
	if broker.subscriberCount(8) != 0 {
		t.Fatalf("game 8 subscriber count = %d, want 0", broker.subscriberCount(8))
	}
}

func TestHandlerRejectsInvalidRequestsBeforeStreaming(t *testing.T) {
	tests := []struct {
		name    string
		handler *Handler
		method  string
		path    string
		status  int
	}{
		{name: "method", handler: &Handler{}, method: http.MethodPost, path: "/game/9/events", status: http.StatusBadRequest},
		{name: "path", handler: &Handler{}, method: http.MethodGet, path: "/game/events", status: http.StatusBadRequest},
		{name: "unavailable", handler: &Handler{}, method: http.MethodGet, path: "/game/9/events", status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := httptest.NewRecorder()
			test.handler.ServeHTTP(writer, httptest.NewRequest(test.method, test.path, nil))
			if writer.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%q", writer.Code, test.status, writer.Body.String())
			}
		})
	}
}

func TestEventPathAndTerminalStateValidation(t *testing.T) {
	if gameID, err := gameIDFromEventPath("/game/42/events/"); err != nil || gameID != 42 {
		t.Fatalf("valid path game ID = %d, err=%v", gameID, err)
	}
	for _, path := range []string{"/game/events", "/games/42/events", "/game/0/events", "/game/not-a-number/events"} {
		if _, err := gameIDFromEventPath(path); err == nil {
			t.Fatalf("path %q was accepted", path)
		}
	}

	projector := ProjectorFuncs{Snapshot: func(_ Viewer, _ []byte) ([]byte, error) {
		return []byte(`{"terminal":"public"}`), nil
	}}
	if _, terminal, err := terminalFromState(Viewer{}, []byte(`{"terminal":{"ended":false}}`), projector); err != nil || terminal {
		t.Fatalf("nonterminal state returned terminal=%v err=%v", terminal, err)
	}
	event, terminal, err := terminalFromState(Viewer{}, []byte(`{"terminal":{"ended":true}}`), projector)
	if err != nil || !terminal || string(event.Data) != `{"terminal":"public"}` {
		t.Fatalf("terminal event = %#v terminal=%v err=%v", event, terminal, err)
	}
	if _, _, err := terminalFromState(Viewer{}, []byte(`{"terminal":`), projector); err == nil {
		t.Fatal("malformed terminal state was accepted")
	}
}

func TestHandlerWritesCurrentProjectedResync(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := NewHandler(sqlx.NewDb(db, "sqlmock"), auth.NewSessionStore(sqlx.NewDb(db, "sqlmock")), ProjectorFuncs{Snapshot: func(_ Viewer, _ []byte) ([]byte, error) {
		return []byte(`{"visible":"current"}`), nil
	}}, nil)
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(int64(22), 4, now, now.Add(time.Hour)))
	mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(int64(42), int64(4)))
	mock.ExpectQuery(regexp.QuoteMeta(snapshotForResyncQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "canonical_state_json"}).AddRow(int64(3), `{"terminal":{"ended":false}}`))
	mock.ExpectQuery("SELECT last_event_sequence, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(int64(3), `{"terminal":{"ended":false}}`))

	request := httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil)
	request.Header.Set("X-Gameplay-Resync", "current")
	writer := streamRecorder{httptest.NewRecorder()}
	handler.ServeHTTP(writer, request)
	body := writer.Body.String()
	if writer.Code != http.StatusOK || !strings.Contains(body, "event: resync") || !strings.Contains(body, `{"visible":"current"}`) {
		t.Fatalf("current resync status=%d body=%q", writer.Code, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandlerRejectsNonmemberAndNonStreamingWriter(t *testing.T) {
	t.Run("nonmember", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		xdb := sqlx.NewDb(db, "sqlmock")
		handler := NewHandler(xdb, auth.NewSessionStore(xdb), ProjectorFuncs{}, nil)
		now := time.Now()
		mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(int64(22), 4, now, now.Add(time.Hour)))
		mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}))
		writer := httptest.NewRecorder()
		handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil))
		if writer.Code != http.StatusForbidden {
			t.Fatalf("nonmember status = %d, want 403", writer.Code)
		}
	})

	t.Run("writer without flusher", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		xdb := sqlx.NewDb(db, "sqlmock")
		handler := NewHandler(xdb, auth.NewSessionStore(xdb), ProjectorFuncs{}, nil)
		now := time.Now()
		mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(int64(22), 4, now, now.Add(time.Hour)))
		mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(int64(42), int64(4)))
		writer := &nonStreamingWriter{}
		handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil))
		if writer.status != http.StatusInternalServerError {
			t.Fatalf("non-streaming writer status = %d, want 500", writer.status)
		}
	})
}
