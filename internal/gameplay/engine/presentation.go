package engine

import "encoding/json"

const (
	// Presentation windows bound snapshot/persistence growth. Missing older history
	// is unavailable, not a license to reconstruct past scores or replay rewards.
	NotificationWindow = 128
	StatisticsWindow   = 256
)

// Notification is an immutable owner-authorized semantic result. It is not a
// canonical event payload or a request to reveal target terrain to other viewers.
type Notification struct {
	ID        string `json:"id"`
	PlayerID  string `json:"player_id"`
	Kind      string `json:"kind"`
	Turn      int    `json:"turn"`
	CityID    string `json:"city_id"`
	TileID    string `json:"tile_id"`
	ContentID string `json:"content_id"`
	UnitID    string `json:"unit_id,omitempty"`
}

// StatisticsSample records actual completed-turn nation facts, not forecasts.
// Score uses exactly the current score-victory calculation. Foreign samples are
// private; ranking against unknown nations must remain unavailable in the client.
type StatisticsSample struct {
	Turn       int `json:"turn"`
	Score      int `json:"score"`
	CityCount  int `json:"city_count"`
	UnitCount  int `json:"unit_count"`
	Population int `json:"population"`
}

func cityScore(city City) int {
	return city.Yields.Victory + city.Yields.Food + city.Yields.Production + city.Yields.Gold
}

func recordTurnStatistics(state *State) {
	for index := range state.Nations {
		nation := &state.Nations[index]
		if len(nation.Statistics) > 0 && nation.Statistics[len(nation.Statistics)-1].Turn == state.Phase.Turn {
			continue
		}
		sample := StatisticsSample{Turn: state.Phase.Turn}
		for _, city := range state.Cities {
			if nationForPlayer(*state, city.OwnerID) != nation.ID {
				continue
			}
			sample.CityCount++
			sample.Population += city.Population
			sample.Score += cityScore(city)
		}
		for _, unit := range state.Units {
			if nationForPlayer(*state, unit.OwnerID) == nation.ID {
				sample.UnitCount++
			}
		}
		nation.Statistics = append(nation.Statistics, sample)
		if len(nation.Statistics) > StatisticsWindow {
			nation.Statistics = append([]StatisticsSample(nil), nation.Statistics[len(nation.Statistics)-StatisticsWindow:]...)
		}
	}
}

func recordProductionNotification(state *State, city City, item ProductionItem, unitID string) error {
	recipe := struct {
		Revision  int64  `json:"revision"`
		PlayerID  string `json:"player_id"`
		CityID    string `json:"city_id"`
		EntryID   string `json:"entry_id"`
		Turn      int    `json:"turn"`
		ContentID string `json:"content_id"`
		UnitID    string `json:"unit_id"`
	}{state.Revision + 1, city.OwnerID, city.ID, item.EntryID, state.Phase.Turn, item.ID, unitID}
	raw, err := json.Marshal(recipe)
	if err != nil {
		return err
	}
	notification := Notification{ID: "production:" + hashBytes(raw), PlayerID: city.OwnerID, Kind: "production.completed", Turn: state.Phase.Turn, CityID: city.ID, TileID: city.TileID, ContentID: item.ID, UnitID: unitID}
	for _, existing := range state.Notifications {
		if existing.ID == notification.ID {
			return nil
		}
	}
	state.Notifications = append(state.Notifications, notification)
	// Retain a bounded window per recipient, not a global window that lets one
	// busy empire evict every other player's pending results.
	count, first := 0, 0
	for index := len(state.Notifications) - 1; index >= 0; index-- {
		if state.Notifications[index].PlayerID != city.OwnerID {
			continue
		}
		count++
		if count == NotificationWindow {
			first = index
			break
		}
	}
	if count == NotificationWindow {
		retained := make([]Notification, 0, len(state.Notifications))
		for index, notice := range state.Notifications {
			if index >= first || notice.PlayerID != city.OwnerID {
				retained = append(retained, notice)
			}
		}
		state.Notifications = retained
	}
	return nil
}
