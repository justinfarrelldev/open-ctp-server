package game

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// CreateGameArgs represents the expected structure of the request body for creating a game.
//
// @Description Structure for the game creation request payload.
type CreateGameArgs struct {
	// PasswordProtected indicates whether the game is password-protected.
	// If true, a password must be provided.
	PasswordProtected bool `json:"password_protected"`
	// Password is the password for the game.
	// This field is required if PasswordProtected is true.
	// It must be longer than 6 characters.
	Password string `json:"password"`
}

const ERROR_PASSWORD_TOO_SHORT = "password must be longer than 6 characters"
const ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD = "password is required when password_protected is true"

// CreateGame handles the creation of a new game.
//
// @Summary Create a new game
// @Description This endpoint creates a new multiplayer game, optionally protected by a password.
// @Tags game
// @Accept json
// @Produce json
// @Param body body CreateGameArgs true "Game creation request body"
// @Success 201 {string} string "Game successfully created"
// @Failure 400 {object} error "Bad Request"
// @Failure 500 {object} error "Internal Server Error"
// @Router /game/create_game [post]
func CreateGame(w http.ResponseWriter, r *http.Request, db *sql.DB) error {

	if r.Method != "POST" {
		return errors.New("invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	game := CreateGameArgs{}
	err := decoder.Decode(&game)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if game.PasswordProtected && game.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}

	if !game.PasswordProtected && len(game.Password) > 0 {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New("password was provided despite password_protected being set to false")
	}

	if len(game.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_TOO_SHORT)
	}

	// TODO salt & hash password here / handle it in Supabase or something then actually store the game somewhere

	w.WriteHeader(http.StatusCreated)
	fmt.Println("Successfully created game!")
	return nil
}
