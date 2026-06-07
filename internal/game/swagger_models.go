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

// WorldGenerationHostParamsDocs documents the serialized world host parameters returned by /world endpoints.
type WorldGenerationHostParamsDocs struct {
	Style            string `json:"style"`
	RulesJSON        string `json:"rules_json"`
	ExclusionsJSON   string `json:"exclusions_json"`
	StartingSettlers int    `json:"starting_settlers"`
	StartingGold     int    `json:"starting_gold"`
	Seed             *int64 `json:"seed,omitempty"`
	Wetness          *int   `json:"wetness,omitempty"`
	MapSize          string `json:"map_size,omitempty"`
	Climate          string `json:"climate,omitempty"`
	GeneratorVersion string `json:"generator_version,omitempty"`
}

// WorldTileDocs documents one serialized world tile returned by /world endpoints.
type WorldTileDocs struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Terrain string `json:"terrain"`
}

// WorldPlayerSpawnDocs documents one persisted player spawn returned by /world endpoints.
type WorldPlayerSpawnDocs struct {
	AccountID int64  `json:"account_id"`
	Country   string `json:"country,omitempty"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
}

// WorldReadResponseDocs documents the persisted generated world returned by GET /world/{game_id}.
type WorldReadResponseDocs struct {
	GameID           int64                         `json:"game_id"`
	GameStatus       string                        `json:"game_status"`
	GeneratorVersion string                        `json:"generator_version"`
	HostParams       WorldGenerationHostParamsDocs `json:"host_params"`
	PlayerOrder      []int64                       `json:"player_order"`
	Spawns           []WorldPlayerSpawnDocs        `json:"spawns"`
	Tiles            []WorldTileDocs               `json:"tiles"`
}

// WorldSpaceLayerErrorResponseDocs documents the fixed JSON placeholder for the future space layer.
type WorldSpaceLayerErrorResponseDocs struct {
	Error string `json:"error" example:"space layer is not supported for CTP2 worlds"`
}
