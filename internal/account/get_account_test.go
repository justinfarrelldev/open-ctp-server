package account

import (
	"database/sql"
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
	sessionID := "1"

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
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedResponse := `{"name":"John Doe","info":"Some info","location":"Some location","email":"john.doe@example.com","experience_level":5}`
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedResponse)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	accountID := int64(1)
	sessionID := "1"

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
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "no account exists with the ID 1"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetAccount_InvalidMethod(t *testing.T) {
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
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "invalid request; request must be a GET request"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetAccount_DecodeError(t *testing.T) {
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
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "invalid account_id"
	if strings.ReplaceAll(strings.TrimSpace(rr.Body.String()), "\n", "") != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.ReplaceAll(strings.TrimSpace(rr.Body.String()), "\n", ""), expectedError)
	}
}

func TestGetAccount_MissingAccountID(t *testing.T) {
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
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetAccount(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "account_id is required"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}
