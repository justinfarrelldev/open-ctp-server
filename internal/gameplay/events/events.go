// Package events provides durable, viewer-projected gameplay event delivery.
package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/jmoiron/sqlx"
)

// Viewer identifies the authenticated player receiving a projection.
type Viewer struct {
	GameID    int64
	PlayerID  int64
	AccountID int64
}

// Event is a durable gameplay event read in authoritative sequence order.
type Event struct {
	GameID      int64
	Sequence    int64
	Revision    int64
	Ordinal     int
	Type        string
	PayloadJSON []byte
}

// SSEEvent is a fully projected event ready for SSE serialization.
type SSEEvent struct {
	ID   int64
	Name string
	Data []byte
}

// Projector transforms canonical durable data into a viewer-safe representation.
// Implementations must never return canonical data for a viewer.
type Projector interface {
	ProjectEvent(Viewer, Event) ([]byte, error)
	ProjectSnapshot(Viewer, []byte) ([]byte, error)
}

// ProjectorFuncs adapts projection functions for tests and HTTP integration.
type ProjectorFuncs struct {
	Event    func(Viewer, Event) ([]byte, error)
	Snapshot func(Viewer, []byte) ([]byte, error)
}

func (p ProjectorFuncs) ProjectEvent(viewer Viewer, event Event) ([]byte, error) {
	if p.Event == nil {
		return []byte(`{}`), nil
	}
	return p.Event(viewer, event)
}

func (p ProjectorFuncs) ProjectSnapshot(viewer Viewer, state []byte) ([]byte, error) {
	if p.Snapshot == nil {
		return []byte(`{}`), nil
	}
	return p.Snapshot(viewer, state)
}

// Wakeup is an optional latency signal. Durable database rows remain authoritative.
type Wakeup interface {
	Publish(gameID int64)
	Subscribe(gameID int64) (<-chan struct{}, func())
}

// Dispatcher emits latency wakeups for committed, undelivered outbox rows.
type Dispatcher struct {
	db     *sqlx.DB
	wakeup Wakeup
}

func NewDispatcher(db *sqlx.DB, wakeup Wakeup) *Dispatcher {
	return &Dispatcher{db: db, wakeup: wakeup}
}

type outboxEntry struct {
	GameID   int64 `db:"game_id"`
	Sequence int64 `db:"sequence"`
}

const undeliveredOutboxQuery = `SELECT o.game_id, o.sequence
FROM gameplay_event_outbox o
WHERE o.delivered_at IS NULL
ORDER BY o.game_id ASC, o.sequence ASC
FOR UPDATE SKIP LOCKED`

const markOutboxDeliveredQuery = `UPDATE gameplay_event_outbox
SET delivered_at = NOW(), delivery_attempts = delivery_attempts + 1
WHERE game_id = $1 AND sequence = $2 AND delivered_at IS NULL`

// DispatchOnce wakes listeners only for rows committed before this transaction.
// Marking happens after Publish; a failed transaction leaves rows retryable.
func (d *Dispatcher) DispatchOnce(ctx context.Context) error {
	if d == nil || d.db == nil {
		return errors.New("gameplay event dispatcher requires a database")
	}
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning event outbox dispatch: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Drain and close the SELECT before reusing this transaction's connection
	// for UPDATEs. PostgreSQL row locks remain held until the transaction ends.
	var pending []outboxEntry
	if err := tx.SelectContext(ctx, &pending, undeliveredOutboxQuery); err != nil {
		return fmt.Errorf("reading committed gameplay event outbox: %w", err)
	}
	for _, entry := range pending {
		if d.wakeup != nil {
			d.wakeup.Publish(entry.GameID)
		}
		if _, err := tx.ExecContext(ctx, markOutboxDeliveredQuery, entry.GameID, entry.Sequence); err != nil {
			return fmt.Errorf("marking gameplay event outbox delivery: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing gameplay event outbox delivery: %w", err)
	}
	committed = true
	return nil
}

// Store reads durable gameplay events and snapshots for a stream.
type Store struct{ db *sqlx.DB }

func NewStore(db *sqlx.DB) *Store { return &Store{db: db} }

const firstEventSequenceQuery = `SELECT min(sequence) FROM gameplay_events WHERE game_id = $1`
const eventsAfterCursorQuery = `SELECT game_id, sequence, revision, ordinal, event_type, payload_json
FROM gameplay_events
WHERE game_id = $1 AND sequence > $2
ORDER BY sequence ASC`
const snapshotForResyncQuery = `SELECT revision, last_event_sequence, canonical_state_json
FROM world_turn_states
WHERE game_id = $1 AND state_schema_version = 1`

// Resume returns strictly increasing durable events after cursor. An expired cursor
// is replaced by exactly one viewer-projected resync with the current revision.
func (s *Store) Resume(ctx context.Context, viewer Viewer, cursor int64, projector Projector) ([]SSEEvent, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("gameplay event store requires a database")
	}
	if viewer.GameID <= 0 || viewer.PlayerID <= 0 {
		return nil, errors.New("gameplay event viewer is required")
	}
	if cursor < 0 {
		return nil, errors.New("event cursor must be nonnegative")
	}
	if projector == nil {
		return nil, errors.New("gameplay event projector is required")
	}

	var first sql.NullInt64
	if err := s.db.QueryRowxContext(ctx, firstEventSequenceQuery, viewer.GameID).Scan(&first); err != nil {
		return nil, fmt.Errorf("reading first gameplay event sequence: %w", err)
	}
	if first.Valid && (cursor == 0 || cursor < first.Int64-1) {
		return s.resync(ctx, viewer, projector)
	}

	rows, err := s.db.QueryxContext(ctx, eventsAfterCursorQuery, viewer.GameID, cursor)
	if err != nil {
		return nil, fmt.Errorf("reading gameplay events after cursor: %w", err)
	}
	defer rows.Close()
	result := make([]SSEEvent, 0)
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.GameID, &event.Sequence, &event.Revision, &event.Ordinal, &event.Type, &event.PayloadJSON); err != nil {
			return nil, fmt.Errorf("scanning gameplay event: %w", err)
		}
		if event.Type == "terminal" {
			continue
		}
		data, err := projector.ProjectEvent(viewer, event)
		if err != nil {
			return nil, fmt.Errorf("projecting gameplay event: %w", err)
		}
		result = append(result, SSEEvent{ID: event.Sequence, Name: event.Type, Data: data})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating gameplay events: %w", err)
	}
	return result, nil
}

const terminalStateQuery = `SELECT last_event_sequence, canonical_state_json
FROM world_turn_states
WHERE game_id = $1 AND state_schema_version = 1`

// Terminal reads terminal state durably and projects it before SSE serialization.
func (s *Store) Terminal(ctx context.Context, viewer Viewer, projector Projector) (SSEEvent, bool, error) {
	if s == nil || s.db == nil {
		return SSEEvent{}, false, errors.New("gameplay event store requires a database")
	}
	var sequence int64
	var canonicalState []byte
	if err := s.db.QueryRowxContext(ctx, terminalStateQuery, viewer.GameID).Scan(&sequence, &canonicalState); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SSEEvent{}, false, nil
		}
		return SSEEvent{}, false, fmt.Errorf("reading gameplay terminal state: %w", err)
	}
	event, ended, err := terminalFromState(viewer, canonicalState, projector)
	if err != nil || !ended {
		return event, ended, err
	}
	event.ID = sequence
	return event, true, nil
}

func (s *Store) resync(ctx context.Context, viewer Viewer, projector Projector) ([]SSEEvent, error) {
	var revision, cursor int64
	var canonicalState []byte
	if err := s.db.QueryRowxContext(ctx, snapshotForResyncQuery, viewer.GameID).Scan(&revision, &cursor, &canonicalState); err != nil {
		return nil, fmt.Errorf("reading gameplay state for resync: %w", err)
	}
	projection, err := projector.ProjectSnapshot(viewer, canonicalState)
	if err != nil {
		return nil, fmt.Errorf("projecting gameplay state for resync: %w", err)
	}
	data, err := json.Marshal(struct {
		CurrentRevision int64           `json:"current_revision"`
		EventCursor     int64           `json:"event_cursor"`
		Snapshot        json.RawMessage `json:"snapshot"`
	}{CurrentRevision: revision, EventCursor: cursor, Snapshot: projection})
	if err != nil {
		return nil, fmt.Errorf("encoding gameplay resync: %w", err)
	}
	return []SSEEvent{{ID: cursor, Name: "resync", Data: data}}, nil
}

// ParseLastEventID accepts only an empty string or a base-10 nonnegative sequence.
func ParseLastEventID(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	cursor, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || cursor < 0 {
		return 0, errors.New("invalid Last-Event-ID")
	}
	return cursor, nil
}

// NewConnection subscribes to optional wakeups and guarantees idempotent cleanup.
func NewConnection(ctx context.Context, wakeup Wakeup, gameID int64) (context.Context, func()) {
	connection, cancel := context.WithCancel(ctx)
	var unsubscribe func()
	if wakeup != nil {
		_, unsubscribe = wakeup.Subscribe(gameID)
	}
	var once sync.Once
	return connection, func() {
		once.Do(func() {
			if unsubscribe != nil {
				unsubscribe()
			}
			cancel()
		})
	}
}

// TerminalEvent serializes terminal state only through the viewer projection seam.
func TerminalEvent(viewer Viewer, canonicalTerminal []byte, projector Projector) (SSEEvent, error) {
	if projector == nil {
		return SSEEvent{}, errors.New("gameplay event projector is required")
	}
	projection, err := projector.ProjectSnapshot(viewer, canonicalTerminal)
	if err != nil {
		return SSEEvent{}, fmt.Errorf("projecting terminal gameplay state: %w", err)
	}
	return SSEEvent{Name: "terminal", Data: projection}, nil
}
