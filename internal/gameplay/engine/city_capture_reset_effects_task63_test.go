package engine

import "testing"

func TestRetailCityCaptureClearsOwnerSpecificCityEffects(t *testing.T) {
	state := fixtureState()
	state.Cities = []City{{ID: "target", OwnerID: "player-b", TileID: "tile-b", Population: 4, Effects: []Effect{
		{ID: "franchise:player-a", Kind: EffectFranchise},
		{ID: "injunction", Kind: EffectInjunction, Value: 1},
		{ID: "watchful", Kind: EffectWatchful, Value: 4},
		{ID: "happiness-attack", Kind: EffectHappiness, Value: -5, Turns: 2},
		{ID: "base-happiness", Kind: EffectHappiness, Value: 75},
	}}}

	captureCityAt(&state, "player-a", "tile-b")
	city := state.Cities[0]
	if hasEffect(city.Effects, EffectFranchise) || hasEffect(city.Effects, EffectInjunction) || hasEffect(city.Effects, EffectWatchful) || hasTimedHappinessAttack(city, -5, 2) {
		t.Fatalf("capture retained owner-specific/timed effects: %+v", city.Effects)
	}
	if !hasEffect(city.Effects, EffectHappiness) {
		t.Fatalf("capture removed non-timed base happiness effect: %+v", city.Effects)
	}
}
