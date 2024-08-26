package game

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ExpectedBody struct {
	password_protected bool
	password           string // Since this
}

func createGame(w http.ResponseWriter, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	game := ExpectedBody{}
	err := decoder.Decode(&game)

	if err != nil {
		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	// TODO salt & hash password here / handle it in Supabase or something then actually store the game somewhere

	return nil
}
