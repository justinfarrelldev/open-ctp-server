package account

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/DATA-DOG/go-sqlmock"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func TestGetAccount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	var sessionID int64 = 1

	expectedAccount := Account{
		Name:            "John Doe",
		Info:            "Some info",
		Location:        "Some location",
		Email:           "john.doe@example.com",
		ExperienceLevel: ExperienceLevel(5),
	}

	createdAt := time.Now()
	expiresAt := time.Now().Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))

	mock.ExpectQuery("SELECT name, info, location, email, experience_level FROM account WHERE id = \\$1").
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"name", "info", "location", "email", "experience_level"}).
			AddRow(expectedAccount.Name, expectedAccount.Info, expectedAccount.Location, expectedAccount.Email, int(expectedAccount.ExperienceLevel)))

	req, err := http.NewRequest("GET", "/account/get_account?account_id=1&session_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	GetAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("handler returned wrong content type: got %q want %q", got, "application/json")
	}

	var actualAccount Account
	if err := json.Unmarshal(rr.Body.Bytes(), &actualAccount); err != nil {
		t.Fatalf("expected valid JSON response, got %q (%v)", rr.Body.String(), err)
	}
	if actualAccount != expectedAccount {
		t.Errorf("handler returned unexpected account: got %+v want %+v", actualAccount, expectedAccount)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	var sessionID int64 = 1

	createdAt := time.Now()
	expiresAt := time.Now().Add(6 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).
			AddRow(sessionID, accountID, createdAt, expiresAt))

	mock.ExpectQuery("SELECT name, info, location, email, experience_level FROM account WHERE id = \\$1").
		WithArgs(accountID).
		WillReturnError(sql.ErrNoRows)

	req, err := http.NewRequest("GET", "/account/get_account?account_id=1&session_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	GetAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	expectedError := "no account exists with the ID 1"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetAccountHandler_InvalidMethodReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/account/get_account?account_id=1&session_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	GetAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "invalid request; request must be a GET request"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetAccountHandler_InvalidAccountIDReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("GET", "/account/get_account?account_id=invalid&session_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	GetAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "invalid account_id"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetAccountHandler_MissingSessionIDReturnsUnauthorized(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("GET", "/account/get_account?account_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	GetAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	expectedError := "session_id is required"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetAccountHandler_InvalidSessionIDReturnsUnauthorized(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("GET", "/account/get_account?account_id=1&session_id=invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	GetAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	expectedError := "invalid session_id"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetAccountHandler_MissingAccountIDReturnsBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("GET", "/account/get_account?session_id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{DB: sqlxDB}

	GetAccountHandler(rr, req, sqlxDB, mockStore)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "account_id is required"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}
