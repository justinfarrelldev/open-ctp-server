package world

import (
	"database/sql"
	"sync"
	"testing"
	"time"
)

func TestTurnRecalculationDeterministic(t *testing.T) {
	firstStore := newRecalculationMemoryStore(FixtureWorldSnapshot())
	secondStore := newRecalculationMemoryStore(FixtureWorldSnapshot())

	firstResult, err := NewTurnRecalculationService(firstStore).Recalculate(44)
	if err != nil {
		t.Fatalf("first Recalculate() error = %v", err)
	}
	secondResult, err := NewTurnRecalculationService(secondStore).Recalculate(44)
	if err != nil {
		t.Fatalf("second Recalculate() error = %v", err)
	}

	AssertCanonicalJSONEqual(t, firstResult.Snapshot.Turn, secondResult.Snapshot.Turn)
	AssertCanonicalJSONStable(t, firstResult.Snapshot.Turn)
}

func TestTurnRecalculationMutatesOnlyTurnState(t *testing.T) {
	initial := FixtureWorldSnapshot()
	store := newRecalculationMemoryStore(initial)

	updated, err := NewTurnRecalculationService(store).Recalculate(initial.GameID)
	if err != nil {
		t.Fatalf("Recalculate() error = %v", err)
	}

	if updated.Snapshot.Turn.RecalculationCount != initial.Turn.RecalculationCount+1 {
		t.Fatalf("recalculation_count=%d want %d", updated.Snapshot.Turn.RecalculationCount, initial.Turn.RecalculationCount+1)
	}
	if !updated.Snapshot.Turn.LastRecalculatedAt.After(initial.Turn.LastRecalculatedAt) {
		t.Fatalf("last_recalculated_at=%s must be after %s", updated.Snapshot.Turn.LastRecalculatedAt, initial.Turn.LastRecalculatedAt)
	}
	if updated.Snapshot.Turn.Pollution.UpdatedAt != updated.Snapshot.Turn.LastRecalculatedAt {
		t.Fatalf("pollution updated_at=%s want %s", updated.Snapshot.Turn.Pollution.UpdatedAt, updated.Snapshot.Turn.LastRecalculatedAt)
	}

	gotImmutable := updated.Snapshot.Clone()
	gotImmutable.Turn = initial.Turn.Clone()
	AssertCanonicalJSONEqual(t, gotImmutable, initial)
}

func TestRecalcRequiresExistingWorld(t *testing.T) {
	service := NewTurnRecalculationService(&recalculationMissingStore{})

	_, err := service.Recalculate(44)
	if err == nil {
		t.Fatal("Recalculate() error = nil, want sql.ErrNoRows")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("Recalculate() error = %v want %v", err, sql.ErrNoRows)
	}
}

func TestRecalcSerialization(t *testing.T) {
	store := newRecalculationMemoryStore(FixtureWorldSnapshot())
	store.blockFirstUpdate = make(chan struct{})
	store.firstUpdateEntered = make(chan struct{})
	service := NewTurnRecalculationService(store)

	results := make(chan MutableTurnState, 2)
	errs := make(chan error, 2)

	go func() {
		updated, err := service.Recalculate(44)
		if err == nil {
			results <- updated.Snapshot.Turn
		}
		errs <- err
	}()

	<-store.firstUpdateEntered

	go func() {
		updated, err := service.Recalculate(44)
		if err == nil {
			results <- updated.Snapshot.Turn
		}
		errs <- err
	}()

	time.Sleep(50 * time.Millisecond)
	if got := store.readCountValue(); got != 1 {
		t.Fatalf("concurrent recalculation reached Read() %d times before first update completed; want serialization", got)
	}

	close(store.blockFirstUpdate)

	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("Recalculate() error = %v", err)
		}
	}
	close(results)

	counts := make(map[int]int)
	for turn := range results {
		counts[turn.RecalculationCount]++
	}
	if counts[2] != 1 || counts[3] != 1 {
		t.Fatalf("unexpected recalculation counts: %+v", counts)
	}

	stored, err := store.Read(44)
	if err != nil {
		t.Fatalf("Read() after serialization error = %v", err)
	}
	if stored.Snapshot.Turn.RecalculationCount != 3 {
		t.Fatalf("final recalculation_count=%d want 3", stored.Snapshot.Turn.RecalculationCount)
	}
	AssertCanonicalJSONStable(t, stored.Snapshot.Turn)
}

type recalculationMissingStore struct{}

func (s *recalculationMissingStore) Read(gameID int64) (*StoredWorld, error) {
	return nil, sql.ErrNoRows
}

func (s *recalculationMissingStore) UpdateTurnState(gameID int64, turn MutableTurnState) error {
	return nil
}

type recalculationMemoryStore struct {
	mu                 sync.Mutex
	stored             *StoredWorld
	readCount          int
	updateCount        int
	blockFirstUpdate   chan struct{}
	firstUpdateEntered chan struct{}
}

func newRecalculationMemoryStore(snapshot WorldSnapshot) *recalculationMemoryStore {
	stored := &StoredWorld{
		Snapshot:          snapshot.Clone(),
		SnapshotCreatedAt: time.Date(2026, time.June, 6, 18, 0, 0, 0, time.UTC),
		SnapshotUpdatedAt: time.Date(2026, time.June, 6, 18, 1, 0, 0, time.UTC),
		TurnCreatedAt:     time.Date(2026, time.June, 6, 18, 2, 0, 0, time.UTC),
		TurnUpdatedAt:     time.Date(2026, time.June, 6, 18, 3, 0, 0, time.UTC),
	}
	return &recalculationMemoryStore{stored: stored}
}

func (s *recalculationMemoryStore) Read(gameID int64) (*StoredWorld, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stored == nil || s.stored.Snapshot.GameID != gameID {
		return nil, sql.ErrNoRows
	}
	s.readCount++
	return cloneStoredWorld(s.stored), nil
}

func (s *recalculationMemoryStore) UpdateTurnState(gameID int64, turn MutableTurnState) error {
	s.mu.Lock()
	s.updateCount++
	updateCount := s.updateCount
	block := s.blockFirstUpdate
	entered := s.firstUpdateEntered
	s.mu.Unlock()

	if updateCount == 1 && entered != nil {
		close(entered)
	}
	if updateCount == 1 && block != nil {
		<-block
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stored == nil || s.stored.Snapshot.GameID != gameID {
		return sql.ErrNoRows
	}
	s.stored.Snapshot.Turn = turn.Clone()
	s.stored.TurnUpdatedAt = turn.LastRecalculatedAt
	return nil
}

func (s *recalculationMemoryStore) readCountValue() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readCount
}
