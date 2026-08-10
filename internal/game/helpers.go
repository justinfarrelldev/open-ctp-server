package game

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	ErrorGameOwnerRequired   = "only the game owner may perform this action"
	ErrorGamePlayerRequired  = "only joined game players may perform this action"
	ErrorCountryAlreadyTaken = "country has already been selected by another player"
)

const gameSelectByID = `SELECT id, created_at, updated_at, started_at, lobby_id, owner_account_id, name, password_protected, password, is_locked, is_public, style, rules_json, exclusions_json, status, starting_settlers, starting_gold, map_size, climate, wetness
FROM games
WHERE id = $1`

const gameSelectByIDForUpdate = `SELECT id, created_at, updated_at, started_at, lobby_id, owner_account_id, name, password_protected, password, is_locked, is_public, style, rules_json, exclusions_json, status, starting_settlers, starting_gold, map_size, climate, wetness
FROM games
WHERE id = $1
FOR UPDATE`

const gamePlayersSelectByGameID = `SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms, controller_kind FROM game_players WHERE game_id = $1 ORDER BY joined_at ASC, id ASC`

const gamePlayerByAccountQuery = `SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms, controller_kind FROM game_players WHERE game_id = $1 AND account_id = $2`

const lobbyByIDForGameQuery = `SELECT id, closed_at FROM lobbies WHERE id = $1`

const lobbyMembersForGameQuery = `SELECT account_id FROM lobby_members WHERE lobby_id = $1 ORDER BY joined_at ASC, id ASC`

type gameGetter interface {
	Get(dest any, query string, args ...any) error
}

type gameSelector interface {
	Select(dest any, query string, args ...any) error
}

type gameReader interface {
	gameGetter
	gameSelector
}

type gamePatch struct {
	Name              *string
	IsLocked          *bool
	IsPublic          *bool
	Style             *string
	MapSize           *string
	Climate           *string
	Wetness           *int
	RulesJSON         json.RawMessage
	HasRulesJSON      bool
	ExclusionsJSON    json.RawMessage
	HasExclusionsJSON bool
}

type lobbyDependency struct {
	ID       int64      `db:"id"`
	ClosedAt *time.Time `db:"closed_at"`
}

func (p gamePatch) hasAnyField() bool {
	return p.Name != nil || p.IsLocked != nil || p.IsPublic != nil || p.Style != nil || p.MapSize != nil || p.Climate != nil || p.Wetness != nil || p.HasRulesJSON || p.HasExclusionsJSON
}

func beginGameTx(db *sqlx.DB) (*sqlx.Tx, error) {
	return db.Beginx()
}

func decodeGameJSONBody(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func requireGameSession(w http.ResponseWriter, store *auth.SessionStore, body auth.SessionIDBody) (*auth.Session, error) {
	return auth.RequireSessionFromJSONBody(w, store, body)
}

func requireGameQuerySession(w http.ResponseWriter, r *http.Request, store *auth.SessionStore) (*auth.Session, error) {
	return auth.RequireSessionFromQuery(w, r, store)
}

func getGameByID(executor gameGetter, gameID int64) (*Game, error) {
	var game Game
	if err := executor.Get(&game, gameSelectByID, gameID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("an error occurred while getting the game with the ID %d: %v", gameID, err)
	}
	return &game, nil
}

func getGameByIDForUpdate(executor gameGetter, gameID int64) (*Game, error) {
	var game Game
	if err := executor.Get(&game, gameSelectByIDForUpdate, gameID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("an error occurred while locking the game with the ID %d: %v", gameID, err)
	}
	return &game, nil
}

func listGamePlayersByGameID(executor gameSelector, gameID int64) ([]GamePlayer, error) {
	players := make([]GamePlayer, 0)
	if err := executor.Select(&players, gamePlayersSelectByGameID, gameID); err != nil {
		return nil, fmt.Errorf("an error occurred while getting game players for the game with the ID %d: %v", gameID, err)
	}
	return players, nil
}

func getGamePlayerByAccount(executor gameGetter, gameID int64, accountID int64) (*GamePlayer, error) {
	var player GamePlayer
	if err := executor.Get(&player, gamePlayerByAccountQuery, gameID, accountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("an error occurred while getting the game player for the game with the ID %d and account with the ID %d: %v", gameID, accountID, err)
	}
	return &player, nil
}

func getGameWithPlayers(executor gameReader, gameID int64) (*Game, error) {
	game, err := getGameByID(executor, gameID)
	if err != nil {
		return nil, err
	}

	players, err := listGamePlayersByGameID(executor, gameID)
	if err != nil {
		return nil, err
	}
	game.Players = players
	return game, nil
}

func ensureGameOwner(w http.ResponseWriter, session *auth.Session, game *Game) error {
	if game.OwnerAccountID == nil || int64(session.AccountID) != *game.OwnerAccountID {
		return writeGameStatusError(w, http.StatusForbidden, ErrorGameOwnerRequired)
	}
	return nil
}

func ensureGamePlayer(w http.ResponseWriter, executor gameGetter, gameID int64, accountID int64) (*GamePlayer, error) {
	player, err := getGamePlayerByAccount(executor, gameID, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, writeGameStatusError(w, http.StatusForbidden, ErrorGamePlayerRequired)
		}
		return nil, err
	}
	return player, nil
}

func ensureGameSetupMutable(w http.ResponseWriter, game *Game) error {
	if game.Status != DefaultGameStatus {
		return writeGameStatusError(w, http.StatusConflict, fmt.Sprintf("game %d is not in setup", game.ID))
	}
	return nil
}

func writeGameJSONResponse(w http.ResponseWriter, status int, payload any) error {
	responseBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding the response body: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(responseBytes)
	return nil
}

func parseGameIDQuery(r *http.Request) (int64, error) {
	gameIDStr := r.URL.Query().Get("game_id")
	if strings.TrimSpace(gameIDStr) == "" {
		return 0, errors.New("game_id is required")
	}

	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil || gameID <= 0 {
		return 0, errors.New("invalid game_id")
	}

	return gameID, nil
}

func buildGamePatchUpdate(patch gamePatch) (string, []any, error) {
	fields := make([]string, 0, 7)
	params := make([]any, 0, 6)
	paramIndex := 1

	if patch.Name != nil {
		trimmed := strings.TrimSpace(*patch.Name)
		if trimmed == "" {
			return "", nil, errors.New("name must be specified")
		}
		fields = append(fields, fmt.Sprintf("name = $%d", paramIndex))
		params = append(params, trimmed)
		paramIndex++
	}
	if patch.IsLocked != nil {
		fields = append(fields, fmt.Sprintf("is_locked = $%d", paramIndex))
		params = append(params, *patch.IsLocked)
		paramIndex++
	}
	if patch.IsPublic != nil {
		fields = append(fields, fmt.Sprintf("is_public = $%d", paramIndex))
		params = append(params, *patch.IsPublic)
		paramIndex++
	}
	if patch.Style != nil {
		trimmed := strings.TrimSpace(*patch.Style)
		if trimmed == "" {
			return "", nil, errors.New("style must be specified")
		}
		if !supportedGameStyle(trimmed) {
			return "", nil, fmt.Errorf("unsupported game style %q", trimmed)
		}
		fields = append(fields, fmt.Sprintf("style = $%d", paramIndex))
		params = append(params, trimmed)
		paramIndex++
	}
	if patch.MapSize != nil {
		value := strings.TrimSpace(*patch.MapSize)
		if value != "duel" && value != "small" && value != "medium" && value != "large" && value != "huge" {
			return "", nil, errors.New("unsupported map_size")
		}
		fields = append(fields, fmt.Sprintf("map_size = $%d", paramIndex))
		params = append(params, value)
		paramIndex++
	}
	if patch.Climate != nil {
		value := strings.TrimSpace(*patch.Climate)
		if value != "arid" && value != "cold" && value != "temperate" && value != "tropical" {
			return "", nil, errors.New("unsupported climate")
		}
		fields = append(fields, fmt.Sprintf("climate = $%d", paramIndex))
		params = append(params, value)
		paramIndex++
	}
	if patch.Wetness != nil {
		if *patch.Wetness < 0 || *patch.Wetness > 100 {
			return "", nil, errors.New("wetness must be between 0 and 100")
		}
		fields = append(fields, fmt.Sprintf("wetness = $%d", paramIndex))
		params = append(params, *patch.Wetness)
		paramIndex++
	}
	if patch.HasRulesJSON {
		normalized, err := normalizeJSONArray(patch.RulesJSON, []byte(defaultRulesJSON))
		if err != nil {
			return "", nil, errors.New(ErrorRulesJSONInvalid)
		}
		fields = append(fields, fmt.Sprintf("rules_json = $%d", paramIndex))
		params = append(params, normalized)
		paramIndex++
	}
	if patch.HasExclusionsJSON {
		normalized, err := normalizeJSONObject(patch.ExclusionsJSON, []byte(defaultExclusionsJSON))
		if err != nil {
			return "", nil, errors.New(ErrorExclusionsJSONInvalid)
		}
		fields = append(fields, fmt.Sprintf("exclusions_json = $%d", paramIndex))
		params = append(params, normalized)
		paramIndex++
	}

	if len(fields) == 0 {
		return "", nil, errors.New("at least one field to update must be specified")
	}

	fields = append(fields, "updated_at = NOW()")
	return strings.Join(fields, ", "), params, nil
}

func supportedGameStyle(style string) bool {
	for _, mode := range rules.SupportedManifest().Modes {
		if style == mode {
			return true
		}
	}
	return false
}

func validateLobbyDependency(executor gameGetter, lobbyID int64) error {
	var lobby lobbyDependency
	if err := executor.Get(&lobby, lobbyByIDForGameQuery, lobbyID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("lobby %d is unavailable", lobbyID)
		}
		return fmt.Errorf("an error occurred while getting the lobby with the ID %d: %v", lobbyID, err)
	}
	if lobby.ClosedAt != nil {
		return fmt.Errorf("lobby %d is unavailable", lobbyID)
	}
	return nil
}

func listLobbyAccountIDs(executor gameSelector, lobbyID int64) ([]int64, error) {
	rows := make([]struct {
		AccountID int64 `db:"account_id"`
	}, 0)
	if err := executor.Select(&rows, lobbyMembersForGameQuery, lobbyID); err != nil {
		return nil, fmt.Errorf("an error occurred while getting lobby members for the lobby with the ID %d: %v", lobbyID, err)
	}

	accountIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		accountIDs = append(accountIDs, row.AccountID)
	}
	return accountIDs, nil
}

func insertGamePlayer(tx *sqlx.Tx, gameID int64, accountID int64) error {
	if _, err := tx.Exec(`INSERT INTO game_players (game_id, account_id) VALUES ($1, $2)`, gameID, accountID); err != nil {
		return fmt.Errorf("an error occurred while storing the game player in the database: %v", err)
	}
	return nil
}

func resolveGamePlayerAccountIDs(executor gameReader, ownerAccountID int64, lobbyID *int64) ([]int64, error) {
	accountIDs := []int64{ownerAccountID}
	if lobbyID != nil {
		if err := validateLobbyDependency(executor, *lobbyID); err != nil {
			return nil, err
		}
		lobbyAccountIDs, err := listLobbyAccountIDs(executor, *lobbyID)
		if err != nil {
			return nil, err
		}
		if len(lobbyAccountIDs) > 0 {
			accountIDs = lobbyAccountIDs
		}
		ownerPresent := false
		for _, accountID := range accountIDs {
			if accountID == ownerAccountID {
				ownerPresent = true
				break
			}
		}
		if !ownerPresent {
			accountIDs = append([]int64{ownerAccountID}, accountIDs...)
		}
	}

	seen := make(map[int64]struct{}, len(accountIDs))
	uniqueAccountIDs := make([]int64, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		if _, ok := seen[accountID]; ok {
			continue
		}
		seen[accountID] = struct{}{}
		uniqueAccountIDs = append(uniqueAccountIDs, accountID)
	}
	return uniqueAccountIDs, nil
}

func seedGamePlayers(tx *sqlx.Tx, gameID int64, accountIDs []int64) error {
	for _, accountID := range accountIDs {
		if err := insertGamePlayer(tx, gameID, accountID); err != nil {
			return err
		}
	}
	return nil
}

func applyGamePatch(w http.ResponseWriter, db *sqlx.DB, store *auth.SessionStore, sessionBody auth.SessionIDBody, gameID int64, patch gamePatch) error {
	session, err := requireGameSession(w, store, sessionBody)
	if err != nil {
		return err
	}

	setClause, params, err := buildGamePatchUpdate(patch)
	if err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, err.Error())
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game update: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	game, err := getGameByIDForUpdate(tx, gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", gameID))
		}
		return err
	}

	if err := ensureGameOwner(w, session, game); err != nil {
		return err
	}
	if err := ensureGameSetupMutable(w, game); err != nil {
		return err
	}

	params = append(params, gameID)
	query := fmt.Sprintf("UPDATE games SET %s WHERE id = $%d", setClause, len(params))
	if _, err := tx.Exec(query, params...); err != nil {
		return fmt.Errorf("an error occurred while updating the game with the ID %d: %v", gameID, err)
	}

	updatedGame, err := getGameWithPlayers(tx, gameID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game update: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, updatedGame)
}
