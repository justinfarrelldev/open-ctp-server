package engine

import (
	"encoding/json"
	"testing"
)

func task10State() State {
	state := fixtureState()
	state.Tiles = []Tile{
		{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true},
		{ID: "b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-a", Open: true},
		{ID: "c", X: 2, Y: 0, Terrain: TerrainPlains, OwnerID: "player-b", Open: true},
	}
	state.Units = []Unit{
		{ID: "settler-a", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "settler", Kind: EffectSettler}}},
		{ID: "settler-b", OwnerID: "player-b", TileID: "c", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "settler", Kind: EffectSettler}}},
	}
	return state
}

func task10Apply(t *testing.T, state State, actor, kind string, payload any) State {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "task10-" + kind + "-" + actor, ActorPlayerID: actor, Type: kind, Payload: body}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Apply(%s): %v", kind, err)
	}
	return next
}

func TestRetailIncreaseBrokeragesAmplifiesBrokerageCommerce(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_BROKERAGE", Effects: []Effect{{ID: "commerce", Kind: EffectYieldGoldPercent, Value: 20}, {ID: "brokerage", Kind: EffectBrokerage}}}}}}

	withoutWonder, err := cityYield(state, state.Cities[0])
	if err != nil {
		t.Fatal(err)
	}
	state.Wonders = []Wonder{{ID: "WONDER_ELECTRONIC_COMMERCE_ACT", OwnerID: "player-a", CityID: "city-a", Scope: "nation", Effects: []Effect{{ID: "brokerages", Kind: EffectIncreaseBrokerages, Value: 50}}}}
	withWonder, err := cityYield(state, state.Cities[0])
	if err != nil {
		t.Fatal(err)
	}
	if withoutWonder.Gold != 2 || withWonder.Gold != 3 {
		t.Fatalf("brokerage commerce gold = %d/%d, want 2/3", withoutWonder.Gold, withWonder.Gold)
	}
}

func TestRetailSettlerEffectsSeedFoundedCitySizeAndBuildings(t *testing.T) {
	state := task10State()
	state.Units[0].Effects = []Effect{{ID: "settle", Kind: EffectSettler}, {ID: "settle-size", Kind: EffectSettleSize, Value: 3}, {ID: "IMPROVE_GRANARY", Kind: EffectSettleBuilding}}
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	if len(state.Cities) != 1 || state.Cities[0].Population != 3 || len(state.Cities[0].Buildings) != 1 || state.Cities[0].Buildings[0].ID != "IMPROVE_GRANARY" {
		t.Fatalf("founded city = %+v", state.Cities)
	}
}

func TestTask10FoundCityConsumesSettlerAndRejectsIllegalSite(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	if len(state.Units) != 1 || state.Units[0].ID != "settler-b" || len(state.Cities) != 1 {
		t.Fatalf("founding must atomically consume settler and create city: %#v", state)
	}
	if state.Cities[0].Name != "Alpha" || state.Cities[0].OwnerID != "player-a" || state.Cities[0].TileID != "a" {
		t.Fatalf("city ownership/site mismatch: %#v", state.Cities[0])
	}
	foreign := task10State()
	foreign.Units[0].TileID = "c"
	if _, _, _, err := Apply(foreign, Command{SchemaVersion: CommandSchemaV1, ID: "bad-site", ActorPlayerID: "player-a", Type: CommandFoundCity, Payload: []byte(`{"unit_id":"settler-a","city_id":"bad","name":"Bad"}`)}, foreign.Rules, fixtureSources()); err == nil {
		t.Fatal("foreign city site must fail closed")
	}
}

func TestTask10ProductionEffectsAndCancellationVectors(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state = task10Apply(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "granary", Kind: ProductionBuilding, Cost: 3, Effects: []Effect{{ID: "granary-food", Kind: EffectYieldFood, Value: 2}}}})
	state, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.Cities[0].Yields.Food; got != 4 { // plains base 2 plus completed building 2
		t.Fatalf("food=%d want 4", got)
	}
	if len(state.Cities[0].Buildings) != 1 || state.Cities[0].Buildings[0].ID != "granary" {
		t.Fatalf("completed building missing: %#v", state.Cities[0])
	}
	state = task10Apply(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "slow", Kind: ProductionUnit, Cost: 99}})
	state.Cities[0].StoredProduction = 7
	state = task10Apply(t, state, "player-a", CommandCancelProduction, CancelProductionPayload{CityID: "city-a", ItemID: "slow"})
	if state.Cities[0].StoredProduction != 0 || len(state.Cities[0].Queue) != 0 {
		t.Fatalf("cancellation must remove item with no refund: %#v", state.Cities[0])
	}
}

func TestTask10RejectsUnknownEffectKinds(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	payload, err := json.Marshal(QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "opaque", Kind: ProductionBuilding, Cost: 1, Effects: []Effect{{ID: "opaque", Kind: "script"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "unknown-effect", ActorPlayerID: "player-a", Type: CommandQueueProduction, Payload: payload}, state.Rules, fixtureSources()); err == nil {
		t.Fatal("unknown effect kind must fail closed")
	}
}

func TestTask10RoundEffectPipelineIsStable(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Nations = []Nation{{ID: "nation-a", Effects: []Effect{{ID: "nation-production", Kind: EffectProduction, Value: 2}}}}
	state = task10Apply(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "global-food", Kind: ProductionWonder, Cost: 3, Scope: "global", Effects: []Effect{{ID: "global-food", Kind: EffectYieldFood, Value: 1}}}})
	first, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RecalculateRound(first)
	if err != nil {
		t.Fatal(err)
	}
	firstHash, err := CanonicalStateHash(first)
	if err != nil {
		t.Fatal(err)
	}
	secondHash, err := CanonicalStateHash(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != secondHash || first.Cities[0].Yields != (Yield{Food: 3, Production: 5, Gold: 1}) {
		t.Fatalf("round effect pipeline must be stable: %#v %s %s", first.Cities[0].Yields, firstHash, secondHash)
	}
}

func TestTask10ProductionAndEffectFailureVectors(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	city, ok := cityByID(state, "city-a")
	if !ok {
		t.Fatal("missing city")
	}
	for _, terrain := range []string{TerrainPlains, TerrainForest, TerrainOcean} {
		state.Tiles[0].Terrain = terrain
		if _, err := terrainYield(state, "a"); err != nil {
			t.Fatalf("terrain %s: %v", terrain, err)
		}
	}
	state.Tiles[0].Terrain = "unsupported"
	if _, err := terrainYield(state, "a"); err == nil {
		t.Fatal("unsupported terrain must fail closed")
	}
	state.Tiles[0].Terrain = TerrainPlains
	city.WorkedTileIDs = []string{"b"}
	if yield, err := cityYield(state, *city); err != nil || yield.Food != 3 || yield.Production != 5 {
		t.Fatalf("worked-tile yield=%#v err=%v", yield, err)
	}
	city.WorkedTileIDs = []string{"b", "c"}
	city.Population = 1
	if yield, err := cityYield(state, *city); err != nil || yield.Food != 3 || yield.Production != 5 {
		t.Fatalf("population-limited worked tile yield=%#v err=%v", yield, err)
	}
	city.Population = 2
	if yield, err := cityYield(state, *city); err != nil || yield.Food != 5 || yield.Production != 8 {
		t.Fatalf("population-2 worked tile yield=%#v err=%v", yield, err)
	}
	city.WorkedTileIDs = []string{"a"}
	if _, err := cityYield(state, *city); err == nil {
		t.Fatal("duplicate worked tile must fail")
	}
	city.WorkedTileIDs = nil
	city.Buildings = []Building{{ID: "prereq"}}
	if !hasPrerequisite(state, *city, "prereq") || hasPrerequisite(state, *city, "missing") {
		t.Fatal("building prerequisite resolution is incorrect")
	}
	state.Wonders = []Wonder{{ID: "wonder-prereq", OwnerID: "player-a", CityID: "city-a"}}
	if !hasPrerequisite(state, *city, "wonder-prereq") {
		t.Fatal("wonder prerequisite resolution is incorrect")
	}
	for _, item := range []ProductionItem{{}, {ID: "bad", Kind: "unknown", Cost: 1}, {ID: "bad", Kind: ProductionBuilding, Cost: 0}, {ID: "bad", Kind: ProductionBuilding, Cost: 1, Scope: "script"}} {
		if validateProductionItem(item) == nil {
			t.Fatalf("invalid item accepted: %#v", item)
		}
	}
	city.Queue = []ProductionItem{{ID: "unit", Kind: ProductionUnit, Cost: 1}}
	city.StoredProduction = 1
	if err := completeProduction(&state, city, nil); err != nil || len(city.Queue) != 0 {
		t.Fatalf("unit completion=%v city=%#v", err, city)
	}
	city.Queue = []ProductionItem{{ID: "duplicate-wonder", Kind: ProductionWonder, Cost: 1}}
	city.StoredProduction = 1
	state.Wonders = []Wonder{{ID: "duplicate-wonder", OwnerID: "player-b", CityID: "other"}}
	if err := completeProduction(&state, city, nil); err != nil || city.StoredProduction != 0 {
		t.Fatalf("wonder loser cancellation=%v city=%#v", err, city)
	}
	if err := cancelProduction(&state, "player-a", CancelProductionPayload{CityID: "city-a", ItemID: "missing"}); err == nil {
		t.Fatal("missing cancellation must fail")
	}
	if err := validateEffects([]Effect{{ID: "", Kind: EffectVision}}); err == nil {
		t.Fatal("effect id is required")
	}
	if got := addYield(Yield{Food: 1, Production: 2}, Yield{Food: 2, Gold: 3}); got != (Yield{Food: 3, Production: 2, Gold: 3}) {
		t.Fatalf("yield addition=%#v", got)
	}
}

func TestTask10CityCommandAuthorizationAndQueueRejections(t *testing.T) {
	state := task10State()
	if err := foundCity(&state, "player-a", FoundCityPayload{}); err == nil {
		t.Fatal("incomplete foundation payload accepted")
	}
	state.Units[0].Effects = nil
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"}); err == nil {
		t.Fatal("non-settler founded city")
	}
	state = task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-b", CityID: "city-a", Name: "Beta"}); err == nil {
		t.Fatal("duplicate city id accepted")
	}
	if _, err := applyCityCommand(&state, Command{ActorPlayerID: "player-a", Type: "unknown"}, nil); err == nil {
		t.Fatal("unknown city command accepted")
	}
	if err := queueProduction(&state, "player-b", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "x", Kind: ProductionUnit, Cost: 1}}, nil); err == nil {
		t.Fatal("foreign city queue accepted")
	}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "locked", Kind: ProductionBuilding, Cost: 1, Prerequisite: "missing"}}, nil); err == nil {
		t.Fatal("missing prerequisite accepted")
	}
	state.Wonders = []Wonder{{ID: "built", OwnerID: "player-b", CityID: "other"}}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "built", Kind: ProductionWonder, Cost: 1}}, nil); err == nil {
		t.Fatal("existing wonder accepted")
	}
	item := ProductionItem{ID: "repeat", Kind: ProductionUnit, Cost: 1}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: item}, nil); err != nil {
		t.Fatal(err)
	}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "city-a", Item: item}, nil); err == nil {
		t.Fatal("duplicate queue item accepted")
	}
	if err := cancelProduction(&state, "player-b", CancelProductionPayload{CityID: "city-a", ItemID: "repeat"}); err == nil {
		t.Fatal("foreign cancellation accepted")
	}
	if _, err := applyCityCommand(&state, Command{ActorPlayerID: "player-a", Type: CommandQueueProduction, Payload: []byte("{")}, nil); err == nil {
		t.Fatal("malformed city payload accepted")
	}
}

func TestTask10CanonicalCityStateValidationVectors(t *testing.T) {
	base := task10State()
	base = task10Apply(t, base, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	cases := []struct {
		name string
		edit func(*State)
	}{
		{"invalid recalculation", func(s *State) { s.RecalculatedTurn = 2 }},
		{"duplicate wonder", func(s *State) {
			s.Wonders = []Wonder{{ID: "w", OwnerID: "player-a", CityID: "city-a"}, {ID: "w", OwnerID: "player-b", CityID: "city-a"}}
		}},
		{"missing city id", func(s *State) { s.Cities[0].ID = "" }},
		{"missing city owner", func(s *State) { s.Cities[0].OwnerID = "" }},
		{"missing city tile", func(s *State) { s.Cities[0].TileID = "" }},
		{"unknown city tile", func(s *State) { s.Cities[0].TileID = "missing" }},
		{"unknown city effect", func(s *State) { s.Cities[0].Effects = []Effect{{ID: "x", Kind: "script"}} }},
		{"invalid building", func(s *State) { s.Cities[0].Buildings = []Building{{}} }},
		{"negative building upkeep", func(s *State) { s.Cities[0].Buildings = []Building{{ID: "b", Upkeep: -1}} }},
		{"invalid queued item", func(s *State) { s.Cities[0].Queue = []ProductionItem{{}} }},
		{"invalid wonder", func(s *State) { s.Wonders = []Wonder{{ID: "w"}} }},
		{"invalid wonder scope", func(s *State) {
			s.Wonders = []Wonder{{ID: "w", OwnerID: "player-a", CityID: "city-a", Scope: "script"}}
		}},
		{"invalid nation effect", func(s *State) { s.Nations = []Nation{{ID: "nation-a", Effects: []Effect{{ID: "x", Kind: "script"}}}} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, err := canonicalClone(base)
			if err != nil {
				t.Fatal(err)
			}
			tc.edit(&state)
			if err := validateCityState(state); err == nil {
				t.Fatal("invalid canonical city state was accepted")
			}
		})
	}
}

func TestTask10WonderScopeObsolescenceAndDeterministicRace(t *testing.T) {
	state := task10State()
	state = task10Apply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Alpha"})
	state.Phase.ActivePlayerIndex = 1
	state = task10Apply(t, state, "player-b", CommandFoundCity, FoundCityPayload{UnitID: "settler-b", CityID: "city-b", Name: "Beta"})
	wonder := ProductionItem{ID: "great-library", Kind: ProductionWonder, Cost: 3, Effects: []Effect{{ID: "library-vision", Kind: EffectVision, Value: 1}}}
	state = task10Apply(t, state, "player-b", CommandQueueProduction, QueueProductionPayload{CityID: "city-b", Item: wonder})
	state.Phase.ActivePlayerIndex = 0
	state = task10Apply(t, state, "player-a", CommandQueueProduction, QueueProductionPayload{CityID: "city-a", Item: wonder})
	state, err := RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Wonders) != 1 || state.Wonders[0].OwnerID != "player-a" {
		t.Fatalf("canonical wonder race winner=%#v; stable city tuple must select city-a", state.Wonders)
	}
	for _, city := range state.Cities {
		if city.ID == "city-b" && (len(city.Queue) != 0 || city.StoredProduction != 0) {
			t.Fatalf("wonder loser must be cancelled without refund: %#v", city)
		}
	}
	state.Cities[0].Buildings = []Building{{ID: "obsolete-granary", ObsoletedBy: "future-tech", Effects: []Effect{{ID: "obsolete-food", Kind: EffectYieldFood, Value: 2}}}}
	state.Cities[0].Obsolete = []string{"future-tech"}
	state.RecalculatedTurn = 0
	state, err = RecalculateRound(state)
	if err != nil {
		t.Fatal(err)
	}
	if state.Cities[0].Yields.Food != 2 {
		t.Fatalf("obsolete building must not apply effects, food=%d", state.Cities[0].Yields.Food)
	}
}
