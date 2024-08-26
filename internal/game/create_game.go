package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ExpectedBody struct {
	PasswordProtected bool   `json:"password_protected"`
	Password          string `json:"password"`
}

func CreateGame(w http.ResponseWriter, r *http.Request) error {

	if r.Method != "POST" {
		return errors.New("invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	game := ExpectedBody{}
	err := decoder.Decode(&game)

	if err != nil {
		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if game.PasswordProtected && game.Password == "" {
		return errors.New("password is required when password_protected is true")
	}

	if !game.PasswordProtected && len(game.Password) > 0 {
		return errors.New("password was provided despite password_protected being set to false")
	}

	// TODO salt & hash password here / handle it in Supabase or something then actually store the game somewhere

	fmt.Println("Successfully created game!")
	return nil
}
