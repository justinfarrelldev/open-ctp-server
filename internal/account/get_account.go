package account

import (
	"database/sql"
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
// @Description This endpoint gets a multiplayer account's info and requires a valid session.
// @Tags account
// @Accept json
// @Produce json
// @Param account_id query int true "account ID"
// @Param session_id query int true "session ID"
//
// @Success 200 {object} account.Account "Account successfully retrieved"
//
// @Failure 400 {object} error "Bad Request"
// @Failure 401 {object} error "Unauthorized"
// @Failure 404 {object} error "Not Found"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/get_account [get]
func GetAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeAccountStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	queryParams := r.URL.Query()
	accountIdStr := queryParams.Get("account_id")
	if accountIdStr == "" {
		return writeAccountStatusError(w, http.StatusBadRequest, "account_id is required")
	}

	accountId, err := strconv.ParseInt(accountIdStr, 10, 64)
	if err != nil {
		return writeAccountStatusError(w, http.StatusBadRequest, "invalid account_id")
	}

	_, err = auth.RequireSessionFromQuery(w, r, store)
	if err != nil {
		return err
	}

	var (
		name            string
		info            string
		location        string
		email           string
		experienceLevel int
	)

	if err := db.QueryRow("SELECT name, info, location, email, experience_level FROM account WHERE id = $1", accountId).
		Scan(&name, &info, &location, &email, &experienceLevel); err != nil {
		if err == sql.ErrNoRows {
			return writeAccountStatusError(w, http.StatusNotFound, fmt.Sprintf("no account exists with the ID %d", accountId))
		}
		return fmt.Errorf("an error occurred while getting the account with the ID %d: %v", accountId, err)
	}

	account := Account{
		Name:            name,
		Info:            info,
		Location:        location,
		Email:           email,
		ExperienceLevel: ExperienceLevel(experienceLevel),
	}

	return writeAccountJSONResponse(w, http.StatusOK, account)
}
