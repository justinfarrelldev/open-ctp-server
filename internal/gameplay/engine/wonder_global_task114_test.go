package engine

import "testing"

func TestRetailEmbassiesEverywhereEvenAtWarGrantsEffectiveEmbassy(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}},
		Nations: []Nation{{ID: "nation-a"}, {ID: "nation-b"}},
		Policy:  TurnPolicyState{Wars: []TurnPolicyWar{{A: "player-a", B: "player-b"}}},
		Wonders: []Wonder{{ID: "WONDER_GLOBAL_SURVEILLANCE", OwnerID: "player-a", Effects: []Effect{{ID: "embassy", Kind: EffectEmbassiesEverywhere}}}},
	}

	if !HasEffectiveEmbassyWith(state, "player-a", "player-b") {
		t.Fatal("embassies-everywhere-even-at-war wonder did not grant an effective embassy while at war")
	}
	state.Wonders = append(state.Wonders, Wonder{ID: "WONDER_CLOSE", OwnerID: "player-b", Effects: []Effect{{ID: "close", Kind: EffectCloseEmbassies}}})
	if HasEffectiveEmbassyWith(state, "player-a", "player-b") {
		t.Fatal("close-embassies target still accepted an effective embassy")
	}
}
