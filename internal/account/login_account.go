package account

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type LoginAccountArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginAccountResponse struct {
	SessionID int64     `json:"session_id"`
	AccountID int64     `json:"account_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func LoginAccount(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeAccountStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	args := LoginAccountArgs{}
	if err := decoder.Decode(&args); err != nil {
		return writeAccountStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body: "+err.Error())
	}

	args.Email = strings.TrimSpace(args.Email)
	if args.Email == "" || args.Password == "" {
		return writeAccountStatusError(w, http.StatusBadRequest, "email and password are required")
	}

	var accountID int64
	var encodedHash string
	var encodedSalt string
	err := db.QueryRow("SELECT a.id, p.hash, p.salt FROM account a JOIN passwords p ON p.account_email = a.email WHERE a.email = $1", args.Email).Scan(&accountID, &encodedHash, &encodedSalt)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeAccountStatusError(w, http.StatusUnauthorized, ErrorInvalidEmailOrPassword)
		}
		return fmt.Errorf("an error occurred while retrieving login credentials: %v", err)
	}

	storedHash, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return fmt.Errorf("error decoding stored hash: %v", err)
	}

	storedSalt, err := base64.StdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return fmt.Errorf("error decoding stored salt: %v", err)
	}

	if err := auth.Hasher.Compare(storedHash, storedSalt, []byte(args.Password)); err != nil {
		return writeAccountStatusError(w, http.StatusUnauthorized, ErrorInvalidEmailOrPassword)
	}

	session, err := store.CreateSession(int(accountID))
	if err != nil {
		return fmt.Errorf("an error occurred while creating a session: %v", err)
	}

	response := LoginAccountResponse{
		SessionID: session.ID,
		AccountID: accountID,
		ExpiresAt: session.ExpiresAt,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding the login response: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseBytes)
	return nil
}
