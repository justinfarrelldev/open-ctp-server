package lobby

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
)

// CreateLobbyArgs represents the expected structure of the request body for creating a lobby for use within the server.
//
// @Description Structure for the lobby creation request payload.
type CreateLobbyArgs struct {
	// The lobby to create.
	Lobby Lobby `json:"lobby"`
	// The password for the lobby to be created
	Password string `json:"password"`
}

const ERROR_PASSWORD_TOO_SHORT = "password must be longer than 6 characters"
const ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD = "password is required"

func isEmailValid(email string, db *sql.DB) (bool, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false, errors.New("an error occurred while checking whether the email for the lobby is valid: " + err.Error())
	}

	result, err := db.Query("SELECT * from lobby WHERE email = $1", email)

	if err != nil {
		return false, errors.New("an error occurred while checking whether the email for the lobby is unique: " + err.Error())
	}

	defer result.Close()

	if result.Next() {
		// If result.Next() returns true, there is at least one row, so the email is not unique.
		return false, nil
	}

	// If no rows were found, the email is unique (or not in the database).
	return true, nil
}

// CreateLobby handles the creation of a new lobby.
//
// @Summary Create a new lobby
// @Description This endpoint creates a new multiplayer lobby, protected by a password.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body CreateLobbyArgs true "lobby creation request body"
// @Success 201 {string} string "Lobby successfully created"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/create_lobby [post]
func CreateLobby(w http.ResponseWriter, r *http.Request, db *sql.DB) error {

	if r.Method != "POST" {
		return errors.New("invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	lobby := CreateLobbyArgs{}
	err := decoder.Decode(&lobby)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if lobby.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}

	if len(lobby.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_TOO_SHORT)
	}

	// TODO store password if needed
	return nil
}

func storeLobby(lobby *Lobby, db *sql.DB) error {
	result, err := db.Query(
		"INSERT INTO lobby (id, name, owner_name, is_closed, is_muted, is_public) VALUES ($1, $2, $3, $4, $5, $6)",
		lobby.ID, lobby.Name, lobby.OwnerName, lobby.IsClosed, lobby.IsMuted, lobby.IsPublic,
	)
	if err != nil {
		return errors.New("an error occurred while inserting a lobby into the database: " + err.Error())
	}

	defer result.Close()
	return nil
}
