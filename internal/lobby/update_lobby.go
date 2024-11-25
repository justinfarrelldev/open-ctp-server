package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/jmoiron/sqlx"
)

// UpdateLobbyArgs represents the expected structure of the request body for updating a lobby.
//
// @Description Structure for the lobby update request payload.
type UpdateLobbyArgs struct {
	// The lobby to update.
	Lobby *LobbyParam `json:"lobby"`
	// The lobby ID for the lobby that will be updated.
	LobbyId *int64 `json:"lobby_id"`
}

// UpdateLobby updates a lobby by the lobby ID.
//
// @Summary Updates a lobby
// @Description This endpoint updates a lobby's info.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body UpdateLobbyArgs true "lobby update request body"
// @Success 200 {string} string "Successfully updated lobby!"
// @Failure 400 {string} string "lobby_id must be specified"
// @Failure 500 {string} string "an error occurred while decoding the request body: <error message>"
// @Router /lobby/update_lobby [put]
func UpdateLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB) error {

	if r.Method != http.MethodPut {
		return errors.New("invalid request; request must be a PUT request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := UpdateLobbyArgs{}
	err := decoder.Decode(&args)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if args.LobbyId == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("lobby_id must be specified")
	}

	if args.Lobby == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("lobby must be specified")
	}

	// Use reflection to check if at least one field other than LobbyId is set
	v := reflect.ValueOf(args)
	numFields := v.NumField()
	anyFieldSet := false

	for i := 0; i < numFields; i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() && v.Type().Field(i).Name != "LobbyId" {
			anyFieldSet = true
			break
		}
	}

	if !anyFieldSet {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("at least one field to update must be specified")
	}

	query := "UPDATE lobby SET "
	params := []interface{}{}
	paramIndex := 1

	if args.Lobby.Name != nil {
		query += fmt.Sprintf("name = $%d, ", paramIndex)
		params = append(params, args.Lobby.Name)
		paramIndex++
	}
	if args.Lobby.OwnerName != nil {
		query += fmt.Sprintf("owner_name = $%d, ", paramIndex)
		params = append(params, args.Lobby.OwnerName)
		paramIndex++
	}
	if args.Lobby.IsClosed != nil {
		query += fmt.Sprintf("is_closed = $%d, ", paramIndex)
		params = append(params, args.Lobby.IsClosed)
		paramIndex++
	}
	if args.Lobby.IsMuted != nil {
		query += fmt.Sprintf("is_muted = $%d, ", paramIndex)
		params = append(params, args.Lobby.IsMuted)
		paramIndex++
	}
	if args.Lobby.IsPublic != nil {
		query += fmt.Sprintf("is_public = $%d, ", paramIndex)
		params = append(params, args.Lobby.IsPublic)
		paramIndex++
	}

	// Remove the trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d", paramIndex)
	params = append(params, args.LobbyId)

	_, err = db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("an error occurred while updating the lobby with the ID %d: %v", args.LobbyId, err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated lobby!"))
	return nil
}
