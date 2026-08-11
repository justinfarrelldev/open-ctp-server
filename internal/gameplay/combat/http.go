// Package combat exposes the authenticated, viewer-safe combat command adapter.
package combat

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/service"
)

type executor interface {
	ExecuteCommand(context.Context, *auth.Session, service.CommandRequest) (service.Result, error)
}

type request struct {
	SessionID        *int64 `json:"session_id"`
	ExpectedRevision int64  `json:"expected_revision"`
	IdempotencyKey   string `json:"idempotency_key"`
	CommandID        string `json:"command_id"`
	AttackerArmyID   string `json:"attacker_army_id"`
	DefenderArmyID   string `json:"defender_army_id"`
}

func (r request) GetSessionID() *int64 { return r.SessionID }

type authenticator func(http.ResponseWriter, *http.Request, request) (*auth.Session, error)

type Handler struct {
	executor     executor
	sessions     *auth.SessionStore
	authenticate authenticator
}

func NewHandler(commandService *service.Service, sessions *auth.SessionStore) *Handler {
	return &Handler{executor: commandService, sessions: sessions, authenticate: func(w http.ResponseWriter, _ *http.Request, body request) (*auth.Session, error) {
		return auth.RequireSessionFromJSONBody(w, sessions, body)
	}}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request; request must be a POST request", http.StatusMethodNotAllowed)
		return
	}
	gameID, err := gameIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "game id must be specified", http.StatusBadRequest)
		return
	}
	if h == nil || h.executor == nil || h.sessions == nil || h.authenticate == nil {
		http.Error(w, "gameplay combat is unavailable", http.StatusInternalServerError)
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var body request
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "malformed combat request", http.StatusBadRequest)
		return
	}
	session, err := h.authenticate(w, r, body)
	if err != nil {
		return
	}
	payload, err := json.Marshal(struct {
		AttackerArmyID string `json:"attacker_army_id"`
		DefenderArmyID string `json:"defender_army_id"`
	}{AttackerArmyID: body.AttackerArmyID, DefenderArmyID: body.DefenderArmyID})
	if err != nil {
		http.Error(w, "malformed combat request", http.StatusBadRequest)
		return
	}
	result, err := h.executor.ExecuteCommand(r.Context(), session, service.CommandRequest{GameID: gameID, ExpectedRevision: body.ExpectedRevision, IdempotencyKey: body.IdempotencyKey, CommandID: body.CommandID, Type: engine.CommandCombat, Payload: payload})
	if err != nil {
		status := service.StatusCode(err)
		if status == 0 {
			status = http.StatusInternalServerError
		}
		http.Error(w, "gameplay combat is unavailable", status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func gameIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "game" || parts[2] != "combat" {
		return 0, errors.New("invalid combat path")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid game id")
	}
	return id, nil
}
