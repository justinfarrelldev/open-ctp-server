package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailGaiaCountdownRequiresConfiguredComputersSatellitesTowersAndCoverage(t *testing.T) {
	runtime := &rules.RuntimeRules{Gaia: rules.RuntimeGaia{ActivationTurns: 2, WonderID: "WONDER_SOLARIS", BuildingID: "IMPROVE_GAIA_COMPUTER", MinBuildingCount: 2, SatelliteBuildingID: "IMPROVE_POWER_SATELLITE", MinSatelliteCount: 2, MaxSatelliteCount: 4, TowerImprovementID: "TILEIMP_PROCESSING_TOWER", MinTowerCount: 1, MinTowerCoveragePercent: 50, TowerMinRadius: 1, TowerMaxRadius: 2}}
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive}},
		Nations: []Nation{{ID: "nation-a"}},
		Tiles:   []Tile{{ID: "west", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "center", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "east", X: 2, Y: 0, Terrain: TerrainPlains, Open: true}},
		Cities: []City{
			{ID: "core", OwnerID: "player-a", TileID: "west", Buildings: []Building{{ID: "IMPROVE_GAIA_COMPUTER"}, {ID: "IMPROVE_POWER_SATELLITE"}}},
			{ID: "solar", OwnerID: "player-a", TileID: "center", Buildings: []Building{{ID: "IMPROVE_POWER_SATELLITE"}}},
		},
		Wonders:  []Wonder{{ID: "WONDER_SOLARIS", OwnerID: "player-a", CityID: "solar"}},
		Overlays: []TileOverlay{{ID: "tower", Kind: "improvement", TileID: "center", OwnerID: "player-a", ImprovementID: "TILEIMP_PROCESSING_TOWER"}},
	}

	updateGaiaCountdown(&state, runtime, 10)
	if got := state.Nations[0].GaiaCompletionTurn; got != 0 {
		t.Fatalf("Gaia countdown started with one computer: %d", got)
	}
	state.Cities[1].Buildings = append(state.Cities[1].Buildings, Building{ID: "IMPROVE_GAIA_COMPUTER"})
	updateGaiaCountdown(&state, runtime, 10)
	if got := state.Nations[0].GaiaCompletionTurn; got != 12 {
		t.Fatalf("Gaia countdown = %d, want 12 after all configured retail prerequisites", got)
	}

	state.Overlays[0].Remaining = 1
	updateGaiaCountdown(&state, runtime, 11)
	if got := state.Nations[0].GaiaCompletionTurn; got != 0 {
		t.Fatalf("Gaia countdown survived incomplete tower: %d", got)
	}
}
