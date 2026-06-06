package lobby

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
