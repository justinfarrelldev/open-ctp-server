package account

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
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
	updateArgs := UpdateAccountArgs{
		Password: &password,
	}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/account/update_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB)
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
	password := "password123"
	name := "Updated Name"
	info := "Updated Info"
	location := "Updated Location"
	email := "updated@example.com"
	experienceLevel := 2

	updateArgs := UpdateAccountArgs{
		AccountId: &accountID,
		Password:  &password,
		Account: &AccountParam{
			Name:            &name,
			Info:            &info,
			Location:        &location,
			Email:           &email,
			ExperienceLevel: (*ExperienceLevel)(&experienceLevel),
		},
	}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/account/update_account", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	hasher := auth.NewArgon2idHash(1, 32, 64*1024, 32, 256)
	hashSalt, err := hasher.GenerateHash([]byte(password), nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when generating hash and salt", err)
	}

	storedHash := base64.StdEncoding.EncodeToString(hashSalt.Hash)
	storedSalt := base64.StdEncoding.EncodeToString(hashSalt.Salt)

	mock.ExpectQuery("SELECT hash, salt FROM passwords WHERE id = \\$1").
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"hash", "salt"}).AddRow(storedHash, storedSalt))

	mock.ExpectExec("UPDATE account SET name = \\$1, info = \\$2, location = \\$3, email = \\$4, experience_level = \\$5 WHERE id = \\$6").
		WithArgs(name, info, location, email, experienceLevel, accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateAccount(w, r, sqlxDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedResponse := "Successfully updated account!"
	if strings.TrimSpace(rr.Body.String()) != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedResponse)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
