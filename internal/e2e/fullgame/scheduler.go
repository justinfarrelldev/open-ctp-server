package fullgame

import (
	"context"
	"errors"
	"fmt"
)

type ScheduledStep struct {
	Role    string
	Command string
	Payload []byte
	AI      bool
}

type SchedulerTransport interface {
	Command(context.Context, ScheduledStep) (int64, error)
	DriveAI(context.Context) (int64, error)
	ResumeSSE(context.Context, int64) (int64, error)
	Restart(context.Context) error
	Terminal(context.Context) (Artifacts, bool, error)
}

func ExecuteSchedule(ctx context.Context, steps []ScheduledStep, checkpoint, maxActions int, transport SchedulerTransport) (Artifacts, error) {
	if transport == nil || checkpoint < 1 || maxActions < 1 || len(steps) == 0 || len(steps) > maxActions {
		return Artifacts{}, errors.New("fullgame schedule is invalid")
	}
	var cursor int64
	for index, step := range steps {
		if step.AI {
			var err error
			cursor, err = transport.DriveAI(ctx)
			if err != nil {
				return Artifacts{}, fmt.Errorf("AI step %d: %w", index, err)
			}
		} else {
			if step.Role == "" || step.Command == "" {
				return Artifacts{}, fmt.Errorf("public step %d is invalid", index)
			}
			next, err := transport.Command(ctx, step)
			if err != nil {
				return Artifacts{}, fmt.Errorf("public step %d: %w", index, err)
			}
			cursor = next
		}
		if index+1 == checkpoint {
			if err := transport.Restart(ctx); err != nil {
				return Artifacts{}, fmt.Errorf("restart checkpoint: %w", err)
			}
		}
		resumed, err := transport.ResumeSSE(ctx, cursor)
		if err != nil || resumed < cursor {
			return Artifacts{}, errors.New("SSE cursor did not resume")
		}
		cursor = resumed
	}
	artifacts, ended, err := transport.Terminal(ctx)
	if err != nil || !ended || artifacts.TerminalReason == "" || artifacts.TerminalWinner == "" {
		return Artifacts{}, errors.New("natural terminal state was not observed")
	}
	return artifacts, nil
}
