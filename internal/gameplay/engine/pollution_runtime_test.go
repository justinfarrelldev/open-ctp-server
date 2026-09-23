package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailCityPollutionUsesPopulationProductionAndReducers(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.Turn = state.RecalculatedTurn + 1
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 10, Buildings: []Building{
		{ID: "IMPROVE_ECO_TRANSIT", Effects: []Effect{{ID: "eco-transit", Kind: EffectPollutionPopulationPercent, Value: -30}}},
		{ID: "IMPROVE_RECYCLER", Effects: []Effect{{ID: "recycler-population", Kind: EffectPollutionPopulationPercent, Value: -20}, {ID: "recycler-production", Kind: EffectPollutionProductionPercent, Value: -20}}},
		{ID: "IMPROVE_MATTER_DECOMPILER", Effects: []Effect{{ID: "matter", Kind: EffectPollutionProductionPercent, Value: -50}}},
	}}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityProduction: 10}}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if got := tileByIDForTest(next.Tiles, "tile-a").Pollution; got != 8 {
		t.Fatalf("city tile pollution = %d, want population 10-50%% plus production 10-70%%", got)
	}
}

func TestRetailCityPollutionAppliesGovernmentCoefficientBeforeBuildingReducers(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.Turn = state.RecalculatedTurn + 1
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Nations[0].Effects = []Effect{{ID: "government:GOVERNMENT_DEMOCRACY:pollution", Kind: EffectPollutionPercent, Value: -20}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 10, Buildings: []Building{
		{ID: "IMPROVE_ECO_TRANSIT", Effects: []Effect{{ID: "eco-transit", Kind: EffectPollutionPopulationPercent, Value: -50}}},
		{ID: "IMPROVE_MATTER_DECOMPILER", Effects: []Effect{{ID: "matter", Kind: EffectPollutionProductionPercent, Value: -50}}},
	}}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityProduction: 10}}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if got := tileByIDForTest(next.Tiles, "tile-a").Pollution; got != 8 {
		t.Fatalf("city tile pollution = %d, want government 80%% applied before 50%% building reducers", got)
	}
}

func TestCityPollutionRequiresRetailPollutionTriggers(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.Turn = state.RecalculatedTurn + 1
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 10}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityProduction: 10}}}, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_SYNTHETIC", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SYNTHETIC", Domain: DomainLand, Movement: 1, Health: 1}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if got := tileByIDForTest(next.Tiles, "tile-a").Pollution; got != 0 {
		t.Fatalf("synthetic city tile pollution = %d, want no runtime-trigger pollution integration", got)
	}
}

func TestRetailPollutionTriggersDisasterDuringRoundCompletion(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}}
	state.Tiles[0].Pollution = 2
	state.Pollution = &PollutionState{History: []int64{3999999}}
	runtime := &rules.RuntimeRules{PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000, 6000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER", "POLLUTION_OZONE_DISASTER"}}}}

	advanceRound(&state, runtime)
	if state.Pollution.Phase != 1 || state.Pollution.LastDisaster != "POLLUTION_FLOOD_DISASTER" || state.Pollution.LastPollutionLevel != 4000001 || len(state.Pollution.History) != rules.RetailPollutionHistoryTurns || state.Pollution.History[0] != 4000001 {
		t.Fatalf("pollution state = %+v", state.Pollution)
	}
}

func TestRetailPollutionFloodDisasterConvertsLandAndRemovesOccupants(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{
		{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true, Pollution: 1},
		{ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, Open: true, Pollution: 9},
	}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-a", TileID: "tile-b"}}
	state.Units = []Unit{{ID: "unit-b", OwnerID: "player-a", TileID: "tile-b", Domain: DomainLand}}
	state.Pollution = &PollutionState{History: []int64{3999999}}
	runtime := &rules.RuntimeRules{PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}}}}

	advanceRound(&state, runtime)
	if got := tileByIDForTest(state.Tiles, "tile-b").Terrain; got != TerrainOcean {
		t.Fatalf("flooded tile terrain = %q, want %q", got, TerrainOcean)
	}
	if len(state.Cities) != 0 || len(state.Units) != 0 {
		t.Fatalf("occupants after flood: cities=%+v units=%+v", state.Cities, state.Units)
	}
}

func TestRetailPollutionOzoneDisasterConvertsLandToDeadTerrain(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{
		{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true, Pollution: 4},
		{ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, Open: true, Pollution: 7},
	}
	state.Pollution = &PollutionState{History: []int64{3999999}}
	runtime := &rules.RuntimeRules{
		Movement:          rules.RuntimeMovement{Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_DEAD", Domains: []string{DomainLand}}}},
		PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_OZONE_DISASTER"}}},
	}

	advanceRound(&state, runtime)
	if got := tileByIDForTest(state.Tiles, "tile-b").Terrain; got != "TERRAIN_DEAD" {
		t.Fatalf("ozone tile terrain = %q, want TERRAIN_DEAD", got)
	}
}

func tileByIDForTest(tiles []Tile, id string) Tile {
	for _, tile := range tiles {
		if tile.ID == id {
			return tile
		}
	}
	return Tile{}
}

func TestRetailPollutionDefaultsToSmallMapTriggerUntilWorldMapSizeIsModeled(t *testing.T) {
	state := runtimeStartedState(t)
	state.Tiles[0].Pollution = 2
	state.Pollution = &PollutionState{History: []int64{3999999}}
	runtime := &rules.RuntimeRules{PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{
		{ID: "POLLUTION_TRIGGER_LIST_HUGE_MAP", TriggerLevels: []int64{35000000}, DisasterTypes: []string{"POLLUTION_OZONE_DISASTER"}},
		{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}},
	}}

	advanceRound(&state, runtime)
	if state.Pollution.LastDisaster != "POLLUTION_FLOOD_DISASTER" {
		t.Fatalf("last disaster = %q, want small-map flood trigger", state.Pollution.LastDisaster)
	}
}
