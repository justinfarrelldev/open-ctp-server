package engine

import (
	"fmt"
	"testing"
)

func TestClientPresentationWindowsAreBoundedPerOwner(t *testing.T) {
	state := task10State()
	state.Nations = []Nation{{ID: "nation-a"}, {ID: "nation-b"}}
	cityA := City{ID: "home-a", OwnerID: "player-a", TileID: "a"}
	cityB := City{ID: "home-b", OwnerID: "player-b", TileID: "b"}
	item := ProductionItem{ID: "unit", Kind: ProductionUnit}
	if err := recordProductionNotification(&state, cityB, item, "b-unit"); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < NotificationWindow+5; index++ {
		state.Revision++
		if err := recordProductionNotification(&state, cityA, item, fmt.Sprintf("a-unit-%d", index)); err != nil {
			t.Fatal(err)
		}
	}
	counts := map[string]int{}
	for _, notice := range state.Notifications {
		counts[notice.PlayerID]++
	}
	if counts["player-a"] != NotificationWindow || counts["player-b"] != 1 {
		t.Fatalf("one empire evicted another's results: %v", counts)
	}
	for turn := 1; turn <= StatisticsWindow+3; turn++ {
		state.Phase.Turn = turn
		recordTurnStatistics(&state, state.Phase.Turn)
	}
	for _, nation := range state.Nations {
		if len(nation.Statistics) != StatisticsWindow || nation.Statistics[0].Turn != 4 {
			t.Fatalf("statistics window is not bounded: %d", len(nation.Statistics))
		}
	}
}

func TestClientProductionNotificationsAndStatisticsPersistWithoutDuplicateRecalculation(t *testing.T) {
	state := task10State()
	state.Nations = []Nation{{ID: "nation-a"}, {ID: "nation-b"}}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2, Queue: []ProductionItem{{ID: "granary", Kind: ProductionBuilding, Cost: 1}}}}
	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Notifications) != 1 || next.Notifications[0].PlayerID != "player-a" || next.Notifications[0].ContentID != "granary" || next.Notifications[0].CityID != "home" || next.Notifications[0].ID == "" {
		t.Fatalf("completion notification absent: %+v", next.Notifications)
	}
	if len(next.Nations[0].Statistics) != 1 || next.Nations[0].Statistics[0].CityCount != 1 || next.Nations[0].Statistics[0].Turn != state.Phase.Turn {
		t.Fatal("completed-turn statistics not recorded")
	}
	raw, err := CanonicalStateJSON(next)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := RecalculateRound(restored)
	if err != nil {
		t.Fatal(err)
	}
	if len(repeat.Notifications) != 1 || repeat.Notifications[0].ID != next.Notifications[0].ID || len(repeat.Nations[0].Statistics) != 1 {
		t.Fatal("resync/recalculation duplicated history or result")
	}
	if len(state.Notifications) != 0 || len(state.Nations[0].Statistics) != 0 {
		t.Fatal("history mutated the caller's state")
	}
}
