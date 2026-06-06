package lobby

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type lobbyEvent struct {
	name    string
	payload lobbyEventPayload
}

type lobbyEventPayload struct {
	LobbyID   int64  `json:"lobby_id"`
	AccountID int64  `json:"account_id"`
	Message   string `json:"message,omitempty"`
}

type lobbyBroker struct {
	mu          sync.Mutex
	nextClient  int64
	subscribers map[int64]map[int64]chan lobbyEvent
}

func newLobbyBroker() *lobbyBroker {
	return &lobbyBroker{
		subscribers: make(map[int64]map[int64]chan lobbyEvent),
	}
}

var lobbyBrokerInstance = newLobbyBroker()

func (b *lobbyBroker) subscribe(lobbyID int64) (<-chan lobbyEvent, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextClient++
	clientID := b.nextClient
	clientEvents := make(chan lobbyEvent, 8)
	if b.subscribers[lobbyID] == nil {
		b.subscribers[lobbyID] = make(map[int64]chan lobbyEvent)
	}
	b.subscribers[lobbyID][clientID] = clientEvents

	unsubscribed := false
	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if unsubscribed {
			return
		}
		unsubscribed = true
		clients := b.subscribers[lobbyID]
		if clients == nil {
			return
		}
		channel, ok := clients[clientID]
		if !ok {
			return
		}
		delete(clients, clientID)
		close(channel)
		if len(clients) == 0 {
			delete(b.subscribers, lobbyID)
		}
	}

	return clientEvents, unsubscribe
}

func (b *lobbyBroker) publish(lobbyID int64, event lobbyEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, clientEvents := range b.subscribers[lobbyID] {
		select {
		case clientEvents <- event:
		default:
		}
	}
}

func (b *lobbyBroker) subscriberCount(lobbyID int64) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subscribers[lobbyID])
}

func writeLobbySSEEvent(w http.ResponseWriter, event lobbyEvent) error {
	body, err := json.Marshal(event.payload)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding lobby event: %v", err)
	}

	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.name, body); err != nil {
		return err
	}

	return nil
}
