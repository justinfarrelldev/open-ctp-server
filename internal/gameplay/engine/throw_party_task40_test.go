package engine

import "testing"

func TestThrowPartyRaisesRegardAndCostsGold(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].Gold = 150
	state.Cities = []City{{ID: "foreign", OwnerID: "player-b", TileID: "tile-b"}}
	state.Units = []Unit{{ID: "diplomat", OwnerID: "player-a", TileID: "tile-a", Movement: 1, Effects: []Effect{{ID: "party", Kind: EffectThrowParty}}}}

	eventType, _, err := applySpecialCommand(&state, Command{Type: CommandThrowParty, ActorPlayerID: "player-a", Payload: mustPayload(t, SpecialActionPayload{UnitID: "diplomat", CityID: "foreign"})}, fixtureSources(), nil)
	if err != nil {
		t.Fatalf("throw party: %v", err)
	}
	if eventType != EventPartyThrown {
		t.Fatalf("event type = %q, want %q", eventType, EventPartyThrown)
	}
	if state.Nations[0].Gold != 50 {
		t.Fatalf("actor gold = %d, want 50 after retail party cost", state.Nations[0].Gold)
	}
	if state.Nations[1].LastPartyTurn != state.Phase.Turn {
		t.Fatalf("target last party turn = %d, want %d", state.Nations[1].LastPartyTurn, state.Phase.Turn)
	}
	relation := relationForPlayers(state, "player-a", "player-b")
	if relation == nil || relation.RegardBonusBToA != retailPartyRegardBonus || relation.RegardTurnsBToA != retailPartyRegardDuration {
		t.Fatalf("relation after party = %#v, want target-to-actor regard bonus for duration", relation)
	}
	if relation.RegardBonusAToB != 0 || relation.RegardTurnsAToB != 0 {
		t.Fatalf("party should not improve actor-to-target regard: %#v", relation)
	}
	if state.Units[0].Movement != 0 {
		t.Fatalf("diplomat movement = %d, want spent", state.Units[0].Movement)
	}
}

func TestThrowPartyRequiresGoldAndTargetCooldown(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].Gold = retailPartyCost - 1
	state.Cities = []City{{ID: "foreign", OwnerID: "player-b", TileID: "tile-b"}}
	state.Units = []Unit{{ID: "diplomat", OwnerID: "player-a", TileID: "tile-a", Movement: 1, Effects: []Effect{{ID: "party", Kind: EffectThrowParty}}}}
	payload := mustPayload(t, SpecialActionPayload{UnitID: "diplomat", CityID: "foreign"})

	if _, _, err := applySpecialCommand(&state, Command{Type: CommandThrowParty, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), nil); err == nil {
		t.Fatal("throw party succeeded without required gold")
	}
	state.Nations[0].Gold = retailPartyCost
	state.Nations[1].LastPartyTurn = state.Phase.Turn
	if _, _, err := applySpecialCommand(&state, Command{Type: CommandThrowParty, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), nil); err == nil {
		t.Fatal("throw party succeeded while target diplomacy was on party cooldown")
	}
}
