package auth

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// hashSalt represents a salt and a hash in the same data type for password storage.
//
// @Description Structure containing both a salt and a hash for password storage.
type hashSalt struct {
	Hash []byte
	Salt []byte
}

type credentialQueryExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type argon2idHash struct {
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
func NewArgon2idHash(time, saltLen uint32, memory uint32, threads uint8, keyLen uint32) *argon2idHash {
	return &argon2idHash{
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
func (a *argon2idHash) GenerateHash(password, salt []byte) (*hashSalt, error) {
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
	return &hashSalt{Hash: hash, Salt: salt}, nil

}

// Compare generated hash with store hash.
func (a *argon2idHash) Compare(hash, salt, password []byte) error {
	// Generate hash for comparison.
	hashSalt, err := a.GenerateHash(password, salt)

	if err != nil {
		return err
	}

	// Compare the generated hash with the stored hash.
	// If they don't match return error.
	if !bytes.Equal(hash, hashSalt.Hash) {
		return errors.New("hash doesn't match")

	}
	return nil
}

func StoreHashAndSalt(hashSalt *hashSalt, accountEmail string, db credentialQueryExecutor) error {
	result, err := db.Query("INSERT INTO passwords (account_email, hash, salt) VALUES ($1, $2, $3)", accountEmail, base64.StdEncoding.EncodeToString(hashSalt.Hash), base64.StdEncoding.EncodeToString(hashSalt.Salt))
	if err != nil {
		return errors.New("an error occurred while inserting a hash-salt pair into the database: " + err.Error())
	}

	defer result.Close()
	return nil
}

func GetHashAndSalt(accountEmail string, db credentialQueryExecutor) ([]byte, []byte, error) {
	var storedHash string
	var storedSalt string

	err := db.QueryRow("SELECT hash, salt FROM passwords WHERE account_email = $1", accountEmail).Scan(&storedHash, &storedSalt)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving account credentials: %w", err)
	}

	storedHashBytes, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding stored hash: %w", err)
	}

	storedSaltBytes, err := base64.StdEncoding.DecodeString(storedSalt)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding stored salt: %w", err)
	}

	return storedHashBytes, storedSaltBytes, nil
}

func UpdateHashAndSalt(hashSalt *hashSalt, accountEmail string, db credentialQueryExecutor) error {
	_, err := db.Exec("UPDATE passwords SET hash = $1, salt = $2 WHERE account_email = $3", base64.StdEncoding.EncodeToString(hashSalt.Hash), base64.StdEncoding.EncodeToString(hashSalt.Salt), accountEmail)
	if err != nil {
		return errors.New("an error occurred while updating a hash-salt pair in the database: " + err.Error())
	}

	return nil
}

var Hasher = NewArgon2idHash(1, 32, 64*1024, 32, 256)
