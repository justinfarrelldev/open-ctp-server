package lobby

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetLobby_NotFound(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lobbySelectByID)).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodGet, "/lobby/get_lobby", strings.NewReader(`{"lobby_id":1}`))
	rr := httptest.NewRecorder()

	err := GetLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusNotFound {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetLobby_Success(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lobbySelectByID)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(int64(1), "Test Lobby", "Owner", int64(7), false, false, true, now, now, now, nil))
	expectLobbyMemberCount(mock, int64(1), 1)
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodGet, "/lobby/get_lobby", strings.NewReader(`{"lobby_id":1}`))
	rr := httptest.NewRecorder()

	err := GetLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}
}
