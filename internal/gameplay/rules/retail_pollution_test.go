package rules

import "testing"

func TestRetailPollutionBeginTurnMatchesCTP2TriggerSemantics(t *testing.T) {
	list := RetailPollutionTriggerList{
		ID:            "POLLUTION_TRIGGER_LIST_SMALL_MAP",
		TriggerLevels: []int64{4000000, 6000000},
		DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER", "POLLUTION_OZONE_DISASTER"},
	}
	state := RetailPollutionState{Phase: 0, History: []int64{3999999, 3800000}}

	next, event := RetailPollutionBeginTurn(state, list, []int64{2}, false)
	if !event.Triggered || event.DisasterType != "POLLUTION_FLOOD_DISASTER" || event.TriggerLevel != 4000000 {
		t.Fatalf("event = %+v, want flood trigger at 4000000", event)
	}
	if next.Phase != 1 {
		t.Fatalf("phase = %d, want 1", next.Phase)
	}

	_, equalEvent := RetailPollutionBeginTurn(RetailPollutionState{History: []int64{3999998}}, list, []int64{2}, false)
	if equalEvent.Triggered {
		t.Fatalf("pollution equal to trigger should not trigger: %+v", equalEvent)
	}

	blocked, blockedEvent := RetailPollutionBeginTurn(state, list, []int64{2000000}, true)
	if blockedEvent.Triggered {
		t.Fatalf("world-pollution reduction should suppress disasters: %+v", blockedEvent)
	}
	if blocked.Phase != 0 {
		t.Fatalf("suppressed phase = %d, want 0", blocked.Phase)
	}
}

func TestRetailPollutionEndRoundMatchesCTP2HistoryAccumulation(t *testing.T) {
	state := RetailPollutionState{History: []int64{100, 90, 80, 70, 60}}
	next := RetailPollutionEndRound(state, []int64{10, 15})
	if got, want := next.History, []int64{125, 100, 90, 80, 70}; !sameInt64s(got, want) {
		t.Fatalf("history = %#v, want %#v", got, want)
	}
	if next.Trend == "" {
		t.Fatalf("trend was not recalculated")
	}
}

func TestRetailPollutionRoundsToNextDisasterMatchesCTP2Projection(t *testing.T) {
	list := RetailPollutionTriggerList{TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}}
	state := RetailPollutionState{History: []int64{3900000, 3800000}}
	if got := RetailPollutionRoundsToNextDisaster(state, list, false); got != 1 {
		t.Fatalf("rounds = %d, want 1", got)
	}
	if got := RetailPollutionRoundsToNextDisaster(RetailPollutionState{History: []int64{3900000, 3900000}}, list, false); got != 0 {
		t.Fatalf("flat pollution rounds = %d, want 0", got)
	}
	if got := RetailPollutionRoundsToNextDisaster(state, list, true); got != 0 {
		t.Fatalf("gaia-controlled rounds = %d, want 0", got)
	}
}

func TestRetailPollutionUsesReleasedGameTriggerList(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	list, ok := db.PollutionTriggerList("POLLUTION_TRIGGER_LIST_SMALL_MAP")
	if !ok {
		t.Fatalf("small map trigger list missing")
	}
	_, event := RetailPollutionBeginTurn(RetailPollutionState{History: []int64{3999999}}, list, []int64{2}, false)
	if !event.Triggered || event.DisasterType != "POLLUTION_FLOOD_DISASTER" {
		t.Fatalf("released trigger event = %+v, want flood", event)
	}
}
