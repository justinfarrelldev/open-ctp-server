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

func TestUpdateAccount_InvalidMethod(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/account/update_account", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "invalid request; request must be a PUT request"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestUpdateAccount_DecodeError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("PUT", "/account/update_account", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "an error occurred while decoding the request body:invalid character 'i' looking for beginning of value"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestUpdateAccount_MissingAccountID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	password := "John Doe Updated"
	updateArgs := UpdateAccountArgs{Password: &password}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/account/update_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB, mockStore)
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

func TestUpdateAccount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	var sessionID int64 = 1
	password := "password123"
	name := "Updated Name"
	info := "Updated Info"
	location := "Updated Location"
	email := "updated@example.com"
	experienceLevel := ExperienceLevel(2)

	updateArgs := UpdateAccountArgs{
		AccountId: &accountID,
		Password:  &password,
		Account: &AccountParam{
			Name:            &name,
			Info:            &info,
			Location:        &location,
			Email:           &email,
			ExperienceLevel: &experienceLevel,
		},
		SessionId: &sessionID,
	}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/account/update_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	storedHash, storedSalt := buildStoredPassword(t, password)
	createdAt := time.Now()
	expiresAt := time.Now().Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("current@example.com"))
	mock.ExpectQuery("SELECT \\* from account WHERE email = \\$1").
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT hash, salt FROM passwords WHERE account_email = $1")).
		WithArgs("current@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"hash", "salt"}).AddRow(storedHash, storedSalt))
	mock.ExpectExec("UPDATE account SET name = \\$1, info = \\$2, location = \\$3, email = \\$4, experience_level = \\$5 WHERE id = \\$6").
		WithArgs(name, info, location, email, experienceLevel, accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE passwords SET account_email = $1 WHERE account_email = $2")).
		WithArgs(email, "current@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mockStore := &auth.SessionStore{DB: sqlxDB}

	UpdateAccountHandler(rr, req, sqlxDB, mockStore)

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
	if response != "Successfully updated account!" {
		t.Errorf("handler returned unexpected body: got %v want %v", response, "Successfully updated account!")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateAccount_MissingPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	var sessionID int64 = 1
	name := "Updated Name"

	updateArgs := UpdateAccountArgs{
		AccountId: &accountID,
		Account:   &AccountParam{Name: &name},
		SessionId: &sessionID,
	}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/account/update_account", bytes.NewBuffer(body))
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

	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "the password for the account must be specified"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestUpdateAccount_MissingAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	var sessionID int64 = 1
	password := "password123"

	updateArgs := UpdateAccountArgs{AccountId: &accountID, Password: &password, SessionId: &sessionID}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/account/update_account", bytes.NewBuffer(body))
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

	mockStore := &auth.SessionStore{DB: sqlxDB}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "account must be specified"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestUpdateAccount_InvalidExperienceLevel(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	accountID := int64(1)
	sessionID := int64(1)
	password := "password123"
	experienceLevel := ExperienceLevel(9)
	args := UpdateAccountArgs{
		AccountId: &accountID,
		Password:  &password,
		SessionId: &sessionID,
		Account:   &AccountParam{ExperienceLevel: &experienceLevel},
	}

	body, _ := json.Marshal(args)
	req := httptest.NewRequest(http.MethodPut, "/account/update_account", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(sessionID, accountID, createdAt, expiresAt))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("current@example.com"))
	mock.ExpectRollback()

	err = UpdateAccount(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "experience_level must be between 0 and 5 (0=easy, 5=impossible)" {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUpdateAccount_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error creating db: %v", err)
	}
	defer db.Close()

	accountID := int64(1)
	sessionID := int64(1)
	password := "password123"
	email := "taken@example.com"
	args := UpdateAccountArgs{
		AccountId: &accountID,
		Password:  &password,
		SessionId: &sessionID,
		Account:   &AccountParam{Email: &email},
	}

	body, _ := json.Marshal(args)
	req := httptest.NewRequest(http.MethodPut, "/account/update_account", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	createdAt := time.Now()
	expiresAt := createdAt.Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(sessionID, accountID, createdAt, expiresAt))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM account WHERE id = $1")).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("current@example.com"))
	mock.ExpectQuery("SELECT \\* from account WHERE email = \\$1").
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(42, email))
	mock.ExpectRollback()

	err = UpdateAccount(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "the provided email is already in use" {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
