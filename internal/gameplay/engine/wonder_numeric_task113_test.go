package engine

import "testing"

func TestRetailNumericWonderEffectsRespectDestructionAndObsolescence(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_OBSOLETE"}}},
		Wonders: []Wonder{{ID: "WONDER_NUMERIC", OwnerID: "player-a", ObsoletedBy: "ADVANCE_OBSOLETE", Effects: []Effect{{ID: "bonus", Kind: EffectBonusGold, Value: 100}}}},
	}

	if got := ownerWonderEffectSum(state, "player-a", EffectBonusGold); got != 0 {
		t.Fatalf("obsolete wonder bonus = %d, want 0", got)
	}
	state.Nations[0].Advances = nil
	state.Wonders[0].Destroyed = true
	if got := ownerWonderEffectSum(state, "player-a", EffectBonusGold); got != 0 {
		t.Fatalf("destroyed wonder bonus = %d, want 0", got)
	}
	state.Wonders[0].Destroyed = false
	if got := ownerWonderEffectSum(state, "player-a", EffectBonusGold); got != 100 {
		t.Fatalf("active wonder bonus = %d, want 100", got)
	}
}
