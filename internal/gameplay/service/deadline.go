package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

// DeadlineWorker submits due long-turn passes through ExecuteCommand after its
// persisted deadline lookup. The receipt and revision guards make restarts and
// races with a manual pass harmless.
type DeadlineWorker struct {
	service     *Service
	Clock       engine.Clock
	AfterCommit func(context.Context) error
}

func NewDeadlineWorker(service *Service) *DeadlineWorker {
	return &DeadlineWorker{service: service, Clock: service.sources.Clock}
}

func (w *DeadlineWorker) RunOnce(ctx context.Context) error {
	if w == nil || w.service == nil || w.service.store == nil || w.Clock == nil {
		return fmt.Errorf("gameplay deadline worker is not configured")
	}
	deadlines, err := w.service.store.FindDueLongTurnDeadlines(w.Clock.Now().UTC().Unix())
	if err != nil {
		return err
	}
	for _, deadline := range deadlines {
		accountID, err := deadlineAccountID(deadline.PlayerID)
		if err != nil {
			return err
		}
		key := deadlineReceiptKey(deadline.GameID, deadline.PlayerID, deadline.DeadlineAt)
		_, err = w.service.ExecuteCommand(ctx, &auth.Session{AccountID: int(accountID), ExpiresAt: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)}, CommandRequest{
			GameID:           deadline.GameID,
			ExpectedRevision: deadline.Revision,
			IdempotencyKey:   key,
			CommandID:        key,
			Type:             engine.CommandPass,
			Payload:          json.RawMessage(`{}`),
		})
		if err != nil {
			if StatusCode(err) == 409 {
				continue
			}
			return fmt.Errorf("passing due long-turn deadline for game %d: %w", deadline.GameID, err)
		}
		if w.AfterCommit != nil {
			if err := w.AfterCommit(ctx); err != nil {
				return fmt.Errorf("dispatching due long-turn deadline for game %d: %w", deadline.GameID, err)
			}
		}
	}
	return nil
}

func deadlineReceiptKey(gameID int64, playerID string, deadlineAt int64) string {
	return fmt.Sprintf("deadline:%d:%s:%d", gameID, playerID, deadlineAt)
}

func deadlineAccountID(playerID string) (int64, error) {
	accountID, ok := strings.CutPrefix(playerID, "account-")
	if !ok || accountID == "" {
		return 0, fmt.Errorf("due deadline has invalid player %q", playerID)
	}
	parsed, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("due deadline has invalid player %q", playerID)
	}
	return parsed, nil
}
