package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestFoundCityRecordsRetailFounder(t *testing.T) {
	state := runtimeStartedState(t)
	state.Units = []Unit{{ID: "settler", OwnerID: "player-a", TileID: "tile-a", Effects: []Effect{{ID: "settle", Kind: EffectSettler}}}}

	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler", CityID: "city-a", Name: "Alpha"}); err != nil {
		t.Fatalf("foundCity: %v", err)
	}
	if got := state.Cities[0].FounderID; got != "player-a" {
		t.Fatalf("founder id = %q, want founding player", got)
	}
}

func TestCityRecapturedFeatAppliesWhenFounderRetakesCity(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-b", FounderID: "player-a", TileID: "tile-b", Population: 1}}
	runtime := &rules.RuntimeRules{Feats: []rules.RuntimeFeatDefinition{{ID: "FEAT_CITY_RECAPTURED", Duration: 10, Effects: []rules.RuntimeStarterEffect{{ID: "happy", Kind: EffectHappiness, Value: 1}}}}}

	captured := captureCityAt(&state, "player-a", "tile-b", runtime)
	if captured != "city-a" || state.Cities[0].OwnerID != "player-a" {
		t.Fatalf("captured=%q city=%#v, want recaptured city owned by founder", captured, state.Cities[0])
	}
	if !hasEffectValueTurns(state.Nations[0].Effects, EffectHappiness, 1, 10) {
		t.Fatalf("founder nation effects after recapture = %#v, want FEAT_CITY_RECAPTURED happiness", state.Nations[0].Effects)
	}
}

func TestCityRecapturedFeatDoesNotApplyToNonFounderCapture(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-b", FounderID: "player-b", TileID: "tile-b", Population: 1}}
	runtime := &rules.RuntimeRules{Feats: []rules.RuntimeFeatDefinition{{ID: "FEAT_CITY_RECAPTURED", Duration: 10, Effects: []rules.RuntimeStarterEffect{{ID: "happy", Kind: EffectHappiness, Value: 1}}}}}

	captureCityAt(&state, "player-a", "tile-b", runtime)
	if hasEffectValueTurns(state.Nations[0].Effects, EffectHappiness, 1, 10) {
		t.Fatalf("recapture feat applied for non-founder capture: %#v", state.Nations[0].Effects)
	}
}
