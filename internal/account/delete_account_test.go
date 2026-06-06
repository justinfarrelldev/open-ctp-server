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

func TestDeleteAccount_InvalidMethod(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/account/delete_account", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "invalid request; request must be a DELETE request"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteAccount_DecodeError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("DELETE", "/account/delete_account", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "an error occurred while decoding the request body: invalid character 'i' looking for beginning of value"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteAccount_MissingAccountID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sessionID := int64(1)
	deleteArgs := DeleteAccountArgs{SessionId: &sessionID}

	body, _ := json.Marshal(deleteArgs)
	req, err := http.NewRequest("DELETE", "/account/delete_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "account_id must be specified"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteAccount_MissingSessionID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	var acctID int64 = 1
	deleteArgs := DeleteAccountArgs{AccountId: &acctID}

	body, _ := json.Marshal(deleteArgs)
	req, err := http.NewRequest("DELETE", "/account/delete_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	expectedError := "a valid session_id must be specified"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteAccount_SessionNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	sessionID := int64(1)
	deleteArgs := DeleteAccountArgs{AccountId: &accountID, SessionId: &sessionID}

	body, _ := json.Marshal(deleteArgs)
	req, err := http.NewRequest("DELETE", "/account/delete_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows(nil))
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	expectedError := "session not found"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteAccount_SessionExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	sessionID := int64(1)
	deleteArgs := DeleteAccountArgs{AccountId: &accountID, SessionId: &sessionID}

	body, _ := json.Marshal(deleteArgs)
	req, err := http.NewRequest("DELETE", "/account/delete_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	createdAt := time.Now().Add(-2 * time.Hour)
	expiresAt := time.Now().Add(-1 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	expectedError := "session has expired"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteAccount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	sessionID := int64(1)
	deleteArgs := DeleteAccountArgs{AccountId: &accountID, SessionId: &sessionID}

	body, _ := json.Marshal(deleteArgs)
	req, err := http.NewRequest("DELETE", "/account/delete_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	createdAt := time.Now()
	expiresAt := time.Now().Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("captain@example.com"))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE account_id = $1")).
		WithArgs(accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM passwords WHERE account_email = $1")).
		WithArgs("captain@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM account WHERE id = \\$1").
		WithArgs(accountID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mockStore := &auth.SessionStore{DB: sqlxDB}

	DeleteAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("handler returned wrong content type: got %q want %q", got, "application/json")
	}

	var response string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid JSON response, got %q (%v)", rr.Body.String(), err)
	}
	if response != "Successfully deleted account!" {
		t.Errorf("handler returned unexpected body: got %v want %v", response, "Successfully deleted account!")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteAccount_NoAccountExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	sessionID := int64(1)
	deleteArgs := DeleteAccountArgs{AccountId: &accountID, SessionId: &sessionID}

	body, _ := json.Marshal(deleteArgs)
	req, err := http.NewRequest("DELETE", "/account/delete_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	createdAt := time.Now()
	expiresAt := time.Now().Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("captain@example.com"))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE account_id = $1")).
		WithArgs(accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM passwords WHERE account_email = $1")).
		WithArgs("captain@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM account WHERE id = \\$1").
		WithArgs(accountID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "no rows were affected when the DELETE query ran for the account with ID 1"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}
