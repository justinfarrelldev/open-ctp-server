package fullgame

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

type AuditArtifacts struct {
	TerminalWinner string
	TerminalReason string
	StateHash      string
	EventHash      string
}

const auditStateQuery = `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`
const auditEventsQuery = `SELECT event_id, payload_json::text FROM gameplay_events WHERE game_id = $1 ORDER BY sequence, ordinal`

func AuditTerminal(ctx context.Context, db *sqlx.DB, gameID int64) (AuditArtifacts, error) {
	if db == nil || gameID < 1 {
		return AuditArtifacts{}, errors.New("post-terminal audit is invalid")
	}
	var raw string
	if err := db.GetContext(ctx, &raw, auditStateQuery, gameID); err != nil {
		return AuditArtifacts{}, err
	}
	state, err := engine.DecodeCanonicalStateJSON([]byte(raw))
	if err != nil {
		return AuditArtifacts{}, err
	}
	return auditState(ctx, db, gameID, state)
}

func auditState(ctx context.Context, db *sqlx.DB, gameID int64, state engine.State) (AuditArtifacts, error) {
	if !state.Terminal.Ended {
		return AuditArtifacts{}, errors.New("post-terminal audit requires terminal state")
	}
	rows, err := db.QueryxContext(ctx, auditEventsQuery, gameID)
	if err != nil {
		return AuditArtifacts{}, err
	}
	defer rows.Close()
	events := []engine.Event{}
	for rows.Next() {
		var id, raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return AuditArtifacts{}, err
		}
		if id == "" {
			return AuditArtifacts{}, errors.New("legacy gameplay event has no canonical ID")
		}
		var event engine.Event
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			return AuditArtifacts{}, err
		}
		if event.ID != id {
			return AuditArtifacts{}, fmt.Errorf("gameplay event ID mismatch")
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return AuditArtifacts{}, err
	}
	stateHash, err := engine.CanonicalStateHash(state)
	if err != nil {
		return AuditArtifacts{}, err
	}
	eventHash, err := engine.CanonicalEventsHash(events)
	if err != nil {
		return AuditArtifacts{}, err
	}
	return AuditArtifacts{TerminalWinner: state.Terminal.WinningPlayerID, TerminalReason: state.Terminal.Reason, StateHash: stateHash, EventHash: eventHash}, nil
}
