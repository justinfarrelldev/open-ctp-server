package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailTerraformRequiresTerrainAddAndRemoveAdvances(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].PublicWorks = 600
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a", Open: true}}
	state.Units = []Unit{{ID: "worker", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "terraform", Kind: EffectSpecialTerraform}}}}
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{
			{ID: "TERRAIN_GRASSLAND", TransformAddTime: 1, TransformAddMaterials: 400, TransformAddAdvanceID: "ADVANCE_AGRICULTURAL_REVOLUTION"},
			{ID: TerrainForest, TransformRemoveTime: 1, TransformRemoveMaterials: 200, TransformRemoveAdvanceID: "ADVANCE_AGRICULTURAL_REVOLUTION"},
		}},
		TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_TERRAFORM_GRASSLAND", BuildTurns: 1, PublicWorksCost: 1, TargetTerrain: "TERRAIN_GRASSLAND"}},
	}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TILEIMP_TERRAFORM_GRASSLAND"}, &runtime); err == nil {
		t.Fatal("terraform accepted without the source RemoveAdvance and target AddAdvance")
	}
	state.Nations[0].Advances = []string{"ADVANCE_AGRICULTURAL_REVOLUTION"}
	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TILEIMP_TERRAFORM_GRASSLAND"}, &runtime); err != nil {
		t.Fatalf("terraform rejected after add/remove advances: %v", err)
	}
}

func TestRetailOzoneDisasterOnlyKillsCanDieTerrain(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{
		{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true, Pollution: 4},
		{ID: "tile-b", X: 1, Y: 0, Terrain: "TERRAIN_MOUNTAIN", Open: true, Pollution: 9},
	}
	state.Pollution = &PollutionState{History: []int64{3999999}}
	runtime := &rules.RuntimeRules{
		Movement: rules.RuntimeMovement{Terrains: []rules.RuntimeTerrain{
			{ID: "TERRAIN_DEAD", Domains: []string{DomainLand}},
			{ID: "TERRAIN_MOUNTAIN", Domains: []string{DomainLand}, CanDie: false},
			{ID: TerrainPlains, Domains: []string{DomainLand}, CanDie: true},
		}},
		PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_OZONE_DISASTER"}}},
	}

	advanceRound(&state, runtime)
	if got := tileByIDForTest(state.Tiles, "tile-a").Terrain; got != "TERRAIN_DEAD" {
		t.Fatalf("can-die tile terrain = %q, want TERRAIN_DEAD", got)
	}
	if got := tileByIDForTest(state.Tiles, "tile-b").Terrain; got != "TERRAIN_MOUNTAIN" {
		t.Fatalf("non-can-die tile terrain = %q, want TERRAIN_MOUNTAIN", got)
	}
}
