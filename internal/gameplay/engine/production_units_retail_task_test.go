package engine

import "testing"

func TestProductionCompletionCreatesRuntimeUnitAtCity(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Queue: []ProductionItem{{ID: "UNIT_SYNTHETIC", Kind: ProductionUnit, Cost: 1}}}}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatalf("RecalculateRound() error = %v", err)
	}
	if len(next.Cities[0].Queue) != 0 || next.Cities[0].StoredProduction != 2 {
		t.Fatalf("city production state = %#v", next.Cities[0])
	}
	if len(next.Units) != 1 {
		t.Fatalf("units = %#v, want one produced unit", next.Units)
	}
	unit := next.Units[0]
	if unit.ID != "unit:city-a:UNIT_SYNTHETIC:1" || unit.OwnerID != "player-a" || unit.TileID != "tile-a" {
		t.Fatalf("produced unit identity/location = %+v", unit)
	}
	if unit.Domain != DomainLand || unit.Movement != 2 || unit.Attack != 4 || unit.Defense != 4 || unit.MaxHealth != 10 || unit.Health != 10 || unit.VisibilityRange != 2 {
		t.Fatalf("produced unit stats = %+v", unit)
	}
	if !hasEffect(unit.Effects, EffectSettler) || !hasEffect(unit.Effects, EffectSpecialEmbassy) {
		t.Fatalf("produced unit effects = %#v", unit.Effects)
	}
}
