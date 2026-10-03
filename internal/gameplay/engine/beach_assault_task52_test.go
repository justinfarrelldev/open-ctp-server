package engine

import "testing"

func TestRetailDisembarkZeroesMovementForNonBeachAssaultCargo(t *testing.T) {
	state := task8State()
	next, _, err := applyTask8(t, state, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}

	next, _, err = applyTask8(t, next, CommandDisembark, map[string]string{"unit_id": "cargo", "tile_id": "e"})
	if err != nil {
		t.Fatal(err)
	}
	cargo, ok := unitByID(next, "cargo")
	if !ok {
		t.Fatalf("cargo missing after disembark: %#v", next.Units)
	}
	if cargo.Movement != 0 {
		t.Fatalf("non-beach-assault cargo movement after disembark = %d, want 0", cargo.Movement)
	}
}

func TestRetailBeachAssaultCargoKeepsPostMoveMovementAfterSeaDisembark(t *testing.T) {
	state := task8State()
	for i := range state.Units {
		if state.Units[i].ID == "cargo" {
			state.Units[i].Effects = append(state.Units[i].Effects, Effect{ID: "beach", Kind: EffectCanBeachAssault})
		}
	}
	next, _, err := applyTask8(t, state, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}

	next, _, err = applyTask8(t, next, CommandDisembark, map[string]string{"unit_id": "cargo", "tile_id": "e"})
	if err != nil {
		t.Fatal(err)
	}
	cargo, ok := unitByID(next, "cargo")
	if !ok {
		t.Fatalf("cargo missing after disembark: %#v", next.Units)
	}
	if cargo.Movement != 2 {
		t.Fatalf("beach-assault cargo movement after disembark = %d, want 2", cargo.Movement)
	}
}
