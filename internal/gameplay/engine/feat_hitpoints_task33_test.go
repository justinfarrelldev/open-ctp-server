package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestNationFeatIncreaseHitPointsAppliesToProducedUnits(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].Effects = []Effect{{ID: "feat:conquer", Kind: EffectIncreaseHP, Value: 1, Turns: 25}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", StoredProduction: 1, Queue: []ProductionItem{{ID: "UNIT_WARRIOR", Kind: ProductionUnit, Cost: 1}}}}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_WARRIOR", Domain: DomainLand, Movement: 100, Health: 10}}}

	if err := completeProduction(&state, &state.Cities[0], runtime); err != nil {
		t.Fatalf("completeProduction: %v", err)
	}
	produced, ok := unitByID(state, "unit:city-a:UNIT_WARRIOR:1")
	if !ok {
		t.Fatalf("produced units = %#v, want UNIT_WARRIOR", state.Units)
	}
	if produced.Health != 11 || produced.MaxHealth != 11 {
		t.Fatalf("produced unit hp = %d/%d, want 11/11 from active hit-point feat", produced.Health, produced.MaxHealth)
	}
}
