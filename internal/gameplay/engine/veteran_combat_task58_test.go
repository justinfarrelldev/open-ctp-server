package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailVeteranCoefficientIncreasesAttackAndBombardStrength(t *testing.T) {
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{VeteranCoefPercent: 50}}
	unit := Unit{ID: "vet", Attack: 10, Ranged: 8, Veteran: true}
	if got := combatAttackValueWithRuntime(runtime, unit, false); got != 15 {
		t.Fatalf("veteran melee attack = %d, want 15", got)
	}
	if got := combatAttackValueWithRuntime(runtime, unit, true); got != 12 {
		t.Fatalf("veteran ranged attack = %d, want 12", got)
	}
	if got := combatVeteranAdjustedAttackWithRuntime(runtime, unit, 20); got != 30 {
		t.Fatalf("veteran adjusted modified attack = %d, want coefficient to scale full attack strength", got)
	}
}

func TestRetailVeteranCoefficientDefaultsToReleasedConstant(t *testing.T) {
	unit := Unit{ID: "vet", Attack: 10, Veteran: true}
	if got := combatAttackValueWithRuntime(nil, unit, false); got != 15 {
		t.Fatalf("default veteran attack = %d, want retail 50%% bonus => 15", got)
	}
}

func TestRetailVeteranCoefficientChangesCombatHitChance(t *testing.T) {
	state := fixtureState()
	state.RNG = []RNGDomainState{{Domain: RNGDomainCombat, State: 1}}
	state.Units = []Unit{
		{ID: "attacker", OwnerID: "player-a", Attack: 10, Strength: 1, Firepower: 1, Health: 3, MaxHealth: 3, Veteran: true},
		{ID: "defender", OwnerID: "player-b", Defense: 10, Armor: 1, Health: 3, MaxHealth: 3},
	}
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{VeteranCoefPercent: 50}}
	result := &CombatResult{}
	combatStrikeWithRuntime(&state, Sources{RNG: fixedValueRNG{value: 599}}, []*Unit{&state.Units[0]}, []*Unit{&state.Units[1]}, false, result, runtime)
	if state.Units[1].Health != 2 {
		t.Fatalf("veteran strike target health = %d, want one hit at retail 15/(15+10) odds", state.Units[1].Health)
	}
}

type fixedValueRNG struct{ value uint64 }

func (r fixedValueRNG) Next(domain string, state uint64) (uint64, uint64) { return state + 1, r.value }
