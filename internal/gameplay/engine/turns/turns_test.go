package turns

import (
	"reflect"
	"testing"
	"time"
)

func fixture(mode string) State {
	return State{
		Mode:        mode,
		Version:     ContractVersion,
		Turn:        1,
		ActiveOrder: []string{"player-a", "player-b", "player-c"},
		Players: []Player{
			{ID: "player-a", Eligible: true},
			{ID: "player-b", Eligible: true},
			{ID: "player-c", Eligible: false},
		},
	}
}

func action(actor, class, unit string, sequence int64) Command {
	return Command{Kind: CommandAction, ActorID: actor, ActionClass: class, UnitOrArmyID: unit, ReceiptSequence: sequence}
}

func TestSequentialPolicyPhaseTable(t *testing.T) {
	policy, err := New("sequential", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	state := fixture("sequential")
	state, result, err := policy.Apply(state, action("player-a", "economic", "unit-2", 2), time.Unix(0, 0))
	if err != nil || !result.Immediate || state.ActivePlayer() != "player-a" {
		t.Fatalf("immediate action = %#v, %#v, %v", state, result, err)
	}
	state, _, err = policy.Apply(state, Command{Kind: CommandPass, ActorID: "player-a"}, time.Unix(0, 0))
	if err != nil || state.ActivePlayer() != "player-b" || state.Recalculations != 0 {
		t.Fatalf("first pass = %#v, %v", state, err)
	}
	state, result, err = policy.Apply(state, Command{Kind: CommandPass, ActorID: "player-b"}, time.Unix(0, 0))
	if err != nil || !result.Resolved || state.Turn != 2 || state.ActivePlayer() != "player-a" || state.Recalculations != 1 {
		t.Fatalf("final pass = %#v, %#v, %v", state, result, err)
	}
}

func TestLongTurnDeadlineIsOneIdempotentSystemPass(t *testing.T) {
	policy, err := New("long_turn", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0).UTC()
	state, _, err := policy.Start(fixture("long_turn"), now)
	if err != nil || state.Deadline == nil || state.Deadline.PlayerID != "player-a" {
		t.Fatalf("start = %#v, %v", state, err)
	}
	first, result, err := policy.Expire(state, now.Add(time.Minute))
	if err != nil || !result.SystemPass || first.ActivePlayer() != "player-b" {
		t.Fatalf("expire = %#v, %#v, %v", first, result, err)
	}
	retry, retryResult, err := policy.Expire(first, now.Add(time.Minute))
	if err != nil || retryResult.SystemPass || !reflect.DeepEqual(first, retry) {
		t.Fatalf("retry = %#v, %#v, %v", retry, retryResult, err)
	}
	manual, _, err := policy.Apply(state, Command{Kind: CommandPass, ActorID: "player-a"}, now)
	if err != nil || manual.ActivePlayer() != "player-b" {
		t.Fatalf("manual pass = %#v, %v", manual, err)
	}
	afterManual, result, err := policy.Expire(manual, now)
	if err != nil || result.SystemPass || !reflect.DeepEqual(manual, afterManual) {
		t.Fatalf("manual supersedes deadline = %#v, %#v, %v", afterManual, result, err)
	}
}

func TestSimultaneousPolicyQueuesAndResolvesByStableTuple(t *testing.T) {
	policy, err := New("simultaneous", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(0, 0).UTC()
	state := fixture("simultaneous")
	for _, command := range []Command{action("player-b", "hostile", "army-9", 9), action("player-a", "economic", "unit-3", 3), action("player-a", "economic", "unit-1", 1)} {
		var result Result
		state, result, err = policy.Apply(state, command, now)
		if err != nil || result.Immediate {
			t.Fatalf("queue %#v: %#v, %v", command, result, err)
		}
	}
	state, _, err = policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-a"}, now)
	if err != nil {
		t.Fatal(err)
	}
	var result Result
	state, result, err = policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-b"}, now)
	if err != nil || !result.Resolved {
		t.Fatalf("resolve = %#v, %v", result, err)
	}
	want := []Command{action("player-a", "economic", "unit-1", 1), action("player-a", "economic", "unit-3", 3), action("player-b", "hostile", "army-9", 9)}
	if !reflect.DeepEqual(result.ResolvedCommands, want) {
		t.Fatalf("tuple order = %#v; want %#v", result.ResolvedCommands, want)
	}
	if state.ResolutionLocked || len(state.Queue) != 0 || state.Recalculations != 1 {
		t.Fatalf("resolved state = %#v", state)
	}
}

func TestHybridRejectsCrossComponentHostileAndCompletesComponents(t *testing.T) {
	policy, err := New("hybrid", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(0, 0).UTC()
	state := fixture("hybrid")
	state.Wars = []War{{A: "player-a", B: "player-b"}}
	state, _, err = policy.Apply(state, action("player-a", "economic", "unit-1", 1), now)
	if err != nil {
		t.Fatal(err)
	}
	state, _, err = policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-a"}, now)
	if err != nil {
		t.Fatal(err)
	}
	state, result, err := policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-b"}, now)
	if err != nil || !result.Resolved || state.Stage != StageConflict {
		t.Fatalf("planning resolve = %#v, %#v, %v", state, result, err)
	}
	_, _, err = policy.Apply(state, Command{Kind: CommandAction, ActorID: "player-a", Hostile: true, TargetPlayerID: "player-c", ReceiptSequence: 2}, now)
	if Status(err) != 422 {
		t.Fatalf("cross component status = %v (%v)", Status(err), err)
	}
	for _, actor := range []string{"player-a", "player-b"} {
		state, _, err = policy.Apply(state, Command{Kind: CommandPass, ActorID: actor}, now)
		if err != nil {
			t.Fatal(err)
		}
	}
	if state.Stage != StagePlanning || state.Turn != 2 || state.Recalculations != 1 {
		t.Fatalf("completed hybrid = %#v", state)
	}
}

func TestPlanningEarlyEndIsClearedByFurtherEligibleCommand(t *testing.T) {
	policy, err := New("simultaneous", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	state := fixture("simultaneous")
	now := time.Unix(0, 0)
	state, _, err = policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-a"}, now)
	if err != nil || !state.HasEarlyEnd("player-a") {
		t.Fatalf("early end = %#v, %v", state, err)
	}
	state, _, err = policy.Apply(state, action("player-a", "economic", "unit-1", 1), now)
	if err != nil || state.HasEarlyEnd("player-a") {
		t.Fatalf("new action should clear early end = %#v, %v", state, err)
	}
}
