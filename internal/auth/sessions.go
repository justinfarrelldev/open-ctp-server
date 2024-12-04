package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/jmoiron/sqlx"
)

// Session represents a user session
type Session struct {
	ID        string    `db:"id" json:"id"`
	AccountID int       `db:"account_id" json:"account_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}

// SessionStore handles session-related database operations
type SessionStore struct {
	db *sqlx.DB
}

// NewSessionStore creates a new SessionStore
func NewSessionStore(db *sqlx.DB) *SessionStore {
	return &SessionStore{db: db}
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
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        sessionID,
		AccountID: accountID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(12 * time.Hour), // Session expires in 12 hours
	}

	query := `INSERT INTO sessions (id, account_id, created_at, expires_at) VALUES (:id, :account_id, :created_at, :expires_at)`
	_, err = s.db.NamedExec(query, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session by its ID
// @Summary Get a session
// @Description Get a session by its ID
// @Tags sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Session
// @Failure 404 {object} error
// @Router /sessions/{id} [get]
func (s *SessionStore) GetSession(sessionID string) (*Session, error) {
	var session Session
	query := `SELECT * FROM sessions WHERE id = $1`
	err := s.db.Get(&session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
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
func (s *SessionStore) DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := s.db.Exec(query, sessionID)
	if err != nil {
		return err
	}
	return nil
}

// generateSessionID generates a unique session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
