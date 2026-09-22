package world

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

type rowsAffectedErrorResult struct{}

func (rowsAffectedErrorResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (rowsAffectedErrorResult) RowsAffected() (int64, error) {
	return 0, errors.New("rows affected failed")
}

type stubWorldGetter struct {
	err error
}

func (g stubWorldGetter) Get(dest any, query string, args ...any) error {
	return g.err
}

func TestWorldStoreGuardClauses(t *testing.T) {
	snapshot := FixtureWorldSnapshot()
	turn := FixtureMutableTurnState()

	if err := CreateInTx(nil, snapshot); err == nil || err.Error() != "world store requires a transaction" {
		t.Fatalf("CreateInTx(nil) error = %v", err)
	}
	if err := (&Store{}).Create(snapshot); err == nil || err.Error() != "world store requires a database" {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := (&Store{}).Read(snapshot.GameID); err == nil || err.Error() != "world store requires a database" {
		t.Fatalf("Read() error = %v", err)
	}
	if _, err := NewStore(nil).Read(0); err == nil || err.Error() != "world store requires a database" {
		t.Fatalf("Read(nil db) error = %v", err)
	}
	if _, err := NewStore(sqlmockDB(t)).Read(0); err == nil || err.Error() != "game_id must be positive" {
		t.Fatalf("Read(gameID=0) error = %v", err)
	}
	if err := (&Store{}).UpdateTurnState(snapshot.GameID, turn); err == nil || err.Error() != "world store requires a database" {
		t.Fatalf("UpdateTurnState() error = %v", err)
	}
	if err := NewStore(sqlmockDB(t)).UpdateTurnState(0, turn); err == nil || err.Error() != "game_id must be positive" {
		t.Fatalf("UpdateTurnState(gameID=0) error = %v", err)
	}
}

func TestCreateInTx_ErrorPaths(t *testing.T) {
	t.Run("normalize error", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		mock.ExpectBegin()
		mock.ExpectRollback()
		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatalf("Beginx() error = %v", err)
		}
		defer tx.Rollback()

		badSnapshot := FixtureWorldSnapshot()
		badSnapshot.GameID = 0

		err = CreateInTx(tx, badSnapshot)
		if err == nil || err.Error() != "game_id must be positive" {
			t.Fatalf("CreateInTx() error = %v", err)
		}
	})

	t.Run("existing world lookup error", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		mock.ExpectBegin()
		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatalf("Beginx() error = %v", err)
		}
		defer tx.Rollback()

		snapshot := FixtureWorldSnapshot()
		mock.ExpectQuery(regexp.QuoteMeta(worldSnapshotExistsQuery)).
			WithArgs(snapshot.GameID).
			WillReturnError(errors.New("lookup failed"))
		mock.ExpectRollback()

		err = CreateInTx(tx, snapshot)
		if err == nil || !strings.Contains(err.Error(), "checking for an existing world") {
			t.Fatalf("CreateInTx() error = %v", err)
		}
	})

	t.Run("snapshot insert error", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		mock.ExpectBegin()
		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatalf("Beginx() error = %v", err)
		}
		defer tx.Rollback()

		snapshot := FixtureWorldSnapshot()
		mock.ExpectQuery(regexp.QuoteMeta(worldSnapshotExistsQuery)).
			WithArgs(snapshot.GameID).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectExec(regexp.QuoteMeta(worldSnapshotInsertQuery)).
			WillReturnError(errors.New("insert snapshot failed"))
		mock.ExpectRollback()

		err = CreateInTx(tx, snapshot)
		if err == nil || !strings.Contains(err.Error(), "storing the world snapshot") {
			t.Fatalf("CreateInTx() error = %v", err)
		}
	})
}

func TestWorldStoreUpdateTurnState(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		store := NewStore(sqlxDB)
		turn := FixtureMutableTurnState()

		activePlayerOrderJSON := MustCanonicalString(t, turn.ActivePlayerOrder)
		ratingsJSON := MustCanonicalString(t, turn.Ratings)
		pollutionJSON := MustCanonicalString(t, turn.Pollution)

		mock.ExpectExec("UPDATE world_turn_states").
			WithArgs(
				turn.TurnNumber,
				turn.CurrentPlayerIndex,
				activePlayerOrderJSON,
				turn.RecalculationCount,
				"99",
				ratingsJSON,
				pollutionJSON,
				turn.LastRecalculatedAt,
				int64(44),
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := store.UpdateTurnState(44, turn); err != nil {
			t.Fatalf("UpdateTurnState() error = %v", err)
		}
	})

	t.Run("update error", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		store := NewStore(sqlxDB)
		turn := FixtureMutableTurnState()

		mock.ExpectExec("UPDATE world_turn_states").WillReturnError(errors.New("update failed"))

		err := store.UpdateTurnState(44, turn)
		if err == nil || !strings.Contains(err.Error(), "an error occurred while updating world turn state for game 44") {
			t.Fatalf("error = %v want wrapped update error", err)
		}
	})

	t.Run("rows affected error", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		store := NewStore(sqlxDB)
		turn := FixtureMutableTurnState()

		mock.ExpectExec("UPDATE world_turn_states").WillReturnResult(rowsAffectedErrorResult{})

		err := store.UpdateTurnState(44, turn)
		if err == nil || !strings.Contains(err.Error(), "confirming world turn state update") {
			t.Fatalf("error = %v want rows affected wrapper", err)
		}
	})

	t.Run("no rows updated", func(t *testing.T) {
		sqlxDB, mock := newWorldSQLMock(t)
		store := NewStore(sqlxDB)
		turn := FixtureMutableTurnState()

		mock.ExpectExec("UPDATE world_turn_states").WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.UpdateTurnState(44, turn)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("error = %v want sql.ErrNoRows", err)
		}
	})
}

func TestNormalizeWorldSnapshotForStorageAndDecodeHelpers(t *testing.T) {
	t.Run("uses config generator version when trimmed snapshot version is blank", func(t *testing.T) {
		snapshot := FixtureWorldSnapshot()
		snapshot.GameStatus = " started "
		snapshot.GeneratorVersion = " "

		normalized, err := normalizeWorldSnapshotForStorage(snapshot)
		if err != nil {
			t.Fatalf("normalizeWorldSnapshotForStorage() error = %v", err)
		}
		if normalized.GameStatus != "started" {
			t.Fatalf("GameStatus = %q want %q", normalized.GameStatus, "started")
		}
		if normalized.GeneratorVersion != snapshot.Config.GeneratorVersion {
			t.Fatalf("GeneratorVersion = %q want %q", normalized.GeneratorVersion, snapshot.Config.GeneratorVersion)
		}
	})

	t.Run("validates required world snapshot fields", func(t *testing.T) {
		badGameID := FixtureWorldSnapshot()
		badGameID.GameID = 0
		if _, err := normalizeWorldSnapshotForStorage(badGameID); err == nil || err.Error() != "game_id must be positive" {
			t.Fatalf("bad game id error = %v", err)
		}

		missingStatus := FixtureWorldSnapshot()
		missingStatus.GameStatus = "   "
		if _, err := normalizeWorldSnapshotForStorage(missingStatus); err == nil || err.Error() != "game_status must be specified" {
			t.Fatalf("missing status error = %v", err)
		}

		missingVersion := FixtureWorldSnapshot()
		missingVersion.GeneratorVersion = ""
		missingVersion.Config.GeneratorVersion = ""
		if _, err := normalizeWorldSnapshotForStorage(missingVersion); err == nil || err.Error() != "generator_version must be specified" {
			t.Fatalf("missing generator version error = %v", err)
		}
	})

	t.Run("decode json helpers handle blank and invalid input", func(t *testing.T) {
		values, err := decodeJSONSlice[int64]("   ")
		if err != nil {
			t.Fatalf("decodeJSONSlice(blank) error = %v", err)
		}
		if values != nil {
			t.Fatalf("decodeJSONSlice(blank) = %#v want nil", values)
		}

		if _, err := decodeJSONSlice[int64]("not-json"); err == nil {
			t.Fatal("expected decodeJSONSlice invalid JSON error")
		}

		decoded := struct {
			Value string `json:"value"`
		}{}
		if err := decodeJSONObject("   ", &decoded); err != nil {
			t.Fatalf("decodeJSONObject(blank) error = %v", err)
		}
		if decoded.Value != "" {
			t.Fatalf("decoded blank object = %#v want zero value", decoded)
		}
		if err := decodeJSONObject("{", &decoded); err == nil {
			t.Fatal("expected decodeJSONObject invalid JSON error")
		}
	})
}

func TestWorldStoreRowLoadersWrapErrors(t *testing.T) {
	if _, err := getWorldSnapshotRow(stubWorldGetter{err: sql.ErrNoRows}, 44); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("getWorldSnapshotRow(sql.ErrNoRows) error = %v", err)
	}
	if _, err := getWorldTurnStateRow(stubWorldGetter{err: sql.ErrNoRows}, 44); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("getWorldTurnStateRow(sql.ErrNoRows) error = %v", err)
	}

	_, err := getWorldSnapshotRow(stubWorldGetter{err: errors.New("snapshot down")}, 44)
	if err == nil || !strings.Contains(err.Error(), "loading the world snapshot for game 44") {
		t.Fatalf("snapshot loader error = %v", err)
	}

	_, err = getWorldTurnStateRow(stubWorldGetter{err: errors.New("turn down")}, 44)
	if err == nil || !strings.Contains(err.Error(), "loading the world turn state for game 44") {
		t.Fatalf("turn loader error = %v", err)
	}
}

func TestAssembleStoredWorld_ErrorPaths(t *testing.T) {
	snapshotRow, turnRow := fixtureStoredWorldRows(t)

	t.Run("invalid player order", func(t *testing.T) {
		badSnapshot := snapshotRow
		badSnapshot.PlayerOrderJSON = "not-json"

		_, err := assembleStoredWorld(badSnapshot, turnRow)
		if err == nil || !strings.Contains(err.Error(), "decoding world player order") {
			t.Fatalf("error = %v want player order decode wrapper", err)
		}
	})

	t.Run("invalid active player order", func(t *testing.T) {
		badTurn := turnRow
		badTurn.ActivePlayerOrderJSON = "not-json"

		_, err := assembleStoredWorld(snapshotRow, badTurn)
		if err == nil || !strings.Contains(err.Error(), "decoding active player order") {
			t.Fatalf("error = %v want active player order decode wrapper", err)
		}
	})

	t.Run("invalid ratings", func(t *testing.T) {
		badTurn := turnRow
		badTurn.RatingsJSON = "{"

		_, err := assembleStoredWorld(snapshotRow, badTurn)
		if err == nil || !strings.Contains(err.Error(), "decoding world ratings") {
			t.Fatalf("error = %v want ratings decode wrapper", err)
		}
	})

	t.Run("invalid pollution", func(t *testing.T) {
		badTurn := turnRow
		badTurn.PollutionJSON = "{"

		_, err := assembleStoredWorld(snapshotRow, badTurn)
		if err == nil || !strings.Contains(err.Error(), "decoding world pollution") {
			t.Fatalf("error = %v want pollution decode wrapper", err)
		}
	})

	t.Run("invalid rng state", func(t *testing.T) {
		badTurn := turnRow
		badTurn.RNGState = "not-a-number"

		_, err := assembleStoredWorld(snapshotRow, badTurn)
		if err == nil || !strings.Contains(err.Error(), "decoding RNG state") {
			t.Fatalf("error = %v want rng decode wrapper", err)
		}
	})

	t.Run("invalid spawns", func(t *testing.T) {
		badSnapshot := snapshotRow
		badSnapshot.SpawnsJSON = "not-json"

		_, err := assembleStoredWorld(badSnapshot, turnRow)
		if err == nil || !strings.Contains(err.Error(), "decoding world spawns") {
			t.Fatalf("error = %v want spawns decode wrapper", err)
		}
	})

	t.Run("invalid tiles", func(t *testing.T) {
		badSnapshot := snapshotRow
		badSnapshot.TilesJSON = "not-json"

		_, err := assembleStoredWorld(badSnapshot, turnRow)
		if err == nil || !strings.Contains(err.Error(), "decoding world tiles") {
			t.Fatalf("error = %v want tiles decode wrapper", err)
		}
	})
}

func fixtureStoredWorldRows(t *testing.T) (worldSnapshotRow, worldTurnStateRow) {
	t.Helper()
	snapshot := FixtureWorldSnapshot()
	turn := snapshot.Turn
	playerOrderJSON := MustCanonicalString(t, snapshot.PlayerOrder)
	spawnsJSON := MustCanonicalString(t, snapshot.Spawns)
	tilesJSON := MustCanonicalString(t, snapshot.Tiles)
	activePlayerOrderJSON := MustCanonicalString(t, turn.ActivePlayerOrder)
	ratingsJSON := MustCanonicalString(t, turn.Ratings)
	pollutionJSON := MustCanonicalString(t, turn.Pollution)

	return worldSnapshotRow{
			GameID:           snapshot.GameID,
			GameStatus:       snapshot.GameStatus,
			GeneratorVersion: snapshot.GeneratorVersion,
			Style:            snapshot.Config.Style,
			RulesJSON:        snapshot.Config.RulesJSON,
			ExclusionsJSON:   snapshot.Config.ExclusionsJSON,
			StartingSettlers: snapshot.Config.StartingSettlers,
			StartingGold:     snapshot.Config.StartingGold,
			Seed:             snapshot.Config.Seed,
			Wetness:          snapshot.Config.Wetness,
			MapSize:          snapshot.Config.MapSize,
			Climate:          snapshot.Config.Climate,
			PlayerOrderJSON:  playerOrderJSON,
			SpawnsJSON:       spawnsJSON,
			TilesJSON:        tilesJSON,
		}, worldTurnStateRow{
			GameID:                snapshot.GameID,
			TurnNumber:            turn.TurnNumber,
			CurrentPlayerIndex:    turn.CurrentPlayerIndex,
			ActivePlayerOrderJSON: activePlayerOrderJSON,
			RecalculationCount:    turn.RecalculationCount,
			RNGState:              "99",
			RatingsJSON:           ratingsJSON,
			PollutionJSON:         pollutionJSON,
			LastRecalculatedAt:    turn.LastRecalculatedAt,
		}
}

func sqlmockDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return sqlx.NewDb(db, "sqlmock")
}
