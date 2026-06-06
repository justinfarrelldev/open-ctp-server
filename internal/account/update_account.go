package account

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
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
	// A valid session ID for the account (so we know they are signed in)
	SessionId *int64 `json:"session_id"`
}

func (a UpdateAccountArgs) GetSessionID() *int64 {
	return a.SessionId
}

// UpdateAccount updates an account by the account ID.
//
// @Summary Updates an account
// @Description This endpoint updates an account's info for the authenticated account owner.
// @Tags account
// @Accept json
// @Produce json
// @Param body body UpdateAccountArgs true "account update request body"
// @Success 200 {string} string "Successfully updated account!"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/update_account [put]
func UpdateAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPut {
		return writeAccountStatusError(w, http.StatusBadRequest, "invalid request; request must be a PUT request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := UpdateAccountArgs{}
	err := decoder.Decode(&args)
	if err != nil {
		return writeAccountStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}

	if args.AccountId == nil {
		return writeAccountStatusError(w, http.StatusBadRequest, "account_id must be specified")
	}

	session, err := auth.RequireSessionFromJSONBody(w, store, args)
	if err != nil {
		return err
	}

	if err := ensureSessionOwnsAccount(w, session, *args.AccountId); err != nil {
		return err
	}

	if args.Password == nil {
		return writeAccountStatusError(w, http.StatusBadRequest, "the password for the account must be specified")
	}

	if args.Account == nil {
		return writeAccountStatusError(w, http.StatusBadRequest, "account must be specified")
	}

	anyFieldSet := args.Account.Name != nil ||
		args.Account.Info != nil ||
		args.Account.Location != nil ||
		args.Account.Email != nil ||
		args.Account.ExperienceLevel != nil
	if !anyFieldSet {
		return writeAccountStatusError(w, http.StatusBadRequest, "at least one field to update must be specified")
	}

	tx, err := beginAccountTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting account update: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	currentEmail, err := getAccountEmailByID(*args.AccountId, tx)
	if err != nil {
		return err
	}

	if args.Account.ExperienceLevel != nil {
		if err := validateExperienceLevel(*args.Account.ExperienceLevel); err != nil {
			return writeAccountStatusError(w, http.StatusBadRequest, err.Error())
		}
	}

	if args.Account.Email != nil {
		if err := validateEmail(*args.Account.Email, &currentEmail, tx); err != nil {
			return writeAccountStatusError(w, http.StatusBadRequest, err.Error())
		}
	}

	storedHash, storedSalt, err := auth.GetHashAndSalt(currentEmail, tx)
	if err != nil {
		return err
	}

	if err := auth.Hasher.Compare(storedHash, storedSalt, []byte(*args.Password)); err != nil {
		return writeAccountStatusError(w, http.StatusUnauthorized, ErrorCurrentPasswordIncorrect)
	}

	query := "UPDATE account SET "
	params := []any{}
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

	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d", paramIndex)
	params = append(params, args.AccountId)

	_, err = tx.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("an error occurred while updating the account with the ID %d: %v", *args.AccountId, err)
	}

	if args.Account.Email != nil && *args.Account.Email != currentEmail {
		_, err = tx.Exec("UPDATE passwords SET account_email = $1 WHERE account_email = $2", *args.Account.Email, currentEmail)
		if err != nil {
			return fmt.Errorf("an error occurred while synchronizing updated account credentials: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing account update: %v", err)
	}
	committed = true

	return writeAccountJSONResponse(w, http.StatusOK, "Successfully updated account!")
}
