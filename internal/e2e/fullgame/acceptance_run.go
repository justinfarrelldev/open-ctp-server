package fullgame

import (
	"context"
	"fmt"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
)

type AcceptanceExecutor interface {
	Execute(context.Context, AcceptanceConfig, *testkit.Scenario) (Artifacts, error)
}
type AcceptanceResult struct {
	Name                     string
	First, Second, Restarted Artifacts
}

func RunAcceptance(ctx context.Context, config AcceptanceConfig, executor AcceptanceExecutor) (AcceptanceResult, error) {
	if err := config.validate(); err != nil {
		return AcceptanceResult{}, err
	}
	restartPlan := config.Restart
	config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartDisabled}
	first, err := runAcceptanceOnce(ctx, config, executor, "clean-1")
	if err != nil {
		return AcceptanceResult{}, err
	}
	second, err := runAcceptanceOnce(ctx, config, executor, "clean-2")
	if err != nil {
		return AcceptanceResult{}, err
	}
	config.Restart = restartPlan
	restarted, err := runAcceptanceOnce(ctx, config, executor, "restart")
	if err != nil {
		return AcceptanceResult{}, err
	}
	if first != second || first != restarted {
		return AcceptanceResult{}, fmt.Errorf("acceptance %s deterministic artifact mismatch: first=%+v second=%+v restarted=%+v", config.Name, first, second, restarted)
	}
	if first.TerminalWinner != config.ExpectedWinner || first.TerminalReason != config.ExpectedReason {
		return AcceptanceResult{}, fmt.Errorf("acceptance %s terminal metadata = %s/%s, want %s/%s", config.Name, first.TerminalWinner, first.TerminalReason, config.ExpectedWinner, config.ExpectedReason)
	}
	if config.ExpectedStateHash != "" && first.StateHash != config.ExpectedStateHash {
		return AcceptanceResult{}, fmt.Errorf("acceptance %s state hash = %s, want %s", config.Name, first.StateHash, config.ExpectedStateHash)
	}
	if config.ExpectedEventHash != "" && first.EventHash != config.ExpectedEventHash {
		return AcceptanceResult{}, fmt.Errorf("acceptance %s event hash = %s, want %s", config.Name, first.EventHash, config.ExpectedEventHash)
	}
	if first.AICommands == 0 || first.AITranscript == "" || first.AITranscript == "[]" {
		return AcceptanceResult{}, fmt.Errorf("acceptance %s has no production AI gateway audit", config.Name)
	}
	return AcceptanceResult{Name: config.Name, First: first, Second: second, Restarted: restarted}, nil
}
func runAcceptanceOnce(parent context.Context, config AcceptanceConfig, executor AcceptanceExecutor, suffix string) (Artifacts, error) {
	ctx, cancel := context.WithTimeout(parent, config.Budgets.Deadline)
	defer cancel()
	kit := testkit.NewScenario(config.Name+"-"+suffix, int64(config.Seed))
	defer kit.Close()
	artifacts, err := executor.Execute(ctx, config, kit)
	if err != nil {
		return Artifacts{}, fmt.Errorf("%s %s: %w", config.Name, suffix, err)
	}
	if err := ctx.Err(); err != nil {
		return Artifacts{}, fmt.Errorf("%s %s exceeded %s budget: %w", config.Name, suffix, config.Budgets.Deadline, err)
	}
	return artifacts, nil
}
