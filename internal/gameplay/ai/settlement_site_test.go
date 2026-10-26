package ai

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestPlannerSettlementMatchesViewerSitePolicy(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*visibility.Projection)
		allowed bool
	}{
		{"foreign national", func(*visibility.Projection) {}, true},
		{"own influence", func(p *visibility.Projection) {
			id, owner := "city", "ai"
			p.Tiles[0].InfluenceCityID, p.Tiles[0].OwnerID = &id, &owner
		}, false},
		{"foreign influence", func(p *visibility.Projection) { id := "city"; p.Tiles[0].InfluenceCityID = &id }, false},
		{"adjacent foreign city", func(p *visibility.Projection) {
			p.Cities = []visibility.CityProjection{{ID: "city", OwnerID: "other", TileID: "near"}}
		}, false},
		{"occupied", func(p *visibility.Projection) {
			p.Cities = []visibility.CityProjection{{ID: "city", OwnerID: "other", TileID: "site"}}
		}, false},
		{"stale site", func(p *visibility.Projection) { p.Tiles[0].Stale = true }, false},
		{"missing site", func(p *visibility.Projection) { p.Tiles = nil }, false},
		{"stale adjacent city", func(p *visibility.Projection) {
			p.Cities = []visibility.CityProjection{{ID: "city", OwnerID: "other", TileID: "near", Stale: true}}
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			foreign := "other"
			p := visibility.Projection{
				Tiles: []visibility.TileProjection{{ID: "site", OwnerID: &foreign}, {ID: "near", X: 1, Y: 1}},
				Units: []visibility.UnitProjection{{ID: "settler", OwnerID: "ai", TileID: "site", Effects: []engine.Effect{{Kind: engine.EffectSettler}}}},
				Rules: &visibility.RulesProjection{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER"}}},
			}
			tc.mutate(&p)
			raw, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			candidates, err := Plan(raw, "ai")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, c := range candidates {
				found = found || c.Type == engine.CommandFoundCity
			}
			if found != tc.allowed {
				t.Fatalf("AI founding=%v, want %v: %+v", found, tc.allowed, candidates)
			}
		})
	}
}
