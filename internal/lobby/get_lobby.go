package lobby

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// GetLobbyArgs represents the expected structure of the request body for getting a lobby.
//
// @Description Structure for the lobby acquisition request payload.
type GetLobbyArgs struct {
	// The lobby ID for the lobby that will be retrieved.
	LobbyId int8 `json:"lobby_id"`
}

// GetLobby gets a lobby by the lobby ID.
//
// @Summary Gets a lobby
// @Description This endpoint gets a multiplayer lobby's info.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body GetLobbyArgs true "lobby acquisition request body"
//
// @Success 200 {object} lobby.Lobby "Lobby successfully retrieved"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/get_lobby [get]
func GetLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {

	if r.Method != "GET" {
		return errors.New("invalid request; request must be a GET request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	argsGotten := GetLobbyArgs{}
	err := decoder.Decode(&argsGotten)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while decoding the request body: " + err.Error())
	}

	var lobby Lobby

	query := "SELECT id, name, owner_name, is_closed, is_muted, is_public FROM lobby WHERE id = $1"
	if err := db.Get(&lobby, query, argsGotten.LobbyId); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no lobby exists with the ID %d", argsGotten.LobbyId)
		}
		return fmt.Errorf("an error occurred while getting the lobby with the ID %d: %v", argsGotten.LobbyId, err)
	}

	lobbyBytes, err := json.Marshal(lobby)
	if err != nil {
		return fmt.Errorf("Error marshalling struct: %v", err)
	}

	w.Write(lobbyBytes)

	fmt.Println("Successfully got lobby!")
	return nil
}
