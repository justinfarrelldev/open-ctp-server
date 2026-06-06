package lobby

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateLobby_InvalidMethod(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/lobby/update_lobby", nil)
	rr := httptest.NewRecorder()

	err := UpdateLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateLobby_MissingLobbyID(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/lobby/update_lobby", strings.NewReader(`{"lobby":{"name":"Renamed"},"session_id":1}`))
	rr := httptest.NewRecorder()

	err := UpdateLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}
