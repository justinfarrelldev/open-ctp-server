package auth

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestCreateSession_NilDatabase(t *testing.T) {
	store := NewSessionStore(nil)

	session, err := store.CreateSession(1)
	if session != nil {
		t.Fatalf("expected nil session, got %#v", session)
	}
	if err == nil || err.Error() != "Database connection is nil" {
		t.Fatalf("error = %v want %q", err, "Database connection is nil")
	}
}

func TestCreateSessionTx_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), 7, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := sqlxDB.Beginx()
	if err != nil {
		t.Fatalf("Beginx() error = %v", err)
	}

	session, err := NewSessionStore(sqlxDB).CreateSessionTx(tx, 7)
	if err != nil {
		t.Fatalf("CreateSessionTx() error = %v", err)
	}
	if session == nil || session.AccountID != 7 {
		t.Fatalf("session = %#v want account 7", session)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDeleteSessionHelperErrorsAndTxWrappers(t *testing.T) {
	t.Run("delete sessions for account error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New() error = %v", err)
		}
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		store := NewSessionStore(sqlxDB)
		mock.ExpectExec("DELETE FROM sessions WHERE account_id = \\$1").
			WithArgs(int64(42)).
			WillReturnError(sql.ErrConnDone)

		err = store.DeleteSessionsForAccount(42)
		if err == nil || !strings.Contains(err.Error(), sql.ErrConnDone.Error()) {
			t.Fatalf("error = %v want wrapped exec error", err)
		}
	})

	t.Run("delete sessions for account except error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New() error = %v", err)
		}
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		store := NewSessionStore(sqlxDB)
		mock.ExpectExec("DELETE FROM sessions WHERE account_id = \\$1 AND id <> \\$2").
			WithArgs(int64(42), int64(99)).
			WillReturnError(sql.ErrConnDone)

		err = store.DeleteSessionsForAccountExcept(42, 99)
		if err == nil || !strings.Contains(err.Error(), sql.ErrConnDone.Error()) {
			t.Fatalf("error = %v want wrapped exec error", err)
		}
	})

	t.Run("delete sessions for account tx wrapper", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New() error = %v", err)
		}
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM sessions WHERE account_id = \\$1").
			WithArgs(int64(42)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatalf("Beginx() error = %v", err)
		}
		if err := NewSessionStore(sqlxDB).DeleteSessionsForAccountTx(tx, 42); err != nil {
			t.Fatalf("DeleteSessionsForAccountTx() error = %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("Commit() error = %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("delete sessions for account except tx wrapper", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New() error = %v", err)
		}
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM sessions WHERE account_id = \\$1 AND id <> \\$2").
			WithArgs(int64(42), int64(99)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatalf("Beginx() error = %v", err)
		}
		if err := NewSessionStore(sqlxDB).DeleteSessionsForAccountExceptTx(tx, 42, 99); err != nil {
			t.Fatalf("DeleteSessionsForAccountExceptTx() error = %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("Commit() error = %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})
}
