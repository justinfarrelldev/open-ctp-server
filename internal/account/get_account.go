package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// GetAccountArgs represents the expected structure of the request body for getting an account.
//
// @Description Structure for the account acquisition request payload.
type GetAccountArgs struct {
	// The account ID for the account that will be retrieved.
	AccountId int64 `json:"account_id"`
	// A valid session ID for the account (so we know they are signed in)
	SessionId int64 `json:"session_id"`
}

// GetAccount gets an account by the account ID.
//
// @Summary Gets an account
// @Description This endpoint gets a multiplayer account's info.
// @Tags account
// @Accept json
// @Produce json
// @Param account_id query int true "account ID"
// @Param session_id query int true "session ID"
//
// @Success 200 {object} account.Account "Account successfully retrieved"
//
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/get_account [get]
func GetAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != "GET" {
		return errors.New("invalid request; request must be a GET request")
	}

	queryParams := r.URL.Query()
	accountIdStr := queryParams.Get("account_id")
	if accountIdStr == "" {
		return errors.New("account_id is required")
	}

	accountId, err := strconv.ParseInt(accountIdStr, 10, 64)
	if err != nil {
		return errors.New("invalid account_id")
	}

	sessionIdStr := queryParams.Get("session_id")
	if sessionIdStr == "" {
		return errors.New("session_id is required")
	}

	sessionId, err := strconv.ParseInt(sessionIdStr, 10, 64)
	if err != nil {
		return errors.New("invalid session_id")
	}

	session, err := store.GetSession(sessionId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while retrieving the session: " + err.Error())
	}

	if session == nil {
		w.WriteHeader(http.StatusForbidden)
		return errors.New("session not found")
	}

	if session.IsExpired() {
		w.WriteHeader(http.StatusForbidden)
		return errors.New("session has expired")
	}

	// Refresh the session
	_, err = store.RefreshSession(session.AccountID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.New("an error occurred while refreshing the session: " + err.Error())
	}

	var (
		name            string
		info            string
		location        string
		email           string
		experienceLevel int
	)

	fmt.Println("account id is", accountId)

	if err := db.QueryRow("SELECT name, info, location, email, experience_level FROM account WHERE id = $1", accountId).
		Scan(&name, &info, &location, &email, &experienceLevel); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no account exists with the ID %d", accountId)
		}
		return fmt.Errorf("an error occurred while getting the account with the ID %d: %v", accountId, err)
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
