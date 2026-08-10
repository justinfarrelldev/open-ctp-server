package world

import "time"

// WorldRatings is the deterministic, read-only rating summary scaffold.
type WorldRatings struct {
	Overall   int `json:"overall"`
	Science   int `json:"science"`
	Military  int `json:"military"`
	Economic  int `json:"economic"`
	Pollution int `json:"pollution"`
}

// WorldPollution is the deterministic pollution summary scaffold.
type WorldPollution struct {
	GlobalPollution    int       `json:"global_pollution"`
	TilePollutionCount int       `json:"tile_pollution_count"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MutableTurnState is the per-turn mutable state scaffold for future recalculation work.
type MutableTurnState struct {
	TurnNumber         int            `json:"turn_number"`
	CurrentPlayerIndex int            `json:"current_player_index"`
	ActivePlayerOrder  []int64        `json:"active_player_order,omitempty"`
	RecalculationCount int            `json:"recalculation_count"`
	RNGState           uint64         `json:"rng_state"`
	Ratings            WorldRatings   `json:"ratings"`
	Pollution          WorldPollution `json:"pollution"`
	LastRecalculatedAt time.Time      `json:"last_recalculated_at"`
}

// Clone returns a deep copy of the mutable turn state.
func (s MutableTurnState) Clone() MutableTurnState {
	clone := s
	if s.ActivePlayerOrder != nil {
		clone.ActivePlayerOrder = append([]int64(nil), s.ActivePlayerOrder...)
	}
	return clone
}

// WorldSnapshot is the immutable world state persisted and serialized later.
// RulesetPin is the immutable canonical import identity supplied when creating a v1 world snapshot.
type RulesetPin struct {
	CanonicalJSON []byte
	ContentHash   string
	Version       string
}

type WorldSnapshot struct {
	GameID           int64            `json:"game_id"`
	GameStatus       string           `json:"game_status"`
	GeneratorVersion string           `json:"generator_version,omitempty"`
	Config           WorldConfig      `json:"config"`
	PlayerOrder      []int64          `json:"player_order"`
	Spawns           []PlayerSpawn    `json:"spawns,omitempty"`
	Tiles            []WorldTile      `json:"tiles,omitempty"`
	Turn             MutableTurnState `json:"turn_state"`
}

// Clone returns a deep copy of the world snapshot.
func (s WorldSnapshot) Clone() WorldSnapshot {
	clone := s
	clone.Config = s.Config.Clone()
	clone.Turn = s.Turn.Clone()
	if s.PlayerOrder != nil {
		clone.PlayerOrder = append([]int64(nil), s.PlayerOrder...)
	}
	if s.Spawns != nil {
		clone.Spawns = append([]PlayerSpawn(nil), s.Spawns...)
	}
	if s.Tiles != nil {
		clone.Tiles = append([]WorldTile(nil), s.Tiles...)
	}
	return clone
}

// CanonicalJSON returns the stable JSON encoding of the snapshot.
func (s WorldSnapshot) CanonicalJSON() ([]byte, error) {
	return CanonicalJSON(s)
}

// CanonicalJSON returns the stable JSON encoding of the turn state.
func (s MutableTurnState) CanonicalJSON() ([]byte, error) {
	return CanonicalJSON(s)
}
