// Package gameplayhttp exposes authenticated, gateway-only gameplay commands.
package gameplayhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	SessionID        *int64          `json:"session_id"`
	ExpectedRevision int64           `json:"expected_revision"`
	IdempotencyKey   string          `json:"idempotency_key"`
	CommandID        string          `json:"command_id"`
	Payload          json.RawMessage `json:"-"`
}

func (r request) GetSessionID() *int64 { return r.SessionID }

type authenticator func(http.ResponseWriter, *http.Request, request) (*auth.Session, error)

// CommandRequestDocs is the shared authenticated gameplay command envelope. The
// route selects the command; actor identity is derived from the session.
type CommandRequestDocs struct {
	SessionID        int64  `json:"session_id"`
	ExpectedRevision int64  `json:"expected_revision"`
	IdempotencyKey   string `json:"idempotency_key"`
	CommandID        string `json:"command_id"`
}

// CommandResponseDocs is the viewer-safe gameplay receipt envelope.
type CommandResponseDocs struct {
	Revision         int64          `json:"revision"`
	ReceiptID        string         `json:"receipt_id"`
	EventCursor      int64          `json:"event_cursor"`
	ViewerProjection map[string]any `json:"viewer_projection"`
}

type Handler struct {
	executor     executor
	authenticate authenticator
}

func NewHandler(commandService *service.Service, sessions *auth.SessionStore) *Handler {
	return &Handler{executor: commandService, authenticate: func(w http.ResponseWriter, _ *http.Request, body request) (*auth.Session, error) {
		return auth.RequireSessionFromJSONBody(w, sessions, body)
	}}
}

var commands = map[string]struct {
	kind    string
	allowed map[string]bool
}{
	"pass":              {engine.CommandPass, fields()},
	"set-war":           {engine.CommandSetWar, fields("target_player_id", "at_war")},
	"establish-contact": {engine.CommandEstablishContact, fields("target_player_id")},
	"propose-agreement": {engine.CommandProposeAgreement, fields("proposal_id", "target_player_id", "kind")},
	"respond-agreement": {engine.CommandRespondAgreement, fields("proposal_id", "accept")},
	"set-government":    {engine.CommandSetGovernment, fields("government_id")},
	"set-science-rate":  {engine.CommandSetScienceRate, fields("science_percent")},
	"set-social-policy": {engine.CommandSetSocialPolicy, fields("workday", "wages", "rations")},
	"set-readiness":     {engine.CommandSetReadiness, fields("readiness")},
	"round-complete":    {engine.CommandRoundComplete, fields()},
	"set-research":      {engine.CommandSetResearch, fields("advance_id")},
	"move":              {engine.CommandMoveUnit, fields("unit_id", "path")},
	"disband":           {engine.CommandDisband, fields("unit_id")},
	"embark":            {engine.CommandEmbark, fields("unit_id", "transport_id")},
	"disembark":         {engine.CommandDisembark, fields("unit_id", "tile_id")},
	"fortify":           {engine.CommandFortify, fields("unit_id")},
	"found-city":        {engine.CommandFoundCity, fields("unit_id", "city_id", "name")},
	"army-create":       {engine.CommandArmyCreate, fields("army_id", "unit_ids")},
	"army-join":         {engine.CommandArmyJoin, fields("army_id", "unit_id")},
	"army-leave":        {engine.CommandArmyLeave, fields("army_id", "unit_id")},
	"combat":            {engine.CommandCombat, fields("attacker_army_id", "defender_army_id")},
	"establish-embassy": {engine.CommandEstablishEmbassy, fields("unit_id", "city_id")},
	"soothsaying":       {engine.CommandSoothsaying, fields("unit_id", "city_id")},
	"steal-technology":  {engine.CommandStealTechnology, fields("unit_id", "city_id")},
	"terraform":         {engine.CommandTerraform, fields("unit_id", "tile_id", "target_terrain")},
	"build-improvement": {engine.CommandBuildImprovement, fields("unit_id", "tile_id", "improvement_id")},
	"cancel-terraform":  {engine.CommandCancelTerraform, fields("tile_id")},
	"queue-production":  {engine.CommandQueueProduction, fields("city_id", "item")},
	"cancel-production": {engine.CommandCancelProduction, fields("city_id", "item_id")},
	"set-specialists":   {engine.CommandSetSpecialists, fields("city_id", "specialists")},
	"leave":             {engine.CommandLeaveGame, fields()},
	"end-early":         {engine.CommandEndGame, fields()},
}

func fields(names ...string) map[string]bool {
	allowed := make(map[string]bool, len(names)+4)
	for _, name := range []string{"session_id", "expected_revision", "idempotency_key", "command_id"} {
		allowed[name] = true
	}
	for _, name := range names {
		allowed[name] = true
	}
	return allowed
}

// ServeHTTP accepts the documented command routes (pass, move, disband, embark,
// disembark, fortify, found-city, army-*, combat, special actions, terraform, production,
// leave, and end-early) through one authenticated gateway adapter.
//
// @Summary Submit a gameplay command
// @Description The server derives player identity from session_id and submits one canonical command. Unknown fields and client authority are rejected.
// @Tags gameplay
// @Accept json
// @Produce json
// @Param game_id path int true "game ID"
// @Param command path string true "command name"
// @Param command body CommandRequestDocs true "command envelope plus command-specific payload"
// @Success 200 {object} CommandResponseDocs
// @Failure 400 {object} map[string]string "malformed or unknown command field"
// @Failure 401 {object} map[string]string "missing or invalid session"
// @Failure 403 {object} map[string]string "game membership or controller is not authorized"
// @Failure 409 {object} map[string]string "stale revision or idempotency conflict"
// @Failure 422 {object} map[string]string "invalid command phase or terminal game"
// @Failure 429 {object} ratelimit.Message "rate limit exceeded"
// @Router /game/{game_id}/{command} [post]
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed)
		return
	}
	gameID, commandName, err := route(r.URL.Path)
	if err != nil || h == nil || h.executor == nil || h.authenticate == nil {
		writeError(w, http.StatusBadRequest)
		return
	}
	definition, ok := commands[commandName]
	if !ok {
		writeError(w, http.StatusNotFound)
		return
	}
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil || len(raw) == 0 || decoder.Decode(&struct{}{}) != io.EOF {
		writeError(w, http.StatusBadRequest)
		return
	}
	for field := range raw {
		if !definition.allowed[field] {
			writeError(w, http.StatusBadRequest)
			return
		}
	}
	bodyBytes, _ := json.Marshal(raw)
	var body struct {
		SessionID        *int64 `json:"session_id"`
		ExpectedRevision int64  `json:"expected_revision"`
		IdempotencyKey   string `json:"idempotency_key"`
		CommandID        string `json:"command_id"`
	}
	if json.Unmarshal(bodyBytes, &body) != nil {
		writeError(w, http.StatusBadRequest)
		return
	}
	payload := make(map[string]json.RawMessage, len(raw))
	for key, value := range raw {
		if key != "session_id" && key != "expected_revision" && key != "idempotency_key" && key != "command_id" {
			payload[key] = value
		}
	}
	payloadBytes, _ := json.Marshal(payload)
	req := request{SessionID: body.SessionID, ExpectedRevision: body.ExpectedRevision, IdempotencyKey: body.IdempotencyKey, CommandID: body.CommandID, Payload: payloadBytes}
	session, err := h.authenticate(w, r, req)
	if err != nil {
		return
	}
	result, err := h.executor.ExecuteCommand(r.Context(), session, service.CommandRequest{GameID: gameID, ExpectedRevision: req.ExpectedRevision, IdempotencyKey: req.IdempotencyKey, CommandID: req.CommandID, Type: definition.kind, Payload: req.Payload})
	if err != nil {
		writeError(w, service.StatusCode(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func route(path string) (int64, string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "game" {
		return 0, "", errors.New("invalid gameplay path")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return 0, "", errors.New("invalid game id")
	}
	return id, parts[2], nil
}

func writeError(w http.ResponseWriter, status int) {
	if status < 400 || status > 599 {
		status = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: "gameplay request rejected"})
}
