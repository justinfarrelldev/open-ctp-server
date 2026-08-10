package turns

import (
	"errors"
	"testing"
	"time"
)

func TestPolicyRejectsInvalidAuthorityAndDeadlineStates(t *testing.T) {
	if _, err := New("unknown", time.Minute); err == nil {
		t.Fatal("unknown mode accepted")
	}
	if _, err := New(ModeLongTurn, 0); err == nil {
		t.Fatal("zero long-turn deadline accepted")
	}
	policy, _ := New(ModeSequential, time.Minute)
	bad := fixture(ModeSequential)
	bad.Version = "v0"
	if _, _, err := policy.Start(bad, time.Unix(0, 0)); err == nil {
		t.Fatal("bad version accepted")
	}
	bad = fixture(ModeSequential)
	bad.ActiveIndex = -1
	if _, _, err := policy.Apply(bad, Command{}, time.Unix(0, 0)); err == nil {
		t.Fatal("bad index accepted")
	}
	state := fixture(ModeSequential)
	if _, _, err := policy.Apply(state, Command{Kind: CommandPass, ActorID: "player-b"}, time.Unix(0, 0)); Status(err) != 422 {
		t.Fatalf("wrong active status: %v", err)
	}
	if _, _, err := policy.Apply(state, Command{Kind: "bad", ActorID: "player-a"}, time.Unix(0, 0)); Status(err) != 422 {
		t.Fatalf("bad sequential command: %v", err)
	}
	state.ResolutionLocked = true
	if _, _, err := policy.Apply(state, action("player-a", "economic", "u", 1), time.Unix(0, 0)); Status(err) != 409 {
		t.Fatalf("lock status: %v", err)
	}
	if Status(errors.New("other")) != 0 {
		t.Fatal("non-policy error status")
	}
	if (State{ActiveIndex: 2, ActiveOrder: []string{"a"}}).ActivePlayer() != "" {
		t.Fatal("invalid active player")
	}
}

func TestPlanningDeadlineAndHybridConflictBranches(t *testing.T) {
	now := time.Unix(10, 0)
	policy, _ := New(ModeSimultaneous, time.Minute)
	state, _, err := policy.Start(fixture(ModeSimultaneous), now)
	if err != nil || state.Deadline == nil {
		t.Fatalf("start deadline: %#v %v", state, err)
	}
	unchanged, result, err := policy.Expire(state, now)
	if err != nil || result.Resolved || unchanged.Deadline == nil {
		t.Fatalf("premature expiry: %#v %#v %v", unchanged, result, err)
	}
	state, result, err = policy.Expire(state, now.Add(time.Minute))
	if err != nil || !result.Resolved || state.Turn != 2 {
		t.Fatalf("deadline resolution: %#v %#v %v", state, result, err)
	}

	hybrid, _ := New(ModeHybrid, time.Minute)
	state = fixture(ModeHybrid)
	state.Wars = []War{{A: "player-a", B: "player-b"}}
	state, _, _ = hybrid.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-a"}, now)
	state, _, _ = hybrid.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-b"}, now)
	if _, _, err := hybrid.Apply(state, Command{Kind: "bad", ActorID: "player-a"}, now); Status(err) != 422 {
		t.Fatalf("bad conflict command: %v", err)
	}
	immediate, conflictResult, err := hybrid.Apply(state, Command{Kind: CommandAction, ActorID: "player-a", Hostile: false, ReceiptSequence: 3}, now)
	if err != nil || !conflictResult.Immediate || immediate.ActivePlayer() != "player-a" {
		t.Fatalf("connected conflict action = %#v, %#v, %v", immediate, conflictResult, err)
	}
	state.ComponentIndex = 9
	if _, _, err := hybrid.Apply(state, Command{Kind: CommandPass, ActorID: "player-a"}, now); Status(err) != 422 {
		t.Fatalf("missing component: %v", err)
	}
	if smallest(nil) != "" || smallest([]string{"z", "a"}) != "a" {
		t.Fatal("component minimum ordering")
	}
	if actionPriority("non_hostile") != 1 || actionPriority("hostile") != 2 || actionPriority("other") != 3 {
		t.Fatal("unknown action priority")
	}

	// The stable resolver must not use enqueue order for equal phase actions.
	state = fixture(ModeSimultaneous)
	state.Players[2].Eligible = true // keeps a third planner open while A clears its vote.
	state, _, _ = policy.Apply(state, action("player-b", "non_hostile", "unit-2", 2), now)
	state, _, _ = policy.Apply(state, action("player-a", "economic", "unit-2", 4), now)
	state, _, _ = policy.Apply(state, action("player-a", "economic", "unit-1", 3), now)
	state, _, _ = policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-b"}, now)
	// Repeating early-end is an idempotent retry, not a second vote.
	state, retry, err := policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-b"}, now)
	if err != nil || retry.Resolved || len(state.EarlyEnds) != 1 {
		t.Fatalf("early-end retry = %#v, %#v, %v", state, retry, err)
	}
	// A new eligible command clears only that owner's early-end marker.
	state, _, err = policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-a"}, now)
	if err != nil {
		t.Fatal(err)
	}
	state, _, err = policy.Apply(state, action("player-a", "economic", "unit-3", 5), now)
	if err != nil || state.HasEarlyEnd("player-a") || !state.HasEarlyEnd("player-b") {
		t.Fatalf("clear owner marker = %#v, %v", state, err)
	}
	state, _, err = policy.Apply(state, Command{Kind: CommandEarlyEnd, ActorID: "player-a"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := policy.Apply(state, action("player-a", "economic", "unit-4", 6), now); Status(err) != 0 {
		t.Fatalf("next planning phase action: %v", err)
	}
	commands := []Command{
		action("player-a", "economic", "unit-1", 8),
		action("player-a", "economic", "unit-1", 7),
	}
	sortCommands(commands)
	if commands[0].ReceiptSequence != 7 {
		t.Fatalf("receipt sequence tie-break = %#v", commands)
	}
	if (policy.(interface{ Mode() string })).Mode() != ModeSimultaneous {
		t.Fatal("mode contract")
	}
	if (&Error{StatusCode: 422, Message: "invalid"}).Error() != "invalid" {
		t.Fatal("policy error text")
	}
	if (State{Players: []Player{{ID: "a", Eligible: true}}}).eligible("missing") {
		t.Fatal("unknown player eligible")
	}
	if firstEligibleIndex(State{ActiveOrder: []string{"missing"}}) != 0 {
		t.Fatal("missing eligible fallback")
	}
}
