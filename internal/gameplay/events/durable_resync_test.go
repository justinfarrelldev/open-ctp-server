package events

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestResyncUsesDurableEventCursorNotSnapshotRevision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(snapshotForResyncQuery)).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "last_event_sequence", "canonical_state_json"}).AddRow(8, 17, `{}`))
	items, err := NewStore(sqlx.NewDb(db, "sqlmock")).resync(context.Background(), Viewer{GameID: 9, PlayerID: 4}, ProjectorFuncs{})
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Revision int64 `json:"current_revision"`
		Cursor   int64 `json:"event_cursor"`
	}
	if err := json.Unmarshal(items[0].Data, &body); err != nil {
		t.Fatal(err)
	}
	if items[0].ID != 17 || body.Cursor != 17 || body.Revision != 8 {
		t.Fatalf("revision masqueraded as durable cursor: %+v %s", items[0], items[0].Data)
	}
	cursor, ok := writeProjectedItems(httptest.NewRecorder(), items, 0)
	if !ok || cursor != 17 {
		t.Fatal("stream failed to adopt resync durable cursor")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
