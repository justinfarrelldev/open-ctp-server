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

func storedWorldToReadResponse(stored *world.StoredWorld) world.WorldReadResponse {
	snapshot := stored.Snapshot.Clone()
	return world.WorldReadResponse{
		GameID:           snapshot.GameID,
		GameStatus:       snapshot.GameStatus,
		GeneratorVersion: snapshot.GeneratorVersion,
		HostParams:       snapshot.Config.ToHostParams(),
		PlayerOrder:      append([]int64(nil), snapshot.PlayerOrder...),
		Spawns:           append([]world.PlayerSpawn(nil), snapshot.Spawns...),
		Tiles:            append([]world.WorldTile(nil), snapshot.Tiles...),
	}
}

// GetWorld returns the persisted generated world for an authenticated joined player.
//
// @Summary Get a generated world
// @Description This endpoint returns the persisted generated world state for any authenticated account that is still joined to the game after the game has started.
// @Tags world
// @Accept json
// @Produce json
// @Param game_id path int true "game ID"
// @Param session_id query int true "session ID"
// @Success 200 {object} world.WorldReadResponse
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /world/{game_id} [get]
func GetWorld(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	gameID, _, err := requireReadableWorldGame(w, r, db, store)
	if err != nil {
		return err
	}

	stored, err := world.NewStore(db).Read(gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("world for game %d was not found", gameID))
		}
		return err
	}

	return writeGameJSONResponse(w, http.StatusOK, storedWorldToReadResponse(stored))
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
// @Success 404 {object} world.ErrorResponse
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
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
