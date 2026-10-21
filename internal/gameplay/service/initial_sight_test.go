package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type initialSightCapture struct{ state engine.State }

func (capture *initialSightCapture) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	return ok && json.Unmarshal(raw, &capture.state) == nil
}

func TestClientMoveRemembersPreCommandSightForLegacyEmptyFog(t *testing.T) {
	svc, mock := newMockService(t)
	pin, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state := engine.State{SchemaVersion: engine.StateSchemaV1, Rules: engine.RulesPin{Version: pin.CompatibilityVersion, ContentHash: pin.ContentHash, Definitions: pin.CanonicalJSON}, Phase: engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerOrder: []string{"account-9"}}, Players: []engine.Player{{ID: "account-9", NationID: "a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}, RNG: []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}}, Units: []engine.Unit{{ID: "scout", OwnerID: "account-9", TileID: "t0", Domain: engine.DomainLand, Movement: 4}}}
	for x := 0; x < 6; x++ {
		state.Tiles = append(state.Tiles, engine.Tile{ID: fmt.Sprintf("t%d", x), X: x, Terrain: engine.TerrainPlains, Open: true})
	}
	state.Tiles[0].River = true
	state.Tiles[0].GoodID = "observed-origin"
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	capture := &initialSightCapture{}
	req := CommandRequest{GameID: 44, ExpectedRevision: 0, IdempotencyKey: "initial-move", CommandID: "initial-move", Type: engine.CommandMoveUnit, Payload: []byte(`{"unit_id":"scout","path":["t1","t2","t3"]}`)}
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version"}).AddRow(44, 1, 0, 0, string(raw), string(pin.CanonicalJSON), pin.ContentHash, pin.CompatibilityVersion))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WithArgs(capture, int64(1), int64(44), int64(0)).WillReturnResult(sqlmock.NewResult(0, 1))
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
	found := false
	for _, tile := range projection.Tiles {
		if tile.ID == "t0" && tile.Stale && tile.River != nil && *tile.River && tile.GoodID != nil && *tile.GoodID == "observed-origin" {
			found = true
		}
	}
	if !found {
		t.Fatalf("first move discarded previously visible origin facts: %s", result.ViewerProjection)
	}
	found = false
	for _, observation := range capture.state.Fog {
		if observation.PlayerID == "account-9" && observation.TileID == "t0" && observation.River != nil && *observation.River {
			found = true
		}
	}
	if !found {
		t.Fatal("pre-command sight was not persisted")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
