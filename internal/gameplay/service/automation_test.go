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
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type automationStateMatcher struct{ state engine.State }

func (m *automationStateMatcher) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		return false
	}
	var err error
	m.state, err = engine.DecodeCanonicalStateJSON(raw)
	return err == nil
}

func TestAutomationCommandsPersistPrivateResultsAndReplayWithoutReexecution(t *testing.T) {
	for _, kind := range []string{engine.CommandSetMovementOrder, engine.CommandCancelMovementOrder, engine.CommandSetGovernor, engine.CommandMoveUnit} {
		t.Run(kind, func(t *testing.T) {
			svc, mock := newMockService(t)
			state := testState()
			imported, err := rules.ServerOwnedSyntheticRuleset()
			if err != nil {
				t.Fatal(err)
			}
			runtime, err := rules.CompileRuntime(imported.CanonicalJSON, imported.ContentHash)
			if err != nil {
				t.Fatal(err)
			}
			runtime.AI = &rules.RuntimeAI{BuildingBuildLists: []rules.RuntimeAIList{{ID: "BUILDING_BUILD_LIST_SCIENCE", Items: []string{"LIBRARY"}}}}
			runtime.Production = append(runtime.Production, rules.RuntimeProduction{ID: "LIBRARY", Kind: engine.ProductionBuilding, Cost: 20})
			defs, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
			if err != nil {
				t.Fatal(err)
			}
			state.Rules = engine.RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: defs}
			state.Tiles = []engine.Tile{{ID: "t0", Terrain: engine.TerrainPlains, Open: true}, {ID: "t1", X: 1, Terrain: engine.TerrainPlains, Open: true, Ruin: &engine.Ruin{Profile: engine.VillageRuinProfile, TypeRoll: 0, ValueRoll: 5000}}, {ID: "t2", X: 2, Terrain: engine.TerrainPlains, Open: true}, {ID: "t3", X: 3, Terrain: engine.TerrainPlains, Open: true}}
			state.Units = []engine.Unit{{ID: "unit", OwnerID: "7", TileID: "t0", Domain: engine.DomainLand, Movement: 1, MaxMovement: 1}}
			state.Nations = []engine.Nation{{ID: "nation-a"}}
			state.Cities = []engine.City{{ID: "home", OwnerID: "7", TileID: "t0", Population: 2}}
			req := request()
			req.Type = kind
			switch kind {
			case engine.CommandSetMovementOrder:
				req.Payload = []byte(`{"unit_id":"unit","target_x":3,"target_y":0}`)
			case engine.CommandCancelMovementOrder:
				state.MovementOrders = []engine.MovementOrder{{UnitID: "unit", PlayerID: "7", UnitIDs: []string{"unit"}, TileID: "t0", TargetX: 3, CommandID: "old-journey"}}
				req.Payload = []byte(`{"unit_id":"unit"}`)
			case engine.CommandSetGovernor:
				req.Payload = []byte(`{"city_id":"home","enabled":true,"priority":"science"}`)
			case engine.CommandMoveUnit:
				req.Payload = []byte(`{"unit_id":"unit","path":["t1"]}`)
			}
			raw, err := engine.CanonicalStateJSON(state)
			if err != nil {
				t.Fatal(err)
			}
			persisted := &automationStateMatcher{}
			mock.ExpectBegin()
			expectIdentity(mock, "human")
			mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
			mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version"}).AddRow(44, 1, 0, 0, string(raw), string(defs), hash, state.Rules.Version))
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
			if kind == engine.CommandSetMovementOrder && (len(persisted.state.MovementOrders) != 1 || len(projection.MovementOrders) != 1) {
				t.Fatal("journey not persisted/projected")
			}
			if kind == engine.CommandCancelMovementOrder && (len(persisted.state.MovementOrders) != 0 || len(projection.MovementOrders) != 0) {
				t.Fatal("cancellation not persisted/projected")
			}
			if kind == engine.CommandSetGovernor && (len(projection.Cities) != 1 || projection.Cities[0].Governor == nil || len(projection.Cities[0].Queue) != 1) {
				t.Fatal("governor not persisted/projected")
			}
			if kind == engine.CommandMoveUnit || kind == engine.CommandSetMovementOrder {
				if persisted.state.Nations[0].Gold != 225 || projection.Diplomacy.Gold != 225 || len(projection.Notifications) != 1 {
					t.Fatal("authoritative ruin reward missing")
				}
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
				t.Fatal("replay regenerated side effects/receipt")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
