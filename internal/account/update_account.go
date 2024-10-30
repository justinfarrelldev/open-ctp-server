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
	Name *string `json:"name,omitempty"`
	// The new info for the account.
	Info *string `json:"info,omitempty"`
	// The new location for the account.
	Location *string `json:"location,omitempty"`
	// The new email for the account.
	Email *string `json:"email,omitempty"`
	// The new experience level for the account.
	ExperienceLevel *int `json:"experience_level,omitempty"`
}

// UpdateAccount updates an account by the account ID.
//
// @Summary Updates an account
// @Description This endpoint updates an account's info.
// @Tags account
// @Accept json
// @Produce json
// @Param body body UpdateAccountArgs true "account update request body"
// @Success 200 {string} string "Successfully updated account!"
// @Failure 400 {string} string "account_id must be specified"
// @Failure 500 {string} string "an error occurred while decoding the request body: <error message>"
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

	if args.AccountId == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("account_id must be specified")
	}

	query := "UPDATE account SET "
	params := []interface{}{}
	paramIndex := 1

	if args.Name != nil {
		query += fmt.Sprintf("name = $%d, ", paramIndex)
		params = append(params, *args.Name)
		paramIndex++
	}
	if args.Info != nil {
		query += fmt.Sprintf("info = $%d, ", paramIndex)
		params = append(params, *args.Info)
		paramIndex++
	}
	if args.Location != nil {
		query += fmt.Sprintf("location = $%d, ", paramIndex)
		params = append(params, *args.Location)
		paramIndex++
	}
	if args.Email != nil {
		query += fmt.Sprintf("email = $%d, ", paramIndex)
		params = append(params, *args.Email)
		paramIndex++
	}
	if args.ExperienceLevel != nil {
		query += fmt.Sprintf("experience_level = $%d, ", paramIndex)
		params = append(params, *args.ExperienceLevel)
		paramIndex++
	}

	// Remove the trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d", paramIndex)
	params = append(params, args.AccountId)

	_, err = db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("an error occurred while updating the account with the ID %d: %v", args.AccountId, err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully updated account!"))
	return nil
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getIntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
