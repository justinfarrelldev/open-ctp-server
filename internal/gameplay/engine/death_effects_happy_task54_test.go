package engine

import "testing"

func TestRetailCombatDeathWithoutDeathEffectsHappyDoesNotAffectPeaceMovement(t *testing.T) {
	state := combatState()
	state.Nations = append(state.Nations, Nation{ID: "nation-b"})
	state.Units = []Unit{
		{ID: "attacker", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Attack: 1000, Defense: 1, Strength: 10, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "defender", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Attack: 0, Defense: 0, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"attacker"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"defender"}}}

	next, _, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	nation := nationForID(&next, "nation-b")
	if nation == nil {
		t.Fatal("nation-b missing")
	}
	if nation.PeaceMovement != nil && (nation.PeaceMovement.UnitsLostHome != 0 || nation.PeaceMovement.UnitsLostOverseas != 0) {
		t.Fatalf("non DeathEffectsHappy combat casualty affected peace movement: %#v", nation.PeaceMovement)
	}
}

func TestRetailCombatDeathWithDeathEffectsHappyAffectsPeaceMovement(t *testing.T) {
	state := combatState()
	state.Nations = append(state.Nations, Nation{ID: "nation-b"})
	state.Units = []Unit{
		{ID: "attacker", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Attack: 1000, Defense: 1, Strength: 10, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "defender", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Attack: 0, Defense: 0, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1, Effects: []Effect{{ID: "happy", Kind: EffectDeathEffectsHappy}}},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"attacker"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"defender"}}}

	next, _, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	nation := nationForID(&next, "nation-b")
	if nation == nil || nation.PeaceMovement == nil || nation.PeaceMovement.UnitsLostHome+nation.PeaceMovement.UnitsLostOverseas != 1 {
		t.Fatalf("DeathEffectsHappy combat casualty did not affect peace movement: %#v", nation)
	}
}
