package lobby

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// DeleteLobbyArgs represents the expected structure of the request body for deleting a lobby.
//
// @Description Structure for the lobby deletion request payload.
type DeleteLobbyArgs struct {
	LobbyId   int64  `json:"lobby_id"`
	SessionId *int64 `json:"session_id"`
}

func (a DeleteLobbyArgs) GetSessionID() *int64 {
	return a.SessionId
}

// DeleteLobby deletes a lobby for the current owner.
//
// @Summary Deletes a lobby
// @Description This endpoint deletes a lobby only for the authenticated lobby owner. It enforces ownership before deletion and does not allow non-owners to remove a lobby.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body DeleteLobbyArgs true "lobby deletion request body"
// @Success 200 {string} string "Successfully deleted lobby!"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 404 {object} error "Not Found"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/delete_lobby [delete]
func DeleteLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodDelete {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a DELETE request")
	}

	args := DeleteLobbyArgs{}
	if err := decodeJSONBody(r, &args); err != nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body: "+err.Error())
	}
	if args.LobbyId == 0 {
		return writeLobbyStatusError(w, http.StatusBadRequest, "lobby_id must be specified")
	}

	session, err := requireLobbySession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginLobbyTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting lobby deletion: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	lobby, err := getLobbyByID(tx, args.LobbyId)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeLobbyStatusError(w, http.StatusNotFound, fmt.Sprintf("no lobby exists with the ID %d", args.LobbyId))
		}
		return err
	}
	if err := closeLobbyIfInactive(tx, lobby, timeNow()); err != nil {
		return err
	}
	if err := ensureLobbyOwner(w, session, lobby); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM lobbies WHERE id = $1", args.LobbyId)
	if err != nil {
		return fmt.Errorf("an error occurred while deleting the lobby with the ID %d: %v", args.LobbyId, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while checking the affected rows: %v", err)
	}
	if rowsAffected == 0 {
		return writeLobbyStatusError(w, http.StatusNotFound, fmt.Sprintf("no lobby exists with the ID %d", args.LobbyId))
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing lobby deletion: %v", err)
	}
	committed = true

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Successfully deleted lobby!"))
	return nil
}
