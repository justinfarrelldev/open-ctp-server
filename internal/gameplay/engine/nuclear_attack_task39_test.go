package engine

import "testing"

func TestNuclearAttackUnitNukesForeignCityAndIsConsumed(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "target", OwnerID: "player-b", TileID: "tile-b", Population: 6}}
	state.Units = []Unit{
		{ID: "missile", OwnerID: "player-a", TileID: "tile-a", Domain: DomainAir, Movement: 100, Health: 1, MaxHealth: 1, Effects: []Effect{{ID: "nuke", Kind: EffectNuclearAttack}}},
		{ID: "defender", OwnerID: "player-b", TileID: "tile-b", Domain: DomainLand, Health: 10, MaxHealth: 10},
	}

	eventType, err := applyUnitCommand(&state, Command{Type: CommandNuclearAttack, ActorPlayerID: "player-a", Payload: mustPayload(t, nuclearAttackPayload{UnitID: "missile", TargetTileID: "tile-b"})}, nil)
	if err != nil {
		t.Fatalf("nuclear attack: %v", err)
	}
	if eventType != EventNuclearAttack {
		t.Fatalf("event type = %q, want %q", eventType, EventNuclearAttack)
	}
	if _, ok := unitByID(state, "missile"); ok {
		t.Fatalf("nuclear missile survived attack: %#v", state.Units)
	}
	if city, ok := cityByID(state, "target"); !ok || city.Population >= 6 {
		t.Fatalf("target city after nuke = %#v ok=%v, want damaged city", city, ok)
	}
	if _, ok := unitByID(state, "defender"); ok {
		t.Fatalf("defender on nuked tile survived: %#v", state.Units)
	}
}

func TestNuclearAttackProtectionConsumesMissileWithoutDamagingCity(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "target", OwnerID: "player-b", TileID: "tile-b", Population: 6, Buildings: []Building{{ID: "shield", Effects: []Effect{{ID: "nuke-shield", Kind: EffectProtectFromNukes}}}}}}
	state.Units = []Unit{{ID: "missile", OwnerID: "player-a", TileID: "tile-a", Domain: DomainAir, Movement: 100, Health: 1, MaxHealth: 1, Effects: []Effect{{ID: "nuke", Kind: EffectNuclearAttack}}}}

	eventType, err := applyUnitCommand(&state, Command{Type: CommandNuclearAttack, ActorPlayerID: "player-a", Payload: mustPayload(t, nuclearAttackPayload{UnitID: "missile", TargetTileID: "tile-b"})}, nil)
	if err != nil {
		t.Fatalf("nuclear attack: %v", err)
	}
	if eventType != EventNuclearAttackFailed {
		t.Fatalf("event type = %q, want %q", eventType, EventNuclearAttackFailed)
	}
	if _, ok := unitByID(state, "missile"); ok {
		t.Fatalf("protected nuclear attack did not consume missile: %#v", state.Units)
	}
	if city, ok := cityByID(state, "target"); !ok || city.Population != 6 {
		t.Fatalf("protected city after nuke = %#v ok=%v, want unchanged population 6", city, ok)
	}
}
