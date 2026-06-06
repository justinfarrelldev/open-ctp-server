package game

// CreateGameDocsArgs documents the request payload for creating a game.
type CreateGameDocsArgs struct {
	SessionId         *int64         `json:"session_id,omitempty"`
	Name              string         `json:"name"`
	OwnerAccountID    *int64         `json:"owner_account_id,omitempty"`
	LobbyID           *int64         `json:"lobby_id,omitempty"`
	PasswordProtected bool           `json:"password_protected"`
	Password          string         `json:"password,omitempty"`
	IsLocked          *bool          `json:"is_locked,omitempty"`
	IsPublic          *bool          `json:"is_public,omitempty"`
	Style             *string        `json:"style,omitempty"`
	RulesJSON         []any          `json:"rules_json,omitempty"`
	ExclusionsJSON    map[string]any `json:"exclusions_json,omitempty"`
	StartingSettlers  *int           `json:"starting_settlers,omitempty"`
	StartingGold      *int           `json:"starting_gold,omitempty"`
}

// GameUpdateDocsArgs documents the nested mutable fields accepted by update_game.
type GameUpdateDocsArgs struct {
	Name           *string        `json:"name,omitempty"`
	IsLocked       *bool          `json:"is_locked,omitempty"`
	IsPublic       *bool          `json:"is_public,omitempty"`
	Style          *string        `json:"style,omitempty"`
	RulesJSON      []any          `json:"rules_json,omitempty"`
	ExclusionsJSON map[string]any `json:"exclusions_json,omitempty"`
}

// UpdateGameDocsArgs documents the request payload for update_game.
type UpdateGameDocsArgs struct {
	GameId    *int64              `json:"game_id"`
	SessionId *int64              `json:"session_id"`
	Game      *GameUpdateDocsArgs `json:"game"`
}

// UpdateRulesDocsArgs documents the request payload for update_rules.
type UpdateRulesDocsArgs struct {
	GameId    int64  `json:"game_id"`
	SessionId *int64 `json:"session_id"`
	RulesJSON []any  `json:"rules_json"`
}

// UpdateExclusionsDocsArgs documents the request payload for update_exclusions.
type UpdateExclusionsDocsArgs struct {
	GameId         int64          `json:"game_id"`
	SessionId      *int64         `json:"session_id"`
	ExclusionsJSON map[string]any `json:"exclusions_json"`
}
