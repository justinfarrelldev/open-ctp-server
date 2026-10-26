package store

import (
	"database/sql"
	"errors"
	"fmt"
)

const gameplaySchedulingAggregateQuery = aiSchedulingSelect + ` AND state.game_id = $1`

// FindGameplaySchedulingAggregate reads only one committed game's scheduling
// fields. An ended/deleted game has no due work, including a notification racing
// its deletion. Errors remain retryable through periodic durable discovery.
func (s *Store) FindGameplaySchedulingAggregate(gameID int64) (StartedAggregate, error) {
	if s == nil || s.db == nil {
		return StartedAggregate{}, errors.New("gameplay store requires a database")
	}
	var result StartedAggregate
	if err := s.db.Get(&result, gameplaySchedulingAggregateQuery, gameID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StartedAggregate{}, nil
		}
		return StartedAggregate{}, fmt.Errorf("reading game AI scheduling state: %w", err)
	}
	return result, nil
}
