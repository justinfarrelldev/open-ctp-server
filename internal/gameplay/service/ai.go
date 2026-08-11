package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

// AIGateway adapts the non-authoritative AI driver to the ordinary command gateway.
type AIGateway struct{ service *Service }

func NewAIGateway(service *Service) *AIGateway { return &AIGateway{service: service} }

func (g *AIGateway) Submit(ctx context.Context, submission ai.Submission) (ai.SubmissionResult, error) {
	result, err := g.service.ExecuteAICommand(ctx, submission.ActorPlayerID, CommandRequest{
		GameID:           submission.GameID,
		ExpectedRevision: submission.ExpectedRevision,
		IdempotencyKey:   submission.IdempotencyKey,
		CommandID:        submission.CommandID,
		Type:             submission.Type,
		Payload:          submission.Payload,
	})
	if err != nil {
		if StatusCode(err) == 409 && g != nil && g.service != nil && g.service.store != nil {
			if aggregate, readErr := g.service.store.ReadAggregate(submission.GameID); readErr == nil {
				if state, decodeErr := engine.DecodeCanonicalStateJSON(aggregate.CanonicalStateJSON); decodeErr == nil && state.Terminal.Ended {
					return ai.SubmissionResult{}, ai.ErrTerminal
				}
			}
		}
		return ai.SubmissionResult{}, err
	}
	return ai.SubmissionResult{Revision: result.Revision, Terminal: result.Terminal}, nil
}

// Projection returns exactly the visibility projection used by player-facing reads; canonical state
// remains inside this adapter and is never passed to the AI driver.
func (g *AIGateway) Projection(ctx context.Context, gameID int64, playerID string) ([]byte, error) {
	if g == nil || g.service == nil || g.service.store == nil {
		return nil, ai.ErrTerminal
	}
	accountID := strings.TrimPrefix(playerID, "account-")
	if accountID == playerID {
		return nil, commandError(400, "AI actor player is invalid")
	}
	if parsed, err := strconv.ParseInt(accountID, 10, 64); err != nil || parsed < 1 {
		return nil, commandError(400, "AI actor player is invalid")
	}
	aggregate, err := g.service.store.ReadAggregate(gameID)
	if err != nil {
		return nil, err
	}
	state, err := engine.DecodeCanonicalStateJSON(aggregate.CanonicalStateJSON)
	if err != nil {
		return nil, commandError(422, "gameplay state is invalid: "+err.Error())
	}
	return visibility.AIInput(state, playerID)
}

// ExplorationHistory reconstructs only this AI actor's accepted movement steps.
func (g *AIGateway) ExplorationHistory(ctx context.Context, gameID int64, playerID string) ([]ai.ExplorationStep, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if g == nil || g.service == nil || g.service.store == nil {
		return nil, ai.ErrTerminal
	}
	accountID := strings.TrimPrefix(playerID, "account-")
	if accountID == playerID {
		return nil, commandError(400, "AI actor player is invalid")
	}
	account, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil || account < 1 {
		return nil, commandError(400, "AI actor player is invalid")
	}
	actorPlayerID, err := g.service.store.GamePlayerID(gameID, account)
	if err != nil {
		return nil, err
	}
	stored, err := g.service.store.AIExplorationHistory(gameID, actorPlayerID)
	if err != nil {
		return nil, err
	}
	history := make([]ai.ExplorationStep, len(stored))
	for index, step := range stored {
		history[index] = ai.ExplorationStep{CommandType: step.CommandType, DestinationTileID: step.DestinationTileID, Sequence: step.Sequence}
	}
	return history, nil
}
