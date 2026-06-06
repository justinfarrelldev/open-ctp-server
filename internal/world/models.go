package world

import "time"

// StoredWorld is the store-facing aggregate that keeps immutable world data and
// mutable turn-state timestamps separate from the transport contract.
type StoredWorld struct {
	Snapshot          WorldSnapshot `json:"snapshot"`
	SnapshotCreatedAt time.Time     `json:"snapshot_created_at"`
	SnapshotUpdatedAt time.Time     `json:"snapshot_updated_at"`
	TurnCreatedAt     time.Time     `json:"turn_created_at"`
	TurnUpdatedAt     time.Time     `json:"turn_updated_at"`
}

type worldSnapshotRow struct {
	GameID           int64     `db:"game_id"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	GameStatus       string    `db:"game_status"`
	GeneratorVersion string    `db:"generator_version"`
	Style            string    `db:"style"`
	RulesJSON        string    `db:"rules_json"`
	ExclusionsJSON   string    `db:"exclusions_json"`
	StartingSettlers int       `db:"starting_settlers"`
	StartingGold     int       `db:"starting_gold"`
	Seed             *int64    `db:"seed"`
	Wetness          *int      `db:"wetness"`
	MapSize          string    `db:"map_size"`
	Climate          string    `db:"climate"`
	PlayerOrderJSON  string    `db:"player_order_json"`
	SpawnsJSON       string    `db:"spawns_json"`
	TilesJSON        string    `db:"tiles_json"`
}

type worldTurnStateRow struct {
	GameID                int64     `db:"game_id"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
	TurnNumber            int       `db:"turn_number"`
	CurrentPlayerIndex    int       `db:"current_player_index"`
	ActivePlayerOrderJSON string    `db:"active_player_order_json"`
	RecalculationCount    int       `db:"recalculation_count"`
	RNGState              string    `db:"rng_state"`
	RatingsJSON           string    `db:"ratings_json"`
	PollutionJSON         string    `db:"pollution_json"`
	LastRecalculatedAt    time.Time `db:"last_recalculated_at"`
}
