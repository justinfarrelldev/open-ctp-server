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

const maxCommandBodyBytes = 256 << 10

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
	"pass":                  {engine.CommandPass, fields()},
	"set-war":               {engine.CommandSetWar, fields("target_player_id", "at_war")},
	"establish-contact":     {engine.CommandEstablishContact, fields("target_player_id")},
	"propose-agreement":     {engine.CommandProposeAgreement, fields("proposal_id", "target_player_id", "kind")},
	"respond-agreement":     {engine.CommandRespondAgreement, fields("proposal_id", "accept")},
	"set-government":        {engine.CommandSetGovernment, fields("government_id")},
	"set-science-rate":      {engine.CommandSetScienceRate, fields("science_percent")},
	"set-public-works-rate": {engine.CommandSetPublicWorksRate, fields("public_works_percent")},
	"set-social-policy":     {engine.CommandSetSocialPolicy, fields("workday", "wages", "rations")},
	"set-readiness":         {engine.CommandSetReadiness, fields("readiness")},
	"round-complete":        {engine.CommandRoundComplete, fields()},
	"set-research":          {engine.CommandSetResearch, fields("advance_id")},
	"set-research-goal":     {engine.CommandSetResearchGoal, fields("advance_id")},
	"move":                  {engine.CommandMoveUnit, fields("unit_id", "path", "target_x", "target_y")},
	"set-movement-order":    {engine.CommandSetMovementOrder, fields("unit_id", "target_x", "target_y")},
	"cancel-movement-order": {engine.CommandCancelMovementOrder, fields("unit_id")},
	"disband":               {engine.CommandDisband, fields("unit_id")},
	"embark":                {engine.CommandEmbark, fields("unit_id", "transport_id")},
	"disembark":             {engine.CommandDisembark, fields("unit_id", "tile_id")},
	"gift-unit":             {engine.CommandGiftUnit, fields("unit_id", "target_player_id")},
	"pirate-trade":          {engine.CommandPirateTrade, fields("unit_id", "route_id")},
	"space-launch":          {engine.CommandSpaceLaunch, fields("unit_id", "target_tile_id")},
	"fortify":               {engine.CommandFortify, fields("unit_id")},
	"sleep":                 {engine.CommandSleep, fields("unit_id")},
	"wake":                  {engine.CommandWake, fields("unit_id")},
	"rename-city":           {engine.CommandRenameCity, fields("city_id", "name")},
	"disband-city":          {engine.CommandDisbandCity, fields("city_id")},
	"found-city":            {engine.CommandFoundCity, fields("unit_id", "city_id", "name")},
	"army-create":           {engine.CommandArmyCreate, fields("army_id", "unit_ids")},
	"army-join":             {engine.CommandArmyJoin, fields("army_id", "unit_id")},
	"army-leave":            {engine.CommandArmyLeave, fields("army_id", "unit_id")},
	"combat":                {engine.CommandCombat, fields("attacker_army_id", "defender_army_id")},
	"nuclear-attack":        {engine.CommandNuclearAttack, fields("unit_id", "target_tile_id")},
	"establish-embassy":     {engine.CommandEstablishEmbassy, fields("unit_id", "city_id")},
	"soothsaying":           {engine.CommandSoothsaying, fields("unit_id", "city_id")},
	"cause-unhappiness":     {engine.CommandCauseUnhappiness, fields("unit_id", "city_id")},
	"advertise":             {engine.CommandAdvertise, fields("unit_id", "city_id")},
	"incite-revolution":     {engine.CommandInciteRevolution, fields("unit_id", "city_id")},
	"indulgence-sales":      {engine.CommandIndulgenceSales, fields("unit_id", "city_id")},
	"slave-uprising":        {engine.CommandSlaveUprising, fields("unit_id", "city_id")},
	"throw-party":           {engine.CommandThrowParty, fields("unit_id", "city_id")},
	"steal-technology":      {engine.CommandStealTechnology, fields("unit_id", "city_id", "advance_id")},
	"terraform":             {engine.CommandTerraform, fields("unit_id", "tile_id", "target_terrain")},
	"build-improvement":     {engine.CommandBuildImprovement, fields("unit_id", "tile_id", "improvement_id")},
	"cancel-terraform":      {engine.CommandCancelTerraform, fields("tile_id")},
	"assassinate-ruler":     {engine.CommandAssassinateRuler, fields("unit_id", "city_id")},
	"bio-terror":            {engine.CommandBioTerror, fields("unit_id", "city_id")},
	"bombard":               {engine.CommandBombard, fields("unit_id", "target_tile_id")},
	"convert-city":          {engine.CommandConvertCity, fields("unit_id", "city_id")},
	"create-franchise":      {engine.CommandCreateFranchise, fields("unit_id", "city_id")},
	"create-park":           {engine.CommandCreatePark, fields("unit_id", "tile_id")},
	"expel":                 {engine.CommandExpel, fields("unit_id", "target_unit_id")},
	"injoin":                {engine.CommandInjoin, fields("unit_id", "city_id")},
	"investigate-city":      {engine.CommandInvestigateCity, fields("unit_id", "city_id")},
	"nano-terror":           {engine.CommandNanoTerror, fields("unit_id", "city_id")},
	"plague":                {engine.CommandPlague, fields("unit_id", "city_id")},
	"plant-nuke":            {engine.CommandPlantNuke, fields("unit_id", "city_id")},
	"reform":                {engine.CommandReform, fields("unit_id", "city_id")},
	"settler-slave-raid":    {engine.CommandSettlerSlaveRaid, fields("unit_id", "city_id")},
	"slave-raid":            {engine.CommandSlaveRaid, fields("unit_id", "city_id")},
	"sue":                   {engine.CommandSue, fields("unit_id", "city_id", "target_unit_id")},
	"underground-railway":   {engine.CommandUndergroundRailway, fields("unit_id", "city_id")},
	"upgrade-unit":          {engine.CommandUpgradeUnit, fields("unit_id", "upgrade_to")},
	"pillage":               {engine.CommandPillage, fields("unit_id")},
	"queue-production":      {engine.CommandQueueProduction, fields("city_id", "item")},
	"cancel-production":     {engine.CommandCancelProduction, fields("city_id", "item_id", "entry_id")},
	"edit-production":       {engine.CommandEditProduction, fields("city_id", "entries")},
	"rush-buy":              {engine.CommandRushBuy, fields("city_id")},
	"set-specialists":       {engine.CommandSetSpecialists, fields("city_id", "specialists")},
	"create-trade-route":    {engine.CommandCreateTradeRoute, fields("source_city_id", "destination_city_id")},
	"leave":                 {engine.CommandLeaveGame, fields()},
	"end-early":             {engine.CommandEndGame, fields()},
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
// @Failure 413 {object} map[string]string "command body exceeds 256 KiB"
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
	r.Body = http.MaxBytesReader(w, r.Body, maxCommandBodyBytes)
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		writeBodyError(w, err)
		return
	}
	if len(raw) == 0 {
		writeError(w, http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeBodyError(w, err)
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

func writeBodyError(w http.ResponseWriter, err error) {
	var oversized *http.MaxBytesError
	if errors.As(err, &oversized) {
		writeError(w, http.StatusRequestEntityTooLarge)
		return
	}
	writeError(w, http.StatusBadRequest)
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
