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

func TestCreateLobbyHandler_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobby := Lobby{
		Name:      "Test Lobby",
		OwnerName: "Owner",
		IsClosed:  false,
		IsMuted:   false,
		IsPublic:  true,
	}

	mock.ExpectQuery("INSERT INTO lobby \\(name, owner_name, is_closed, is_muted, is_public\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
		WithArgs(lobby.Name, lobby.OwnerName, lobby.IsClosed, lobby.IsMuted, lobby.IsPublic).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	req, err := http.NewRequest("POST", "/lobby/create_lobby", strings.NewReader(`{"lobby": {"name": "Test Lobby", "owner_name": "Owner", "is_closed": false, "is_muted": false, "is_public": true}, "password": "password123"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateLobbyHandler(w, r, sqlxDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestCreateLobbyHandler_InvalidMethod(t *testing.T) {
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateLobbyHandler(w, r, sqlxDB)
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

func TestCreateLobbyHandler_DecodeError(t *testing.T) {
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateLobbyHandler(w, r, sqlxDB)
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

func TestCreateLobbyHandler_PasswordTooShort(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/lobby/create_lobby", strings.NewReader(`{"lobby": {"name": "Test Lobby", "owner_name": "Owner", "is_closed": false, "is_muted": false, "is_public": true}, "password": "123"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateLobbyHandler(w, r, sqlxDB)
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

func TestCreateLobbyHandler_PasswordRequired(t *testing.T) {
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateLobbyHandler(w, r, sqlxDB)
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

func TestGetLobbyHandler_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	lobby := Lobby{
		ID:        1,
		Name:      "Test Lobby",
		OwnerName: "Owner",
		IsClosed:  false,
		IsMuted:   false,
		IsPublic:  true,
	}

	mock.ExpectQuery("SELECT id, name, owner_name, is_closed, is_muted, is_public FROM lobby WHERE id = \\$1").
		WithArgs(lobby.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "is_closed", "is_muted", "is_public"}).
			AddRow(lobby.ID, lobby.Name, lobby.OwnerName, lobby.IsClosed, lobby.IsMuted, lobby.IsPublic))

	req, err := http.NewRequest("GET", "/lobby/get_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GetLobbyHandler(w, r, sqlxDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetLobbyHandler_InvalidMethod(t *testing.T) {
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
		GetLobbyHandler(w, r, sqlxDB)
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

func TestGetLobbyHandler_DecodeError(t *testing.T) {
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
		GetLobbyHandler(w, r, sqlxDB)
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

func TestGetLobbyHandler_LobbyNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, owner_name, is_closed, is_muted, is_public FROM lobby WHERE id = \\$1").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	req, err := http.NewRequest("GET", "/lobby/get_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GetLobbyHandler(w, r, sqlxDB)
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

func TestDeleteLobbyHandler_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM lobby WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req, err := http.NewRequest("DELETE", "/lobby/delete_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteLobbyHandler(w, r, sqlxDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Successfully deleted lobby!"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expected)
	}
}

func TestDeleteLobbyHandler_InvalidMethod(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("POST", "/lobby/delete_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteLobbyHandler(w, r, sqlxDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "invalid request; request must be a DELETE request"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteLobbyHandler_DecodeError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("DELETE", "/lobby/delete_lobby", strings.NewReader(`{"lobby_id": "invalid"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteLobbyHandler(w, r, sqlxDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "an error occurred while decoding the request body: json: cannot unmarshal string into Go struct field DeleteLobbyArgs.lobby_id of type int64"
	if strings.TrimSpace(rr.Body.String()) != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v", strings.TrimSpace(rr.Body.String()), expectedError)
	}
}

func TestDeleteLobbyHandler_LobbyNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM lobby WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 0))

	req, err := http.NewRequest("DELETE", "/lobby/delete_lobby", strings.NewReader(`{"lobby_id": 1}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteLobbyHandler(w, r, sqlxDB)
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

func TestDeleteLobbyHandler_LobbyIdRequired(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	req, err := http.NewRequest("DELETE", "/lobby/delete_lobby", strings.NewReader(`{"lobby_id": 0}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteLobbyHandler(w, r, sqlxDB)
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
