package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailCityCaptureTransfersBuiltWondersToConqueror(t *testing.T) {
	state := fixtureState()
	state.Cities = []City{
		{ID: "captured", OwnerID: "player-b", TileID: "tile-b", Population: 4},
		{ID: "new-owner-city", OwnerID: "player-a", TileID: "tile-a", Population: 4},
	}
	state.Wonders = []Wonder{{ID: "WONDER_INTERNET", OwnerID: "player-b", CityID: "captured", Effects: []Effect{{ID: "IMPROVE_COMPUTER_CENTER", Kind: EffectBuildingEverywhere}}}}
	runtime := &rules.RuntimeRules{Production: []rules.RuntimeProduction{{ID: "IMPROVE_ADVANCED_LAB", Kind: ProductionBuilding, Cost: 1, RequiredBuildings: []string{"IMPROVE_COMPUTER_CENTER"}}}}

	if got := captureCityAt(&state, "player-a", "tile-b", runtime); got != "captured" {
		t.Fatalf("captured city id = %q", got)
	}
	if state.Wonders[0].OwnerID != "player-a" {
		t.Fatalf("wonder owner after capture = %q, want player-a", state.Wonders[0].OwnerID)
	}
	if err := requireProductionUnlock(state, state.Cities[1], ProductionItem{Kind: ProductionBuilding, ID: "IMPROVE_ADVANCED_LAB"}, runtime); err != nil {
		t.Fatalf("transferred building-everywhere wonder did not affect new owner: %v", err)
	}
}

func TestRetailCityCaptureReappliesWonderAddEffectsForNewOwner(t *testing.T) {
	state := fixtureState()
	state.Cities = []City{{ID: "captured", OwnerID: "player-b", TileID: "tile-b", Population: 4}}
	state.Units = []Unit{{ID: "new-owner-unit", OwnerID: "player-a", TileID: "tile-a", Health: 5, MaxHealth: 5}}
	state.Wonders = []Wonder{{ID: "WONDER_RAMAYANA", OwnerID: "player-b", CityID: "captured", Effects: []Effect{{ID: "hp", Kind: EffectIncreaseHP, Value: 2}}}}

	captureCityAt(&state, "player-a", "tile-b")
	if state.Units[0].Health != 7 || state.Units[0].MaxHealth != 7 {
		t.Fatalf("new owner unit HP after wonder capture = %d/%d, want 7/7", state.Units[0].Health, state.Units[0].MaxHealth)
	}
}
