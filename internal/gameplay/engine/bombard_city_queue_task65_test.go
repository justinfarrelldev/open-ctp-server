package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailBombardDestroyedBuildingRemovesNowIllegalQueueItems(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{{ID: "artillery", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 1, BombRounds: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}}}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "target", Population: 1,
		Buildings: []Building{{ID: "IMPROVE_FACTORY", Effects: []Effect{{ID: "factory-production", Kind: EffectYieldProduction, Value: 5}}}},
		Queue: []ProductionItem{
			{ID: "UNIT_FACTORY_ONLY", Kind: ProductionUnit, Cost: 1},
			{ID: "UNIT_GUARD", Kind: ProductionUnit, Cost: 1},
		},
	}}
	runtime := &rules.RuntimeRules{
		Constants:  rules.RuntimeConstants{BombardDestroyBuildingPercent: 100},
		Production: []rules.RuntimeProduction{{ID: "UNIT_FACTORY_ONLY", Kind: ProductionUnit, Cost: 1, RequiredBuildings: []string{"IMPROVE_FACTORY"}}, {ID: "UNIT_GUARD", Kind: ProductionUnit, Cost: 1}},
		Units:      []rules.RuntimeUnitDefinition{{ID: "UNIT_FACTORY_ONLY", Domain: DomainLand, Movement: 1, Health: 1}, {ID: "UNIT_GUARD", Domain: DomainLand, Movement: 1, Health: 1}},
	}
	payload, _ := json.Marshal(bombardPayload{UnitID: "artillery", TargetTileID: "target"})

	if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), runtime); err != nil {
		t.Fatalf("bombard city err=%v", err)
	}
	if len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].ID != "UNIT_GUARD" {
		t.Fatalf("queue after building destruction = %#v, want only still-legal guard", state.Cities[0].Queue)
	}
}
