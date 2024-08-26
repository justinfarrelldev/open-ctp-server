package game

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGame_PasswordTooShort(t *testing.T) {
	// Create a test request with a password that is less than 6 characters
	body := ExpectedBody{
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

	// Call the function to test
	err = CreateGame(rr, req)

	// Check if the error is what we expect
	expectedError := ERROR_PASSWORD_TOO_SHORT
	if err == nil || err.Error() != expectedError {
		t.Errorf("CreateGame() error = %v, wantErr %v", err, expectedError)
	}
}
