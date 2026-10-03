package engine

import "testing"

func TestAdvertiseAppliesRetailUnhappinessAttack(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "target", OwnerID: "player-b", TileID: "tile-b", Happiness: 75}}
	state.Units = []Unit{{ID: "corp", OwnerID: "player-a", TileID: "tile-a", Movement: 1, Effects: []Effect{{ID: "advertise", Kind: EffectAdvertise}, {ID: "cause", Kind: EffectSpecialCauseUnhappiness, Value: 100}, {ID: "timer", Kind: EffectSpecialCauseUnhappinessTimer, Value: 2}, {ID: "amount", Kind: EffectSpecialCauseUnhappinessAmount, Value: 5}}}}

	eventType, _, err := applySpecialCommand(&state, Command{Type: CommandAdvertise, ActorPlayerID: "player-a", Payload: mustPayload(t, SpecialActionPayload{UnitID: "corp", CityID: "target"})}, fixtureSources(), nil)
	if err != nil {
		t.Fatalf("advertise: %v", err)
	}
	if eventType != EventAdvertised {
		t.Fatalf("event type = %q, want %q", eventType, EventAdvertised)
	}
	if state.Units[0].Movement != 0 {
		t.Fatalf("advertiser movement = %d, want spent", state.Units[0].Movement)
	}
	found := false
	for _, effect := range state.Cities[0].Effects {
		if effect.Kind == EffectHappiness && effect.Value == -5 && effect.Turns == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("city effects after advertise = %#v", state.Cities[0].Effects)
	}
}

func TestAdvertiseRequiresAdvertiseFlag(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "target", OwnerID: "player-b", TileID: "tile-b"}}
	state.Units = []Unit{{ID: "agitator", OwnerID: "player-a", TileID: "tile-a", Movement: 1, Effects: []Effect{{ID: "cause", Kind: EffectSpecialCauseUnhappiness, Value: 100}}}}

	if _, _, err := applySpecialCommand(&state, Command{Type: CommandAdvertise, ActorPlayerID: "player-a", Payload: mustPayload(t, SpecialActionPayload{UnitID: "agitator", CityID: "target"})}, fixtureSources(), nil); err == nil {
		t.Fatal("advertise succeeded without retail Advertise flag")
	}
}
