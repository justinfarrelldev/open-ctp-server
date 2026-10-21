package visibility

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestClientCityArtUsesObservedAgeAndCapitalNotFreshHiddenChanges(t *testing.T) {
	state := clientMetadataState()
	state.Nations = []engine.Nation{{ID: "b", CurrentAgeID: "AGE_ANCIENT"}}
	state.Cities[0].Buildings = []engine.Building{{ID: "capitol", Effects: []engine.Effect{{ID: "capitol", Kind: engine.EffectCapitol}}}}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	state.Units[0].TileID = "t10"
	state.Nations[0].CurrentAgeID = "AGE_MODERN"
	state.Cities[0].Buildings = nil
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	city := metadataWire(t, projection.Cities[0])
	art, ok := city["art"].(map[string]any)
	if !ok || art["age_id"] != "AGE_ANCIENT" || art["capital"] != true {
		t.Fatalf("remembered art was absent or refreshed from hidden state: %v", city)
	}
	state.Units[0].TileID = "t0"
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	art, ok = metadataWire(t, projection.Cities[0])["art"].(map[string]any)
	if !ok || art["age_id"] != "AGE_MODERN" || art["capital"] != false {
		t.Fatalf("re-observation did not refresh art: %v", art)
	}
}

func TestClientDistinctOrdersAreAdvertisedAndSleepIsPrivate(t *testing.T) {
	state := clientMetadataState()
	pin, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = engine.RulesPin{Version: pin.CompatibilityVersion, ContentHash: pin.ContentHash, Definitions: pin.CanonicalJSON}
	state.Cities = []engine.City{{ID: "home", Name: "Home", OwnerID: "a", TileID: "t0", Population: 2}, {ID: "other", Name: "Other", OwnerID: "a", TileID: "t2", Population: 2}}
	state.Units[0].Asleep = true
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{engine.CommandRenameCity, engine.CommandDisbandCity, engine.CommandWake} {
		if !hasAction(projection.LegalActions, kind) {
			t.Errorf("distinct command not advertised: %s", kind)
		}
	}
	if !projection.Units[0].Asleep {
		t.Fatal("owner sleep state lost")
	}
	state.Units = append(state.Units, engine.Unit{ID: "observer", OwnerID: "b", TileID: "t1"})
	foreign, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	for _, unit := range foreign.Units {
		if unit.ID == "scout" && unit.Asleep {
			t.Fatal("foreign private sleep order disclosed")
		}
	}
}
