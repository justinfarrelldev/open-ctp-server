package game

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestCreateGame_PasswordTooShort(t *testing.T) {
	// Create a test request with a password that is less than 6 characters
	body := CreateGameArgs{
		PasswordProtected: true,
		Password:          "123", // This password is less than 6 characters
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/game", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// DB is not needed for this test
	var mockDB *sqlx.DB = nil

	// Call the function to test
	err = CreateGame(rr, req, mockDB)

	// Check if the error is what we expect
	expectedError := ERROR_PASSWORD_TOO_SHORT
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateGame() error = %v, wantErr %v", err, expectedError)
	}
}

func TestCreateGame_PasswordRequiredWhenPasswordProtectedIsTrue(t *testing.T) {
	// Create a test request with password protected set to true but no password
	body := CreateGameArgs{
		PasswordProtected: true,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/game", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// DB is not needed for this test
	var mockDB *sqlx.DB = nil

	// Call the function to test
	err = CreateGame(rr, req, mockDB)

	// Check if the error is what we expect
	expectedError := ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateGame() error = %v, wantErr %v", err, expectedError)
	}
}

func TestCreateGame_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	game := CreateGameArgs{
		PasswordProtected: true,
		Password:          "password123",
	}

	mock.ExpectQuery("INSERT INTO game \\(password_protected, password\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(game.PasswordProtected, game.Password).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	jsonBody, _ := json.Marshal(game)
	req, err := http.NewRequest("POST", "/game/create_game", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := CreateGame(w, r, sqlxDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}
