package lobby

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func TestUpdateLobby_InvalidMethod(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/lobby/update_lobby", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateLobby(w, r, sqlxDB, mockStore)
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

func TestUpdateLobby_DecodeError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("PUT", "/lobby/update_lobby", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateLobby(w, r, sqlxDB, mockStore)
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

func TestUpdateLobby_MissingLobbyID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	name := "Updated Lobby"
	updateArgs := UpdateLobbyArgs{
		Lobby: &LobbyParam{
			Name: &name,
		},
	}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/lobby/update_lobby", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "lobby_id must be specified"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestUpdateLobby_MissingLobby(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobbyID := int64(1)
	updateArgs := UpdateLobbyArgs{
		LobbyId: &lobbyID,
	}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/lobby/update_lobby", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expectedError := "lobby must be specified"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestUpdateLobby_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobbyID := int64(1)
	name := "Updated Lobby"
	ownerName := "Updated Owner"
	isClosed := true
	isMuted := false
	isPublic := true

	updateArgs := UpdateLobbyArgs{
		LobbyId: &lobbyID,
		Lobby: &LobbyParam{
			Name:      &name,
			OwnerName: &ownerName,
			IsClosed:  &isClosed,
			IsMuted:   &isMuted,
			IsPublic:  &isPublic,
		},
	}

	body, _ := json.Marshal(updateArgs)
	req, err := http.NewRequest("PUT", "/lobby/update_lobby", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	mock.ExpectExec("UPDATE lobby SET name = \\$1, owner_name = \\$2, is_closed = \\$3, is_muted = \\$4, is_public = \\$5 WHERE id = \\$6").
		WithArgs(name, ownerName, isClosed, isMuted, isPublic, lobbyID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := UpdateLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedResponse := "Successfully updated lobby!"
	if strings.TrimSpace(rr.Body.String()) != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedResponse)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
