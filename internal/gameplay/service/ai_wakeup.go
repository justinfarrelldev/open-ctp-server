package service

import (
	"context"
	"log"
	"sync"
	"time"

	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

// Notifications are bounded latency hints, never canonical scheduling authority.
// Overflow is safe: the periodic durable discovery still finds all started games.
const maxPendingAIGames = 1024

type targetedAggregateFinder interface {
	FindGameplaySchedulingAggregate(int64) (gameplaystore.StartedAggregate, error)
}
type aiWakeQueue struct {
	mu      sync.Mutex
	pending map[int64]bool
	order   []int64
	ready   chan struct{}
}

func newAIWakeQueue() aiWakeQueue {
	return aiWakeQueue{pending: make(map[int64]bool), ready: make(chan struct{}, 1)}
}

// Notify may be called concurrently only after a successful commit. No state or
// command is carried in the notification; duplicates coalesce by game identity.
func (w *AIWorker) Notify(gameID int64) {
	if w == nil || gameID < 1 {
		return
	}
	q := &w.queue
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.pending == nil {
		q.pending = make(map[int64]bool)
	}
	if q.ready == nil {
		q.ready = make(chan struct{}, 1)
	}
	if !q.pending[gameID] && len(q.pending) < maxPendingAIGames {
		q.pending[gameID] = true
		q.order = append(q.order, gameID)
	}
	select {
	case q.ready <- struct{}{}:
	default:
	}
}
func (w *AIWorker) Wakeups() <-chan struct{} {
	q := &w.queue
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.ready == nil {
		q.ready = make(chan struct{}, 1)
	}
	return q.ready
}
func (w *AIWorker) takePending() int64 {
	q := &w.queue
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.order) == 0 {
		return 0
	}
	gameID := q.order[0]
	q.order = q.order[1:]
	delete(q.pending, gameID)
	return gameID
}
func (q *aiWakeQueue) hasPending() bool { q.mu.Lock(); defer q.mu.Unlock(); return len(q.pending) > 0 }

// Run replaces poll-only admission with immediate committed-game wakeups. The
// same single worker and per-game revision gateway still serialize AI actions.
// A periodic sweep recovers restarts, missed/overflowed hints and newly started
// games without requiring a human command. Work remains bounded by MaxDrives.
func (w *AIWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := w.RunOnce(ctx); err != nil && ctx.Err() == nil {
			log.Printf("gameplay AI driver failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-w.Wakeups():
		case <-ticker.C:
		}
	}
}
