package auth

import (
	"errors"
	"net/http"
	"strconv"
)

var ErrSessionIDRequired = errors.New("session_id is required")
var ErrInvalidSessionID = errors.New("invalid session_id")
var ErrInvalidBodySessionID = errors.New("a valid session_id must be specified")
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionExpired = errors.New("session has expired")

type SessionIDBody interface {
	GetSessionID() *int64
}

func SessionIDFromJSONBody(body SessionIDBody) (int64, error) {
	sessionID := body.GetSessionID()
	if sessionID == nil || *sessionID <= 0 {
		return 0, ErrInvalidBodySessionID
	}

	return *sessionID, nil
}

func SessionIDFromQuery(r *http.Request) (int64, error) {
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		return 0, ErrSessionIDRequired
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil || sessionID <= 0 {
		return 0, ErrInvalidSessionID
	}

	return sessionID, nil
}

func RequireSessionFromJSONBody(w http.ResponseWriter, store *SessionStore, body SessionIDBody) (*Session, error) {
	sessionID, err := SessionIDFromJSONBody(body)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, err
	}

	return RequireSession(w, store, sessionID)
}

func RequireSessionFromQuery(w http.ResponseWriter, r *http.Request, store *SessionStore) (*Session, error) {
	sessionID, err := SessionIDFromQuery(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, err
	}

	return RequireSession(w, store, sessionID)
}

func RequireSession(w http.ResponseWriter, store *SessionStore, sessionID int64) (*Session, error) {
	session, err := store.GetSession(sessionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, errors.New("an error occurred while retrieving the session: " + err.Error())
	}

	if session == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, ErrSessionExpired
	}

	return session, nil
}
