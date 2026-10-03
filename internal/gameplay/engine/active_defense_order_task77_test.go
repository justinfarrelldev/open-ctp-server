package engine

import (
	"encoding/json"
	"testing"
)

func TestRetailActiveDefenseInterceptorsSortByAttackStrength(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "middle", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 2, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "bomber", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}},
		{ID: "a-weak", OwnerID: "player-b", TileID: "target", Domain: DomainLand, Attack: 1, Defense: 0, Ranged: 100, ActiveDefenseRange: 2, Strength: 1, Armor: 1, Health: 20, MaxHealth: 20, Effects: []Effect{{ID: "defend-land", Kind: EffectDefendLand}}},
		{ID: "z-strong", OwnerID: "player-b", TileID: "target", Domain: DomainLand, Attack: 10, Defense: 0, Ranged: 100, ActiveDefenseRange: 2, Strength: 10, Armor: 1, Health: 20, MaxHealth: 20, Effects: []Effect{{ID: "defend-land", Kind: EffectDefendLand}}},
	}
	payload, _ := json.Marshal(bombardPayload{UnitID: "bomber", TargetTileID: "target"})

	if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), nil); err != nil {
		t.Fatalf("bombard err=%v", err)
	}
	if _, ok := unitByID(state, "bomber"); ok {
		t.Fatalf("strong active defender should destroy the bombarder before weaker interceptors fire: units=%+v", state.Units)
	}
	weak, _ := unitByID(state, "a-weak")
	strong, _ := unitByID(state, "z-strong")
	if weak.UsedActiveDefense || !strong.UsedActiveDefense {
		t.Fatalf("active-defense order did not prefer highest attack: weak=%+v strong=%+v", weak, strong)
	}
}
