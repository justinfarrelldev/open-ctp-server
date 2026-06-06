package account

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// func TestCreateAccountHandler_Success(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer db.Close()

// 	account := Account{
// 		Name:            "Test User",
// 		Info:            "Some info",
// 		Location:        "Some location",
// 		Email:           "test@example.com",
// 		ExperienceLevel: Beginner,
// 	}

// 	mock.ExpectQuery("INSERT INTO account \\(name, info, location, email, experience_level\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
// 		WithArgs(account.Name, account.Info, account.Location, account.Email, account.ExperienceLevel).
// 		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

// 	req, err := http.NewRequest("POST", "/account/create_account", strings.NewReader(`{"name": "Test User", "info": "Some info", "location": "Some location", "email": "test@example.com", "experience_level": 0}`))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	sqlxDB := sqlx.NewDb(db, "sqlmock")
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		CreateAccountHandler(w, r, sqlxDB)
// 	})

// 	handler.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusCreated {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
// 	}
// }

func TestCreateAccountHandler_InvalidMethod(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("GET", "/account/create_account", strings.NewReader(`{"name": "Test User", "info": "Some info", "location": "Some location", "email": "test@example.com", "experience_level": 0}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateAccountHandler(w, r, sqlxDB, store)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "invalid request; request must be a POST request"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestCreateAccountHandler_DecodeError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/account/create_account", strings.NewReader(`{ "account": {"name": "Test User", "info": "Some info", "location": "Some location", "email": "test@example.com", "experience_level": "invalid"}, "password": "fake" }`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateAccountHandler(w, r, sqlxDB, store)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "an error occurred while decoding the request body:json: cannot unmarshal string into Go struct field Account.account.experience_level of type account.ExperienceLevel"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestCreateAccountHandler_EmailRequired(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/account/create_account", strings.NewReader(`{"account": {"name": "Test User", "info": "Some info", "location": "Some location", "experience_level": 0}, "password": "test password" }`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateAccountHandler(w, r, sqlxDB, store)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "an error occurred while checking whether the email for the account is valid: mail: no address"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestCreateAccountHandler_SuccessWritesSessionID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/account/create_account", strings.NewReader(`{"account":{"name":"Test User","info":"Some info","location":"Some location","email":"test@example.com","experience_level":3},"password":"password123"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	store := auth.NewSessionStore(sqlxDB)

	mock.ExpectQuery("SELECT \\* from account WHERE email = \\$1").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows(nil))
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO account \\(name, info, location, email, experience_level\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\) RETURNING id").
		WithArgs("Test User", "Some info", "Some location", "test@example.com", ExperienceLevel(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO passwords \\(account_email, hash, salt\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"account_email"}).AddRow("test@example.com"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO sessions (id, account_id, created_at, expires_at) VALUES (?, ?, ?, ?) ")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	CreateAccountHandler(rr, req, sqlxDB, store)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("handler returned wrong content type: got %q want %q", got, "application/json")
	}

	var sessionID int64
	if err := json.Unmarshal(rr.Body.Bytes(), &sessionID); err != nil {
		t.Fatalf("expected JSON integer session response, got %q (%v)", rr.Body.String(), err)
	}
	if sessionID <= 0 {
		t.Fatalf("expected positive session ID, got %d", sessionID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %s", err)
	}
}
