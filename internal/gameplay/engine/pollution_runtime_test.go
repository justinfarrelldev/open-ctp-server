package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

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
