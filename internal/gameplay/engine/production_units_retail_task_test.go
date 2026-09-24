package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailPollutersToParksWonderConvertsMostPollutedCities(t *testing.T) {
	state := State{
		Tiles: []Tile{
			{ID: "gaia-tile", Terrain: TerrainPlains, Open: true, Pollution: 99},
			{ID: "dirty-a", Terrain: TerrainPlains, Open: true, Pollution: 40},
			{ID: "dirty-b", Terrain: TerrainPlains, Open: true, Pollution: 30},
			{ID: "clean", Terrain: TerrainPlains, Open: true, Pollution: 0},
		},
		Cities: []City{
			{ID: "gaia-city", OwnerID: "player-a", TileID: "gaia-tile", StoredProduction: 1, Queue: []ProductionItem{{ID: "WONDER_GAIA", Kind: ProductionWonder, Cost: 1, Effects: []Effect{{ID: "parks", Kind: EffectPollutersToParks, Value: 2}}}}},
			{ID: "dirty-city-a", OwnerID: "player-b", TileID: "dirty-a"},
			{ID: "dirty-city-b", OwnerID: "player-c", TileID: "dirty-b"},
			{ID: "clean-city", OwnerID: "player-d", TileID: "clean"},
		},
		Units: []Unit{{ID: "polluter-guard", OwnerID: "player-b", TileID: "dirty-a"}, {ID: "clean-guard", OwnerID: "player-d", TileID: "clean"}},
	}
	if err := completeProduction(&state, &state.Cities[0], nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := cityByID(state, "dirty-city-a"); ok {
		t.Fatalf("dirty-city-a survived polluters-to-parks: %#v", state.Cities)
	}
	if _, ok := cityByID(state, "dirty-city-b"); ok {
		t.Fatalf("dirty-city-b survived polluters-to-parks: %#v", state.Cities)
	}
	if _, ok := cityByID(state, "gaia-city"); !ok {
		t.Fatalf("wonder city must be excluded: %#v", state.Cities)
	}
	if tileByIDMust(t, state, "dirty-a").Pollution != 0 || tileByIDMust(t, state, "dirty-b").Pollution != 0 || parkForTile(state, "dirty-a") == nil || parkForTile(state, "dirty-b") == nil {
		t.Fatalf("parked polluted tiles = tiles:%#v overlays:%#v", state.Tiles, state.Overlays)
	}
	if _, ok := unitByID(state, "polluter-guard"); ok {
		t.Fatalf("unit on converted city tile survived: %#v", state.Units)
	}
	if _, ok := unitByID(state, "clean-guard"); !ok {
		t.Fatalf("unit away from converted cities removed: %#v", state.Units)
	}
}

func tileByIDMust(t *testing.T, state State, id string) Tile {
	t.Helper()
	tile, ok := tileByID(state, id)
	if !ok {
		t.Fatalf("tile %s missing", id)
	}
	return tile
}

func TestRetailFreeSlavesWonderRemovesSlaverUnitsEverywhere(t *testing.T) {
	state := State{
		Tiles:  []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}},
		Cities: []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", StoredProduction: 1, Queue: []ProductionItem{{ID: "WONDER_EMANCIPATION", Kind: ProductionWonder, Cost: 1, Effects: []Effect{{ID: "free", Kind: EffectFreeSlaves}}}}}},
		Units: []Unit{
			{ID: "slaver-a", OwnerID: "player-a", TileID: "tile-a", Effects: []Effect{{ID: "slave", Kind: EffectSpecialSlaveRaid}}},
			{ID: "slaver-b", OwnerID: "player-b", TileID: "tile-a", Effects: []Effect{{ID: "settler-slave", Kind: EffectSpecialSettlerSlaveRaid}}},
			{ID: "guard", OwnerID: "player-b", TileID: "tile-a"},
		},
	}
	if err := completeProduction(&state, &state.Cities[0], nil); err != nil {
		t.Fatal(err)
	}
	if len(state.Units) != 1 || state.Units[0].ID != "guard" {
		t.Fatalf("units after free-slaves wonder = %#v", state.Units)
	}
}

func TestRetailProhibitSlaversBlocksSlaverProduction(t *testing.T) {
	state := State{
		Tiles:   []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}},
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a"}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}},
		Wonders: []Wonder{{ID: "WONDER_EMANCIPATION", OwnerID: "player-a", CityID: "city-a", Effects: []Effect{{ID: "prohibit", Kind: EffectProhibitSlavers}}}},
	}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SLAVER", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "slave", Kind: EffectSpecialSlaveRaid}}}}}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_SLAVER", Kind: ProductionUnit, Cost: 1}}, runtime); err == nil {
		t.Fatal("slaver production allowed under prohibit-slavers wonder")
	}
}

func TestRetailCloseEmbassiesWonderRemovesForeignEmbassiesAndWars(t *testing.T) {
	state := State{
		Tiles:   []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}},
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}, {ID: "player-c", NationID: "nation-c"}},
		Nations: []Nation{{ID: "nation-a", Embassies: []string{"nation-b"}}, {ID: "nation-b", Embassies: []string{"nation-a"}}, {ID: "nation-c", Embassies: []string{"nation-a"}}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", StoredProduction: 1, Queue: []ProductionItem{{ID: "WONDER_CLOSE", Kind: ProductionWonder, Cost: 1, Effects: []Effect{{ID: "close", Kind: EffectCloseEmbassies}}}}}},
		Policy:  TurnPolicyState{Wars: []TurnPolicyWar{{A: "player-a", B: "player-b"}, {A: "player-b", B: "player-c"}}},
	}
	if err := completeProduction(&state, &state.Cities[0], nil); err != nil {
		t.Fatal(err)
	}
	if containsString(state.Nations[1].Embassies, "nation-a") || containsString(state.Nations[2].Embassies, "nation-a") {
		t.Fatalf("foreign embassies with close-embassy owner remain: %#v", state.Nations)
	}
	if !containsString(state.Nations[0].Embassies, "nation-b") {
		t.Fatalf("owner's outbound embassies should remain explicit: %#v", state.Nations[0].Embassies)
	}
	if len(state.Policy.Wars) != 1 || state.Policy.Wars[0].A != "player-b" || state.Policy.Wars[0].B != "player-c" {
		t.Fatalf("wars after close embassies = %#v", state.Policy.Wars)
	}
}

func TestRetailIncreaseHPWonderRaisesExistingAndFutureUnitHealth(t *testing.T) {
	state := State{
		Phase:   PhaseState{Turn: 1},
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a"}},
		Tiles:   []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}},
		Cities:  []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", StoredProduction: 1, Queue: []ProductionItem{{ID: "WONDER_HP", Kind: ProductionWonder, Cost: 1, Scope: "nation", Effects: []Effect{{ID: "hp", Kind: EffectIncreaseHP, Value: 3}}}}}},
		Units:   []Unit{{ID: "veteran", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Health: 8, MaxHealth: 10}},
	}
	if err := completeProduction(&state, &state.Cities[0], nil); err != nil {
		t.Fatal(err)
	}
	if state.Units[0].Health != 11 || state.Units[0].MaxHealth != 13 {
		t.Fatalf("existing unit health = %d/%d, want 11/13", state.Units[0].Health, state.Units[0].MaxHealth)
	}

	state.Cities[0].StoredProduction = 1
	state.Cities[0].Queue = []ProductionItem{{ID: "UNIT_INFANTRY", Kind: ProductionUnit, Cost: 1}}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_INFANTRY", Domain: DomainLand, Movement: 1, Health: 10}}}
	if err := completeProduction(&state, &state.Cities[0], runtime); err != nil {
		t.Fatal(err)
	}
	produced, ok := unitByID(state, "unit:city-a:UNIT_INFANTRY:1")
	if !ok || produced.Health != 13 || produced.MaxHealth != 13 {
		t.Fatalf("produced unit health = %#v, want 13/13", produced)
	}
}

func TestProductionCompletionCopiesRuntimeRangedStats(t *testing.T) {
	state := State{Phase: PhaseState{Turn: 7}}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_RANGED", Domain: DomainLand, Movement: 1, Attack: 3, Defense: 2, Ranged: 9, BombardRange: 2, BombRounds: 4, ActiveDefenseRange: 3, Strength: 3, Armor: 1, Health: 10, Upkeep: 4}}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}
	if err := completeUnitProduction(&state, &city, ProductionItem{ID: "UNIT_RANGED", Kind: ProductionUnit}, runtime); err != nil {
		t.Fatalf("completeUnitProduction() error = %v", err)
	}
	if len(state.Units) != 1 || state.Units[0].Movement != 1 || state.Units[0].MaxMovement != 1 || state.Units[0].Ranged != 9 || state.Units[0].BombardRange != 2 || state.Units[0].BombRounds != 4 || state.Units[0].ActiveDefenseRange != 3 || state.Units[0].Upkeep != 4 {
		t.Fatalf("produced ranged unit = %#v", state.Units)
	}
}

func TestRetailPopulationRemovingUnitsReduceCityPopulation(t *testing.T) {
	state := State{Phase: PhaseState{Turn: 7}}
	city := City{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "pop", Kind: EffectBuildingRemovesPopulation}}}}}

	if err := completeUnitProduction(&state, &city, ProductionItem{ID: "UNIT_SETTLER", Kind: ProductionUnit}, runtime); err != nil {
		t.Fatalf("completeUnitProduction() error = %v", err)
	}
	if got := city.Population; got != 1 {
		t.Fatalf("city population after settler = %d, want 1", got)
	}
}

func TestProductionCompletionCreatesRuntimeUnitAtCity(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Queue: []ProductionItem{{ID: "UNIT_SYNTHETIC", Kind: ProductionUnit, Cost: 1}}}}

	next, err := RecalculateRound(state)
	if err != nil {
		t.Fatalf("RecalculateRound() error = %v", err)
	}
	if len(next.Cities[0].Queue) != 0 || next.Cities[0].StoredProduction != 2 {
		t.Fatalf("city production state = %#v", next.Cities[0])
	}
	if len(next.Units) != 1 {
		t.Fatalf("units = %#v, want one produced unit", next.Units)
	}
	unit := next.Units[0]
	if unit.ID != "unit:city-a:UNIT_SYNTHETIC:1" || unit.OwnerID != "player-a" || unit.TileID != "tile-a" {
		t.Fatalf("produced unit identity/location = %+v", unit)
	}
	if unit.Domain != DomainLand || unit.Movement != 2 || unit.Attack != 4 || unit.Defense != 4 || unit.MaxHealth != 10 || unit.Health != 10 || unit.VisibilityRange != 2 {
		t.Fatalf("produced unit stats = %+v", unit)
	}
	if !hasEffect(unit.Effects, EffectSettler) || !hasEffect(unit.Effects, EffectSpecialEmbassy) {
		t.Fatalf("produced unit effects = %#v", unit.Effects)
	}
}
