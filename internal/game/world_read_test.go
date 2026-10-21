package game

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

const testWorldSnapshotSelectByGameIDQuery = `SELECT
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

const testWorldTurnStateSelectByGameIDQuery = `SELECT
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

// Authenticated membership is not permission to download hidden geography, the
// generation seed or foreign spawns. No snapshot query may run on this endpoint.
func TestGetWorldRejectsCanonicalMapDownload(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
	req.SetPathValue("game_id", "44")
	rr := httptest.NewRecorder()
	GetWorldHandler(rr, req, sqlxDB, store)
	if rr.Code != http.StatusGone || !strings.Contains(rr.Body.String(), "viewer-safe") {
		t.Fatalf("canonical map endpoint remains accessible: %d %s", rr.Code, rr.Body.String())
	}
	for _, secret := range []string{"\"tiles\"", "\"seed\"", "\"spawns\""} {
		if strings.Contains(rr.Body.String(), secret) {
			t.Fatalf("canonical data leaked: %s", rr.Body.String())
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetWorldRequiresAuthAndMembership(t *testing.T) {
	t.Run("requires auth", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		req := httptest.NewRequest(http.MethodGet, "/world/44", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()
		GetWorldHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
		if strings.TrimSpace(rr.Body.String()) != auth.ErrSessionIDRequired.Error() {
			t.Fatalf("got body %q want %q", rr.Body.String(), auth.ErrSessionIDRequired.Error())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("requires membership", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectGameSessionLookup(mock, 55, 7)
		mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
			WithArgs(int64(44)).
			WillReturnRows(gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).
			WithArgs(int64(44), int64(7)).
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()
		GetWorldHandler(rr, req, sqlxDB, store)

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
}

func TestRetiredWorldEndpointDoesNotProbeSnapshotExistence(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
	req.SetPathValue("game_id", "44")
	rr := httptest.NewRecorder()
	GetWorldHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusGone {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusGone, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestWorldSpaceLayerPlaceholder(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))

	req := httptest.NewRequest(http.MethodGet, "/world/44/space?session_id=55", nil)
	req.SetPathValue("game_id", "44")
	rr := httptest.NewRecorder()
	GetWorldSpaceLayerHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("got content type %q want application/json", contentType)
	}
	if strings.TrimSpace(rr.Body.String()) != `{"error":"space layer is not supported for CTP2 worlds"}` {
		t.Fatalf("got body %q want %q", rr.Body.String(), `{"error":"space layer is not supported for CTP2 worlds"}`)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func expectReadableWorldAccess(mock sqlmock.Sqlmock, sessionID int64, accountID int, gameRows *sqlmock.Rows) {
	expectGameSessionLookup(mock, sessionID, accountID)
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRows)
	mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).
		WithArgs(int64(44), int64(accountID)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: int64(accountID), country: "Greeks", ping: 0},
		))
}

func expectWorldRead(mock sqlmock.Sqlmock, snapshot world.WorldSnapshot, turn world.MutableTurnState, snapshotCreatedAt time.Time, snapshotUpdatedAt time.Time, turnCreatedAt time.Time, turnUpdatedAt time.Time) {
	playerOrderJSON, err := world.CanonicalString(snapshot.PlayerOrder)
	if err != nil {
		panic(err)
	}
	spawnsJSON, err := world.CanonicalString(snapshot.Spawns)
	if err != nil {
		panic(err)
	}
	tilesJSON, err := world.CanonicalString(snapshot.Tiles)
	if err != nil {
		panic(err)
	}
	activePlayerOrderJSON, err := world.CanonicalString(turn.ActivePlayerOrder)
	if err != nil {
		panic(err)
	}
	ratingsJSON, err := world.CanonicalString(turn.Ratings)
	if err != nil {
		panic(err)
	}
	pollutionJSON, err := world.CanonicalString(turn.Pollution)
	if err != nil {
		panic(err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(testWorldSnapshotSelectByGameIDQuery)).
		WithArgs(snapshot.GameID).
		WillReturnRows(sqlmock.NewRows([]string{"game_id", "created_at", "updated_at", "game_status", "generator_version", "style", "rules_json", "exclusions_json", "starting_settlers", "starting_gold", "seed", "wetness", "map_size", "climate", "player_order_json", "spawns_json", "tiles_json"}).
			AddRow(snapshot.GameID, snapshotCreatedAt, snapshotUpdatedAt, snapshot.GameStatus, snapshot.GeneratorVersion, snapshot.Config.Style, snapshot.Config.RulesJSON, snapshot.Config.ExclusionsJSON, snapshot.Config.StartingSettlers, snapshot.Config.StartingGold, snapshot.Config.Seed, snapshot.Config.Wetness, snapshot.Config.MapSize, snapshot.Config.Climate, playerOrderJSON, spawnsJSON, tilesJSON))
	mock.ExpectQuery(regexp.QuoteMeta(testWorldTurnStateSelectByGameIDQuery)).
		WithArgs(snapshot.GameID).
		WillReturnRows(sqlmock.NewRows([]string{"game_id", "created_at", "updated_at", "turn_number", "current_player_index", "active_player_order_json", "recalculation_count", "rng_state", "ratings_json", "pollution_json", "last_recalculated_at"}).
			AddRow(snapshot.GameID, turnCreatedAt, turnUpdatedAt, turn.TurnNumber, turn.CurrentPlayerIndex, activePlayerOrderJSON, turn.RecalculationCount, "99", ratingsJSON, pollutionJSON, turn.LastRecalculatedAt))
}
