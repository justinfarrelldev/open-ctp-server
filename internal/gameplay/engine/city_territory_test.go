package engine

import (
	"fmt"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func territoryState(t *testing.T) (State, *rules.RuntimeRules) {
	t.Helper()
	state := civilizationNameState(t)
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.CityTerritory = &rules.RuntimeCityTerritory{Version: rules.CityTerritoryVersion, BorderIntRadius: 4, BorderSquaredRadius: 20, Sizes: []rules.RuntimeCityInfluenceSize{{ID: "zero", Population: 0}, {ID: "one", Population: 6, IntRadius: 1, SquaredRadius: 2}, {ID: "two", Population: 18, IntRadius: 2, SquaredRadius: 5}}}
	defs, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: defs}
	state.Tiles = nil
	for x := -5; x <= 5; x++ {
		for y := -5; y <= 5; y++ {
			state.Tiles = append(state.Tiles, Tile{ID: fmt.Sprintf("%d,%d", x, y), X: x, Y: y, Terrain: TerrainPlains, Open: true})
		}
	}
	state.Units[0].TileID = "0,0"
	state.Units[1].TileID = "5,5"
	return state, &runtime
}

func TestFoundingClaimsNationalTerritoryAndDistinctCityInfluence(t *testing.T) {
	state, runtime := territoryState(t)
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}, runtime); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ id, owner, influence string }{{"0,0", "player-a", "home"}, {"1,1", "player-a", "home"}, {"3,0", "player-a", ""}, {"4,3", "", ""}} {
		tile, _ := tileByID(state, tc.id)
		if tile.OwnerID != tc.owner || tile.InfluenceCityID != tc.influence {
			t.Fatalf("tile %s = %+v, want owner %s influence %s", tc.id, tile, tc.owner, tc.influence)
		}
	}
}

func TestNationalFloodStopsAtForeignOwnershipAndMapHoles(t *testing.T) {
	state, runtime := territoryState(t)
	for i := range state.Tiles {
		if state.Tiles[i].X == 2 {
			state.Tiles[i].OwnerID = "player-b"
		}
	}
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}, runtime); err != nil {
		t.Fatal(err)
	}
	foreign, _ := tileByID(state, "2,0")
	behind, _ := tileByID(state, "3,0")
	if foreign.OwnerID != "player-b" || behind.OwnerID != "" {
		t.Fatalf("foreign wall stolen/crossed: %+v %+v", foreign, behind)
	}
}

func TestTerritoryFloodCannotJumpAcrossMissingMapCells(t *testing.T) {
	state, runtime := territoryState(t)
	kept := state.Tiles[:0]
	for _, tile := range state.Tiles {
		if tile.X != 1 {
			kept = append(kept, tile)
		}
	}
	state.Tiles = kept
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}, runtime); err != nil {
		t.Fatal(err)
	}
	tile, _ := tileByID(state, "2,0")
	if tile.OwnerID != "" || tile.InfluenceCityID != "" {
		t.Fatal("territory jumped a disconnected map gap")
	}
}

func TestCityInfluenceGrowthShrinkCaptureAndRemoval(t *testing.T) {
	state, runtime := territoryState(t)
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}, runtime); err != nil {
		t.Fatal(err)
	}
	state.Cities[0].Population = 7
	refreshCityInfluence(&state, runtime)
	expanded, _ := tileByID(state, "2,0")
	if expanded.InfluenceCityID != "home" {
		t.Fatal("population threshold did not expand influence")
	}
	state.Cities[0].Population = 1
	refreshCityInfluence(&state, runtime)
	shrunk, _ := tileByID(state, "2,0")
	if shrunk.InfluenceCityID != "" || shrunk.OwnerID != "player-a" {
		t.Fatal("shrink failed to distinguish influence from historical national ownership")
	}
	if id := captureCityAt(&state, "player-b", "0,0", runtime); id != "home" {
		t.Fatal("capture failed")
	}
	center, _ := tileByID(state, "0,0")
	near, _ := tileByID(state, "1,0")
	outer, _ := tileByID(state, "3,0")
	if center.OwnerID != "player-b" || near.OwnerID != "player-b" || near.InfluenceCityID != "home" || outer.OwnerID != "player-a" {
		t.Fatalf("captured territory disagrees with original influence transfer: %+v %+v %+v", center, near, outer)
	}
	if id := captureCityAt(&state, "player-a", "0,0", runtime); id != "home" {
		t.Fatal("recapture failed")
	}
	if tile, _ := tileByID(state, "1,0"); tile.OwnerID != "player-a" || tile.InfluenceCityID != "home" {
		t.Fatal("recapture failed to transfer assigned territory")
	}
	removeCityByID(&state, "home")
	refreshCityInfluence(&state, runtime)
	for _, tile := range state.Tiles {
		if tile.InfluenceCityID != "" {
			t.Fatal("removed city left dangling influence")
		}
	}
	center, _ = tileByID(state, "0,0")
	if center.OwnerID != "player-a" {
		t.Fatal("city loss incorrectly erased national claim")
	}
}

func TestPlayerDeparturePreservesCityInfluenceForAITakeover(t *testing.T) {
	state, runtime := territoryState(t)
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}, runtime); err != nil {
		t.Fatal(err)
	}
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "leave-city-owner", ActorPlayerID: "player-a", Type: CommandLeaveGame, Payload: []byte(`{}`)}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if tile, _ := tileByID(next, "0,0"); tile.InfluenceCityID != "home" || tile.OwnerID != "player-a" {
		t.Fatal("disconnect erased a city retained for AI takeover")
	}
	if err := ValidateState(next); err != nil {
		t.Fatal(err)
	}
}

func TestDisbandCommandClearsMissingCityInfluence(t *testing.T) {
	state, runtime := territoryState(t)
	// This bounded fixture's source starter is not a disbandable settler.
	// Supply an explicit legal definition, rather than bypassing disband checks.
	runtime.Movement = rules.RuntimeMovement{Configured: true, OpenAccess: true, Terrains: []rules.RuntimeTerrain{{ID: TerrainPlains, Cost: 1, Domains: []string{DomainLand}}}}
	runtime.Units = append(runtime.Units, rules.RuntimeUnitDefinition{ID: "UNIT_TEST_SETTLER", Domain: DomainLand, Movement: 1, Health: 10, ProductionCost: 1, Effects: []rules.RuntimeStarterEffect{{ID: "settler", Kind: EffectSettler}}})
	defs, hash, err := rules.CanonicalRuntimeDefinitions(*runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules.Definitions = defs
	state.Rules.ContentHash = hash
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}, runtime); err != nil {
		t.Fatal(err)
	}
	state.Cities = append(state.Cities, City{ID: "other", Name: "Antium", OwnerID: "player-a", TileID: "-5,-5", Population: 1})
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "disband-home", ActorPlayerID: "player-a", Type: CommandDisbandCity, Payload: []byte(`{"city_id":"home"}`)}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	for _, tile := range next.Tiles {
		if tile.InfluenceCityID == "home" {
			t.Fatal("disband left dangling influence")
		}
	}
	if err := ValidateState(next); err != nil {
		t.Fatal(err)
	}
}

func TestNeighborCityDoesNotStealExistingInfluence(t *testing.T) {
	state, runtime := territoryState(t)
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}, runtime); err != nil {
		t.Fatal(err)
	}
	state.Units = append(state.Units, Unit{ID: "second", OwnerID: "player-a", TileID: "2,0", Domain: DomainLand, Effects: []Effect{{Kind: EffectSettler}}})
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "second", CityID: "second"}, runtime); err != nil {
		t.Fatal(err)
	}
	shared, _ := tileByID(state, "1,0")
	center, _ := tileByID(state, "2,0")
	if shared.InfluenceCityID != "home" || center.InfluenceCityID != "second" {
		t.Fatalf("neighbor reassigned existing influence/center: %+v %+v", shared, center)
	}
}
