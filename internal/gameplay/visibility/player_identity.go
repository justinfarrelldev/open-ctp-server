package visibility

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"sort"
)

// These are public lobby/participant identities, not inferred from visible units.
// ColorIndex is the persisted zero-based palette slot; nil is an older save with
// no assigned slot. The client chooses artwork/RGBA, not geopolitical ownership.
type PlayerIdentity struct {
	ID             string `json:"id"`
	NationID       string `json:"nation_id"`
	CivilizationID string `json:"civilization_id,omitempty"`
	ColorIndex     *int   `json:"color_index,omitempty"`
}

func playerIdentities(state engine.State) []PlayerIdentity {
	civilizations := map[string]string{}
	for _, nation := range state.Nations {
		civilizations[nation.ID] = nation.CivilizationID
	}
	result := make([]PlayerIdentity, 0, len(state.Players))
	for _, player := range state.Players {
		result = append(result, PlayerIdentity{ID: player.ID, NationID: player.NationID, CivilizationID: civilizations[player.NationID], ColorIndex: copyValue(player.ColorIndex)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
