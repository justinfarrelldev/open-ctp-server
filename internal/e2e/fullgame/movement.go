package fullgame

import (
	"encoding/json"
	"fmt"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func PlanVisibleRoute(publicProjection []byte, unitID, targetTileID string) ([]string, error) {
	var projection visibility.Projection
	if err := json.Unmarshal(publicProjection, &projection); err != nil {
		return nil, fmt.Errorf("decode public projection: %w", err)
	}
	return visibility.Plan(projection, unitID, targetTileID)
}
