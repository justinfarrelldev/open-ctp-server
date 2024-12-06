package account

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func TestCreateAccount_InvalidMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "/account/create_account", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	var mockDB *sqlx.DB = nil
	var mockStore *auth.SessionStore = nil

	_, err = CreateAccount(rr, req, mockDB, mockStore)

	expectedError := "invalid request; request must be a POST request"
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, expectedError)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCreateAccount_DecodeError(t *testing.T) {
	invalidJSON := `{"account": {"name": "Test User", "info": "Test Info", "location": "Test Location", "email": "test@example.com", "experience_level": 3}, "password": 123}`
	req, err := http.NewRequest("POST", "/account/create_account", bytes.NewBufferString(invalidJSON))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	var mockDB *sqlx.DB = nil
	var mockStore *auth.SessionStore = nil

	_, err = CreateAccount(rr, req, mockDB, mockStore)

	expectedError := "an error occurred while decoding the request body:json: cannot unmarshal number into Go struct field CreateAccountArgs.password of type string"
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, expectedError)
	}

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestCreateAccount_PasswordTooShort(t *testing.T) {
	account := CreateAccountArgs{
		Account: Account{
			Name:            "Test User",
			Info:            "Test Info",
			Location:        "Test Location",
			Email:           "test@example.com",
			ExperienceLevel: 3,
		},
		Password: "123",
	}

	jsonBody, _ := json.Marshal(account)
	req, err := http.NewRequest("POST", "/account/create_account", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	var mockDB *sqlx.DB = nil
	var mockStore *auth.SessionStore = nil

	_, err = CreateAccount(rr, req, mockDB, mockStore)

	expectedError := ERROR_PASSWORD_TOO_SHORT
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, expectedError)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCreateAccount_PasswordRequired(t *testing.T) {
	account := CreateAccountArgs{
		Account: Account{
			Name:            "Test User",
			Info:            "Test Info",
			Location:        "Test Location",
			Email:           "test@example.com",
			ExperienceLevel: 3,
		},
	}

	jsonBody, _ := json.Marshal(account)
	req, err := http.NewRequest("POST", "/account/create_account", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	var mockDB *sqlx.DB = nil
	var mockStore *auth.SessionStore = nil

	_, err = CreateAccount(rr, req, mockDB, mockStore)

	expectedError := ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, expectedError)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
func TestCreateAccount_Success(t *testing.T) {
	account := CreateAccountArgs{
		Account: Account{
			Name:            "Test User",
			Info:            "Test Info",
			Location:        "Test Location",
			Email:           "test@example.com",
			ExperienceLevel: 3,
		},
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(account)
	req, err := http.NewRequest("POST", "/account/create_account", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT \\* from account WHERE email = \\$1").
		WithArgs(account.Account.Email).
		WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectQuery("INSERT INTO account \\(name, info, location, email, experience_level\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
		WithArgs(account.Account.Name, account.Account.Info, account.Account.Location, account.Account.Email, account.Account.ExperienceLevel).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery("INSERT INTO passwords \\(account_email, hash, salt\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectExec("INSERT INTO sessions \\(id, account_id, created_at, expires_at\\) VALUES \\(\\?, \\?, \\?, \\?\\)").
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	_, err = CreateAccount(rr, req, sqlxDB, mockStore)
	if err != nil {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, nil)
	}

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateAccount_ExperienceLevelTooLow(t *testing.T) {
	account := CreateAccountArgs{
		Account: Account{
			Name:            "Test User",
			Info:            "Test Info",
			Location:        "Test Location",
			Email:           "test@example.com",
			ExperienceLevel: -1,
		},
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(account)
	req, err := http.NewRequest("POST", "/account/create_account", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	var mockDB *sqlx.DB = nil
	var mockStore *auth.SessionStore = nil

	_, err = CreateAccount(rr, req, mockDB, mockStore)

	expectedError := "experience_level must be between 0 and 5 (0=easy, 5=impossible)"
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, expectedError)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCreateAccount_ExperienceLevelTooHigh(t *testing.T) {
	account := CreateAccountArgs{
		Account: Account{
			Name:            "Test User",
			Info:            "Test Info",
			Location:        "Test Location",
			Email:           "test@example.com",
			ExperienceLevel: 6,
		},
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(account)
	req, err := http.NewRequest("POST", "/account/create_account", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	var mockDB *sqlx.DB = nil
	var mockStore *auth.SessionStore = nil

	_, err = CreateAccount(rr, req, mockDB, mockStore)

	expectedError := "experience_level must be between 0 and 5 (0=easy, 5=impossible)"
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, expectedError)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCreateAccount_InvalidEmail(t *testing.T) {
	account := CreateAccountArgs{
		Account: Account{
			Name:            "Test User",
			Info:            "Test Info",
			Location:        "Test Location",
			Email:           "invalid-email",
			ExperienceLevel: 3,
		},
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(account)
	req, err := http.NewRequest("POST", "/account/create_account", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	var mockDB *sqlx.DB = nil
	var mockStore *auth.SessionStore = nil

	_, err = CreateAccount(rr, req, mockDB, mockStore)

	expectedError := "an error occurred while checking whether the email for the account is valid: mail: missing '@' or angle-addr"
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateAccount() error = %v, wantErr %v", err, expectedError)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
