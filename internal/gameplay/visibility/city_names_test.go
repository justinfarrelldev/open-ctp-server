package visibility

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestFoundCitySuggestionNeverContainsEntityIDOrForeignName(t *testing.T) {
	p := Projection{Tiles: []TileProjection{{ID: "t0"}}, Units: []UnitProjection{{ID: "account-123-unit-settler-technical", OwnerID: "a", TileID: "t0", Effects: []engine.Effect{{ID: "settler", Kind: engine.EffectSettler}}}}, Cities: []CityProjection{{ID: "foreign", OwnerID: "b", Name: "City 1"}}}
	for _, action := range unitActions(p, "a") {
		if action.Command == engine.CommandFoundCity {
			var payload engine.FoundCityPayload
			if err := json.Unmarshal(action.Payload, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Name != "City 1" || payload.UnitID != p.Units[0].ID || payload.CityID == "" {
				t.Fatalf("bad display/identity separation: %+v", payload)
			}
			return
		}
	}
	t.Fatal("found action missing")
}
