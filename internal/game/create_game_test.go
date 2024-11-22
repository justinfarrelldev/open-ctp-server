package game

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
