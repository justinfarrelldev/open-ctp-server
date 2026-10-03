package engine

import (
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailCanPirateSetsPiratingArmyAndPaysWasteGold(t *testing.T) {
	state := State{
		Phase:       PhaseState{Mode: "classic", ActivePlayerOrder: []string{"player-a", "player-b"}},
		Players:     []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive}, {ID: "player-b", NationID: "nation-b", Lifecycle: PlayerActive}, {ID: "player-c", NationID: "nation-c", Lifecycle: PlayerActive}},
		Nations:     []Nation{{ID: "nation-a"}, {ID: "nation-b"}, {ID: "nation-c"}},
		Tiles:       []Tile{{ID: "tile-a", Open: true}, {ID: "tile-b", Open: true}},
		Cities:      []City{{ID: "city-b", OwnerID: "player-b", TileID: "tile-b"}, {ID: "city-c", OwnerID: "player-c", TileID: "tile-a"}},
		Units:       []Unit{{ID: "privateer", OwnerID: "player-a", ArmyID: "privateers", TileID: "tile-b", Movement: 1, Effects: []Effect{{ID: "pirate", Kind: EffectCanPirate}}}},
		Armies:      []Army{{ID: "privateers", OwnerID: "player-a", UnitIDs: []string{"privateer"}}},
		TradeRoutes: []TradeRoute{{ID: "route-b-c", SourceCityID: "city-b", DestinationCityID: "city-c", Gold: 10}},
	}
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{PiracyWastePercent: 80}}

	eventType, err := applyUnitCommand(&state, Command{Type: CommandPirateTrade, ActorPlayerID: "player-a", Payload: mustPayload(t, map[string]string{"unit_id": "privateer", "route_id": "route-b-c"})}, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if eventType != EventTradeRoutePirated || state.TradeRoutes[0].PiratingArmyID != "privateers" || state.Units[0].Movement != 0 {
		t.Fatalf("piracy command outcome event=%q route=%#v unit=%#v", eventType, state.TradeRoutes[0], state.Units[0])
	}

	advanceTreasury(&state, runtime)
	if got := state.Nations[0].Gold; got != 8 {
		t.Fatalf("piracy treasury = %d, want route value 10 * 80%% waste coefficient", got)
	}
}

func TestRetailCanPirateRejectsOwnRoutesAndNonPirates(t *testing.T) {
	state := State{
		Phase:       PhaseState{Mode: "classic", ActivePlayerOrder: []string{"player-a"}},
		Players:     []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive}},
		Nations:     []Nation{{ID: "nation-a"}},
		Tiles:       []Tile{{ID: "tile-a", Open: true}, {ID: "tile-b", Open: true}},
		Cities:      []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}, {ID: "city-b", OwnerID: "player-a", TileID: "tile-b"}},
		Units:       []Unit{{ID: "ship", OwnerID: "player-a", ArmyID: "ships", TileID: "tile-a", Movement: 1}},
		Armies:      []Army{{ID: "ships", OwnerID: "player-a", UnitIDs: []string{"ship"}}},
		TradeRoutes: []TradeRoute{{ID: "own-route", SourceCityID: "city-a", DestinationCityID: "city-b", Gold: 10}},
	}

	_, err := applyUnitCommand(&state, Command{Type: CommandPirateTrade, ActorPlayerID: "player-a", Payload: mustPayload(t, map[string]string{"unit_id": "ship", "route_id": "own-route"})}, nil)
	if err == nil || !strings.Contains(err.Error(), "piracy") {
		t.Fatalf("non-pirate/own-route piracy error = %v", err)
	}
}
