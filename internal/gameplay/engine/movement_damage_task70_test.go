package engine

import "testing"

func TestRetailLossMoveToDmgTwoHasMinimumTwoMovement(t *testing.T) {
	unit := Unit{ID: "damaged", OwnerID: "player-a", Domain: DomainLand, MaxMovement: 6, Health: 1, MaxHealth: 6, Effects: []Effect{{ID: "loss-two", Kind: EffectLossMoveToDmgTwo}}}
	if got := refreshedUnitMovement(fixtureState(), unit, 0); got != 2 {
		t.Fatalf("refreshed movement = %d, want minimum 2 for LossMoveToDmgTwo", got)
	}
}
