package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// CreateAccountArgs represents the expected structure of the request body for creating an account for use within the server.
//
// @Description Structure for the account creation request payload.
type CreateAccountArgs struct {
	// The account to create.
	Account Account `json:"account"`
	// The password for the account to be created
	Password string `json:"password"`
}

const ERROR_PASSWORD_TOO_SHORT = "password must be longer than 6 characters"
const ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD = "password is required"

// CreateAccount handles the creation of a new account.
//
// @Summary Create a new account
// @Description This endpoint creates a new multiplayer account, protected by a password.
// @Tags account
// @Accept json
// @Produce json
// @Param body body CreateAccountArgs true "account creation request body"
// @Success 201 {integer} int64 "Account successfully created"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/create_account [post]
func CreateAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) (*int64, error) {
	if r.Method != http.MethodPost {
		return nil, writeAccountStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	account := CreateAccountArgs{}
	if err := decoder.Decode(&account); err != nil {
		return nil, writeAccountStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}

	if err := validateExperienceLevel(account.Account.ExperienceLevel); err != nil {
		return nil, writeAccountStatusError(w, http.StatusBadRequest, err.Error())
	}

	if account.Password == "" {
		return nil, writeAccountStatusError(w, http.StatusBadRequest, ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}

	if len(account.Password) < 6 {
		return nil, writeAccountStatusError(w, http.StatusBadRequest, ERROR_PASSWORD_TOO_SHORT)
	}

	hashSalt, err := auth.Hasher.GenerateHash([]byte(account.Password), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error saving a password: ", err.Error())
		return nil, errors.New("an error occurred while saving the password. Please try again later")
	}

	if err := validateEmail(account.Account.Email, nil, db); err != nil {
		return nil, writeAccountStatusError(w, http.StatusBadRequest, err.Error())
	}

	tx, err := beginAccountTx(db)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while starting account creation: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	accountID, err := storeAccount(&account.Account, tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error saving an account: ", err.Error())
		return nil, errors.New("an error occurred while creating the account. Please try again at a later time")
	}

	if err := auth.StoreHashAndSalt(hashSalt, account.Account.Email, tx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error saving a password: ", err.Error())
		return nil, errors.New("an error occurred while saving the password. Please try again at a later time")
	}

	session, err := store.CreateSessionTx(tx, *accountID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error creating a session: ", err.Error())
		return nil, errors.New("an error occurred while creating a session. Please try again at a later time")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("an error occurred while finalizing account creation: %v", err)
	}
	committed = true

	w.Header().Set("X-Account-ID", fmt.Sprintf("%d", *accountID))
	if err := writeAccountJSONResponse(w, http.StatusCreated, session.ID); err != nil {
		return nil, err
	}

	return &session.ID, nil
}

func storeAccount(account *Account, db accountQueryExecutor) (accountId *int, err error) {
	var id int
	err = db.QueryRow("INSERT INTO account (name, info, location, email, experience_level) VALUES ($1, $2, $3, $4, $5) RETURNING id", account.Name, account.Info, account.Location, account.Email, account.ExperienceLevel).Scan(&id)
	if err != nil {
		return nil, errors.New("an error occurred while inserting an account into the database: " + err.Error())
	}

	return &id, nil
}
