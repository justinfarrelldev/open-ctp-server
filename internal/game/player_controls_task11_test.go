package game

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func gameRowWithLobby(gameID int64, ownerAccountID int64, lobbyID *int64, status string, name string, startedAt *time.Time, isLocked bool, isPublic bool, style string, rulesJSON string, exclusionsJSON string, settlers int, gold int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "created_at", "updated_at", "started_at", "lobby_id", "owner_account_id", "name", "password_protected", "password", "is_locked", "is_public", "style", "rules_json", "exclusions_json", "status", "starting_settlers", "starting_gold"}).
		AddRow(gameID, time.Now(), time.Now(), startedAt, lobbyID, ownerAccountID, name, false, "", isLocked, isPublic, style, rulesJSON, exclusionsJSON, status, settlers, gold)
}

func gamePlayerRows(gameID int64, players ...struct {
	id        int64
	accountID int64
	country   string
	ping      int
}) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "game_id", "account_id", "joined_at", "country", "ping_ms"})
	for _, player := range players {
		rows.AddRow(player.id, time.Now(), time.Now(), gameID, player.accountID, time.Now(), player.country, player.ping)
	}
	return rows
}

func TestCreateGame_WithLobbyCopiesLobbyMembersIntoGamePlayers(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	lobbyID := int64(18)

	body := map[string]any{
		"session_id":         55,
		"name":               "Task 11 Game",
		"lobby_id":           lobbyID,
		"password_protected": false,
	}
	jsonBody, _ := json.Marshal(body)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, closed_at FROM lobbies WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "closed_at"}).AddRow(lobbyID, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT account_id FROM lobby_members WHERE lobby_id = $1 ORDER BY joined_at ASC, id ASC`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(int64(7)).AddRow(int64(8)))
	mock.ExpectQuery(regexp.QuoteMeta(insertGameQuery)).
		WithArgs(
			"Task 11 Game",
			int64(7),
			lobbyID,
			false,
			"",
			false,
			false,
			DefaultGameStyle,
			defaultRulesJSON,
			defaultExclusionsJSON,
			DefaultGameStatus,
			DefaultStartingSettlers,
			DefaultStartingGold,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(44))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO game_players (game_id, account_id) VALUES ($1, $2)`)).
		WithArgs(int64(44), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO game_players (game_id, account_id) VALUES ($1, $2)`)).
		WithArgs(int64(44), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, &lobbyID, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 ORDER BY joined_at ASC, id ASC`)).
		WithArgs(int64(44)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: 7, country: "", ping: 0},
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 2, accountID: 8, country: "", ping: 0},
		))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	CreateGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusCreated {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var response Game
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(response.Players) != 2 {
		t.Fatalf("got %d players want 2", len(response.Players))
	}
	if response.Players[0].Ping != 0 || response.Players[1].Ping != 0 {
		t.Fatalf("expected default ping 0 in response, got %+v", response.Players)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetGame_AllowsJoinedPlayerAndIncludesPingValues(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectGameSessionLookup(mock, 88, 8)
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, true, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 AND account_id = $2`)).
		WithArgs(int64(44), int64(8)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 2, accountID: 8, country: "Romans", ping: 140},
		))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 ORDER BY joined_at ASC, id ASC`)).
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
			}{id: 2, accountID: 8, country: "Romans", ping: 140},
		))

	req := httptest.NewRequest(http.MethodGet, "/game/get_game?game_id=44&session_id=88", nil)
	rr := httptest.NewRecorder()

	GetGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}

	var response Game
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(response.Players) != 2 {
		t.Fatalf("got %d players want 2", len(response.Players))
	}
	if response.Players[1].Country != "Romans" || response.Players[1].Ping != 140 {
		t.Fatalf("unexpected player payload: %+v", response.Players[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestKickPlayer_RemovesPlayerAndKickedPlayerCanNoLongerReadGame(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 AND account_id = $2`)).
		WithArgs(int64(44), int64(8)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 2, accountID: 8, country: "Romans", ping: 0},
		))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM game_players WHERE game_id = $1 AND account_id = $2`)).
		WithArgs(int64(44), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 ORDER BY joined_at ASC, id ASC`)).
		WithArgs(int64(44)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: 7, country: "Greeks", ping: 0},
		))
	mock.ExpectCommit()

	expectGameSessionLookup(mock, 88, 8)
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 AND account_id = $2`)).
		WithArgs(int64(44), int64(8)).
		WillReturnError(sql.ErrNoRows)

	kickReq := httptest.NewRequest(http.MethodPost, "/game/kick_player", strings.NewReader(`{"game_id":44,"session_id":55,"account_id":8}`))
	kickRR := httptest.NewRecorder()
	KickGamePlayerHandler(kickRR, kickReq, sqlxDB, store)

	if kickRR.Code != http.StatusOK {
		t.Fatalf("kick got status %d want %d", kickRR.Code, http.StatusOK)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/game/get_game?game_id=44&session_id=88", nil)
	getRR := httptest.NewRecorder()
	GetGameHandler(getRR, getReq, sqlxDB, store)

	if getRR.Code != http.StatusForbidden {
		t.Fatalf("get got status %d want %d", getRR.Code, http.StatusForbidden)
	}
	if strings.TrimSpace(getRR.Body.String()) != ErrorGamePlayerRequired {
		t.Fatalf("got body %q want %q", getRR.Body.String(), ErrorGamePlayerRequired)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestSelectCountry_RejectsDuplicateAndStartedGameChanges(t *testing.T) {
	t.Run("duplicate country", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectGameSessionLookup(mock, 88, 8)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
			WithArgs(int64(44)).
			WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 AND account_id = $2`)).
			WithArgs(int64(44), int64(8)).
			WillReturnRows(gamePlayerRows(44,
				struct {
					id        int64
					accountID int64
					country   string
					ping      int
				}{id: 2, accountID: 8, country: "", ping: 0},
			))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM game_players WHERE game_id = $1 AND account_id <> $2 AND lower(country) = lower($3)`)).
			WithArgs(int64(44), int64(8), "Romans").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
		mock.ExpectRollback()

		req := httptest.NewRequest(http.MethodPost, "/game/select_country", strings.NewReader(`{"game_id":44,"session_id":88,"country":"Romans"}`))
		rr := httptest.NewRecorder()
		SelectGameCountryHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusConflict {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusConflict)
		}
		if strings.TrimSpace(rr.Body.String()) != ErrorCountryAlreadyTaken {
			t.Fatalf("got body %q want %q", rr.Body.String(), ErrorCountryAlreadyTaken)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("started game", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)
		startedAt := time.Now()

		expectGameSessionLookup(mock, 88, 8)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
			WithArgs(int64(44)).
			WillReturnRows(gameRowWithLobby(44, 7, nil, "started", "Task 11 Game", &startedAt, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
		mock.ExpectRollback()

		req := httptest.NewRequest(http.MethodPost, "/game/select_country", strings.NewReader(`{"game_id":44,"session_id":88,"country":"Romans"}`))
		rr := httptest.NewRecorder()
		SelectGameCountryHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusConflict {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusConflict)
		}
		if strings.TrimSpace(rr.Body.String()) != "game 44 is not in setup" {
			t.Fatalf("got body %q", rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})
}

func TestUpdateStartingResources_RequireOwnerAndNonNegativeValues(t *testing.T) {
	t.Run("non-owner cannot update gold", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectGameSessionLookup(mock, 88, 8)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
			WithArgs(int64(44)).
			WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
		mock.ExpectRollback()

		req := httptest.NewRequest(http.MethodPost, "/game/update_starting_gold", strings.NewReader(`{"game_id":44,"session_id":88,"starting_gold":250}`))
		rr := httptest.NewRecorder()
		UpdateGameStartingGoldHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusForbidden)
		}
		if strings.TrimSpace(rr.Body.String()) != ErrorGameOwnerRequired {
			t.Fatalf("got body %q want %q", rr.Body.String(), ErrorGameOwnerRequired)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("negative settlers rejected", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectGameSessionLookup(mock, 55, 7)

		req := httptest.NewRequest(http.MethodPost, "/game/update_starting_settlers", strings.NewReader(`{"game_id":44,"session_id":55,"starting_settlers":-1}`))
		rr := httptest.NewRecorder()
		UpdateGameStartingSettlersHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
		}
		if strings.TrimSpace(rr.Body.String()) != ErrorNegativeSettlers {
			t.Fatalf("got body %q want %q", rr.Body.String(), ErrorNegativeSettlers)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected sql execution: %v", err)
		}
	})
}

func TestUpdatePing_PersistsNonNegativeValue(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectGameSessionLookup(mock, 88, 8)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 AND account_id = $2`)).
		WithArgs(int64(44), int64(8)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 2, accountID: 8, country: "Romans", ping: 0},
		))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE game_players SET ping_ms = $1, updated_at = NOW() WHERE game_id = $2 AND account_id = $3`)).
		WithArgs(140, int64(44), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, DefaultStartingSettlers, DefaultStartingGold))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms FROM game_players WHERE game_id = $1 ORDER BY joined_at ASC, id ASC`)).
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
			}{id: 2, accountID: 8, country: "Romans", ping: 140},
		))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/game/update_ping", strings.NewReader(`{"game_id":44,"session_id":88,"ping":140}`))
	rr := httptest.NewRecorder()
	UpdateGamePingHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}

	var response Game
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Players[1].Ping != 140 {
		t.Fatalf("got ping %d want 140", response.Players[1].Ping)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestStartGame_StartsSetupGameAndBlocksLaterSetupMutations(t *testing.T) {
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
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRowWithLobby(44, 7, &lobbyID, "started", "Task 11 Game", &startedAt, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	mock.ExpectRollback()

	startReq := httptest.NewRequest(http.MethodPost, "/game/start_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	startRR := httptest.NewRecorder()
	StartGameHandler(startRR, startReq, sqlxDB, store)

	if startRR.Code != http.StatusOK {
		t.Fatalf("start got status %d want %d", startRR.Code, http.StatusOK)
	}

	var started Game
	if err := json.Unmarshal(startRR.Body.Bytes(), &started); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if started.Status != "started" {
		t.Fatalf("got status %q want started", started.Status)
	}

	goldReq := httptest.NewRequest(http.MethodPost, "/game/update_starting_gold", strings.NewReader(`{"game_id":44,"session_id":55,"starting_gold":300}`))
	goldRR := httptest.NewRecorder()
	UpdateGameStartingGoldHandler(goldRR, goldReq, sqlxDB, store)

	if goldRR.Code != http.StatusConflict {
		t.Fatalf("gold update got status %d want %d", goldRR.Code, http.StatusConflict)
	}
	if strings.TrimSpace(goldRR.Body.String()) != "game 44 is not in setup" {
		t.Fatalf("got body %q", goldRR.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
