package engine

import "testing"

func TestRetailNonFlankerMeleeDoesNotRetargetEmptyRows(t *testing.T) {
	state := fixtureState()
	attacker := &Unit{ID: "attacker", OwnerID: "player-a", Attack: 100, Strength: 10, Armor: 1, Firepower: 1, Health: 10, MaxHealth: 10}
	deadSameRow := &Unit{ID: "dead-row", OwnerID: "player-b", Defense: 0, Strength: 1, Armor: 1, Health: 0, MaxHealth: 10}
	otherRow := &Unit{ID: "other-row", OwnerID: "player-b", Defense: 0, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10}

	combatStrikeWithRuntime(&state, fixtureSources(), []*Unit{attacker}, []*Unit{deadSameRow, otherRow}, false, &CombatResult{}, nil)
	if otherRow.Health != 10 {
		t.Fatalf("non-flanker retargeted an empty row: health=%d", otherRow.Health)
	}
}

func TestRetailFlankerMeleeTargetsNearestOccupiedRow(t *testing.T) {
	state := fixtureState()
	attacker := &Unit{ID: "flanker", OwnerID: "player-a", Attack: 100, Strength: 10, Armor: 1, Firepower: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "flanker", Kind: EffectIsFlanker}}}
	deadSameRow := &Unit{ID: "dead-row", OwnerID: "player-b", Defense: 0, Strength: 1, Armor: 1, Health: 0, MaxHealth: 10}
	otherRow := &Unit{ID: "other-row", OwnerID: "player-b", Defense: 0, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10}

	combatStrikeWithRuntime(&state, fixtureSources(), []*Unit{attacker}, []*Unit{deadSameRow, otherRow}, false, &CombatResult{}, nil)
	if otherRow.Health >= 10 {
		t.Fatalf("flanker did not attack nearest occupied row: health=%d", otherRow.Health)
	}
}
