package lobby

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type GetLobbyArgs struct {
	LobbyId int64 `json:"lobby_id"`
}

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
