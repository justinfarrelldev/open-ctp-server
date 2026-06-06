package account

import (
	"bytes"
	"encoding/base64"
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

type loginResponse struct {
	SessionID int64     `json:"session_id"`
	AccountID int64     `json:"account_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func buildStoredPassword(t *testing.T, password string) (string, string) {
	t.Helper()

	hashSalt, err := auth.Hasher.GenerateHash([]byte(password), nil)
	if err != nil {
		t.Fatalf("failed generating password hash: %v", err)
	}

	return base64.StdEncoding.EncodeToString(hashSalt.Hash), base64.StdEncoding.EncodeToString(hashSalt.Salt)
}

func TestLoginAccount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"email":"captain@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/account/login", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	storedHash, storedSalt := buildStoredPassword(t, "password123")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, p.hash, p.salt FROM account a JOIN passwords p ON p.account_email = a.email WHERE a.email = $1")).
		WithArgs("captain@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hash", "salt"}).AddRow(7, storedHash, storedSalt))

	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), 7, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = LoginAccount(rr, req, sqlxDB, store)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response loginResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got %v", err)
	}

	if response.AccountID != 7 {
		t.Fatalf("expected account_id 7, got %d", response.AccountID)
	}

	if response.SessionID == 0 {
		t.Fatal("expected non-zero session_id")
	}

	if response.ExpiresAt.IsZero() {
		t.Fatal("expected expires_at to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestLoginAccount_WrongPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"email":"captain@example.com","password":"wrongpass"}`
	req := httptest.NewRequest(http.MethodPost, "/account/login", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	storedHash, storedSalt := buildStoredPassword(t, "password123")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, p.hash, p.salt FROM account a JOIN passwords p ON p.account_email = a.email WHERE a.email = $1")).
		WithArgs("captain@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hash", "salt"}).AddRow(7, storedHash, storedSalt))

	err = LoginAccount(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if err.Error() != "invalid email or password" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoginAccount_EmailNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"email":"missing@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/account/login", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, p.hash, p.salt FROM account a JOIN passwords p ON p.account_email = a.email WHERE a.email = $1")).
		WithArgs("missing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hash", "salt"}))

	err = LoginAccount(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if err.Error() != "invalid email or password" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoginAccountHandler_InvalidPasswordReturnsUnauthorized(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"email":"captain@example.com","password":"wrongpass"}`
	req := httptest.NewRequest(http.MethodPost, "/account/login", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	storedHash, storedSalt := buildStoredPassword(t, "password123")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, p.hash, p.salt FROM account a JOIN passwords p ON p.account_email = a.email WHERE a.email = $1")).
		WithArgs("captain@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hash", "salt"}).AddRow(7, storedHash, storedSalt))

	LoginAccountHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "invalid email or password" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}
