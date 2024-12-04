package auth

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

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

	sessionID := "test-session-id"
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
		t.Errorf("expected session ID %s, got %s", sessionID, session.ID)
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

	sessionID := "non-existent-session-id"
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

	sessionID := "test-session-id"
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

	sessionID := "test-session-id"
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
