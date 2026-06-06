package lobby

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// LeaveLobbyArgs represents the expected structure of the request body for leaving a lobby.
//
// @Description Structure for the lobby leave request payload.
type LeaveLobbyArgs struct {
	SessionId *int64 `json:"session_id"`
}

func (a LeaveLobbyArgs) GetSessionID() *int64 {
	return a.SessionId
}

// LeaveLobby removes the authenticated account from its current open lobby.
//
// @Summary Leaves the current lobby
// @Description This endpoint removes the authenticated account from its current open lobby. If the owner leaves, ownership transfers to the longest-waiting remaining member, or the lobby is closed if it becomes empty.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body LeaveLobbyArgs true "lobby leave request body"
// @Success 200 {string} string "Successfully left lobby!"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/leave_lobby [post]
func LeaveLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := LeaveLobbyArgs{}
	if err := decodeJSONBody(r, &args); err != nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}

	session, err := requireLobbySession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginLobbyTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting lobby leave: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	accountID := int64(session.AccountID)
	openLobbyID, err := findOpenLobbyForAccount(tx, accountID)
	if err != nil {
		return err
	}
	if openLobbyID == nil {
		return writeLobbyStatusError(w, http.StatusConflict, "account is not currently in an open lobby")
	}

	lobby, err := getLobbyByID(tx, *openLobbyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeLobbyStatusError(w, http.StatusNotFound, fmt.Sprintf("no lobby exists with the ID %d", *openLobbyID))
		}
		return err
	}

	if _, err := tx.Exec("DELETE FROM lobby_members WHERE lobby_id = $1 AND account_id = $2", *openLobbyID, accountID); err != nil {
		return fmt.Errorf("an error occurred while removing lobby membership for lobby %d: %v", *openLobbyID, err)
	}

	if lobby.OwnerAccountID == accountID {
		nextOwner := lobbyOwnerRow{}
		err = tx.Get(&nextOwner, `SELECT a.id, a.name, lm.joined_at
FROM lobby_members lm
JOIN account a ON a.id = lm.account_id
WHERE lm.lobby_id = $1
ORDER BY lm.joined_at ASC, lm.id ASC
LIMIT 1`, *openLobbyID)
		switch {
		case err == sql.ErrNoRows:
			if _, err := tx.Exec("UPDATE lobbies SET is_closed = TRUE, closed_at = NOW(), last_activity_at = NOW(), updated_at = NOW() WHERE id = $1", *openLobbyID); err != nil {
				return fmt.Errorf("an error occurred while closing empty lobby %d: %v", *openLobbyID, err)
			}
		case err != nil:
			return fmt.Errorf("an error occurred while selecting the next lobby owner for lobby %d: %v", *openLobbyID, err)
		default:
			if _, err := tx.Exec("UPDATE lobbies SET owner_account_id = $1, owner_name = $2, last_activity_at = NOW(), updated_at = NOW() WHERE id = $3", nextOwner.AccountID, nextOwner.Name, *openLobbyID); err != nil {
				return fmt.Errorf("an error occurred while transferring lobby ownership for lobby %d: %v", *openLobbyID, err)
			}
		}
	} else {
		if _, err := tx.Exec("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1", *openLobbyID); err != nil {
			return fmt.Errorf("an error occurred while updating lobby activity for the ID %d: %v", *openLobbyID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing lobby leave: %v", err)
	}
	committed = true

	lobbyBrokerInstance.publish(*openLobbyID, lobbyEvent{
		name: "player_left",
		payload: lobbyEventPayload{
			LobbyID:   *openLobbyID,
			AccountID: accountID,
		},
	})

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Successfully left lobby!"))
	return nil
}
