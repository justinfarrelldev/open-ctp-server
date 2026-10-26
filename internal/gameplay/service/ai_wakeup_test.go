package service

import (
	"context"
	"encoding/json"
	"testing"

	gameplayai "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

type targetedAIStore struct {
	fakeAIWorkerStore
	targets []int64
	reads   int
}

func (s *targetedAIStore) FindGameplaySchedulingAggregate(gameID int64) (gameplaystore.StartedAggregate, error) {
	s.targets = append(s.targets, gameID)
	s.reads++
	state := aiWorkerState("sequential", 0, nil, []engine.Player{{ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-2"})
	if s.reads > 1 {
		state.Players[0].Controller = engine.ControllerHuman
	}
	raw, err := json.Marshal(state)
	return gameplaystore.StartedAggregate{GameID: gameID, Revision: int64(6 + s.reads), CanonicalStateJSON: raw}, err
}

func TestAICommitWakeupTargetsOneGameAndPreservesDriveSemantics(t *testing.T) {
	store := &targetedAIStore{}
	driver := &fakeAIDriver{result: gameplayai.Result{Commands: 2, Revision: 9}}
	worker := NewAIWorker(store, driver)
	worker.Notify(44)
	worker.Notify(44) // hints coalesce, commands must not duplicate
	select {
	case <-worker.Wakeups():
	default:
		t.Fatal("committed game does not wake worker immediately")
	}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.calls != 0 || len(store.targets) != 2 || store.targets[0] != 44 || store.targets[1] != 44 {
		t.Fatalf("targeted work scanned unrelated worlds: all=%d targeted=%v", store.calls, store.targets)
	}
	if len(driver.calls) != 1 || driver.calls[0] != (fakeAIDriveCall{gameID: 44, playerID: "account-2", revision: 7}) {
		t.Fatalf("AI drive changed: %+v", driver.calls)
	}
}

func TestAIWakeupHintsAreBoundedAndRetainDistinctGames(t *testing.T) {
	worker := NewAIWorker(&fakeAIWorkerStore{}, &fakeAIDriver{})
	worker.Notify(0)
	for game := int64(1); game <= maxPendingAIGames+10; game++ {
		worker.Notify(game)
		worker.Notify(game)
	}
	if len(worker.queue.pending) != maxPendingAIGames {
		t.Fatalf("unbounded pending games: %d", len(worker.queue.pending))
	}
	for game := int64(1); game <= maxPendingAIGames; game++ {
		if got := worker.takePending(); got != game {
			t.Fatalf("FIFO fairness: got %d want %d", got, game)
		}
	}
	if got := worker.takePending(); got != 0 {
		t.Fatal("empty queue returned a game")
	}
}
