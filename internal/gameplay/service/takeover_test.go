package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
)

func TestPendingTakeoverWorkerExecutesUnlockedPendingTakeoverOnce(t *testing.T) {
	svc, mock := newMockService(t)
	state := pendingTakeoverState(false)
	mock.ExpectQuery(regexp.QuoteMeta(store.PendingTakeoversQuery)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "revision", "player_id"}).AddRow(44, 0, "account-9"))
	expectTakeoverCommand(t, mock, state, "takeover:44:account-9:0")

	worker := NewPendingTakeoverWorker(svc)
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPendingTakeoverWorkerRetriesLockedPendingTakeoverAtLaterBoundary(t *testing.T) {
	svc, mock := newMockService(t)
	locked := pendingTakeoverState(true)
	unlocked := pendingTakeoverState(false)
	mock.ExpectQuery(regexp.QuoteMeta(store.PendingTakeoversQuery)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "revision", "player_id"}).AddRow(44, 0, "account-9"))
	expectLockedTakeoverCommand(t, mock, locked, "takeover:44:account-9:0")
	mock.ExpectQuery(regexp.QuoteMeta(store.PendingTakeoversQuery)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "revision", "player_id"}).AddRow(44, 0, "account-9"))
	expectTakeoverCommand(t, mock, unlocked, "takeover:44:account-9:0")

	worker := NewPendingTakeoverWorker(svc)
	commits := 0
	worker.AfterCommit = func(context.Context) error {
		commits++
		return nil
	}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if commits != 0 {
		t.Fatalf("locked takeover dispatched %d commits, want 0", commits)
	}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if commits != 1 {
		t.Fatalf("later unlocked takeover dispatched %d commits, want 1", commits)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPendingTakeoverWorkerIgnoresStaleRevisionConflicts(t *testing.T) {
	svc, mock := newMockService(t)
	stale := pendingTakeoverState(false)
	stale.Revision = 1
	stale.EventSequence = 1
	mock.ExpectQuery(regexp.QuoteMeta(store.PendingTakeoversQuery)).WillReturnRows(sqlmock.NewRows([]string{"game_id", "revision", "player_id"}).AddRow(44, 0, "account-9"))
	expectLockedTakeoverCommand(t, mock, stale, "takeover:44:account-9:0")

	worker := NewPendingTakeoverWorker(svc)
	worker.AfterCommit = func(context.Context) error {
		t.Fatal("stale takeover dispatched events")
		return nil
	}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func pendingTakeoverState(locked bool) engine.State {
	state := testState()
	state.Players[0].ID = "account-9"
	state.Players[0].TakeoverPending = true
	state.Phase.ActivePlayerOrder = []string{"account-9"}
	state.Phase.ResolutionLocked = locked
	return state
}

func expectTakeoverCommand(t *testing.T, mock sqlmock.Sqlmock, state engine.State, key string) {
	t.Helper()
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), key).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectExec(regexp.QuoteMeta(updateAggregateQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertReceiptQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertEventQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertOutboxQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateControllerQuery)).WithArgs("ai", int64(44), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateGameStatusQuery)).WithArgs("started", int64(44)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func expectLockedTakeoverCommand(t *testing.T, mock sqlmock.Sqlmock, state engine.State, key string) {
	t.Helper()
	mock.ExpectBegin()
	expectIdentity(mock, "human")
	mock.ExpectQuery(regexp.QuoteMeta(receiptQuery)).WithArgs(int64(44), int64(7), key).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(aggregateForUpdateQuery)).WithArgs(int64(44)).WillReturnRows(aggregateRows(t, state))
	mock.ExpectRollback()
}

func TestTakeoverAccountIDRejectsInvalidPlayers(t *testing.T) {
	for _, playerID := range []string{"", "player-9", "account-0", "account-bad"} {
		if _, err := takeoverAccountID(playerID); err == nil {
			t.Fatalf("takeoverAccountID(%q) unexpectedly succeeded", playerID)
		}
	}
}

func TestTakeoverReceiptKeyIsStable(t *testing.T) {
	if key := takeoverReceiptKey(44, "account-9", 3); key != "takeover:44:account-9:3" {
		t.Fatalf("takeover receipt key = %q", key)
	}
	if key := takeoverReceiptKey(44, "account-9", 3); key != takeoverReceiptKey(44, "account-9", 3) {
		t.Fatalf("takeover receipt key replay changed: %q", key)
	}
}
