package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func settlementProjection() Projection {
	foreign := "b"
	return Projection{
		Tiles: []TileProjection{{ID: "site", Terrain: engine.TerrainPlains, OwnerID: &foreign}, {ID: "near", X: 1, Y: 1}, {ID: "far", X: 3}},
		Units: []UnitProjection{{ID: "settler", OwnerID: "a", TileID: "site", Domain: engine.DomainLand, Movement: 100, Effects: []engine.Effect{{Kind: engine.EffectSettler}}}},
	}
}

func settlementActions(p Projection) []ActionProjection {
	var result []ActionProjection
	for _, a := range unitActions(p, "a") {
		if a.Command == engine.CommandFoundCity {
			result = append(result, a)
		}
	}
	return result
}

func TestSettlementActionsUseCurrentViewerSiteFacts(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*Projection)
		allowed bool
	}{
		{"foreign national", func(*Projection) {}, true},
		{"own national", func(p *Projection) { owner := "a"; p.Tiles[0].OwnerID = &owner }, true},
		{"neutral", func(p *Projection) { p.Tiles[0].OwnerID = nil }, true},
		{"foreign influence", func(p *Projection) { id := "foreign-city"; p.Tiles[0].InfluenceCityID = &id }, false},
		{"own influence", func(p *Projection) {
			id, owner := "own-city", "a"
			p.Tiles[0].InfluenceCityID, p.Tiles[0].OwnerID = &id, &owner
		}, false},
		{"occupied", func(p *Projection) { p.Cities = []CityProjection{{ID: "city", OwnerID: "b", TileID: "site"}} }, false},
		{"adjacent", func(p *Projection) { p.Cities = []CityProjection{{ID: "city", OwnerID: "b", TileID: "near"}} }, false},
		{"distant", func(p *Projection) { p.Cities = []CityProjection{{ID: "city", OwnerID: "b", TileID: "far"}} }, true},
		{"stale site", func(p *Projection) { p.Tiles[0].Stale = true }, false},
		{"missing site", func(p *Projection) { p.Tiles = nil }, false},
		{"stale adjacent city memory", func(p *Projection) {
			p.Cities = []CityProjection{{ID: "city", OwnerID: "b", TileID: "near", Stale: true}}
		}, true},
		{"foreign settler", func(p *Projection) { p.Units[0].OwnerID = "b" }, false},
		{"nonsettler", func(p *Projection) { p.Units[0].Effects = nil }, false},
		{"army member", func(p *Projection) { p.Units[0].ArmyID = "army" }, false},
		{"transported", func(p *Projection) { p.Units[0].TransportID = "transport" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := settlementProjection()
			tc.mutate(&p)
			if got := len(settlementActions(p)) > 0; got != tc.allowed {
				t.Fatalf("advertised=%v, want %v", got, tc.allowed)
			}
		})
	}
}

func TestSettlementAdvertisedForeignSitePayloadExecutes(t *testing.T) {
	state := fixtureState()
	state.Cities = nil
	state.Units = []engine.Unit{{ID: "settler", OwnerID: "a", TileID: "0,0", Domain: engine.DomainLand, Movement: 100, Effects: []engine.Effect{{Kind: engine.EffectSettler}}}}
	state.Tiles[0].OwnerID = "b"
	p, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	var advertised *ActionProjection
	for _, a := range p.LegalActions {
		if a.Command == engine.CommandFoundCity {
			copy := a
			advertised = &copy
			break
		}
	}
	if advertised == nil {
		t.Fatal("foreign empty-influence site has no founding action")
	}
	next, _, _, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: "found-advertised", ActorPlayerID: "a", Type: advertised.Command, Payload: advertised.Payload}, state.Rules, engine.Sources{Clock: topologyClock{}, IDs: &topologyIDs{}, RNG: topologyRNG{}})
	if err != nil {
		t.Fatalf("server rejected its advertised payload: %v", err)
	}
	var payload engine.FoundCityPayload
	if err := json.Unmarshal(advertised.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(next.Cities) != 1 || len(next.Units) != 0 || next.Cities[0].ID != payload.CityID || next.Cities[0].Name != payload.Name {
		t.Fatalf("advertised city identity not preserved: %+v", next.Cities)
	}
}

func TestSettlementActionFilteringDoesNotReadHiddenCityState(t *testing.T) {
	// Both worlds project the same current site and no foreign city. Changing an
	// unseen city must not change the action list or disclose its identity.
	base := fixtureState()
	base.Units = []engine.Unit{{ID: "settler", OwnerID: "a", TileID: "0,0", Domain: engine.DomainLand, Effects: []engine.Effect{{Kind: engine.EffectSettler}}}}
	base.Tiles = append(base.Tiles, engine.Tile{ID: "far", X: 10, Terrain: engine.TerrainPlains, Open: true})
	base.Cities = []engine.City{{ID: "HIDDEN_CITY_TOKEN", OwnerID: "b", TileID: "far"}}
	withCity, err := Project(base, "a")
	if err != nil {
		t.Fatal(err)
	}
	base.Cities = nil
	withoutCity, err := Project(base, "a")
	if err != nil {
		t.Fatal(err)
	}
	left, _ := json.Marshal(withCity.LegalActions)
	right, _ := json.Marshal(withoutCity.LegalActions)
	if !bytes.Equal(left, right) || bytes.Contains(left, []byte("HIDDEN_CITY_TOKEN")) {
		t.Fatalf("hidden city affected actions: %s / %s", left, right)
	}
}
