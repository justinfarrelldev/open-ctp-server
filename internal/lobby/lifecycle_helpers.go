package lobby

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

const lobbySelectByID = `SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`

const openLobbyForAccountQuery = `SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`

func beginLobbyTx(db *sqlx.DB) (*sqlx.Tx, error) {
	return db.Beginx()
}

func decodeJSONBody(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func requireLobbySession(w http.ResponseWriter, store *auth.SessionStore, body auth.SessionIDBody) (*auth.Session, error) {
	return auth.RequireSessionFromJSONBody(w, store, body)
}

func getLobbyByID(tx *sqlx.Tx, lobbyID int64) (*Lobby, error) {
	var lobby Lobby
	if err := tx.Get(&lobby, lobbySelectByID, lobbyID); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("an error occurred while getting the lobby with the ID %d: %v", lobbyID, err)
	}

	return &lobby, nil
}

func getLobbyByIDNoTx(db *sqlx.DB, lobbyID int64) (*Lobby, error) {
	tx, err := beginLobbyTx(db)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	lobby, err := getLobbyByID(tx, lobbyID)
	if err != nil {
		return nil, err
	}

	if err := closeLobbyIfInactive(tx, lobby, time.Now()); err != nil {
		return nil, err
	}
	if lobby.ClosedAt != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("an error occurred while finalizing lobby close: %v", err)
		}
		committed = true
		return lobby, nil
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("an error occurred while finalizing lobby read: %v", err)
	}
	committed = true
	return lobby, nil
}

func findOpenLobbyForAccount(tx *sqlx.Tx, accountID int64) (*int64, error) {
	var lobbyIDs []int64
	if err := tx.Select(&lobbyIDs, openLobbyForAccountQuery, accountID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("an error occurred while checking current lobby membership for account %d: %v", accountID, err)
	}

	now := timeNow()
	for _, lobbyID := range lobbyIDs {
		lobby, err := getLobbyByID(tx, lobbyID)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}

		if err := closeLobbyIfInactive(tx, lobby, now); err != nil {
			return nil, err
		}
		if lobby.ClosedAt == nil && !lobby.IsClosed {
			return &lobbyID, nil
		}
	}

	return nil, nil
}

func ensureLobbyOwner(w http.ResponseWriter, session *auth.Session, lobby *Lobby) error {
	if int64(session.AccountID) != lobby.OwnerAccountID {
		return writeLobbyStatusError(w, http.StatusForbidden, "only the lobby owner may perform this action")
	}
	return nil
}

func closeLobbyIfInactive(tx *sqlx.Tx, lobby *Lobby, now time.Time) error {
	if lobby.ClosedAt != nil {
		return nil
	}

	connectedSubscriberCount := lobbyBrokerInstance.subscriberCount(lobby.ID)
	if !lobby.ShouldCloseForInactivity(now, connectedSubscriberCount) {
		return nil
	}

	if _, err := tx.Exec("UPDATE lobbies SET is_closed = TRUE, closed_at = NOW(), updated_at = NOW() WHERE id = $1 AND closed_at IS NULL", lobby.ID); err != nil {
		return fmt.Errorf("an error occurred while closing inactive lobby %d: %v", lobby.ID, err)
	}

	closedAt := now
	lobby.IsClosed = true
	lobby.ClosedAt = &closedAt
	return nil
}

func writeLobbyJSON(w http.ResponseWriter, lobby *Lobby) error {
	body, err := json.Marshal(lobby)
	if err != nil {
		return fmt.Errorf("error marshalling lobby: %v", err)
	}

	_, _ = w.Write(body)
	return nil
}

func lookupAccountName(tx *sqlx.Tx, accountID int64) (string, error) {
	var name string
	if err := tx.Get(&name, "SELECT name FROM account WHERE id = $1", accountID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no account exists with the ID %d", accountID)
		}
		return "", fmt.Errorf("an error occurred while getting the account with the ID %d: %v", accountID, err)
	}
	return name, nil
}
