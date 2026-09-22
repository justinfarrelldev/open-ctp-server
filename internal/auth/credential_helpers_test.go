package auth

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGetHashAndSalt_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	wantHash := []byte("hash-bytes")
	wantSalt := []byte("salt-bytes")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT hash, salt FROM passwords WHERE account_email = $1")).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"hash", "salt"}).
			AddRow(base64.StdEncoding.EncodeToString(wantHash), base64.StdEncoding.EncodeToString(wantSalt)))

	gotHash, gotSalt, err := GetHashAndSalt("test@example.com", sqlxDB)
	if err != nil {
		t.Fatalf("GetHashAndSalt() error = %v", err)
	}
	if string(gotHash) != string(wantHash) {
		t.Fatalf("hash = %q want %q", gotHash, wantHash)
	}
	if string(gotSalt) != string(wantSalt) {
		t.Fatalf("salt = %q want %q", gotSalt, wantSalt)
	}
}

func TestGetHashAndSalt_QueryOrDecodeErrors(t *testing.T) {
	tests := []struct {
		name       string
		rowHash    string
		rowSalt    string
		queryErr   error
		wantErrSub string
	}{
		{
			name:       "query error",
			queryErr:   sql.ErrNoRows,
			wantErrSub: "error retrieving account credentials",
		},
		{
			name:       "invalid hash",
			rowHash:    "%%%",
			rowSalt:    base64.StdEncoding.EncodeToString([]byte("salt")),
			wantErrSub: "error decoding stored hash",
		},
		{
			name:       "invalid salt",
			rowHash:    base64.StdEncoding.EncodeToString([]byte("hash")),
			rowSalt:    "%%%",
			wantErrSub: "error decoding stored salt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New() error = %v", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			expect := mock.ExpectQuery(regexp.QuoteMeta("SELECT hash, salt FROM passwords WHERE account_email = $1")).
				WithArgs("test@example.com")
			if tt.queryErr != nil {
				expect.WillReturnError(tt.queryErr)
			} else {
				expect.WillReturnRows(sqlmock.NewRows([]string{"hash", "salt"}).AddRow(tt.rowHash, tt.rowSalt))
			}

			_, _, err = GetHashAndSalt("test@example.com", sqlxDB)
			if err == nil || !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Fatalf("error = %v want substring %q", err, tt.wantErrSub)
			}
		})
	}
}

func TestUpdateHashAndSalt(t *testing.T) {
	hashSalt := &hashSalt{Hash: []byte("hash"), Salt: []byte("salt")}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New() error = %v", err)
		}
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mock.ExpectExec(regexp.QuoteMeta("UPDATE passwords SET hash = $1, salt = $2 WHERE account_email = $3")).
			WithArgs(base64.StdEncoding.EncodeToString(hashSalt.Hash), base64.StdEncoding.EncodeToString(hashSalt.Salt), "test@example.com").
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := UpdateHashAndSalt(hashSalt, "test@example.com", sqlxDB); err != nil {
			t.Fatalf("UpdateHashAndSalt() error = %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New() error = %v", err)
		}
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mock.ExpectExec(regexp.QuoteMeta("UPDATE passwords SET hash = $1, salt = $2 WHERE account_email = $3")).
			WithArgs(base64.StdEncoding.EncodeToString(hashSalt.Hash), base64.StdEncoding.EncodeToString(hashSalt.Salt), "test@example.com").
			WillReturnError(sql.ErrConnDone)

		err = UpdateHashAndSalt(hashSalt, "test@example.com", sqlxDB)
		if !errors.Is(err, sql.ErrConnDone) && (err == nil || !strings.Contains(err.Error(), "an error occurred while updating a hash-salt pair")) {
			t.Fatalf("error = %v want wrapped update error", err)
		}
	})
}
