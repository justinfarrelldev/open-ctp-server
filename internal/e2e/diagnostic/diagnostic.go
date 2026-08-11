// Package diagnostic provides detached, read-only snapshots for E2E assertions.
package diagnostic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

const stateQuery = `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`
const playersQuery = `SELECT id, created_at, updated_at, game_id, account_id, joined_at, country, ping_ms, controller_kind FROM game_players WHERE game_id = $1 ORDER BY joined_at ASC, id ASC`

type Snapshot struct {
	Game        Game
	Players     []Player
	Countries   []Country
	Cities      []City
	Units       []Unit
	Controllers []Controller
	Phase       Phase
	Revision    int64
	Terminal    Terminal
	EventCursor int64
}

type Game struct {
	ID               int64
	StateSchema      int
	RulesVersion     string
	RulesContentHash string
}

type Player struct {
	ID        string
	NationID  string
	Lifecycle string
}

type Country struct {
	PlayerID string
	RecordID string
}

type City struct {
	ID            string
	OwnerID       string
	TileID        string
	WorkedTileIDs []string
	BuildingIDs   []string
}

type Unit struct {
	ID           string
	OwnerID      string
	TileID       string
	ArmyID       string
	TransportID  string
	CargoDomains []string
	Movement     int
	Fortified    bool
}

type Controller struct {
	PlayerID string
	Kind     string
}

type Phase struct {
	Mode              string
	ModeVersion       string
	Turn              int
	ActivePlayerIndex int
	ActivePlayerOrder []string
	BaseRevision      int64
	DeadlineID        string
	ResolutionLocked  bool
}

type Terminal struct {
	Ended           bool
	Reason          string
	WinningPlayerID string
}

type playerRecord struct {
	playerID   string
	country    string
	controller string
}

func Read(ctx context.Context, db *sqlx.DB, gameID int64) (Snapshot, error) {
	if db == nil || gameID < 1 {
		return Snapshot{}, errors.New("diagnostic read is invalid")
	}
	var raw string
	if err := db.GetContext(ctx, &raw, stateQuery, gameID); err != nil {
		return Snapshot{}, fmt.Errorf("read diagnostic state: %w", err)
	}
	state, err := engine.DecodeCanonicalStateJSON([]byte(raw))
	if err != nil {
		return Snapshot{}, fmt.Errorf("decode diagnostic state: %w", err)
	}
	records, err := readPlayerRecords(ctx, db, gameID)
	if err != nil {
		return Snapshot{}, err
	}
	return detach(gameID, state, records)
}

func readPlayerRecords(ctx context.Context, db *sqlx.DB, gameID int64) (map[string]playerRecord, error) {
	rows, err := db.QueryxContext(ctx, playersQuery, gameID)
	if err != nil {
		return nil, fmt.Errorf("read diagnostic players: %w", err)
	}
	defer rows.Close()
	records := make(map[string]playerRecord)
	for rows.Next() {
		var id, returnedGameID, accountID, pingMS int64
		var createdAt, updatedAt, joinedAt time.Time
		var country, controller string
		if err := rows.Scan(&id, &createdAt, &updatedAt, &returnedGameID, &accountID, &joinedAt, &country, &pingMS, &controller); err != nil {
			return nil, fmt.Errorf("scan diagnostic player: %w", err)
		}
		if id < 1 || returnedGameID != gameID || accountID < 1 || country == "" || controller == "" {
			return nil, errors.New("diagnostic player record is invalid")
		}
		playerID := fmt.Sprintf("account-%d", accountID)
		if _, exists := records[playerID]; exists {
			return nil, fmt.Errorf("duplicate diagnostic player %s", playerID)
		}
		records[playerID] = playerRecord{playerID: playerID, country: country, controller: controller}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate diagnostic players: %w", err)
	}
	return records, nil
}

func detach(gameID int64, state engine.State, records map[string]playerRecord) (Snapshot, error) {
	snapshot := Snapshot{
		Game:    Game{ID: gameID, StateSchema: state.SchemaVersion, RulesVersion: state.Rules.Version, RulesContentHash: state.Rules.ContentHash},
		Players: make([]Player, 0, len(state.Players)), Countries: make([]Country, 0, len(state.Players)),
		Cities: make([]City, 0, len(state.Cities)), Units: make([]Unit, 0, len(state.Units)), Controllers: make([]Controller, 0, len(state.Players)),
		Phase:    Phase{Mode: state.Phase.Mode, ModeVersion: state.Phase.ModeVersion, Turn: state.Phase.Turn, ActivePlayerIndex: state.Phase.ActivePlayerIndex, ActivePlayerOrder: cloneStrings(state.Phase.ActivePlayerOrder), BaseRevision: state.Phase.BaseRevision, DeadlineID: state.Phase.DeadlineID, ResolutionLocked: state.Phase.ResolutionLocked},
		Revision: state.Revision, Terminal: Terminal{Ended: state.Terminal.Ended, Reason: state.Terminal.Reason, WinningPlayerID: state.Terminal.WinningPlayerID}, EventCursor: state.EventSequence,
	}
	for _, player := range state.Players {
		record, ok := records[player.ID]
		if !ok {
			return Snapshot{}, fmt.Errorf("diagnostic metadata missing for player %s", player.ID)
		}
		snapshot.Players = append(snapshot.Players, Player{ID: player.ID, NationID: player.NationID, Lifecycle: player.Lifecycle})
		snapshot.Countries = append(snapshot.Countries, Country{PlayerID: record.playerID, RecordID: record.country})
		snapshot.Controllers = append(snapshot.Controllers, Controller{PlayerID: record.playerID, Kind: record.controller})
	}
	if len(records) != len(state.Players) {
		return Snapshot{}, errors.New("diagnostic player metadata does not match canonical players")
	}
	for _, city := range state.Cities {
		buildings := make([]string, len(city.Buildings))
		for index := range city.Buildings {
			buildings[index] = city.Buildings[index].ID
		}
		snapshot.Cities = append(snapshot.Cities, City{ID: city.ID, OwnerID: city.OwnerID, TileID: city.TileID, WorkedTileIDs: cloneStrings(city.WorkedTileIDs), BuildingIDs: buildings})
	}
	for _, unit := range state.Units {
		snapshot.Units = append(snapshot.Units, Unit{ID: unit.ID, OwnerID: unit.OwnerID, TileID: unit.TileID, ArmyID: unit.ArmyID, TransportID: unit.TransportID, CargoDomains: cloneStrings(unit.CargoDomains), Movement: unit.Movement, Fortified: unit.Fortified})
	}
	return snapshot, nil
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string(nil), values...)
}
