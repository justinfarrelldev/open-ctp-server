package lobby

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// UpdateLobbyArgs represents the expected structure of the request body for updating a lobby.
//
// @Description Structure for the lobby update request payload.
type UpdateLobbyArgs struct {
	Lobby     *LobbyParam `json:"lobby"`
	LobbyId   *int64      `json:"lobby_id"`
	SessionId *int64      `json:"session_id"`
}

func (a UpdateLobbyArgs) GetSessionID() *int64 {
	return a.SessionId
}

// UpdateLobby updates mutable lobby fields for the current owner.
//
// @Summary Updates a lobby
// @Description This endpoint lets only the authenticated lobby owner update the lobby name, mute state, or public state. If the lobby has already auto-closed from inactivity, the endpoint returns HTTP 410 Gone.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body UpdateLobbyArgs true "lobby update request body"
// @Success 200 {string} string "Successfully updated lobby!"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 410 {object} error "Gone"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/update_lobby [put]
func UpdateLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPut {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a PUT request")
	}

	args := UpdateLobbyArgs{}
	if err := decodeJSONBody(r, &args); err != nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}
	if args.LobbyId == nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "lobby_id must be specified")
	}
	if args.Lobby == nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "lobby must be specified")
	}

	anyFieldSet := args.Lobby.Name != nil || args.Lobby.IsMuted != nil || args.Lobby.IsPublic != nil
	if !anyFieldSet {
		return writeLobbyStatusError(w, http.StatusBadRequest, "at least one field to update must be specified")
	}

	session, err := requireLobbySession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginLobbyTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting lobby update: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	lobby, err := getLobbyByID(tx, *args.LobbyId)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeLobbyStatusError(w, http.StatusNotFound, fmt.Sprintf("no lobby exists with the ID %d", *args.LobbyId))
		}
		return err
	}

	if err := closeLobbyIfInactive(tx, lobby, timeNow()); err != nil {
		return err
	}
	if lobby.ClosedAt != nil || lobby.IsClosed {
		return writeLobbyStatusError(w, http.StatusGone, fmt.Sprintf("lobby %d is closed", *args.LobbyId))
	}
	if err := ensureLobbyOwner(w, session, lobby); err != nil {
		return err
	}

	fields := make([]string, 0, 5)
	params := make([]any, 0, 4)
	paramIndex := 1
	if args.Lobby.Name != nil {
		trimmed := strings.TrimSpace(*args.Lobby.Name)
		if trimmed == "" {
			return writeLobbyStatusError(w, http.StatusBadRequest, "lobby.name must not be empty")
		}
		fields = append(fields, fmt.Sprintf("name = $%d", paramIndex))
		params = append(params, trimmed)
		paramIndex++
	}
	if args.Lobby.IsMuted != nil {
		fields = append(fields, fmt.Sprintf("is_muted = $%d", paramIndex))
		params = append(params, *args.Lobby.IsMuted)
		paramIndex++
	}
	if args.Lobby.IsPublic != nil {
		fields = append(fields, fmt.Sprintf("is_public = $%d", paramIndex))
		params = append(params, *args.Lobby.IsPublic)
		paramIndex++
	}
	fields = append(fields, "updated_at = NOW()", "last_activity_at = NOW()")
	params = append(params, *args.LobbyId)

	query := fmt.Sprintf("UPDATE lobbies SET %s WHERE id = $%d", strings.Join(fields, ", "), paramIndex)
	if _, err := tx.Exec(query, params...); err != nil {
		return fmt.Errorf("an error occurred while updating the lobby with the ID %d: %v", *args.LobbyId, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing lobby update: %v", err)
	}
	committed = true

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Successfully updated lobby!"))
	return nil
}

var timeNow = func() time.Time { return time.Now() }
