package store

import (
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestAISchedulingReadsBoundedMetadataNotUnrelatedWorlds(t *testing.T) {
	if strings.Contains(StartedGameplayAggregatesQuery, "state.canonical_state_json::text") || !strings.Contains(StartedGameplayAggregatesQuery, "jsonb_build_object") {
		t.Fatal("AI discovery still transfers complete worlds")
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := New(sqlx.NewDb(db, "sqlmock"))
	mock.ExpectQuery(regexp.QuoteMeta(gameplaySchedulingAggregateQuery)).WithArgs(int64(44)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "revision", "canonical_state_json"}).AddRow(44, 7, `{"players":[],"phase":{"mode":"sequential"}}`))
	aggregate, err := store.FindGameplaySchedulingAggregate(44)
	if err != nil || aggregate.GameID != 44 || aggregate.Revision != 7 {
		t.Fatalf("metadata: %+v err=%v", aggregate, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(gameplaySchedulingAggregateQuery)).WithArgs(int64(45)).WillReturnError(sql.ErrNoRows)
	aggregate, err = store.FindGameplaySchedulingAggregate(45)
	if err != nil || aggregate.GameID != 0 {
		t.Fatalf("ended/deleted game: %+v err=%v", aggregate, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
