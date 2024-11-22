package lobby

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"

	argon2 "golang.org/x/crypto/argon2"
)

// CreateLobbyArgs represents the expected structure of the request body for creating a lobby for use within the server.
//
// @Description Structure for the lobby creation request payload.
type CreateLobbyArgs struct {
	// The lobby to create.
	Lobby Lobby `json:"lobby"`
	// The password for the lobby to be created
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
		return false, errors.New("an error occurred while checking whether the email for the lobby is valid: " + err.Error())
	}

	result, err := db.Query("SELECT * from lobby WHERE email = $1", email)

	if err != nil {
		return false, errors.New("an error occurred while checking whether the email for the lobby is unique: " + err.Error())
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

// CreateLobby handles the creation of a new lobby.
//
// @Summary Create a new lobby
// @Description This endpoint creates a new multiplayer lobby, protected by a password.
// @Tags lobby
// @Accept json
// @Produce json
// @Param body body CreateLobbyArgs true "lobby creation request body"
// @Success 201 {string} string "Lobby successfully created"
// @Failure 400 {object} error "Bad Request"
// @Failure 403 {object} error "Forbidden"
// @Failure 500 {object} error "Internal Server Error"
// @Router /lobby/create_lobby [post]
func CreateLobby(w http.ResponseWriter, r *http.Request, db *sql.DB) error {

	if r.Method != "POST" {
		return errors.New("invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	lobby := CreateLobbyArgs{}
	err := decoder.Decode(&lobby)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return errors.New("an error occurred while decoding the request body:" + err.Error())
	}

	if lobby.Password == "" {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_REQUIRED_BUT_NO_PASSWORD)
	}

	if len(lobby.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)

		return errors.New(ERROR_PASSWORD_TOO_SHORT)
	}

	// TODO store password if needed
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

func storeHashAndSalt(hashSalt *HashSalt, lobbyEmail string, db *sql.DB) error {
	result, err := db.Query("INSERT INTO passwords (lobby_email, hash, salt) VALUES ($1, $2, $3)", lobbyEmail, base64.StdEncoding.EncodeToString(hashSalt.hash), base64.StdEncoding.EncodeToString(hashSalt.salt))
	if err != nil {
		return errors.New("an error occurred while inserting a hash-salt pair into the database: " + err.Error())
	}

	defer result.Close()
	return nil
}

func storeLobby(lobby *Lobby, db *sql.DB) error {
	result, err := db.Query(
		"INSERT INTO lobby (id, name, owner_name, is_closed, is_muted, is_public) VALUES ($1, $2, $3, $4, $5, $6)",
		lobby.ID, lobby.Name, lobby.OwnerName, lobby.IsClosed, lobby.IsMuted, lobby.IsPublic,
	)
	if err != nil {
		return errors.New("an error occurred while inserting a lobby into the database: " + err.Error())
	}

	defer result.Close()
	return nil
}
