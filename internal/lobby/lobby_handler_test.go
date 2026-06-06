package lobby

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateLobbyHandler_PropagatesStatusAwareErrors(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/lobby/create_lobby", strings.NewReader(`{"lobby":{"name":"Test Lobby"},"password":"password123","session_id":0}`))
	rr := httptest.NewRecorder()

	CreateLobbyHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestJoinLobbyHandler_PropagatesBadRequest(t *testing.T) {
	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/lobby/join_lobby", strings.NewReader(`{"session_id":1}`))
	rr := httptest.NewRecorder()

	JoinLobbyHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}
