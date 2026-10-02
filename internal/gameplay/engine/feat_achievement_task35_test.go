package engine

import (
	"reflect"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailFeatAchievementPreventsRepeatedAward(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{
		{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 1, Buildings: []Building{{ID: "IMPROVE_THEATER"}}},
		{ID: "city-b", OwnerID: "player-a", TileID: "tile-b", Population: 1, StoredProduction: 1, Queue: []ProductionItem{{ID: "IMPROVE_THEATER", Kind: ProductionBuilding, Cost: 1}}},
	}
	runtime := &rules.RuntimeRules{
		Production: []rules.RuntimeProduction{{ID: "IMPROVE_THEATER", Kind: ProductionBuilding, Cost: 1}},
		Feats:      []rules.RuntimeFeatDefinition{{ID: "FEAT_THEATERS", Duration: 15, BuildingID: "IMPROVE_THEATER", BuildingCount: 2, Effects: []rules.RuntimeStarterEffect{{ID: "feat-happiness", Kind: EffectHappiness, Value: 5}}}},
	}

	if err := completeProduction(&state, &state.Cities[1], runtime); err != nil {
		t.Fatalf("first completeProduction: %v", err)
	}
	if !reflect.DeepEqual(state.AchievedFeatIDs, []string{"FEAT_THEATERS"}) {
		t.Fatalf("achieved feats after first award = %#v, want FEAT_THEATERS", state.AchievedFeatIDs)
	}
	if !hasEffectValueTurns(state.Nations[0].Effects, EffectHappiness, 5, 16) {
		t.Fatalf("nation effects after first feat = %#v, want active theater feat", state.Nations[0].Effects)
	}

	state.Nations[0].Effects = nil
	state.Tiles = append(state.Tiles, Tile{ID: "tile-c", X: 2, Y: 0, Terrain: TerrainPlains, Open: true})
	state.Cities = append(state.Cities, City{ID: "city-c", OwnerID: "player-a", TileID: "tile-c", Population: 1, StoredProduction: 1, Queue: []ProductionItem{{ID: "IMPROVE_THEATER", Kind: ProductionBuilding, Cost: 1}}})
	if err := completeProduction(&state, &state.Cities[2], runtime); err != nil {
		t.Fatalf("second completeProduction: %v", err)
	}
	if len(state.Nations[0].Effects) != 0 {
		t.Fatalf("nation effects after repeat feat = %#v, want no repeated award after achievement", state.Nations[0].Effects)
	}
}
