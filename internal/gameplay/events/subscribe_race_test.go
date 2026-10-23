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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := NewHandler(xdb, auth.NewSessionStore(xdb), ProjectorFuncs{Snapshot: func(_ Viewer, _ []byte) ([]byte, error) {
		// A commit here is after the initial durable read, but before the old
		// implementation subscribed: its wakeup was lost until another command.
		if broker.subscriberCount(9) != 1 {
			t.Error("initial read raced ahead of wakeup subscription")
		}
		broker.Publish(9)
		cancel()
		return []byte(`{}`), nil
	}}, broker)
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(22, 4, now, now.Add(time.Hour)))
	mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(42, 4))
	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(snapshotForResyncQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "canonical_state_json"}).AddRow(1, `{}`))
	writer := streamRecorder{httptest.NewRecorder()}
	handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil).WithContext(ctx))
	if broker.subscriberCount(9) != 0 {
		t.Fatal("subscription leaked on disconnect")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
