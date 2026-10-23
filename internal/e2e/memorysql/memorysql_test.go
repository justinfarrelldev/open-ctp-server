package memorysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	gameplayevents "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

func TestSessionLookupReturnsStoredSessionAndRejectsUnknownID(t *testing.T) {
	db, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("INSERT INTO sessions (id, account_id, created_at, expires_at) VALUES (?, ?, ?, ?)", int64(9), int64(3), time.Unix(1, 0), time.Unix(2, 0)); err != nil {
		t.Fatal(err)
	}
	var accountID int64
	if err := db.QueryRow("SELECT * FROM sessions WHERE id = $1", int64(9)).Scan(new(int64), &accountID, new(time.Time), new(time.Time)); err != nil {
		t.Fatal(err)
	}
	if accountID != 3 {
		t.Fatalf("account ID = %d, want 3", accountID)
	}
	if err := db.QueryRow("SELECT * FROM sessions WHERE id = $1", int64(10)).Scan(new(int64), new(int64), new(time.Time), new(time.Time)); err != sql.ErrNoRows {
		t.Fatalf("unknown session error = %v, want sql.ErrNoRows", err)
	}
	if telemetry.Unknown != nil {
		t.Fatal(telemetry.Unknown)
	}
}

func TestResetDiscardsNamedStoreState(t *testing.T) {
	name := t.Name()
	first, _, err := Open(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Exec("INSERT INTO sessions (id, account_id, created_at, expires_at) VALUES (?, ?, ?, ?)", int64(9), int64(3), time.Unix(1, 0), time.Unix(2, 0)); err != nil {
		first.Close()
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	Reset(name)
	second, telemetry, err := Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if err := second.QueryRow("SELECT * FROM sessions WHERE id = $1", int64(9)).Scan(new(int64), new(int64), new(time.Time), new(time.Time)); err != sql.ErrNoRows {
		t.Fatalf("reset store retained session: %v", err)
	}
	if telemetry.Unknown != nil {
		t.Fatal(telemetry.Unknown)
	}
}

func TestRollbackDiscardsInsertedAccount(t *testing.T) {
	db, _, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Query("INSERT INTO account (name, info, location, email, experience_level) VALUES ($1, $2, $3, $4, $5) RETURNING id", "owner", "", "", "owner@example.invalid", 1); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT * from account WHERE email = $1", "owner@example.invalid").Scan(new(int64)); err != sql.ErrNoRows {
		t.Fatalf("rolled back account error = %v, want sql.ErrNoRows", err)
	}
}

func TestOrderedLobbyMembershipRowsAndEmptyResult(t *testing.T) {
	db, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const query = `SELECT account_id FROM lobby_members WHERE lobby_id = $1 ORDER BY joined_at ASC, id ASC`
	for _, accountID := range []int64{42, 7, 99} {
		if _, err := db.Exec("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)", int64(12), accountID); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := db.Query(query, int64(12))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []int64
	for rows.Next() {
		var accountID int64
		if err := rows.Scan(&accountID); err != nil {
			t.Fatal(err)
		}
		got = append(got, accountID)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	want := []int64{42, 7, 99}
	if len(got) != len(want) {
		t.Fatalf("membership row count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("membership row %d = %d, want %d", index, got[index], want[index])
		}
	}
	if err := db.QueryRow(query, int64(13)).Scan(new(int64)); err != sql.ErrNoRows {
		t.Fatalf("empty membership error = %v, want sql.ErrNoRows", err)
	}
	if telemetry.Unknown != nil {
		t.Fatal(telemetry.Unknown)
	}
}

func TestGameSetupUpdatePersistsStartableState(t *testing.T) {
	db, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const createGame = `INSERT INTO games (
		name,
		owner_account_id,
		lobby_id,
		password_protected,
		password,
		is_locked,
		is_public,
		style,
		rules_json,
		exclusions_json,
		status,
		starting_settlers,
		starting_gold
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
	) RETURNING id`
	var gameID int64
	if err := db.QueryRow(createGame, "setup-game", int64(1), int64(0), false, "", false, true, "", "{}", "{}", "setup", int64(1), int64(0)).Scan(&gameID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE games SET style = $1, map_size = $2, climate = $3, wetness = $4, updated_at = NOW() WHERE id = $5`, "sequential", "duel", "temperate", 50, gameID); err != nil {
		t.Fatal(err)
	}

	storesMu.Lock()
	stored := stores[t.Name()]
	storesMu.Unlock()
	stored.mu.Lock()
	game := stored.games[gameID]
	stored.mu.Unlock()
	if game.status != "setup" || game.style != "sequential" || game.mapSize != "duel" || game.climate != "temperate" || game.wetness == nil || *game.wetness != 50 {
		t.Fatalf("game setup state = %#v, want persisted startable setup", game)
	}
	if telemetry.Unknown != nil {
		t.Fatal(telemetry.Unknown)
	}
}

func TestGameplayAggregateCommandRecordsSurviveRollbackAndRestart(t *testing.T) {
	db, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	seedGameplayAggregate(t, db, 41)
	store := gameplaystore.New(sqlx.NewDb(db, DriverName))
	aggregate, err := store.ReadAggregate(41)
	if err != nil {
		t.Fatal(err)
	}
	if aggregate.Revision != 1 || aggregate.LastEventSequence != 1 || aggregate.Ruleset.ContentHash != rulesHash {
		t.Fatalf("aggregate = %#v, want revisioned rules pin", aggregate)
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`UPDATE world_turn_states SET canonical_state_json = $1::jsonb, revision = revision + 1, last_event_sequence = $2, updated_at = NOW() WHERE game_id = $3 AND state_schema_version = 1 AND revision = $4`, `{"revision":2}`, 2, 41, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO gameplay_command_receipts (game_id, actor_player_id, idempotency_key, payload_sha256, result_json, resulting_revision, resulting_event_sequence) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`, 41, 8, "rollback", payloadHash, `{}`, 2, 2); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	aggregate, err = store.ReadAggregate(41)
	if err != nil {
		t.Fatal(err)
	}
	if aggregate.Revision != 1 || aggregate.LastEventSequence != 1 {
		t.Fatalf("rollback aggregate = %#v", aggregate)
	}
	restarted, _, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	aggregate, err = gameplaystore.New(sqlx.NewDb(restarted, DriverName)).ReadAggregate(41)
	if err != nil || aggregate.Revision != 1 || telemetry.Unknown != nil {
		t.Fatalf("restart aggregate = %#v, err = %v, unknown = %v", aggregate, err, telemetry.Unknown)
	}
}

func TestGameplayReceiptReplayAndConflictArePersisted(t *testing.T) {
	db, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	seedGameplayAggregate(t, db, 43)
	const insertReceipt = `INSERT INTO gameplay_command_receipts (game_id, actor_player_id, idempotency_key, payload_sha256, result_json, resulting_revision, resulting_event_sequence) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`
	if _, err := db.Exec(insertReceipt, 43, 8, "replay-key", payloadHash, `{"revision":2}`, 2, 2); err != nil {
		t.Fatal(err)
	}
	var result string
	if err := db.QueryRow(`SELECT game_id, actor_player_id, idempotency_key, payload_sha256, result_json::text AS result_json, resulting_revision, resulting_event_sequence FROM gameplay_command_receipts WHERE game_id = $1 AND actor_player_id = $2 AND idempotency_key = $3`, 43, 8, "replay-key").Scan(new(int64), new(int64), new(string), new(string), &result, new(int64), new(int64)); err != nil {
		t.Fatal(err)
	}
	if result != `{"revision":2}` {
		t.Fatalf("replayed receipt result = %s", result)
	}
	if _, err := db.Exec(insertReceipt, 43, 8, "replay-key", rulesHash, `{"revision":3}`, 3, 3); err == nil {
		t.Fatal("duplicate receipt with conflicting payload was accepted")
	}
	if telemetry.Unknown != nil {
		t.Fatal(telemetry.Unknown)
	}
}

func TestAIExplorationHistorySurvivesRestartAndExcludesOtherActors(t *testing.T) {
	db, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const insertHistory = `INSERT INTO gameplay_ai_exploration_history (game_id, actor_player_id, sequence, command_type, destination_tile_id) VALUES ($1, $2, $3, $4, $5)`
	for _, step := range []struct {
		actor, sequence int64
		destination     string
	}{{8, 3, "visible-a"}, {8, 5, "visible-b"}, {9, 4, "foreign-step"}} {
		if _, err := db.Exec(insertHistory, 43, step.actor, step.sequence, "move-unit", step.destination); err != nil {
			t.Fatal(err)
		}
	}
	store := gameplaystore.New(sqlx.NewDb(db, DriverName))
	steps, err := store.AIExplorationHistory(43, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 || steps[0].DestinationTileID != "visible-a" || steps[1].DestinationTileID != "visible-b" {
		t.Fatalf("actor-scoped history = %#v", steps)
	}
	foreign, err := store.AIExplorationHistory(43, 9)
	if err != nil || len(foreign) != 1 || foreign[0].DestinationTileID != "foreign-step" {
		t.Fatalf("foreign actor history = %#v, err=%v", foreign, err)
	}
	restarted, _, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	rebuilt, err := gameplaystore.New(sqlx.NewDb(restarted, DriverName)).AIExplorationHistory(43, 8)
	if err != nil || len(rebuilt) != len(steps) || rebuilt[0] != steps[0] || rebuilt[1] != steps[1] {
		t.Fatalf("restart history = %#v, err=%v, want %#v", rebuilt, err, steps)
	}
	if telemetry.Unknown != nil {
		t.Fatal(telemetry.Unknown)
	}
}

func TestGameplayRecordsProvideOrderedOutboxAndSSEResume(t *testing.T) {
	db, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	seedGameplayAggregate(t, db, 42)
	for _, record := range []struct {
		sequence int64
		eventID  string
	}{
		{2, "event-2"}, {3, "event-3"},
	} {
		if _, err := db.Exec(`INSERT INTO gameplay_events (game_id, sequence, revision, ordinal, event_id, event_type, payload_json) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`, 42, record.sequence, record.sequence, 0, record.eventID, "command", `{"visible":true}`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO gameplay_event_outbox (game_id, sequence) VALUES ($1, $2)`, 42, record.sequence); err != nil {
			t.Fatal(err)
		}
	}
	stream := gameplayevents.NewStore(sqlx.NewDb(db, DriverName))
	resumed, err := stream.Resume(context.Background(), gameplayevents.Viewer{GameID: 42, PlayerID: 8, AccountID: 8}, 1, gameplayevents.ProjectorFuncs{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resumed) != 2 || resumed[0].ID != 2 || resumed[1].ID != 3 {
		t.Fatalf("SSE resume = %#v, want ordered events after Last-Event-ID", resumed)
	}
	rows, err := db.Query(`SELECT o.game_id, o.sequence FROM gameplay_event_outbox o WHERE o.delivered_at IS NULL ORDER BY o.game_id ASC, o.sequence ASC FOR UPDATE SKIP LOCKED`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var sequences []int64
	for rows.Next() {
		var gameID, sequence int64
		if err := rows.Scan(&gameID, &sequence); err != nil {
			t.Fatal(err)
		}
		sequences = append(sequences, sequence)
	}
	if len(sequences) != 2 || sequences[0] != 2 || sequences[1] != 3 {
		t.Fatalf("undelivered outbox = %v", sequences)
	}
	if _, err := db.Exec(`UPDATE gameplay_event_outbox SET delivered_at = NOW(), delivery_attempts = delivery_attempts + 1 WHERE game_id = $1 AND sequence = $2 AND delivered_at IS NULL`, 42, 2); err != nil {
		t.Fatal(err)
	}
	if telemetry.Unknown != nil {
		t.Fatal(telemetry.Unknown)
	}
}

const rulesHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const payloadHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func TestDriverInterfacesRejectUnsupportedOperationsAndCompletedTransactions(t *testing.T) {
	if _, err := (memoryDriver{}).Open(t.Name() + "-missing"); err == nil {
		t.Fatal("driver opened an unknown named store")
	}
	database, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	raw, err := (memoryDriver{}).Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	connection := raw.(*conn)
	defer connection.Close()
	if err := connection.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Prepare("SELECT 1"); err == nil {
		t.Fatal("prepared statement was accepted")
	}

	rawTransaction, err := connection.Begin()
	if err != nil {
		t.Fatal(err)
	}
	transaction := rawTransaction.(*tx)
	queryRows, err := transaction.QueryContext(context.Background(), "SELECT * FROM sessions WHERE id = $1", []driver.NamedValue{{Ordinal: 1, Value: int64(999)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := queryRows.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(context.Background(), "UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1", []driver.NamedValue{{Ordinal: 1, Value: int64(1)}}); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err == nil {
		t.Fatal("completed transaction committed twice")
	}

	rawTransaction, err = connection.Begin()
	if err != nil {
		t.Fatal(err)
	}
	transaction = rawTransaction.(*tx)
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Rollback(); err == nil {
		t.Fatal("completed transaction rolled back twice")
	}
	if _, err := connection.QueryContext(context.Background(), "SELECT unsupported", nil); err == nil {
		t.Fatal("unsupported query was accepted")
	}
	if telemetry.Unknown == nil {
		t.Fatal("unsupported query was not recorded in telemetry")
	}
	if _, err := connection.ExecContext(context.Background(), "UPDATE unsupported", nil); err == nil {
		t.Fatal("unsupported exec was accepted")
	}
}

func TestDriverArgumentCoercionIsDeterministic(t *testing.T) {
	if got := stringArg([]driver.NamedValue{{Value: []byte("bytes")}}, 0); got != "[98 121 116 101 115]" {
		t.Fatalf("byte string = %q", got)
	}
	if got := stringArg([]driver.NamedValue{{Value: nil}}, 0); got != "" {
		t.Fatalf("nil string = %q", got)
	}
	when := time.Unix(123, 0).UTC()
	if got := timeArg([]driver.NamedValue{{Value: when}}, 0); !got.Equal(when) {
		t.Fatalf("time value = %v, want %v", got, when)
	}
	if got := timeArg([]driver.NamedValue{{Value: "invalid"}}, 0); !got.IsZero() {
		t.Fatalf("invalid time = %v, want zero", got)
	}
	values := []struct {
		value any
		want  int64
	}{
		{value: int64(1), want: 1},
		{value: int(2), want: 2},
		{value: int32(3), want: 0},
		{value: float64(4), want: 0},
		{value: []byte("5"), want: 0},
		{value: "6", want: 0},
		{value: "invalid", want: 0},
	}
	for _, test := range values {
		if got := intArgFromValue(test.value); got != test.want {
			t.Fatalf("intArgFromValue(%T(%v)) = %d, want %d", test.value, test.value, got, test.want)
		}
	}
	if got := intArg([]driver.NamedValue{{Value: "7"}}, 0); got != 0 {
		t.Fatalf("intArg string = %d, want 0", got)
	}
	if got := intArg(nil, 0); got != 0 {
		t.Fatalf("missing intArg = %d, want 0", got)
	}
}

func TestDriverMissingDuplicateAndDeliverySemantics(t *testing.T) {
	database, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	assertRows := func(query string, want int64, args ...any) {
		t.Helper()
		result, err := database.Exec(query, args...)
		if err != nil {
			t.Fatalf("exec %q: %v", query, err)
		}
		got, err := result.RowsAffected()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("exec %q affected %d rows, want %d", query, got, want)
		}
	}

	assertRows("UPDATE game_players SET controller_kind = 'ai', updated_at = NOW() WHERE id = $1", 0, int64(99))
	assertRows("UPDATE game_players SET controller_kind = $1, controller_changed_at = NOW() WHERE game_id = $2 AND id = $3", 0, "ai", int64(99), int64(1))
	assertRows("UPDATE games SET status = 'ended', ended_at = NOW(), ended_by = $1, end_reason = $2, updated_at = NOW() WHERE id = $3", 0, int64(1), "test", int64(99))
	assertRows("UPDATE games SET status = $1, updated_at = NOW() WHERE id = $2", 0, "started", int64(99))
	assertRows("UPDATE games SET status = $1, started_at = NOW(), updated_at = NOW() WHERE id = $2", 0, "started", int64(99))
	assertRows("UPDATE games SET rules_json = $1, updated_at = NOW() WHERE id = $2", 0, "[]", int64(99))
	assertRows("UPDATE games SET starting_settlers = $1, updated_at = NOW() WHERE id = $2", 0, int64(2), int64(99))
	assertRows("UPDATE games SET style = $1, map_size = $2, climate = $3, wetness = $4, updated_at = NOW() WHERE id = $5", 0, "sequential", "duel", "temperate", int64(50), int64(99))

	assertRows("INSERT INTO world_snapshots (game_id) VALUES ($1)", 1, int64(9))
	if _, err := database.Exec("INSERT INTO world_snapshots (game_id) VALUES ($1)", int64(9)); err == nil {
		t.Fatal("duplicate world snapshot was accepted")
	}
	assertRows("INSERT INTO world_turn_states (game_id) VALUES ($1)", 1, int64(9))
	if _, err := database.Exec("INSERT INTO world_turn_states (game_id) VALUES ($1)", int64(9)); err == nil {
		t.Fatal("duplicate world turn state was accepted")
	}
	assertRows("INSERT INTO gameplay_event_outbox (game_id, sequence) VALUES ($1, $2)", 1, int64(9), int64(1))
	if _, err := database.Exec("INSERT INTO gameplay_event_outbox (game_id, sequence) VALUES ($1, $2)", int64(9), int64(1)); err == nil {
		t.Fatal("duplicate gameplay outbox record was accepted")
	}
	assertRows("UPDATE gameplay_event_outbox SET delivered_at = NOW(), delivery_attempts = delivery_attempts + 1 WHERE game_id = $1 AND sequence = $2 AND delivered_at IS NULL", 1, int64(9), int64(1))
	assertRows("UPDATE gameplay_event_outbox SET delivered_at = NOW(), delivery_attempts = delivery_attempts + 1 WHERE game_id = $1 AND sequence = $2 AND delivered_at IS NULL", 0, int64(9), int64(1))
	assertRows("UPDATE gameplay_event_outbox SET delivered_at = NOW(), delivery_attempts = delivery_attempts + 1 WHERE game_id = $1 AND sequence = $2 AND delivered_at IS NULL", 0, int64(9), int64(2))
	if telemetry.Unknown == nil {
		t.Fatal("duplicate persistence errors were not recorded")
	}
}

func TestDriverEmptyStoreQueriesFailClosedWithoutFabricatedRows(t *testing.T) {
	database, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	assertNoRows := func(query string, args ...any) {
		t.Helper()
		rows, err := database.Query(query, args...)
		if err != nil {
			t.Fatalf("query %q: %v", query, err)
		}
		defer rows.Close()
		if rows.Next() {
			t.Fatalf("query %q fabricated a row", query)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("query %q iteration: %v", query, err)
		}
	}

	assertNoRows("SELECT id, status, owner_account_id FROM games WHERE id = $1", int64(99))
	assertNoRows("SELECT id, controller_kind FROM game_players WHERE game_id = $1 AND account_id = $2", int64(99), int64(1))
	assertNoRows("SELECT id, account_id FROM game_players WHERE game_id = $1 AND account_id = $2", int64(99), int64(1))
	assertNoRows("SELECT id, controller_kind FROM game_players WHERE game_id = $1 AND id = $2", int64(99), int64(1))
	assertNoRows("SELECT id FROM game_players WHERE game_id = $1 AND account_id = $2", int64(99), int64(1))
	assertNoRows("SELECT * FROM account WHERE email = $1", "missing@example.invalid")
	assertNoRows("SELECT name FROM account WHERE id = $1", int64(99))
	assertNoRows("SELECT id, closed_at FROM lobbies WHERE id = $1", int64(99))
	assertNoRows("SELECT game_id FROM world_snapshots WHERE game_id = $1", int64(99))
	assertNoRows("SELECT revision, last_event_sequence, canonical_state_json FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1", int64(99))
	assertNoRows("SELECT last_event_sequence, canonical_state_json FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1", int64(99))
	assertNoRows("SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1", int64(99))
	assertNoRows("SELECT game_id, actor_player_id, idempotency_key, payload_sha256, result_json::text AS result_json, resulting_revision, resulting_event_sequence FROM gameplay_command_receipts WHERE game_id = $1 AND actor_player_id = $2 AND idempotency_key = $3", int64(99), int64(1), "missing")
	assertNoRows("SELECT state.game_id, state.revision, player.value->>'id' AS player_id FROM world_turn_states AS state JOIN games ON games.id = state.game_id CROSS JOIN LATERAL jsonb_array_elements(state.canonical_state_json->'players') AS player(value) WHERE state.state_schema_version = 1 AND games.status = 'started' AND COALESCE(player.value->>'controller', '') = 'human' AND COALESCE(player.value->>'takeover_pending', 'false') = 'true' ORDER BY state.game_id ASC, player_id ASC")
	assertNoRows("SELECT game_id, sequence, revision, ordinal, event_type, payload_json FROM gameplay_events WHERE game_id = $1 AND sequence > $2 ORDER BY sequence ASC", int64(99), int64(0))

	var minimum sql.NullInt64
	if err := database.QueryRow("SELECT min(sequence) FROM gameplay_events WHERE game_id = $1", int64(99)).Scan(&minimum); err != nil {
		t.Fatal(err)
	}
	if minimum.Valid {
		t.Fatalf("empty event minimum = %d, want NULL", minimum.Int64)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("recognized empty queries produced unknown SQL: %v", telemetry.Unknown)
	}
}

func TestDueLongTurnDeadlinesFilterAndSortPersistedPolicyState(t *testing.T) {
	database, telemetry, err := Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	states := []struct {
		gameID   int64
		mode     string
		playerID string
		deadline int64
	}{
		{gameID: 20, mode: "long_turn", playerID: "player-20", deadline: 100},
		{gameID: 10, mode: "long_turn", playerID: "player-10", deadline: 50},
		{gameID: 30, mode: "long_turn", playerID: "player-30", deadline: 200},
		{gameID: 40, mode: "sequential", playerID: "player-40", deadline: 25},
	}
	for _, state := range states {
		seedGameplayAggregate(t, database, state.gameID)
		canonical := fmt.Sprintf(`{"phase":{"mode":%q},"policy":{"deadline":{"player_id":%q,"at_unix":%d}}}`, state.mode, state.playerID, state.deadline)
		result, err := database.Exec(`UPDATE world_turn_states SET canonical_state_json = $1::jsonb, revision = revision + 1, last_event_sequence = $2, updated_at = NOW() WHERE game_id = $3 AND state_schema_version = 1 AND revision = $4`, canonical, int64(2), state.gameID, int64(1))
		if err != nil {
			t.Fatal(err)
		}
		if rows, err := result.RowsAffected(); err != nil || rows != 1 {
			t.Fatalf("update game %d affected %d rows, err=%v", state.gameID, rows, err)
		}
	}

	deadlines, err := gameplaystore.New(sqlx.NewDb(database, DriverName)).FindDueLongTurnDeadlines(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(deadlines) != 2 || deadlines[0].GameID != 10 || deadlines[0].PlayerID != "player-10" || deadlines[0].DeadlineAt != 50 || deadlines[1].GameID != 20 || deadlines[1].PlayerID != "player-20" || deadlines[1].DeadlineAt != 100 {
		t.Fatalf("due deadlines = %#v", deadlines)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("deadline query used unknown SQL: %v", telemetry.Unknown)
	}
}

func seedGameplayAggregate(t *testing.T, db *sql.DB, gameID int64) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO world_snapshots (game_id, game_status, generator_version, style, rules_json, exclusions_json, starting_settlers, starting_gold, seed, wetness, map_size, climate, player_order_json, spawns_json, tiles_json, ruleset_canonical_json, ruleset_content_hash, ruleset_version, updated_at) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb, $15::jsonb, $16::jsonb, $17, $18, NOW())`, gameID, "started", "test", "sequential", `{}`, `{}`, 1, 0, 1, 50, "duel", "temperate", `[]`, `[]`, `[]`, `{}`, rulesHash, "v1"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO world_turn_states (game_id, turn_number, current_player_index, active_player_order_json, recalculation_count, rng_state, ratings_json, pollution_json, last_recalculated_at, state_schema_version, revision, last_event_sequence, canonical_state_json, updated_at) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7::jsonb, $8::jsonb, $9, $10, $11, $12, $13::jsonb, NOW())`, gameID, 1, 0, `[]`, 0, 1, `{}`, `{}`, time.Unix(1, 0), 1, 1, 1, `{"revision":1}`); err != nil {
		t.Fatal(err)
	}
}
