// Package memorysql is an in-process, fail-closed SQL driver for public E2E tests.
package memorysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

const DriverName = "e2e-memory"

var (
	storesMu sync.Mutex
	stores   = map[string]*store{}
)

func init() { sql.Register(DriverName, memoryDriver{}) }

type Telemetry struct {
	mu       sync.Mutex
	Connects int
	Queries  []string
	Unknown  error
}

type store struct {
	mu          sync.Mutex
	nextID      int64
	accounts    map[string]int64
	names       map[int64]string
	passwords   map[string]struct{}
	lobbies     map[int64]lobby
	members     map[int64][]int64
	games       map[int64]game
	players     map[int64][]int64
	countries   map[int64]map[int64]string
	controllers map[int64]map[int64]string
	snapshots   map[int64][]driver.Value
	turns       map[int64][]driver.Value
	receipts    map[string][]driver.Value
	exploration map[string][]driver.Value
	events      map[string][]driver.Value
	outbox      map[string]outboxRecord
	sessions    map[string]session
	telemetry   *Telemetry
}

func Open(name string) (*sql.DB, *Telemetry, error) {
	storesMu.Lock()
	if _, ok := stores[name]; !ok {
		stores[name] = &store{nextID: 1, accounts: map[string]int64{}, names: map[int64]string{}, passwords: map[string]struct{}{}, lobbies: map[int64]lobby{}, members: map[int64][]int64{}, games: map[int64]game{}, players: map[int64][]int64{}, countries: map[int64]map[int64]string{}, controllers: map[int64]map[int64]string{}, snapshots: map[int64][]driver.Value{}, turns: map[int64][]driver.Value{}, receipts: map[string][]driver.Value{}, exploration: map[string][]driver.Value{}, events: map[string][]driver.Value{}, outbox: map[string]outboxRecord{}, sessions: map[string]session{}, telemetry: &Telemetry{}}
	}
	telemetry := stores[name].telemetry
	storesMu.Unlock()
	db, err := sql.Open(DriverName, name)
	return db, telemetry, err
}

// Reset discards a closed named test store so the next Open starts from an empty deterministic state.
func Reset(name string) {
	storesMu.Lock()
	defer storesMu.Unlock()
	delete(stores, name)
}

type memoryDriver struct{}

func (memoryDriver) Open(name string) (driver.Conn, error) {
	storesMu.Lock()
	defer storesMu.Unlock()
	store, ok := stores[name]
	if !ok {
		return nil, fmt.Errorf("memorysql: unknown named store %q", name)
	}
	store.telemetry.mu.Lock()
	store.telemetry.Connects++
	store.telemetry.mu.Unlock()
	return &conn{store: store}, nil
}

type session struct {
	id        int64
	accountID int64
	createdAt time.Time
	expiresAt time.Time
}

type outboxRecord struct {
	gameID, sequence int64
	delivered        bool
	attempts         int64
}

type game struct {
	id, ownerID, lobbyID                           int64
	name, style, rulesJSON, exclusionsJSON, status string
	startingSettlers, startingGold                 int64
	mapSize, climate                               string
	wetness                                        *int64
	startedAt                                      *time.Time
}

type lobby struct {
	id, ownerID     int64
	name, ownerName string
	closed          bool
	muted, public   bool
	createdAt       time.Time
}

type conn struct {
	store *store
	tx    *tx
}

func (c *conn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("memorysql: prepared statements are unsupported")
}
func (c *conn) Close() error                                                 { return nil }
func (c *conn) Begin() (driver.Tx, error)                                    { return c.begin() }
func (c *conn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) { return c.begin() }
func (c *conn) begin() (driver.Tx, error) {
	if c.tx != nil {
		return nil, errors.New("memorysql: nested transactions are unsupported")
	}
	transaction := &tx{conn: c}
	c.tx = transaction
	return transaction, nil
}
func (c *conn) Ping(context.Context) error { return nil }

func (c *conn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.query(query, args)
}
func (c *conn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.exec(query, args)
}

func (c *conn) query(query string, args []driver.NamedValue) (driver.Rows, error) {
	c.record(query)
	normalized := normalize(query)
	c.store.mu.Lock()
	defer c.store.mu.Unlock()
	switch normalized {
	case "select state.game_id, state.revision, player.value->>'id' as player_id from world_turn_states as state join games on games.id = state.game_id cross join lateral jsonb_array_elements(state.canonical_state_json->'players') as player(value) where state.state_schema_version = 1 and games.status = 'started' and coalesce(player.value->>'controller', '') = 'human' and coalesce(player.value->>'takeover_pending', 'false') = 'true' order by state.game_id asc, player_id asc":
		values := make([][]driver.Value, 0)
		for gameID, turn := range c.store.turns {
			game, ok := c.store.games[gameID]
			if !ok || game.status != "started" {
				continue
			}
			var state struct {
				Players []struct {
					ID              string `json:"id"`
					Controller      string `json:"controller"`
					TakeoverPending bool   `json:"takeover_pending"`
				} `json:"players"`
			}
			canonical, ok := turn[12].(string)
			if !ok {
				canonical = string(turn[12].([]byte))
			}
			if json.Unmarshal([]byte(canonical), &state) != nil {
				continue
			}
			for _, player := range state.Players {
				if player.Controller != "human" || !player.TakeoverPending {
					continue
				}
				values = append(values, []driver.Value{int64(gameID), turn[10], player.ID})
			}
		}
		sort.Slice(values, func(i, j int) bool {
			leftGame, rightGame := intArgFromValue(values[i][0]), intArgFromValue(values[j][0])
			if leftGame != rightGame {
				return leftGame < rightGame
			}
			return values[i][2].(string) < values[j][2].(string)
		})
		return &rows{columns: []string{"game_id", "revision", "player_id"}, values: values}, nil
	case "select game_id, revision, canonical_state_json #>> '{policy,deadline,player_id}' as player_id, (canonical_state_json #>> '{policy,deadline,at_unix}')::bigint as deadline_at from world_turn_states where state_schema_version = 1 and canonical_state_json #>> '{phase,mode}' = 'long_turn' and (canonical_state_json #>> '{policy,deadline,at_unix}')::bigint <= $1 order by deadline_at asc, game_id asc":
		now := intArg(args, 0)
		values := make([][]driver.Value, 0)
		for gameID, turn := range c.store.turns {
			var state struct {
				Phase struct {
					Mode string `json:"mode"`
				} `json:"phase"`
				Policy struct {
					Deadline *struct {
						PlayerID string `json:"player_id"`
						AtUnix   int64  `json:"at_unix"`
					} `json:"deadline"`
				} `json:"policy"`
			}
			canonical, ok := turn[12].(string)
			if !ok {
				canonical = string(turn[12].([]byte))
			}
			if json.Unmarshal([]byte(canonical), &state) != nil || state.Phase.Mode != "long_turn" || state.Policy.Deadline == nil || state.Policy.Deadline.AtUnix > int64(now) {
				continue
			}
			values = append(values, []driver.Value{int64(gameID), turn[10], state.Policy.Deadline.PlayerID, state.Policy.Deadline.AtUnix})
		}
		sort.Slice(values, func(i, j int) bool {
			return intArgFromValue(values[i][3]) < intArgFromValue(values[j][3]) || intArgFromValue(values[i][3]) == intArgFromValue(values[j][3]) && intArgFromValue(values[i][0]) < intArgFromValue(values[j][0])
		})
		return &rows{columns: []string{"game_id", "revision", "player_id", "deadline_at"}, values: values}, nil
	case "select state.game_id, state.revision, state.canonical_state_json::text as canonical_state_json from world_turn_states as state join games on games.id = state.game_id where state.state_schema_version = 1 and games.status = 'started' order by state.game_id asc":
		values := make([][]driver.Value, 0)
		for gameID, turn := range c.store.turns {
			game, ok := c.store.games[gameID]
			if !ok || game.status != "started" {
				continue
			}
			values = append(values, []driver.Value{int64(gameID), turn[10], turn[12]})
		}
		sort.Slice(values, func(i, j int) bool { return intArgFromValue(values[i][0]) < intArgFromValue(values[j][0]) })
		return &rows{columns: []string{"game_id", "revision", "canonical_state_json"}, values: values}, nil
	case "select state.game_id, state.state_schema_version, state.revision, state.last_event_sequence, state.canonical_state_json::text as canonical_state_json, snapshot.ruleset_canonical_json::text as ruleset_canonical_json, snapshot.ruleset_content_hash, snapshot.ruleset_version from world_turn_states as state join world_snapshots as snapshot on snapshot.game_id = state.game_id where state.game_id = $1", "select state.game_id, state.state_schema_version, state.revision, state.last_event_sequence, state.canonical_state_json::text as canonical_state_json, snapshot.ruleset_canonical_json::text as ruleset_canonical_json, snapshot.ruleset_content_hash, snapshot.ruleset_version from world_turn_states as state join world_snapshots as snapshot on snapshot.game_id = state.game_id where state.game_id = $1 for update of state":
		gameID := intArg(args, 0)
		turn, hasTurn := c.store.turns[gameID]
		snapshot, hasSnapshot := c.store.snapshots[gameID]
		columns := []string{"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version"}
		if !hasTurn || !hasSnapshot {
			return &rows{columns: columns}, nil
		}
		return &rows{columns: columns, values: [][]driver.Value{{turn[0], turn[9], turn[10], turn[11], turn[12], snapshot[15], snapshot[16], snapshot[17]}}}, nil
	case "select game_id, actor_player_id, idempotency_key, payload_sha256, result_json::text as result_json, resulting_revision, resulting_event_sequence from gameplay_command_receipts where game_id = $1 and actor_player_id = $2 and idempotency_key = $3":
		key := fmt.Sprintf("%d:%d:%s", intArg(args, 0), intArg(args, 1), stringArg(args, 2))
		columns := []string{"game_id", "actor_player_id", "idempotency_key", "payload_sha256", "result_json", "resulting_revision", "resulting_event_sequence"}
		if receipt, ok := c.store.receipts[key]; ok {
			return &rows{columns: columns, values: [][]driver.Value{{receipt[0], receipt[1], receipt[2], receipt[3], receipt[4], receipt[5], receipt[6]}}}, nil
		}
		return &rows{columns: columns}, nil
	case "select game_id, actor_player_id, sequence, command_type, destination_tile_id from gameplay_ai_exploration_history where game_id = $1 and actor_player_id = $2 order by sequence asc":
		gameID, actorPlayerID := intArg(args, 0), intArg(args, 1)
		values := make([][]driver.Value, 0)
		for _, step := range c.store.exploration {
			if intArgFromValue(step[0]) == gameID && intArgFromValue(step[1]) == actorPlayerID {
				values = append(values, []driver.Value{step[0], step[1], step[2], step[3], step[4]})
			}
		}
		sort.Slice(values, func(i, j int) bool { return intArgFromValue(values[i][2]) < intArgFromValue(values[j][2]) })
		return &rows{columns: []string{"game_id", "actor_player_id", "sequence", "command_type", "destination_tile_id"}, values: values}, nil
	case "select min(sequence) from gameplay_events where game_id = $1":
		gameID := intArg(args, 0)
		var min driver.Value
		for _, event := range c.store.events {
			if intArgFromValue(event[0]) != gameID {
				continue
			}
			if min == nil || intArgFromValue(event[1]) < intArgFromValue(min) {
				min = event[1]
			}
		}
		return &rows{columns: []string{"min"}, values: [][]driver.Value{{min}}}, nil
	case "select game_id, sequence, revision, ordinal, event_type, payload_json from gameplay_events where game_id = $1 and sequence > $2 order by sequence asc":
		gameID, cursor := intArg(args, 0), intArg(args, 1)
		values := gameplayEventRows(c.store.events, gameID, cursor, false)
		return &rows{columns: []string{"game_id", "sequence", "revision", "ordinal", "event_type", "payload_json"}, values: values}, nil
	case "select event_id, payload_json::text from gameplay_events where game_id = $1 order by sequence, ordinal":
		gameID := intArg(args, 0)
		values := gameplayEventRows(c.store.events, gameID, -1, true)
		return &rows{columns: []string{"event_id", "payload_json"}, values: values}, nil
	case "select revision, last_event_sequence, canonical_state_json from world_turn_states where game_id = $1 and state_schema_version = 1":
		turn, ok := c.store.turns[intArg(args, 0)]
		if !ok || intArgFromValue(turn[9]) != 1 {
			return &rows{columns: []string{"revision", "last_event_sequence", "canonical_state_json"}}, nil
		}
		return &rows{columns: []string{"revision", "last_event_sequence", "canonical_state_json"}, values: [][]driver.Value{{turn[10], turn[11], turn[12]}}}, nil
	case "select last_event_sequence, canonical_state_json from world_turn_states where game_id = $1 and state_schema_version = 1":
		turn, ok := c.store.turns[intArg(args, 0)]
		if !ok || intArgFromValue(turn[9]) != 1 {
			return &rows{columns: []string{"last_event_sequence", "canonical_state_json"}}, nil
		}
		return &rows{columns: []string{"last_event_sequence", "canonical_state_json"}, values: [][]driver.Value{{turn[11], turn[12]}}}, nil
	case "select canonical_state_json::text from world_turn_states where game_id = $1 and state_schema_version = 1":
		turn, ok := c.store.turns[intArg(args, 0)]
		if !ok || intArgFromValue(turn[9]) != 1 {
			return &rows{columns: []string{"canonical_state_json"}}, nil
		}
		return &rows{columns: []string{"canonical_state_json"}, values: [][]driver.Value{{turn[12]}}}, nil
	case "select o.game_id, o.sequence from gameplay_event_outbox o where o.delivered_at is null order by o.game_id asc, o.sequence asc", "select o.game_id, o.sequence from gameplay_event_outbox o where o.delivered_at is null order by o.game_id asc, o.sequence asc for update skip locked":
		keys := make([]string, 0, len(c.store.outbox))
		for key, record := range c.store.outbox {
			if !record.delivered {
				keys = append(keys, key)
			}
		}
		sort.Slice(keys, func(i, j int) bool {
			left, right := c.store.outbox[keys[i]], c.store.outbox[keys[j]]
			return left.gameID < right.gameID || left.gameID == right.gameID && left.sequence < right.sequence
		})
		values := make([][]driver.Value, 0, len(keys))
		for _, key := range keys {
			record := c.store.outbox[key]
			values = append(values, []driver.Value{record.gameID, record.sequence})
		}
		return &rows{columns: []string{"game_id", "sequence"}, values: values}, nil
	case "select game_id from world_snapshots where game_id = $1":
		gameID := intArg(args, 0)
		if _, ok := c.store.snapshots[gameID]; ok {
			return &rows{columns: []string{"game_id"}, values: [][]driver.Value{{gameID}}}, nil
		}
		return &rows{columns: []string{"game_id"}}, nil
	case "select game_id, created_at, updated_at, game_status, generator_version, style, rules_json::text as rules_json, exclusions_json::text as exclusions_json, starting_settlers, starting_gold, seed, wetness, map_size, climate, player_order_json::text as player_order_json, spawns_json::text as spawns_json, tiles_json::text as tiles_json from world_snapshots where game_id = $1":
		stored, ok := c.store.snapshots[intArg(args, 0)]
		columns := []string{"game_id", "created_at", "updated_at", "game_status", "generator_version", "style", "rules_json", "exclusions_json", "starting_settlers", "starting_gold", "seed", "wetness", "map_size", "climate", "player_order_json", "spawns_json", "tiles_json"}
		if !ok {
			return &rows{columns: columns}, nil
		}
		now := time.Now().UTC()
		return &rows{columns: columns, values: [][]driver.Value{{stored[0], now, now, stored[1], stored[2], stored[3], stored[4], stored[5], stored[6], stored[7], stored[8], stored[9], stored[10], stored[11], stored[12], stored[13], stored[14]}}}, nil
	case "select game_id, created_at, updated_at, turn_number, current_player_index, active_player_order_json::text as active_player_order_json, recalculation_count, rng_state, ratings_json::text as ratings_json, pollution_json::text as pollution_json, last_recalculated_at from world_turn_states where game_id = $1":
		stored, ok := c.store.turns[intArg(args, 0)]
		columns := []string{"game_id", "created_at", "updated_at", "turn_number", "current_player_index", "active_player_order_json", "recalculation_count", "rng_state", "ratings_json", "pollution_json", "last_recalculated_at"}
		if !ok {
			return &rows{columns: columns}, nil
		}
		now := time.Now().UTC()
		return &rows{columns: columns, values: [][]driver.Value{{stored[0], now, now, stored[1], stored[2], stored[3], stored[4], stored[5], stored[6], stored[7], stored[8]}}}, nil
	case "select id, status, owner_account_id from games where id = $1":
		game, ok := c.store.games[intArg(args, 0)]
		if !ok {
			return &rows{columns: []string{"id", "status", "owner_account_id"}}, nil
		}
		return &rows{columns: []string{"id", "status", "owner_account_id"}, values: [][]driver.Value{{game.id, game.status, game.ownerID}}}, nil
	case "select id, controller_kind from game_players where game_id = $1 and account_id = $2", "select id, account_id from game_players where game_id = $1 and account_id = $2":
		gameID, accountID := intArg(args, 0), intArg(args, 1)
		for index, playerID := range c.store.players[gameID] {
			if playerID != accountID {
				continue
			}
			if strings.HasPrefix(normalized, "select id, controller_kind") {
				controller := c.store.controllers[gameID][accountID]
				if controller == "" {
					controller = "human"
				}
				return &rows{columns: []string{"id", "controller_kind"}, values: [][]driver.Value{{int64(index + 1), controller}}}, nil
			}
			return &rows{columns: []string{"id", "account_id"}, values: [][]driver.Value{{int64(index + 1), accountID}}}, nil
		}
		if strings.HasPrefix(normalized, "select id, controller_kind") {
			return &rows{columns: []string{"id", "controller_kind"}}, nil
		}
		return &rows{columns: []string{"id", "account_id"}}, nil
	case "select id, controller_kind from game_players where game_id = $1 and id = $2":
		players := c.store.players[intArg(args, 0)]
		index := intArg(args, 1) - 1
		if index < 0 || index >= int64(len(players)) {
			return &rows{columns: []string{"id", "controller_kind"}}, nil
		}
		accountID := players[index]
		controller := c.store.controllers[intArg(args, 0)][accountID]
		if controller == "" {
			controller = "human"
		}
		return &rows{columns: []string{"id", "controller_kind"}, values: [][]driver.Value{{intArg(args, 1), controller}}}, nil
	case "select id from game_players where game_id = $1 and account_id = $2":
		for index, accountID := range c.store.players[intArg(args, 0)] {
			if accountID == intArg(args, 1) {
				return &rows{columns: []string{"id"}, values: [][]driver.Value{{int64(index + 1)}}}, nil
			}
		}
		return &rows{columns: []string{"id"}}, nil
	case "select * from account where email = $1":
		email := stringArg(args, 0)
		if id, ok := c.store.accounts[email]; ok {
			return &rows{columns: []string{"id"}, values: [][]driver.Value{{id}}}, nil
		}
		return &rows{columns: []string{"id"}}, nil
	case "insert into account (name, info, location, email, experience_level) values ($1, 'server-managed ai player', 'local', $2, 0) returning id":
		email := stringArg(args, 1)
		if _, exists := c.store.accounts[email]; exists {
			return nil, c.unknown(query, "duplicate account")
		}
		id := c.store.nextID
		c.store.nextID++
		c.store.accounts[email] = id
		c.store.names[id] = stringArg(args, 0)
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { delete(c.store.accounts, email); delete(c.store.names, id) })
		}
		return &rows{columns: []string{"id"}, values: [][]driver.Value{{id}}}, nil
	case "insert into account (name, info, location, email, experience_level) values ($1, $2, $3, $4, $5) returning id":
		email := stringArg(args, 3)
		if _, exists := c.store.accounts[email]; exists {
			return nil, c.unknown(query, "duplicate account")
		}
		id := c.store.nextID
		c.store.nextID++
		c.store.accounts[email] = id
		c.store.names[id] = stringArg(args, 0)
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { delete(c.store.accounts, email); delete(c.store.names, id) })
		}
		return &rows{columns: []string{"id"}, values: [][]driver.Value{{id}}}, nil
	case "insert into passwords (account_email, hash, salt) values ($1, $2, $3)":
		c.store.passwords[stringArg(args, 0)] = struct{}{}
		return &rows{columns: []string{"account_email"}, values: [][]driver.Value{{stringArg(args, 0)}}}, nil
	case "select lm.lobby_id from lobby_members lm join lobbies l on l.id = lm.lobby_id where lm.account_id = $1 and l.closed_at is null order by lm.joined_at asc, lm.id asc":
		accountID := intArg(args, 0)
		for lobbyID, members := range c.store.members {
			for _, member := range members {
				if member == accountID && !c.store.lobbies[lobbyID].closed {
					return &rows{columns: []string{"lobby_id"}, values: [][]driver.Value{{lobbyID}}}, nil
				}
			}
		}
		return &rows{columns: []string{"lobby_id"}}, nil
	case "select name from account where id = $1":
		if name, ok := c.store.names[intArg(args, 0)]; ok {
			return &rows{columns: []string{"name"}, values: [][]driver.Value{{name}}}, nil
		}
		return &rows{columns: []string{"name"}}, nil
	case "insert into lobbies (name, owner_name, owner_account_id, is_closed, is_muted, is_public, updated_at, last_activity_at) values ($1, $2, $3, false, $4, $5, now(), now()) returning id":
		id := c.store.nextID
		c.store.nextID++
		c.store.lobbies[id] = lobby{id: id, name: stringArg(args, 0), ownerName: stringArg(args, 1), ownerID: intArg(args, 2), public: true, createdAt: time.Now().UTC()}
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { delete(c.store.lobbies, id) })
		}
		return &rows{columns: []string{"id"}, values: [][]driver.Value{{id}}}, nil
	case "select id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms, controller_kind from game_players where game_id = $1 order by joined_at asc, id asc":
		gameID := intArg(args, 0)
		now := time.Now().UTC()
		players := c.store.players[gameID]
		values := make([][]driver.Value, 0, len(players))
		for index, accountID := range players {
			controller := c.store.controllers[gameID][accountID]
			if controller == "" {
				controller = "human"
			}
			values = append(values, []driver.Value{int64(index + 1), now, now, gameID, accountID, now, c.store.countries[gameID][accountID], 0, controller})
		}
		return &rows{columns: []string{"id", "created_at", "updated_at", "game_id", "account_id", "joined_at", "country", "ping_ms", "controller_kind"}, values: values}, nil
	case "select account_id, controller_kind from game_players where game_id = $1":
		gameID := intArg(args, 0)
		players := c.store.players[gameID]
		values := make([][]driver.Value, 0, len(players))
		for _, accountID := range players {
			controller := c.store.controllers[gameID][accountID]
			if controller == "" {
				controller = "human"
			}
			values = append(values, []driver.Value{accountID, controller})
		}
		return &rows{columns: []string{"account_id", "controller_kind"}, values: values}, nil
	case "select id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms, controller_kind from game_players where game_id = $1 and account_id = $2":
		gameID, accountID := intArg(args, 0), intArg(args, 1)
		for index, playerID := range c.store.players[gameID] {
			if playerID == accountID {
				now := time.Now().UTC()
				controller := c.store.controllers[gameID][accountID]
				if controller == "" {
					controller = "human"
				}
				return &rows{columns: []string{"id", "created_at", "updated_at", "game_id", "account_id", "joined_at", "country", "ping_ms", "controller_kind"}, values: [][]driver.Value{{int64(index + 1), now, now, gameID, accountID, now, c.store.countries[gameID][accountID], 0, controller}}}, nil
			}
		}
		return &rows{columns: []string{"id", "created_at", "updated_at", "game_id", "account_id", "joined_at", "country", "ping_ms", "controller_kind"}}, nil
	case "select id from game_players where game_id = $1 and account_id <> $2 and lower(country) = lower($3)":
		gameID, accountID, country := intArg(args, 0), intArg(args, 1), strings.ToLower(stringArg(args, 2))
		for index, playerID := range c.store.players[gameID] {
			if playerID != accountID && strings.ToLower(c.store.countries[gameID][playerID]) == country {
				return &rows{columns: []string{"id"}, values: [][]driver.Value{{int64(index + 1)}}}, nil
			}
		}
		return &rows{columns: []string{"id"}}, nil
	case "select id, created_at, updated_at, started_at, lobby_id, owner_account_id, name, password_protected, password, is_locked, is_public, style, rules_json, exclusions_json, status, starting_settlers, starting_gold, map_size, climate, wetness from games where id = $1", "select id, created_at, updated_at, started_at, lobby_id, owner_account_id, name, password_protected, password, is_locked, is_public, style, rules_json, exclusions_json, status, starting_settlers, starting_gold, map_size, climate, wetness from games where id = $1 for update":
		columns := []string{"id", "created_at", "updated_at", "started_at", "lobby_id", "owner_account_id", "name", "password_protected", "password", "is_locked", "is_public", "style", "rules_json", "exclusions_json", "status", "starting_settlers", "starting_gold", "map_size", "climate", "wetness"}
		if game, ok := c.store.games[intArg(args, 0)]; ok {
			now := time.Now().UTC()
			var startedAt, lobbyID, wetness driver.Value
			if game.startedAt != nil {
				startedAt = *game.startedAt
			}
			if game.lobbyID > 0 {
				lobbyID = game.lobbyID
			}
			if game.wetness != nil {
				wetness = *game.wetness
			}
			return &rows{columns: columns, values: [][]driver.Value{{game.id, now, now, startedAt, lobbyID, game.ownerID, game.name, false, "", false, true, game.style, game.rulesJSON, game.exclusionsJSON, game.status, game.startingSettlers, game.startingGold, game.mapSize, game.climate, wetness}}}, nil
		}
		return &rows{columns: columns}, nil
	case "insert into games ( name, owner_account_id, lobby_id, password_protected, password, is_locked, is_public, style, rules_json, exclusions_json, status, starting_settlers, starting_gold ) values ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13 ) returning id":
		id := c.store.nextID
		c.store.nextID++
		c.store.games[id] = game{id: id, name: stringArg(args, 0), ownerID: intArg(args, 1), lobbyID: intArg(args, 2), style: stringArg(args, 7), rulesJSON: stringArg(args, 8), exclusionsJSON: stringArg(args, 9), status: stringArg(args, 10), startingSettlers: intArg(args, 11), startingGold: intArg(args, 12)}
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { delete(c.store.games, id) })
		}
		return &rows{columns: []string{"id"}, values: [][]driver.Value{{id}}}, nil
	case "select account_id from lobby_members where lobby_id = $1 order by joined_at asc, id asc":
		members := c.store.members[intArg(args, 0)]
		values := make([][]driver.Value, 0, len(members))
		for _, member := range members {
			values = append(values, []driver.Value{member})
		}
		return &rows{columns: []string{"account_id"}, values: values}, nil
	case "select id, closed_at from lobbies where id = $1":
		if lobby, ok := c.store.lobbies[intArg(args, 0)]; ok {
			var closedAt driver.Value
			if lobby.closed {
				closedAt = lobby.createdAt
			}
			return &rows{columns: []string{"id", "closed_at"}, values: [][]driver.Value{{lobby.id, closedAt}}}, nil
		}
		return &rows{columns: []string{"id", "closed_at"}}, nil
	case "select id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at from lobby where id = $1":
		if lobby, ok := c.store.lobbies[intArg(args, 0)]; ok {
			return &rows{columns: []string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}, values: [][]driver.Value{{lobby.id, lobby.name, lobby.ownerName, lobby.ownerID, lobby.closed, lobby.muted, lobby.public, lobby.createdAt, lobby.createdAt, lobby.createdAt, nil}}}, nil
		}
		return &rows{columns: []string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}}, nil
	case "select * from sessions where id = $1":
		if session, ok := c.store.sessions[stringArg(args, 0)]; ok {
			return &rows{columns: []string{"id", "account_id", "created_at", "expires_at"}, values: [][]driver.Value{{session.id, session.accountID, session.createdAt, session.expiresAt}}}, nil
		}
		return &rows{columns: []string{"id", "account_id", "created_at", "expires_at"}}, nil
	default:
		return nil, c.unknown(query, "unsupported query")
	}
}

func (c *conn) insertGamePlayer(gameID, accountID int64, country, controller string) driver.Result {
	c.store.mu.Lock()
	defer c.store.mu.Unlock()
	c.store.players[gameID] = append(c.store.players[gameID], accountID)
	if c.store.countries[gameID] == nil {
		c.store.countries[gameID] = map[int64]string{}
	}
	if c.store.controllers[gameID] == nil {
		c.store.controllers[gameID] = map[int64]string{}
	}
	c.store.countries[gameID][accountID] = country
	c.store.controllers[gameID][accountID] = controller
	if c.tx != nil {
		c.tx.undo = append(c.tx.undo, func() {
			players := c.store.players[gameID]
			c.store.players[gameID] = players[:len(players)-1]
			delete(c.store.countries[gameID], accountID)
			delete(c.store.controllers[gameID], accountID)
		})
	}
	return driver.RowsAffected(1)
}

func (c *conn) exec(query string, args []driver.NamedValue) (driver.Result, error) {
	c.record(query)
	normalized := normalize(query)
	if normalized == "update game_players set controller_kind = 'ai', updated_at = now() where id = $1" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		playerID := intArg(args, 0)
		for gameID, players := range c.store.players {
			index := playerID - 1
			if index < 0 || index >= int64(len(players)) {
				continue
			}
			if c.store.controllers[gameID] == nil {
				c.store.controllers[gameID] = map[int64]string{}
			}
			accountID := players[index]
			previous := c.store.controllers[gameID][accountID]
			c.store.controllers[gameID][accountID] = "ai"
			if c.tx != nil {
				c.tx.undo = append(c.tx.undo, func() { c.store.controllers[gameID][accountID] = previous })
			}
			return driver.RowsAffected(1), nil
		}
		return driver.RowsAffected(0), nil
	}
	if normalized == "update game_players set controller_kind = $1, controller_changed_at = now() where game_id = $2 and id = $3" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID, playerID := intArg(args, 1), intArg(args, 2)
		players := c.store.players[gameID]
		index := playerID - 1
		if index < 0 || index >= int64(len(players)) {
			return driver.RowsAffected(0), nil
		}
		if c.store.controllers[gameID] == nil {
			c.store.controllers[gameID] = map[int64]string{}
		}
		accountID := players[index]
		previous := c.store.controllers[gameID][accountID]
		c.store.controllers[gameID][accountID] = stringArg(args, 0)
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { c.store.controllers[gameID][accountID] = previous })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update games set status = 'ended', ended_at = now(), ended_by = $1, end_reason = $2, updated_at = now() where id = $3" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 2)
		stored, ok := c.store.games[gameID]
		if !ok {
			return driver.RowsAffected(0), nil
		}
		previous := stored.status
		stored.status = "ended"
		c.store.games[gameID] = stored
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { game := c.store.games[gameID]; game.status = previous; c.store.games[gameID] = game })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update games set status = $1, updated_at = now() where id = $2" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 1)
		stored, ok := c.store.games[gameID]
		if !ok {
			return driver.RowsAffected(0), nil
		}
		previous := stored.status
		stored.status = stringArg(args, 0)
		c.store.games[gameID] = stored
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { game := c.store.games[gameID]; game.status = previous; c.store.games[gameID] = game })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update world_turn_states set canonical_state_json = $1::jsonb, revision = revision + 1, last_event_sequence = $2, updated_at = now() where game_id = $3 and state_schema_version = 1 and revision = $4" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 2)
		turn, ok := c.store.turns[gameID]
		if !ok || intArgFromValue(turn[10]) != intArg(args, 3) {
			return driver.RowsAffected(0), nil
		}
		previous := append([]driver.Value(nil), turn...)
		turn[12], turn[11], turn[10] = args[0].Value, args[1].Value, intArg(args, 3)+1
		c.store.turns[gameID] = turn
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { c.store.turns[gameID] = previous })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update games set status = $1, started_at = now(), updated_at = now() where id = $2" {
		gameID := intArg(args, 1)
		game, ok := c.store.games[gameID]
		if !ok {
			return driver.RowsAffected(0), nil
		}
		previous := game.status
		game.status = stringArg(args, 0)
		c.store.games[gameID] = game
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { game := c.store.games[gameID]; game.status = previous; c.store.games[gameID] = game })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "insert into game_players (game_id, account_id) values ($1, $2)" {
		gameID, accountID := intArg(args, 0), intArg(args, 1)
		return c.insertGamePlayer(gameID, accountID, "", "human"), nil
	}
	if normalized == "insert into game_players (game_id, account_id, country, controller_kind) values ($1, $2, $3, 'ai')" {
		gameID, accountID := intArg(args, 0), intArg(args, 1)
		return c.insertGamePlayer(gameID, accountID, stringArg(args, 2), "ai"), nil
	}
	if normalized == "update lobbies set last_activity_at = now(), updated_at = now() where id = $1" {
		return driver.RowsAffected(1), nil
	}
	if normalized == "update games set rules_json = $1, updated_at = now() where id = $2" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 1)
		stored, ok := c.store.games[gameID]
		if !ok {
			return driver.RowsAffected(0), nil
		}
		previous := stored.rulesJSON
		stored.rulesJSON = stringArg(args, 0)
		c.store.games[gameID] = stored
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { game := c.store.games[gameID]; game.rulesJSON = previous; c.store.games[gameID] = game })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update games set starting_settlers = $1, updated_at = now() where id = $2" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 1)
		stored, ok := c.store.games[gameID]
		if !ok {
			return driver.RowsAffected(0), nil
		}
		previous := stored.startingSettlers
		stored.startingSettlers = intArg(args, 0)
		c.store.games[gameID] = stored
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { game := c.store.games[gameID]; game.startingSettlers = previous; c.store.games[gameID] = game })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update games set style = $1, map_size = $2, climate = $3, wetness = $4, updated_at = now() where id = $5" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 4)
		stored, ok := c.store.games[gameID]
		if !ok {
			return driver.RowsAffected(0), nil
		}
		wetness := intArg(args, 3)
		stored.style, stored.mapSize, stored.climate, stored.wetness = stringArg(args, 0), stringArg(args, 1), stringArg(args, 2), &wetness
		c.store.games[gameID] = stored
		return driver.RowsAffected(1), nil
	}
	if normalized == "update games set status = $1, started_at = now(), updated_at = now() where id = $2" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 1)
		stored, ok := c.store.games[gameID]
		if !ok {
			return driver.RowsAffected(0), nil
		}
		now := time.Now().UTC()
		stored.status, stored.startedAt = stringArg(args, 0), &now
		c.store.games[gameID] = stored
		return driver.RowsAffected(1), nil
	}
	if normalized == "update game_players set country = $1, updated_at = now() where game_id = $2 and account_id = $3" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID, accountID := intArg(args, 1), intArg(args, 2)
		if c.store.countries[gameID] == nil {
			c.store.countries[gameID] = map[int64]string{}
		}
		c.store.countries[gameID][accountID] = stringArg(args, 0)
		return driver.RowsAffected(1), nil
	}
	if normalized == "insert into lobby_members (lobby_id, account_id) values ($1, $2)" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		lobbyID, accountID := intArg(args, 0), intArg(args, 1)
		c.store.members[lobbyID] = append(c.store.members[lobbyID], accountID)
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { members := c.store.members[lobbyID]; c.store.members[lobbyID] = members[:len(members)-1] })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "insert into sessions (id, account_id, created_at, expires_at) values (?, ?, ?, ?)" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		c.store.sessions[stringArg(args, 0)] = session{id: intArg(args, 0), accountID: intArg(args, 1), createdAt: timeArg(args, 2), expiresAt: timeArg(args, 3)}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update world_turn_states set canonical_state_json = $1::jsonb, revision = revision + 1, last_event_sequence = $2, updated_at = now() where game_id = $3 and state_schema_version = 1 and revision = $4" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 2)
		stored, ok := c.store.turns[gameID]
		if !ok || intArg(args, 3) != intArgFromValue(stored[10]) {
			return driver.RowsAffected(0), nil
		}
		stored[10] = intArgFromValue(stored[10]) + 1
		stored[11], stored[12] = args[1].Value, args[0].Value
		c.store.turns[gameID] = stored
		return driver.RowsAffected(1), nil
	}
	if strings.HasPrefix(normalized, "insert into world_snapshots (") {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 0)
		if _, exists := c.store.snapshots[gameID]; exists {
			return nil, c.unknown(query, "duplicate world snapshot")
		}
		values := make([]driver.Value, len(args))
		for index := range args {
			values[index] = args[index].Value
		}
		c.store.snapshots[gameID] = values
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { delete(c.store.snapshots, gameID) })
		}
		return driver.RowsAffected(1), nil
	}
	if strings.HasPrefix(normalized, "insert into world_turn_states (") {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID := intArg(args, 0)
		if _, exists := c.store.turns[gameID]; exists {
			return nil, c.unknown(query, "duplicate world turn state")
		}
		values := make([]driver.Value, len(args))
		for index := range args {
			values[index] = args[index].Value
		}
		c.store.turns[gameID] = values
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { delete(c.store.turns, gameID) })
		}
		return driver.RowsAffected(1), nil
	}
	if strings.HasPrefix(normalized, "update world_turn_states set canonical_state_json = $1::jsonb, revision = revision + 1, last_event_sequence = $2, updated_at = now() where game_id = $3 and state_schema_version = 1 and revision = $4") {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID, expectedRevision := intArg(args, 2), intArg(args, 3)
		stored, ok := c.store.turns[gameID]
		if !ok || int64(stored[10].(int64)) != expectedRevision {
			return driver.RowsAffected(0), nil
		}
		previous := append([]driver.Value(nil), stored...)
		stored[10], stored[11], stored[12] = expectedRevision+1, intArg(args, 1), stringArg(args, 0)
		c.store.turns[gameID] = stored
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { c.store.turns[gameID] = previous })
		}
		return driver.RowsAffected(1), nil
	}
	if strings.HasPrefix(normalized, "insert into gameplay_command_receipts (") {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		key := fmt.Sprintf("%d:%d:%s", intArg(args, 0), intArg(args, 1), stringArg(args, 2))
		if err := c.persistGameplayRecord(c.store.receipts, key, args); err != nil {
			return nil, err
		}
		return driver.RowsAffected(1), nil
	}
	if strings.HasPrefix(normalized, "insert into gameplay_ai_exploration_history (") {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		key := fmt.Sprintf("%d:%d:%d", intArg(args, 0), intArg(args, 1), intArg(args, 2))
		if err := c.persistGameplayRecord(c.store.exploration, key, args); err != nil {
			return nil, err
		}
		return driver.RowsAffected(1), nil
	}
	if strings.HasPrefix(normalized, "insert into gameplay_events (") {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		key := fmt.Sprintf("%d:%d", intArg(args, 0), intArg(args, 1))
		if err := c.persistGameplayRecord(c.store.events, key, args); err != nil {
			return nil, err
		}
		return driver.RowsAffected(1), nil
	}
	if strings.HasPrefix(normalized, "insert into gameplay_event_outbox (") {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		gameID, sequence := intArg(args, 0), intArg(args, 1)
		key := fmt.Sprintf("%d:%d", gameID, sequence)
		if _, exists := c.store.outbox[key]; exists {
			return nil, fmt.Errorf("memorysql: duplicate gameplay outbox record %q", key)
		}
		c.store.outbox[key] = outboxRecord{gameID: gameID, sequence: sequence}
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { delete(c.store.outbox, key) })
		}
		return driver.RowsAffected(1), nil
	}
	if normalized == "update gameplay_event_outbox set delivered_at = now(), delivery_attempts = delivery_attempts + 1 where game_id = $1 and sequence = $2 and delivered_at is null" {
		c.store.mu.Lock()
		defer c.store.mu.Unlock()
		key := fmt.Sprintf("%d:%d", intArg(args, 0), intArg(args, 1))
		record, ok := c.store.outbox[key]
		if !ok || record.delivered {
			return driver.RowsAffected(0), nil
		}
		previous := record
		record.delivered, record.attempts = true, record.attempts+1
		c.store.outbox[key] = record
		if c.tx != nil {
			c.tx.undo = append(c.tx.undo, func() { c.store.outbox[key] = previous })
		}
		return driver.RowsAffected(1), nil
	}
	return nil, c.unknown(query, "unsupported exec")
}

func (c *conn) persistGameplayRecord(records map[string][]driver.Value, key string, args []driver.NamedValue) error {
	if _, exists := records[key]; exists {
		return fmt.Errorf("memorysql: duplicate gameplay record %q", key)
	}
	values := make([]driver.Value, len(args))
	for index := range args {
		values[index] = args[index].Value
	}
	records[key] = values
	if c.tx != nil {
		c.tx.undo = append(c.tx.undo, func() { delete(records, key) })
	}
	return nil
}

func gameplayEventRows(records map[string][]driver.Value, gameID, cursor int64, audit bool) [][]driver.Value {
	selected := make([][]driver.Value, 0)
	for _, event := range records {
		if intArgFromValue(event[0]) != gameID || !audit && intArgFromValue(event[1]) <= cursor {
			continue
		}
		selected = append(selected, event)
	}
	sort.Slice(selected, func(i, j int) bool {
		leftSequence, rightSequence := intArgFromValue(selected[i][1]), intArgFromValue(selected[j][1])
		if leftSequence != rightSequence {
			return leftSequence < rightSequence
		}
		return intArgFromValue(selected[i][3]) < intArgFromValue(selected[j][3])
	})
	values := make([][]driver.Value, 0, len(selected))
	for _, event := range selected {
		if audit {
			values = append(values, []driver.Value{event[4], event[6]})
		} else {
			values = append(values, []driver.Value{event[0], event[1], event[2], event[3], event[5], event[6]})
		}
	}
	return values
}

func (c *conn) record(query string) {
	c.store.telemetry.mu.Lock()
	c.store.telemetry.Queries = append(c.store.telemetry.Queries, query)
	c.store.telemetry.mu.Unlock()
}
func (c *conn) unknown(query, reason string) error {
	err := fmt.Errorf("memorysql: %s: %s", reason, query)
	c.store.telemetry.mu.Lock()
	c.store.telemetry.Unknown = err
	c.store.telemetry.mu.Unlock()
	return err
}

type tx struct {
	conn *conn
	undo []func()
}

func (t *tx) Commit() error {
	if t.conn.tx != t {
		return errors.New("memorysql: completed transaction")
	}
	t.conn.tx = nil
	return nil
}
func (t *tx) Rollback() error {
	if t.conn.tx != t {
		return errors.New("memorysql: completed transaction")
	}
	for index := len(t.undo) - 1; index >= 0; index-- {
		t.undo[index]()
	}
	t.conn.tx = nil
	return nil
}
func (t *tx) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return t.conn.QueryContext(ctx, query, args)
}
func (t *tx) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return t.conn.ExecContext(ctx, query, args)
}

type rows struct {
	columns []string
	values  [][]driver.Value
	pos     int
}

func (r rows) Columns() []string { return r.columns }
func (r *rows) Close() error     { return nil }
func (r *rows) Next(dest []driver.Value) error {
	if r.pos >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.pos])
	r.pos++
	return nil
}

func normalize(query string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	return strings.TrimSuffix(normalized, " for update")
}
func stringArg(args []driver.NamedValue, index int) string {
	if index >= len(args) || args[index].Value == nil {
		return ""
	}
	return fmt.Sprint(args[index].Value)
}
func timeArg(args []driver.NamedValue, index int) time.Time {
	if index >= len(args) {
		return time.Time{}
	}
	value, _ := args[index].Value.(time.Time)
	return value
}

func intArgFromValue(value driver.Value) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return 0
	}
}

func intArg(args []driver.NamedValue, index int) int64 {
	if index >= len(args) {
		return 0
	}
	switch value := args[index].Value.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	default:
		return 0
	}
}
