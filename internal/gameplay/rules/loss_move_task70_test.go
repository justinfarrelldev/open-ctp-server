package rules

import "testing"

func TestRetailUnitLossMoveToDmgTwoProjectsMarker(t *testing.T) {
	effects := retailUnitEffects(RetailUnit{ID: "UNIT_TEST", Flags: []string{"LossMoveToDmgTwo"}})
	for _, effect := range effects {
		if effect.Kind == "loss_move_to_dmg_two" {
			return
		}
	}
	t.Fatalf("effects = %#v, want loss_move_to_dmg_two marker", effects)
}
