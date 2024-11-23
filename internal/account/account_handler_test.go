package account

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateAccountHandler(w, r, sqlxDB)
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateAccountHandler(w, r, sqlxDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateAccountHandler(w, r, sqlxDB)
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
