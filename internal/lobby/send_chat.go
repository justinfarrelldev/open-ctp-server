package lobby

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type SendChatArgs struct {
	Message   string `json:"message"`
	SessionId *int64 `json:"session_id"`
}

func (a SendChatArgs) GetSessionID() *int64 {
	return a.SessionId
}

func SendChat(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) error {
	if r.Method != http.MethodPost {
		return writeLobbyStatusError(w, http.StatusBadRequest, "invalid request; request must be a POST request")
	}

	args := SendChatArgs{}
	if err := decodeJSONBody(r, &args); err != nil {
		return writeLobbyStatusError(w, http.StatusBadRequest, "an error occurred while decoding the request body:"+err.Error())
	}

	message := strings.TrimSpace(args.Message)
	if message == "" {
		return writeLobbyStatusError(w, http.StatusBadRequest, "message must be specified")
	}

	session, err := requireLobbySession(w, store, args)
	if err != nil {
		return err
	}

	tx, err := beginLobbyTx(db)
	if err != nil {
		return fmt.Errorf("an error occurred while starting lobby chat send: %v", err)
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
	if openLobbyID == nil {
		return writeLobbyStatusError(w, http.StatusConflict, "account is not currently in an open lobby")
	}

	lobbyRecord, err := getLobbyByID(tx, *openLobbyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return writeLobbyStatusError(w, http.StatusNotFound, fmt.Sprintf("no lobby exists with the ID %d", *openLobbyID))
		}
		return err
	}
	if err := closeLobbyIfInactive(tx, lobbyRecord, timeNow()); err != nil {
		return err
	}
	if lobbyRecord.ClosedAt != nil || lobbyRecord.IsClosed {
		return writeLobbyStatusError(w, http.StatusGone, fmt.Sprintf("lobby %d is closed", *openLobbyID))
	}
	if lobbyRecord.IsMuted {
		return writeLobbyStatusError(w, http.StatusForbidden, fmt.Sprintf("lobby %d is muted", *openLobbyID))
	}

	var messageID int64
	var createdAt time.Time
	if err := tx.QueryRowx("INSERT INTO lobby_messages (lobby_id, account_id, message) VALUES ($1, $2, $3) RETURNING id, created_at", *openLobbyID, accountID, message).Scan(&messageID, &createdAt); err != nil {
		return fmt.Errorf("an error occurred while storing chat message for lobby %d: %v", *openLobbyID, err)
	}
	if _, err := tx.Exec("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1", *openLobbyID); err != nil {
		return fmt.Errorf("an error occurred while updating lobby activity for the ID %d: %v", *openLobbyID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing lobby chat send: %v", err)
	}
	committed = true

	_ = messageID
	_ = createdAt
	lobbyBrokerInstance.publish(*openLobbyID, lobbyEvent{
		name: "chat_message",
		payload: lobbyEventPayload{
			LobbyID:   *openLobbyID,
			AccountID: accountID,
			Message:   message,
		},
	})

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Successfully sent chat message!"))
	return nil
}
