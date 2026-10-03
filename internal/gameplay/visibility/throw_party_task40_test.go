package visibility

import (
	"bytes"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestThrowPartyLegalActionUsesPublicGoldAndDiplomacy(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "home", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "foreign-tile", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Nations = []engine.Nation{{ID: "a", Gold: 100}, {ID: "b"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "peace"}}
	state.Cities = []engine.City{{ID: "foreign", OwnerID: "b", TileID: "foreign-tile"}}
	state.Units = []engine.Unit{{ID: "diplomat", OwnerID: "a", TileID: "home", Movement: 1, Effects: []engine.Effect{{ID: "party", Kind: engine.EffectThrowParty}}}}
	var err error
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandThrowParty && action.EntityID == "diplomat" && action.TargetID == "foreign" && bytes.Contains(action.Payload, []byte(`"city_id":"foreign"`)) {
			return
		}
	}
	t.Fatalf("throw-party action missing from %#v", projection.LegalActions)
}

func TestThrowPartyRegardProjectionIsDirectional(t *testing.T) {
	state := fixtureState()
	state.Nations = []engine.Nation{{ID: "a"}, {ID: "b"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, RegardBonusBToA: 10, RegardTurnsBToA: 10}}
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
	if len(projectionA.Diplomacy.Relations) != 1 || projectionA.Diplomacy.Relations[0].Regard != "neutral" {
		t.Fatalf("actor projection regard = %#v, want neutral", projectionA.Diplomacy.Relations)
	}
	if len(projectionB.Diplomacy.Relations) != 1 || projectionB.Diplomacy.Relations[0].Regard != "friendly" {
		t.Fatalf("target projection regard = %#v, want friendly", projectionB.Diplomacy.Relations)
	}
}

func TestThrowPartyLegalActionHiddenWhenGoldOrDiplomacyDisallowIt(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "home", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "foreign-tile", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Nations = []engine.Nation{{ID: "a", Gold: 99}, {ID: "b"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "peace"}}
	state.Cities = []engine.City{{ID: "foreign", OwnerID: "b", TileID: "foreign-tile"}}
	state.Units = []engine.Unit{{ID: "diplomat", OwnerID: "a", TileID: "home", Movement: 1, Effects: []engine.Effect{{ID: "party", Kind: engine.EffectThrowParty}}}}
	var err error
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if hasAction(projection.LegalActions, engine.CommandThrowParty) {
		t.Fatalf("throw-party action exposed without public gold: %#v", projection.LegalActions)
	}

	state.Nations[0].Gold = 100
	state.Policy.Wars = []engine.TurnPolicyWar{{A: "a", B: "b"}}
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if hasAction(projection.LegalActions, engine.CommandThrowParty) {
		t.Fatalf("throw-party action exposed during hot war: %#v", projection.LegalActions)
	}
}
