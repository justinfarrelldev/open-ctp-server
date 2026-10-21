package engine

import "testing"

func TestClientCancelFutureProductionPreservesCurrentProgress(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", StoredProduction: 40, Queue: []ProductionItem{{ID: "hoplite", Kind: ProductionUnit, Cost: 100}, {ID: "granary", Kind: ProductionBuilding, Cost: 50}}}}
	if err := cancelProduction(&state, "player-a", CancelProductionPayload{CityID: "home", ItemID: "granary"}); err != nil {
		t.Fatal(err)
	}
	if state.Cities[0].StoredProduction != 40 || len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].ID != "hoplite" {
		t.Fatalf("tail removal destroyed current production: %+v", state.Cities[0])
	}
}
