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
	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func newGameSQLMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return sqlx.NewDb(db, "sqlmock"), mock
}

func newGameSessionStore(db *sqlx.DB) *auth.SessionStore {
	return auth.NewSessionStore(db)
}

func expectGameSessionLookup(mock sqlmock.Sqlmock, sessionID int64, accountID int) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM sessions WHERE id = $1`)).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "created_at", "expires_at"}).AddRow(sessionID, accountID, time.Now(), time.Now().Add(time.Hour)))
}

func gameRow(gameID int64, ownerAccountID int64, status string, name string, isLocked bool, isPublic bool, style string, rulesJSON string, exclusionsJSON string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "created_at", "updated_at", "started_at", "lobby_id", "owner_account_id", "name", "password_protected", "password", "is_locked", "is_public", "style", "rules_json", "exclusions_json", "status", "starting_settlers", "starting_gold"}).
		AddRow(gameID, time.Now(), time.Now(), nil, nil, ownerAccountID, name, false, "", isLocked, isPublic, style, rulesJSON, exclusionsJSON, status, DefaultStartingSettlers, DefaultStartingGold)
}

func TestCreateGame_UsesAuthenticatedSessionOwnerAndReturnsCreatedGame(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	body := map[string]any{
		"session_id":         55,
		"name":               "ctp2 default setup",
		"owner_account_id":   999,
		"password_protected": false,
		"password":           "",
	}
	jsonBody, _ := json.Marshal(body)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(insertGameQuery)).
		WithArgs(
			"ctp2 default setup",
			int64(7),
			nil,
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
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRow(44, 7, DefaultGameStatus, "ctp2 default setup", false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON))
	mock.ExpectQuery(regexp.QuoteMeta(gamePlayersSelectByGameID)).
		WithArgs(int64(44)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: 7, country: "", ping: 0},
		))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	CreateGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusCreated {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusCreated)
	}

	var response Game
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.OwnerAccountID == nil || *response.OwnerAccountID != 7 {
		t.Fatalf("got owner_account_id %v want 7", response.OwnerAccountID)
	}
	if response.Style != DefaultGameStyle || response.RulesJSON != defaultRulesJSON || response.ExclusionsJSON != defaultExclusionsJSON {
		t.Fatalf("unexpected persisted defaults in response: %+v", response)
	}
	if len(response.Players) != 1 || response.Players[0].Ping != 0 {
		t.Fatalf("unexpected players in response: %+v", response.Players)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCreateGame_RejectsNegativeStartingSettlers(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	settlers := -1
	body := CreateGameArgs{SessionId: int64Ptr(55), StartingSettlers: &settlers}
	jsonBody, _ := json.Marshal(body)

	expectGameSessionLookup(mock, 55, 7)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	CreateGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if strings.TrimSpace(rr.Body.String()) != ErrorNegativeSettlers {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorNegativeSettlers)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsNegativeStartingGold(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	gold := -1
	body := CreateGameArgs{SessionId: int64Ptr(55), StartingGold: &gold}
	jsonBody, _ := json.Marshal(body)

	expectGameSessionLookup(mock, 55, 7)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	CreateGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if strings.TrimSpace(rr.Body.String()) != ErrorNegativeGold {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorNegativeGold)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsInvalidRulesJSONValue(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	jsonBody := []byte(`{"session_id":55,"rules_json":"not-json-array"}`)

	expectGameSessionLookup(mock, 55, 7)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	CreateGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if strings.TrimSpace(rr.Body.String()) != ErrorRulesJSONInvalid {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorRulesJSONInvalid)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsInvalidExclusionsJSONValue(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	jsonBody := []byte(`{"session_id":55,"exclusions_json":"not-json-object"}`)

	expectGameSessionLookup(mock, 55, 7)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	CreateGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if strings.TrimSpace(rr.Body.String()) != ErrorExclusionsJSONInvalid {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorExclusionsJSONInvalid)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestGameProtectedEndpoints_RejectMissingOrInvalidSessions(t *testing.T) {
	tests := []struct {
		name       string
		request    *http.Request
		invoke     func(http.ResponseWriter, *http.Request, *sqlx.DB, *auth.SessionStore)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "create missing session",
			request:    httptest.NewRequest(http.MethodPost, "/game/create_game", strings.NewReader(`{"name":"test"}`)),
			invoke:     CreateGameHandler,
			wantStatus: http.StatusUnauthorized,
			wantBody:   auth.ErrInvalidBodySessionID.Error(),
		},
		{
			name:       "get invalid session query",
			request:    httptest.NewRequest(http.MethodGet, "/game/get_game?game_id=1&session_id=invalid", nil),
			invoke:     GetGameHandler,
			wantStatus: http.StatusUnauthorized,
			wantBody:   auth.ErrInvalidSessionID.Error(),
		},
		{
			name:       "rename missing session",
			request:    httptest.NewRequest(http.MethodPost, "/game/rename_game", strings.NewReader(`{"game_id":1,"name":"new"}`)),
			invoke:     RenameGameHandler,
			wantStatus: http.StatusUnauthorized,
			wantBody:   auth.ErrInvalidBodySessionID.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlxDB, mock := newGameSQLMock(t)
			store := newGameSessionStore(sqlxDB)
			rr := httptest.NewRecorder()

			tt.invoke(rr, tt.request, sqlxDB, store)

			if rr.Code != tt.wantStatus {
				t.Fatalf("got status %d want %d", rr.Code, tt.wantStatus)
			}
			if strings.TrimSpace(rr.Body.String()) != tt.wantBody {
				t.Fatalf("got body %q want %q", rr.Body.String(), tt.wantBody)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unexpected sql execution: %v", err)
			}
		})
	}
}

func TestGetGame_ReturnsPersistedFieldsForOwner(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRow(44, 7, DefaultGameStatus, "Persisted", true, true, "scenario", `["fast"]`, `{"units":["settler"]}`))
	mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).
		WithArgs(int64(44), int64(7)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: 7, country: "Greeks", ping: 0},
		))
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

	req := httptest.NewRequest(http.MethodGet, "/game/get_game?game_id=44&session_id=55", nil)
	rr := httptest.NewRecorder()

	GetGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}

	var response Game
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !response.IsLocked || !response.IsPublic || response.Style != "scenario" || response.RulesJSON != `["fast"]` || response.ExclusionsJSON != `{"units":["settler"]}` {
		t.Fatalf("unexpected response: %+v", response)
	}
	if len(response.Players) != 1 || response.Players[0].Country != "Greeks" || response.Players[0].Ping != 0 {
		t.Fatalf("unexpected players response: %+v", response.Players)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUpdateGame_PersistsAndReturnsUpdatedSetupFields(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	body := `{"game_id":44,"session_id":55,"game":{"name":"Renamed","is_locked":true,"is_public":true,"style":"scenario","rules_json":["fast"],"exclusions_json":{"units":["settler"]}}}`

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRow(44, 7, DefaultGameStatus, "Original", false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE games SET name = $1, is_locked = $2, is_public = $3, style = $4, rules_json = $5, exclusions_json = $6, updated_at = NOW() WHERE id = $7`)).
		WithArgs("Renamed", true, true, "scenario", `["fast"]`, `{"units":["settler"]}`, int64(44)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRow(44, 7, DefaultGameStatus, "Renamed", true, true, "scenario", `["fast"]`, `{"units":["settler"]}`))
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
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPut, "/game/update_game", strings.NewReader(body))
	rr := httptest.NewRecorder()

	UpdateGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusOK)
	}

	var response Game
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Name != "Renamed" || !response.IsLocked || !response.IsPublic || response.Style != "scenario" || response.RulesJSON != `["fast"]` || response.ExclusionsJSON != `{"units":["settler"]}` {
		t.Fatalf("unexpected updated response: %+v", response)
	}
	if len(response.Players) != 1 || response.Players[0].Country != "Greeks" {
		t.Fatalf("unexpected players response: %+v", response.Players)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestDeleteGame_RemovesGameAndFutureGetReturnsNotFound(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRow(44, 7, DefaultGameStatus, "Delete Me", false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM games WHERE id = $1`)).
		WithArgs(int64(44)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnError(sql.ErrNoRows)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/game/delete_game", strings.NewReader(`{"game_id":44,"session_id":55}`))
	deleteRR := httptest.NewRecorder()
	DeleteGameHandler(deleteRR, deleteReq, sqlxDB, store)

	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete got status %d want %d", deleteRR.Code, http.StatusOK)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/game/get_game?game_id=44&session_id=55", nil)
	getRR := httptest.NewRecorder()
	GetGameHandler(getRR, getReq, sqlxDB, store)

	if getRR.Code != http.StatusNotFound {
		t.Fatalf("get got status %d want %d", getRR.Code, http.StatusNotFound)
	}
	if strings.TrimSpace(getRR.Body.String()) != "no game exists with the ID 44" {
		t.Fatalf("got body %q", getRR.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGameMutations_RejectNonOwners(t *testing.T) {
	tests := []struct {
		name    string
		request *http.Request
		invoke  func(http.ResponseWriter, *http.Request, *sqlx.DB, *auth.SessionStore)
	}{
		{name: "rename", request: httptest.NewRequest(http.MethodPost, "/game/rename_game", strings.NewReader(`{"game_id":44,"session_id":55,"name":"new"}`)), invoke: RenameGameHandler},
		{name: "delete", request: httptest.NewRequest(http.MethodDelete, "/game/delete_game", strings.NewReader(`{"game_id":44,"session_id":55}`)), invoke: DeleteGameHandler},
		{name: "style", request: httptest.NewRequest(http.MethodPost, "/game/update_style", strings.NewReader(`{"game_id":44,"session_id":55,"style":"scenario"}`)), invoke: UpdateGameStyleHandler},
		{name: "rules", request: httptest.NewRequest(http.MethodPost, "/game/update_rules", strings.NewReader(`{"game_id":44,"session_id":55,"rules_json":["fast"]}`)), invoke: UpdateGameRulesHandler},
		{name: "exclusions", request: httptest.NewRequest(http.MethodPost, "/game/update_exclusions", strings.NewReader(`{"game_id":44,"session_id":55,"exclusions_json":{"units":["settler"]}}`)), invoke: UpdateGameExclusionsHandler},
		{name: "public", request: httptest.NewRequest(http.MethodPost, "/game/set_public", strings.NewReader(`{"game_id":44,"session_id":55,"is_public":true}`)), invoke: SetGamePublicHandler},
		{name: "lock", request: httptest.NewRequest(http.MethodPost, "/game/lock_game", strings.NewReader(`{"game_id":44,"session_id":55,"is_locked":true}`)), invoke: LockGameHandler},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlxDB, mock := newGameSQLMock(t)
			store := newGameSessionStore(sqlxDB)

			expectGameSessionLookup(mock, 55, 7)
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
				WithArgs(int64(44)).
				WillReturnRows(gameRow(44, 8, DefaultGameStatus, "Owned by someone else", false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON))
			mock.ExpectRollback()

			rr := httptest.NewRecorder()
			tt.invoke(rr, tt.request, sqlxDB, store)

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
	}
}

func TestGameMutations_RejectNonSetupGames(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectGameSessionLookup(mock, 55, 7)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRow(44, 7, "active", "No Longer Mutable", false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON))
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/game/rename_game", strings.NewReader(`{"game_id":44,"session_id":55,"name":"new"}`))
	rr := httptest.NewRecorder()

	RenameGameHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusConflict {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusConflict)
	}
	if strings.TrimSpace(rr.Body.String()) != "game 44 is not in setup" {
		t.Fatalf("got body %q", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAddAI_AlwaysReturns404WithExactMessage(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/game/add_ai", strings.NewReader(`{"game_id":44}`))
	rr := httptest.NewRecorder()

	AddAIHandler(rr, req, nil, nil)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusNotFound)
	}
	if strings.TrimSpace(rr.Body.String()) != "AI players are not supported yet" {
		t.Fatalf("got body %q", rr.Body.String())
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
