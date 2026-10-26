package visibility

import "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"

// FoundingSiteAvailable applies execution's site policy to current viewer facts.
// It is shared by advertised actions and AI planning. It is not an authoritative
// guarantee: hidden obstacles and intervening commands are checked at execution.
func FoundingSiteAvailable(projection Projection, tileID string) bool {
	return foundingSiteAvailable(projection, projectedTiles(projection), tileID)
}

// Action construction already indexes visible tiles for the other unit orders.
func foundingSiteAvailable(projection Projection, tiles map[string]TileProjection, tileID string) bool {
	site, ok := tiles[tileID]
	if !ok || site.Stale {
		return false
	}
	engineSite := engine.Tile{ID: site.ID, X: site.X, Y: site.Y}
	if site.InfluenceCityID != nil {
		engineSite.InfluenceCityID = *site.InfluenceCityID
	}
	centers := make([]engine.Tile, 0, len(projection.Cities))
	for _, city := range projection.Cities {
		if city.Stale {
			continue
		}
		if center, ok := tiles[city.TileID]; ok && !center.Stale {
			centers = append(centers, engine.Tile{ID: center.ID, X: center.X, Y: center.Y})
		}
	}
	return engine.CitySiteAvailable(engineSite, centers)
}
