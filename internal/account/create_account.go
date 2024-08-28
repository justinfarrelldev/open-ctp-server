package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// CreateAccountArgs represents the expected structure of the request body for creating an account for use within the server.
//
// @Description Structure for the account creation request payload.
type CreateAccountArgs struct {
	// The account to create.
	account Account `json:"account"`
	// The password for the account to be created
	Password string `json:"password"`
}

const ERROR_PASSWORD_TOO_SHORT = "password must be longer than 6 characters"
const ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD = "password is required"

// CreateAccount handles the creation of a new account.
//
// @Summary Create a new account
// @Description This endpoint creates a new multiplayer account, protected by a password.
// @Tags account
// @Accept json
// @Produce json
// @Param body body CreateAccountArgs true "account creation request body"
// @Success 201 {string} string "Account successfully created"
// @Failure 400 {object} error "Bad Request"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/create_account [post]
func CreateAccount(w http.ResponseWriter, r *http.Request, db *sql.DB) error {

	if r.Method != "POST" {
		return errors.New("invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	account := CreateAccountArgs{}
	err := decoder.Decode(&account)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if account.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}

	if len(account.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_TOO_SHORT)
	}

	// TODO check if the account name is already taken in Supabase

	// TODO salt & hash password here / handle it in Supabase or something then actually store the game somewhere

	w.WriteHeader(http.StatusCreated)
	fmt.Println("Successfully created game!")
	return nil
}
