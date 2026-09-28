package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailRushBuySpendsGoldAndCompletesQueueNextTurn(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].Gold = 96
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", StoredProduction: 40, Effects: []Effect{{ID: "workshop", Kind: EffectYieldProduction, Value: 10}}, Queue: []ProductionItem{{ID: "IMPROVE_SYNTHETIC", Kind: ProductionBuilding, Cost: 100}}}}
	payload := json.RawMessage(`{"city_id":"city-a"}`)
	rushed, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "rush-buy", ActorPlayerID: "player-a", Type: CommandRushBuy, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("rush buy failed: %v", err)
	}
	if rushed.Nations[0].Gold != 0 || rushed.Cities[0].StoredProduction != 100 || len(events) != 1 || events[0].Type != EventProductionRushed {
		t.Fatalf("rush buy outcome = gold:%d city:%+v events:%+v", rushed.Nations[0].Gold, rushed.Cities[0], events)
	}
	completed := runtimeApply(t, rushed, "player-a", CommandRoundComplete, struct{}{})
	if len(completed.Cities[0].Buildings) != 1 || completed.Cities[0].Buildings[0].ID != "IMPROVE_SYNTHETIC" {
		t.Fatalf("rush-bought production did not complete next turn: %+v", completed.Cities[0])
	}
}

func TestRetailRushBuyUsesGovernmentRushModifier(t *testing.T) {
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_RETAIL", BuildingRushModifier: 4}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Players = []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations = []Nation{{ID: "nation-a", GovernmentID: "GOVERNMENT_RETAIL", Gold: 383}}
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", StoredProduction: 40, Queue: []ProductionItem{{ID: "IMPROVE_SYNTHETIC", Kind: ProductionBuilding, Cost: 100}}}}
	if err := runtimeApplyExpectError(t, state, "player-a", CommandRushBuy, RushBuyPayload{CityID: "city-a"}); err == nil {
		t.Fatal("rush buy should require government-modified 384 gold")
	}
	state.Nations[0].Gold = 384
	rushed := runtimeApply(t, state, "player-a", CommandRushBuy, RushBuyPayload{CityID: "city-a"})
	if rushed.Nations[0].Gold != 0 || rushed.Cities[0].StoredProduction != 100 {
		t.Fatalf("rush buy with modifier = gold:%d city:%+v, want exact 384 gold spend", rushed.Nations[0].Gold, rushed.Cities[0])
	}
}
