package fullgame

import (
	"context"
	"testing"
)

type schedulerTransport struct{ restarted bool }

func (s *schedulerTransport) Command(_ context.Context, _ ScheduledStep) (int64, error) {
	return 1, nil
}
func (s *schedulerTransport) DriveAI(_ context.Context) (int64, error) { return 2, nil }
func (s *schedulerTransport) ResumeSSE(_ context.Context, cursor int64) (int64, error) {
	return cursor, nil
}
func (s *schedulerTransport) Restart(context.Context) error { s.restarted = true; return nil }
func (s *schedulerTransport) Terminal(context.Context) (Artifacts, bool, error) {
	return Artifacts{TerminalWinner: "account-1", TerminalReason: "score_2300"}, true, nil
}

func TestExecuteScheduleRequiresRestartSSEAIAndTerminal(t *testing.T) {
	transport := &schedulerTransport{}
	_, err := ExecuteSchedule(context.Background(), []ScheduledStep{{Role: "human-1", Command: "found-city"}, {AI: true}}, 2, 2, transport)
	if err != nil || !transport.restarted {
		t.Fatalf("ExecuteSchedule() = %v, restarted=%v", err, transport.restarted)
	}
}
