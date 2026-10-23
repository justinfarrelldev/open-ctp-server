package auth

import (
	"bytes"
	"errors"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestSessionStoreNeverLogsBearerIDsOrRawDatabaseErrors(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewSessionStore(sqlx.NewDb(db, "sqlmock"))
	id := int64(91234567890123456)
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM sessions").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(id, 7, now, now.Add(time.Hour)))
	if _, err := store.GetSession(id); err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("INSERT INTO sessions").WillReturnResult(sqlmock.NewResult(0, 1))
	session, err := store.CreateSession(7)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM sessions WHERE account_id").WithArgs(int64(7), id).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.DeleteSessionsForAccountExcept(7, id); err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM sessions WHERE id").WithArgs(id).WillReturnError(errors.New("SECRET-DB-CREDENTIAL"))
	if err := store.DeleteSession(id); err == nil {
		t.Fatal("missing injected failure")
	}
	if strings.Contains(output.String(), strconv.FormatInt(id, 10)) || strings.Contains(output.String(), strconv.FormatInt(session.ID, 10)) || strings.Contains(output.String(), "SECRET-DB-CREDENTIAL") {
		t.Fatalf("credential leaked: %s", output.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
