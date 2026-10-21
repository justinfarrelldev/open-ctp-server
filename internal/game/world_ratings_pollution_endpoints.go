package game

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// GetWorldRatings retires the legacy whole-world terrain-derived aggregate.
// Authentication/membership checks and migration guidance are shared with GetWorld.
//
// @Summary Retired canonical world ratings
// @Description Legacy ratings count undiscovered terrain. Use the viewer-safe gameplay projection's own statistics instead.
// @Tags world
// @Produce json
// @Param game_id path int true "game ID"
// @Param session_id query int true "session ID"
// @Failure 400 {string} string "invalid request"
// @Failure 401 {string} string "missing or invalid session"
// @Failure 403 {string} string "joined game player required"
// @Failure 404 {string} string "game not found"
// @Failure 409 {string} string "game has not started"
// @Failure 410 {object} map[string]string "use viewer-safe gameplay current resync"
// @Failure 500 {string} string "internal server error"
// @Router /world/{game_id}/ratings [get]
func GetWorldRatings(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	return GetWorld(w, r, db, store)
}

// GetWorldPollution retires the legacy aggregate computed from hidden terrain,
// which is unrelated to observed engine tile pollution or own economic history.
//
// @Summary Retired canonical world pollution
// @Description Legacy pollution weights undiscovered terrain. Use legally observed gameplay tile pollution instead.
// @Tags world
// @Produce json
// @Param game_id path int true "game ID"
// @Param session_id query int true "session ID"
// @Failure 400 {string} string "invalid request"
// @Failure 401 {string} string "missing or invalid session"
// @Failure 403 {string} string "joined game player required"
// @Failure 404 {string} string "game not found"
// @Failure 409 {string} string "game has not started"
// @Failure 410 {object} map[string]string "use viewer-safe gameplay current resync"
// @Failure 500 {string} string "internal server error"
// @Router /world/{game_id}/pollution [get]
func GetWorldPollution(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	return GetWorld(w, r, db, store)
}
