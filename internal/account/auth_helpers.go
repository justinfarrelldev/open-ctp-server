package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

const ErrorInvalidEmailOrPassword = "invalid email or password"
const ErrorCurrentPasswordIncorrect = "current password is incorrect"
const ErrorSessionAccountMismatch = "session does not own the target account"

type accountQueryExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func writeAccountStatusError(w http.ResponseWriter, status int, message string) error {
	w.WriteHeader(status)
	return errors.New(message)
}

func writeAccountJSONResponse(w http.ResponseWriter, status int, payload any) error {
	responseBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding the response body: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(responseBytes)
	return nil
}

func ensureSessionOwnsAccount(w http.ResponseWriter, session *auth.Session, accountID int64) error {
	if int64(session.AccountID) != accountID {
		return writeAccountStatusError(w, http.StatusForbidden, ErrorSessionAccountMismatch)
	}

	return nil
}

func getAccountEmailByID(accountID int64, db accountQueryExecutor) (string, error) {
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

func validateExperienceLevel(level ExperienceLevel) error {
	if level < 0 || level > 5 {
		return errors.New("experience_level must be between 0 and 5 (0=easy, 5=impossible)")
	}

	return nil
}

func isEmailAvailable(email string, currentEmail *string, db accountQueryExecutor) (bool, error) {
	result, err := db.Query("SELECT * from account WHERE email = $1", email)
	if err != nil {
		return false, errors.New("an error occurred while checking whether the email for the account is unique: " + err.Error())
	}
	defer result.Close()

	if !result.Next() {
		return true, nil
	}

	if currentEmail != nil && email == *currentEmail {
		return true, nil
	}

	return false, nil
}

func validateEmail(email string, currentEmail *string, db accountQueryExecutor) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("an error occurred while checking whether the email for the account is valid: " + err.Error())
	}

	isAvailable, err := isEmailAvailable(email, currentEmail, db)
	if err != nil {
		return err
	}

	if !isAvailable {
		return errors.New("the provided email is already in use")
	}

	return nil
}

func beginAccountTx(db *sqlx.DB) (*sqlx.Tx, error) {
	return db.Beginx()
}
