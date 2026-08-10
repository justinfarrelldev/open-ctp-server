package memorysql

import (
	"context"
	"database/sql"
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

func seedGameplayAggregate(t *testing.T, db *sql.DB, gameID int64) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO world_snapshots (game_id, game_status, generator_version, style, rules_json, exclusions_json, starting_settlers, starting_gold, seed, wetness, map_size, climate, player_order_json, spawns_json, tiles_json, ruleset_canonical_json, ruleset_content_hash, ruleset_version, updated_at) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb, $15::jsonb, $16::jsonb, $17, $18, NOW())`, gameID, "started", "test", "sequential", `{}`, `{}`, 1, 0, 1, 50, "duel", "temperate", `[]`, `[]`, `[]`, `{}`, rulesHash, "v1"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO world_turn_states (game_id, turn_number, current_player_index, active_player_order_json, recalculation_count, rng_state, ratings_json, pollution_json, last_recalculated_at, state_schema_version, revision, last_event_sequence, canonical_state_json, updated_at) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7::jsonb, $8::jsonb, $9, $10, $11, $12, $13::jsonb, NOW())`, gameID, 1, 0, `[]`, 0, 1, `{}`, `{}`, time.Unix(1, 0), 1, 1, 1, `{"revision":1}`); err != nil {
		t.Fatal(err)
	}
}
