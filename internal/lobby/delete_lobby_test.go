package lobby

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeleteLobby_InvalidMethod(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/lobby/delete_lobby", strings.NewReader(`{"lobby_id":1,"session_id":2}`))
	rr := httptest.NewRecorder()

	err := DeleteLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteLobby_LobbyIDRequired(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/lobby/delete_lobby", strings.NewReader(`{"lobby_id":0,"session_id":2}`))
	rr := httptest.NewRecorder()

	err := DeleteLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}
