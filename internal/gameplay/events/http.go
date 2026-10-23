package events

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// Handler serves the authenticated gameplay SSE endpoint.
type Handler struct {
	store     *Store
	db        *sqlx.DB
	sessions  *auth.SessionStore
	projector Projector
	wakeup    Wakeup
}

func NewHandler(db *sqlx.DB, sessions *auth.SessionStore, projector Projector, wakeup Wakeup) *Handler {
	return &Handler{store: NewStore(db), db: db, sessions: sessions, projector: projector, wakeup: wakeup}
}

// ServeHTTP implements GET /game/{id}/events. The database sequence is always
// the stream authority; wakeups merely cause an earlier durable reread.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid request; request must be a GET request", http.StatusBadRequest)
		return
	}
	gameID, err := gameIDFromEventPath(r.URL.Path)
	if err != nil {
		http.Error(w, "game id must be specified", http.StatusBadRequest)
		return
	}
	if h == nil || h.db == nil || h.sessions == nil || h.projector == nil {
		http.Error(w, "gameplay event stream is unavailable", http.StatusInternalServerError)
		return
	}
	session, err := auth.RequireSessionFromQuery(w, r, h.sessions)
	if err != nil {
		return
	}
	viewer, err := h.viewer(r, gameID, int64(session.AccountID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "game player membership is required", http.StatusForbidden)
			return
		}
		http.Error(w, "unable to authenticate game player", http.StatusInternalServerError)
		return
	}
	cursor, err := ParseLastEventID(r.Header.Get("Last-Event-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "response writer does not support streaming", http.StatusInternalServerError)
		return
	}

	current := r.Header.Get("X-Gameplay-Resync") == "current"
	var wakeups <-chan struct{}
	if !current && h.wakeup != nil {
		// Subscribe BEFORE the first durable read. A commit between reading and
		// subscribing would otherwise lose its only wakeup until a later turn.
		var unsubscribe func()
		wakeups, unsubscribe = h.wakeup.Subscribe(gameID)
		defer unsubscribe()
	}
	mode := "stream"
	if current {
		mode = "current"
	}
	w.Header().Set("X-Gameplay-Read-Mode", mode)
	started := time.Now()
	var items []SSEEvent
	timing := "initial_read"
	if current {
		items, err = h.store.resync(r.Context(), viewer, h.projector)
		timing = "snapshot"
	} else {
		items, err = h.store.Resume(r.Context(), viewer, cursor, h.projector)
	}
	w.Header().Set("Server-Timing", fmt.Sprintf("%s;dur=%.3f", timing, float64(time.Since(started))/float64(time.Millisecond)))
	w.Header().Set("Cache-Control", "no-store")
	if err != nil || current && len(items) != 1 {
		http.Error(w, "gameplay event read is unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	cursor, ok = writeProjectedItems(w, items, cursor)
	if !ok {
		return
	}
	flusher.Flush()
	if h.writeTerminal(w, r, viewer) {
		return
	}
	flusher.Flush()
	if current || h.wakeup == nil {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case _, open := <-wakeups:
			if !open {
				return
			}
			var wrote bool
			cursor, wrote = h.writeAvailable(w, r, viewer, cursor)
			if !wrote || h.writeTerminal(w, r, viewer) {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) viewer(r *http.Request, gameID, accountID int64) (Viewer, error) {
	var playerID int64
	if err := h.db.QueryRowxContext(r.Context(), `SELECT id, account_id FROM game_players WHERE game_id = $1 AND account_id = $2`, gameID, accountID).Scan(&playerID, &accountID); err != nil {
		return Viewer{}, err
	}
	return Viewer{GameID: gameID, PlayerID: playerID, AccountID: accountID}, nil
}

func (h *Handler) writeAvailable(w http.ResponseWriter, r *http.Request, viewer Viewer, cursor int64) (int64, bool) {
	items, err := h.store.Resume(r.Context(), viewer, cursor, h.projector)
	if err != nil {
		return cursor, false
	}
	return writeProjectedItems(w, items, cursor)
}

func writeProjectedItems(w http.ResponseWriter, items []SSEEvent, cursor int64) (int64, bool) {
	for _, item := range items {
		if err := writeSSEEvent(w, item); err != nil {
			return cursor, false
		}
		if item.Name != "resync" && item.ID > cursor {
			cursor = item.ID
		}
	}
	return cursor, true
}

func (h *Handler) writeTerminal(w http.ResponseWriter, r *http.Request, viewer Viewer) bool {
	event, ended, err := h.store.Terminal(r.Context(), viewer, h.projector)
	if err != nil || !ended {
		return false
	}
	return writeSSEEvent(w, event) != nil || ended
}

func gameIDFromEventPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "game" || parts[2] != "events" {
		return 0, errors.New("invalid gameplay event path")
	}
	gameID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || gameID <= 0 {
		return 0, errors.New("invalid game id")
	}
	return gameID, nil
}

func writeSSEEvent(w http.ResponseWriter, event SSEEvent) error {
	if event.ID > 0 {
		if _, err := fmt.Fprintf(w, "id: %d\n", event.ID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Name, event.Data); err != nil {
		return err
	}
	return nil
}

// Broker is a process-local latency wakeup. It has no event payload or cursor
// authority; reconnects always reread the durable event sequence.
type Broker struct {
	mu          sync.Mutex
	nextClient  int64
	subscribers map[int64]map[int64]chan struct{}
}

func NewBroker() *Broker { return &Broker{subscribers: make(map[int64]map[int64]chan struct{})} }

func (b *Broker) Subscribe(gameID int64) (<-chan struct{}, func()) {
	if b == nil {
		return nil, func() {}
	}
	b.mu.Lock()
	b.nextClient++
	clientID := b.nextClient
	listeners := b.subscribers[gameID]
	if listeners == nil {
		listeners = make(map[int64]chan struct{})
		b.subscribers[gameID] = listeners
	}
	ch := make(chan struct{}, 1)
	listeners[clientID] = ch
	b.mu.Unlock()
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			listeners := b.subscribers[gameID]
			if listeners == nil {
				return
			}
			if registered, ok := listeners[clientID]; ok {
				delete(listeners, clientID)
				close(registered)
			}
			if len(listeners) == 0 {
				delete(b.subscribers, gameID)
			}
		})
	}
}

func (b *Broker) Publish(gameID int64) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers[gameID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (b *Broker) subscriberCount(gameID int64) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subscribers[gameID])
}

// terminalFromState remains deliberately separate so terminal payloads are
// projected exactly like snapshots rather than exposing canonical terminal data.
func terminalFromState(viewer Viewer, canonicalState []byte, projector Projector) (SSEEvent, bool, error) {
	var state struct {
		Terminal struct {
			Ended bool `json:"ended"`
		} `json:"terminal"`
	}
	if err := json.Unmarshal(canonicalState, &state); err != nil {
		return SSEEvent{}, false, err
	}
	if !state.Terminal.Ended {
		return SSEEvent{}, false, nil
	}
	event, err := TerminalEvent(viewer, canonicalState, projector)
	return event, true, err
}
