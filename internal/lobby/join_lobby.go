package lobby

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// JoinLobbyArgs represents the expected structure of the request body for joining a lobby.
//
// @Description Structure for the lobby join request payload.
type JoinLobbyArgs struct {
	LobbyId   int64  `json:"lobby_id"`
	SessionId *int64 `json:"session_id"`
}

func (a JoinLobbyArgs) GetSessionID() *int64 {
	return a.SessionId
}

// JoinLobby adds the authenticated account to a lobby.
//
// @Summary Joins a lobby
// @Description This endpoint joins the authenticated account to the requested lobby, rejects the request if the account is already in another open lobby, and returns HTTP 410 Gone when the target lobby has auto-closed from inactivity.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body JoinLobbyArgs true "lobby join request body"
// @Success 200 {string} string "Successfully joined lobby!"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 404 {object} error "Not Found"
// @Failure 409 {object} error "Conflict"
// @Failure 410 {object} error "Gone"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/join_lobby [post]
func JoinLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := JoinLobbyArgs{}
	if err := decodeJSONBody(r, &args); err != nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
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
		return fmt.Errorf("an error occurred while starting lobby join: %v", err)
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
	if openLobbyID != nil {
		return writeLobbyStatusError(w, http.StatusConflict, "account is already in an open lobby")
	}

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
	if lobby.ClosedAt != nil || lobby.IsClosed {
		return writeLobbyStatusError(w, http.StatusGone, fmt.Sprintf("lobby %d is closed", args.LobbyId))
	}

	if _, err := tx.Exec("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)", args.LobbyId, accountID); err != nil {
		return fmt.Errorf("an error occurred while joining the lobby with the ID %d: %v", args.LobbyId, err)
	}
	if _, err := tx.Exec("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1", args.LobbyId); err != nil {
		return fmt.Errorf("an error occurred while updating lobby activity for the ID %d: %v", args.LobbyId, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing lobby join: %v", err)
	}
	committed = true

	lobbyBrokerInstance.publish(args.LobbyId, lobbyEvent{
		name: "player_connected",
		payload: lobbyEventPayload{
			LobbyID:   args.LobbyId,
			AccountID: accountID,
		},
	})

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Successfully joined lobby!"))
	return nil
}
