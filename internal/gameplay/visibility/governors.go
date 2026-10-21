package visibility

import "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"

func governorActions(projection Projection, playerID string) []ActionProjection {
	if projection.Rules == nil {
		return nil
	}
	var actions []ActionProjection
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.Stale {
			continue
		}
		for _, priority := range projection.Rules.GovernorPriorities {
			actions = append(actions, action(engine.CommandSetGovernor, "set-governor", city.ID, priority, map[string]any{"city_id": city.ID, "enabled": true, "priority": priority}, []string{"city_id", "enabled", "priority"}))
		}
		if city.Governor != nil && city.Governor.Enabled {
			actions = append(actions, action(engine.CommandSetGovernor, "set-governor", city.ID, "disabled", map[string]any{"city_id": city.ID, "enabled": false}, []string{"city_id", "enabled"}))
		}
	}
	return actions
}
