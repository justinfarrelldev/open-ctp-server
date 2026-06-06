package account

import (
	"bytes"
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

func TestResetPassword_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	accountID := int64(7)
	sessionID := int64(55)
	body := `{"account_id":7,"session_id":55,"current_password":"password123","new_password":"new-password-456"}`
	req := httptest.NewRequest(http.MethodPost, "/account/reset_password", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	storedHash, storedSalt := buildStoredPassword(t, "password123")
	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("captain@example.com"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT hash, salt FROM passwords WHERE account_email = $1")).
		WithArgs("captain@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"hash", "salt"}).AddRow(storedHash, storedSalt))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE passwords SET hash = $1, salt = $2 WHERE account_email = $3")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "captain@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE account_id = $1 AND id <> $2")).
		WithArgs(accountID, sessionID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	ResetPasswordHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("handler returned wrong content type: got %q want %q", got, "application/json")
	}

	var response string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid JSON response, got %q (%v)", rr.Body.String(), err)
	}
	if response != "Successfully reset password!" {
		t.Fatalf("unexpected response: %q", response)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestResetPassword_WrongCurrentPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	accountID := int64(7)
	sessionID := int64(55)
	body := `{"account_id":7,"session_id":55,"current_password":"wrongpass","new_password":"new-password-456"}`
	req := httptest.NewRequest(http.MethodPost, "/account/reset_password", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	storedHash, storedSalt := buildStoredPassword(t, "password123")
	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("captain@example.com"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT hash, salt FROM passwords WHERE account_email = $1")).
		WithArgs("captain@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"hash", "salt"}).AddRow(storedHash, storedSalt))
	mock.ExpectRollback()

	err = ResetPassword(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if err.Error() != "current password is incorrect" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResetPasswordHandler_MissingSessionIDReturnsUnauthorized(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"account_id":7,"current_password":"password123","new_password":"new-password-456"}`
	req := httptest.NewRequest(http.MethodPost, "/account/reset_password", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	ResetPasswordHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "a valid session_id must be specified" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestResetPasswordHandler_DecodeErrorReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `invalid json`
	req := httptest.NewRequest(http.MethodPost, "/account/reset_password", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	ResetPasswordHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "an error occurred while decoding the request body: invalid character 'i' looking for beginning of value" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestResetPasswordHandler_MissingAccountIDReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"session_id":55,"current_password":"password123","new_password":"new-password-456"}`
	req := httptest.NewRequest(http.MethodPost, "/account/reset_password", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	ResetPasswordHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "account_id must be specified" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestResetPasswordHandler_MissingCurrentPasswordReturnsBadRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"account_id":7,"session_id":55,"new_password":"new-password-456"}`
	req := httptest.NewRequest(http.MethodPost, "/account/reset_password", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(int64(55), int64(7), createdAt, expiresAt))

	ResetPasswordHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "current_password must be specified" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestResetPasswordHandler_MissingNewPasswordReturnsBadRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"account_id":7,"session_id":55,"current_password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/account/reset_password", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(int64(55), int64(7), createdAt, expiresAt))

	ResetPasswordHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "new_password must be specified" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestDeleteAccount_SessionAccountMismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	accountID := int64(7)
	sessionID := int64(55)
	args := DeleteAccountArgs{AccountId: &accountID, SessionId: &sessionID}
	body, _ := json.Marshal(args)
	req := httptest.NewRequest(http.MethodDelete, "/account/delete_account", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, 99, createdAt, expiresAt))

	err = DeleteAccount(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}

	if err.Error() != "session does not own the target account" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAccountHandler_InvalidSessionIDReturnsUnauthorized(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	requestBody := `{"account_id":7,"session_id":-1}`
	req := httptest.NewRequest(http.MethodDelete, "/account/delete_account", bytes.NewBufferString(requestBody))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	DeleteAccountHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "a valid session_id must be specified" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}
