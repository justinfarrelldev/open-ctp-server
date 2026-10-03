package visibility

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRetailIncreaseRegardWonderImprovesForeignRegardTowardOwner(t *testing.T) {
	state := fixtureState()
	state.Nations = []engine.Nation{{ID: "a"}, {ID: "b"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true}}
	state.Wonders = []engine.Wonder{{ID: "WONDER_RAMAYANA", OwnerID: "b", CityID: "city-b", Scope: "nation", Effects: []engine.Effect{{ID: "regard", Kind: engine.EffectIncreaseRegard, Value: 50}}}}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projectionA, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	projectionB, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(projectionA.Diplomacy.Relations) != 1 || projectionA.Diplomacy.Relations[0].Regard != "friendly" {
		t.Fatalf("foreigner regard toward wonder owner = %#v, want friendly", projectionA.Diplomacy.Relations)
	}
	if len(projectionB.Diplomacy.Relations) != 1 || projectionB.Diplomacy.Relations[0].Regard != "neutral" {
		t.Fatalf("wonder owner regard toward foreigner = %#v, want neutral", projectionB.Diplomacy.Relations)
	}
}
