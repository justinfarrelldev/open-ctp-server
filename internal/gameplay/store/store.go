// Package store persists the single mutable gameplay aggregate. Task 5 must lock the
// aggregate row with LoadAggregateForUpdate and use PersistCommandInTx in its command
// transaction; optimistic revision checks complement the PostgreSQL row lock.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	StateSchemaLegacy = 0
	StateSchemaV1     = 1
)

var (
	ErrStaleRevision = errors.New("gameplay aggregate revision is stale")
	ErrUnknownSchema = errors.New("unknown gameplay state schema version")
)

const selectAggregateQuery = `SELECT
	state.game_id,
	state.state_schema_version,
	state.revision,
	state.last_event_sequence,
	state.canonical_state_json::text AS canonical_state_json,
	snapshot.ruleset_canonical_json::text AS ruleset_canonical_json,
	snapshot.ruleset_content_hash,
	snapshot.ruleset_version
FROM world_turn_states AS state
JOIN world_snapshots AS snapshot ON snapshot.game_id = state.game_id
WHERE state.game_id = $1`

const selectAggregateForUpdateQuery = selectAggregateQuery + ` FOR UPDATE OF state`

const PendingTakeoversQuery = `SELECT
	state.game_id,
	state.revision,
	player.value->>'id' AS player_id
FROM world_turn_states AS state
JOIN games ON games.id = state.game_id
CROSS JOIN LATERAL jsonb_array_elements(state.canonical_state_json->'players') AS player(value)
WHERE state.state_schema_version = 1
	AND games.status = 'started'
	AND COALESCE(player.value->>'controller', '') = 'human'
	AND COALESCE(player.value->>'takeover_pending', 'false') = 'true'
ORDER BY state.game_id ASC, player_id ASC`

const DueLongTurnDeadlinesQuery = `SELECT
	game_id,
	revision,
	canonical_state_json #>> '{policy,deadline,player_id}' AS player_id,
	(canonical_state_json #>> '{policy,deadline,at_unix}')::bigint AS deadline_at
FROM world_turn_states
WHERE state_schema_version = 1
	AND canonical_state_json #>> '{phase,mode}' = 'long_turn'
	AND (canonical_state_json #>> '{policy,deadline,at_unix}')::bigint <= $1
ORDER BY deadline_at ASC, game_id ASC`

// Discovery transfers scheduling metadata only, never terrain, fog, rules or
// unit/city state for unrelated games. The AI still fetches its full authorized
// projection before choosing a command; these fields confer no execution rights.
const aiSchedulingSelect = `SELECT
	state.game_id,
	state.revision,
	jsonb_build_object(
		'players', state.canonical_state_json->'players',
		'phase', state.canonical_state_json->'phase',
		'policy', jsonb_build_object(
			'stage', state.canonical_state_json #> '{policy,stage}',
			'early_ends', state.canonical_state_json #> '{policy,early_ends}'),
		'terminal', state.canonical_state_json->'terminal'
	)::text AS canonical_state_json
FROM world_turn_states AS state
JOIN games ON games.id = state.game_id
WHERE state.state_schema_version = 1
	AND games.status = 'started'`

const StartedGameplayAggregatesQuery = aiSchedulingSelect + `
ORDER BY state.game_id ASC`

const selectReceiptQuery = `SELECT
  game_id, actor_player_id, idempotency_key, payload_sha256,
  result_json::text AS result_json, resulting_revision, resulting_event_sequence
FROM gameplay_command_receipts
WHERE game_id = $1 AND actor_player_id = $2 AND idempotency_key = $3`

const selectAIExplorationHistoryQuery = `SELECT
  game_id, actor_player_id, sequence, command_type, destination_tile_id
FROM gameplay_ai_exploration_history
WHERE game_id = $1 AND actor_player_id = $2
ORDER BY sequence ASC`

const selectGamePlayerIDQuery = `SELECT id FROM game_players
WHERE game_id = $1 AND account_id = $2`

const updateAggregateQuery = `UPDATE world_turn_states
SET
	canonical_state_json = $1::jsonb,
	revision = revision + 1,
	last_event_sequence = $2,
	updated_at = NOW()
WHERE game_id = $3 AND state_schema_version = 1 AND revision = $4`

const insertReceiptQuery = `INSERT INTO gameplay_command_receipts (
	game_id, actor_player_id, idempotency_key, payload_sha256,
	result_json, resulting_revision, resulting_event_sequence
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`

const insertAIExplorationStepQuery = `INSERT INTO gameplay_ai_exploration_history (
	game_id, actor_player_id, sequence, command_type, destination_tile_id
) VALUES ($1, $2, $3, $4, $5)`

const insertEventQuery = `INSERT INTO gameplay_events (
	game_id, sequence, revision, ordinal, event_id, event_type, payload_json
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`

const insertOutboxQuery = `INSERT INTO gameplay_event_outbox (game_id, sequence) VALUES ($1, $2)`

// RulesetPin is the immutable canonical import artifact used by a v1 aggregate.
type RulesetPin struct {
	CanonicalJSON []byte
	ContentHash   string
	Version       string
}

// Aggregate is the one mutable gameplay authority per game.
type Aggregate struct {
	GameID             int64
	StateSchemaVersion int
	Revision           int64
	LastEventSequence  int64
	CanonicalStateJSON []byte
	Ruleset            RulesetPin
}

type PendingTakeover struct {
	GameID   int64  `db:"game_id"`
	Revision int64  `db:"revision"`
	PlayerID string `db:"player_id"`
}

// DueDeadline is a long-turn deadline discovered from canonical policy state.
type DueDeadline struct {
	GameID     int64  `db:"game_id"`
	Revision   int64  `db:"revision"`
	PlayerID   string `db:"player_id"`
	DeadlineAt int64  `db:"deadline_at"`
}

// StartedAggregate is the minimal canonical gameplay state needed by autonomous workers.
type StartedAggregate struct {
	GameID             int64  `db:"game_id"`
	Revision           int64  `db:"revision"`
	CanonicalStateJSON []byte `db:"canonical_state_json"`
}

type CommandReceipt struct {
	GameID                 int64  `db:"game_id"`
	ActorPlayerID          int64  `db:"actor_player_id"`
	IdempotencyKey         string `db:"idempotency_key"`
	PayloadSHA256          string `db:"payload_sha256"`
	ResultJSON             []byte `db:"result_json"`
	ResultingRevision      int64  `db:"resulting_revision"`
	ResultingEventSequence int64  `db:"resulting_event_sequence"`
}

// AIExplorationStep records one accepted AI movement destination for restart-safe planning.
type AIExplorationStep struct {
	GameID            int64  `db:"game_id"`
	ActorPlayerID     int64  `db:"actor_player_id"`
	Sequence          int64  `db:"sequence"`
	CommandType       string `db:"command_type"`
	DestinationTileID string `db:"destination_tile_id"`
}

type Event struct {
	GameID      int64
	Sequence    int64
	Revision    int64
	Ordinal     int
	ID          string
	Type        string
	PayloadJSON []byte
}

type Store struct{ db *sqlx.DB }

func New(db *sqlx.DB) *Store { return &Store{db: db} }

func (s *Store) BeginCommand() (*sqlx.Tx, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("gameplay store requires a database")
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("starting gameplay command transaction: %w", err)
	}
	return tx, nil
}

func (s *Store) ReadAggregate(gameID int64) (Aggregate, error) {
	if s == nil || s.db == nil {
		return Aggregate{}, errors.New("gameplay store requires a database")
	}
	return loadAggregate(s.db, selectAggregateQuery, gameID)
}

// FindPendingTakeovers reads persisted pending human-to-AI transitions without mutating state.
func (s *Store) FindPendingTakeovers() ([]PendingTakeover, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("gameplay store requires a database")
	}
	var takeovers []PendingTakeover
	if err := s.db.Select(&takeovers, PendingTakeoversQuery); err != nil {
		return nil, fmt.Errorf("reading pending takeovers: %w", err)
	}
	return takeovers, nil
}

// FindDueLongTurnDeadlines reads due long-turn deadlines from canonical policy state.
// Callers must submit the pass through the command gateway; this query never mutates state.
func (s *Store) FindDueLongTurnDeadlines(nowUnix int64) ([]DueDeadline, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("gameplay store requires a database")
	}
	var deadlines []DueDeadline
	if err := s.db.Select(&deadlines, DueLongTurnDeadlinesQuery, nowUnix); err != nil {
		return nil, fmt.Errorf("reading due long-turn deadlines: %w", err)
	}
	return deadlines, nil
}

// FindStartedGameplayAggregates reads started games that may need autonomous workers.
func (s *Store) FindStartedGameplayAggregates() ([]StartedAggregate, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("gameplay store requires a database")
	}
	var aggregates []StartedAggregate
	if err := s.db.Select(&aggregates, StartedGameplayAggregatesQuery); err != nil {
		return nil, fmt.Errorf("reading started gameplay aggregates: %w", err)
	}
	return aggregates, nil
}

// LoadAggregateForUpdate is intentionally transaction-only: callers use the locked
// aggregate plus PersistCommandInTx for a single expected-revision command commit.
func LoadAggregateForUpdate(tx *sqlx.Tx, gameID int64) (Aggregate, error) {
	if tx == nil {
		return Aggregate{}, errors.New("gameplay store requires a transaction")
	}
	return loadAggregate(tx, selectAggregateForUpdateQuery, gameID)
}

func FindReceipt(tx *sqlx.Tx, gameID, actorPlayerID int64, key string) (CommandReceipt, error) {
	if tx == nil {
		return CommandReceipt{}, errors.New("gameplay store requires a transaction")
	}
	var receipt CommandReceipt
	if err := tx.Get(&receipt, selectReceiptQuery, gameID, actorPlayerID, key); err != nil {
		return CommandReceipt{}, err
	}
	return receipt, nil
}

// AIExplorationHistory returns accepted movement metadata for exactly one game and AI actor.
func (s *Store) AIExplorationHistory(gameID, actorPlayerID int64) ([]AIExplorationStep, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("gameplay store requires a database")
	}
	if gameID < 1 || actorPlayerID < 1 {
		return nil, errors.New("AI exploration history scope is invalid")
	}
	var steps []AIExplorationStep
	if err := s.db.Select(&steps, selectAIExplorationHistoryQuery, gameID, actorPlayerID); err != nil {
		return nil, fmt.Errorf("reading AI exploration history: %w", err)
	}
	return steps, nil
}

// GamePlayerID resolves a game-scoped player identity from its account identity.
func (s *Store) GamePlayerID(gameID, accountID int64) (int64, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("gameplay store requires a database")
	}
	if gameID < 1 || accountID < 1 {
		return 0, errors.New("game player scope is invalid")
	}
	var playerID int64
	if err := s.db.Get(&playerID, selectGamePlayerIDQuery, gameID, accountID); err != nil {
		return 0, fmt.Errorf("resolving game player: %w", err)
	}
	return playerID, nil
}

func (s *Store) PersistCommand(expectedRevision int64, after Aggregate, receipt CommandReceipt, events []Event, exploration ...AIExplorationStep) error {
	tx, err := s.BeginCommand()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := PersistCommandInTx(tx, expectedRevision, after, receipt, events, exploration...); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("finalizing gameplay command: %w", err)
	}
	committed = true
	return nil
}

// PersistCommandInTx stages one authoritative revision, receipt, events, and delivery
// outbox rows. It does not commit or notify; callers must commit before dispatching.
func PersistCommandInTx(tx *sqlx.Tx, expectedRevision int64, after Aggregate, receipt CommandReceipt, events []Event, exploration ...AIExplorationStep) error {
	if tx == nil {
		return errors.New("gameplay store requires a transaction")
	}
	if err := ValidateAggregate(after); err != nil {
		return err
	}
	if after.Revision != expectedRevision+1 {
		return fmt.Errorf("gameplay revision must advance exactly one from %d", expectedRevision)
	}
	if err := validateReceipt(after, receipt); err != nil {
		return err
	}
	if err := validateEvents(after, events); err != nil {
		return err
	}
	if err := validateAIExplorationSteps(after, receipt, exploration); err != nil {
		return err
	}
	result, err := tx.Exec(updateAggregateQuery, after.CanonicalStateJSON, after.LastEventSequence, after.GameID, expectedRevision)
	if err != nil {
		return fmt.Errorf("updating gameplay aggregate: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("confirming gameplay aggregate update: %w", err)
	}
	if rows != 1 {
		return ErrStaleRevision
	}
	if _, err := tx.Exec(insertReceiptQuery, receipt.GameID, receipt.ActorPlayerID, receipt.IdempotencyKey, receipt.PayloadSHA256, receipt.ResultJSON, receipt.ResultingRevision, receipt.ResultingEventSequence); err != nil {
		return fmt.Errorf("storing gameplay command receipt: %w", err)
	}
	for _, step := range exploration {
		if _, err := tx.Exec(insertAIExplorationStepQuery, step.GameID, step.ActorPlayerID, step.Sequence, step.CommandType, step.DestinationTileID); err != nil {
			return fmt.Errorf("storing AI exploration history: %w", err)
		}
	}
	for _, event := range events {
		if _, err := tx.Exec(insertEventQuery, event.GameID, event.Sequence, event.Revision, event.Ordinal, event.ID, event.Type, event.PayloadJSON); err != nil {
			return fmt.Errorf("storing gameplay event: %w", err)
		}
		if _, err := tx.Exec(insertOutboxQuery, event.GameID, event.Sequence); err != nil {
			return fmt.Errorf("storing gameplay event outbox row: %w", err)
		}
	}
	return nil
}

type aggregateGetter interface {
	Get(any, string, ...any) error
}

func loadAggregate(getter aggregateGetter, query string, gameID int64) (Aggregate, error) {
	if gameID <= 0 {
		return Aggregate{}, errors.New("game_id must be positive")
	}
	var row struct {
		GameID               int64   `db:"game_id"`
		StateSchemaVersion   int     `db:"state_schema_version"`
		Revision             int64   `db:"revision"`
		LastEventSequence    int64   `db:"last_event_sequence"`
		CanonicalStateJSON   string  `db:"canonical_state_json"`
		RulesetCanonicalJSON *string `db:"ruleset_canonical_json"`
		RulesetContentHash   *string `db:"ruleset_content_hash"`
		RulesetVersion       *string `db:"ruleset_version"`
	}
	if err := getter.Get(&row, query, gameID); err != nil {
		return Aggregate{}, err
	}
	aggregate := Aggregate{GameID: row.GameID, StateSchemaVersion: row.StateSchemaVersion, Revision: row.Revision, LastEventSequence: row.LastEventSequence, CanonicalStateJSON: []byte(row.CanonicalStateJSON)}
	if row.RulesetCanonicalJSON != nil {
		aggregate.Ruleset.CanonicalJSON = []byte(*row.RulesetCanonicalJSON)
	}
	if row.RulesetContentHash != nil {
		aggregate.Ruleset.ContentHash = *row.RulesetContentHash
	}
	if row.RulesetVersion != nil {
		aggregate.Ruleset.Version = *row.RulesetVersion
	}
	if aggregate.StateSchemaVersion != StateSchemaLegacy && aggregate.StateSchemaVersion != StateSchemaV1 {
		return Aggregate{}, fmt.Errorf("%w: %d", ErrUnknownSchema, aggregate.StateSchemaVersion)
	}
	if aggregate.StateSchemaVersion == StateSchemaV1 {
		if err := ValidateAggregate(aggregate); err != nil {
			return Aggregate{}, err
		}
	}
	return aggregate, nil
}

func ValidateAggregate(aggregate Aggregate) error {
	if aggregate.GameID <= 0 {
		return errors.New("game_id must be positive")
	}
	if aggregate.StateSchemaVersion != StateSchemaV1 {
		return fmt.Errorf("%w: %d", ErrUnknownSchema, aggregate.StateSchemaVersion)
	}
	if aggregate.Revision < 0 || aggregate.LastEventSequence < 0 {
		return errors.New("gameplay revision and sequence must be nonnegative")
	}
	if err := requireJSONObject("canonical state", aggregate.CanonicalStateJSON); err != nil {
		return err
	}
	if err := requireJSONObject("ruleset canonical JSON", aggregate.Ruleset.CanonicalJSON); err != nil {
		return err
	}
	if !isSHA256(aggregate.Ruleset.ContentHash) || strings.TrimSpace(aggregate.Ruleset.Version) == "" {
		return errors.New("v1 gameplay aggregate requires a pinned SHA-256 ruleset hash and version")
	}
	return nil
}

func requireJSONObject(name string, value []byte) error {
	if len(value) == 0 {
		return fmt.Errorf("%s must be non-empty canonical JSON", name)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(value, &object); err != nil || object == nil {
		return fmt.Errorf("%s must be a JSON object", name)
	}
	return nil
}

func validateReceipt(after Aggregate, receipt CommandReceipt) error {
	if receipt.GameID != after.GameID || receipt.ActorPlayerID <= 0 || strings.TrimSpace(receipt.IdempotencyKey) == "" || !isSHA256(receipt.PayloadSHA256) {
		return errors.New("invalid gameplay command receipt")
	}
	if receipt.ResultingRevision != after.Revision || receipt.ResultingEventSequence != after.LastEventSequence {
		return errors.New("gameplay command receipt result does not match aggregate")
	}
	return requireJSONObject("receipt result", receipt.ResultJSON)
}

func validateAIExplorationSteps(after Aggregate, receipt CommandReceipt, steps []AIExplorationStep) error {
	for _, step := range steps {
		if step.GameID != after.GameID || step.ActorPlayerID != receipt.ActorPlayerID || step.Sequence != receipt.ResultingEventSequence || step.Sequence < 1 || strings.TrimSpace(step.CommandType) == "" || strings.TrimSpace(step.DestinationTileID) == "" {
			return errors.New("invalid AI exploration history step")
		}
	}
	return nil
}

func validateEvents(after Aggregate, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	firstSequence := after.LastEventSequence - int64(len(events)) + 1
	for index, event := range events {
		if event.GameID != after.GameID || event.Revision != after.Revision || event.Sequence != firstSequence+int64(index) || event.Ordinal != index || strings.TrimSpace(event.ID) == "" || strings.TrimSpace(event.Type) == "" {
			return errors.New("gameplay events must be ordered and match the aggregate")
		}
		if err := requireJSONObject("event payload", event.PayloadJSON); err != nil {
			return err
		}
	}
	return nil
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}
