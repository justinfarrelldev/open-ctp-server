package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestRequireSessionFromJSONBody_Missing(t *testing.T) {
	store := NewSessionStore(nil)
	response := httptest.NewRecorder()

	session, err := RequireSessionFromJSONBody(response, store, testSessionBody{})
	if !errors.Is(err, ErrInvalidBodySessionID) {
		t.Fatalf("expected ErrInvalidBodySessionID, got %v", err)
	}
	if session != nil {
		t.Fatalf("expected nil session, got %#v", session)
	}
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRequireSessionFromJSONBody_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)
	response := httptest.NewRecorder()
	createdAt := time.Now()
	expiresAt := createdAt.Add(2 * time.Hour)
	var sessionID int64 = 77

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, 9, createdAt, expiresAt))

	session, err := RequireSessionFromJSONBody(response, store, testSessionBody{SessionId: &sessionID})
	if err != nil {
		t.Fatalf("RequireSessionFromJSONBody() error = %v", err)
	}
	if session == nil || session.ID != sessionID {
		t.Fatalf("session = %#v want id %d", session, sessionID)
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", response.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRequireSessionFromQuery_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)
	response := httptest.NewRecorder()
	createdAt := time.Now()
	expiresAt := createdAt.Add(2 * time.Hour)
	var sessionID int64 = 88

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, 4, createdAt, expiresAt))

	session, err := RequireSessionFromQuery(response, httptest.NewRequest(http.MethodGet, "/events?session_id=88", nil), store)
	if err != nil {
		t.Fatalf("RequireSessionFromQuery() error = %v", err)
	}
	if session == nil || session.ID != sessionID {
		t.Fatalf("session = %#v want id %d", session, sessionID)
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", response.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRequireSession_InternalStoreError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)
	response := httptest.NewRecorder()
	var sessionID int64 = 100

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnError(sql.ErrConnDone)

	session, err := RequireSession(response, store, sessionID)
	if session != nil {
		t.Fatalf("expected nil session, got %#v", session)
	}
	if err == nil || !strings.Contains(err.Error(), "an error occurred while retrieving the session") {
		t.Fatalf("error = %v want wrapped retrieval error", err)
	}
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d want %d", response.Code, http.StatusInternalServerError)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
