package store

import (
	"errors"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func newSQLMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return sqlx.NewDb(db, "sqlmock"), mock
}

func fixtureAggregate() Aggregate {
	return Aggregate{
		GameID:             44,
		StateSchemaVersion: StateSchemaV1,
		Revision:           0,
		LastEventSequence:  0,
		CanonicalStateJSON: []byte(`{"phase":"setup"}`),
		Ruleset: RulesetPin{
			CanonicalJSON: []byte(`{"version":"ctp2-v1"}`),
			ContentHash:   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			Version:       "ctp2-v1",
		},
	}
}

func TestStorePersistCommandCommitFailureRollsBackStagedWrites(t *testing.T) {
	db, mock := newSQLMock(t)
	store := New(db)
	before := fixtureAggregate()
	after := before
	after.Revision = 1
	after.LastEventSequence = 1
	after.CanonicalStateJSON = []byte(`{"phase":"turn"}`)
	receipt := CommandReceipt{GameID: 44, ActorPlayerID: 7, IdempotencyKey: "key-1", PayloadSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", ResultJSON: []byte(`{"ok":true}`), ResultingRevision: 1, ResultingEventSequence: 1}
	event := Event{GameID: 44, Sequence: 1, Revision: 1, Ordinal: 0, ID: "event-1", Type: "turn_started", PayloadJSON: []byte(`{"turn":1}`)}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).
		WithArgs(after.CanonicalStateJSON, after.LastEventSequence, after.GameID, before.Revision).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).
		WithArgs(receipt.GameID, receipt.ActorPlayerID, receipt.IdempotencyKey, receipt.PayloadSHA256, receipt.ResultJSON, receipt.ResultingRevision, receipt.ResultingEventSequence).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).
		WithArgs(event.GameID, event.Sequence, event.Revision, event.Ordinal, event.ID, event.Type, event.PayloadJSON).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).
		WithArgs(event.GameID, event.Sequence).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := store.PersistCommand(before.Revision, after, receipt, []Event{event})
	if err == nil || !regexp.MustCompile(`finalizing gameplay command.*commit failed`).MatchString(err.Error()) {
		t.Fatalf("PersistCommand() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestStorePersistCommandRollsBackAfterStagedWriteFailure(t *testing.T) {
	db, mock := newSQLMock(t)
	store := New(db)
	before := fixtureAggregate()
	after := before
	after.Revision = 1
	after.LastEventSequence = 1
	receipt := CommandReceipt{GameID: 44, ActorPlayerID: 7, IdempotencyKey: "key-1", PayloadSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", ResultJSON: []byte(`{"ok":true}`), ResultingRevision: 1, ResultingEventSequence: 1}
	event := Event{GameID: 44, Sequence: 1, Revision: 1, Ordinal: 0, ID: "event-1", Type: "turn_started", PayloadJSON: []byte(`{"turn":1}`)}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnError(errors.New("event insert failed"))
	mock.ExpectRollback()

	err := store.PersistCommand(before.Revision, after, receipt, []Event{event})
	if err == nil || !regexp.MustCompile(`storing gameplay event.*event insert failed`).MatchString(err.Error()) {
		t.Fatalf("PersistCommand() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestValidateAggregateRejectsUnpinnedOrInvalidV1State(t *testing.T) {
	aggregate := fixtureAggregate()
	aggregate.Ruleset.ContentHash = ""
	if err := ValidateAggregate(aggregate); err == nil {
		t.Fatal("ValidateAggregate() accepted v1 aggregate without an immutable ruleset hash")
	}

	aggregate = fixtureAggregate()
	aggregate.CanonicalStateJSON = []byte(`[]`)
	if err := ValidateAggregate(aggregate); err == nil {
		t.Fatal("ValidateAggregate() accepted non-object canonical state")
	}
}

func TestReadAggregateRejectsUnknownSchemaVersion(t *testing.T) {
	db, mock := newSQLMock(t)
	store := New(db)
	mock.ExpectQuery(regexp.QuoteMeta(selectAggregateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{
		"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version",
	}).AddRow(44, 99, 0, 0, `{}`, `{}`, "hash", "v1"))
	if _, err := store.ReadAggregate(44); err == nil {
		t.Fatal("ReadAggregate() accepted an unknown state schema version")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestAggregateUsesCanonicalStateAsItsOnlyDynamicAuthority(t *testing.T) {
	aggregateType := reflect.TypeOf(Aggregate{})
	for _, field := range []string{"DeadlineAt", "VisibilityMemoryJSON"} {
		if _, exists := aggregateType.FieldByName(field); exists {
			t.Fatalf("Aggregate exposes %s outside CanonicalStateJSON", field)
		}
	}
	for _, query := range []string{selectAggregateQuery, updateAggregateQuery} {
		if strings.Contains(query, "deadline_at") || strings.Contains(query, "visibility_memory_json") {
			t.Fatalf("gameplay aggregate SQL exposes a competing dynamic authority: %s", query)
		}
	}
}

func TestGameplayMigrationTerminalConstraintRequiresCompleteMetadata(t *testing.T) {
	migration, err := os.ReadFile("../../../supabase/migrations/20260719120000_gameplay_persistence_task2.sql")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	body := string(migration)
	for _, fragment := range []string{
		`"status" = 'ended' and "ended_at" is not null and "ended_by" is not null and "end_reason" is not null`,
		`"status" <> 'ended' and "ended_at" is null and "ended_by" is null and "end_reason" is null`,
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("terminal metadata constraint is missing SQL contract %q", fragment)
		}
	}
}

func TestReadAggregateRejectsV1WithoutImmutableRulesetPin(t *testing.T) {
	db, mock := newSQLMock(t)
	store := New(db)
	mock.ExpectQuery(regexp.QuoteMeta(selectAggregateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{
		"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version",
	}).AddRow(44, StateSchemaV1, 0, 0, `{}`, nil, nil, nil))

	if _, err := store.ReadAggregate(44); err == nil {
		t.Fatal("ReadAggregate() accepted v1 state without its immutable snapshot ruleset pin")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestReadAggregateAcceptsInitializedV1StateWithImmutablePin(t *testing.T) {
	db, mock := newSQLMock(t)
	store := New(db)
	pinJSON := `{"version":"ctp2-v1"}`
	mock.ExpectQuery(regexp.QuoteMeta(selectAggregateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{
		"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version",
	}).AddRow(44, StateSchemaV1, 0, 0, `{"phase":"setup"}`, pinJSON, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "ctp2-v1"))

	aggregate, err := store.ReadAggregate(44)
	if err != nil {
		t.Fatalf("ReadAggregate() error = %v", err)
	}
	if aggregate.StateSchemaVersion != StateSchemaV1 || aggregate.Revision != 0 || aggregate.LastEventSequence != 0 {
		t.Fatalf("unexpected initialized v1 aggregate: %+v", aggregate)
	}
	if string(aggregate.CanonicalStateJSON) != `{"phase":"setup"}` || string(aggregate.Ruleset.CanonicalJSON) != pinJSON {
		t.Fatalf("v1 aggregate read lost canonical state or ruleset pin: %+v", aggregate)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGameplayMigrationHasNoDeadlineOrFogAuthorityColumns(t *testing.T) {
	migration, err := os.ReadFile("../../../supabase/migrations/20260719120000_gameplay_persistence_task2.sql")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	for _, forbidden := range []string{"deadline_at", "visibility_memory_json"} {
		if strings.Contains(string(migration), forbidden) {
			t.Fatalf("migration contains competing mutable authority %q", forbidden)
		}
	}
}

func TestAIExplorationHistoryMigrationIsScopedMetadataOnly(t *testing.T) {
	migration, err := os.ReadFile("../../../supabase/migrations/20260810160000_gameplay_ai_exploration_history_task17.sql")
	if err != nil {
		t.Fatal(err)
	}
	body := string(migration)
	for _, required := range []string{
		`primary key ("game_id", "actor_player_id", "sequence")`,
		`foreign key ("game_id", "actor_player_id")`,
		`(game_id, actor_player_id, sequence)`,
		`"command_type" text not null`,
		`"destination_tile_id" text not null`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("AI exploration migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"canonical_state", "runner", "campaign", "foreign_player", "target_city"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("AI exploration migration contains forbidden authority %q", forbidden)
		}
	}
}

func TestAIExplorationHistoryIsTransactionalAndActorScoped(t *testing.T) {
	db, mock := newSQLMock(t)
	commandStore := New(db)
	before := fixtureAggregate()
	after := before
	after.Revision, after.LastEventSequence = 1, 1
	after.CanonicalStateJSON = []byte(`{"phase":"turn"}`)
	receipt := CommandReceipt{GameID: 44, ActorPlayerID: 7, IdempotencyKey: "ai-move", PayloadSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", ResultJSON: []byte(`{"ok":true}`), ResultingRevision: 1, ResultingEventSequence: 1}
	step := AIExplorationStep{GameID: 44, ActorPlayerID: 7, Sequence: 1, CommandType: "move-unit", DestinationTileID: "visible-step"}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertAIExplorationStepQuery)).WithArgs(step.GameID, step.ActorPlayerID, step.Sequence, step.CommandType, step.DestinationTileID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := commandStore.PersistCommand(before.Revision, after, receipt, nil, step); err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectAIExplorationHistoryQuery)).WithArgs(int64(44), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "actor_player_id", "sequence", "command_type", "destination_tile_id"}).AddRow(44, 7, 1, "move-unit", "visible-step"))
	steps, err := commandStore.AIExplorationHistory(44, 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0] != step {
		t.Fatalf("AI exploration history = %#v, want %#v", steps, step)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
