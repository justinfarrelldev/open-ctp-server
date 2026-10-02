package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestBuildingFeatAppliesWhenRequiredBuildingCountIsCompleted(t *testing.T) {
	state := buildingFeatState(t)

	next := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if len(next.Cities[1].Buildings) != 1 || next.Cities[1].Buildings[0].ID != "IMPROVE_THEATER" {
		t.Fatalf("completed city buildings = %#v, want IMPROVE_THEATER", next.Cities[1].Buildings)
	}
	if !hasEffectValueTurns(next.Nations[0].Effects, EffectHappiness, 5, 15) {
		t.Fatalf("nation effects after building feat = %#v, want 15-turn +5 happiness", next.Nations[0].Effects)
	}
}

func buildingFeatState(t *testing.T) State {
	t.Helper()
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Movement:             rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityProduction: 1}}},
		Feats:                []rules.RuntimeFeatDefinition{{ID: "FEAT_THEATERS", Duration: 15, BuildingID: "IMPROVE_THEATER", BuildingCount: 2, Effects: []rules.RuntimeStarterEffect{{ID: "feat-happiness", Kind: EffectHappiness, Value: 5}}}},
		Production:           []rules.RuntimeProduction{{ID: "IMPROVE_THEATER", Kind: ProductionBuilding, Cost: 1}, {ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}},
	}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := runtimeStartedState(t)
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}
	state.Cities = []City{
		{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 1, Buildings: []Building{{ID: "IMPROVE_THEATER"}}},
		{ID: "city-b", OwnerID: "player-a", TileID: "tile-b", Population: 1, Queue: []ProductionItem{{ID: "IMPROVE_THEATER", Kind: ProductionBuilding, Cost: 1}}},
	}
	return state
}
