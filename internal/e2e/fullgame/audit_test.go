package fullgame

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

func auditStateFixture() engine.State {
	return engine.State{SchemaVersion: engine.StateSchemaV1, Rules: engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}, Phase: engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", ActivePlayerOrder: []string{"a"}}, Players: []engine.Player{{ID: "a", NationID: "a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}, RNG: []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}}, Terminal: engine.TerminalState{Ended: true, Reason: "conquest", WinningPlayerID: "a"}}
}

func TestAuditTerminalReconstructsAuditFreeHashesReadOnly(t *testing.T) {
	state := auditStateFixture()
	stateRaw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	events := []engine.Event{
		{SchemaVersion: engine.EventSchemaV1, ID: "start-event-1", Sequence: 1, Ordinal: 0, Type: "round.completed", ActorPlayerID: "a", CommandID: "game-start", Payload: json.RawMessage(`{}`), AuditAt: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)},
		{SchemaVersion: engine.EventSchemaV1, ID: "terminal-event-2", Sequence: 2, Ordinal: 0, Type: "terminal", ActorPlayerID: "a", CommandID: "command-terminal", Payload: json.RawMessage(`{"reason":"conquest"}`), AuditAt: time.Date(2030, 1, 2, 3, 4, 6, 0, time.UTC)},
	}
	eventRaw := make([][]byte, len(events))
	for index, event := range events {
		eventRaw[index], err = json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
	}
	wantState, err := engine.CanonicalStateHash(state)
	if err != nil {
		t.Fatal(err)
	}
	wantEvents, err := engine.CanonicalEventsHash(events)
	if err != nil {
		t.Fatal(err)
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	mock.ExpectQuery(regexp.QuoteMeta(auditStateQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"canonical_state_json"}).AddRow(string(stateRaw)))
	mock.ExpectQuery(regexp.QuoteMeta(auditEventsQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"event_id", "payload_json"}).AddRow(events[0].ID, string(eventRaw[0])).AddRow(events[1].ID, string(eventRaw[1])))
	got, err := AuditTerminal(context.Background(), xdb, 9)
	if err != nil {
		t.Fatal(err)
	}
	if got.StateHash != wantState || got.EventHash != wantEvents || got.TerminalReason != "conquest" {
		t.Fatalf("audit = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAuditTerminalRejectsLegacyEventWithoutID(t *testing.T) {
	state := auditStateFixture()
	stateRaw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	mock.ExpectQuery(regexp.QuoteMeta(auditStateQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"canonical_state_json"}).AddRow(string(stateRaw)))
	mock.ExpectQuery(regexp.QuoteMeta(auditEventsQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"event_id", "payload_json"}).AddRow("", `{}`))
	if _, err := AuditTerminal(context.Background(), xdb, 9); err == nil {
		t.Fatal("legacy event ID was accepted")
	}
}

func TestAuditTerminalRejectsMismatchedPersistedEventID(t *testing.T) {
	stateRaw, err := engine.CanonicalStateJSON(auditStateFixture())
	if err != nil {
		t.Fatal(err)
	}
	eventRaw, err := json.Marshal(engine.Event{SchemaVersion: engine.EventSchemaV1, ID: "payload-event-id", Sequence: 1, Type: "terminal"})
	if err != nil {
		t.Fatal(err)
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlmock")
	mock.ExpectQuery(regexp.QuoteMeta(auditStateQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"canonical_state_json"}).AddRow(string(stateRaw)))
	mock.ExpectQuery(regexp.QuoteMeta(auditEventsQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"event_id", "payload_json"}).AddRow("stored-event-id", string(eventRaw)))
	if _, err := AuditTerminal(context.Background(), xdb, 9); err == nil {
		t.Fatal("mismatched gameplay event ID was accepted")
	}
}
