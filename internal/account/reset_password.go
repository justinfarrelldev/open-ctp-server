package account

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type ResetPasswordArgs struct {
	AccountId       *int64 `json:"account_id"`
	SessionId       *int64 `json:"session_id"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (a ResetPasswordArgs) GetSessionID() *int64 {
	return a.SessionId
}

func ResetPassword(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeAccountStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := ResetPasswordArgs{}
	if err := decoder.Decode(&args); err != nil {
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

	if args.CurrentPassword == "" {
		return writeAccountStatusError(w, http.StatusBadRequest, "current_password must be specified")
	}

	if args.NewPassword == "" {
		return writeAccountStatusError(w, http.StatusBadRequest, "new_password must be specified")
	}

	if len(args.NewPassword) < 6 {
		return writeAccountStatusError(w, http.StatusBadRequest, ERROR_PASSWORD_TOO_SHORT)
	}

	accountEmail, err := getAccountEmailByID(*args.AccountId, db)
	if err != nil {
		return err
	}

	storedHash, storedSalt, err := auth.GetHashAndSalt(accountEmail, db)
	if err != nil {
		return err
	}

	if err := auth.Hasher.Compare(storedHash, storedSalt, []byte(args.CurrentPassword)); err != nil {
		return writeAccountStatusError(w, http.StatusUnauthorized, ErrorCurrentPasswordIncorrect)
	}

	hashSalt, err := auth.Hasher.GenerateHash([]byte(args.NewPassword), nil)
	if err != nil {
		return fmt.Errorf("an error occurred while hashing the new password: %v", err)
	}

	if err := auth.UpdateHashAndSalt(hashSalt, accountEmail, db); err != nil {
		return err
	}

	if err := store.DeleteSessionsForAccountExcept(*args.AccountId, session.ID); err != nil {
		return fmt.Errorf("an error occurred while invalidating old sessions: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Successfully reset password!"))
	return nil
}
