package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type queueReceiptState struct{ state engine.State }

func (m *queueReceiptState) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		return false
	}
	if json.Unmarshal(raw, &m.state) != nil || len(m.state.Cities) != 1 {
		return false
	}
	city := m.state.Cities[0]
	return city.StoredProduction == 40 && len(city.Queue) == 2 && city.Queue[0].EntryID == "home:queue:1" && city.Queue[1].EntryID == "home:queue:3"
}

func TestClientQueueEditPersistsStableIdentityAndReplaysExactReceipt(t *testing.T) {
	svc, mock := newMockService(t)
	state := testState()
	state.Tiles = []engine.Tile{{ID: "tile", Terrain: engine.TerrainPlains, Open: true}}
	state.Cities = []engine.City{{ID: "home", OwnerID: "7", TileID: "tile", StoredProduction: 40, Queue: []engine.ProductionItem{{ID: "warrior", Kind: engine.ProductionUnit, Cost: 100}, {ID: "warrior", Kind: engine.ProductionUnit, Cost: 100}, {ID: "warrior", Kind: engine.ProductionUnit, Cost: 100}}}}
	req := request()
	req.Type = engine.CommandEditProduction
	req.Payload = []byte(`{"city_id":"home","entries":[{"entry_id":"home:queue:1"},{"entry_id":"home:queue:3"}]}`)
	persisted := &queueReceiptState{}
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WithArgs(persisted, int64(1), int64(44), int64(0)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	result, err := svc.ExecuteCommand(context.Background(), testSession(), req)
	if err != nil {
		t.Fatal(err)
	}
	var projection visibility.Projection
	if err := json.Unmarshal(result.ViewerProjection, &projection); err != nil {
		t.Fatal(err)
	}
	if len(projection.Cities) != 1 || projection.Cities[0].Queue[1].EntryID != "home:queue:3" || projection.Cities[0].StoredProduction != 40 {
		t.Fatalf("receipt and persisted queue disagree: %s", result.ViewerProjection)
	}
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnRows(sqlmock.NewRows([]string{"game_id", "actor_player_id", "idempotency_key", "payload_sha256", "result_json", "resulting_revision", "resulting_event_sequence"}).AddRow(44, 7, "key-1", payloadHash(req), string(result.Receipt), 1, 1))
	mock.ExpectCommit()
	replay, err := svc.ExecuteCommand(context.Background(), testSession(), req)
	if err != nil {
		t.Fatal(err)
	}
	if string(replay.Receipt) != string(result.Receipt) {
		t.Fatal("queue replay regenerated a different receipt")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
