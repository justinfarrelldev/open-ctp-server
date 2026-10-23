// Package service owns the one authenticated transactional gameplay mutation gateway.
package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

const (
	lockGameQuery           = `SELECT id, status, owner_account_id FROM games WHERE id = $1 FOR UPDATE`
	lockPlayerQuery         = `SELECT id, controller_kind FROM game_players WHERE game_id = $1 AND account_id = $2 FOR UPDATE`
	lockAIPlayerQuery       = `SELECT id, controller_kind FROM game_players WHERE game_id = $1 AND account_id = $2 FOR UPDATE`
	receiptQuery            = `SELECT game_id, actor_player_id, idempotency_key, payload_sha256, result_json::text AS result_json, resulting_revision, resulting_event_sequence FROM gameplay_command_receipts WHERE game_id = $1 AND actor_player_id = $2 AND idempotency_key = $3`
	aggregateForUpdateQuery = `SELECT state.game_id, state.state_schema_version, state.revision, state.last_event_sequence, state.canonical_state_json::text AS canonical_state_json, snapshot.ruleset_canonical_json::text AS ruleset_canonical_json, snapshot.ruleset_content_hash, snapshot.ruleset_version FROM world_turn_states AS state JOIN world_snapshots AS snapshot ON snapshot.game_id = state.game_id WHERE state.game_id = $1 FOR UPDATE OF state`
	updateAggregateQuery    = `UPDATE world_turn_states SET canonical_state_json = $1::jsonb, revision = revision + 1, last_event_sequence = $2, updated_at = NOW() WHERE game_id = $3 AND state_schema_version = 1 AND revision = $4`
	insertReceiptQuery      = `INSERT INTO gameplay_command_receipts ( game_id, actor_player_id, idempotency_key, payload_sha256, result_json, resulting_revision, resulting_event_sequence ) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`
	insertEventQuery        = `INSERT INTO gameplay_events ( game_id, sequence, revision, ordinal, event_id, event_type, payload_json ) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`
	insertOutboxQuery       = `INSERT INTO gameplay_event_outbox (game_id, sequence) VALUES ($1, $2)`
	updateGameStatusQuery   = `UPDATE games SET status = $1, updated_at = NOW() WHERE id = $2`
	updateControllerQuery   = `UPDATE game_players SET controller_kind = $1, controller_changed_at = NOW() WHERE game_id = $2 AND id = $3`
	endGameQuery            = `UPDATE games SET status = 'ended', ended_at = NOW(), ended_by = $1, end_reason = $2, updated_at = NOW() WHERE id = $3`
	terminalWinnerQuery     = `SELECT id FROM game_players WHERE game_id = $1 AND account_id = $2`
)

type CommandRequest struct {
	GameID           int64
	ExpectedRevision int64
	IdempotencyKey   string
	CommandID        string
	Type             string
	Payload          json.RawMessage
}

type Result struct {
	Revision         int64                `json:"revision"`
	ReceiptID        string               `json:"receipt_id"`
	EventCursor      int64                `json:"event_cursor"`
	ViewerProjection json.RawMessage      `json:"viewer_projection"`
	Combat           *engine.CombatResult `json:"combat,omitempty"`
	Terminal         bool                 `json:"terminal,omitempty"`
	Receipt          json.RawMessage      `json:"-"`
}

type Error struct {
	Status        int
	Message       string // private; HTTP must use PublicRejection
	Code          string
	CorrelationID string
	Cause         error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }
func StatusCode(err error) int {
	var target *Error
	if errors.As(err, &target) {
		return target.Status
	}
	return 500
}
func commandError(status int, message string) error { return &Error{Status: status, Message: message} }

type Service struct {
	store          *store.Store
	sources        engine.Sources
	now            func() time.Time
	ObserveCommand func(CommandTiming)
}

func New(commandStore *store.Store, sources engine.Sources) *Service {
	return &Service{store: commandStore, sources: sources, now: time.Now, ObserveCommand: logCommandTiming}
}

// ExecuteCommand is the only accepted gameplay mutation path. The session is server-authenticated
// input; no actor, combat result, visibility, or RNG authority is accepted from CommandRequest.
// ExecuteCommand is the only accepted human gameplay mutation path. The session is server-authenticated
// input; no actor, combat result, visibility, or RNG authority is accepted from CommandRequest.
func (s *Service) ExecuteCommand(ctx context.Context, session *auth.Session, request CommandRequest) (result Result, err error) {
	defer func() { err = attachDiagnostic(err, request) }()
	if err := validateRequest(ctx, session, request); err != nil {
		return Result{}, err
	}
	accountID := int64(session.AccountID)
	return s.executeControllerCommand(ctx, request, engine.ControllerHuman, &accountID, func(tx *sqlx.Tx) (lockedPlayer, error) {
		return lockPlayer(tx, request.GameID, accountID)
	})
}

// ExecuteAICommand is intentionally package-facing infrastructure rather than an HTTP endpoint.
// It uses the same transactional command gateway, receipts, events, revision, and authority checks.
func (s *Service) ExecuteAICommand(ctx context.Context, actorPlayerID string, request CommandRequest) (Result, error) {
	if err := validateAIRequest(ctx, actorPlayerID, request); err != nil {
		return Result{}, err
	}
	accountID := strings.TrimPrefix(actorPlayerID, "account-")
	actorID, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil || actorID < 1 {
		return Result{}, commandError(400, "AI actor is invalid")
	}
	return s.executeControllerCommand(ctx, request, engine.ControllerAI, &actorID, func(tx *sqlx.Tx) (lockedPlayer, error) {
		return lockAIPlayer(tx, request.GameID, actorID)
	})
}

func (s *Service) executeControllerCommand(ctx context.Context, request CommandRequest, controller string, actorAccountID *int64, lockActor func(*sqlx.Tx) (lockedPlayer, error)) (result Result, resultErr error) {
	if s == nil || s.store == nil {
		return Result{}, fmt.Errorf("gameplay command service is not configured")
	}
	meter := newCommandMeter(s.now, request)
	outcome := "accepted"
	defer func() {
		if resultErr != nil {
			outcome = rejectionCode(resultErr)
		}
		metric := meter.finish(outcome)
		if s.ObserveCommand != nil {
			s.ObserveCommand(metric)
		}
	}()
	tx, err := s.store.BeginCommand()
	if err != nil {
		return Result{}, fmt.Errorf("beginning gameplay command: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	game, err := lockGame(tx, request.GameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, commandError(404, "game not found")
		}
		return Result{}, err
	}
	if game.Status != "started" {
		return Result{}, commandError(409, "game is not accepting gameplay commands")
	}
	player, err := lockActor(tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, commandError(403, "game player membership is required")
		}
		return Result{}, err
	}
	if player.ControllerKind != string(controller) {
		return Result{}, commandError(403, "player controller is not authorized to submit this command")
	}
	if commandRequiresOwner(request.Type) && (actorAccountID == nil || game.OwnerAccountID == nil || *game.OwnerAccountID != *actorAccountID) {
		return Result{}, commandError(403, "game owner authorization is required")
	}
	payloadSHA := payloadHash(request)
	receipt, err := findReceipt(tx, request.GameID, player.ID, request.IdempotencyKey)
	if err == nil {
		if receipt.PayloadSHA256 != payloadSHA {
			return Result{}, codedError(409, "idempotency_conflict", "idempotency key was already used with a different payload")
		}
		result, err := decodeStoredResult(receipt.ResultJSON)
		if err != nil {
			return Result{}, fmt.Errorf("decoding stored gameplay receipt: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return Result{}, fmt.Errorf("finalizing replayed gameplay command: %w", err)
		}
		committed = true
		outcome = "replayed"
		return result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}
	aggregate, err := store.LoadAggregateForUpdate(tx, request.GameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, commandError(404, "gameplay state not found")
		}
		return Result{}, err
	}
	if aggregate.Revision != request.ExpectedRevision {
		return Result{}, codedError(409, "revision_conflict", "gameplay revision is stale")
	}
	state, err := engine.DecodeCanonicalStateJSON(aggregate.CanonicalStateJSON)
	if err != nil {
		return Result{}, commandError(422, "gameplay state is invalid")
	}
	if state.Revision != aggregate.Revision || state.EventSequence != aggregate.LastEventSequence {
		return Result{}, codedError(409, "revision_conflict", "gameplay state revision is stale")
	}
	actorID := strconv.FormatInt(player.ID, 10)
	if actorAccountID != nil {
		canonicalActor := fmt.Sprintf("account-%d", *actorAccountID)
		if _, exists := statePlayer(state, canonicalActor); exists {
			actorID = canonicalActor
		}
	}
	if err := validateAuthority(state, actorID, controller, request.Type); err != nil {
		return Result{}, err
	}
	// Preserve pre-transition sight as well as post-transition sight. This also
	// safely upgrades legacy saves with empty/incomplete memory, but only for
	// cells currently visible before this command, never hidden backfills.
	meter.step(&meter.metric.Visibility)
	state, err = visibility.Update(state)
	if err != nil {
		return Result{}, fmt.Errorf("recording pre-command visibility: %w", err)
	}
	meter.step(&meter.metric.Engine)
	next, engineEvents, applied, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: request.CommandID, ActorPlayerID: actorID, IdempotencyKey: request.IdempotencyKey, Type: request.Type, Payload: canonicalPayload(request.Payload)}, rulesPin(aggregate.Ruleset), s.sources)
	if err != nil {
		return Result{}, engineRejection(err)
	}
	meter.step(&meter.metric.Visibility)
	next, err = visibility.Update(next)
	if err != nil {
		return Result{}, fmt.Errorf("updating gameplay visibility: %w", err)
	}
	projection, err := visibility.AIInput(next, actorID)
	if err != nil {
		return Result{}, fmt.Errorf("projecting gameplay result: %w", err)
	}
	meter.step(&meter.metric.Persistence)
	nextJSON, err := engine.CanonicalStateJSON(next)
	if err != nil {
		return Result{}, fmt.Errorf("encoding gameplay state: %w", err)
	}
	result = Result{Revision: applied.Revision, ReceiptID: request.IdempotencyKey, EventCursor: applied.EventSequence, ViewerProjection: projection, Combat: applied.Combat, Terminal: next.Terminal.Ended}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return Result{}, fmt.Errorf("encoding gameplay result: %w", err)
	}
	persistedEvents, err := toStoreEvents(request.GameID, engineEvents, applied.Revision)
	if err != nil {
		return Result{}, fmt.Errorf("encoding gameplay events: %w", err)
	}
	after := aggregate
	after.Revision, after.LastEventSequence, after.CanonicalStateJSON = next.Revision, next.EventSequence, nextJSON
	receipt = store.CommandReceipt{GameID: request.GameID, ActorPlayerID: player.ID, IdempotencyKey: request.IdempotencyKey, PayloadSHA256: payloadSHA, ResultJSON: resultJSON, ResultingRevision: result.Revision, ResultingEventSequence: result.EventCursor}
	step, recordsExploration, err := acceptedAIExplorationStep(request, controller, receipt)
	if err != nil {
		return Result{}, fmt.Errorf("deriving AI exploration history: %w", err)
	}
	var exploration []store.AIExplorationStep
	if recordsExploration {
		exploration = append(exploration, step)
	}
	if err := store.PersistCommandInTx(tx, aggregate.Revision, after, receipt, persistedEvents, exploration...); err != nil {
		if errors.Is(err, store.ErrStaleRevision) {
			return Result{}, codedError(409, "revision_conflict", "gameplay revision is stale")
		}
		return Result{}, fmt.Errorf("persisting gameplay command: %w", err)
	}
	if request.Type == engine.CommandTakeoverAI {
		if _, err := tx.Exec(updateControllerQuery, string(engine.ControllerAI), request.GameID, player.ID); err != nil {
			return Result{}, fmt.Errorf("persisting AI controller takeover: %w", err)
		}
	}
	if next.Terminal.Ended {
		winnerID, err := terminalWinnerID(tx, request.GameID, next.Terminal.WinningPlayerID)
		if err != nil {
			return Result{}, fmt.Errorf("terminal winner player is invalid: %w", err)
		}
		if _, err := tx.Exec(endGameQuery, winnerID, next.Terminal.Reason, request.GameID); err != nil {
			return Result{}, fmt.Errorf("persisting terminal gameplay status: %w", err)
		}
	} else if _, err := tx.Exec(updateGameStatusQuery, "started", request.GameID); err != nil {
		return Result{}, fmt.Errorf("persisting gameplay status: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("finalizing gameplay command: %w", err)
	}
	committed = true
	result.Receipt = resultJSON
	return result, nil
}

func validateRequest(ctx context.Context, session *auth.Session, request CommandRequest) error {
	if err := ctx.Err(); err != nil {
		return commandError(429, "gameplay command deadline exceeded")
	}
	if session == nil || session.IsExpired() {
		return commandError(401, "authenticated session is required")
	}
	if request.GameID <= 0 || request.ExpectedRevision < 0 || strings.TrimSpace(request.IdempotencyKey) == "" || strings.TrimSpace(request.CommandID) == "" || strings.TrimSpace(request.Type) == "" || !json.Valid(request.Payload) {
		return commandError(400, "malformed gameplay command")
	}
	return nil
}

func validateAIRequest(ctx context.Context, actorPlayerID string, request CommandRequest) error {
	if err := ctx.Err(); err != nil {
		return commandError(429, "gameplay command deadline exceeded")
	}
	accountID := strings.TrimPrefix(actorPlayerID, "account-")
	if accountID == actorPlayerID || strings.TrimSpace(accountID) == "" {
		return commandError(400, "AI actor player is invalid")
	}
	if parsed, err := strconv.ParseInt(accountID, 10, 64); err != nil || parsed < 1 {
		return commandError(400, "AI actor player is invalid")
	}
	if request.GameID <= 0 || request.ExpectedRevision < 0 || strings.TrimSpace(request.IdempotencyKey) == "" || strings.TrimSpace(request.CommandID) == "" || strings.TrimSpace(request.Type) == "" || !json.Valid(request.Payload) {
		return commandError(400, "malformed gameplay command")
	}
	return nil
}

type lockedGame struct {
	ID             int64  `db:"id"`
	Status         string `db:"status"`
	OwnerAccountID *int64 `db:"owner_account_id"`
}
type lockedPlayer struct {
	ID             int64  `db:"id"`
	ControllerKind string `db:"controller_kind"`
}

func lockGame(tx *sqlx.Tx, gameID int64) (lockedGame, error) {
	var game lockedGame
	return game, tx.Get(&game, lockGameQuery, gameID)
}
func lockPlayer(tx *sqlx.Tx, gameID, accountID int64) (lockedPlayer, error) {
	var player lockedPlayer
	return player, tx.Get(&player, lockPlayerQuery, gameID, accountID)
}

func lockAIPlayer(tx *sqlx.Tx, gameID, playerID int64) (lockedPlayer, error) {
	var player lockedPlayer
	return player, tx.Get(&player, lockAIPlayerQuery, gameID, playerID)
}
func findReceipt(tx *sqlx.Tx, gameID, actorID int64, key string) (store.CommandReceipt, error) {
	var receipt store.CommandReceipt
	return receipt, tx.Get(&receipt, receiptQuery, gameID, actorID, key)
}

func acceptedAIExplorationStep(request CommandRequest, controller string, receipt store.CommandReceipt) (store.AIExplorationStep, bool, error) {
	if controller != string(engine.ControllerAI) || request.Type != engine.CommandMoveUnit {
		return store.AIExplorationStep{}, false, nil
	}
	var payload struct {
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(request.Payload, &payload); err != nil || len(payload.Path) == 0 || strings.TrimSpace(payload.Path[0]) == "" {
		return store.AIExplorationStep{}, false, errors.New("accepted AI move is missing its path step")
	}
	return store.AIExplorationStep{
		GameID:            receipt.GameID,
		ActorPlayerID:     receipt.ActorPlayerID,
		Sequence:          receipt.ResultingEventSequence,
		CommandType:       request.Type,
		DestinationTileID: payload.Path[0],
	}, true, nil
}

// commandRequiresOwner reserves authority validation for owner-only lifecycle commands;
// mechanics remain the engine's responsibility and unsupported commands are rejected there.
func commandRequiresOwner(commandType string) bool { return commandType == "end-game" }

func validateAuthority(state engine.State, actorID, controller, commandType string) error {
	if state.Terminal.Ended {
		return commandError(409, "game is already terminal")
	}
	for _, player := range state.Players {
		if player.ID != actorID || player.Lifecycle != engine.PlayerActive || player.Controller != controller {
			continue
		}
		if commandType == engine.CommandEndGame || state.Phase.Mode == "simultaneous" || (state.Phase.Mode == "hybrid" && state.Policy.Stage != "conflict") {
			return nil
		}
		if len(state.Phase.ActivePlayerOrder) == 0 || state.Phase.ActivePlayerIndex < 0 || state.Phase.ActivePlayerIndex >= len(state.Phase.ActivePlayerOrder) || state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex] != actorID {
			return codedError(422, "phase_restricted", "command is not allowed in this phase")
		}
		return nil
	}
	return commandError(403, "player is not authorized to control this command")
}

func statePlayer(state engine.State, id string) (engine.Player, bool) {
	for _, player := range state.Players {
		if player.ID == id {
			return player, true
		}
	}
	return engine.Player{}, false
}

func terminalWinnerID(tx *sqlx.Tx, gameID int64, winner string) (int64, error) {
	if tx == nil || gameID < 1 || winner == "" {
		return 0, errors.New("terminal winner is invalid")
	}
	if accountID, canonical := strings.CutPrefix(winner, "account-"); canonical {
		id, err := strconv.ParseInt(accountID, 10, 64)
		if err != nil || id < 1 {
			return 0, errors.New("terminal winner account is invalid")
		}
		var playerID int64
		if err := tx.Get(&playerID, terminalWinnerQuery, gameID, id); err != nil {
			return 0, err
		}
		return playerID, nil
	}
	id, err := strconv.ParseInt(winner, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("terminal winner player is invalid")
	}
	return id, nil
}

func rulesPin(pin store.RulesetPin) engine.RulesPin {
	result := engine.RulesPin{Version: pin.Version, ContentHash: pin.ContentHash}
	canonical, err := rules.CanonicalDefinitions(pin.CanonicalJSON)
	if err != nil {
		return result
	}
	digest := sha256.Sum256(canonical)
	if hex.EncodeToString(digest[:]) == pin.ContentHash {
		result.Definitions = json.RawMessage(canonical)
	}
	return result
}

func canonicalPayload(payload json.RawMessage) json.RawMessage {
	var value any
	if json.Unmarshal(payload, &value) != nil {
		return nil
	}
	body, _ := json.Marshal(value)
	return body
}
func payloadHash(request CommandRequest) string {
	body, _ := json.Marshal(struct {
		GameID           int64           `json:"game_id"`
		ExpectedRevision int64           `json:"expected_revision"`
		CommandID        string          `json:"command_id"`
		Type             string          `json:"type"`
		Payload          json.RawMessage `json:"payload"`
	}{request.GameID, request.ExpectedRevision, request.CommandID, request.Type, canonicalPayload(request.Payload)})
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
func decodeStoredResult(body []byte) (Result, error) {
	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		return Result{}, err
	}
	result.Receipt = append([]byte(nil), body...)
	return result, nil
}
func toStoreEvents(gameID int64, events []engine.Event, revision int64) ([]store.Event, error) {
	output := make([]store.Event, 0, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return nil, err
		}
		output = append(output, store.Event{GameID: gameID, Sequence: event.Sequence, Revision: revision, Ordinal: event.Ordinal, ID: event.ID, Type: event.Type, PayloadJSON: payload})
	}
	return output, nil
}
