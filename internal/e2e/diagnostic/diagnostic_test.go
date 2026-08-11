package diagnostic

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestReadReturnsDetachedCanonicalSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	sqlDB := sqlx.NewDb(db, "sqlmock")
	state := diagnosticState()
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		mock.ExpectQuery(regexp.QuoteMeta(stateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"canonical_state_json"}).AddRow(string(raw)))
		mock.ExpectQuery(regexp.QuoteMeta(playersQuery)).WithArgs(int64(44)).WillReturnRows(playerRows())
	}

	first, err := Read(context.Background(), sqlDB, 44)
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, first)
	first.Players[0].ID = "mutated-player"
	first.Countries[0].RecordID = "mutated-country"
	first.Controllers[0].Kind = "mutated-controller"
	first.Cities[0].WorkedTileIDs[0] = "mutated-worked-tile"
	first.Cities[0].BuildingIDs[0] = "mutated-building"
	first.Units[0].CargoDomains[0] = "mutated-domain"
	first.Phase.ActivePlayerOrder[0] = "mutated-order"

	second, err := Read(context.Background(), sqlDB, 44)
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, second)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReadRejectsMismatchedPersistedPlayerMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	sqlDB := sqlx.NewDb(db, "sqlmock")
	raw, err := json.Marshal(diagnosticState())
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(stateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"canonical_state_json"}).AddRow(string(raw)))
	mock.ExpectQuery(regexp.QuoteMeta(playersQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "game_id", "account_id", "joined_at", "country", "ping_ms", "controller_kind"}))
	if _, err := Read(context.Background(), sqlDB, 44); err == nil {
		t.Fatal("diagnostic read accepted missing persisted player metadata")
	}
}

func diagnosticState() engine.State {
	return engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		Revision:      7, EventSequence: 9,
		Tiles:   []engine.Tile{{ID: "tile-a", X: 1, Y: 2, Terrain: engine.TerrainPlains, Open: true}},
		Units:   []engine.Unit{{ID: "unit-a", OwnerID: "account-7", TileID: "tile-a", Domain: engine.DomainLand, Movement: 2, CargoDomains: []string{engine.DomainLand}, Fortified: true}},
		Cities:  []engine.City{{ID: "city-a", OwnerID: "account-7", TileID: "tile-a", WorkedTileIDs: []string{"tile-a"}, Buildings: []engine.Building{{ID: "granary"}}}},
		Nations: []engine.Nation{{ID: "nation-a"}},
		Players: []engine.Player{{ID: "account-7", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}},
		Phase:   engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 3, ActivePlayerOrder: []string{"account-7"}, BaseRevision: 6},
		RNG:     []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
	}
}

func playerRows() *sqlmock.Rows {
	now := time.Unix(10, 0).UTC()
	return sqlmock.NewRows([]string{"id", "created_at", "updated_at", "game_id", "account_id", "joined_at", "country", "ping_ms", "controller_kind"}).
		AddRow(int64(70), now, now, int64(44), int64(7), now, "synthetic-country-7", int64(0), engine.ControllerHuman)
}

func assertSnapshot(t *testing.T, snapshot Snapshot) {
	t.Helper()
	if snapshot.Game.ID != 44 || snapshot.Game.StateSchema != engine.StateSchemaV1 || snapshot.Revision != 7 || snapshot.EventCursor != 9 {
		t.Fatalf("unexpected game snapshot: %#v", snapshot)
	}
	if len(snapshot.Players) != 1 || snapshot.Players[0].ID != "account-7" || snapshot.Players[0].NationID != "nation-a" {
		t.Fatalf("unexpected players: %#v", snapshot.Players)
	}
	if len(snapshot.Countries) != 1 || snapshot.Countries[0].RecordID != "synthetic-country-7" {
		t.Fatalf("unexpected countries: %#v", snapshot.Countries)
	}
	if len(snapshot.Controllers) != 1 || snapshot.Controllers[0].Kind != engine.ControllerHuman {
		t.Fatalf("unexpected controllers: %#v", snapshot.Controllers)
	}
	if len(snapshot.Cities) != 1 || snapshot.Cities[0].WorkedTileIDs[0] != "tile-a" || snapshot.Cities[0].BuildingIDs[0] != "granary" {
		t.Fatalf("unexpected cities: %#v", snapshot.Cities)
	}
	if len(snapshot.Units) != 1 || snapshot.Units[0].CargoDomains[0] != engine.DomainLand || !snapshot.Units[0].Fortified {
		t.Fatalf("unexpected units: %#v", snapshot.Units)
	}
	if snapshot.Phase.Turn != 3 || snapshot.Phase.ActivePlayerOrder[0] != "account-7" || snapshot.Terminal.Ended {
		t.Fatalf("unexpected phase/terminal: phase=%#v terminal=%#v", snapshot.Phase, snapshot.Terminal)
	}
}
