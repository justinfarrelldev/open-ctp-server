package engine

import (
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestTask8PublicWorksImprovementChangesMovementYieldAndPollution(t *testing.T) {
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].PublicWorks = 5
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a", Open: true, Pollution: 3}, {ID: "tile-c", X: 9, Y: 9, Terrain: TerrainPlains, OwnerID: "player-b", Open: true}}
	state.Units = []Unit{{ID: "worker", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 3, Effects: []Effect{{ID: "terraform", Kind: EffectSpecialTerraform}}}, {ID: "guard", OwnerID: "player-b", TileID: "tile-c", Domain: DomainLand, Movement: 1}}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "tile-c"}}
	runtime := mustRuntimeRules(t, state)

	if cost, ok := terrainCost(state, runtime, "tile-b"); !ok || cost != 2 {
		t.Fatalf("forest cost before road = %d, %v", cost, ok)
	}
	queued := runtimeApply(t, state, "player-a", CommandBuildImprovement, BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD"})
	if queued.Nations[0].PublicWorks != 4 || len(queued.Overlays) != 1 || queued.Overlays[0].Remaining != 1 {
		t.Fatalf("queued road state = %#v", queued)
	}
	completed := runtimeApply(t, queued, "player-a", CommandRoundComplete, struct{}{})
	if completed.Overlays[0].Remaining != 0 {
		t.Fatalf("road did not complete: %#v", completed.Overlays)
	}
	if cost, ok := terrainCost(completed, runtime, "tile-b"); !ok || cost != 1 {
		t.Fatalf("forest cost after road = %d, %v", cost, ok)
	}

	farmQueued := runtimeApply(t, completed, "player-a", CommandBuildImprovement, BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TERRAIN_IMPROVEMENT_SYNTHETIC_FARM"})
	farmCompleted := runtimeApply(t, farmQueued, "player-a", CommandRoundComplete, struct{}{})
	yield, err := terrainYield(farmCompleted, "tile-b")
	if err != nil || yield.Food != 2 || yield.Production != 2 {
		t.Fatalf("farm yield = %#v err=%v", yield, err)
	}

	cleanupQueued := runtimeApply(t, farmCompleted, "player-a", CommandBuildImprovement, BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP"})
	cleanupCompleted := runtimeApply(t, cleanupQueued, "player-a", CommandRoundComplete, struct{}{})
	if cleanupCompleted.Tiles[1].Pollution != 1 {
		t.Fatalf("cleanup pollution = %d, want 1", cleanupCompleted.Tiles[1].Pollution)
	}
	if got := len(cleanupCompleted.Overlays); got != 2 {
		t.Fatalf("cleanup one-shot overlay was not removed: %#v", cleanupCompleted.Overlays)
	}
}

func TestRetailTerrainImprovementRequiresEnabledAdvance(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].PublicWorks = 2
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a", Open: true}}
	state.Units = []Unit{{ID: "worker", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "terraform", Kind: EffectSpecialTerraform}}}}
	runtime := &rules.RuntimeRules{TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_FARMS", BuildTurns: 1, PublicWorksCost: 1, RequiredAdvanceID: "ADVANCE_AGRICULTURE"}}}

	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TILEIMP_FARMS"}, runtime); err == nil {
		t.Fatal("advance-gated improvement was accepted before advance was known")
	}
	state.Nations[0].Advances = []string{"ADVANCE_AGRICULTURE"}
	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TILEIMP_FARMS"}, runtime); err != nil {
		t.Fatalf("advance-gated improvement rejected after advance: %v", err)
	}
}

func TestRetailTerraformImprovementUsesTerrainTransformCostAndTurns(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].PublicWorks = 600
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a", Open: true}}
	state.Units = []Unit{{ID: "worker", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "terraform", Kind: EffectSpecialTerraform}}}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainForest, TransformRemoveTime: 1, TransformRemoveMaterials: 200}, {ID: "TERRAIN_GRASSLAND", TransformAddTime: 1, TransformAddMaterials: 400}}}, TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_TERRAFORM_GRASSLAND", BuildTurns: 1, PublicWorksCost: 1, TargetTerrain: "TERRAIN_GRASSLAND"}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	queued := runtimeApply(t, state, "player-a", CommandBuildImprovement, BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TILEIMP_TERRAFORM_GRASSLAND"})
	if queued.Nations[0].PublicWorks != 0 || len(queued.Overlays) != 1 || queued.Overlays[0].Remaining != 2 {
		t.Fatalf("queued terraform public works/overlay = %d/%+v, want retail transform cost 600 and turns 2", queued.Nations[0].PublicWorks, queued.Overlays)
	}
}

func TestRetailTerraformImprovementChangesTileTerrain(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].PublicWorks = 2
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a", Open: true}}
	state.Units = []Unit{{ID: "worker", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "terraform", Kind: EffectSpecialTerraform}}}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_TERRAFORM_GRASSLAND", BuildTurns: 1, PublicWorksCost: 1, TargetTerrain: "TERRAIN_GRASSLAND"}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	queued := runtimeApply(t, state, "player-a", CommandBuildImprovement, BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TILEIMP_TERRAFORM_GRASSLAND"})
	if queued.Overlays[0].TargetTerrain != "TERRAIN_GRASSLAND" || queued.Tiles[1].Terrain != TerrainForest {
		t.Fatalf("queued terraform = tiles:%+v overlays:%+v", queued.Tiles, queued.Overlays)
	}
	completed := runtimeApply(t, queued, "player-a", CommandRoundComplete, struct{}{})
	if completed.Tiles[1].Terrain != "TERRAIN_GRASSLAND" || len(completed.Overlays) != 0 {
		t.Fatalf("completed terraform = tiles:%+v overlays:%+v", completed.Tiles, completed.Overlays)
	}
}

func TestTerrainImprovementRuntimeCantBuildAndExcludes(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].PublicWorks = 10
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: "TERRAIN_WATER_SHELF", OwnerID: "player-a", Open: true}, {ID: "tile-c", X: 1, Y: 1, Terrain: "TERRAIN_PLAINS", OwnerID: "player-a", Open: true}}
	state.Units = []Unit{{ID: "worker", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 3, Effects: []Effect{{ID: "terraform", Kind: EffectSpecialTerraform}}}}
	runtime := &rules.RuntimeRules{TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{
		{ID: "TILEIMP_FARMS", Class: "Farm", BuildTurns: 1, PublicWorksCost: 1, CantBuildOn: []string{"TERRAIN_WATER_SHELF"}, Excludes: []string{"Mine"}},
		{ID: "TILEIMP_MINES", Class: "Mine", BuildTurns: 1, PublicWorksCost: 1},
	}}

	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: "TILEIMP_FARMS"}, runtime); err == nil {
		t.Fatal("farm build on cant-build-on terrain was accepted")
	}
	state.Overlays = []TileOverlay{{ID: "mine", TileID: "tile-c", Kind: "improvement", OwnerID: "player-a", ImprovementID: "TILEIMP_MINES"}}
	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{UnitID: "worker", TileID: "tile-c", ImprovementID: "TILEIMP_FARMS"}, runtime); err == nil {
		t.Fatal("farm build over excluded mine class was accepted")
	}
	state.Overlays = nil
	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{UnitID: "worker", TileID: "tile-c", ImprovementID: "TILEIMP_FARMS"}, runtime); err != nil {
		t.Fatalf("eligible farm build error = %v", err)
	}
}

func TestTask8InvalidTerrainImprovementCommandsFailClosed(t *testing.T) {
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].PublicWorks = 0
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a", Open: true}}
	state.Units = []Unit{{ID: "worker", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "terraform", Kind: EffectSpecialTerraform}}}}
	for _, tc := range []struct {
		name   string
		mutate func(State) State
		want   string
	}{
		{name: "unsupported improvement", mutate: func(s State) State { s.Nations[0].PublicWorks = 2; return s }, want: "terrain action"},
		{name: "insufficient public works", mutate: func(s State) State { return s }, want: "terrain action"},
		{name: "no worker effect", mutate: func(s State) State { s.Nations[0].PublicWorks = 2; s.Units[0].Effects = nil; return s }, want: "terrain action"},
		{name: "pending duplicate", mutate: func(s State) State {
			s.Nations[0].PublicWorks = 2
			s.Overlays = []TileOverlay{{ID: "pending", TileID: "tile-b", Kind: "improvement", OwnerID: "player-a", Remaining: 1, ImprovementID: "TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD"}}
			return s
		}, want: "terrain action"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base, err := canonicalClone(state)
			if err != nil {
				t.Fatal(err)
			}
			candidate := tc.mutate(base)
			improvementID := "TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD"
			if tc.name == "unsupported improvement" {
				improvementID = "TERRAIN_IMPROVEMENT_SYNTHETIC_MISSING"
			}
			err = runtimeApplyExpectError(t, candidate, "player-a", CommandBuildImprovement, BuildImprovementPayload{UnitID: "worker", TileID: "tile-b", ImprovementID: improvementID})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
	bad := state
	bad.Tiles[0].Pollution = -1
	if err := ValidateState(bad); err == nil {
		t.Fatal("negative tile pollution accepted")
	}
}

func serverOwnedRulesPin(t *testing.T) RulesPin {
	t.Helper()
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	return RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
}

func mustRuntimeRules(t *testing.T, state State) *rules.RuntimeRules {
	t.Helper()
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	return &runtime
}
