package game

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

	expectSuccessfulStartGame(mock, snapshot, turn, &lobbyID, &startedAt)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByIDForUpdate)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, &lobbyID, "started", "Task 11 Game", &startedAt, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	mock.ExpectRollback()

	startReq := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	startRR := httptest.NewRecorder()
	StartGameHandler(startRR, startReq, sqlxDB, store)
	if startRR.Code != http.StatusOK {
		t.Fatalf("first start got status %d want %d body=%q", startRR.Code, http.StatusOK, startRR.Body.String())
	}

	duplicateReq := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	duplicateRR := httptest.NewRecorder()
	StartGameHandler(duplicateRR, duplicateReq, sqlxDB, store)
	if duplicateRR.Code != http.StatusConflict {
		t.Fatalf("second start got status %d want %d body=%q", duplicateRR.Code, http.StatusConflict, duplicateRR.Body.String())
	}
	if strings.TrimSpace(duplicateRR.Body.String()) != "game 44 is not in setup" {
		t.Fatalf("got duplicate body %q", duplicateRR.Body.String())
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

	expectSuccessfulStartGame(mock, snapshot, turn, &lobbyID, &startedAt)

	req := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	rr := httptest.NewRecorder()
	StartGameHandler(rr, req, sqlxDB, store)

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
		Style:            " standard_ctp2 ",
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
	if input.HostParams.Style != "standard_ctp2" || input.HostParams.RulesJSON != `{"fast":true}` || input.HostParams.ExclusionsJSON != `{"units":["settler"]}` {
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

func expectSuccessfulStartGame(mock sqlmock.Sqlmock, snapshot world.WorldSnapshot, turn world.MutableTurnState, lobbyID *int64, startedAt *time.Time) {
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
	expectWorldCreate(mock, snapshot, turn)
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

func expectWorldCreate(mock sqlmock.Sqlmock, snapshot world.WorldSnapshot, turn world.MutableTurnState) {
	playerOrderJSON, err := world.CanonicalString(snapshot.PlayerOrder)
	if err != nil {
		panic(err)
	}
	spawnsJSON, err := world.CanonicalString(snapshot.Spawns)
	if err != nil {
		panic(err)
	}
	tilesJSON, err := world.CanonicalString(snapshot.Tiles)
	if err != nil {
		panic(err)
	}
	activePlayerOrderJSON, err := world.CanonicalString(turn.ActivePlayerOrder)
	if err != nil {
		panic(err)
	}
	ratingsJSON, err := world.CanonicalString(turn.Ratings)
	if err != nil {
		panic(err)
	}
	pollutionJSON, err := world.CanonicalString(turn.Pollution)
	if err != nil {
		panic(err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(testWorldSnapshotExistsQuery)).
		WithArgs(snapshot.GameID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(testWorldSnapshotInsertQuery)).
		WithArgs(
			snapshot.GameID,
			snapshot.GameStatus,
			snapshot.GeneratorVersion,
			snapshot.Config.Style,
			snapshot.Config.RulesJSON,
			snapshot.Config.ExclusionsJSON,
			snapshot.Config.StartingSettlers,
			snapshot.Config.StartingGold,
			snapshot.Config.Seed,
			snapshot.Config.Wetness,
			snapshot.Config.MapSize,
			snapshot.Config.Climate,
			playerOrderJSON,
			spawnsJSON,
			tilesJSON,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(testWorldTurnStateInsertQuery)).
		WithArgs(
			snapshot.GameID,
			turn.TurnNumber,
			turn.CurrentPlayerIndex,
			activePlayerOrderJSON,
			turn.RecalculationCount,
			strconv.FormatUint(turn.RNGState, 10),
			ratingsJSON,
			pollutionJSON,
			turn.LastRecalculatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
}
