package visibility

import (
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestRetailCanPirateProjectsVisibleTouchedForeignTradeRoute(t *testing.T) {
	state := engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerOrder: []string{"player-a", "player-b"}},
		Players:       []engine.Player{{ID: "player-a", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "player-b", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "player-c", NationID: "nation-c", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}},
		Nations:       []engine.Nation{{ID: "nation-a"}, {ID: "nation-b"}, {ID: "nation-c"}},
		RNG:           []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
		Tiles:         []engine.Tile{{ID: "tile-b", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "tile-c", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}},
		Cities:        []engine.City{{ID: "city-b", OwnerID: "player-b", TileID: "tile-b"}, {ID: "city-c", OwnerID: "player-c", TileID: "tile-c"}},
		Units:         []engine.Unit{{ID: "privateer", OwnerID: "player-a", ArmyID: "privateers", TileID: "tile-b", Movement: 1, Effects: []engine.Effect{{ID: "pirate", Kind: engine.EffectCanPirate}}}},
		Armies:        []engine.Army{{ID: "privateers", OwnerID: "player-a", UnitIDs: []string{"privateer"}}},
		TradeRoutes:   []engine.TradeRoute{{ID: "route-b-c", SourceCityID: "city-b", DestinationCityID: "city-c", Gold: 10}},
		Fog:           []engine.FogObservation{{PlayerID: "player-a", TileID: "tile-b", Turn: 1}, {PlayerID: "player-a", TileID: "tile-c", Turn: 1}},
	}
	projection, err := Project(state, "player-a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandPirateTrade && strings.Contains(string(action.Payload), "privateer") && strings.Contains(string(action.Payload), "route-b-c") {
			return
		}
	}
	t.Fatalf("pirate-trade action missing: %#v", projection.LegalActions)
}
