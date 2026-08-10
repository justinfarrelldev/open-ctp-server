package lobby

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateLobby_InvalidMethod(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/lobby/create_lobby", strings.NewReader(`{"lobby":{"name":"Test Lobby"},"password":"password123","session_id":1}`))
	rr := httptest.NewRecorder()

	err := CreateLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateLobby_PasswordRequired(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/lobby/create_lobby", strings.NewReader(`{"lobby":{"name":"Test Lobby"},"session_id":1}`))
	rr := httptest.NewRecorder()

	err := CreateLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateLobby_SuccessReturnsOnlyLobbyID(t *testing.T) {
	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	mock.ExpectQuery("SELECT \\* FROM sessions WHERE id = \\$1").WithArgs(int64(1)).WillReturnRows(lobbySessionRows(1, 7, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT lm.lobby_id").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}))
	mock.ExpectQuery("SELECT name FROM account WHERE id = \\$1").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("owner"))
	mock.ExpectQuery("INSERT INTO lobbies").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(91)))
	mock.ExpectExec("INSERT INTO lobby_members").WithArgs(int64(91), int64(7)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/lobby/create_lobby", strings.NewReader(`{"lobby":{"name":"Test Lobby"},"password":"password123","session_id":1}`))
	rr := httptest.NewRecorder()
	if err := CreateLobby(rr, req, sqlxDB, store); err != nil {
		t.Fatal(err)
	}
	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d", rr.Code)
	}
	var response struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode safe lobby identity: %v", err)
	}
	if response.ID != 91 {
		t.Fatalf("lobby ID=%d want=91", response.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
