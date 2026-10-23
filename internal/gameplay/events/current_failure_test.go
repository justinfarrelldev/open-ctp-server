package events

import (
	"errors"
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

func TestFiniteCurrentReadFailsBeforeCommittingSuccessHeaders(t *testing.T) {
	testInitialReadFailure(t, true)
}
func TestStreamInitialReadFailsBeforeCommittingSuccessHeaders(t *testing.T) {
	testInitialReadFailure(t, false)
}

func testInitialReadFailure(t *testing.T, current bool) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	handler := NewHandler(xdb, auth.NewSessionStore(xdb), ProjectorFuncs{}, nil)
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM sessions`).WithArgs(int64(22)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(22, 4, now, now.Add(time.Hour)))
	mock.ExpectQuery("SELECT id, account_id FROM game_players").WithArgs(int64(9), 4).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(42, 4))
	mode, timing := "stream", "initial_read;dur="
	if current {
		mock.ExpectQuery(regexp.QuoteMeta(snapshotForResyncQuery)).WithArgs(int64(9)).WillReturnError(errors.New("SECRET database detail"))
		mode, timing = "current", "snapshot;dur="
	} else {
		mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnError(errors.New("SECRET database detail"))
	}
	request := httptest.NewRequest(http.MethodGet, "/game/9/events?session_id=22", nil)
	if current {
		request.Header.Set("X-Gameplay-Resync", "current")
	}
	writer := streamRecorder{httptest.NewRecorder()}
	handler.ServeHTTP(writer, request)
	if writer.Code != http.StatusServiceUnavailable || strings.Contains(writer.Body.String(), "SECRET") {
		t.Fatalf("ambiguous empty-success snapshot: %d %s", writer.Code, writer.Body.String())
	}
	if writer.Header().Get("X-Gameplay-Read-Mode") != mode || !strings.Contains(writer.Header().Get("Server-Timing"), timing) {
		t.Fatalf("missing read diagnostics %v", writer.Header())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
