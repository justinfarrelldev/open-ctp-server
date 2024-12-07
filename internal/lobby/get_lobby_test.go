package lobby

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func TestGetLobby_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobbyID := int64(1)
	expectedLobby := Lobby{
		ID:             lobbyID,
		Name:           "Test Lobby",
		OwnerName:      "Owner",
		OwnerAccountId: "1",
		IsClosed:       false,
		IsMuted:        false,
		IsPublic:       true,
	}

	mock.ExpectQuery("SELECT id, name, owner_name, is_closed, is_muted, is_public FROM lobby WHERE id = \\$1").
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public"}).
			AddRow(expectedLobby.ID, expectedLobby.Name, expectedLobby.OwnerName, expectedLobby.OwnerAccountId, expectedLobby.IsClosed, expectedLobby.IsMuted, expectedLobby.IsPublic))

	req, err := http.NewRequest("GET", "/lobby/get_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetLobby(w, r, sqlxDB, mockStore)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedResponse := `{"id":1,"name":"Test Lobby","owner_name":"Owner","owner_account_id":"1","is_closed":false,"is_muted":false,"is_public":true}`
	if strings.TrimSpace(rr.Body.String()) != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedResponse)
	}
}

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

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetLobby(w, r, sqlxDB, mockStore)
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

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetLobby(w, r, sqlxDB, mockStore)
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

	mockStore := &auth.SessionStore{
		DB: sqlxDB,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := GetLobby(w, r, sqlxDB, mockStore)
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
