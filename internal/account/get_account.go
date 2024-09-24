package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetAccountArgs represents the expected structure of the request body for getting an account.
//
// @Description Structure for the account acquisition request payload.
type GetAccountArgs struct {
	// The account ID for the account that will be retrieved.
	AccountId int64 `json:"account_id"`
}

// GetAccount gets an account by the account ID.
//
// @Summary Gets an account
// @Description This endpoint gets a multiplayer account's info.
// @Tags account
// @Accept json
// @Produce json
// @Param body body GetAccountArgs true "account acquisition request body"
// @Success 200 {string} string "Account successfully retrieved"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/get_account [get]
func GetAccount(w http.ResponseWriter, r *http.Request, db *sql.DB) error {

	if r.Method != "GET" {
		return errors.New("invalid request; request must be a GET request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	argsGotten := GetAccountArgs{}
	err := decoder.Decode(&argsGotten)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	// TODO add sqlx so we don't have to manually provision row results from .Scan
	var (
		name            string
		info            string
		location        string
		email           string
		experienceLevel int
	)

	if err := db.QueryRow("SELECT name, info, location, email, experience_level FROM account WHERE id = $1", argsGotten.AccountId).
		Scan(&name, &info, &location, &email, &experienceLevel); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no account exists with the ID %d", argsGotten.AccountId)
		}
		return fmt.Errorf("an error occurred while getting the account with the ID %d: %v", argsGotten.AccountId, err)
	}

	// Now assemble the variables into the Account struct
	account := Account{
		Name:            name,
		Info:            info,
		Location:        location,
		Email:           email,
		ExperienceLevel: ExperienceLevel(experienceLevel),
	}

	accountBytes, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("Error marshalling struct: %v", err)
	}

	w.Write(accountBytes)

	fmt.Println("Successfully got account!")
	return nil
}
