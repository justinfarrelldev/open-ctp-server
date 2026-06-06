package world

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func TestWorldReadContract(t *testing.T) {
	t.Run("returns normalized world payload for joined player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		req = withWorldReadContractState(req, worldReadContractStateReady)
		rr := httptest.NewRecorder()

		GetWorldHandler(rr, req, nil, nil)

		if rr.Code != http.StatusOK {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
		}

		var response WorldReadResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if response.GameID != 44 {
			t.Fatalf("got game_id %d want 44", response.GameID)
		}
		if response.GameStatus != "started" {
			t.Fatalf("got game_status %q want started", response.GameStatus)
		}
		if strings.TrimSpace(response.GeneratorVersion) == "" {
			t.Fatal("generator_version is required in /world responses")
		}
		if response.HostParams.Style == "" || response.HostParams.RulesJSON == "" || response.HostParams.ExclusionsJSON == "" {
			t.Fatalf("host params must preserve setup defaults, got %+v", response.HostParams)
		}
		if response.HostParams.Seed == nil {
			t.Fatal("host_params.seed is required in /world responses")
		}
		if response.HostParams.Wetness == nil {
			t.Fatal("host_params.wetness is required in /world responses")
		}
		if len(response.PlayerOrder) == 0 {
			t.Fatal("player_order is required in /world responses")
		}
		if len(response.Spawns) == 0 {
			t.Fatal("spawns are required in /world responses")
		}
		if len(response.Tiles) == 0 {
			t.Fatal("tiles are required in /world responses")
		}
	})

	t.Run("rejects missing game id path value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/world/?session_id=55", nil)
		req.SetPathValue("game_id", "")
		rr := httptest.NewRecorder()

		GetWorldHandler(rr, req, nil, nil)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
		}
		if strings.TrimSpace(rr.Body.String()) != "game_id must be specified" {
			t.Fatalf("got body %q want %q", rr.Body.String(), "game_id must be specified")
		}
	})

	t.Run("rejects invalid game id path value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/world/not-a-number?session_id=55", nil)
		req.SetPathValue("game_id", "not-a-number")
		rr := httptest.NewRecorder()

		GetWorldHandler(rr, req, nil, nil)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
		}
		if strings.TrimSpace(rr.Body.String()) != "invalid game_id" {
			t.Fatalf("got body %q want %q", rr.Body.String(), "invalid game_id")
		}
	})

	t.Run("rejects invalid session query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=invalid", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()

		GetWorldHandler(rr, req, nil, nil)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusUnauthorized)
		}
		if strings.TrimSpace(rr.Body.String()) != auth.ErrInvalidSessionID.Error() {
			t.Fatalf("got body %q want %q", rr.Body.String(), auth.ErrInvalidSessionID.Error())
		}
	})

	t.Run("rejects non member access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		req = withWorldReadContractState(req, worldReadContractStateNonMember)
		rr := httptest.NewRecorder()

		GetWorldHandler(rr, req, nil, nil)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusForbidden)
		}
		if strings.TrimSpace(rr.Body.String()) != errorGamePlayerRequired {
			t.Fatalf("got body %q want %q", rr.Body.String(), errorGamePlayerRequired)
		}
	})

	t.Run("rejects setup games before the world becomes active", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		req = withWorldReadContractState(req, worldReadContractStateSetupGame)
		rr := httptest.NewRecorder()

		GetWorldHandler(rr, req, nil, nil)

		if rr.Code != http.StatusConflict {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusConflict)
		}
		if strings.TrimSpace(rr.Body.String()) != "game 44 has not started" {
			t.Fatalf("got body %q want %q", rr.Body.String(), "game 44 has not started")
		}
	})

	t.Run("returns not found when world is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		req = withWorldReadContractState(req, worldReadContractStateMissing)
		rr := httptest.NewRecorder()

		GetWorldHandler(rr, req, nil, nil)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusNotFound)
		}
		if !strings.Contains(strings.ToLower(rr.Body.String()), "world") {
			t.Fatalf("expected missing-world response body, got %q", rr.Body.String())
		}
	})
}

func TestWorldSpaceLayerPlaceholder(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/world/44/space?session_id=55", nil)
	req.SetPathValue("game_id", "44")
	rr := httptest.NewRecorder()

	GetWorldSpaceLayerHandler(rr, req, nil, nil)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("got content type %q want application/json", contentType)
	}

	wantCanonical := canonicalJSON(t, []byte(`{ "error": "space layer is not supported for CTP2 worlds" }`))
	gotCanonical := canonicalJSON(t, rr.Body.Bytes())
	if gotCanonical != wantCanonical {
		t.Fatalf("got body %s want %s", gotCanonical, wantCanonical)
	}
}

func canonicalJSON(t *testing.T, raw []byte) string {
	t.Helper()
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		t.Fatalf("json.Compact() error = %v raw=%q", err, string(raw))
	}
	return compact.String()
}
