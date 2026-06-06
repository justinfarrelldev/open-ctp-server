package world

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

var ErrWorldAlreadyExists = errors.New("world already exists")

const worldSnapshotExistsQuery = `SELECT game_id FROM world_snapshots WHERE game_id = $1`

const worldSnapshotInsertQuery = `INSERT INTO world_snapshots (
	game_id,
	game_status,
	generator_version,
	style,
	rules_json,
	exclusions_json,
	starting_settlers,
	starting_gold,
	seed,
	wetness,
	map_size,
	climate,
	player_order_json,
	spawns_json,
	tiles_json,
	updated_at
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5::jsonb,
	$6::jsonb,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13::jsonb,
	$14::jsonb,
	$15::jsonb,
	NOW()
)`

const worldTurnStateInsertQuery = `INSERT INTO world_turn_states (
	game_id,
	turn_number,
	current_player_index,
	active_player_order_json,
	recalculation_count,
	rng_state,
	ratings_json,
	pollution_json,
	last_recalculated_at,
	updated_at
) VALUES (
	$1,
	$2,
	$3,
	$4::jsonb,
	$5,
	$6,
	$7::jsonb,
	$8::jsonb,
	$9,
	NOW()
)`

const worldSnapshotSelectByGameIDQuery = `SELECT
	game_id,
	created_at,
	updated_at,
	game_status,
	generator_version,
	style,
	rules_json::text AS rules_json,
	exclusions_json::text AS exclusions_json,
	starting_settlers,
	starting_gold,
	seed,
	wetness,
	map_size,
	climate,
	player_order_json::text AS player_order_json,
	spawns_json::text AS spawns_json,
	tiles_json::text AS tiles_json
FROM world_snapshots
WHERE game_id = $1`

const worldTurnStateSelectByGameIDQuery = `SELECT
	game_id,
	created_at,
	updated_at,
	turn_number,
	current_player_index,
	active_player_order_json::text AS active_player_order_json,
	recalculation_count,
	rng_state,
	ratings_json::text AS ratings_json,
	pollution_json::text AS pollution_json,
	last_recalculated_at
FROM world_turn_states
WHERE game_id = $1`

const worldTurnStateUpdateQuery = `UPDATE world_turn_states
SET
	turn_number = $1,
	current_player_index = $2,
	active_player_order_json = $3::jsonb,
	recalculation_count = $4,
	rng_state = $5,
	ratings_json = $6::jsonb,
	pollution_json = $7::jsonb,
	last_recalculated_at = $8,
	updated_at = NOW()
WHERE game_id = $9`

type worldGetter interface {
	Get(dest any, query string, args ...any) error
}

// Store persists world snapshots separately from mutable turn-state rows.
type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(snapshot WorldSnapshot) error {
	if s == nil || s.db == nil {
		return errors.New("world store requires a database")
	}

	normalized, err := normalizeWorldSnapshotForStorage(snapshot)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("an error occurred while starting world create: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var existingGameID int64
	if err := tx.Get(&existingGameID, worldSnapshotExistsQuery, normalized.GameID); err == nil {
		return fmt.Errorf("%w for game %d", ErrWorldAlreadyExists, normalized.GameID)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("an error occurred while checking for an existing world for game %d: %v", normalized.GameID, err)
	}

	if err := insertWorldSnapshot(tx, normalized); err != nil {
		return err
	}
	if err := insertWorldTurnState(tx, normalized.GameID, normalized.Turn); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("an error occurred while finalizing world create: %v", err)
	}
	committed = true
	return nil
}

func (s *Store) Read(gameID int64) (*StoredWorld, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("world store requires a database")
	}
	if gameID <= 0 {
		return nil, errors.New("game_id must be positive")
	}

	snapshotRow, err := getWorldSnapshotRow(s.db, gameID)
	if err != nil {
		return nil, err
	}
	turnRow, err := getWorldTurnStateRow(s.db, gameID)
	if err != nil {
		return nil, err
	}

	return assembleStoredWorld(snapshotRow, turnRow)
}

func (s *Store) UpdateTurnState(gameID int64, turn MutableTurnState) error {
	if s == nil || s.db == nil {
		return errors.New("world store requires a database")
	}
	if gameID <= 0 {
		return errors.New("game_id must be positive")
	}

	normalizedTurn := turn.Clone()
	activePlayerOrderJSON, err := CanonicalString(normalizedTurn.ActivePlayerOrder)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world active player order: %v", err)
	}
	ratingsJSON, err := CanonicalString(normalizedTurn.Ratings)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world ratings: %v", err)
	}
	pollutionJSON, err := CanonicalString(normalizedTurn.Pollution)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world pollution: %v", err)
	}

	result, err := s.db.Exec(
		worldTurnStateUpdateQuery,
		normalizedTurn.TurnNumber,
		normalizedTurn.CurrentPlayerIndex,
		activePlayerOrderJSON,
		normalizedTurn.RecalculationCount,
		strconv.FormatUint(normalizedTurn.RNGState, 10),
		ratingsJSON,
		pollutionJSON,
		normalizedTurn.LastRecalculatedAt,
		gameID,
	)
	if err != nil {
		return fmt.Errorf("an error occurred while updating world turn state for game %d: %v", gameID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("an error occurred while confirming world turn state update for game %d: %v", gameID, err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func insertWorldSnapshot(executor sqlx.Execer, snapshot WorldSnapshot) error {
	playerOrderJSON, err := CanonicalString(snapshot.PlayerOrder)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world player order: %v", err)
	}
	spawnsJSON, err := CanonicalString(snapshot.Spawns)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world spawns: %v", err)
	}
	tilesJSON, err := CanonicalString(snapshot.Tiles)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world tiles: %v", err)
	}

	if _, err := executor.Exec(
		worldSnapshotInsertQuery,
		snapshot.GameID,
		snapshot.GameStatus,
		snapshot.GeneratorVersion,
		snapshot.Config.Style,
		snapshot.Config.RulesJSON,
		snapshot.Config.ExclusionsJSON,
		snapshot.Config.StartingSettlers,
		snapshot.Config.StartingGold,
		snapshot.Config.Seed,
		snapshot.Config.Wetness,
		snapshot.Config.MapSize,
		snapshot.Config.Climate,
		playerOrderJSON,
		spawnsJSON,
		tilesJSON,
	); err != nil {
		return fmt.Errorf("an error occurred while storing the world snapshot in the database: %v", err)
	}

	return nil
}

func insertWorldTurnState(executor sqlx.Execer, gameID int64, turn MutableTurnState) error {
	normalizedTurn := turn.Clone()
	activePlayerOrderJSON, err := CanonicalString(normalizedTurn.ActivePlayerOrder)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world active player order: %v", err)
	}
	ratingsJSON, err := CanonicalString(normalizedTurn.Ratings)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world ratings: %v", err)
	}
	pollutionJSON, err := CanonicalString(normalizedTurn.Pollution)
	if err != nil {
		return fmt.Errorf("an error occurred while encoding world pollution: %v", err)
	}

	if _, err := executor.Exec(
		worldTurnStateInsertQuery,
		gameID,
		normalizedTurn.TurnNumber,
		normalizedTurn.CurrentPlayerIndex,
		activePlayerOrderJSON,
		normalizedTurn.RecalculationCount,
		strconv.FormatUint(normalizedTurn.RNGState, 10),
		ratingsJSON,
		pollutionJSON,
		normalizedTurn.LastRecalculatedAt,
	); err != nil {
		return fmt.Errorf("an error occurred while storing the world turn state in the database: %v", err)
	}

	return nil
}

func getWorldSnapshotRow(executor worldGetter, gameID int64) (worldSnapshotRow, error) {
	var row worldSnapshotRow
	if err := executor.Get(&row, worldSnapshotSelectByGameIDQuery, gameID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return worldSnapshotRow{}, err
		}
		return worldSnapshotRow{}, fmt.Errorf("an error occurred while loading the world snapshot for game %d: %v", gameID, err)
	}
	return row, nil
}

func getWorldTurnStateRow(executor worldGetter, gameID int64) (worldTurnStateRow, error) {
	var row worldTurnStateRow
	if err := executor.Get(&row, worldTurnStateSelectByGameIDQuery, gameID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return worldTurnStateRow{}, err
		}
		return worldTurnStateRow{}, fmt.Errorf("an error occurred while loading the world turn state for game %d: %v", gameID, err)
	}
	return row, nil
}

func assembleStoredWorld(snapshotRow worldSnapshotRow, turnRow worldTurnStateRow) (*StoredWorld, error) {
	playerOrder, err := decodeJSONSlice[int64](snapshotRow.PlayerOrderJSON)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decoding world player order for game %d: %v", snapshotRow.GameID, err)
	}
	spawns, err := decodeJSONSlice[PlayerSpawn](snapshotRow.SpawnsJSON)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decoding world spawns for game %d: %v", snapshotRow.GameID, err)
	}
	tiles, err := decodeJSONSlice[WorldTile](snapshotRow.TilesJSON)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decoding world tiles for game %d: %v", snapshotRow.GameID, err)
	}
	activePlayerOrder, err := decodeJSONSlice[int64](turnRow.ActivePlayerOrderJSON)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decoding active player order for game %d: %v", snapshotRow.GameID, err)
	}

	var ratings WorldRatings
	if err := decodeJSONObject(turnRow.RatingsJSON, &ratings); err != nil {
		return nil, fmt.Errorf("an error occurred while decoding world ratings for game %d: %v", snapshotRow.GameID, err)
	}
	var pollution WorldPollution
	if err := decodeJSONObject(turnRow.PollutionJSON, &pollution); err != nil {
		return nil, fmt.Errorf("an error occurred while decoding world pollution for game %d: %v", snapshotRow.GameID, err)
	}

	rngState, err := strconv.ParseUint(strings.TrimSpace(turnRow.RNGState), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decoding RNG state for game %d: %v", snapshotRow.GameID, err)
	}

	config := WorldConfig{
		Style:            snapshotRow.Style,
		RulesJSON:        snapshotRow.RulesJSON,
		ExclusionsJSON:   snapshotRow.ExclusionsJSON,
		StartingSettlers: snapshotRow.StartingSettlers,
		StartingGold:     snapshotRow.StartingGold,
		Seed:             snapshotRow.Seed,
		Wetness:          snapshotRow.Wetness,
		MapSize:          snapshotRow.MapSize,
		Climate:          snapshotRow.Climate,
		GeneratorVersion: snapshotRow.GeneratorVersion,
	}

	stored := &StoredWorld{
		Snapshot: WorldSnapshot{
			GameID:           snapshotRow.GameID,
			GameStatus:       snapshotRow.GameStatus,
			GeneratorVersion: snapshotRow.GeneratorVersion,
			Config:           config,
			PlayerOrder:      playerOrder,
			Spawns:           spawns,
			Tiles:            tiles,
			Turn: MutableTurnState{
				TurnNumber:         turnRow.TurnNumber,
				CurrentPlayerIndex: turnRow.CurrentPlayerIndex,
				ActivePlayerOrder:  activePlayerOrder,
				RecalculationCount: turnRow.RecalculationCount,
				RNGState:           rngState,
				Ratings:            ratings,
				Pollution:          pollution,
				LastRecalculatedAt: turnRow.LastRecalculatedAt,
			},
		},
		SnapshotCreatedAt: snapshotRow.CreatedAt,
		SnapshotUpdatedAt: snapshotRow.UpdatedAt,
		TurnCreatedAt:     turnRow.CreatedAt,
		TurnUpdatedAt:     turnRow.UpdatedAt,
	}

	return stored, nil
}

func normalizeWorldSnapshotForStorage(snapshot WorldSnapshot) (WorldSnapshot, error) {
	normalized := snapshot.Clone()
	if normalized.GameID <= 0 {
		return WorldSnapshot{}, errors.New("game_id must be positive")
	}

	config, err := NormalizeWorldConfig(normalized.Config)
	if err != nil {
		return WorldSnapshot{}, fmt.Errorf("normalize world config: %w", err)
	}
	normalized.Config = config
	normalized.GameStatus = strings.TrimSpace(normalized.GameStatus)
	normalized.GeneratorVersion = strings.TrimSpace(normalized.GeneratorVersion)
	if normalized.GeneratorVersion == "" {
		normalized.GeneratorVersion = normalized.Config.GeneratorVersion
	}
	if normalized.GeneratorVersion == "" {
		return WorldSnapshot{}, errors.New("generator_version must be specified")
	}
	if normalized.GameStatus == "" {
		return WorldSnapshot{}, errors.New("game_status must be specified")
	}

	return normalized, nil
}

func decodeJSONSlice[T any](raw string) ([]T, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var values []T
	if err := json.Unmarshal([]byte(trimmed), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func decodeJSONObject(raw string, dest any) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = "{}"
	}
	return json.Unmarshal([]byte(trimmed), dest)
}
