package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailPollutionTriggersDisasterDuringRoundCompletion(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}}
	state.Tiles[0].Pollution = 2
	state.Pollution = &PollutionState{History: []int64{3999999}}
	runtime := &rules.RuntimeRules{PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000, 6000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER", "POLLUTION_OZONE_DISASTER"}}}}

	advanceRound(&state, runtime)
	if state.Pollution.Phase != 1 || state.Pollution.LastDisaster != "POLLUTION_FLOOD_DISASTER" || state.Pollution.LastPollutionLevel != 4000001 || len(state.Pollution.History) != rules.RetailPollutionHistoryTurns || state.Pollution.History[0] != 4000001 {
		t.Fatalf("pollution state = %+v", state.Pollution)
	}
}

func TestRetailPollutionDefaultsToSmallMapTriggerUntilWorldMapSizeIsModeled(t *testing.T) {
	state := runtimeStartedState(t)
	state.Tiles[0].Pollution = 2
	state.Pollution = &PollutionState{History: []int64{3999999}}
	runtime := &rules.RuntimeRules{PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{
		{ID: "POLLUTION_TRIGGER_LIST_HUGE_MAP", TriggerLevels: []int64{35000000}, DisasterTypes: []string{"POLLUTION_OZONE_DISASTER"}},
		{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}},
	}}

	advanceRound(&state, runtime)
	if state.Pollution.LastDisaster != "POLLUTION_FLOOD_DISASTER" {
		t.Fatalf("last disaster = %q, want small-map flood trigger", state.Pollution.LastDisaster)
	}
}
