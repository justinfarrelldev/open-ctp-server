package engine

// CitySiteAvailable implements the released settlement site restrictions:
// UnitData::Settle rejects any existing city influence; Player::CreateCity also
// rejects an existing or adjacent city. National ownership is not a restriction.
// Adjacency uses the same bounded tile geometry as other server commands.
//
// Execution supplies all city centers. Viewer-side callers MUST supply only
// currently observable centers, never hidden state or stale city memory. This
// predicate checks the site only, not unit eligibility or command admission.
func CitySiteAvailable(site Tile, cityCenters []Tile) bool {
	if site.ID == "" || site.InfluenceCityID != "" {
		return false
	}
	for _, center := range cityCenters {
		if center.ID == site.ID || adjacent(site, center) {
			return false
		}
	}
	return true
}

func foundingSiteAvailable(state State, site Tile) bool {
	cityTiles := make(map[string]bool, len(state.Cities))
	for _, city := range state.Cities {
		cityTiles[city.TileID] = true
	}
	centers := make([]Tile, 0, len(state.Cities))
	for _, tile := range state.Tiles {
		if cityTiles[tile.ID] {
			centers = append(centers, tile)
		}
	}
	return CitySiteAvailable(site, centers)
}
