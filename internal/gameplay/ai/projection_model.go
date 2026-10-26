package ai

import (
	"context"
	"encoding/json"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

// projectionModelSource is an optional in-process optimization, not canonical
// state access: it returns exactly the ordinary authorized viewer model. Keeping
// the byte source supports external/legacy controllers and all existing fixtures.
type projectionModelSource interface {
	ProjectionModel(context.Context, int64, string) (visibility.Projection, error)
}

func (d *Driver) projectionModel(ctx context.Context, gameID int64, playerID string) (visibility.Projection, error) {
	if typed, ok := d.projections.(projectionModelSource); ok {
		return typed.ProjectionModel(ctx, gameID, playerID)
	}
	input, err := d.projections.Projection(ctx, gameID, playerID)
	if err != nil {
		return visibility.Projection{}, err
	}
	closeout, err := planningCloseout(input)
	if err != nil {
		return visibility.Projection{}, err
	}
	if closeout.PlanningStatus == visibility.PlanningStatusClosed || closeout.PlanningStatus == visibility.PlanningStatusQueued {
		return visibility.Projection{Revision: closeout.Revision, PlanningStatus: closeout.PlanningStatus}, nil
	}
	var projection visibility.Projection
	if err := json.Unmarshal(input, &projection); err != nil {
		return visibility.Projection{}, err
	}
	return projection, nil
}
