package engine

import (
	"encoding/json"
	"testing"
)

func TestRetailCounterBombardConsidersAllDefendersOnTargetTile(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "bomber", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 1, BombRounds: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-air", Kind: EffectCanBombardAir}}},
		{ID: "air-target", OwnerID: "player-b", TileID: "target", Domain: DomainAir, Defense: 0, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "counter-gun", OwnerID: "player-b", TileID: "target", Domain: DomainLand, Defense: 0, Ranged: 100, BombardRange: 1, BombRounds: 1, Strength: 5, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "counter", Kind: EffectCanCounterBombard}, {ID: "bombard-land", Kind: EffectCanBombardLand}}},
	}
	payload, _ := json.Marshal(bombardPayload{UnitID: "bomber", TargetTileID: "target"})

	if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), nil); err != nil {
		t.Fatalf("bombard err=%v", err)
	}
	bomber, ok := unitByID(state, "bomber")
	if !ok || bomber.Health >= 10 {
		t.Fatalf("counter-bombarder outside attacker's eligible target set did not fire: bomber=%+v ok=%v", bomber, ok)
	}
}
