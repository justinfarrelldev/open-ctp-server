package game

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

func TestWorldRatingsEndpoint(t *testing.T) {
	t.Run("returns persisted ratings for joined player", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)
		snapshot := world.FixtureWorldSnapshot()
		turn := snapshot.Turn.Clone()
		snapshotCreatedAt := time.Date(2026, time.June, 6, 18, 0, 0, 0, time.UTC)
		snapshotUpdatedAt := snapshotCreatedAt.Add(2 * time.Minute)
		turnCreatedAt := snapshotCreatedAt.Add(3 * time.Minute)
		turnUpdatedAt := snapshotCreatedAt.Add(4 * time.Minute)

		expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		expectWorldRead(mock, snapshot, turn, snapshotCreatedAt, snapshotUpdatedAt, turnCreatedAt, turnUpdatedAt)

		req := httptest.NewRequest(http.MethodGet, "/world/44/ratings?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()

		GetWorldRatingsHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusOK {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
		}
		if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
			t.Fatalf("got content type %q want application/json", contentType)
		}

		var response WorldRatingsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := WorldRatingsResponse{
			Overall:   snapshot.Turn.Ratings.Overall,
			Science:   snapshot.Turn.Ratings.Science,
			Military:  snapshot.Turn.Ratings.Military,
			Economic:  snapshot.Turn.Ratings.Economic,
			Pollution: snapshot.Turn.Ratings.Pollution,
		}
		if response != want {
			t.Fatalf("got ratings %+v want %+v", response, want)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("rejects non-member access", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectGameSessionLookup(mock, 55, 7)
		mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
			WithArgs(int64(44)).
			WillReturnRows(gameRowWithLobby(44, 7, nil, "started", "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).
			WithArgs(int64(44), int64(7)).
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/world/44/ratings?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()

		GetWorldRatingsHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusForbidden, rr.Body.String())
		}
		if strings.TrimSpace(rr.Body.String()) != ErrorGamePlayerRequired {
			t.Fatalf("got body %q want %q", rr.Body.String(), ErrorGamePlayerRequired)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("returns not found when persisted world is missing", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		mock.ExpectQuery(regexp.QuoteMeta(testWorldSnapshotSelectByGameIDQuery)).
			WithArgs(int64(44)).
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/world/44/ratings?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()

		GetWorldRatingsHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusNotFound, rr.Body.String())
		}
		if strings.TrimSpace(rr.Body.String()) != "world for game 44 was not found" {
			t.Fatalf("got body %q want %q", rr.Body.String(), "world for game 44 was not found")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})
}

func TestWorldPollutionEndpoint(t *testing.T) {
	t.Run("returns persisted pollution summary for joined player", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)
		snapshot := world.FixtureWorldSnapshot()
		turn := snapshot.Turn.Clone()
		snapshotCreatedAt := time.Date(2026, time.June, 6, 18, 0, 0, 0, time.UTC)
		snapshotUpdatedAt := snapshotCreatedAt.Add(2 * time.Minute)
		turnCreatedAt := snapshotCreatedAt.Add(3 * time.Minute)
		turnUpdatedAt := snapshotCreatedAt.Add(4 * time.Minute)

		expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		expectWorldRead(mock, snapshot, turn, snapshotCreatedAt, snapshotUpdatedAt, turnCreatedAt, turnUpdatedAt)

		req := httptest.NewRequest(http.MethodGet, "/world/44/pollution?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()

		GetWorldPollutionHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusOK {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
		}

		var response WorldPollutionResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if response.GameID != snapshot.GameID {
			t.Fatalf("got game_id %d want %d", response.GameID, snapshot.GameID)
		}
		if response.Turn != snapshot.Turn.TurnNumber {
			t.Fatalf("got turn %d want %d", response.Turn, snapshot.Turn.TurnNumber)
		}
		if response.GlobalPollution != snapshot.Turn.Pollution.GlobalPollution {
			t.Fatalf("got global_pollution %d want %d", response.GlobalPollution, snapshot.Turn.Pollution.GlobalPollution)
		}
		if response.TilePollutionCount != snapshot.Turn.Pollution.TilePollutionCount {
			t.Fatalf("got tile_pollution_count %d want %d", response.TilePollutionCount, snapshot.Turn.Pollution.TilePollutionCount)
		}
		if !response.UpdatedAt.Equal(snapshot.Turn.Pollution.UpdatedAt) {
			t.Fatalf("got updated_at %s want %s", response.UpdatedAt, snapshot.Turn.Pollution.UpdatedAt)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("falls back to last recalculated time when pollution update time is unset", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)
		snapshot := world.FixtureWorldSnapshot()
		turn := snapshot.Turn.Clone()
		fallback := time.Date(2026, time.July, 8, 9, 10, 11, 0, time.UTC)
		snapshot.Turn.Pollution.UpdatedAt = time.Time{}
		snapshot.Turn.LastRecalculatedAt = fallback
		turn.Pollution.UpdatedAt = time.Time{}
		turn.LastRecalculatedAt = fallback
		snapshotCreatedAt := time.Date(2026, time.June, 6, 18, 0, 0, 0, time.UTC)
		snapshotUpdatedAt := snapshotCreatedAt.Add(2 * time.Minute)
		turnCreatedAt := snapshotCreatedAt.Add(3 * time.Minute)
		turnUpdatedAt := snapshotCreatedAt.Add(4 * time.Minute)

		expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		expectWorldRead(mock, snapshot, turn, snapshotCreatedAt, snapshotUpdatedAt, turnCreatedAt, turnUpdatedAt)

		req := httptest.NewRequest(http.MethodGet, "/world/44/pollution?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()

		GetWorldPollutionHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusOK {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
		}

		var response WorldPollutionResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if !response.UpdatedAt.Equal(fallback) {
			t.Fatalf("got updated_at %s want %s", response.UpdatedAt, fallback)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("rejects setup games before world reads become active", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectGameSessionLookup(mock, 55, 7)
		mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
			WithArgs(int64(44)).
			WillReturnRows(gameRowWithLobby(44, 7, nil, DefaultGameStatus, "Task 11 Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).
			WithArgs(int64(44), int64(7)).
			WillReturnRows(gamePlayerRows(44,
				struct {
					id        int64
					accountID int64
					country   string
					ping      int
				}{id: 1, accountID: 7, country: "Greeks", ping: 0},
			))

		req := httptest.NewRequest(http.MethodGet, "/world/44/pollution?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()

		GetWorldPollutionHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusConflict {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusConflict, rr.Body.String())
		}
		if strings.TrimSpace(rr.Body.String()) != "game 44 has not started" {
			t.Fatalf("got body %q want %q", rr.Body.String(), "game 44 has not started")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})
}
