package lobby

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

func ConnectLobby(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodGet {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a GET request")
	}

	lobbyIDStr := r.URL.Query().Get("lobby_id")
	lobbyID, err := strconv.ParseInt(lobbyIDStr, 10, 64)
	if err != nil || lobbyID <= 0 {
		return writeLobbyStatusError(w, http.StatusBadRequest, "lobby_id must be specified")
	}

	session, err := auth.RequireSessionFromQuery(w, r, store)
	if err != nil {
		return err
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return writeLobbyStatusError(w, http.StatusInternalServerError, "response writer does not support streaming")
	}

	tx, err := beginLobbyTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting lobby stream connect: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	accountID := int64(session.AccountID)
	openLobbyID, err := findOpenLobbyForAccount(tx, accountID)
	if err != nil {
		return err
	}
	if openLobbyID != nil && *openLobbyID != lobbyID {
		return writeLobbyStatusError(w, http.StatusConflict, "account is already in an open lobby")
	}

	lobbyRecord, err := getLobbyByID(tx, lobbyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeLobbyStatusError(w, http.StatusNotFound, fmt.Sprintf("no lobby exists with the ID %d", lobbyID))
		}
		return err
	}
	if err := closeLobbyIfInactive(tx, lobbyRecord, timeNow()); err != nil {
		return err
	}
	if lobbyRecord.ClosedAt != nil || lobbyRecord.IsClosed {
		return writeLobbyStatusError(w, http.StatusGone, fmt.Sprintf("lobby %d is closed", lobbyID))
	}

	if openLobbyID == nil {
		if _, err := tx.Exec("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)", lobbyID, accountID); err != nil {
			return fmt.Errorf("an error occurred while joining the lobby with the ID %d for stream connect: %v", lobbyID, err)
		}
	}
	if _, err := tx.Exec("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1", lobbyID); err != nil {
		return fmt.Errorf("an error occurred while updating lobby activity for the ID %d: %v", lobbyID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing lobby stream connect: %v", err)
	}
	committed = true

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	events, unsubscribe := lobbyBrokerInstance.subscribe(lobbyID)
	defer unsubscribe()

	lobbyBrokerInstance.publish(lobbyID, lobbyEvent{
		name: "player_connected",
		payload: lobbyEventPayload{
			LobbyID:   lobbyID,
			AccountID: accountID,
		},
	})

	for {
		select {
		case <-r.Context().Done():
			return nil
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if err := writeLobbySSEEvent(w, event); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}
