package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	gameplayai "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

const defaultAIWorkerDriveBudget = 32

type startedAggregateFinder interface {
	FindStartedGameplayAggregates() ([]gameplaystore.StartedAggregate, error)
}

type aiTurnDriver interface {
	Drive(context.Context, int64, string, int64) (gameplayai.Result, error)
}

type dueAIPlayer struct {
	GameID   int64
	Revision int64
	PlayerID string
}

// AIWorker drives server-controlled AI players in started games through the ordinary AI driver.
type AIWorker struct {
	store       startedAggregateFinder
	driver      aiTurnDriver
	MaxDrives   int
	AfterCommit func(context.Context) error
	queue       aiWakeQueue
}

func NewAIWorker(store startedAggregateFinder, driver aiTurnDriver) *AIWorker {
	return &AIWorker{store: store, driver: driver, MaxDrives: defaultAIWorkerDriveBudget, queue: newAIWakeQueue()}
}

func (w *AIWorker) RunOnce(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if w == nil || w.store == nil || w.driver == nil {
		return fmt.Errorf("gameplay AI worker is not configured")
	}
	budget := w.MaxDrives
	if budget < 1 {
		budget = defaultAIWorkerDriveBudget
	}
	for drive := 0; drive < budget; drive++ {
		due, err := w.findDuePlayers()
		if err != nil {
			return err
		}
		if len(due) == 0 {
			return nil
		}
		candidate := due[0]
		result, err := w.driver.Drive(ctx, candidate.GameID, candidate.PlayerID, candidate.Revision)
		if result.Commands > 0 {
			// Re-read this game's durable scheduling state for its next player.
			// FIFO hints let other committed games make progress between drives.
			if _, ok := w.store.(targetedAggregateFinder); ok {
				w.Notify(candidate.GameID)
			}
		}
		if result.Commands > 0 && w.AfterCommit != nil {
			if dispatchErr := w.AfterCommit(ctx); dispatchErr != nil {
				return fmt.Errorf("dispatching AI gameplay events for game %d player %q: %w", candidate.GameID, candidate.PlayerID, dispatchErr)
			}
		}
		if errors.Is(err, gameplayai.ErrTerminal) {
			return nil
		}
		if err != nil {
			if harmlessAIDriveError(err) {
				continue
			}
			return fmt.Errorf("driving AI for game %d player %q: %w", candidate.GameID, candidate.PlayerID, err)
		}
	}
	return nil
}

func (w *AIWorker) findDuePlayers() ([]dueAIPlayer, error) {
	if targeted, ok := w.store.(targetedAggregateFinder); ok {
		for gameID := w.takePending(); gameID > 0; gameID = w.takePending() {
			aggregate, err := targeted.FindGameplaySchedulingAggregate(gameID)
			if err != nil {
				return nil, err
			}
			if aggregate.GameID == 0 {
				continue
			} // removed/ended between notification and read
			var state engine.State
			if err := json.Unmarshal(aggregate.CanonicalStateJSON, &state); err != nil {
				return nil, err
			}
			state.Revision = aggregate.Revision
			if due := dueAIPlayersForState(gameID, state); len(due) > 0 {
				return due, nil
			}
			// A completed human turn needs no scan of unrelated worlds.
			if !w.queue.hasPending() {
				return nil, nil
			}
		}
	}
	aggregates, err := w.store.FindStartedGameplayAggregates()
	if err != nil {
		return nil, err
	}
	due := make([]dueAIPlayer, 0)
	for _, aggregate := range aggregates {
		var state engine.State
		if err := json.Unmarshal(aggregate.CanonicalStateJSON, &state); err != nil {
			return nil, fmt.Errorf("decoding started gameplay state for AI worker game %d: %w", aggregate.GameID, err)
		}
		state.Revision = aggregate.Revision
		due = append(due, dueAIPlayersForState(aggregate.GameID, state)...)
	}
	return due, nil
}

func dueAIPlayersForState(gameID int64, state engine.State) []dueAIPlayer {
	if gameID < 1 || state.Terminal.Ended || state.Phase.ResolutionLocked {
		return nil
	}
	activeAI := activeAIPlayers(state)
	if len(activeAI) == 0 {
		return nil
	}
	mode := state.Phase.Mode
	if mode == "sequential" || mode == "long_turn" {
		return dueActiveAI(gameID, state, activeAI)
	}
	if mode == "hybrid" && policyStage(state) == "conflict" {
		return dueActiveAI(gameID, state, activeAI)
	}
	if mode == "simultaneous" || mode == "hybrid" {
		if policyStage(state) != "planning" {
			return nil
		}
		closed := policyClosedPlayers(state)
		return duePlanningAI(gameID, state, activeAI, closed)
	}
	return nil
}

func dueActiveAI(gameID int64, state engine.State, activeAI map[string]bool) []dueAIPlayer {
	if state.Phase.ActivePlayerIndex < 0 || state.Phase.ActivePlayerIndex >= len(state.Phase.ActivePlayerOrder) {
		return nil
	}
	playerID := state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex]
	if !activeAI[playerID] {
		return nil
	}
	return []dueAIPlayer{{GameID: gameID, Revision: state.Revision, PlayerID: playerID}}
}

func duePlanningAI(gameID int64, state engine.State, activeAI map[string]bool, closed map[string]bool) []dueAIPlayer {
	due := make([]dueAIPlayer, 0, len(activeAI))
	for _, playerID := range state.Phase.ActivePlayerOrder {
		if activeAI[playerID] && !closed[playerID] {
			due = append(due, dueAIPlayer{GameID: gameID, Revision: state.Revision, PlayerID: playerID})
		}
	}
	return due
}

func activeAIPlayers(state engine.State) map[string]bool {
	result := make(map[string]bool)
	for _, player := range state.Players {
		if player.Lifecycle == engine.PlayerActive && player.Controller == engine.ControllerAI {
			result[player.ID] = true
		}
	}
	return result
}

func policyStage(state engine.State) string {
	if state.Policy.Stage == "" {
		return "planning"
	}
	return state.Policy.Stage
}

func policyClosedPlayers(state engine.State) map[string]bool {
	closed := make(map[string]bool, len(state.Policy.EarlyEnds))
	for _, playerID := range state.Policy.EarlyEnds {
		closed[playerID] = true
	}
	return closed
}

func harmlessAIDriveError(err error) bool {
	switch StatusCode(err) {
	case 403, 409, 422:
		return true
	default:
		return false
	}
}
