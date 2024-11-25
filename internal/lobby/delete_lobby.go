package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// DeleteLobbyArgs represents the expected structure of the request body for deleting a lobby.
//
// @Description Structure for the lobby deletion request payload.
type DeleteLobbyArgs struct {
	// The lobby ID for the lobby that will be deleted.
	LobbyId int64 `json:"lobby_id"`
}

// DeleteLobby deletes a lobby by the lobby ID.
//
// @Summary Deletes a lobby
// @Description This endpoint deletes a multiplayer lobby.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body DeleteLobbyArgs true "lobby deletion request body"
// @Success 200 {string} string "Successfully deleted lobby!"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/delete_lobby [delete]
func DeleteLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB) error {

	if r.Method != http.MethodDelete {
		return errors.New("invalid request; request must be a DELETE request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := DeleteLobbyArgs{}
	err := decoder.Decode(&args)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while decoding the request body: " + err.Error())
	}

	if args.LobbyId == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("lobby_id must be specified")
	}

	query := "DELETE FROM lobby WHERE id = $1"
	result, err := db.Exec(query, args.LobbyId)
	if err != nil {
		return fmt.Errorf("an error occurred while deleting the lobby with the ID %d: %v", args.LobbyId, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while checking the affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no lobby exists with the ID %d", args.LobbyId)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted lobby!"))
	return nil
}
