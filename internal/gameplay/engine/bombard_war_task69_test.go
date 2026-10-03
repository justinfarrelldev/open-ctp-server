package engine

import (
	"encoding/json"
	"testing"
)

func TestRetailBombardStartsWarAndCancelsTradeRoutes(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "origin"}, {ID: "city-b", OwnerID: "player-b", TileID: "target"}}
	state.TradeRoutes = []TradeRoute{{ID: "combatants", SourceCityID: "city-a", DestinationCityID: "city-b"}}
	state.Units = []Unit{
		{ID: "artillery", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 1, BombRounds: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}},
		{ID: "defender", OwnerID: "player-b", TileID: "target", Domain: DomainLand, Defense: 0, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
	}
	payload, _ := json.Marshal(bombardPayload{UnitID: "artillery", TargetTileID: "target"})

	if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), nil); err != nil {
		t.Fatalf("bombard err=%v", err)
	}
	if len(state.TradeRoutes) != 0 || !playersAtWar(state, "player-a", "player-b") {
		t.Fatalf("bombard did not start war/cancel routes: wars=%#v routes=%#v", state.Policy.Wars, state.TradeRoutes)
	}
}
