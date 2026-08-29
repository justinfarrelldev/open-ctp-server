package fullgame

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestObserveTask9Row(t *testing.T) {
	rowValue, ok := os.LookupEnv("TASK9_ROW")
	if !ok || rowValue == "" {
		t.Skip("set TASK9_ROW=1..12 to observe a single Task 9 catalog row")
	}
	rowNumber, err := strconv.Atoi(rowValue)
	if err != nil {
		t.Fatalf("parse TASK9_ROW %q: %v", rowValue, err)
	}
	scenario, err := task9RowScenario(rowNumber)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateObservationScenario(scenario); err != nil {
		t.Fatalf("row %02d observation preflight: %v", rowNumber, err)
	}
	result, restartEvents, err := observeTask9Row(context.Background(), scenario)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("row=%02d id=%s winner=%s reason=%s state_sha256=%s event_sha256=%s ai_commands=%d", rowNumber, scenario.ID, result.First.TerminalWinner, result.First.TerminalReason, result.First.StateHash, result.First.EventHash, result.First.AICommands)
	for index, event := range restartEvents {
		t.Logf("row=%02d restart[%d]=%+v", rowNumber, index, event)
	}
	t.Logf("row=%02d ai_transcript=%s", rowNumber, result.First.AITranscript)
}

func TestRowSevenObservedScienceGaiaControllerRestartsAtCompletedTurnTwelve(t *testing.T) {
	scenario, err := task9RowScenario(7)
	if err != nil {
		t.Fatal(err)
	}
	result, restartEvents, err := observeTask9Row(context.Background(), scenario)
	if err != nil {
		t.Fatal(err)
	}
	want := AcceptanceRestartEvent{Point: AcceptanceRestartAfterCompletedTurn, CompletedTurn: 12}
	if len(restartEvents) != 1 || restartEvents[0] != want {
		t.Fatalf("row 7 restart events = %+v, want %+v", restartEvents, want)
	}
	if result.First != result.Second || result.First != result.Restarted {
		t.Fatalf("row 7 deterministic artifacts diverged: %+v", result)
	}
}

func observeTask9Row(ctx context.Context, scenario Scenario) (AcceptanceResult, []AcceptanceRestartEvent, error) {
	config, err := scenario.AcceptanceConfig()
	if err != nil {
		return AcceptanceResult{}, nil, err
	}
	config.ExpectedWinner = ""
	config.ExpectedStateHash = ""
	config.ExpectedEventHash = ""
	restartEvents := make([]AcceptanceRestartEvent, 0, 1)
	config.ObserveRestart = func(event AcceptanceRestartEvent) {
		restartEvents = append(restartEvents, event)
	}
	if err := config.validate(); err != nil {
		return AcceptanceResult{}, nil, err
	}
	restartPlan := config.Restart
	config.Restart = AcceptanceRestartPlan{Point: AcceptanceRestartDisabled}
	first, err := runAcceptanceOnce(ctx, config, PublicExecutor{}, "clean-1")
	if err != nil {
		return AcceptanceResult{}, restartEvents, err
	}
	second, err := runAcceptanceOnce(ctx, config, PublicExecutor{}, "clean-2")
	if err != nil {
		return AcceptanceResult{}, restartEvents, err
	}
	config.Restart = restartPlan
	restarted, err := runAcceptanceOnce(ctx, config, PublicExecutor{}, "restart")
	if err != nil {
		return AcceptanceResult{}, restartEvents, err
	}
	if first != second || first != restarted {
		return AcceptanceResult{}, restartEvents, fmt.Errorf("task 9 row %s deterministic artifact mismatch: first=%+v second=%+v restarted=%+v", config.Name, first, second, restarted)
	}
	if first.TerminalWinner == "" || !winnerBelongsToSlots(first.TerminalWinner, config.Slots) {
		return AcceptanceResult{}, restartEvents, fmt.Errorf("task 9 row %s terminal winner %q is not present in the declared slots", config.Name, first.TerminalWinner)
	}
	if first.TerminalReason != config.Rules.Victory {
		return AcceptanceResult{}, restartEvents, fmt.Errorf("task 9 row %s terminal reason = %q, want %q", config.Name, first.TerminalReason, config.Rules.Victory)
	}
	if first.AICommands == 0 || first.AITranscript == "" || first.AITranscript == "[]" {
		return AcceptanceResult{}, restartEvents, fmt.Errorf("task 9 row %s has no production AI gateway audit", config.Name)
	}
	if scenario.hasCompleteObservedArtifacts() {
		if first.TerminalWinner != scenario.Observed.Winner || first.StateHash != scenario.Observed.StateHash || first.EventHash != scenario.Observed.EventHash {
			return AcceptanceResult{}, restartEvents, fmt.Errorf("task 9 row %s observed artifacts drifted: got winner=%s state=%s events=%s want winner=%s state=%s events=%s", config.Name, first.TerminalWinner, first.StateHash, first.EventHash, scenario.Observed.Winner, scenario.Observed.StateHash, scenario.Observed.EventHash)
		}
	}
	return AcceptanceResult{Name: config.Name, First: first, Second: second, Restarted: restarted}, restartEvents, nil
}

func task9RowScenario(rowNumber int) (Scenario, error) {
	rows := CatalogRows01_12()
	if rowNumber < 1 || rowNumber > len(rows) {
		return Scenario{}, fmt.Errorf("TASK9_ROW=%d is outside rows 1-%d", rowNumber, len(rows))
	}
	return rows[rowNumber-1], nil
}
