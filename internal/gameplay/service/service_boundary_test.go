package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

func TestExecuteCommandUsesCanonicalAccountAndVerifiedDefinitions(t *testing.T) {
	svc, mock := newMockService(t)
	definitions := json.RawMessage(`{"profile":"synthetic"}`)
	digest := sha256.Sum256(definitions)
	hash := hex.EncodeToString(digest[:])
	state := testState()
	state.Rules = engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: hash, Definitions: definitions}
	state.Players[0].ID = "account-9"
	state.Phase.ActivePlayerOrder[0] = "account-9"
	body, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), "key-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "state_schema_version", "revision", "last_event_sequence", "canonical_state_json", "ruleset_canonical_json", "ruleset_content_hash", "ruleset_version"}).AddRow(44, 1, 0, 0, string(body), string(definitions), hash, "ctp2-supported-v2"))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if _, err := svc.ExecuteCommand(context.Background(), testSession(), request()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRulesPinPropagatesOnlyHashVerifiedDefinitions(t *testing.T) {
	definitions := json.RawMessage(`{"profile":"synthetic"}`)
	digest := sha256.Sum256(definitions)
	verified := rulesPin(store.RulesetPin{Version: "ctp2-supported-v2", ContentHash: hex.EncodeToString(digest[:]), CanonicalJSON: definitions})
	if string(verified.Definitions) != string(definitions) {
		t.Fatalf("verified definitions = %s", verified.Definitions)
	}
	tampered := rulesPin(store.RulesetPin{Version: verified.Version, ContentHash: verified.ContentHash, CanonicalJSON: json.RawMessage(`{"profile":"tampered"}`)})
	if len(tampered.Definitions) != 0 {
		t.Fatalf("tampered definitions propagated: %s", tampered.Definitions)
	}
}
