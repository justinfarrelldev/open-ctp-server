package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailCityBombardWallDefenseUsesActualBuildingsOnly(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{{ID: "artillery", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 1, BombRounds: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}}}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "target", Population: 2, Effects: []Effect{{ID: "global-wall-marker", Kind: EffectCityWalls, Value: 100}}}}
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{BombardKillPopulationPercent: 100}}
	payload, _ := json.Marshal(bombardPayload{UnitID: "artillery", TargetTileID: "target"})

	if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), runtime); err != nil {
		t.Fatalf("bombard city err=%v", err)
	}
	if state.Cities[0].Population != 1 {
		t.Fatalf("city population after bombard = %d, want global/non-building wall effects ignored for bombard probability", state.Cities[0].Population)
	}
}
