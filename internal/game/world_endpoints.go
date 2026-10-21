package game

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

func parseWorldPathGameID(r *http.Request) (int64, error) {
	gameIDStr := strings.TrimSpace(r.PathValue("game_id"))
	if gameIDStr == "" {
		return 0, errors.New("game_id must be specified")
	}

	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil || gameID <= 0 {
		return 0, errors.New("invalid game_id")
	}

	return gameID, nil
}

func requireReadableWorldGame(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) (int64, *Game, error) {
	gameID, err := parseWorldPathGameID(r)
	if err != nil {
		return 0, nil, writeGameStatusError(w, http.StatusBadRequest, err.Error())
	}

	session, err := requireGameQuerySession(w, r, store)
	if err != nil {
		return 0, nil, err
	}

	game, err := getGameByID(db, gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil, writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("no game exists with the ID %d", gameID))
		}
		return 0, nil, err
	}

	if _, err := ensureGamePlayer(w, db, gameID, int64(session.AccountID)); err != nil {
		return 0, nil, err
	}
	if game.Status == DefaultGameStatus {
		return 0, nil, writeGameStatusError(w, http.StatusConflict, fmt.Sprintf("game %d has not started", gameID))
	}

	return gameID, game, nil
}

// GetWorld retires the unsafe canonical-map download. Joined players must use
// gameplay current resync, which applies the same viewer filtering as receipts.
//
// @Summary Retired canonical world download
// @Description Canonical maps, generation seeds and foreign spawns are private. Use the gameplay events endpoint with X-Gameplay-Resync: current for a viewer-safe snapshot.
// @Tags world
// @Accept json
// @Produce json
// @Param game_id path int true "game ID"
// @Param session_id query int true "session ID"
// @Failure 410 {string} string "use viewer-safe gameplay current resync"
// @Failure 400 {string} string "invalid request method, missing game_id, or invalid game_id"
// @Failure 401 {string} string "missing or invalid session"
// @Failure 403 {string} string "only joined game players may perform this action"
// @Failure 404 {string} string "game or persisted world not found"
// @Failure 409 {string} string "game has not started"
// @Failure 500 {string} string "internal server error"
// @Router /world/{game_id} [get]
func GetWorld(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	if _, _, err := requireReadableWorldGame(w, r, db, store); err != nil {
		return err
	}
	return writeGameStatusError(w, http.StatusGone, "canonical world download is disabled; use viewer-safe gameplay current resync")
}

// GetWorldSpaceLayer returns the documented CTP2 placeholder for the future space layer.
//
// @Summary Get the world space layer
// @Description This endpoint is reserved for future CTP1 space-layer support. Authenticated joined players may call it today, but it always returns the fixed 404 JSON placeholder for CTP2 worlds.
// @Tags world
// @Accept json
// @Produce json
// @Param game_id path int true "game ID"
// @Param session_id query int true "session ID"
// @Failure 400 {string} string "invalid request method, missing game_id, or invalid game_id"
// @Failure 401 {string} string "missing or invalid session"
// @Failure 403 {string} string "only joined game players may perform this action"
// @Failure 404 {object} WorldSpaceLayerErrorResponseDocs "CTP2 placeholder response after auth, membership, and started-game checks"
// @Failure 409 {string} string "game has not started"
// @Failure 500 {string} string "internal server error"
// @Router /world/{game_id}/space [get]
func GetWorldSpaceLayer(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	if _, _, err := requireReadableWorldGame(w, r, db, store); err != nil {
		return err
	}

	return writeGameJSONResponse(w, http.StatusNotFound, world.ErrorResponse{Error: world.SpaceLayerUnsupportedError})
}
