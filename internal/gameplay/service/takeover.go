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

// PendingTakeoverWorker submits persisted pending human-to-AI transitions through
// the ordinary gameplay command path once the canonical policy is unlocked.
type PendingTakeoverWorker struct {
	service     *Service
	AfterCommit func(context.Context) error
}

func NewPendingTakeoverWorker(service *Service) *PendingTakeoverWorker {
	return &PendingTakeoverWorker{service: service}
}

func (w *PendingTakeoverWorker) RunOnce(ctx context.Context) error {
	if w == nil || w.service == nil || w.service.store == nil {
		return fmt.Errorf("gameplay pending takeover worker is not configured")
	}
	takeovers, err := w.service.store.FindPendingTakeovers()
	if err != nil {
		return err
	}
	for _, takeover := range takeovers {
		accountID, err := takeoverAccountID(takeover.PlayerID)
		if err != nil {
			return err
		}
		key := takeoverReceiptKey(takeover.GameID, takeover.PlayerID, takeover.Revision)
		_, err = w.service.ExecuteCommand(ctx, &auth.Session{AccountID: int(accountID), ExpiresAt: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)}, CommandRequest{
			GameID:           takeover.GameID,
			ExpectedRevision: takeover.Revision,
			IdempotencyKey:   key,
			CommandID:        key,
			Type:             engine.CommandTakeoverAI,
			Payload:          json.RawMessage(`{}`),
		})
		if err != nil {
			if harmlessTakeoverConflict(err) {
				continue
			}
			return fmt.Errorf("completing pending takeover for game %d player %q: %w", takeover.GameID, takeover.PlayerID, err)
		}
		if w.AfterCommit != nil {
			if err := w.AfterCommit(ctx); err != nil {
				return fmt.Errorf("dispatching pending takeover for game %d player %q: %w", takeover.GameID, takeover.PlayerID, err)
			}
		}
	}
	return nil
}

func takeoverReceiptKey(gameID int64, playerID string, revision int64) string {
	return fmt.Sprintf("takeover:%d:%s:%d", gameID, playerID, revision)
}

func takeoverAccountID(playerID string) (int64, error) {
	accountID, ok := strings.CutPrefix(playerID, "account-")
	if !ok || accountID == "" {
		return 0, fmt.Errorf("pending takeover has invalid player %q", playerID)
	}
	parsed, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("pending takeover has invalid player %q", playerID)
	}
	return parsed, nil
}

func harmlessTakeoverConflict(err error) bool {
	message := err.Error()
	switch StatusCode(err) {
	case 403:
		return strings.Contains(message, "player controller is not authorized to submit this command") || strings.Contains(message, "player is not authorized to control this command")
	case 409:
		return strings.Contains(message, "gameplay revision is stale")
	case 422:
		return strings.Contains(message, "player has no scheduled takeover") || strings.Contains(message, "takeover is allowed only at a policy boundary") || strings.Contains(message, "player is not active")
	default:
		return false
	}
}
