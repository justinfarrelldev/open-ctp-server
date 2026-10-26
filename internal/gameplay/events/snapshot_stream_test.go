package events

import (
	"context"
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

type snapshotRecorder struct {
	*httptest.ResponseRecorder
	flushes int
	cancel  context.CancelFunc
}

func (w *snapshotRecorder) Flush() {
	w.flushes++
	if w.flushes == 4 {
		w.cancel()
	}
}

// A negotiated persistent subscription receives complete viewer snapshots on the
// same connection; it must never require a second limited HTTP current read.
func TestSnapshotStreamReadsCurrentOnSubscribeAndWakeWithoutCanonicalLeak(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	broker := NewBroker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	projections := 0
	handler := NewHandler(xdb, auth.NewSessionStore(xdb), ProjectorFuncs{Snapshot: func(viewer Viewer, raw []byte) ([]byte, error) {
		if viewer.AccountID != 4 || viewer.GameID != 9 {
			t.Fatalf("wrong viewer: %+v", viewer)
		}
		if broker.subscriberCount(9) != 1 {
			t.Fatal("snapshot read before subscribing loses commits")
		}
		projections++
		if projections == 1 {
			broker.Publish(9)
			return []byte(`{"revision":1,"turn":1,"tiles":[]}`), nil
		}
		return []byte(`{"revision":2,"turn":2,"tiles":[]}`), nil
	}}, broker)
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(22, 4, now, now.Add(time.Hour)))
	mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(42, 4))
	for revision := 1; revision <= 2; revision++ {
		mock.ExpectQuery(regexp.QuoteMeta(snapshotForResyncQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "last_event_sequence", "canonical_state_json"}).AddRow(revision, revision*10, `{"terminal":{"ended":false},"secret":"hidden-enemy"}`))
		mock.ExpectQuery(regexp.QuoteMeta(terminalStateQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(revision*10, `{"terminal":{"ended":false}}`))
	}
	req := httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil).WithContext(ctx)
	req.Header.Set("X-Gameplay-Stream", "snapshots-v1")
	writer := &snapshotRecorder{ResponseRecorder: httptest.NewRecorder(), cancel: cancel}
	handler.ServeHTTP(writer, req)
	body := writer.Body.String()
	if writer.Code != 200 || projections != 2 || strings.Count(body, "event: resync") != 2 || !strings.Contains(body, `"event_cursor":20`) || !strings.Contains(body, `"current_revision":2`) || strings.Contains(body, "hidden-enemy") {
		t.Fatalf("snapshot subscription status=%d projections=%d body=%s", writer.Code, projections, body)
	}
	if writer.Header().Get("X-Gameplay-Stream") != "snapshots-v1" {
		t.Fatal("negotiated stream mode not acknowledged")
	}
	if broker.subscriberCount(9) != 0 {
		t.Fatal("subscription leaked after cancellation")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
