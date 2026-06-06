package account

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

const ErrorInvalidEmailOrPassword = "invalid email or password"
const ErrorCurrentPasswordIncorrect = "current password is incorrect"
const ErrorSessionAccountMismatch = "session does not own the target account"

func writeAccountStatusError(w http.ResponseWriter, status int, message string) error {
	w.WriteHeader(status)
	return errors.New(message)
}

func ensureSessionOwnsAccount(w http.ResponseWriter, session *auth.Session, accountID int64) error {
	if int64(session.AccountID) != accountID {
		return writeAccountStatusError(w, http.StatusForbidden, ErrorSessionAccountMismatch)
	}

	return nil
}

func getAccountEmailByID(accountID int64, db *sqlx.DB) (string, error) {
	var email string
	err := db.QueryRow("SELECT email FROM account WHERE id = $1", accountID).Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("no account exists with the ID %d", accountID)
		}
		return "", fmt.Errorf("an error occurred while getting the account email for the ID %d: %v", accountID, err)
	}

	return email, nil
}
