package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailTileImprovementBorderRadiusClaimsNeutralTiles(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].PublicWorks = 1
	state.Tiles = []Tile{
		{ID: "center", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true},
		{ID: "near", X: 1, Y: 0, Terrain: TerrainPlains, Open: true},
		{ID: "far", X: 3, Y: 0, Terrain: TerrainPlains, Open: true},
		{ID: "enemy", X: 0, Y: 1, Terrain: TerrainPlains, OwnerID: "player-b", Open: true},
	}
	runtime := &rules.RuntimeRules{TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_FORTIFICATIONS", BuildTurns: 1, PublicWorksCost: 1, IntBorderRadius: 2, SquaredBorderRadius: 5}}}
	if _, _, err := buildTerrainImprovement(&state, "player-a", BuildImprovementPayload{TileID: "center", ImprovementID: "TILEIMP_FORTIFICATIONS"}, runtime); err != nil {
		t.Fatalf("buildTerrainImprovement() error = %v", err)
	}

	advanceTerraform(&state)
	if got := tileByIDForTest(state.Tiles, "near").OwnerID; got != "player-a" {
		t.Fatalf("near tile owner = %q, want player-a", got)
	}
	if got := tileByIDForTest(state.Tiles, "far").OwnerID; got != "" {
		t.Fatalf("far tile owner = %q, want neutral", got)
	}
	if got := tileByIDForTest(state.Tiles, "enemy").OwnerID; got != "player-b" {
		t.Fatalf("enemy tile owner = %q, want unchanged player-b", got)
	}
}
