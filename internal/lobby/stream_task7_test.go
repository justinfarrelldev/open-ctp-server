package lobby

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type sseTestWriter struct {
	mu     sync.Mutex
	header http.Header
	status int
	writes chan string
}

func newSSETestWriter() *sseTestWriter {
	return &sseTestWriter{
		header: make(http.Header),
		writes: make(chan string, 16),
	}
}

func (w *sseTestWriter) Header() http.Header {
	return w.header
}

func (w *sseTestWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status == 0 {
		w.status = statusCode
	}
}

func (w *sseTestWriter) Write(body []byte) (int, error) {
	w.mu.Lock()
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.mu.Unlock()
	w.writes <- string(body)
	return len(body), nil
}

func (w *sseTestWriter) Flush() {}

func (w *sseTestWriter) StatusCode() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func waitForStreamChunk(t *testing.T, writes <-chan string, want string) string {
	t.Helper()

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case chunk := <-writes:
			if strings.Contains(chunk, want) {
				return chunk
			}
		case <-timeout.C:
			t.Fatalf("timed out waiting for SSE chunk containing %q", want)
		}
	}
}

func expectOpenLobbyLookupNoRows(mock sqlmock.Sqlmock, accountID int64) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}))
}

func expectOpenLobbyLookup(mock sqlmock.Sqlmock, accountID int64, lobbyID int64) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT lm.lobby_id
FROM lobby_members lm
JOIN lobbies l ON l.id = lm.lobby_id
WHERE lm.account_id = $1 AND l.closed_at IS NULL
ORDER BY lm.joined_at ASC, lm.id ASC`)).
		WithArgs(accountID).
		WillReturnRows(sqlmock.NewRows([]string{"lobby_id"}).AddRow(lobbyID))
}

func expectLobbyLookup(mock sqlmock.Sqlmock, lobbyID int64, ownerAccountID int64, isMuted bool) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, owner_name, owner_account_id, is_closed, is_muted, is_public, created_at, updated_at, last_activity_at, closed_at
FROM lobby
WHERE id = $1`)).
		WithArgs(lobbyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "owner_name", "owner_account_id", "is_closed", "is_muted", "is_public", "created_at", "updated_at", "last_activity_at", "closed_at"}).
			AddRow(lobbyID, "Task 7 Lobby", "Owner", ownerAccountID, false, isMuted, true, time.Now(), time.Now(), time.Now(), nil))
}

func resetLobbyBroker() {
	lobbyBrokerInstance = newLobbyBroker()
}

func brokerSubscriberCount(b *lobbyBroker, lobbyID int64) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subscribers[lobbyID])
}

func TestConnectLobby_RejectsMissingSessionID(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, _, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/lobby/connect_lobby?lobby_id=18", nil)
	rr := httptest.NewRecorder()

	err := ConnectLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestConnectLobby_SetsSSEHeadersJoinsAndEmitsPlayerConnected(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(401)
	accountID := int64(91)
	lobbyID := int64(18)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookupNoRows(mock, accountID)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)")).
		WithArgs(lobbyID, accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/lobby/connect_lobby?session_id=401&lobby_id=18", nil).WithContext(ctx)
	writer := newSSETestWriter()
	done := make(chan error, 1)
	go func() {
		done <- ConnectLobby(writer, req, sqlxDB, store)
	}()

	chunk := waitForStreamChunk(t, writer.writes, "event: player_connected")
	if !strings.Contains(chunk, `"account_id":91`) {
		t.Fatalf("expected account_id in event payload, got %q", chunk)
	}
	if writer.StatusCode() != http.StatusOK {
		t.Fatalf("got status %d want %d", writer.StatusCode(), http.StatusOK)
	}
	if got := writer.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("got content type %q want %q", got, "text/event-stream")
	}
	if got := writer.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("got cache-control %q want %q", got, "no-cache")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connect lobby did not exit after cancel")
	}

	if brokerSubscriberCount(lobbyBrokerInstance, lobbyID) != 0 {
		t.Fatalf("expected no subscribers after disconnect, got %d", brokerSubscriberCount(lobbyBrokerInstance, lobbyID))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestJoinLobby_PublishesPlayerConnectedToExistingSubscriber(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	listenerSessionID := int64(409)
	joinSessionID := int64(410)
	listenerAccountID := int64(98)
	joinAccountID := int64(99)
	lobbyID := int64(25)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(listenerSessionID).
		WillReturnRows(lobbySessionRows(listenerSessionID, listenerAccountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookupNoRows(mock, listenerAccountID)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)")).
		WithArgs(lobbyID, listenerAccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamReq := httptest.NewRequest(http.MethodGet, "/lobby/connect_lobby?session_id=409&lobby_id=25", nil).WithContext(ctx)
	writer := newSSETestWriter()
	done := make(chan error, 1)
	go func() {
		done <- ConnectLobby(writer, streamReq, sqlxDB, store)
	}()

	firstChunk := waitForStreamChunk(t, writer.writes, "event: player_connected")
	if !strings.Contains(firstChunk, `"account_id":98`) {
		t.Fatalf("expected listener account in initial event payload, got %q", firstChunk)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(joinSessionID).
		WillReturnRows(lobbySessionRows(joinSessionID, joinAccountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookupNoRows(mock, joinAccountID)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)")).
		WithArgs(lobbyID, joinAccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	joinReq := httptest.NewRequest(http.MethodPost, "/lobby/join_lobby", strings.NewReader(`{"lobby_id":25,"session_id":410}`))
	joinRR := httptest.NewRecorder()
	if err := JoinLobby(joinRR, joinReq, sqlxDB, store); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if joinRR.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", joinRR.Code, http.StatusOK)
	}

	joinChunk := waitForStreamChunk(t, writer.writes, "event: player_connected")
	if !strings.Contains(joinChunk, `"account_id":99`) {
		t.Fatalf("expected joiner account in broadcast payload, got %q", joinChunk)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connect lobby did not exit after cancel")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestConnectLobby_AllowsTargetConnectAfterExpiredCurrentLobbyIsClosed(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(407)
	accountID := int64(96)
	expiredLobbyID := int64(22)
	targetLobbyID := int64(23)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectExpiredCurrentLobbyLookup(mock, accountID, expiredLobbyID)
	expectLobbyLookup(mock, targetLobbyID, int64(44), false)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO lobby_members (lobby_id, account_id) VALUES ($1, $2)")).
		WithArgs(targetLobbyID, accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(targetLobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamReq := httptest.NewRequest(http.MethodGet, "/lobby/connect_lobby?session_id=407&lobby_id=23", nil).WithContext(ctx)
	writer := newSSETestWriter()
	done := make(chan error, 1)
	go func() {
		done <- ConnectLobby(writer, streamReq, sqlxDB, store)
	}()

	chunk := waitForStreamChunk(t, writer.writes, "event: player_connected")
	if !strings.Contains(chunk, `"account_id":96`) {
		t.Fatalf("expected account_id in event payload, got %q", chunk)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connect lobby did not exit after cancel")
	}
}

func TestSendChat_RejectsMutedLobbyWithoutPersisting(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(402)
	accountID := int64(92)
	lobbyID := int64(19)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookup(mock, accountID, lobbyID)
	expectLobbyLookup(mock, lobbyID, int64(44), true)
	expectLobbyLookup(mock, lobbyID, int64(44), true)
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/lobby/send_chat", strings.NewReader(`{"session_id":402,"message":"hello"}`))
	rr := httptest.NewRecorder()

	err := SendChat(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusForbidden {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusForbidden)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestSendChat_PersistsMessageAndEmitsChatEvent(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	connectSessionID := int64(403)
	chatSessionID := int64(404)
	listenerAccountID := int64(93)
	chatAccountID := int64(94)
	lobbyID := int64(20)
	message := "hello from task 7"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(connectSessionID).
		WillReturnRows(lobbySessionRows(connectSessionID, listenerAccountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookup(mock, listenerAccountID, lobbyID)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamReq := httptest.NewRequest(http.MethodGet, "/lobby/connect_lobby?session_id=403&lobby_id=20", nil).WithContext(ctx)
	writer := newSSETestWriter()
	done := make(chan error, 1)
	go func() {
		done <- ConnectLobby(writer, streamReq, sqlxDB, store)
	}()
	waitForStreamChunk(t, writer.writes, "event: player_connected")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(chatSessionID).
		WillReturnRows(lobbySessionRows(chatSessionID, chatAccountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookup(mock, chatAccountID, lobbyID)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO lobby_messages (lobby_id, account_id, message) VALUES ($1, $2, $3) RETURNING id, created_at")).
		WithArgs(lobbyID, chatAccountID, message).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(501), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	chatReq := httptest.NewRequest(http.MethodPost, "/lobby/send_chat", strings.NewReader(`{"session_id":404,"message":"hello from task 7"}`))
	chatRR := httptest.NewRecorder()

	err := SendChat(chatRR, chatReq, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chatRR.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", chatRR.Code, http.StatusOK)
	}

	chunk := waitForStreamChunk(t, writer.writes, "event: chat_message")
	if !strings.Contains(chunk, message) {
		t.Fatalf("expected chat message in stream, got %q", chunk)
	}
	if !strings.Contains(chunk, `"account_id":94`) {
		t.Fatalf("expected sender account in event payload, got %q", chunk)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connect lobby did not exit after cancel")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestSendChat_ExpiredCurrentLobbyReturnsConflictAfterNormalization(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(408)
	accountID := int64(97)
	expiredLobbyID := int64(24)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectExpiredCurrentLobbyLookup(mock, accountID, expiredLobbyID)
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/lobby/send_chat", strings.NewReader(`{"session_id":408,"message":"hello"}`))
	rr := httptest.NewRecorder()

	err := SendChat(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusConflict {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusConflict)
	}
}

func TestLeaveLobby_ExpiredCurrentLobbyReturnsConflictAfterNormalization(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	sessionID := int64(409)
	accountID := int64(98)
	expiredLobbyID := int64(25)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(sessionID).
		WillReturnRows(lobbySessionRows(sessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectExpiredCurrentLobbyLookup(mock, accountID, expiredLobbyID)
	mock.ExpectRollback()

	req := httptest.NewRequest(http.MethodPost, "/lobby/leave_lobby", strings.NewReader(`{"session_id":409}`))
	rr := httptest.NewRecorder()

	err := LeaveLobby(rr, req, sqlxDB, store)
	if err == nil {
		t.Fatal("expected error")
	}
	if rr.Code != http.StatusConflict {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusConflict)
	}
}

func TestLeaveLobby_EmitsPlayerLeftAndDisconnectCleansUpSubscriber(t *testing.T) {
	resetLobbyBroker()

	sqlxDB, mock, store, cleanup := newLobbyTestStore(t)
	defer cleanup()

	connectSessionID := int64(405)
	leaveSessionID := int64(406)
	accountID := int64(95)
	lobbyID := int64(21)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(connectSessionID).
		WillReturnRows(lobbySessionRows(connectSessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookup(mock, accountID, lobbyID)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamReq := httptest.NewRequest(http.MethodGet, "/lobby/connect_lobby?session_id=405&lobby_id=21", nil).WithContext(ctx)
	writer := newSSETestWriter()
	done := make(chan error, 1)
	go func() {
		done <- ConnectLobby(writer, streamReq, sqlxDB, store)
	}()
	waitForStreamChunk(t, writer.writes, "event: player_connected")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM sessions WHERE id = $1")).
		WithArgs(leaveSessionID).
		WillReturnRows(lobbySessionRows(leaveSessionID, accountID, time.Now().Add(time.Hour)))
	mock.ExpectBegin()
	expectOpenLobbyLookup(mock, accountID, lobbyID)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	expectLobbyLookup(mock, lobbyID, int64(44), false)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM lobby_members WHERE lobby_id = $1 AND account_id = $2")).
		WithArgs(lobbyID, accountID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE lobbies SET last_activity_at = NOW(), updated_at = NOW() WHERE id = $1")).
		WithArgs(lobbyID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	leaveReq := httptest.NewRequest(http.MethodPost, "/lobby/leave_lobby", strings.NewReader(`{"session_id":406}`))
	leaveRR := httptest.NewRecorder()

	err := LeaveLobby(leaveRR, leaveReq, sqlxDB, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if leaveRR.Code != http.StatusOK {
		t.Fatalf("got status %d want %d", leaveRR.Code, http.StatusOK)
	}

	chunk := waitForStreamChunk(t, writer.writes, "event: player_left")
	if !strings.Contains(chunk, `"account_id":95`) {
		t.Fatalf("expected account_id in player_left payload, got %q", chunk)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connect lobby did not exit after cancel")
	}

	if brokerSubscriberCount(lobbyBrokerInstance, lobbyID) != 0 {
		t.Fatalf("expected no subscribers after disconnect, got %d", brokerSubscriberCount(lobbyBrokerInstance, lobbyID))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestLobbyEventJSONShape(t *testing.T) {
	payload := lobbyEventPayload{LobbyID: 1, AccountID: 2, Message: "hello"}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"lobby_id":1,"account_id":2,"message":"hello"}` {
		t.Fatalf("unexpected payload json %s", string(body))
	}
}
