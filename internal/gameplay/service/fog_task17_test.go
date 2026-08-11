package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type fogStateMatcher struct{ saw bool }

func (m *fogStateMatcher) Match(value driver.Value) bool {
	raw, ok := value.([]byte)
	if !ok {
		return false
	}
	var state engine.State
	if err := json.Unmarshal(raw, &state); err != nil {
		return false
	}
	for _, observation := range state.Fog {
		if observation.PlayerID == "account-9" && observation.TileID == "t3" && observation.Turn == 1 {
			m.saw = true
		}
	}
	return m.saw
}

func TestTask17AcceptedMovePersistsRefreshedFogAndProjectsOnlyVisibleTarget(t *testing.T) {
	svc, mock := newMockService(t)
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state := engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON},
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerOrder: []string{"account-9", "account-10"}},
		Players:       []engine.Player{{ID: "account-9", NationID: "a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "account-10", NationID: "b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}},
		RNG:           []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
		Tiles:         []engine.Tile{{ID: "t0", X: 0, Y: 0, Terrain: engine.TerrainPlains}, {ID: "t1", X: 1, Y: 0, Terrain: engine.TerrainPlains}, {ID: "t2", X: 2, Y: 0, Terrain: engine.TerrainPlains}, {ID: "t3", X: 3, Y: 0, Terrain: engine.TerrainPlains}},
		Units:         []engine.Unit{{ID: "scout", OwnerID: "account-9", TileID: "t0", Domain: engine.DomainLand, Movement: 2}, {ID: "hidden", OwnerID: "account-10", TileID: "t3", Domain: engine.DomainLand, Movement: 2}},
	}
	stateJSON, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	request := CommandRequest{GameID: 44, ExpectedRevision: 0, IdempotencyKey: "move-key", CommandID: "move-command", Type: engine.CommandMoveUnit, Payload: json.RawMessage(`{"unit_id":"scout","path":["t1"]}`)}
	matcher := &fogStateMatcher{}
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "move-key").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version"}).AddRow(44, 1, 0, 0, string(stateJSON), string(ruleset.CanonicalJSON), ruleset.ContentHash, ruleset.CompatibilityVersion))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WithArgs(matcher, int64(1), int64(44), int64(0)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := svc.ExecuteCommand(context.Background(), testSession(), request)
	if err != nil {
		t.Fatal(err)
	}
	var projection visibility.Projection
	if err := json.Unmarshal(result.ViewerProjection, &projection); err != nil {
		t.Fatal(err)
	}
	visible := false
	for _, unit := range projection.Units {
		if unit.ID == "hidden" {
			visible = true
		}
	}
	if !visible || !matcher.saw {
		t.Fatalf("accepted move did not durably reveal the legally visible target: projection=%s persisted=%v", result.ViewerProjection, matcher.saw)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAcceptedAIMoveWritesOnlyItsOwnExplorationStep(t *testing.T) {
	svc, mock := newMockService(t)
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state := engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON},
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerOrder: []string{"account-7", "account-8"}},
		Players:       []engine.Player{{ID: "account-7", NationID: "a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}, {ID: "account-8", NationID: "b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}},
		RNG:           []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
		Tiles:         []engine.Tile{{ID: "t0", X: 0, Y: 0, Terrain: engine.TerrainPlains}, {ID: "t1", X: 1, Y: 0, Terrain: engine.TerrainPlains}},
		Units:         []engine.Unit{{ID: "scout", OwnerID: "account-7", TileID: "t0", Domain: engine.DomainLand, Movement: 1}},
	}
	stateJSON, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "started", nil))
	mock.ExpectQuery(regexp.QuoteMeta(lockAIPlayerQuery)).WithArgs(int64(44), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "controller_kind"}).AddRow(7, "ai"))
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "ai-move").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version"}).AddRow(44, 1, 0, 0, string(stateJSON), string(ruleset.CanonicalJSON), ruleset.ContentHash, ruleset.CompatibilityVersion))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO gameplay_ai_exploration_history (
	game_id, actor_player_id, sequence, command_type, destination_tile_id
) VALUES ($1, $2, $3, $4, $5)`)).WithArgs(int64(44), int64(7), int64(1), engine.CommandMoveUnit, "t1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := NewAIGateway(svc).Submit(context.Background(), ai.Submission{GameID: 44, ExpectedRevision: 0, IdempotencyKey: "ai-move", CommandID: "ai-move-command", ActorPlayerID: "account-7", Type: engine.CommandMoveUnit, Payload: json.RawMessage(`{"unit_id":"scout","path":["t1"]}`)})
	if err != nil || result.Revision != 1 {
		t.Fatalf("accepted AI move result=%+v err=%v", result, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
