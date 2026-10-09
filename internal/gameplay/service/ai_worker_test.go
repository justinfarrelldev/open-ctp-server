package service

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	gameplayai "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

func TestDueAIPlayersForStateCoversTurnModes(t *testing.T) {
	tests := []struct {
		name  string
		state engine.State
		want  []dueAIPlayer
	}{
		{
			name:  "sequential active ai",
			state: aiWorkerState("sequential", 1, nil, []engine.Player{{ID: "account-1", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-1", "account-2"}),
			want:  []dueAIPlayer{{GameID: 44, Revision: 7, PlayerID: "account-2"}},
		},
		{
			name:  "sequential inactive ai is skipped",
			state: aiWorkerState("sequential", 0, nil, []engine.Player{{ID: "account-1", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-1", "account-2"}),
		},
		{
			name:  "simultaneous planning drives open and queued ai but not closed ai",
			state: aiWorkerState("simultaneous", 0, &engine.TurnPolicyState{Stage: "planning", Queue: []engine.TurnPolicyCommand{{ActorID: "account-3"}}, EarlyEnds: []string{"account-4"}}, []engine.Player{{ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}, {ID: "account-3", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}, {ID: "account-4", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-2", "account-3", "account-4"}),
			want:  []dueAIPlayer{{GameID: 44, Revision: 7, PlayerID: "account-2"}, {GameID: 44, Revision: 7, PlayerID: "account-3"}},
		},
		{
			name:  "hybrid conflict active ai",
			state: aiWorkerState("hybrid", 1, &engine.TurnPolicyState{Stage: "conflict", Components: [][]string{{"account-1", "account-2"}}}, []engine.Player{{ID: "account-1", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-1", "account-2"}),
			want:  []dueAIPlayer{{GameID: 44, Revision: 7, PlayerID: "account-2"}},
		},
		{
			name: "terminal game has no due ai",
			state: func() engine.State {
				state := aiWorkerState("sequential", 1, nil, []engine.Player{{ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-1", "account-2"})
				state.Terminal.Ended = true
				return state
			}(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dueAIPlayersForState(44, tc.state)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("dueAIPlayersForState() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestAIWorkerDrivesDuePlayersAndDispatchesAfterCommands(t *testing.T) {
	state := aiWorkerState("sequential", 0, nil, []engine.Player{{ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-2"})
	body, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeAIWorkerStore{batches: [][]gameplaystore.StartedAggregate{{{GameID: 44, Revision: 7, CanonicalStateJSON: body}}, {}}}
	driver := &fakeAIDriver{result: gameplayai.Result{Commands: 2, Revision: 9}}
	dispatches := 0
	worker := NewAIWorker(store, driver)
	worker.AfterCommit = func(context.Context) error {
		dispatches++
		return nil
	}

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(driver.calls) != 1 || driver.calls[0] != (fakeAIDriveCall{gameID: 44, playerID: "account-2", revision: 7}) {
		t.Fatalf("driver calls = %#v", driver.calls)
	}
	if dispatches != 1 {
		t.Fatalf("dispatches = %d, want 1", dispatches)
	}
	if store.calls != 2 {
		t.Fatalf("store calls = %d, want worker to requery until no due AI", store.calls)
	}
}

func TestAIWorkerIgnoresHarmlessStaleDriveConflicts(t *testing.T) {
	state := aiWorkerState("sequential", 0, nil, []engine.Player{{ID: "account-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}, []string{"account-2"})
	body, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeAIWorkerStore{batches: [][]gameplaystore.StartedAggregate{{{GameID: 44, Revision: 7, CanonicalStateJSON: body}}, {}}}
	driver := &fakeAIDriver{err: commandError(409, "gameplay revision is stale")}
	worker := NewAIWorker(store, driver)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(driver.calls) != 1 {
		t.Fatalf("driver calls = %#v", driver.calls)
	}
}

func aiWorkerState(mode string, activeIndex int, policy *engine.TurnPolicyState, players []engine.Player, order []string) engine.State {
	state := engine.State{
		Revision: 7,
		Players:  players,
		Phase:    engine.PhaseState{Mode: mode, ActivePlayerIndex: activeIndex, ActivePlayerOrder: order},
	}
	if policy != nil {
		state.Policy = *policy
	}
	return state
}

type fakeAIWorkerStore struct {
	batches [][]gameplaystore.StartedAggregate
	calls   int
}

func (s *fakeAIWorkerStore) FindStartedGameplayAggregates() ([]gameplaystore.StartedAggregate, error) {
	if s.calls >= len(s.batches) {
		s.calls++
		return nil, nil
	}
	batch := s.batches[s.calls]
	s.calls++
	return batch, nil
}

type fakeAIDriveCall struct {
	gameID   int64
	playerID string
	revision int64
}

type fakeAIDriver struct {
	result gameplayai.Result
	err    error
	calls  []fakeAIDriveCall
}

func (d *fakeAIDriver) Drive(ctx context.Context, gameID int64, playerID string, revision int64) (gameplayai.Result, error) {
	if err := ctx.Err(); err != nil {
		return gameplayai.Result{}, err
	}
	d.calls = append(d.calls, fakeAIDriveCall{gameID: gameID, playerID: playerID, revision: revision})
	if d.err != nil {
		return gameplayai.Result{}, d.err
	}
	return d.result, nil
}
