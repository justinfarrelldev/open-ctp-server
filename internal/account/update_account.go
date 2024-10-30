package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// UpdateAccountArgs represents the expected structure of the request body for updating an account.
//
// @Description Structure for the account update request payload.
type UpdateAccountArgs struct {
	// The account ID for the account that will be updated.
	AccountId int64 `json:"account_id"`
	// The new name for the account.
	Name string `json:"name"`
	// The new info for the account.
	Info string `json:"info"`
	// The new location for the account.
	Location string `json:"location"`
	// The new email for the account.
	Email string `json:"email"`
	// The new experience level for the account.
	ExperienceLevel int `json:"experience_level"`
}

// UpdateAccount updates an account by the account ID.
//
// @Summary Updates an account
// @Description This endpoint updates a multiplayer account's info.
// @Tags account
// @Accept json
// @Produce json
// @Param body body UpdateAccountArgs true "account update request body"
//
//	@Success 200 {object} account.Account "Account successfully updated"
//
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/update_account [put]
func UpdateAccount(w http.ResponseWriter, r *http.Request, db *sql.DB) error {

	if r.Method != http.MethodPut {
		return errors.New("invalid request; request must be a PUT request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := UpdateAccountArgs{}
	err := decoder.Decode(&args)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	_, err = db.Exec("UPDATE account SET name = $1, info = $2, location = $3, email = $4, experience_level = $5 WHERE id = $6",
		args.Name, args.Info, args.Location, args.Email, args.ExperienceLevel, args.AccountId)

	if err != nil {
		return fmt.Errorf("an error occurred while updating the account with the ID %d: %v", args.AccountId, err)
	}

	// Now assemble the variables into the Account struct
	account := Account{
		Name:            args.Name,
		Info:            args.Info,
		Location:        args.Location,
		Email:           args.Email,
		ExperienceLevel: ExperienceLevel(args.ExperienceLevel),
	}

	accountBytes, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("Error marshalling struct: %v", err)
	}

	w.Write(accountBytes)

	fmt.Println("Successfully updated account!")
	return nil
}
