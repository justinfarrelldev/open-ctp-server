package engine

import "testing"

func TestRetailRangedAttacksPreferSameRowTarget(t *testing.T) {
	state := fixtureState()
	frontRow := &Unit{ID: "front", OwnerID: "player-a", Ranged: 0, Attack: 0, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10}
	ranged := &Unit{ID: "ranged", OwnerID: "player-a", Ranged: 100, Strength: 10, Armor: 1, Firepower: 1, Health: 10, MaxHealth: 10}
	otherRow := &Unit{ID: "other-row", OwnerID: "player-b", Defense: 0, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10}
	sameRow := &Unit{ID: "same-row", OwnerID: "player-b", Defense: 0, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10}

	combatStrikeWithRuntime(&state, fixtureSources(), []*Unit{frontRow, ranged}, []*Unit{otherRow, sameRow}, true, &CombatResult{}, nil)
	if sameRow.Health >= 10 || otherRow.Health != 10 {
		t.Fatalf("ranged row targeting damaged wrong units: same=%d other=%d", sameRow.Health, otherRow.Health)
	}
}
