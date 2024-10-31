package account

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"

	argon2 "golang.org/x/crypto/argon2"
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

// HashSalt represents a salt and a hash in the same data type for password storage.
//
// @Description Structure containing both a salt and a hash for password storage.
type HashSalt struct {
	hash []byte
	salt []byte
}

const ERROR_PASSWORD_TOO_SHORT = "password must be longer than 6 characters"
const ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD = "password is required"

func isEmailValid(email string, db *sql.DB) (bool, error) {
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

var Hasher = NewArgon2idHash(1, 32, 64*1024, 32, 256)

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
func CreateAccount(w http.ResponseWriter, r *http.Request, db *sql.DB) error {

	if r.Method != "POST" {
		return errors.New("invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	account := CreateAccountArgs{}
	err := decoder.Decode(&account)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if account.Account.ExperienceLevel < 0 || account.Account.ExperienceLevel > 5 {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New("experience_level must be between 0 and 5 (0=easy, 5=impossible)")
	}

	if account.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}

	if len(account.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_TOO_SHORT)
	}

	isValidEmail, err := isEmailValid(account.Account.Email, db)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return err
	}

	if !isValidEmail {
		return errors.New("the provided email is not valid")
	}

	hashSalt, err := Hasher.GenerateHash([]byte(account.Password), nil)
	if err != nil {
		log.Println("error saving a password: ", err.Error())
		return errors.New("an error occurred while saving the password. Please try again later")
	}

	err = storeAccount(&account.Account, db)
	if err != nil {
		log.Println("error saving an account: ", err.Error())
		// Different from the one above for debugging purposes
		return errors.New("an error occurred while creating the account. Please try again at a later time")
	}

	err = storeHashAndSalt(hashSalt, account.Account.Email, db)
	if err != nil {
		log.Println("error saving a password: ", err.Error())
		// Different from the one above for debugging purposes
		return errors.New("an error occurred while saving the password. Please try again at a later time")
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Println("Successfully created account!")
	return nil
}

type Argon2idHash struct {
	// time represents the number of
	// passed over the specified memory.
	time uint32

	// cpu memory to be used.
	memory uint32

	// threads for parallelism aspect
	// of the algorithm.
	threads uint8

	// keyLen of the generate hash key.
	keyLen uint32

	// saltLen the length of the salt used.
	saltLen uint32
}

// NewArgon2idHash constructor function for
// Argon2idHash.
func NewArgon2idHash(time, saltLen uint32, memory uint32, threads uint8, keyLen uint32) *Argon2idHash {
	return &Argon2idHash{
		time:    time,
		saltLen: saltLen,
		memory:  memory,
		threads: threads,
		keyLen:  keyLen,
	}
}

// Salting for Argon2idHash.
func randomSecret(length uint32) ([]byte, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// GenerateHash using the password and provided salt.
// If not salt value provided fallback to random value
// generated of a given length.
func (a *Argon2idHash) GenerateHash(password, salt []byte) (*HashSalt, error) {
	var err error

	// If salt is not provided generate a salt of
	// the configured salt length.
	if len(salt) == 0 {
		salt, err = randomSecret(a.saltLen)
	}

	if err != nil {
		return nil, err
	}

	// Generate hash
	hash := argon2.IDKey(password, salt, a.time, a.memory, a.threads, a.keyLen)

	// Return the generated hash and salt used for storage.
	return &HashSalt{hash: hash, salt: salt}, nil

}

// Compare generated hash with store hash.
func (a *Argon2idHash) Compare(hash, salt, password []byte) error {
	// Generate hash for comparison.
	hashSalt, err := a.GenerateHash(password, salt)

	if err != nil {
		return err
	}

	// Compare the generated hash with the stored hash.
	// If they don't match return error.
	if !bytes.Equal(hash, hashSalt.hash) {
		return errors.New("hash doesn't match")

	}
	return nil
}

func storeHashAndSalt(hashSalt *HashSalt, accountEmail string, db *sql.DB) error {
	result, err := db.Query("INSERT INTO passwords (account_email, hash, salt) VALUES ($1, $2, $3)", accountEmail, base64.StdEncoding.EncodeToString(hashSalt.hash), base64.StdEncoding.EncodeToString(hashSalt.salt))
	if err != nil {
		return errors.New("an error occurred while inserting a hash-salt pair into the database: " + err.Error())
	}

	defer result.Close()
	return nil
}

func storeAccount(account *Account, db *sql.DB) error {
	result, err := db.Query("INSERT INTO account (name, info, location, email, experience_level) VALUES ($1, $2, $3, $4, $5)", account.Name, account.Info, account.Location, account.Email, account.ExperienceLevel)
	if err != nil {
		return errors.New("an error occurred while inserting an account into the database: " + err.Error())
	}

	defer result.Close()
	return nil
}
