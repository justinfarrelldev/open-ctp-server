package world

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func newWorldSQLMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestWorldStoreCreateAndRead(t *testing.T) {
	t.Run("round-trips world snapshot and host params exactly", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		store := NewStore(sqlxDB)
		snapshot := FixtureWorldSnapshot()

		playerOrderJSON := MustCanonicalString(t, snapshot.PlayerOrder)
		spawnsJSON := MustCanonicalString(t, snapshot.Spawns)
		tilesJSON := MustCanonicalString(t, snapshot.Tiles)
		activePlayerOrderJSON := MustCanonicalString(t, snapshot.Turn.ActivePlayerOrder)
		ratingsJSON := MustCanonicalString(t, snapshot.Turn.Ratings)
		pollutionJSON := MustCanonicalString(t, snapshot.Turn.Pollution)

		snapshotCreatedAt := time.Date(2026, time.June, 6, 18, 0, 0, 0, time.UTC)
		snapshotUpdatedAt := snapshotCreatedAt.Add(2 * time.Minute)
		turnCreatedAt := snapshotCreatedAt.Add(3 * time.Minute)
		turnUpdatedAt := snapshotCreatedAt.Add(4 * time.Minute)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(worldSnapshotExistsQuery)).
			WithArgs(snapshot.GameID).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectExec(regexp.QuoteMeta(worldSnapshotInsertQuery)).
			WithArgs(
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
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(worldTurnStateInsertQuery)).
			WithArgs(
				snapshot.GameID,
				snapshot.Turn.TurnNumber,
				snapshot.Turn.CurrentPlayerIndex,
				activePlayerOrderJSON,
				snapshot.Turn.RecalculationCount,
				"99",
				ratingsJSON,
				pollutionJSON,
				snapshot.Turn.LastRecalculatedAt,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		if err := store.Create(snapshot); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		mock.ExpectQuery(regexp.QuoteMeta(worldSnapshotSelectByGameIDQuery)).
			WithArgs(snapshot.GameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"game_id",
				"created_at",
				"updated_at",
				"game_status",
				"generator_version",
				"style",
				"rules_json",
				"exclusions_json",
				"starting_settlers",
				"starting_gold",
				"seed",
				"wetness",
				"map_size",
				"climate",
				"player_order_json",
				"spawns_json",
				"tiles_json",
			}).AddRow(
				snapshot.GameID,
				snapshotCreatedAt,
				snapshotUpdatedAt,
				snapshot.GameStatus,
				snapshot.GeneratorVersion,
				snapshot.Config.Style,
				snapshot.Config.RulesJSON,
				snapshot.Config.ExclusionsJSON,
				snapshot.Config.StartingSettlers,
				snapshot.Config.StartingGold,
				*snapshot.Config.Seed,
				*snapshot.Config.Wetness,
				snapshot.Config.MapSize,
				snapshot.Config.Climate,
				playerOrderJSON,
				spawnsJSON,
				tilesJSON,
			))
		mock.ExpectQuery(regexp.QuoteMeta(worldTurnStateSelectByGameIDQuery)).
			WithArgs(snapshot.GameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"game_id",
				"created_at",
				"updated_at",
				"turn_number",
				"current_player_index",
				"active_player_order_json",
				"recalculation_count",
				"rng_state",
				"ratings_json",
				"pollution_json",
				"last_recalculated_at",
			}).AddRow(
				snapshot.GameID,
				turnCreatedAt,
				turnUpdatedAt,
				snapshot.Turn.TurnNumber,
				snapshot.Turn.CurrentPlayerIndex,
				activePlayerOrderJSON,
				snapshot.Turn.RecalculationCount,
				"99",
				ratingsJSON,
				pollutionJSON,
				snapshot.Turn.LastRecalculatedAt,
			))

		stored, err := store.Read(snapshot.GameID)
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}

		if !reflect.DeepEqual(stored.Snapshot.Config, snapshot.Config) {
			t.Fatalf("config round-trip mismatch\ngot:  %+v\nwant: %+v", stored.Snapshot.Config, snapshot.Config)
		}
		AssertCanonicalJSONEqual(t, stored.Snapshot, snapshot)
		if !stored.SnapshotCreatedAt.Equal(snapshotCreatedAt) || !stored.SnapshotUpdatedAt.Equal(snapshotUpdatedAt) {
			t.Fatalf("unexpected snapshot timestamps: %+v", stored)
		}
		if !stored.TurnCreatedAt.Equal(turnCreatedAt) || !stored.TurnUpdatedAt.Equal(turnUpdatedAt) {
			t.Fatalf("unexpected turn timestamps: %+v", stored)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("rejects duplicate worlds for the same game", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		store := NewStore(sqlxDB)
		snapshot := FixtureWorldSnapshot()

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(worldSnapshotExistsQuery)).
			WithArgs(snapshot.GameID).
			WillReturnRows(sqlmock.NewRows([]string{"game_id"}).AddRow(snapshot.GameID))
		mock.ExpectRollback()

		err := store.Create(snapshot)
		if !errors.Is(err, ErrWorldAlreadyExists) {
			t.Fatalf("Create() error = %v want ErrWorldAlreadyExists", err)
		}
		if err == nil || strings.TrimSpace(err.Error()) != "world already exists for game 44" {
			t.Fatalf("got error %q want %q", err, "world already exists for game 44")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})
}

func TestWorldStoreRollbackOnCreateFailure(t *testing.T) {
	sqlxDB, mock := newWorldSQLMock(t)
	store := NewStore(sqlxDB)
	snapshot := FixtureWorldSnapshot()

	playerOrderJSON := MustCanonicalString(t, snapshot.PlayerOrder)
	spawnsJSON := MustCanonicalString(t, snapshot.Spawns)
	tilesJSON := MustCanonicalString(t, snapshot.Tiles)
	activePlayerOrderJSON := MustCanonicalString(t, snapshot.Turn.ActivePlayerOrder)
	ratingsJSON := MustCanonicalString(t, snapshot.Turn.Ratings)
	pollutionJSON := MustCanonicalString(t, snapshot.Turn.Pollution)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(worldSnapshotExistsQuery)).
		WithArgs(snapshot.GameID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(worldSnapshotInsertQuery)).
		WithArgs(
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
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(worldTurnStateInsertQuery)).
		WithArgs(
			snapshot.GameID,
			snapshot.Turn.TurnNumber,
			snapshot.Turn.CurrentPlayerIndex,
			activePlayerOrderJSON,
			snapshot.Turn.RecalculationCount,
			"99",
			ratingsJSON,
			pollutionJSON,
			snapshot.Turn.LastRecalculatedAt,
		).
		WillReturnError(errors.New("turn insert failed"))
	mock.ExpectRollback()

	err := store.Create(snapshot)
	if err == nil {
		t.Fatal("expected create failure")
	}
	if !strings.Contains(err.Error(), "storing the world turn state") || !strings.Contains(err.Error(), "turn insert failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCreateV1InTxWritesImmutableRulesetPinWithSnapshot(t *testing.T) {
	sqlxDB, mock := newWorldSQLMock(t)
	mock.ExpectBegin()
	tx, err := sqlxDB.Beginx()
	if err != nil {
		t.Fatalf("Beginx() error = %v", err)
	}

	snapshot := FixtureWorldSnapshot()
	pin := RulesetPin{
		CanonicalJSON: []byte(`{"version":"ctp2-v1"}`),
		ContentHash:   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Version:       "ctp2-v1",
	}
	playerOrderJSON := MustCanonicalString(t, snapshot.PlayerOrder)
	spawnsJSON := MustCanonicalString(t, snapshot.Spawns)
	tilesJSON := MustCanonicalString(t, snapshot.Tiles)
	activePlayerOrderJSON := MustCanonicalString(t, snapshot.Turn.ActivePlayerOrder)
	ratingsJSON := MustCanonicalString(t, snapshot.Turn.Ratings)
	pollutionJSON := MustCanonicalString(t, snapshot.Turn.Pollution)

	mock.ExpectQuery(regexp.QuoteMeta(worldSnapshotExistsQuery)).WithArgs(snapshot.GameID).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(worldSnapshotV1InsertQuery)).WithArgs(
		snapshot.GameID, snapshot.GameStatus, snapshot.GeneratorVersion, snapshot.Config.Style,
		snapshot.Config.RulesJSON, snapshot.Config.ExclusionsJSON, snapshot.Config.StartingSettlers,
		snapshot.Config.StartingGold, snapshot.Config.Seed, snapshot.Config.Wetness, snapshot.Config.MapSize,
		snapshot.Config.Climate, playerOrderJSON, spawnsJSON, tilesJSON, string(pin.CanonicalJSON), pin.ContentHash, pin.Version,
	).WillReturnResult(sqlmock.NewResult(0, 1))
	initialCanonicalStateJSON := []byte(`{"phase":"setup"}`)
	mock.ExpectExec(regexp.QuoteMeta(worldTurnStateV1InsertQuery)).WithArgs(
		snapshot.GameID, snapshot.Turn.TurnNumber, snapshot.Turn.CurrentPlayerIndex, activePlayerOrderJSON,
		snapshot.Turn.RecalculationCount, "99", ratingsJSON, pollutionJSON, snapshot.Turn.LastRecalculatedAt,
		StateSchemaV1, int64(0), int64(0), string(initialCanonicalStateJSON),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	if err := CreateV1InTx(tx, snapshot, pin, initialCanonicalStateJSON); err != nil {
		t.Fatalf("CreateV1InTx() error = %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCreateV1InTxRejectsInvalidCanonicalStateBeforeWrites(t *testing.T) {
	sqlxDB, mock := newWorldSQLMock(t)
	mock.ExpectBegin()
	tx, err := sqlxDB.Beginx()
	if err != nil {
		t.Fatalf("Beginx() error = %v", err)
	}
	mock.ExpectRollback()

	pin := RulesetPin{
		CanonicalJSON: []byte(`{"version":"ctp2-v1"}`),
		ContentHash:   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Version:       "ctp2-v1",
	}
	if err := CreateV1InTx(tx, FixtureWorldSnapshot(), pin, []byte(`[]`)); err == nil {
		t.Fatal("CreateV1InTx() accepted array canonical state")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("invalid canonical state performed partial SQL writes: %v", err)
	}
}
