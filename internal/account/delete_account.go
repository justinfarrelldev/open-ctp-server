package account

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// DeleteAccountArgs represents the expected structure of the request body for deleting an account.
//
// @Description Structure for the account deletion request payload.
type DeleteAccountArgs struct {
	// The account ID for the account that will be deleted.
	AccountId *int64 `json:"account_id,omitempty"`
	// A valid session ID for the account (so we know they are signed in)
	SessionId *int64 `json:"session_id,omitempty"`
}

func (a DeleteAccountArgs) GetSessionID() *int64 {
	return a.SessionId
}

// DeleteAccount deletes an account by the account ID.
//
// @Summary Deletes an account
// @Description This endpoint deletes a player account for the authenticated account owner and invalidates that account's sessions after a successful delete.
// @Tags account
// @Accept json
// @Produce json
// @Param body body DeleteAccountArgs true "account deletion request body"
// @Success 200 {string} string "Successfully deleted account!"
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/delete_account [delete]
func DeleteAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodDelete {
		return writeAccountStatusError(w, http.StatusBadRequest, "invalid request; request must be a DELETE request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := DeleteAccountArgs{}
	err := decoder.Decode(&args)
	if err != nil {
		return writeAccountStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body: "+err.Error())
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

	tx, err := beginAccountTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting account deletion: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	accountEmail, err := getAccountEmailByID(*args.AccountId, tx)
	if err != nil {
		return err
	}

	if err := store.DeleteSessionsForAccountTx(tx, *args.AccountId); err != nil {
		return fmt.Errorf("an error occurred while invalidating account sessions: %v", err)
	}

	if _, err := tx.Exec("DELETE FROM passwords WHERE account_email = $1", accountEmail); err != nil {
		return fmt.Errorf("an error occurred while deleting credentials for the account with ID %d: %v", *args.AccountId, err)
	}

	result, err := tx.Exec("DELETE FROM account WHERE id = $1", args.AccountId)
	if err != nil {
		return fmt.Errorf("an error occurred while deleting the account with the ID %d: %v", *args.AccountId, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while checking the affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows were affected when the DELETE query ran for the account with ID %d", *args.AccountId)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing account deletion: %v", err)
	}
	committed = true

	return writeAccountJSONResponse(w, http.StatusOK, "Successfully deleted account!")
}
