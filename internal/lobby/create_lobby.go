package lobby

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// CreateLobbyArgs represents the expected structure of the request body for creating a lobby.
//
// @Description Structure for the lobby creation request payload.
type CreateLobbyArgs struct {
	Lobby     LobbyParam `json:"lobby"`
	Password  string     `json:"password"`
	SessionId *int64     `json:"session_id"`
}

func (a CreateLobbyArgs) GetSessionID() *int64 {
	return a.SessionId
}

const ERROR_PASSWORD_TOO_SHORT = "password must be longer than 6 characters"
const ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD = "password is required"

// CreateLobby creates a new lobby for the authenticated account owner.
//
// @Summary Create a new lobby
// @Description This endpoint creates a new lobby for the authenticated account, makes that account the owner, adds it as the first member, and rejects requests when the account is already in another open lobby.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body CreateLobbyArgs true "lobby creation request body"
// @Success 201 {string} string "Lobby successfully created"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 409 {object} error "Conflict"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/create_lobby [post]
func CreateLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := CreateLobbyArgs{}
	if err := decodeJSONBody(r, &args); err != nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}

	if args.Password == "" {
		return writeLobbyStatusError(w, http.StatusBadRequest, ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}
	if len(args.Password) < 6 {
		return writeLobbyStatusError(w, http.StatusBadRequest, ERROR_PASSWORD_TOO_SHORT)
	}
	if args.Lobby.Name == nil || *args.Lobby.Name == "" {
		return writeLobbyStatusError(w, http.StatusBadRequest, "lobby.name must be specified")
	}

	session, err := requireLobbySession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginLobbyTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting lobby creation: %v", err)
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

	ownerName, err := lookupAccountName(tx, accountID)
	if err != nil {
		return err
	}

	isMuted := false
	if args.Lobby.IsMuted != nil {
		isMuted = *args.Lobby.IsMuted
	}
	isPublic := true
	if args.Lobby.IsPublic != nil {
		isPublic = *args.Lobby.IsPublic
	}

	var lobbyID int64
	if err := tx.QueryRowx(`INSERT INTO lobbies (name, owner_name, owner_account_id, is_closed, is_muted, is_public, updated_at, last_activity_at)
VALUES ($1, $2, $3, FALSE, $4, $5, NOW(), NOW())
RETURNING id`, *args.Lobby.Name, ownerName, accountID, isMuted, isPublic).Scan(&lobbyID); err != nil {
		return fmt.Errorf("an error occurred while storing the lobby in the database: %v", err)
	}

	if _, err := tx.Exec("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)", lobbyID, accountID); err != nil {
		return fmt.Errorf("an error occurred while creating lobby membership: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing lobby creation: %v", err)
	}
	committed = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(struct {
		ID int64 `json:"id"`
	}{ID: lobbyID})
}
