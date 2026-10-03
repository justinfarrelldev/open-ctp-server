package engine

import "testing"

func TestRetailReduceReadinessCostWonderStacksAndCaps(t *testing.T) {
	state := State{Wonders: []Wonder{
		{ID: "WONDER_ACTIVE_A", OwnerID: "player-a", Effects: []Effect{{ID: "reduce-a", Kind: EffectReduceReadinessCost, Value: 60}}},
		{ID: "WONDER_ACTIVE_B", OwnerID: "player-a", Effects: []Effect{{ID: "reduce-b", Kind: EffectReduceReadinessCost, Value: 50}}},
		{ID: "WONDER_DESTROYED", OwnerID: "player-a", Destroyed: true, Effects: []Effect{{ID: "reduce-destroyed", Kind: EffectReduceReadinessCost, Value: 100}}},
		{ID: "WONDER_FOREIGN", OwnerID: "player-b", Effects: []Effect{{ID: "reduce-foreign", Kind: EffectReduceReadinessCost, Value: 100}}},
	}}

	if got := readinessCostReduction(state, map[string]bool{"player-a": true}); got != 100 {
		t.Fatalf("readiness reduction = %d, want capped active owner sum 100", got)
	}
}
