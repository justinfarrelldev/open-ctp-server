package game

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

// WorldRatingsResponse is the stable read-only world ratings summary.
type WorldRatingsResponse struct {
	Overall   int `json:"overall"`
	Science   int `json:"science"`
	Military  int `json:"military"`
	Economic  int `json:"economic"`
	Pollution int `json:"pollution"`
}

// WorldPollutionResponse is the stable read-only world pollution summary.
type WorldPollutionResponse struct {
	GameID             int64     `json:"game_id"`
	Turn               int       `json:"turn"`
	GlobalPollution    int       `json:"global_pollution"`
	TilePollutionCount int       `json:"tile_pollution_count"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func requireStoredReadableWorld(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) (int64, *world.StoredWorld, error) {
	gameID, _, err := requireReadableWorldGame(w, r, db, store)
	if err != nil {
		return 0, nil, err
	}

	stored, err := world.NewStore(db).Read(gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil, writeGameStatusError(w, http.StatusNotFound, fmt.Sprintf("world for game %d was not found", gameID))
		}
		return 0, nil, err
	}

	return gameID, stored, nil
}

func storedWorldToRatingsResponse(stored *world.StoredWorld) WorldRatingsResponse {
	ratings := stored.Snapshot.Turn.Ratings
	return WorldRatingsResponse{
		Overall:   ratings.Overall,
		Science:   ratings.Science,
		Military:  ratings.Military,
		Economic:  ratings.Economic,
		Pollution: ratings.Pollution,
	}
}

func storedWorldPollutionUpdatedAt(stored *world.StoredWorld) time.Time {
	if stored == nil {
		return time.Time{}
	}
	if !stored.Snapshot.Turn.Pollution.UpdatedAt.IsZero() {
		return stored.Snapshot.Turn.Pollution.UpdatedAt
	}
	if !stored.Snapshot.Turn.LastRecalculatedAt.IsZero() {
		return stored.Snapshot.Turn.LastRecalculatedAt
	}
	return stored.TurnUpdatedAt
}

func storedWorldToPollutionResponse(stored *world.StoredWorld) WorldPollutionResponse {
	pollution := stored.Snapshot.Turn.Pollution
	return WorldPollutionResponse{
		GameID:             stored.Snapshot.GameID,
		Turn:               stored.Snapshot.Turn.TurnNumber,
		GlobalPollution:    pollution.GlobalPollution,
		TilePollutionCount: pollution.TilePollutionCount,
		UpdatedAt:          storedWorldPollutionUpdatedAt(stored),
	}
}

// GetWorldRatings returns the persisted read-only world ratings summary for an authenticated joined player.
//
// @Summary Get world ratings
// @Description This endpoint returns the persisted read-only world ratings summary for any authenticated account that is still joined to the started game.
// @Tags world
// @Accept json
// @Produce json
// @Param game_id path int true "game ID"
// @Param session_id query int true "session ID"
// @Success 200 {object} WorldRatingsResponse
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /world/{game_id}/ratings [get]
func GetWorldRatings(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	_, stored, err := requireStoredReadableWorld(w, r, db, store)
	if err != nil {
		return err
	}

	return writeGameJSONResponse(w, http.StatusOK, storedWorldToRatingsResponse(stored))
}

// GetWorldPollution returns the persisted read-only world pollution summary for an authenticated joined player.
//
// @Summary Get world pollution
// @Description This endpoint returns the persisted read-only world pollution summary for any authenticated account that is still joined to the started game.
// @Tags world
// @Accept json
// @Produce json
// @Param game_id path int true "game ID"
// @Param session_id query int true "session ID"
// @Success 200 {object} WorldPollutionResponse
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /world/{game_id}/pollution [get]
func GetWorldPollution(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeGameStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	_, stored, err := requireStoredReadableWorld(w, r, db, store)
	if err != nil {
		return err
	}

	return writeGameJSONResponse(w, http.StatusOK, storedWorldToPollutionResponse(stored))
}
