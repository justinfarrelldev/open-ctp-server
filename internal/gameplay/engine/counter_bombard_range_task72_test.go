package engine

import (
	"encoding/json"
	"testing"
)

func TestRetailCounterBombardHonorsDefenderBombardRange(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "middle", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 2, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "long-gun", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}},
		{ID: "defender", OwnerID: "player-b", TileID: "target", Domain: DomainLand, Defense: 0, Ranged: 100, BombardRange: 1, BombRounds: 1, Strength: 5, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "counter", Kind: EffectCanCounterBombard}, {ID: "bombard-land", Kind: EffectCanBombardLand}}},
	}
	payload, _ := json.Marshal(bombardPayload{UnitID: "long-gun", TargetTileID: "target"})

	if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), nil); err != nil {
		t.Fatalf("bombard err=%v", err)
	}
	attacker, ok := unitByID(state, "long-gun")
	if !ok || attacker.Health != 10 {
		t.Fatalf("out-of-range counter-bombard fired: attacker=%+v ok=%v", attacker, ok)
	}
}
