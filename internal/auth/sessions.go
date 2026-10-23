package auth

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"time"

	"log"

	"github.com/jmoiron/sqlx"
)

// Session represents a user session
type Session struct {
	ID        int64     `db:"id" json:"id"`
	AccountID int       `db:"account_id" json:"account_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}

type sessionNamedExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	NamedExec(query string, arg any) (sql.Result, error)
}

type sessionExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// SessionStore handles session-related database operations. Session IDs are bearer
// credentials: never log them, Session structs, or raw database errors that may
// echo query parameters. Return errors privately; diagnostics use fixed labels.
type SessionStore struct {
	DB *sqlx.DB
}

// NewSessionStore creates a new SessionStore
func NewSessionStore(db *sqlx.DB) *SessionStore {
	if db == nil {
		log.Println("Database connection is nil")
	}
	return &SessionStore{DB: db}
}

func createSession(executor sessionNamedExecutor, accountID int) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		log.Print("Session entropy unavailable")
		return nil, err
	}

	session := &Session{
		ID:        sessionID,
		AccountID: accountID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(12 * time.Hour),
	}

	query := `INSERT INTO sessions (id, account_id, created_at, expires_at) VALUES (:id, :account_id, :created_at, :expires_at)`
	_, err = executor.NamedExec(query, session)
	if err != nil {
		log.Print("Session creation failed")
		return nil, err
	}

	return session, nil
}

func deleteSessionsForAccount(executor sessionExecutor, accountID int64) error {
	query := `DELETE FROM sessions WHERE account_id = $1`
	_, err := executor.Exec(query, accountID)
	if err != nil {
		log.Print("Account session deletion failed")
		return err
	}

	return nil
}

func deleteSessionsForAccountExcept(executor sessionExecutor, accountID int64, exceptSessionID int64) error {
	query := `DELETE FROM sessions WHERE account_id = $1 AND id <> $2`
	_, err := executor.Exec(query, accountID, exceptSessionID)
	if err != nil {
		log.Print("Other account session deletion failed")
		return err
	}

	return nil
}

// CreateSession creates a new session for a user
// @Summary Create a new session
// @Description Create a new session for a user. Expires 12 hours from last interaction.
// @Tags sessions
// @Accept json
// @Produce json
// @Param account_id body int true "User ID"
// @Success 200 {object} Session
// @Failure 400 {object} error
// @Router /sessions [post]
func (s *SessionStore) CreateSession(accountID int) (*Session, error) {
	if s.DB == nil {
		log.Println("Database connection is nil")
		return nil, errors.New("Database connection is nil")
	}

	return createSession(s.DB, accountID)
}

func (s *SessionStore) CreateSessionTx(tx *sqlx.Tx, accountID int) (*Session, error) {
	return createSession(tx, accountID)
}

// GetSession retrieves a session by its ID
// @Summary Get a session
// @Description Get a session by its ID
// @Tags sessions
// @Accept json
// @Produce json
// @Param id path int64 true "Session ID"
// @Success 200 {object} Session
// @Failure 404 {object} error
// @Router /sessions/{id} [get]
func (s *SessionStore) GetSession(sessionID int64) (*Session, error) {
	var session Session
	query := `SELECT * FROM sessions WHERE id = $1`
	err := s.DB.Get(&session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Print("Session lookup failed")
		return nil, err
	}

	return &session, nil
}

// DeleteSession deletes a session by its ID. Should not be exposed to end users via any API endpoints.
// This should instead be used to invalidate any sessions when, for example, a user logs out or an account is deleted.
// @Summary Delete a session
// @Description Delete a session by its ID
// @Tags sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Success 204
// @Failure 404 {object} error
// @Router /sessions/{id} [delete]
func (s *SessionStore) DeleteSession(sessionID int64) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := s.DB.Exec(query, sessionID)
	if err != nil {
		log.Print("Session deletion failed")
		return err
	}

	return nil
}

func (s *SessionStore) DeleteSessionsForAccount(accountID int64) error {
	return deleteSessionsForAccount(s.DB, accountID)
}

func (s *SessionStore) DeleteSessionsForAccountTx(tx *sqlx.Tx, accountID int64) error {
	return deleteSessionsForAccount(tx, accountID)
}

func (s *SessionStore) DeleteSessionsForAccountExcept(accountID int64, exceptSessionID int64) error {
	return deleteSessionsForAccountExcept(s.DB, accountID, exceptSessionID)
}

func (s *SessionStore) DeleteSessionsForAccountExceptTx(tx *sqlx.Tx, accountID int64, exceptSessionID int64) error {
	return deleteSessionsForAccountExcept(tx, accountID, exceptSessionID)
}

// generateSessionID creates a cryptographically secure random session ID
// by generating 8 random bytes and converting them to a 64-bit integer.
func generateSessionID() (int64, error) {
	for {
		// Create a byte array to store our random bytes
		var randomBytes [8]byte

		// Read 8 random bytes using crypto/rand.
		// Clear the sign bit so the resulting int64 stays positive.
		if _, err := rand.Read(randomBytes[:]); err != nil {
			return 0, err
		}
		randomBytes[0] &= 0x7F

		// Convert 8 bytes to int64 by shifting each byte to its position
		// and combining them with bitwise OR operations.
		sessionID := int64(0)
		for i, b := range randomBytes {
			shift := uint(56 - (i * 8)) // 56, 48, 40, 32, 24, 16, 8, 0
			sessionID |= int64(b) << shift
		}

		if sessionID > 0 {
			return sessionID, nil
		}
	}
}

// IsExpired checks if the session has expired
// @Summary Check if session is expired
// @Description Returns true if the session has expired, false otherwise
// @Tags sessions
// @Accept json
// @Produce json
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
