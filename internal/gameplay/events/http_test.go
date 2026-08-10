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
