package events

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func TestStreamSubscribesBeforeInitialDurableRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	broker := NewBroker()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reads := 0
	observe := func() ([]byte, error) {
		// A commit here is after the initial durable read, but before the old
		// implementation subscribed: its wakeup was lost until another command.
		if broker.subscriberCount(9) != 1 {
			t.Error("initial read raced ahead of wakeup subscription")
		}
		reads++
		if reads == 1 {
			broker.Publish(9)
		} else {
			cancel()
		}
		return []byte(`{}`), nil
	}
	handler := NewHandler(xdb, auth.NewSessionStore(xdb), ProjectorFuncs{Snapshot: func(Viewer, []byte) ([]byte, error) { return observe() }, Event: func(Viewer, Event) ([]byte, error) { return observe() }}, broker)
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(22, 4, now, now.Add(time.Hour)))
	mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(42, 4))
	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(snapshotForResyncQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "last_event_sequence", "canonical_state_json"}).AddRow(1, 1, `{}`))
	mock.ExpectQuery(regexp.QuoteMeta(terminalStateQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(1, `{"terminal":{"ended":false}}`))
	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(eventsAfterCursorQuery)).WithArgs(int64(9), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "sequence", "revision", "ordinal", "event_type", "payload_json"}).AddRow(9, 2, 2, 0, "unit_moved", `{}`))
	writer := streamRecorder{httptest.NewRecorder()}
	handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil).WithContext(ctx))
	if reads != 2 {
		t.Fatalf("projection-time commit wasn't reread: %d", reads)
	}
	if broker.subscriberCount(9) != 0 {
		t.Fatal("subscription leaked on disconnect")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
