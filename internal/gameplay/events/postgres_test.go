package events

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func TestDispatcherPostgresConsumesRowsBeforeMarkingDelivery(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("GAMEPLAY_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("set GAMEPLAY_POSTGRES_DSN to run the PostgreSQL outbox regression test")
	}
	// Restrict every connection (including replacements after a disconnect) to
	// temporary tables so unqualified dispatcher queries cannot reach public.
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		if err != nil {
			t.Fatal(err)
		}
	}
	dsn += " search_path=pg_temp"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Keep all statements on one session. The temporary table shadows the
	// application's outbox, so this test never reads or changes real game data.
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, `CREATE TEMPORARY TABLE gameplay_event_outbox (
		game_id bigint NOT NULL,
		sequence bigint NOT NULL,
		delivered_at timestamptz,
		delivery_attempts integer NOT NULL DEFAULT 0,
		PRIMARY KEY (game_id, sequence)
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO gameplay_event_outbox
		(game_id, sequence, delivered_at, delivery_attempts) VALUES
		(7, 1, NOW(), 1), (9, 1, NULL, 0), (7, 3, NULL, 0), (7, 2, NULL, 0)`); err != nil {
		t.Fatal(err)
	}
	wakeup := &recordingWakeup{}
	if err := NewDispatcher(db, wakeup).DispatchOnce(ctx); err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if got, want := wakeup.gameIDs, []int64{7, 7, 9}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wakeup games = %v, want %v", got, want)
	}
	var marked int
	if err := db.GetContext(ctx, &marked, `SELECT count(*) FROM gameplay_event_outbox
		WHERE delivered_at IS NOT NULL AND delivery_attempts = 1`); err != nil {
		t.Fatal(err)
	}
	if marked != 4 {
		t.Fatalf("marked rows = %d, want 4", marked)
	}
	// A restarted dispatcher must not publish or mark these rows again.
	if err := NewDispatcher(db, wakeup).DispatchOnce(ctx); err != nil {
		t.Fatalf("restart DispatchOnce() error = %v", err)
	}
	if len(wakeup.gameIDs) != 3 {
		t.Fatalf("restart republished delivered rows: %v", wakeup.gameIDs)
	}
	if err := db.GetContext(ctx, &marked, `SELECT count(*) FROM gameplay_event_outbox
		WHERE delivered_at IS NOT NULL AND delivery_attempts = 1`); err != nil {
		t.Fatal(err)
	}
	if marked != 4 {
		t.Fatalf("restart changed delivery attempts: marked rows = %d, want 4", marked)
	}
}
