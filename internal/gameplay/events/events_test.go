package events

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

type recordingWakeup struct {
	mu        sync.Mutex
	gameIDs   []int64
	listeners int
}

func (w *recordingWakeup) Publish(gameID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.gameIDs = append(w.gameIDs, gameID)
}

func (w *recordingWakeup) Subscribe(_ int64) (<-chan struct{}, func()) {
	w.mu.Lock()
	w.listeners++
	w.mu.Unlock()
	return make(chan struct{}), func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		w.listeners--
	}
}

func TestDispatcherDeliversCommittedOutboxRowsInSequenceAndIsRestartSafe(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wakeup := &recordingWakeup{}
	dispatcher := NewDispatcher(sqlx.NewDb(db, "sqlmock"), wakeup)
	rows := sqlmock.NewRows([]string{"game_id", "sequence"}).AddRow(int64(7), int64(2)).AddRow(int64(7), int64(3))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT o.game_id, o.sequence").WillReturnRows(rows)
	mock.ExpectExec("UPDATE gameplay_event_outbox").WithArgs(int64(7), int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE gameplay_event_outbox").WithArgs(int64(7), int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := dispatcher.DispatchOnce(context.Background()); err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	wakeup.mu.Lock()
	got := append([]int64(nil), wakeup.gameIDs...)
	wakeup.mu.Unlock()
	if want := []int64{7, 7}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("wakeup games = %v, want %v", got, want)
	}

	// A restarted dispatcher sees only committed-but-undelivered rows. Once marked,
	// a duplicate invocation must not republish anything.
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT o.game_id, o.sequence").WillReturnRows(sqlmock.NewRows([]string{"game_id", "sequence"}))
	mock.ExpectCommit()
	if err := NewDispatcher(sqlx.NewDb(db, "sqlmock"), wakeup).DispatchOnce(context.Background()); err != nil {
		t.Fatalf("restart DispatchOnce() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDispatcherDoesNotPublishWithoutCommittedOutboxRow(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	wakeup := &recordingWakeup{}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT o.game_id, o.sequence").WillReturnRows(sqlmock.NewRows([]string{"game_id", "sequence"}))
	mock.ExpectCommit()
	if err := NewDispatcher(sqlx.NewDb(db, "sqlmock"), wakeup).DispatchOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(wakeup.gameIDs) != 0 {
		t.Fatalf("published uncommitted event: %v", wakeup.gameIDs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStreamResumeExpiredCursorResyncAndNeverLeaksCanonicalPayload(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stream := NewStore(sqlx.NewDb(db, "sqlmock"))
	viewer := Viewer{GameID: 9, PlayerID: 42}
	projector := ProjectorFuncs{
		Event: func(_ Viewer, event Event) ([]byte, error) {
			return json.Marshal(map[string]any{"safe_event": event.Type})
		},
		Snapshot: func(_ Viewer, _ []byte) ([]byte, error) { return []byte(`{"safe_snapshot":true}`), nil },
	}

	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(int64(5)))
	mock.ExpectQuery("SELECT revision, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "canonical_state_json"}).AddRow(int64(8), `{"hidden_entity_id":"enemy-99","coordinates":[7,8]}`))
	items, err := stream.Resume(context.Background(), viewer, 2, projector)
	if err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "resync" || items[0].ID != 8 {
		t.Fatalf("expired cursor items = %#v", items)
	}
	raw := string(items[0].Data)
	if strings.Contains(raw, "enemy-99") || strings.Contains(raw, "coordinates") {
		t.Fatalf("resync leaked canonical bytes: %s", raw)
	}
	if !strings.Contains(raw, "safe_snapshot") || !strings.Contains(raw, "current_revision") {
		t.Fatalf("resync did not include safe metadata: %s", raw)
	}

	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(int64(5)))
	mock.ExpectQuery("SELECT game_id, sequence, revision, ordinal, event_type, payload_json").WithArgs(int64(9), int64(4)).WillReturnRows(
		sqlmock.NewRows([]string{"game_id", "sequence", "revision", "ordinal", "event_type", "payload_json"}).AddRow(int64(9), int64(5), int64(8), 0, "unit_moved", `{"hidden_entity_id":"enemy-99","coordinates":[7,8]}`),
	)
	items, err = stream.Resume(context.Background(), viewer, 4, projector)
	if err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 5 || items[0].Name != "unit_moved" {
		t.Fatalf("resume items = %#v", items)
	}
	if raw := string(items[0].Data); strings.Contains(raw, "enemy-99") || strings.Contains(raw, "coordinates") || !strings.Contains(raw, "safe_event") {
		t.Fatalf("event projection leaked or omitted safe bytes: %s", raw)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStreamTerminalAndDisconnectCleanup(t *testing.T) {
	t.Parallel()
	wakeup := &recordingWakeup{}
	ctx, cancel := context.WithCancel(context.Background())
	events, cleanup := NewConnection(ctx, wakeup, 11)
	if wakeup.listeners != 1 {
		t.Fatalf("listeners = %d, want 1", wakeup.listeners)
	}
	cancel()
	select {
	case <-events.Done():
	case <-time.After(time.Second):
		t.Fatal("connection did not observe cancellation")
	}
	cleanup()
	cleanup()
	if wakeup.listeners != 0 {
		t.Fatalf("listeners after idempotent cleanup = %d, want 0", wakeup.listeners)
	}

	item, err := TerminalEvent(Viewer{GameID: 11, PlayerID: 2}, []byte(`{"ended":true,"reason":"victory","winning_player_id":"hidden-opponent"}`), ProjectorFuncs{
		Snapshot: func(_ Viewer, _ []byte) ([]byte, error) { return []byte(`{"terminal_view":"complete"}`), nil },
	})
	if err != nil {
		t.Fatalf("TerminalEvent() error = %v", err)
	}
	if item.Name != "terminal" || strings.Contains(string(item.Data), "hidden-opponent") || !strings.Contains(string(item.Data), "terminal_view") {
		t.Fatalf("terminal projection = %#v", item)
	}
}

func TestStoreTerminalReadsDurableStateThroughProjection(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stream := NewStore(sqlx.NewDb(db, "sqlmock"))
	mock.ExpectQuery("SELECT last_event_sequence, canonical_state_json").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(int64(12), `{"terminal":{"ended":true,"winning_player_id":"hidden-opponent"}}`))
	event, ended, err := stream.Terminal(context.Background(), Viewer{GameID: 11, PlayerID: 2}, ProjectorFuncs{Snapshot: func(_ Viewer, _ []byte) ([]byte, error) {
		return []byte(`{"terminal_view":"complete"}`), nil
	}})
	if err != nil || !ended || event.ID != 12 || event.Name != "terminal" {
		t.Fatalf("Terminal() = (%#v, %v, %v)", event, ended, err)
	}
	if raw := string(event.Data); strings.Contains(raw, "hidden-opponent") || !strings.Contains(raw, "terminal_view") {
		t.Fatalf("terminal leaked canonical bytes: %s", raw)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestParseLastEventID(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  string
		want int64
		ok   bool
	}{
		{name: "empty", raw: "", want: 0, ok: true},
		{name: "sequence", raw: "42", want: 42, ok: true},
		{name: "negative", raw: "-1"},
		{name: "decimal", raw: "1.5"},
		{name: "spaces", raw: " 2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseLastEventID(tc.raw)
			if (err == nil) != tc.ok || got != tc.want {
				t.Fatalf("ParseLastEventID(%q) = (%d, %v), want (%d, ok=%v)", tc.raw, got, err, tc.want, tc.ok)
			}
		})
	}
}
