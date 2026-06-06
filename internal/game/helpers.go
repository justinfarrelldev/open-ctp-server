package game

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

const ErrorGameOwnerRequired = "only the game owner may perform this action"

const gameSelectByID = `SELECT id, created_at, updated_at, started_at, lobby_id, owner_account_id, name, password_protected, password, is_locked, is_public, style, rules_json, exclusions_json, status, starting_settlers, starting_gold
FROM games
WHERE id = $1`

type gameGetter interface {
	Get(dest any, query string, args ...any) error
}

type gamePatch struct {
	Name              *string
	IsLocked          *bool
	IsPublic          *bool
	Style             *string
	RulesJSON         json.RawMessage
	HasRulesJSON      bool
	ExclusionsJSON    json.RawMessage
	HasExclusionsJSON bool
}

func (p gamePatch) hasAnyField() bool {
	return p.Name != nil || p.IsLocked != nil || p.IsPublic != nil || p.Style != nil || p.HasRulesJSON || p.HasExclusionsJSON
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

func ensureGameOwner(w http.ResponseWriter, session *auth.Session, game *Game) error {
	if game.OwnerAccountID == nil || int64(session.AccountID) != *game.OwnerAccountID {
		return writeGameStatusError(w, http.StatusForbidden, ErrorGameOwnerRequired)
	}
	return nil
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
		fields = append(fields, fmt.Sprintf("style = $%d", paramIndex))
		params = append(params, trimmed)
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

	game, err := getGameByID(tx, gameID)
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

	updatedGame, err := getGameByID(tx, gameID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing game update: %v", err)
	}
	committed = true

	return writeGameJSONResponse(w, http.StatusOK, updatedGame)
}
