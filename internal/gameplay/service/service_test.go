package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
	_ "github.com/lib/pq"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fixedIDs struct{ next int }

func (s *fixedIDs) Next() string { s.next++; return "event-" + string(rune('0'+s.next)) }

type fixedRNG struct{}

func (fixedRNG) Next(domain string, state uint64) (uint64, uint64) { return state + 1, state + 1 }

func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return New(store.New(sqlx.NewDb(db, "sqlmock")), engine.Sources{Clock: fixedClock{time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)}, IDs: &fixedIDs{}, RNG: fixedRNG{}}), mock
}

func testSession() *auth.Session {
	return &auth.Session{AccountID: 9, ExpiresAt: time.Now().Add(time.Hour)}
}

func testState() engine.State {
	return engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		Players:       []engine.Player{{ID: "7", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}},
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerOrder: []string{"7"}},
		RNG:           []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
	}
}

func aggregateRows(t *testing.T, state engine.State) *sqlmock.Rows {
	t.Helper()
	body, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	return sqlmock.NewRows([]string{"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version"}).
		AddRow(44, 1, state.Revision, state.EventSequence, string(body), `{"version":"ctp2-supported-v2"}`, state.Rules.ContentHash, state.Rules.Version)
}

func expectIdentity(mock sqlmock.Sqlmock, controller string) {
	mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "started", 9))
	mock.ExpectQuery(regexp.QuoteMeta(lockPlayerQuery)).WithArgs(int64(44), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"id", "controller_kind"}).AddRow(7, controller))
}

func request() CommandRequest {
	return CommandRequest{GameID: 44, ExpectedRevision: 0, IdempotencyKey: "key-1", CommandID: "command-1", Type: engine.CommandNoop, Payload: json.RawMessage(`{}`)}
}

func TestExecuteCommandRejectsMalformedAndUnauthenticatedBeforeTransaction(t *testing.T) {
	svc, mock := newMockService(t)
	bad := request()
	bad.IdempotencyKey = ""
	if _, err := svc.ExecuteCommand(context.Background(), testSession(), bad); StatusCode(err) != 400 {
		t.Fatalf("malformed status = %d", StatusCode(err))
	}
	if _, err := svc.ExecuteCommand(context.Background(), nil, request()); StatusCode(err) != 401 {
		t.Fatalf("unauthenticated status = %d", StatusCode(err))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteCommandRejectsMembershipAIAndStatusWithoutMutation(t *testing.T) {
	cases := []struct {
		name  string
		setup func(sqlmock.Sqlmock)
		want  int
	}{
		{"nonmember", func(mock sqlmock.Sqlmock) {
			mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "started", 9))
			mock.ExpectQuery(regexp.QuoteMeta(lockPlayerQuery)).WithArgs(int64(44), int64(9)).WillReturnError(sql.ErrNoRows)
		}, 403},
		{"ai", func(mock sqlmock.Sqlmock) { expectIdentity(mock, "ai") }, 403},
		{"setup", func(mock sqlmock.Sqlmock) {
			mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "setup", 9))
		}, 409},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc, mock := newMockService(t)
			mock.ExpectBegin()
			tt.setup(mock)
			mock.ExpectRollback()
			if _, err := svc.ExecuteCommand(context.Background(), testSession(), request()); StatusCode(err) != tt.want {
				t.Fatalf("status = %d want %d (%v)", StatusCode(err), tt.want, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExecuteCommandRejectsPostCombatCityMoveWhenGameEnded(t *testing.T) {
	svc, mock := newMockService(t)
	move := CommandRequest{
		GameID:           44,
		ExpectedRevision: 0,
		IdempotencyKey:   "post-combat-capture-move",
		CommandID:        "post-combat-capture-move",
		Type:             engine.CommandMoveUnit,
		Payload:          json.RawMessage(`{"unit_id":"attacker-melee","path":["city-b"]}`),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "ended", 9))
	mock.ExpectRollback()

	_, err := svc.ExecuteCommand(context.Background(), testSession(), move)
	var domain *Error
	if !errors.As(err, &domain) || domain.Status != 409 || domain.Message != "game is not accepting gameplay commands" {
		t.Fatalf("post-combat city move error = %#v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteCommandCommitsOnceAndReplaysStoredBytes(t *testing.T) {
	svc, mock := newMockService(t)
	var timings []CommandTiming
	svc.ObserveCommand = func(metric CommandTiming) { timings = append(timings, metric) }
	clock := time.Unix(0, 0)
	svc.now = func() time.Time { clock = clock.Add(time.Millisecond); return clock }
	state := testState()
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	result, err := svc.ExecuteCommand(context.Background(), testSession(), request())
	if err != nil {
		t.Fatal(err)
	}
	if result.Revision != 1 || result.EventCursor != 1 || len(result.Receipt) == 0 {
		t.Fatalf("unexpected result: %+v", result)
	}

	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnRows(sqlmock.NewRows([]string{"game_id", "actor_player_id", "idempotency_key", "payload_sha256", "result_json", "resulting_revision", "resulting_event_sequence"}).AddRow(44, 7, "key-1", payloadHash(request()), string(result.Receipt), 1, 1))
	mock.ExpectCommit()
	replayed, err := svc.ExecuteCommand(context.Background(), testSession(), request())
	if err != nil {
		t.Fatal(err)
	}
	if string(replayed.Receipt) != string(result.Receipt) {
		t.Fatalf("replay differs: %s != %s", replayed.Receipt, result.Receipt)
	}
	if len(timings) != 2 || timings[0].Engine <= 0 || timings[0].Visibility <= 0 || timings[0].Persistence <= 0 || timings[1].Engine != 0 || timings[1].Outcome != "replayed" {
		t.Fatalf("ambiguous command timings: %+v", timings)
	}
	for _, metric := range timings {
		if metric.Total != metric.Admission+metric.Engine+metric.Visibility+metric.Persistence || metric.CommandSHA256 == "" {
			t.Fatalf("unbounded/inconsistent timings: %+v", metric)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type simultaneousQueueState struct{}

func (simultaneousQueueState) Match(value driver.Value) bool {
	var body []byte
	switch value := value.(type) {
	case []byte:
		body = value
	case string:
		body = []byte(value)
	default:
		return false
	}
	state, err := engine.DecodeCanonicalStateJSON(body)
	if err != nil || len(state.Policy.Queue) != 1 {
		return false
	}
	queued := state.Policy.Queue[0]
	return queued.ActorID == "account-10" && queued.CommandID == "queued-by-10" && queued.IdempotencyKey == "queue-key" && string(queued.Payload) == `{"schema_version":1,"id":"queued-by-10","actor_player_id":"account-10","idempotency_key":"queue-key","type":"noop","payload":{}}`
}

func TestExecuteCommandQueuesEligibleSimultaneousPlannerInOneTransaction(t *testing.T) {
	svc, mock := newMockService(t)
	state := testState()
	state.Phase.Mode = "simultaneous"
	state.Players = []engine.Player{
		{ID: "account-9", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman},
		{ID: "account-10", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman},
	}
	state.Phase.ActivePlayerOrder = []string{"account-9", "account-10"}
	request := request()
	request.CommandID = "queued-by-10"
	request.IdempotencyKey = "queue-key"
	session := &auth.Session{AccountID: 10, ExpiresAt: time.Now().Add(time.Hour)}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "started", 9))
	mock.ExpectQuery(regexp.QuoteMeta(lockPlayerQuery)).WithArgs(int64(44), int64(10)).WillReturnRows(sqlmock.NewRows([]string{"id", "controller_kind"}).AddRow(8, "human"))
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(8), "queue-key").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WithArgs(simultaneousQueueState{}, int64(1), int64(44), int64(0)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := svc.ExecuteCommand(context.Background(), session, request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Revision != 1 || result.EventCursor != 1 {
		t.Fatalf("queued result = %#v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteCommandPersistsOwnerTerminalMetadata(t *testing.T) {
	svc, mock := newMockService(t)
	end := request()
	end.Type = engine.CommandEndGame
	state := testState()
	state.Players = append(state.Players, engine.Player{ID: "other", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman})
	state.Phase.ActivePlayerOrder = []string{"other"} // owner end is not constrained to the active turn.
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(endGameQuery)).WithArgs(int64(7), "owner_ended", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	result, err := svc.ExecuteCommand(context.Background(), testSession(), end)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnRows(sqlmock.NewRows([]string{"game_id", "actor_player_id", "idempotency_key", "payload_sha256", "result_json", "resulting_revision", "resulting_event_sequence"}).AddRow(44, 7, "key-1", payloadHash(end), string(result.Receipt), 1, 1))
	mock.ExpectCommit()
	if replayed, err := svc.ExecuteCommand(context.Background(), testSession(), end); err != nil || string(replayed.Receipt) != string(result.Receipt) {
		t.Fatalf("terminal receipt replay = %#v, %v", replayed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteCommandPersistsCanonicalAccountTerminalWinner(t *testing.T) {
	svc, mock := newMockService(t)
	end := request()
	end.Type = engine.CommandEndGame
	state := testState()
	state.Players[0].ID = "account-9"
	state.Phase.ActivePlayerOrder = []string{"account-9"}

	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(terminalWinnerQuery)).WithArgs(int64(44), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(7)))
	mock.ExpectExec(regexp.QuoteMeta(endGameQuery)).WithArgs(int64(7), "owner_ended", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if _, err := svc.ExecuteCommand(context.Background(), testSession(), end); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteAICommandUsesTheOrdinaryReceiptGateway(t *testing.T) {
	svc, mock := newMockService(t)
	state := testState()
	state.Players[0].Controller = engine.ControllerAI
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "started", nil))
	mock.ExpectQuery(regexp.QuoteMeta(lockAIPlayerQuery)).WithArgs(int64(44), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "controller_kind"}).AddRow(7, "ai"))
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	result, err := NewAIGateway(svc).Submit(context.Background(), ai.Submission{GameID: 44, ExpectedRevision: 0, IdempotencyKey: "key-1", CommandID: "command-1", ActorPlayerID: "account-7", Type: engine.CommandNoop, Payload: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Revision != 1 {
		t.Fatalf("AI command did not use the standard receipt path: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteAICommandRejectsMalformedCanonicalActorBeforeTransaction(t *testing.T) {
	svc, mock := newMockService(t)
	_, err := NewAIGateway(svc).Submit(context.Background(), ai.Submission{GameID: 44, ActorPlayerID: "not-an-account", Type: engine.CommandNoop, Payload: json.RawMessage(`{}`)})
	if err == nil || StatusCode(err) != 400 {
		t.Fatalf("malformed AI actor error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAIGatewayProjectionExcludesHiddenCanonicalCanaryBeforePlanning(t *testing.T) {
	svc, mock := newMockService(t)
	state := testState()
	state.Players = []engine.Player{
		{ID: "account-7", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
		{ID: "account-8", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman},
	}
	state.Phase.ActivePlayerOrder = []string{"account-7", "account-8"}
	state.Nations = []engine.Nation{{ID: "nation-a"}, {ID: "nation-b"}}
	state.Tiles = []engine.Tile{
		{ID: "origin", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true},
		{ID: "HIDDEN_CANARY_TILE", X: 20, Y: 0, Terrain: engine.TerrainForest, Open: true},
	}
	state.Units = []engine.Unit{
		{ID: "ai-scout", OwnerID: "account-7", TileID: "origin", Domain: engine.DomainLand, Movement: 2},
		{ID: "HIDDEN_CANARY_UNIT", OwnerID: "account-8", TileID: "HIDDEN_CANARY_TILE", Domain: engine.DomainLand, Movement: 2},
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n\tstate.game_id,\n\tstate.state_schema_version,\n\tstate.revision,\n\tstate.last_event_sequence,\n\tstate.canonical_state_json::text AS canonical_state_json,\n\tsnapshot.ruleset_canonical_json::text AS ruleset_canonical_json,\n\tsnapshot.ruleset_content_hash,\n\tsnapshot.ruleset_version\nFROM world_turn_states AS state\nJOIN world_snapshots AS snapshot ON snapshot.game_id = state.game_id\nWHERE state.game_id = $1")).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))

	input, err := NewAIGateway(svc).Projection(context.Background(), 44, "account-7")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(input), "HIDDEN_CANARY") {
		t.Fatalf("AIGateway leaked canonical canary into planning input: %s", input)
	}
	candidates, err := ai.Plan(input, "account-7")
	if err != nil || len(candidates) == 0 {
		t.Fatalf("production AI could not plan from gateway projection: candidates=%#v err=%v", candidates, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAIGatewayHistoryReadsOnlyTheGameScopedAIPlayer(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM game_players
WHERE game_id = $1 AND account_id = $2`)).WithArgs(int64(44), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(19))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT
  game_id, actor_player_id, sequence, command_type, destination_tile_id
FROM gameplay_ai_exploration_history
WHERE game_id = $1 AND actor_player_id = $2
ORDER BY sequence ASC`)).WithArgs(int64(44), int64(19)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "actor_player_id", "sequence", "command_type", "destination_tile_id"}).AddRow(44, 19, 5, engine.CommandMoveUnit, "visible-step"))

	history, err := NewAIGateway(svc).ExplorationHistory(context.Background(), 44, "account-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].DestinationTileID != "visible-step" || history[0].Sequence != 5 {
		t.Fatalf("game-scoped AI history = %#v", history)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTakeoverPersistsControllerWithItsOrdinaryCommand(t *testing.T) {
	svc, mock := newMockService(t)
	state := testState()
	state.Players[0].TakeoverPending = true
	request := request()
	request.Type = engine.CommandTakeoverAI
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateControllerQuery)).WithArgs("ai", int64(44), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if _, err := svc.ExecuteCommand(context.Background(), testSession(), request); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteCommandRejectsReceiptPayloadConflictStaleAndRollsBack(t *testing.T) {
	cases := []struct {
		name  string
		setup func(sqlmock.Sqlmock)
		want  int
	}{
		{"payload conflict", func(mock sqlmock.Sqlmock) {
			expectIdentity(mock, "human")
			mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "actor_player_id", "idempotency_key", "payload_sha256", "result_json", "resulting_revision", "resulting_event_sequence"}).AddRow(44, 7, "key-1", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", `{}`, 1, 1))
		}, 409},
		{"stale", func(mock sqlmock.Sqlmock) {
			expectIdentity(mock, "human")
			mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WillReturnError(sql.ErrNoRows)
			state := testState()
			state.Revision = 1
			mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WillReturnRows(aggregateRows(t, state))
		}, 409},
		{"write failure", func(mock sqlmock.Sqlmock) {
			expectIdentity(mock, "human")
			mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WillReturnError(sql.ErrNoRows)
			mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WillReturnRows(aggregateRows(t, testState()))
			mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnError(errors.New("write failed"))
		}, 500},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc, mock := newMockService(t)
			mock.ExpectBegin()
			tt.setup(mock)
			mock.ExpectRollback()
			if _, err := svc.ExecuteCommand(context.Background(), testSession(), request()); StatusCode(err) != tt.want {
				t.Fatalf("status = %d want %d (%v)", StatusCode(err), tt.want, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExecuteCommandRejectsOwnerTerminalAndIllegalPhase(t *testing.T) {
	cases := []struct {
		name    string
		request CommandRequest
		state   engine.State
		owner   int64
		want    int
	}{
		{"wrong owner", func() CommandRequest { r := request(); r.Type = "end-game"; return r }(), testState(), 10, 403},
		{"terminal", request(), func() engine.State {
			s := testState()
			s.Terminal = engine.TerminalState{Ended: true, Reason: "test", WinningPlayerID: "7"}
			return s
		}(), 9, 409},
		{"illegal phase", request(), func() engine.State {
			s := testState()
			s.Players = append(s.Players, engine.Player{ID: "other", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman})
			s.Phase.ActivePlayerOrder = []string{"other"}
			return s
		}(), 9, 422},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc, mock := newMockService(t)
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(lockGameQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"id", "status", "owner_account_id"}).AddRow(44, "started", tt.owner))
			expectPlayer := true
			if expectPlayer {
				mock.ExpectQuery(regexp.QuoteMeta(lockPlayerQuery)).WithArgs(int64(44), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"id", "controller_kind"}).AddRow(7, "human"))
			}
			if tt.want != 403 {
				mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, tt.state))
			}
			mock.ExpectRollback()
			if _, err := svc.ExecuteCommand(context.Background(), testSession(), tt.request); StatusCode(err) != tt.want {
				t.Fatalf("status = %d want %d (%v)", StatusCode(err), tt.want, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExecuteCommandPostgresLockRace(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("GAMEPLAY_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set GAMEPLAY_POSTGRES_DSN to run the local PostgreSQL lock-race test")
	}
	root, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	schema := "gameplay_service_lock_race"
	setup := []string{
		`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`,
		`CREATE SCHEMA ` + schema,
		`CREATE TABLE ` + schema + `.games (id bigint primary key, status text not null, owner_account_id bigint, updated_at timestamptz not null default now())`,
		`CREATE TABLE ` + schema + `.game_players (id bigint primary key, game_id bigint not null, account_id bigint not null, controller_kind text not null)`,
		`CREATE TABLE ` + schema + `.world_snapshots (game_id bigint primary key, ruleset_canonical_json jsonb, ruleset_content_hash text, ruleset_version text)`,
		`CREATE TABLE ` + schema + `.world_turn_states (game_id bigint primary key, state_schema_version integer not null, revision bigint not null, last_event_sequence bigint not null, canonical_state_json jsonb not null, updated_at timestamptz not null default now())`,
		`CREATE TABLE ` + schema + `.gameplay_command_receipts (game_id bigint, actor_player_id bigint, idempotency_key text, payload_sha256 text, result_json jsonb, resulting_revision bigint, resulting_event_sequence bigint, primary key (game_id, actor_player_id, idempotency_key))`,
		`CREATE TABLE ` + schema + `.gameplay_events (game_id bigint, sequence bigint, revision bigint, ordinal integer, event_id text, event_type text, payload_json jsonb, primary key (game_id, sequence))`,
		`CREATE TABLE ` + schema + `.gameplay_event_outbox (game_id bigint, sequence bigint, primary key (game_id, sequence))`,
	}
	for _, statement := range setup {
		if _, err := root.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	defer root.Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`)
	stateJSON, err := engine.CanonicalStateJSON(testState())
	if err != nil {
		t.Fatal(err)
	}
	fixtures := []string{
		`INSERT INTO ` + schema + `.games VALUES (44, 'started', 9)`,
		`INSERT INTO ` + schema + `.game_players VALUES (7, 44, 9, 'human')`,
		`INSERT INTO ` + schema + `.world_snapshots VALUES (44, '{"version":"ctp2-supported-v2"}', '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef', 'ctp2-supported-v2')`,
	}
	for _, statement := range fixtures {
		if _, err := root.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := root.Exec(`INSERT INTO `+schema+`.world_turn_states (game_id, state_schema_version, revision, last_event_sequence, canonical_state_json) VALUES (44, 1, 0, 0, $1::jsonb)`, string(stateJSON)); err != nil {
		t.Fatal(err)
	}
	open := func() *sqlx.DB {
		db, err := sqlx.Open("postgres", dsn)
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		if _, err := db.Exec(`SET search_path TO ` + schema); err != nil {
			db.Close()
			t.Fatal(err)
		}
		t.Cleanup(func() { db.Close() })
		return db
	}
	barrier := testkit.NewBarrier(2)
	services := []*Service{New(store.New(open()), engine.Sources{Clock: fixedClock{time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)}, IDs: &fixedIDs{}, RNG: fixedRNG{}}), New(store.New(open()), engine.Sources{Clock: fixedClock{time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)}, IDs: &fixedIDs{}, RNG: fixedRNG{}})}
	errs := make(chan error, 2)
	var wait sync.WaitGroup
	for i, svc := range services {
		wait.Add(1)
		go func(i int, svc *Service) {
			defer wait.Done()
			if err := barrier.Wait(context.Background()); err != nil {
				errs <- err
				return
			}
			r := request()
			r.IdempotencyKey = "race-" + string(rune('a'+i))
			_, err := svc.ExecuteCommand(context.Background(), testSession(), r)
			errs <- err
		}(i, svc)
	}
	wait.Wait()
	close(errs)
	var successes, stale int
	for err := range errs {
		if err == nil {
			successes++
		} else if StatusCode(err) == 409 {
			stale++
		} else {
			t.Fatalf("unexpected race error: %v", err)
		}
	}
	if successes != 1 || stale != 1 {
		t.Fatalf("race outcomes success=%d stale=%d", successes, stale)
	}
	var revision, receipts, events, outbox int
	if err := root.Get(&revision, `SELECT revision FROM `+schema+`.world_turn_states WHERE game_id = 44`); err != nil {
		t.Fatal(err)
	}
	if err := root.Get(&receipts, `SELECT count(*) FROM `+schema+`.gameplay_command_receipts`); err != nil {
		t.Fatal(err)
	}
	if err := root.Get(&events, `SELECT count(*) FROM `+schema+`.gameplay_events`); err != nil {
		t.Fatal(err)
	}
	if err := root.Get(&outbox, `SELECT count(*) FROM `+schema+`.gameplay_event_outbox`); err != nil {
		t.Fatal(err)
	}
	if revision != 1 || receipts != 1 || events != 1 || outbox != 1 {
		t.Fatalf("race persisted revision=%d receipts=%d events=%d outbox=%d", revision, receipts, events, outbox)
	}
}
