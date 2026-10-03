package engine

import (
	"fmt"
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

func TestRetailCityPollutionUsesDifficultyThresholdsAndRatios(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.Turn = state.RecalculatedTurn + 1
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 17}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityFood: 34, CityProduction: 301}}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}}}, PollutionModel: &rules.RuntimePollutionModel{PopulationStart: 16, PopulationRatio: 15, ProductionStart: 300, ProductionRatio: 1}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if got := tileByIDForTest(next.Tiles, "tile-a").Pollution; got != 16 {
		t.Fatalf("city tile pollution = %d, want (17-16)*15 plus (301-300)*1", got)
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

func TestRetailCityPollutionAppliesAdvancePollutionModifiers(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.Turn = state.RecalculatedTurn + 1
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Nations[0].Advances = []string{"ADVANCE_CLEAN"}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 10}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Movement: rules.RuntimeMovement{Configured: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 100, Domains: []string{DomainLand}, CityProduction: 10}}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1}}, Advances: []rules.RuntimeAdvanceDefinition{{ID: "ADVANCE_CLEAN", Cost: 1, PollutionSizeModifier: -50, PollutionProductionModifier: -100}}, PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{4000000}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}}}}
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if got := tileByIDForTest(next.Tiles, "tile-a").Pollution; got != 5 {
		t.Fatalf("city tile pollution = %d, want advance modifiers population 50%% plus production 0%%", got)
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

func TestRetailPollutionTriggerUsesWorldTileCountMapSize(t *testing.T) {
	state := runtimeStartedState(t)
	state.Tiles = make([]Tile, 4608)
	for index := range state.Tiles {
		state.Tiles[index] = Tile{ID: fmt.Sprintf("tile-%04d", index), X: index % 96, Y: index / 96, Terrain: TerrainPlains, Open: true}
	}
	state.Tiles[0].Pollution = 2
	state.Pollution = &PollutionState{History: []int64{9}}
	runtime := &rules.RuntimeRules{Movement: rules.RuntimeMovement{Terrains: []rules.RuntimeTerrain{{ID: "TERRAIN_DEAD", Domains: []string{DomainLand}}, {ID: TerrainPlains, Domains: []string{DomainLand}, CanDie: true}}}, PollutionTriggers: []rules.RuntimePollutionTriggerDefinition{
		{ID: "POLLUTION_TRIGGER_LIST_SMALL_MAP", TriggerLevels: []int64{100}, DisasterTypes: []string{"POLLUTION_FLOOD_DISASTER"}},
		{ID: "POLLUTION_TRIGGER_LIST_MEDIUM_MAP", TriggerLevels: []int64{10}, DisasterTypes: []string{"POLLUTION_OZONE_DISASTER"}},
	}}

	advanceRound(&state, runtime)
	if state.Pollution.LastDisaster != "POLLUTION_OZONE_DISASTER" {
		t.Fatalf("last disaster = %q, want medium-map trigger from tile count", state.Pollution.LastDisaster)
	}
}
