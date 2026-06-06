package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

type testSessionBody struct {
	SessionId *int64 `json:"session_id"`
}

func (b testSessionBody) GetSessionID() *int64 {
	return b.SessionId
}

func TestSessionIDFromJSONBody_Success(t *testing.T) {
	var sessionID int64 = 42
	body := testSessionBody{SessionId: &sessionID}

	got, err := SessionIDFromJSONBody(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != sessionID {
		t.Fatalf("expected session ID %d, got %d", sessionID, got)
	}
}

func TestSessionIDFromJSONBody_Missing(t *testing.T) {
	_, err := SessionIDFromJSONBody(testSessionBody{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidBodySessionID) {
		t.Fatalf("expected ErrInvalidBodySessionID, got %v", err)
	}
}

func TestSessionIDFromQuery_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?session_id=42", nil)

	got, err := SessionIDFromQuery(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != 42 {
		t.Fatalf("expected session ID 42, got %d", got)
	}
}

func TestSessionIDFromQuery_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	_, err := SessionIDFromQuery(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrSessionIDRequired) {
		t.Fatalf("expected ErrSessionIDRequired, got %v", err)
	}
}

func TestSessionIDFromQuery_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?session_id=nope", nil)

	_, err := SessionIDFromQuery(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidSessionID) {
		t.Fatalf("expected ErrInvalidSessionID, got %v", err)
	}
}

func TestRequireSessionFromQuery_Missing(t *testing.T) {
	store := NewSessionStore(nil)
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	session, err := RequireSessionFromQuery(response, req, store)
	if !errors.Is(err, ErrSessionIDRequired) {
		t.Fatalf("expected ErrSessionIDRequired, got %v", err)
	}

	if session != nil {
		t.Fatalf("expected nil session, got %#v", session)
	}

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestRequireSessionFromQuery_Invalid(t *testing.T) {
	store := NewSessionStore(nil)
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events?session_id=nope", nil)

	session, err := RequireSessionFromQuery(response, req, store)
	if !errors.Is(err, ErrInvalidSessionID) {
		t.Fatalf("expected ErrInvalidSessionID, got %v", err)
	}

	if session != nil {
		t.Fatalf("expected nil session, got %#v", session)
	}

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestRequireSession_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)
	response := httptest.NewRecorder()

	var sessionID int64 = 100
	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, 1, createdAt, expiresAt))

	session, err := RequireSession(response, store, sessionID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session == nil || session.ID != sessionID {
		t.Fatalf("expected session %d, got %#v", sessionID, session)
	}

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestRequireSession_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)
	response := httptest.NewRecorder()

	var sessionID int64 = 100
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnError(sql.ErrNoRows)

	session, err := RequireSession(response, store, sessionID)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}

	if session != nil {
		t.Fatalf("expected nil session, got %#v", session)
	}

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestRequireSession_Expired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)
	response := httptest.NewRecorder()

	var sessionID int64 = 100
	createdAt := time.Now().Add(-2 * time.Hour)
	expiresAt := time.Now().Add(-1 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, 1, createdAt, expiresAt))

	session, err := RequireSession(response, store, sessionID)
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("expected ErrSessionExpired, got %v", err)
	}

	if session != nil {
		t.Fatalf("expected nil session, got %#v", session)
	}

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}
