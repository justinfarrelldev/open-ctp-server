package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestProductionCompletionCopiesRuntimeRangedStats(t *testing.T) {
	state := State{Phase: PhaseState{Turn: 7}}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_RANGED", Domain: DomainLand, Movement: 1, Attack: 3, Defense: 2, Ranged: 9, BombardRange: 2, BombRounds: 4, ActiveDefenseRange: 3, Strength: 3, Armor: 1, Health: 10, Upkeep: 4}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}
	if err := completeUnitProduction(&state, &city, ProductionItem{ID: "UNIT_RANGED", Kind: ProductionUnit}, runtime); err != nil {
		t.Fatalf("completeUnitProduction() error = %v", err)
	}
	if len(state.Units) != 1 || state.Units[0].Movement != 1 || state.Units[0].MaxMovement != 1 || state.Units[0].Ranged != 9 || state.Units[0].BombardRange != 2 || state.Units[0].BombRounds != 4 || state.Units[0].ActiveDefenseRange != 3 || state.Units[0].Upkeep != 4 {
		t.Fatalf("produced ranged unit = %#v", state.Units)
	}
}

func TestRetailPopulationRemovingUnitsReduceCityPopulation(t *testing.T) {
	state := State{Phase: PhaseState{Turn: 7}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "pop", Kind: EffectBuildingRemovesPopulation}}}}}

	if err := completeUnitProduction(&state, &city, ProductionItem{ID: "UNIT_SETTLER", Kind: ProductionUnit}, runtime); err != nil {
		t.Fatalf("completeUnitProduction() error = %v", err)
	}
	if got := city.Population; got != 1 {
		t.Fatalf("city population after settler = %d, want 1", got)
	}
}

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
