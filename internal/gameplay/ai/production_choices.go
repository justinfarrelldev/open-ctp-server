package ai

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

// Restrict preference ranking to this city's authoritative eligible catalog. A
// stronger but locally illegal unit must not starve a weaker legal defender.
// The payload uses the advertised item, not a reconstructed cost/effect snapshot.
func cityProductionChoices(projection visibility.Projection, city visibility.CityProjection) (visibility.Projection, map[string]engine.ProductionItem, bool) {
	if city.ProductionOptions == nil {
		// Legacy projection fixtures predate production_options. Actual runtime
		// owned-city projections always supply complete or unavailable explicitly.
		return projection, nil, true
	}
	if city.ProductionOptions.Status != "complete" || len(city.ProductionOptions.Items) == 0 || projection.Rules == nil {
		return projection, nil, false
	}
	legal := make(map[string]engine.ProductionItem, len(city.ProductionOptions.Items))
	for _, item := range city.ProductionOptions.Items {
		legal[item.ID] = item
	}
	catalog := *projection.Rules
	catalog.Units = nil
	for _, unit := range projection.Rules.Units {
		if item, ok := legal[unit.ID]; ok && item.Kind == engine.ProductionUnit {
			catalog.Units = append(catalog.Units, unit)
		}
	}
	catalog.Production = nil
	for _, production := range projection.Rules.Production {
		if _, ok := legal[production.ID]; ok {
			catalog.Production = append(catalog.Production, production)
		}
	}
	projection.Rules = &catalog
	return projection, legal, true
}
