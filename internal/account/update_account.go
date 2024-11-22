package account

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// UpdateAccountArgs represents the expected structure of the request body for updating an account.
//
// @Description Structure for the account update request payload.
type UpdateAccountArgs struct {
	// The account to create.
	Account *AccountParam `json:"account"`
	// The password for the account to be created
	Password *string `json:"password"`
	// The account ID for the account that will be updated.
	AccountId *int64 `json:"account_id"`
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

	if args.AccountId == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("account_id must be specified")
	}

	if args.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("the password for the account must be specified")
	}

	if args.Account == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("account must be specified")
	}

	// Get the current password hash and salt from the database
	var storedHash, storedSalt string
	err = db.QueryRow("SELECT hash, salt FROM passwords WHERE id = $1", args.AccountId).Scan(&storedHash, &storedSalt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error retrieving account credentials: %v", err)
	}

	storedHashBytes, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error decoding stored hash: %v", err)
	}
	storedSaltBytes, err := base64.StdEncoding.DecodeString(storedSalt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error decoding stored salt: %v", err)
	}

	err = auth.Hasher.Compare(storedHashBytes, storedSaltBytes, []byte(*args.Password))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error comparing passwords: %v", err)
	}

	// Use reflection to check if at least one field other than AccountId is set
	v := reflect.ValueOf(args)
	numFields := v.NumField()
	anyFieldSet := false

	for i := 0; i < numFields; i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() && v.Type().Field(i).Name != "AccountId" {
			anyFieldSet = true
			break
		}
	}

	if !anyFieldSet {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("at least one field to update must be specified")
	}

	query := "UPDATE account SET "
	params := []interface{}{}
	paramIndex := 1

	if args.Account.Name != nil {
		query += fmt.Sprintf("name = $%d, ", paramIndex)
		params = append(params, args.Account.Name)
		paramIndex++
	}
	if args.Account.Info != nil {
		query += fmt.Sprintf("info = $%d, ", paramIndex)
		params = append(params, args.Account.Info)
		paramIndex++
	}
	if args.Account.Location != nil {
		query += fmt.Sprintf("location = $%d, ", paramIndex)
		params = append(params, args.Account.Location)
		paramIndex++
	}
	if args.Account.Email != nil {
		query += fmt.Sprintf("email = $%d, ", paramIndex)
		params = append(params, args.Account.Email)
		paramIndex++
	}
	if args.Account.ExperienceLevel != nil {
		query += fmt.Sprintf("experience_level = $%d, ", paramIndex)
		params = append(params, args.Account.ExperienceLevel)
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
