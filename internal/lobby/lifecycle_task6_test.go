package lobby

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func newLobbyTestStore(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *auth.SessionStore, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	resetLobbyBroker()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	cleanup := func() {
		resetLobbyBroker()
		_ = db.Close()
	}

	return sqlxDB, mock, store, cleanup
}

func lobbySessionRows(sessionID int64, accountID int64, expiresAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
		AddRow(sessionID, accountID, time.Now(), expiresAt)
}

func expectExpiredCurrentLobbyLookup(mock sqlmock.Sqlmock, accountID int64, lobbyID int64) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}).AddRow(lobbyID))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Expired Current Lobby", "Owner", accountID, false, false, true, time.Now(), time.Now(), time.Now().Add(-2*time.Hour-time.Minute), nil))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET is_closed = TRUE, closed_at = NOW(), updated_at = NOW() WHERE id = $1 AND closed_at IS NULL")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func TestCreateLobby_RejectsInvalidSession(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodPost, "/lobby/create_lobby", strings.NewReader(`{"lobby":{"name":"Test Lobby"},"password":"password123","session_id":0}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	err = CreateLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusUnauthorized)
	}

	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body for direct endpoint call, got %q", strings.TrimSpace(rr.Body.String()))
	}
}

func TestCreateLobby_RejectsSecondLobbyForAccount(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(44)
	accountID := int64(7)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}).AddRow(99))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(int64(99), "Current Lobby", "Owner", accountID, false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectRollback()

	body := `{"lobby":{"name":"Another Lobby","is_muted":false,"is_public":true},"password":"password123","session_id":44}`
	req, err := http.NewRequest(http.MethodPost, "/lobby/create_lobby", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = CreateLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusConflict {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusConflict)
	}
}

func TestCreateLobby_AllowsFreshCreateAfterExpiredCurrentLobbyIsClosed(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(145)
	accountID := int64(17)
	expiredLobbyID := int64(99)
	newLobbyName := "Recovered Lobby"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectExpiredCurrentLobbyLookup(mock, accountID, expiredLobbyID)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT name FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Recovered Owner"))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO lobbies (name, owner_name, owner_account_id, is_closed, is_muted, is_public, updated_at, last_activity_at)
VALUES ($1, $2, $3, FALSE, $4, $5, NOW(), NOW())
RETURNING id`)).
		WithArgs(newLobbyName, "Recovered Owner", accountID, false, true).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(150)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)")).
		WithArgs(int64(150), accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := `{"lobby":{"name":"Recovered Lobby","is_public":true},"password":"password123","session_id":145}`
	req, err := http.NewRequest(http.MethodPost, "/lobby/create_lobby", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = CreateLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rr.Code != http.StatusCreated {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusCreated)
	}
}

func TestUpdateLobby_OwnerCanRenameAndToggleFlags(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(51)
	accountID := int64(9)
	lobbyID := int64(12)
	name := "Renamed Lobby"
	muted := true
	public := false

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Old Lobby", "Owner", accountID, false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET name = $1, is_muted = $2, is_public = $3, updated_at = NOW(), last_activity_at = NOW() WHERE id = $4")).
		WithArgs(name, muted, public, lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	payload := UpdateLobbyArgs{LobbyId: &lobbyID, SessionId: &sessionID, Lobby: &LobbyParam{Name: &name, IsMuted: &muted, IsPublic: &public}}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPut, "/lobby/update_lobby", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = UpdateLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}
}

func TestUpdateLobby_RejectsNonOwner(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(52)
	accountID := int64(20)
	lobbyID := int64(12)
	name := "Renamed Lobby"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Old Lobby", "Owner", int64(99), false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectRollback()

	payload := UpdateLobbyArgs{LobbyId: &lobbyID, SessionId: &sessionID, Lobby: &LobbyParam{Name: &name}}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPut, "/lobby/update_lobby", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = UpdateLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusForbidden {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusForbidden)
	}
}

func TestDeleteLobby_RejectsNonOwner(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(60)
	accountID := int64(5)
	lobbyID := int64(14)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Old Lobby", "Owner", int64(77), false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectRollback()

	req, err := http.NewRequest(http.MethodDelete, "/lobby/delete_lobby", strings.NewReader(`{"lobby_id":14,"session_id":60}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = DeleteLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusForbidden {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusForbidden)
	}
}

func TestJoinLobby_ClosedLobbyReturnsGone(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(70)
	accountID := int64(8)
	lobbyID := int64(21)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Closed Lobby", "Owner", int64(99), true, false, true, time.Now(), time.Now(), time.Now().Add(-2*time.Hour), time.Now().Add(-time.Minute)))
	mock.ExpectRollback()

	req, err := http.NewRequest(http.MethodPost, "/lobby/join_lobby", strings.NewReader(`{"lobby_id":21,"session_id":70}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = JoinLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusGone {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusGone)
	}
}

func TestJoinLobby_EmptyLobbyInactiveAtThirtyMinutesReturnsGone(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(702)
	accountID := int64(80)
	lobbyID := int64(210)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Empty Stale Lobby", "Owner", int64(99), false, false, true, time.Now(), time.Now(), time.Now().Add(-31*time.Minute), nil))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET is_closed = TRUE, closed_at = NOW(), updated_at = NOW() WHERE id = $1 AND closed_at IS NULL")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	req, err := http.NewRequest(http.MethodPost, "/lobby/join_lobby", strings.NewReader(`{"lobby_id":210,"session_id":702}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = JoinLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusGone {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusGone)
	}
}

func TestJoinLobby_SucceedsForOpenLobby(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(71)
	accountID := int64(8)
	lobbyID := int64(22)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Open Lobby", "Owner", int64(99), false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)")).
		WithArgs(lobbyID, accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodPost, "/lobby/join_lobby", strings.NewReader(`{"lobby_id":22,"session_id":71}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = JoinLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}
}

func TestJoinLobby_AllowsTargetJoinAfterExpiredCurrentLobbyIsClosed(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(146)
	accountID := int64(18)
	expiredLobbyID := int64(100)
	targetLobbyID := int64(101)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectExpiredCurrentLobbyLookup(mock, accountID, expiredLobbyID)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(targetLobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(targetLobbyID, "Fresh Target Lobby", "Owner", int64(77), false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)")).
		WithArgs(targetLobbyID, accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(targetLobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodPost, "/lobby/join_lobby", strings.NewReader(`{"lobby_id":101,"session_id":146}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = JoinLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}
}

func TestLeaveLobby_TransfersOwnershipToOldestRemainingMember(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(80)
	accountID := int64(5)
	lobbyID := int64(30)
	joinedAt := time.Now().Add(-time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}).AddRow(lobbyID))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Open Lobby", "Owner", accountID, false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Open Lobby", "Owner", accountID, false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM lobby_members WHERE lobby_id = $1 AND account_id = $2")).
		WithArgs(lobbyID, accountID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT a.id, a.name, lm.joined_at
FROM lobby_members lm
JOIN account a ON a.id = lm.account_id
WHERE lm.lobby_id = $1
ORDER BY lm.joined_at ASC, lm.id ASC
LIMIT 1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "joined_at"}).AddRow(int64(42), "Next Owner", joinedAt))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET owner_account_id = $1, owner_name = $2, last_activity_at = NOW(), updated_at = NOW() WHERE id = $3")).
		WithArgs(int64(42), "Next Owner", lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodPost, "/lobby/leave_lobby", strings.NewReader(`{"session_id":80}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = LeaveLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}
}

func TestLeaveLobby_ClosesLobbyWhenOwnerWasLastMember(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(81)
	accountID := int64(5)
	lobbyID := int64(31)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}).AddRow(lobbyID))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Open Lobby", "Owner", accountID, false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Open Lobby", "Owner", accountID, false, false, true, time.Now(), time.Now(), time.Now(), nil))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM lobby_members WHERE lobby_id = $1 AND account_id = $2")).
		WithArgs(lobbyID, accountID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT a.id, a.name, lm.joined_at
FROM lobby_members lm
JOIN account a ON a.id = lm.account_id
WHERE lm.lobby_id = $1
ORDER BY lm.joined_at ASC, lm.id ASC
LIMIT 1`)).
		WithArgs(lobbyID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET is_closed = TRUE, closed_at = NOW(), last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodPost, "/lobby/leave_lobby", strings.NewReader(`{"session_id":81}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = LeaveLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}
}

func TestGetLobby_RejectsLobbyInactiveAtThirtyMinutesWithoutConnectedSubscribers(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	lobbyID := int64(33)
	inactiveAt := time.Now().Add(-31 * time.Minute)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Inactive Empty Lobby", "Owner", int64(9), false, false, true, time.Now(), time.Now(), inactiveAt, nil))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET is_closed = TRUE, closed_at = NOW(), updated_at = NOW() WHERE id = $1 AND closed_at IS NULL")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodGet, "/lobby/get_lobby", strings.NewReader(`{"lobby_id":33}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = GetLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusGone {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusGone)
	}
}

func TestGetLobby_WithConnectedSubscribersDoesNotCloseAtThirtyMinutes(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	lobbyID := int64(34)
	now := time.Now()
	inactiveAt := now.Add(-31 * time.Minute)
	_, unsubscribe := lobbyBrokerInstance.subscribe(lobbyID)
	defer unsubscribe()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Stale But Connected Lobby", "Owner", int64(10), false, false, true, now, now, inactiveAt, nil))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodGet, "/lobby/get_lobby", strings.NewReader(`{"lobby_id":34}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = GetLobby(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}
}

func TestGetLobby_WithConnectedSubscribersClosesAtTwoHours(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	lobbyID := int64(35)
	inactiveAt := time.Now().Add(-2*time.Hour - time.Minute)
	_, unsubscribe := lobbyBrokerInstance.subscribe(lobbyID)
	defer unsubscribe()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Very Stale Connected Lobby", "Owner", int64(11), false, false, true, time.Now(), time.Now(), inactiveAt, nil))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET is_closed = TRUE, closed_at = NOW(), updated_at = NOW() WHERE id = $1 AND closed_at IS NULL")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req, err := http.NewRequest(http.MethodGet, "/lobby/get_lobby", strings.NewReader(`{"lobby_id":35}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	err = GetLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusGone {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusGone)
	}
}
