package game

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

const testWorldSnapshotSeed = int64(991337)

const testWorldSnapshotExistsQuery = `SELECT game_id FROM world_snapshots WHERE game_id = $1`

const testWorldSnapshotInsertQuery = `INSERT INTO world_snapshots (
	game_id,
	game_status,
	generator_version,
	style,
	rules_json,
	exclusions_json,
	starting_settlers,
	starting_gold,
	seed,
	wetness,
	map_size,
	climate,
	player_order_json,
	spawns_json,
	tiles_json,
	updated_at
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5::jsonb,
	$6::jsonb,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13::jsonb,
	$14::jsonb,
	$15::jsonb,
	NOW()
)`

const testWorldTurnStateInsertQuery = `INSERT INTO world_turn_states (
	game_id,
	turn_number,
	current_player_index,
	active_player_order_json,
	recalculation_count,
	rng_state,
	ratings_json,
	pollution_json,
	last_recalculated_at,
	updated_at
) VALUES (
	$1,
	$2,
	$3,
	$4::jsonb,
	$5,
	$6,
	$7::jsonb,
	$8::jsonb,
	$9,
	NOW()
)`

func TestStartGameGeneratesWorldOnce(t *testing.T) {
	stubStartWorldSeed(t, testWorldSnapshotSeed, nil)

	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	lobbyID := int64(18)
	ownerID := int64(7)
	startedAt := time.Now()
	players := []GamePlayer{
		{AccountID: 7, Country: "Greeks"},
		{AccountID: 8, Country: "Romans"},
	}
	setupGame := &Game{
		ID:               44,
		LobbyID:          &lobbyID,
		OwnerAccountID:   &ownerID,
		Style:            DefaultGameStyle,
		RulesJSON:        defaultRulesJSON,
		ExclusionsJSON:   defaultExclusionsJSON,
		Status:           DefaultGameStatus,
		StartingSettlers: 2,
		StartingGold:     250,
	}
	snapshot, turn := startedWorldFixture(t, setupGame, players)

	expectSuccessfulStartGame(t, mock, snapshot, turn, &lobbyID, &startedAt)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByIDForUpdate)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, &lobbyID, "started", "Task 11 Game", &startedAt, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	mock.ExpectRollback()

	startReq := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	startRR := httptest.NewRecorder()
	if err := StartGameWithSeedAndSources(startRR, startReq, sqlxDB, store, func() (int64, error) { return testWorldSnapshotSeed, nil }, testStartGameplaySources()); err != nil {
		t.Fatalf("StartGameWithSeedAndSources() error = %v", err)
	}
	if startRR.Code != http.StatusOK {
		t.Fatalf("first start got status %d want %d body=%q", startRR.Code, http.StatusOK, startRR.Body.String())
	}

	duplicateReq := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	duplicateRR := httptest.NewRecorder()
	if err := StartGameWithSeedAndSources(duplicateRR, duplicateReq, sqlxDB, store, func() (int64, error) { return testWorldSnapshotSeed, nil }, testStartGameplaySources()); err == nil {
		t.Fatal("StartGameWithSeedAndSources() duplicate error = nil")
	}
	if duplicateRR.Code != http.StatusConflict {
		t.Fatalf("second start got status %d want %d body=%q", duplicateRR.Code, http.StatusConflict, duplicateRR.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestStartGamePersistsWorldInTransaction(t *testing.T) {
	stubStartWorldSeed(t, testWorldSnapshotSeed, nil)

	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	lobbyID := int64(18)
	ownerID := int64(7)
	startedAt := time.Now()
	players := []GamePlayer{
		{AccountID: 7, Country: "Greeks"},
		{AccountID: 8, Country: "Romans"},
	}
	setupGame := &Game{
		ID:               44,
		LobbyID:          &lobbyID,
		OwnerAccountID:   &ownerID,
		Style:            DefaultGameStyle,
		RulesJSON:        defaultRulesJSON,
		ExclusionsJSON:   defaultExclusionsJSON,
		Status:           DefaultGameStatus,
		StartingSettlers: 2,
		StartingGold:     250,
	}
	snapshot, turn := startedWorldFixture(t, setupGame, players)

	expectSuccessfulStartGame(t, mock, snapshot, turn, &lobbyID, &startedAt)

	req := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	rr := httptest.NewRecorder()
	if err := StartGameWithSeedAndSources(rr, req, sqlxDB, store, func() (int64, error) { return testWorldSnapshotSeed, nil }, testStartGameplaySources()); err != nil {
		t.Fatalf("StartGameWithSeedAndSources() error = %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"status":"started"`) {
		t.Fatalf("expected started response, got %q", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestStartGameDoesNotStartOnWorldGenerationFailure(t *testing.T) {
	stubStartWorldSeed(t, testWorldSnapshotSeed, nil)

	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByIDForUpdate)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Broken Game", nil, false, false, "", defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	mock.ExpectQuery(regexp.QuoteMeta(gamePlayersSelectByGameID)).
		WithArgs(int64(44)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: 7, country: "Greeks", ping: 0},
		))
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	rr := httptest.NewRecorder()
	StartGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "style must be specified") {
		t.Fatalf("expected generation failure body, got %q", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestStartGameSerialization(t *testing.T) {
	ownerID := int64(7)
	game := &Game{
		ID:               44,
		OwnerAccountID:   &ownerID,
		Style:            " sequential ",
		RulesJSON:        " { \"fast\" : true } ",
		ExclusionsJSON:   " { \"units\" : [ \"settler\" ] } ",
		Status:           DefaultGameStatus,
		StartingSettlers: 2,
		StartingGold:     250,
	}
	players := []GamePlayer{{AccountID: 8, Country: " Romans "}, {AccountID: 7, Country: "Greeks"}}

	seed := int64(991337)
	input, err := NormalizeWorldGenerationInput(game, players, seed)
	if err != nil {
		t.Fatalf("NormalizeWorldGenerationInput() error = %v", err)
	}
	if input.GeneratorVersion != defaultWorldStartGeneratorVersion {
		t.Fatalf("got generator version %q want %q", input.GeneratorVersion, defaultWorldStartGeneratorVersion)
	}
	if input.HostParams.Seed == nil {
		t.Fatal("expected seed to be populated")
	}
	if *input.HostParams.Seed != seed {
		t.Fatalf("got seed %d want %d", *input.HostParams.Seed, seed)
	}
	if *input.HostParams.Seed == game.ID {
		t.Fatalf("seed must not be derived from public game id %d", game.ID)
	}
	if input.HostParams.Wetness == nil || *input.HostParams.Wetness != defaultWorldStartWetness {
		t.Fatalf("got wetness %v want %d", input.HostParams.Wetness, defaultWorldStartWetness)
	}
	if input.HostParams.MapSize != defaultWorldStartMapSize || input.HostParams.Climate != defaultWorldStartClimate {
		t.Fatalf("unexpected defaults: %+v", input.HostParams)
	}
	if input.HostParams.Style != "sequential" || input.HostParams.RulesJSON != `{"fast":true}` || input.HostParams.ExclusionsJSON != `{"units":["settler"]}` {
		t.Fatalf("expected compact normalized host params, got %+v", input.HostParams)
	}

	snapshot, turn, err := world.GenerateWorld(input)
	if err != nil {
		t.Fatalf("GenerateWorld() error = %v", err)
	}
	snapshot.Turn = turn.Clone()
	snapshot = ApplyWorldSpawnCountries(snapshot, players)
	world.AssertCanonicalJSONStable(t, snapshot)
	for _, spawn := range snapshot.Spawns {
		switch spawn.AccountID {
		case 8:
			if spawn.Country != "Romans" {
				t.Fatalf("got spawn country %q want Romans", spawn.Country)
			}
		case 7:
			if spawn.Country != "Greeks" {
				t.Fatalf("got spawn country %q want Greeks", spawn.Country)
			}
		}
	}
}

func startedWorldFixture(t *testing.T, game *Game, players []GamePlayer) (world.WorldSnapshot, world.MutableTurnState) {
	t.Helper()

	input, err := NormalizeWorldGenerationInput(game, players, testWorldSnapshotSeed)
	if err != nil {
		t.Fatalf("NormalizeWorldGenerationInput() error = %v", err)
	}
	snapshot, turn, err := world.GenerateWorld(input)
	if err != nil {
		t.Fatalf("GenerateWorld() error = %v", err)
	}
	snapshot.Turn = turn.Clone()
	snapshot = ApplyWorldSpawnCountries(snapshot, players)
	return snapshot, turn.Clone()
}

func stubStartWorldSeed(t *testing.T, seed int64, err error) {
	t.Helper()

	original := generateStartWorldSeed
	generateStartWorldSeed = func() (int64, error) {
		return seed, err
	}
	t.Cleanup(func() {
		generateStartWorldSeed = original
	})
}

func expectSuccessfulStartGame(t *testing.T, mock sqlmock.Sqlmock, snapshot world.WorldSnapshot, turn world.MutableTurnState, lobbyID *int64, startedAt *time.Time) {
	t.Helper()
	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByIDForUpdate)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, lobbyID, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	if lobbyID != nil {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, closed_at FROM lobbies WHERE id = $1`)).
			WithArgs(*lobbyID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "closed_at"}).AddRow(*lobbyID, nil))
	}
	mock.ExpectQuery(regexp.QuoteMeta(gamePlayersSelectByGameID)).
		WithArgs(int64(44)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: 7, country: "Greeks", ping: 0},
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 2, accountID: 8, country: "Romans", ping: 0},
		))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT account_id, controller_kind FROM game_players WHERE game_id = $1`)).
		WithArgs(int64(44)).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "controller_kind"}).AddRow(7, "human").AddRow(8, "human"))
	expectWorldCreate(t, mock, snapshot, turn, []GamePlayer{
		{ID: 1, AccountID: 7, Country: "Greeks"},
		{ID: 2, AccountID: 8, Country: "Romans"},
	})
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE games SET status = $1, started_at = NOW(), updated_at = NOW() WHERE id = $2`)).
		WithArgs("started", int64(44)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, lobbyID, "started", "Task 11 Game", startedAt, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	mock.ExpectQuery(regexp.QuoteMeta(gamePlayersSelectByGameID)).
		WithArgs(int64(44)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: 7, country: "Greeks", ping: 0},
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 2, accountID: 8, country: "Romans", ping: 0},
		))
	mock.ExpectCommit()
}

func expectWorldCreate(t *testing.T, mock sqlmock.Sqlmock, snapshot world.WorldSnapshot, _ world.MutableTurnState, players []GamePlayer) {
	t.Helper()
	testState, initialJSON, nextJSON, receipt, event := expectedStartGameplay(t, snapshot, players)
	playerOrderJSON := mustWorldCanonicalString(t, snapshot.PlayerOrder)
	spawnsJSON := mustWorldCanonicalString(t, snapshot.Spawns)
	tilesJSON := mustWorldCanonicalString(t, snapshot.Tiles)
	activePlayerOrderJSON := mustWorldCanonicalString(t, snapshot.Turn.ActivePlayerOrder)
	ratingsJSON := mustWorldCanonicalString(t, snapshot.Turn.Ratings)
	pollutionJSON := mustWorldCanonicalString(t, snapshot.Turn.Pollution)

	mock.ExpectQuery(regexp.QuoteMeta(testWorldSnapshotExistsQuery)).
		WithArgs(snapshot.GameID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO world_snapshots (
	game_id,
	game_status,
	generator_version,
	style,
	rules_json,
	exclusions_json,
	starting_settlers,
	starting_gold,
	seed,
	wetness,
	map_size,
	climate,
	player_order_json,
	spawns_json,
	tiles_json,
	ruleset_canonical_json,
	ruleset_content_hash,
	ruleset_version,
	updated_at
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5::jsonb,
	$6::jsonb,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13::jsonb,
	$14::jsonb,
	$15::jsonb,
	$16::jsonb,
	$17,
	$18,
	NOW()
)`)).WithArgs(
		snapshot.GameID, snapshot.GameStatus, snapshot.GeneratorVersion, snapshot.Config.Style,
		snapshot.Config.RulesJSON, snapshot.Config.ExclusionsJSON, snapshot.Config.StartingSettlers,
		snapshot.Config.StartingGold, snapshot.Config.Seed, snapshot.Config.Wetness, snapshot.Config.MapSize,
		snapshot.Config.Climate, playerOrderJSON, spawnsJSON, tilesJSON, string(testState.Rules.Definitions),
		testState.Rules.ContentHash, testState.Rules.Version,
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO world_turn_states (
	game_id,
	turn_number,
	current_player_index,
	active_player_order_json,
	recalculation_count,
	rng_state,
	ratings_json,
	pollution_json,
	last_recalculated_at,
	state_schema_version,
	revision,
	last_event_sequence,
	canonical_state_json,
	updated_at
) VALUES (
	$1,
	$2,
	$3,
	$4::jsonb,
	$5,
	$6,
	$7::jsonb,
	$8::jsonb,
	$9,
	$10,
	$11,
	$12,
	$13::jsonb,
	NOW()
)`)).WithArgs(
		snapshot.GameID, snapshot.Turn.TurnNumber, snapshot.Turn.CurrentPlayerIndex, activePlayerOrderJSON,
		snapshot.Turn.RecalculationCount, strconv.FormatUint(snapshot.Turn.RNGState, 10), ratingsJSON, pollutionJSON, snapshot.Turn.LastRecalculatedAt,
		world.StateSchemaV1, int64(0), int64(0), string(initialJSON),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE world_turn_states
SET
	canonical_state_json = $1::jsonb,
	revision = revision + 1,
	last_event_sequence = $2,
	updated_at = NOW()
WHERE game_id = $3 AND state_schema_version = 1 AND revision = $4`)).WithArgs(
		nextJSON, int64(1), snapshot.GameID, int64(0),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO gameplay_command_receipts (
	game_id, actor_player_id, idempotency_key, payload_sha256,
	result_json, resulting_revision, resulting_event_sequence
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`)).WithArgs(
		receipt.GameID, receipt.ActorPlayerID, receipt.IdempotencyKey, receipt.PayloadSHA256,
		receipt.ResultJSON, receipt.ResultingRevision, receipt.ResultingEventSequence,
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO gameplay_events (
	game_id, sequence, revision, ordinal, event_id, event_type, payload_json
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`)).WithArgs(
		event.GameID, event.Sequence, event.Revision, event.Ordinal, event.ID, event.Type, event.PayloadJSON,
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO gameplay_event_outbox (game_id, sequence) VALUES ($1, $2)`)).WithArgs(
		event.GameID, event.Sequence,
	).WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectedStartGameplay(t *testing.T, snapshot world.WorldSnapshot, players []GamePlayer) (engine.State, []byte, []byte, gameplaystore.CommandReceipt, gameplaystore.Event) {
	t.Helper()
	pin, err := serverRuntimePin()
	if err != nil {
		t.Fatalf("serverRuntimePin() error = %v", err)
	}
	state, err := initialGameplayState(snapshot, snapshot.Config.Style, pin, players)
	if err != nil {
		t.Fatalf("initialGameplayState() error = %v", err)
	}
	playerIDs := state.Phase.ActivePlayerOrder
	initialJSON, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatalf("CanonicalStateJSON(initial) error = %v", err)
	}
	next, events, result, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: "game-start", ActorPlayerID: playerIDs[0], Type: engine.CommandNoop}, pin, testStartGameplaySources())
	if err != nil {
		t.Fatalf("Apply(initial game-start) error = %v", err)
	}
	nextJSON, err := engine.CanonicalStateJSON(next)
	if err != nil {
		t.Fatalf("CanonicalStateJSON(next) error = %v", err)
	}
	payload, err := json.Marshal(events[0])
	if err != nil {
		t.Fatalf("Marshal(event) error = %v", err)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal(result) error = %v", err)
	}
	commandHash := sha256.Sum256([]byte("game-start"))
	return state, initialJSON, nextJSON, gameplaystore.CommandReceipt{GameID: snapshot.GameID, ActorPlayerID: players[0].ID, IdempotencyKey: "game-start", PayloadSHA256: hex.EncodeToString(commandHash[:]), ResultJSON: resultJSON, ResultingRevision: next.Revision, ResultingEventSequence: next.EventSequence}, gameplaystore.Event{GameID: snapshot.GameID, Sequence: events[0].Sequence, Revision: next.Revision, Ordinal: events[0].Ordinal, ID: events[0].ID, Type: events[0].Type, PayloadJSON: payload}
}

type fixedStartClock struct{}

func (fixedStartClock) Now() time.Time { return time.Date(2026, time.July, 19, 8, 0, 0, 0, time.UTC) }

func testStartGameplaySources() engine.Sources {
	return engine.Sources{Clock: fixedStartClock{}, IDs: &startIDs{}, RNG: startRNG{}}
}

func startRulesJSON(value string) string {
	rules := []byte(value)
	if !json.Valid(rules) || len(rules) == 0 || rules[0] != '{' {
		return `{}`
	}
	return string(rules)
}

func mustWorldCanonicalString(t *testing.T, value any) string {
	t.Helper()
	encoded, err := world.CanonicalString(value)
	if err != nil {
		t.Fatalf("CanonicalString() error = %v", err)
	}
	return encoded
}
