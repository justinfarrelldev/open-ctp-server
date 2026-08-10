package game

import "time"

// Game is the Task 9 persistence model for setup-stage games.
type Game struct {
	ID                int64        `db:"id" json:"id"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
	StartedAt         *time.Time   `db:"started_at" json:"started_at,omitempty"`
	LobbyID           *int64       `db:"lobby_id" json:"lobby_id,omitempty"`
	OwnerAccountID    *int64       `db:"owner_account_id" json:"owner_account_id,omitempty"`
	Name              string       `db:"name" json:"name"`
	PasswordProtected bool         `db:"password_protected" json:"password_protected"`
	Password          string       `db:"password" json:"password"`
	IsLocked          bool         `db:"is_locked" json:"is_locked"`
	IsPublic          bool         `db:"is_public" json:"is_public"`
	Style             string       `db:"style" json:"style"`
	RulesJSON         string       `db:"rules_json" json:"rules_json"`
	ExclusionsJSON    string       `db:"exclusions_json" json:"exclusions_json"`
	Status            string       `db:"status" json:"status"`
	StartingSettlers  int          `db:"starting_settlers" json:"starting_settlers"`
	StartingGold      int          `db:"starting_gold" json:"starting_gold"`
	MapSize           *string      `db:"map_size" json:"map_size,omitempty"`
	Climate           *string      `db:"climate" json:"climate,omitempty"`
	Wetness           *int         `db:"wetness" json:"wetness,omitempty"`
	Players           []GamePlayer `db:"-" json:"players,omitempty"`
}

// GamePlayer is the Task 9 persistence model for game membership rows.
type GamePlayer struct {
	ID         int64     `db:"id" json:"id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	GameID     int64     `db:"game_id" json:"game_id"`
	AccountID  int64     `db:"account_id" json:"account_id"`
	JoinedAt   time.Time `db:"joined_at" json:"joined_at"`
	Country    string    `db:"country" json:"country,omitempty"`
	Ping       int       `db:"ping_ms" json:"ping"`
	Controller string    `db:"controller_kind" json:"controller_kind"`
}
