package lobby

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func TestCreateLobby_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobby := Lobby{
		Name:           "Test Lobby",
		OwnerName:      "Owner",
		OwnerAccountId: "1",
		IsClosed:       false,
		IsMuted:        false,
		IsPublic:       true,
	}

	mock.ExpectQuery("INSERT INTO lobby \\(name, owner_name, owner_account_id, is_closed, is_muted, is_public\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\)").
		WithArgs(lobby.Name, lobby.OwnerName, lobby.OwnerAccountId, lobby.IsClosed, lobby.IsMuted, lobby.IsPublic).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	lobbyBytes, err := json.Marshal(lobby)
	if err != nil {
		t.Fatal(err)
	}
	lobbyJSON := fmt.Sprintf(`{"lobby": %s, "password": "password123"}`, string(lobbyBytes))

	req, err := http.NewRequest("POST", "/lobby/create_lobby", strings.NewReader(lobbyJSON))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := CreateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestCreateLobby_InvalidMethod(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("GET", "/lobby/create_lobby", strings.NewReader(`{"lobby": {"name": "Test Lobby", "owner_name": "Owner", "is_closed": false, "is_muted": false, "is_public": true}, "password": "password123"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := CreateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "invalid request; request must be a POST request"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestCreateLobby_DecodeError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/lobby/create_lobby", strings.NewReader(`{"lobby": {"name": "Test Lobby", "owner_name": "Owner", "is_closed": false, "is_muted": false, "is_public": true}, "password": 123}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := CreateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "an error occurred while decoding the request body:json: cannot unmarshal number into Go struct field CreateLobbyArgs.password of type string"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestCreateLobby_PasswordTooShort(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobby := Lobby{
		Name:           "Test Lobby",
		OwnerName:      "Owner",
		OwnerAccountId: "1",
		IsClosed:       false,
		IsMuted:        false,
		IsPublic:       true,
	}

	lobbyBytes, err := json.Marshal(lobby)
	if err != nil {
		t.Fatal(err)
	}
	lobbyJSON := fmt.Sprintf(`{"lobby": %s, "password": "123"}`, string(lobbyBytes))

	req, err := http.NewRequest("POST", "/lobby/create_lobby", strings.NewReader(lobbyJSON))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := CreateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := ERROR_PASSWORD_TOO_SHORT
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestCreateLobby_PasswordRequired(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/lobby/create_lobby", strings.NewReader(`{"lobby": {"name": "Test Lobby", "owner_name": "Owner", "is_closed": false, "is_muted": false, "is_public": true}}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := CreateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}
