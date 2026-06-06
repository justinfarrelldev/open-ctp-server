package auth

import (
	"database/sql"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGenerateSessionID_PositiveAndAcceptedByValidators(t *testing.T) {
	t.Parallel()

	sessionID, err := generateSessionID()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sessionID <= 0 {
		t.Fatalf("expected positive session ID, got %d", sessionID)
	}

	bodySessionID, err := SessionIDFromJSONBody(testSessionBody{SessionId: &sessionID})
	if err != nil {
		t.Fatalf("expected JSON body validator to accept generated session ID, got %v", err)
	}
	if bodySessionID != sessionID {
		t.Fatalf("expected JSON body session ID %d, got %d", sessionID, bodySessionID)
	}

	req := httptest.NewRequest("GET", "/account/get_account?session_id="+strconv.FormatInt(sessionID, 10), nil)
	querySessionID, err := SessionIDFromQuery(req)
	if err != nil {
		t.Fatalf("expected query validator to accept generated session ID, got %v", err)
	}
	if querySessionID != sessionID {
		t.Fatalf("expected query session ID %d, got %d", sessionID, querySessionID)
	}
}

func TestCreateSession_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	accountID := 1
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), accountID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	session, err := store.CreateSession(accountID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if session.AccountID != accountID {
		t.Errorf("expected account ID %d, got %d", accountID, session.AccountID)
	}

	if session.ID <= 0 {
		t.Fatalf("expected positive session ID, got %d", session.ID)
	}

	if _, err := SessionIDFromJSONBody(testSessionBody{SessionId: &session.ID}); err != nil {
		t.Fatalf("expected created session ID to pass JSON body validation, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateSession_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	accountID := 1
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), accountID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	_, err = store.CreateSession(accountID)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetSession_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	var sessionID int64 = 12345678
	accountID := 1
	createdAt := time.Now()
	expiresAt := createdAt.Add(12 * time.Hour)

	rows := sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
		AddRow(sessionID, accountID, createdAt, expiresAt)
	mock.ExpectQuery("SELECT \\* FROM sessions WHERE id = \\$1").
		WithArgs(sessionID).
		WillReturnRows(rows)

	session, err := store.GetSession(sessionID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if session.ID != sessionID {
		t.Errorf("expected session ID %d, got %d", sessionID, session.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	var sessionID int64 = 87654321
	mock.ExpectQuery("SELECT \\* FROM sessions WHERE id = \\$1").
		WithArgs(sessionID).
		WillReturnError(sql.ErrNoRows)

	session, err := store.GetSession(sessionID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if session != nil {
		t.Errorf("expected nil session, got %v", session)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteSession_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	var sessionID int64 = 12345678
	mock.ExpectExec("DELETE FROM sessions WHERE id = \\$1").
		WithArgs(sessionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.DeleteSession(sessionID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteSession_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	var sessionID int64 = 87654321
	mock.ExpectExec("DELETE FROM sessions WHERE id = \\$1").
		WithArgs(sessionID).
		WillReturnError(sql.ErrConnDone)

	err = store.DeleteSession(sessionID)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteSessionsForAccount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	var accountID int64 = 42
	mock.ExpectExec("DELETE FROM sessions WHERE account_id = \\$1").
		WithArgs(accountID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = store.DeleteSessionsForAccount(accountID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteSessionsForAccountExcept_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := NewSessionStore(sqlxDB)

	var accountID int64 = 42
	var exceptSessionID int64 = 99
	mock.ExpectExec("DELETE FROM sessions WHERE account_id = \\$1 AND id <> \\$2").
		WithArgs(accountID, exceptSessionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.DeleteSessionsForAccountExcept(accountID, exceptSessionID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired - future time",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "expired - past time",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired - current time",
			expiresAt: time.Now(),
			want:      true,
		},
		{
			name:      "not expired - far future",
			expiresAt: time.Now().Add(24 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ExpiresAt: tt.expiresAt,
			}
			if got := session.IsExpired(); got != tt.want {
				t.Errorf("Session.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
