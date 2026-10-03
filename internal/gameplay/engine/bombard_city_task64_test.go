package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailBombardCanDamageUndefendedEnemyCity(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{{ID: "artillery", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 1, BombRounds: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}}}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "target", Population: 2, Buildings: []Building{
		{ID: "IMPROVE_WALLS", Effects: []Effect{{ID: "walls", Kind: EffectCityWalls, Value: 1}}},
		{ID: "IMPROVE_MARKET", Effects: []Effect{{ID: "gold", Kind: EffectYieldGold, Value: 1}}},
	}}}
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{BombardDestroyBuildingPercent: 100, BombardKillPopulationPercent: 100}}
	payload, _ := json.Marshal(bombardPayload{UnitID: "artillery", TargetTileID: "target"})

	eventType, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), runtime)
	if err != nil || eventType != EventBombarded {
		t.Fatalf("bombard city event=%q err=%v", eventType, err)
	}
	if state.Units[0].Movement != 0 {
		t.Fatalf("bombard did not consume movement: %+v", state.Units[0])
	}
	city := state.Cities[0]
	if city.Population != 1 {
		t.Fatalf("city population after bombard = %d, want 1", city.Population)
	}
	if len(city.Buildings) != 1 || city.Buildings[0].ID != "IMPROVE_WALLS" {
		t.Fatalf("city buildings after bombard = %+v, want non-capitol random building destroyed and walls retained by deterministic roll", city.Buildings)
	}
}

func TestRetailBombardCityDoesNotDestroyCapitolBuildings(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{{ID: "artillery", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 1, BombRounds: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}}}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "target", Population: 1, Buildings: []Building{{ID: "IMPROVE_CAPITOL", Effects: []Effect{{ID: "capitol", Kind: EffectCapitol}}}}}}
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{BombardDestroyBuildingPercent: 100, BombardKillPopulationPercent: 100}}
	payload, _ := json.Marshal(bombardPayload{UnitID: "artillery", TargetTileID: "target"})

	if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), runtime); err != nil {
		t.Fatalf("bombard city err=%v", err)
	}
	if len(state.Cities[0].Buildings) != 1 || state.Cities[0].Buildings[0].ID != "IMPROVE_CAPITOL" || state.Cities[0].Population != 1 {
		t.Fatalf("capitol/pop after bombard = buildings %+v population %d", state.Cities[0].Buildings, state.Cities[0].Population)
	}
}
