package lobby

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGetLobby_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobbyID := int8(1)

	mock.ExpectQuery("SELECT id, name, owner_name, is_closed, is_muted, is_public FROM lobby WHERE id = \\$1").
		WithArgs(lobbyID).
		WillReturnError(sql.ErrNoRows)

	req, err := http.NewRequest("GET", "/lobby/get_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetLobby(w, r, sqlxDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "no lobby exists with the ID 1"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestGetLobby_InvalidMethod(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/lobby/get_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetLobby(w, r, sqlxDB)
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

func TestGetLobby_DecodeError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("GET", "/lobby/get_lobby", strings.NewReader(`{"lobby_id": "invalid"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetLobby(w, r, sqlxDB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "an error occurred while decoding the request body: json: cannot unmarshal string into Go struct field GetLobbyArgs.lobby_id of type int8"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}
