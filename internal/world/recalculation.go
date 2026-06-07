package world

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"
)

var recalculationEpoch = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

// TurnStateStore exposes the persisted world turn-state seam used by recalculation.
type TurnStateStore interface {
	Read(gameID int64) (*StoredWorld, error)
	UpdateTurnState(gameID int64, turn MutableTurnState) error
}

// TurnRecalculationService deterministically recalculates only mutable world turn state.
type TurnRecalculationService struct {
	store TurnStateStore

	mu    sync.Mutex
	locks map[int64]*sync.Mutex
}

func NewTurnRecalculationService(store TurnStateStore) *TurnRecalculationService {
	return &TurnRecalculationService{
		store: store,
		locks: make(map[int64]*sync.Mutex),
	}
}

func (s *TurnRecalculationService) Recalculate(gameID int64) (*StoredWorld, error) {
	if s == nil || s.store == nil {
		return nil, errors.New("turn recalculation service requires a store")
	}
	if gameID <= 0 {
		return nil, errors.New("game_id must be positive")
	}

	lock := s.lockForGame(gameID)
	lock.Lock()
	defer lock.Unlock()

	stored, err := s.store.Read(gameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("read world for recalculation: %w", err)
	}

	nextTurn := RecalculateMutableTurnState(stored.Snapshot, stored.Snapshot.Turn)
	if err := s.store.UpdateTurnState(gameID, nextTurn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("update world turn state for recalculation: %w", err)
	}

	updated := cloneStoredWorld(stored)
	updated.Snapshot.Turn = nextTurn
	if updated.TurnUpdatedAt.Before(nextTurn.LastRecalculatedAt) {
		updated.TurnUpdatedAt = nextTurn.LastRecalculatedAt
	}

	return updated, nil
}

func (s *TurnRecalculationService) lockForGame(gameID int64) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	if lock, ok := s.locks[gameID]; ok {
		return lock
	}

	lock := &sync.Mutex{}
	s.locks[gameID] = lock
	return lock
}

func RecalculateMutableTurnState(snapshot WorldSnapshot, turn MutableTurnState) MutableTurnState {
	next := turn.Clone()
	next.RecalculationCount++
	next.LastRecalculatedAt = deterministicRecalculatedAt(turn)
	next.Pollution = recalculatePollution(snapshot, turn, next.LastRecalculatedAt)
	next.Ratings = recalculateRatings(snapshot, next)
	if len(next.ActivePlayerOrder) == 0 && len(snapshot.PlayerOrder) > 0 {
		next.ActivePlayerOrder = append([]int64(nil), snapshot.PlayerOrder...)
	}
	return next
}

func recalculatePollution(snapshot WorldSnapshot, turn MutableTurnState, updatedAt time.Time) WorldPollution {
	terrainWeight := 0
	tilePollutionCount := 0
	for _, tile := range snapshot.Tiles {
		switch tile.Terrain {
		case "swamp":
			terrainWeight += 3
			tilePollutionCount++
		case "mountain":
			terrainWeight += 2
			tilePollutionCount++
		case "desert":
			terrainWeight += 1
			tilePollutionCount++
		case "forest":
			terrainWeight += 1
		}
	}

	activePlayers := len(turn.ActivePlayerOrder)
	if activePlayers == 0 {
		activePlayers = len(snapshot.PlayerOrder)
	}

	globalPollution := terrainWeight + activePlayers*maxInt(turn.TurnNumber, 1) + turn.CurrentPlayerIndex
	return WorldPollution{
		GlobalPollution:    globalPollution,
		TilePollutionCount: tilePollutionCount,
		UpdatedAt:          updatedAt,
	}
}

func recalculateRatings(snapshot WorldSnapshot, turn MutableTurnState) WorldRatings {
	landTiles := 0
	for _, tile := range snapshot.Tiles {
		if isLandTerrain(tile.Terrain) {
			landTiles++
		}
	}

	activePlayers := len(turn.ActivePlayerOrder)
	if activePlayers == 0 {
		activePlayers = len(snapshot.PlayerOrder)
	}

	science := len(snapshot.PlayerOrder)*2 + maxInt(turn.TurnNumber, 1)
	military := len(snapshot.Spawns)*3 + activePlayers
	economic := landTiles + maxInt(snapshot.Config.StartingGold, 0)/100
	overall := science + military + economic - turn.Pollution.TilePollutionCount
	if overall < 0 {
		overall = 0
	}

	return WorldRatings{
		Overall:   overall,
		Science:   science,
		Military:  military,
		Economic:  economic,
		Pollution: turn.Pollution.GlobalPollution,
	}
}

func deterministicRecalculatedAt(turn MutableTurnState) time.Time {
	base := turn.LastRecalculatedAt.UTC()
	if base.IsZero() {
		base = turn.Pollution.UpdatedAt.UTC()
	}
	if base.IsZero() {
		base = recalculationEpoch
	}

	step := time.Duration(maxInt(turn.TurnNumber, 1)+turn.RecalculationCount+1) * time.Second
	return base.Add(step)
}

func cloneStoredWorld(stored *StoredWorld) *StoredWorld {
	if stored == nil {
		return nil
	}
	clone := *stored
	clone.Snapshot = stored.Snapshot.Clone()
	return &clone
}

func maxInt(value, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}
