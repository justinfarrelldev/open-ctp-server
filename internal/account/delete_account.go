package account

import (
	"encoding/json"
	"errors"
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

// DeleteAccount deletes an account by the account ID.
//
// @Summary Deletes an account
// @Description This endpoint deletes a player account.
// @Tags account
// @Accept json
// @Produce json
// @Param body body DeleteAccountArgs true "account deletion request body"
// @Success 200 {string} string "Successfully deleted account!"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/delete_account [delete]
func DeleteAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {

	if r.Method != http.MethodDelete {
		return errors.New("invalid request; request must be a DELETE request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := DeleteAccountArgs{}
	err := decoder.Decode(&args)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while decoding the request body: " + err.Error())
	}

	if args.AccountId == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("account_id must be specified")
	}

	if args.SessionId == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("a valid session_id must be specified")
	}

	fmt.Println("args: ", *args.AccountId)

	session, err := store.GetSession(*args.SessionId)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while retrieving the session: " + err.Error())
	}

	if session == nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("session not found")
	}

	if session.IsExpired() {
		w.WriteHeader(http.StatusForbidden)
		return errors.New("session has expired")
	}

	query := "DELETE FROM account WHERE id = $1"
	result, err := db.Exec(query, args.AccountId)
	if err != nil {
		return fmt.Errorf("an error occurred while deleting the account with the ID %d: %v", args.AccountId, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while checking the affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows were affected when the DELETE query ran for the account with ID %d", *args.AccountId)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted account!"))
	return nil
}
