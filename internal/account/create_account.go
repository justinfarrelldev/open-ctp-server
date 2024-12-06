package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"

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

func isEmailValid(email string, db *sqlx.DB) (bool, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false, errors.New("an error occurred while checking whether the email for the account is valid: " + err.Error())
	}

	result, err := db.Query("SELECT * from account WHERE email = $1", email)

	if err != nil {
		return false, errors.New("an error occurred while checking whether the email for the account is unique: " + err.Error())
	}

	defer result.Close()

	if result.Next() {
		// If result.Next() returns true, there is at least one row, so the email is not unique.
		return false, nil
	}

	// If no rows were found, the email is unique (or not in the database).
	return true, nil
}

// CreateAccount handles the creation of a new account.
//
// @Summary Create a new account
// @Description This endpoint creates a new multiplayer account, protected by a password.
// @Tags account
// @Accept json
// @Produce json
// @Param body body CreateAccountArgs true "account creation request body"
// @Success 201 {string} string "Account successfully created"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /account/create_account [post]
func CreateAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) (*string, error) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)

		return nil, errors.New("invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	account := CreateAccountArgs{}
	err := decoder.Decode(&account)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return nil, errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if account.Account.ExperienceLevel < 0 || account.Account.ExperienceLevel > 5 {
		w.WriteHeader(http.StatusBadRequest)

		return nil, errors.New("experience_level must be between 0 and 5 (0=easy, 5=impossible)")
	}

	if account.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		return nil, errors.New(ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}

	if len(account.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)

		return nil, errors.New(ERROR_PASSWORD_TOO_SHORT)
	}

	isValidEmail, err := isEmailValid(account.Account.Email, db)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	if !isValidEmail {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("the provided email is not valid")
	}

	hashSalt, err := auth.Hasher.GenerateHash([]byte(account.Password), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error saving a password: ", err.Error())
		return nil, errors.New("an error occurred while saving the password. Please try again later")
	}

	accountId, err := storeAccount(&account.Account, db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error saving an account: ", err.Error())
		// Different from the one above for debugging purposes
		return nil, errors.New("an error occurred while creating the account. Please try again at a later time")
	}

	err = auth.StoreHashAndSalt(hashSalt, account.Account.Email, db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Println("error saving a password: ", err.Error())
		// Different from the one above for debugging purposes
		return nil, errors.New("an error occurred while saving the password. Please try again at a later time")
	}

	session, err := store.CreateSession(*accountId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Println("error creating a session: ", err.Error())
		return nil, errors.New("an error occurred while creating a session. Please try again at a later time")
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Println("Successfully created account!")
	return &session.ID, nil
}

func storeAccount(account *Account, db *sqlx.DB) (accountId *int, err error) {
	var id int
	err = db.QueryRow("INSERT INTO account (name, info, location, email, experience_level) VALUES ($1, $2, $3, $4, $5) RETURNING id", account.Name, account.Info, account.Location, account.Email, account.ExperienceLevel).Scan(&id)
	if err != nil {
		return nil, errors.New("an error occurred while inserting an account into the database: " + err.Error())
	}

	return &id, nil
}
