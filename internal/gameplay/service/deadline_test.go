package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

func TestDeadlineWorkerExecutesDueLongTurnPassOnce(t *testing.T) {
	svc, mock := newMockService(t)
	state := testState()
	state.Phase.Mode = "long_turn"
	state.Players = []engine.Player{
		{ID: "account-9", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman},
		{ID: "account-10", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman},
	}
	state.Phase.ActivePlayerOrder = []string{"account-9", "account-10"}
	state.Policy.Deadline = &engine.TurnPolicyDeadline{PlayerID: "account-9", AtUnix: 60}

	mock.ExpectQuery(regexp.QuoteMeta(store.DueLongTurnDeadlinesQuery)).WithArgs(int64(60)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "revision", "player_id", "deadline_at"}).AddRow(44, 0, "account-9", 60))
	expectPassCommand(t, mock, state, "deadline:44:account-9:60")

	worker := NewDeadlineWorker(svc)
	worker.Clock = fixedClock{now: time.Unix(60, 0).UTC()}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeadlineWorkerManualPassOrRestartCannotDoublePass(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectQuery(regexp.QuoteMeta(store.DueLongTurnDeadlinesQuery)).WithArgs(int64(60)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "revision", "player_id", "deadline_at"}))

	worker := NewDeadlineWorker(svc)
	worker.Clock = fixedClock{now: time.Unix(60, 0).UTC()}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectPassCommand(t *testing.T, mock sqlmock.Sqlmock, state engine.State, key string) {
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), key).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}
