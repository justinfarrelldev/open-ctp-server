package engine

import "testing"

func TestRetailCanBeGiftedTransfersUnitToTargetAtGiverCapital(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "capital-tile", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "front", X: 2, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "capital", OwnerID: "player-a", TileID: "capital-tile", Buildings: []Building{{ID: "palace", Effects: []Effect{{ID: "capital", Kind: EffectCapitol}}}}}}
	state.Units = []Unit{{ID: "elephant", OwnerID: "player-a", TileID: "front", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "gift", Kind: EffectCanBeGifted}}}}

	eventType, err := applyUnitCommand(&state, Command{Type: CommandGiftUnit, ActorPlayerID: "player-a", Payload: mustPayload(t, giftUnitPayload{UnitID: "elephant", TargetPlayerID: "player-b"})}, nil)
	if err != nil {
		t.Fatal(err)
	}
	unit, ok := unitByID(state, "elephant")
	if eventType != EventUnitGifted || !ok || unit.OwnerID != "player-b" || unit.TileID != "capital-tile" || unit.TransportID != "" || unit.ArmyID != "" {
		t.Fatalf("gift result event=%q unit=%#v ok=%v", eventType, unit, ok)
	}
}

func TestRetailGiftUnitRequiresCanBeGiftedEffect(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "capital-tile", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "capital", OwnerID: "player-a", TileID: "capital-tile"}}
	state.Units = []Unit{{ID: "warrior", OwnerID: "player-a", TileID: "capital-tile", Domain: DomainLand, Movement: 1}}

	if _, err := applyUnitCommand(&state, Command{Type: CommandGiftUnit, ActorPlayerID: "player-a", Payload: mustPayload(t, giftUnitPayload{UnitID: "warrior", TargetPlayerID: "player-b"})}, nil); err == nil {
		t.Fatal("gift succeeded without CanBeGifted")
	}
}
