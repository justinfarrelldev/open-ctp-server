package engine

import (
	"reflect"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestSailedAroundWorldFeatTriggersWhenSeaUnitVisitsEveryColumn(t *testing.T) {
	state := runtimeStartedState(t)
	state.Tiles = []Tile{{ID: "a", X: 0, Y: 0, Terrain: TerrainOcean, Open: true}, {ID: "b", X: 1, Y: 0, Terrain: TerrainOcean, Open: true}, {ID: "c", X: 2, Y: 0, Terrain: TerrainOcean, Open: true}}
	state.Units = []Unit{{ID: "ship", OwnerID: "player-a", TileID: "a", Domain: DomainSea, Movement: 200, MaxMovement: 200, Health: 1, MaxHealth: 1}}
	runtime := &rules.RuntimeRules{
		Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainOcean, Cost: 100, Domains: []string{DomainSea}}}},
		Feats:    []rules.RuntimeFeatDefinition{{ID: "FEAT_SAILED_AROUND_WORLD", Duration: 25, Effects: []rules.RuntimeStarterEffect{{ID: "boat", Kind: EffectIncreaseBoatMovement, Value: 100}}}},
	}

	if err := moveUnit(&state, "player-a", moveUnitPayload{UnitID: "ship", Path: []string{"b", "c"}}, runtime); err != nil {
		t.Fatalf("moveUnit: %v", err)
	}
	if !reflect.DeepEqual(state.Units[0].CircumnavigatedColumns, []int{0, 1, 2}) {
		t.Fatalf("circumnavigated columns = %#v, want [0 1 2]", state.Units[0].CircumnavigatedColumns)
	}
	if !reflect.DeepEqual(state.AchievedFeatIDs, []string{"FEAT_SAILED_AROUND_WORLD"}) {
		t.Fatalf("achieved feats = %#v, want sailed-around-world", state.AchievedFeatIDs)
	}
	if !hasEffectValueTurns(state.Nations[0].Effects, EffectIncreaseBoatMovement, 100, 25) {
		t.Fatalf("nation effects after circumnavigation = %#v, want boat movement feat", state.Nations[0].Effects)
	}
}

func TestSailedAroundWorldFeatRequiresSeaMovement(t *testing.T) {
	state := runtimeStartedState(t)
	state.Tiles = []Tile{{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "b", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "c", X: 2, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{{ID: "scout", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 200, MaxMovement: 200, Health: 1, MaxHealth: 1}}
	runtime := &rules.RuntimeRules{
		Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}}}},
		Feats:    []rules.RuntimeFeatDefinition{{ID: "FEAT_SAILED_AROUND_WORLD", Duration: 25, Effects: []rules.RuntimeStarterEffect{{ID: "boat", Kind: EffectIncreaseBoatMovement, Value: 100}}}},
	}

	if err := moveUnit(&state, "player-a", moveUnitPayload{UnitID: "scout", Path: []string{"b", "c"}}, runtime); err != nil {
		t.Fatalf("moveUnit: %v", err)
	}
	if len(state.AchievedFeatIDs) != 0 || len(state.Nations[0].Effects) != 0 {
		t.Fatalf("land circumnavigation awarded feat: achieved=%#v effects=%#v", state.AchievedFeatIDs, state.Nations[0].Effects)
	}
}
