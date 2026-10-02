package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestConqueredByForceFeatAppliesWhenLargeNationLosesFinalCity(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "last-city", OwnerID: "player-b", TileID: "tile-b", Population: 1}}
	state.Nations[1].MaxCityCount = 5
	runtime := &rules.RuntimeRules{Feats: []rules.RuntimeFeatDefinition{{ID: "FEAT_CONQUERED_BY_FORCE", Duration: 25, MinimumCivSize: 5, Effects: []rules.RuntimeStarterEffect{{ID: "hp", Kind: EffectIncreaseHP, Value: 1}}}}}

	captured := captureCityAt(&state, "player-a", "tile-b", runtime)
	if captured != "last-city" || state.Cities[0].OwnerID != "player-a" {
		t.Fatalf("captured=%q city=%#v, want final city owned by attacker", captured, state.Cities[0])
	}
	if !hasEffectValueTurns(state.Nations[0].Effects, EffectIncreaseHP, 1, 25) {
		t.Fatalf("attacker nation effects = %#v, want conquered-by-force hit point feat", state.Nations[0].Effects)
	}
}

func TestConqueredByForceFeatRequiresFinalCityAndMinimumCivSize(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-one", OwnerID: "player-b", TileID: "tile-b", Population: 1}, {ID: "city-two", OwnerID: "player-b", TileID: "tile-a", Population: 1}}
	state.Nations[1].MaxCityCount = 5
	runtime := &rules.RuntimeRules{Feats: []rules.RuntimeFeatDefinition{{ID: "FEAT_CONQUERED_BY_FORCE", Duration: 25, MinimumCivSize: 5, Effects: []rules.RuntimeStarterEffect{{ID: "hp", Kind: EffectIncreaseHP, Value: 1}}}}}

	captureCityAt(&state, "player-a", "tile-b", runtime)
	if hasEffectValueTurns(state.Nations[0].Effects, EffectIncreaseHP, 1, 25) {
		t.Fatalf("conquered-by-force feat applied before final city was captured: %#v", state.Nations[0].Effects)
	}

	state.Nations[1].MaxCityCount = 4
	captureCityAt(&state, "player-a", "tile-a", runtime)
	if hasEffectValueTurns(state.Nations[0].Effects, EffectIncreaseHP, 1, 25) {
		t.Fatalf("conquered-by-force feat applied below minimum civ size: %#v", state.Nations[0].Effects)
	}
}
