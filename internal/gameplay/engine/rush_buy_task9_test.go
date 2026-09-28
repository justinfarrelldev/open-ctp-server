package engine

import (
	"encoding/json"
	"testing"
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
