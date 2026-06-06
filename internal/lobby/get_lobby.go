package lobby

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// GetLobbyArgs represents the expected structure of the request body for reading a lobby.
//
// @Description Structure for the lobby read request payload.
type GetLobbyArgs struct {
	LobbyId int64 `json:"lobby_id"`
}

// GetLobby returns the current state of a lobby.
//
// @Summary Gets a lobby
// @Description This endpoint returns a lobby by ID. If the lobby has become inactive, it is closed during the read and the endpoint returns HTTP 410 Gone instead of lobby data.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body GetLobbyArgs true "lobby read request body"
// @Success 200 {object} Lobby "Lobby successfully retrieved"
// @Failure 400 {object} error "Bad Request"
// @Failure 404 {object} error "Not Found"
// @Failure 410 {object} error "Gone"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/get_lobby [get]
func GetLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	args := GetLobbyArgs{}
	if err := decodeJSONBody(r, &args); err != nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body: "+err.Error())
	}
	if args.LobbyId == 0 {
		return writeLobbyStatusError(w, http.StatusBadRequest, "lobby_id must be specified")
	}

	lobby, err := getLobbyByIDNoTx(db, args.LobbyId)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeLobbyStatusError(w, http.StatusNotFound, fmt.Sprintf("no lobby exists with the ID %d", args.LobbyId))
		}
		return err
	}

	if lobby.ClosedAt != nil || lobby.IsClosed {
		return writeLobbyStatusError(w, http.StatusGone, fmt.Sprintf("lobby %d is closed", args.LobbyId))
	}

	return writeLobbyJSON(w, lobby)
}
