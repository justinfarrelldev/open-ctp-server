package game

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type getGameQueryArgs struct {
	GameID int64
}

type UpdateGameArgs struct {
	GameId    *int64          `json:"game_id"`
	SessionId *int64          `json:"session_id"`
	Game      *GameUpdateArgs `json:"game"`
}

func (a UpdateGameArgs) GetSessionID() *int64 {
	return a.SessionId
}

type GameUpdateArgs struct {
	Name           *string         `json:"name,omitempty"`
	IsLocked       *bool           `json:"is_locked,omitempty"`
	IsPublic       *bool           `json:"is_public,omitempty"`
	Style          *string         `json:"style,omitempty"`
	RulesJSON      json.RawMessage `json:"rules_json,omitempty"`
	ExclusionsJSON json.RawMessage `json:"exclusions_json,omitempty"`
}

type DeleteGameArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
}

func (a DeleteGameArgs) GetSessionID() *int64 {
	return a.SessionId
}

type RenameGameArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
	Name      string `json:"name"`
}

func (a RenameGameArgs) GetSessionID() *int64 {
	return a.SessionId
}

type LockGameArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
	IsLocked  *bool  `json:"is_locked"`
}

func (a LockGameArgs) GetSessionID() *int64 {
	return a.SessionId
}

type SetPublicGameArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
	IsPublic  *bool  `json:"is_public"`
}

func (a SetPublicGameArgs) GetSessionID() *int64 {
	return a.SessionId
}

type UpdateStyleArgs struct {
	GameId    int64   `json:"game_id"`
	SessionId *int64  `json:"session_id"`
	Style     string  `json:"style"`
	MapSize   *string `json:"map_size,omitempty"`
	Climate   *string `json:"climate,omitempty"`
	Wetness   *int    `json:"wetness,omitempty"`
}

func (a UpdateStyleArgs) GetSessionID() *int64 {
	return a.SessionId
}

type UpdateRulesArgs struct {
	GameId    int64           `json:"game_id"`
	SessionId *int64          `json:"session_id"`
	RulesJSON json.RawMessage `json:"rules_json"`
}

func (a UpdateRulesArgs) GetSessionID() *int64 {
	return a.SessionId
}

type UpdateExclusionsArgs struct {
	GameId         int64           `json:"game_id"`
	SessionId      *int64          `json:"session_id"`
	ExclusionsJSON json.RawMessage `json:"exclusions_json"`
}

type AddAIArgs struct {
	GameID          int64  `json:"game_id"`
	SessionID       *int64 `json:"session_id"`
	TargetAccountID int64  `json:"target_account_id,omitempty"`
	Country         string `json:"country,omitempty"`
	Name            string `json:"name,omitempty"`
}

func (a AddAIArgs) GetSessionID() *int64 { return a.SessionID }

func (a UpdateExclusionsArgs) GetSessionID() *int64 {
	return a.SessionId
}

// GetGame returns the current persisted game for an authenticated joined player.
//
// @Summary Get a game
// @Description This endpoint returns a game's persisted setup fields for any authenticated account that is still joined to the game, including player country and ping values.
// @Tags game
// @Accept json
// @Produce json
// @Param game_id query int true "game ID"
// @Param session_id query int true "session ID"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/get_game [get]
func GetGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	gameID, err := parseGameIDQuery(r)
	if err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, err.Error())
	}

	session, err := requireGameQuerySession(w, r, store)
	if err != nil {
		return err
	}

	game, err := getGameByID(db, gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", gameID))
		}
		return err
	}

	if _, err := ensureGamePlayer(w, db, gameID, int64(session.AccountID)); err != nil {
		return err
	}

	players, err := listGamePlayersByGameID(db, gameID)
	if err != nil {
		return err
	}
	game.Players = players

	return writeGameJSONResponse(w, http.StatusOK, game)
}

// UpdateGame updates mutable setup fields for the authenticated owner.
//
// @Summary Update a game
// @Description This endpoint updates mutable setup-stage game fields for the authenticated game owner and returns the persisted game state.
// @Tags game
// @Accept json
// @Produce json
// @Param body body UpdateGameDocsArgs true "game update request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/update_game [put]
func UpdateGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPut {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a PUT request")
	}

	args := UpdateGameArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == nil {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}
	if args.Game == nil {
		return writeGameStatusError(w, http.StatusBadRequest, "game must be specified")
	}

	patch := gamePatch{
		Name:              args.Game.Name,
		IsLocked:          args.Game.IsLocked,
		IsPublic:          args.Game.IsPublic,
		Style:             args.Game.Style,
		RulesJSON:         args.Game.RulesJSON,
		HasRulesJSON:      len(args.Game.RulesJSON) > 0,
		ExclusionsJSON:    args.Game.ExclusionsJSON,
		HasExclusionsJSON: len(args.Game.ExclusionsJSON) > 0,
	}
	if !patch.hasAnyField() {
		return writeGameStatusError(w, http.StatusBadRequest, "at least one field to update must be specified")
	}

	return applyGamePatch(w, db, store, args, *args.GameId, patch)
}

// DeleteGame deletes a setup-stage game for the authenticated owner.
//
// @Summary Delete a game
// @Description This endpoint deletes a setup-stage game only for the authenticated game owner.
// @Tags game
// @Accept json
// @Produce json
// @Param body body DeleteGameArgs true "game deletion request body"
// @Success 200 {string} string "Successfully deleted game!"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/delete_game [delete]
func DeleteGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodDelete {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a DELETE request")
	}

	args := DeleteGameArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body: "+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	session, err := requireGameSession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting game deletion: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	game, err := getGameByIDForUpdate(tx, args.GameId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", args.GameId))
		}
		return err
	}

	if err := ensureGameOwner(w, session, game); err != nil {
		return err
	}
	if err := ensureGameSetupMutable(w, game); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM games WHERE id = $1", args.GameId)
	if err != nil {
		return fmt.Errorf("an error occurred while deleting the game with the ID %d: %v", args.GameId, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while checking the affected rows: %v", err)
	}
	if rowsAffected == 0 {
		return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", args.GameId))
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game deletion: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, "Successfully deleted game!")
}

// RenameGame renames a setup-stage game for the authenticated owner.
//
// @Summary Rename a game
// @Description This endpoint updates a setup-stage game's name only for the authenticated game owner and returns the persisted game state.
// @Tags game
// @Accept json
// @Produce json
// @Param body body RenameGameArgs true "game rename request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/rename_game [post]
func RenameGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := RenameGameArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	return applyGamePatch(w, db, store, args, args.GameId, gamePatch{Name: &args.Name})
}

// LockGame updates the locked flag for a setup-stage game.
//
// @Summary Lock or unlock a game
// @Description This endpoint updates a setup-stage game's locked flag only for the authenticated game owner and returns the persisted game state.
// @Tags game
// @Accept json
// @Produce json
// @Param body body LockGameArgs true "game lock request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/lock_game [post]
func LockGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := LockGameArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}
	if args.IsLocked == nil {
		return writeGameStatusError(w, http.StatusBadRequest, "is_locked must be specified")
	}

	return applyGamePatch(w, db, store, args, args.GameId, gamePatch{IsLocked: args.IsLocked})
}

// SetPublic updates the public flag for a setup-stage game.
//
// @Summary Set game public state
// @Description This endpoint updates a setup-stage game's public flag only for the authenticated game owner and returns the persisted game state.
// @Tags game
// @Accept json
// @Produce json
// @Param body body SetPublicGameArgs true "game public-state request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/set_public [post]
func SetPublic(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := SetPublicGameArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}
	if args.IsPublic == nil {
		return writeGameStatusError(w, http.StatusBadRequest, "is_public must be specified")
	}

	return applyGamePatch(w, db, store, args, args.GameId, gamePatch{IsPublic: args.IsPublic})
}

// UpdateStyle updates the style for a setup-stage game.
//
// @Summary Update game style
// @Description This endpoint updates a setup-stage game's style only for the authenticated game owner and returns the persisted game state.
// @Tags game
// @Accept json
// @Produce json
// @Param body body UpdateStyleArgs true "game style update request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/update_style [post]
func UpdateStyle(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := UpdateStyleArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	return applyGamePatch(w, db, store, args, args.GameId, gamePatch{Style: &args.Style, MapSize: args.MapSize, Climate: args.Climate, Wetness: args.Wetness})
}

// UpdateRules updates the rules JSON for a setup-stage game.
//
// @Summary Update game rules
// @Description This endpoint updates a setup-stage game's rules JSON only for the authenticated game owner and returns the persisted game state.
// @Tags game
// @Accept json
// @Produce json
// @Param body body UpdateRulesDocsArgs true "game rules update request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/update_rules [post]
func UpdateRules(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := UpdateRulesArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	return applyGamePatch(w, db, store, args, args.GameId, gamePatch{RulesJSON: args.RulesJSON, HasRulesJSON: len(args.RulesJSON) > 0})
}

// UpdateExclusions updates the exclusions JSON for a setup-stage game.
//
// @Summary Update game exclusions
// @Description This endpoint updates a setup-stage game's exclusions JSON only for the authenticated game owner and returns the persisted game state.
// @Tags game
// @Accept json
// @Produce json
// @Param body body UpdateExclusionsDocsArgs true "game exclusions update request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/update_exclusions [post]
func UpdateExclusions(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := UpdateExclusionsArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameId == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}

	return applyGamePatch(w, db, store, args, args.GameId, gamePatch{ExclusionsJSON: args.ExclusionsJSON, HasExclusionsJSON: len(args.ExclusionsJSON) > 0})
}

const insertManagedAIAccountQuery = `INSERT INTO account (name, info, location, email, experience_level) VALUES ($1, 'Server-managed AI player', 'local', $2, 0) RETURNING id`

const insertManagedAIPlayerQuery = `INSERT INTO game_players (game_id, account_id, country, controller_kind) VALUES ($1, $2, $3, 'ai')`

// AddAI creates or assigns an AI-controlled setup player.
//
// @Summary Add an AI player
// @Description When target_account_id is supplied, this endpoint converts an existing joined setup player to AI control. When target_account_id is omitted, the owner creates a server-managed AI account/player slot with the supplied country.
// @Tags game
// @Accept json
// @Produce json
// @Param body body AddAIArgs true "AI setup request body"
// @Success 200 {object} Game
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Router /game/add_ai [post]
func AddAI(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}
	args := AddAIArgs{}
	if err := decodeGameJSONBody(r, &args); err != nil {
		return writeGameStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.GameID == 0 {
		return writeGameStatusError(w, http.StatusBadRequest, "game_id must be specified")
	}
	creatingManagedAI := args.TargetAccountID == 0
	country := strings.TrimSpace(args.Country)
	if creatingManagedAI && country == "" {
		return writeGameStatusError(w, http.StatusBadRequest, "country must be specified when creating a server-managed AI player")
	}
	session, err := requireGameSession(w, store, args)
	if err != nil {
		return err
	}
	tx, err := beginGameTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting AI setup: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	game, err := getGameByIDForUpdate(tx, args.GameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", args.GameID))
		}
		return err
	}
	if err := ensureGameSetupMutable(w, game); err != nil {
		return err
	}
	if game.OwnerAccountID == nil || *game.OwnerAccountID != int64(session.AccountID) {
		return writeGameStatusError(w, http.StatusForbidden, "only the game owner can assign AI control")
	}
	if creatingManagedAI {
		if err := createManagedAIPlayer(w, tx, args, country); err != nil {
			return err
		}
	} else if err := assignExistingPlayerAI(w, tx, args); err != nil {
		return err
	}
	updated, err := getGameWithPlayers(tx, args.GameID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing AI setup: %v", err)
	}
	committed = true
	return writeGameJSONResponse(w, http.StatusOK, updated)
}

func createManagedAIPlayer(w http.ResponseWriter, tx *sqlx.Tx, args AddAIArgs, country string) error {
	players, err := listGamePlayersByGameID(tx, args.GameID)
	if err != nil {
		return err
	}
	if len(players) >= MaxGamePlayers {
		return writeGameStatusError(w, http.StatusConflict, fmt.Sprintf("game already has the maximum supported %d players", MaxGamePlayers))
	}
	for _, player := range players {
		if strings.EqualFold(strings.TrimSpace(player.Country), country) {
			return writeGameStatusError(w, http.StatusConflict, ErrorCountryAlreadyTaken)
		}
	}
	slot := len(players) + 1
	accountName := strings.TrimSpace(args.Name)
	if accountName == "" {
		accountName = fmt.Sprintf("AI %d", slot)
	}
	email := fmt.Sprintf("ai-game-%d-slot-%d-%d@open-ctp.local", args.GameID, slot, time.Now().UTC().UnixNano())
	var accountID int64
	if err := tx.QueryRowx(insertManagedAIAccountQuery, accountName, email).Scan(&accountID); err != nil {
		return fmt.Errorf("an error occurred while creating the server-managed AI account: %v", err)
	}
	if _, err := tx.Exec(insertManagedAIPlayerQuery, args.GameID, accountID, country); err != nil {
		return fmt.Errorf("an error occurred while creating the server-managed AI player: %v", err)
	}
	return nil
}

func assignExistingPlayerAI(w http.ResponseWriter, tx *sqlx.Tx, args AddAIArgs) error {
	var playerID int64
	if err := tx.Get(&playerID, `SELECT id FROM game_players WHERE game_id = $1 AND account_id = $2 FOR UPDATE`, args.GameID, args.TargetAccountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, "target account is not a game player")
		}
		return err
	}
	if _, err := tx.Exec(`UPDATE game_players SET controller_kind = 'ai', updated_at = NOW() WHERE id = $1`, playerID); err != nil {
		return fmt.Errorf("an error occurred while assigning AI control: %v", err)
	}
	return nil
}
