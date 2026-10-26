package memorysql

import (
	"database/sql/driver"
	"encoding/json"
	"sort"
)

// Deliberately spell out the supported production SQL rather than accepting a
// broad prefix or importing the query constant: new/changed SQL stays fail-closed
// until this adapter implements its exact semantics. Discovery reveals only
// scheduling fields; returning whole worlds would mask privacy/performance bugs.
const aiSchedulingSQL = `SELECT state.game_id, state.revision,
 jsonb_build_object(
 'players', state.canonical_state_json->'players',
 'phase', state.canonical_state_json->'phase',
 'policy', jsonb_build_object(
 'stage', state.canonical_state_json #> '{policy,stage}',
 'early_ends', state.canonical_state_json #> '{policy,early_ends}'),
 'terminal', state.canonical_state_json->'terminal'
 )::text AS canonical_state_json
 FROM world_turn_states AS state JOIN games ON games.id = state.game_id
 WHERE state.state_schema_version = 1 AND games.status = 'started'`

type aiSchedulingJSON struct {
	Players json.RawMessage `json:"players"`
	Phase   json.RawMessage `json:"phase"`
	Policy  struct {
		Stage     json.RawMessage `json:"stage"`
		EarlyEnds json.RawMessage `json:"early_ends"`
	} `json:"policy"`
	Terminal json.RawMessage `json:"terminal"`
}

// query holds the store mutex before entering this read-only adapter.
func (c *conn) aiSchedulingRows(targeted bool, args []driver.NamedValue) (driver.Rows, error) {
	values := make([][]driver.Value, 0)
	for gameID, turn := range c.store.turns {
		game, ok := c.store.games[gameID]
		if !ok || game.status != "started" || targeted && gameID != intArg(args, 0) {
			continue
		}
		canonical, ok := turn[12].(string)
		if !ok {
			canonical = string(turn[12].([]byte))
		}
		var metadata aiSchedulingJSON
		if err := json.Unmarshal([]byte(canonical), &metadata); err != nil {
			return nil, err
		}
		raw, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		values = append(values, []driver.Value{int64(gameID), turn[10], string(raw)})
	}
	sort.Slice(values, func(i, j int) bool { return intArgFromValue(values[i][0]) < intArgFromValue(values[j][0]) })
	return &rows{columns: []string{"game_id", "revision", "canonical_state_json"}, values: values}, nil
}
