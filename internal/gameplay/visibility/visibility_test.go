package visibility

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func fixtureState() engine.State {
	return engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		Revision:      7,
		Players:       []engine.Player{{ID: "a", NationID: "a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}},
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 4, ActivePlayerOrder: []string{"a", "b"}},
		RNG:           []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
		Tiles:         []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "2,0", X: 2, Y: 0, Terrain: engine.TerrainForest, Open: true}},
		Units:         []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "0,0"}, {ID: "HIDDEN_UNIT_TOKEN", OwnerID: "b", TileID: "2,0"}},
		Cities:        []engine.City{{ID: "HIDDEN_CITY_TOKEN", OwnerID: "b", TileID: "2,0"}},
	}
}

func projectionHasUnit(projection Projection, unitID string) bool {
	for _, unit := range projection.Units {
		if unit.ID == unitID {
			return true
		}
	}
	return false
}

func TestRetailStealthProjectionRequiresStealthDetection(t *testing.T) {
	state := fixtureState()
	state.Cities = []engine.City{{ID: "home", OwnerID: "a", TileID: "0,0"}}
	state.Units = []engine.Unit{{ID: "spy", OwnerID: "b", TileID: "1,0", Effects: []engine.Effect{{ID: "stealth", Kind: engine.EffectStealth}}}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projectionHasUnit(projection, "spy") {
		t.Fatalf("stealth unit leaked without detector: %#v", projection.Units)
	}

	state.Units = append(state.Units, engine.Unit{ID: "detector", OwnerID: "a", TileID: "0,0", VisibilityRange: 2, Effects: []engine.Effect{{ID: "see-stealth", Kind: engine.EffectCanSeeStealth}}})
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !projectionHasUnit(projection, "spy") {
		t.Fatalf("stealth unit was hidden from stealth detector: %#v", projection.Units)
	}
}

func TestRetailSubmarineProjectionRequiresUnderwaterDetection(t *testing.T) {
	state := fixtureState()
	state.Cities = []engine.City{{ID: "home", OwnerID: "a", TileID: "0,0"}}
	state.Units = []engine.Unit{{ID: "sub", OwnerID: "b", TileID: "1,0", Effects: []engine.Effect{{ID: "sub", Kind: engine.EffectSubmarine}}}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projectionHasUnit(projection, "sub") {
		t.Fatalf("submarine leaked without detector: %#v", projection.Units)
	}

	state.Units = append(state.Units, engine.Unit{ID: "detector", OwnerID: "a", TileID: "0,0", VisibilityRange: 2, Effects: []engine.Effect{{ID: "underwater", Kind: engine.EffectCanAttackUnderwater}}})
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !projectionHasUnit(projection, "sub") {
		t.Fatalf("submarine was hidden from underwater detector: %#v", projection.Units)
	}
}

func projectionHasTile(projection Projection, tileID string) bool {
	for _, tile := range projection.Tiles {
		if tile.ID == tileID {
			return true
		}
	}
	return false
}

func TestRetailObsoleteGlobalRadarWonderDoesNotRevealAllTiles(t *testing.T) {
	state := fixtureState()
	state.Nations = []engine.Nation{{ID: "a", Advances: []string{"ADVANCE_COUNTER"}}, {ID: "b"}}
	state.Tiles = append(state.Tiles, engine.Tile{ID: "5,0", X: 5, Y: 0, Terrain: engine.TerrainPlains, Open: true})
	state.Cities = append(state.Cities, engine.City{ID: "home", OwnerID: "a", TileID: "0,0"})
	state.Wonders = []engine.Wonder{{ID: "WONDER_RADAR", OwnerID: "a", CityID: "home", Scope: "nation", ObsoletedBy: "ADVANCE_COUNTER", Effects: []engine.Effect{{ID: "radar", Kind: engine.EffectGlobalRadar}}}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projectionHasTile(projection, "5,0") {
		t.Fatalf("obsolete global radar revealed hidden tile: %#v", projection.Tiles)
	}
}

func TestRetailTradeRouteProjectionDoesNotLeakHiddenDestinations(t *testing.T) {
	state := fixtureState()
	state.Tiles = append(state.Tiles, engine.Tile{ID: "5,0", X: 5, Y: 0, Terrain: engine.TerrainPlains, Open: true})
	state.Cities = append(state.Cities, engine.City{ID: "home", OwnerID: "a", TileID: "0,0"}, engine.City{ID: "far", OwnerID: "b", TileID: "5,0"})
	state.TradeRoutes = []engine.TradeRoute{{ID: "hidden-route", SourceCityID: "home", DestinationCityID: "far", CrossesWater: true}}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.TradeRoutes) != 0 {
		t.Fatalf("hidden trade route leaked: %#v", projection.TradeRoutes)
	}

	state.Wonders = []engine.Wonder{{ID: "WONDER_RADAR", OwnerID: "a", CityID: "home", Scope: "nation", Effects: []engine.Effect{{ID: "radar", Kind: engine.EffectGlobalRadar}}}}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.TradeRoutes) != 1 || projection.TradeRoutes[0].ID != "hidden-route" || !projection.TradeRoutes[0].CrossesWater {
		t.Fatalf("visible trade route missing: %#v", projection.TradeRoutes)
	}
}

func TestRetailEmbassiesEverywhereProjectionHonorsCloseEmbassies(t *testing.T) {
	state := fixtureState()
	state.Nations = []engine.Nation{{ID: "a"}, {ID: "b"}}
	state.Cities = append(state.Cities, engine.City{ID: "home-city", OwnerID: "a", TileID: "0,0"})
	state.Wonders = []engine.Wonder{{ID: "WONDER_GLOBAL_SURVEILLANCE_CENTER", OwnerID: "a", CityID: "home-city", Scope: "nation", Effects: []engine.Effect{{ID: "everywhere", Kind: engine.EffectEmbassiesEverywhere}}}}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(projection.EmbassyPlayerIDs, "b") {
		t.Fatalf("embassies-everywhere projection = %#v", projection.EmbassyPlayerIDs)
	}

	state.Wonders = append(state.Wonders, engine.Wonder{ID: "WONDER_CLOSE", OwnerID: "b", CityID: "HIDDEN_CITY_TOKEN", Effects: []engine.Effect{{ID: "close", Kind: engine.EffectCloseEmbassies}}})
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if containsString(projection.EmbassyPlayerIDs, "b") {
		t.Fatalf("close embassies target still projected as embassy: %#v", projection.EmbassyPlayerIDs)
	}
}

func TestRetailGlobalRadarRevealsAllTilesAndUnits(t *testing.T) {
	state := fixtureState()
	state.Cities = append(state.Cities, engine.City{ID: "radar-city", OwnerID: "a", TileID: "0,0"})
	state.Wonders = []engine.Wonder{{ID: "WONDER_RADAR", OwnerID: "a", CityID: "radar-city", Scope: "nation", Effects: []engine.Effect{{ID: "radar", Kind: engine.EffectGlobalRadar}}}}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Tiles) != len(state.Tiles) || !projectionHasUnit(projection, "HIDDEN_UNIT_TOKEN") || !projectionHasCity(projection, "HIDDEN_CITY_TOKEN") {
		t.Fatalf("global radar projection = %+v", projection)
	}
}

func projectionHasCity(projection Projection, id string) bool {
	for _, city := range projection.Cities {
		if city.ID == id {
			return true
		}
	}
	return false
}

func TestObservationMemoryProjectionAndRawLeakBoundaries(t *testing.T) {
	state, err := Update(fixtureState())
	if err != nil {
		t.Fatal(err)
	}
	first, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Tiles) != 3 || len(first.Units) != 2 || len(first.Cities) != 1 {
		t.Fatalf("visible projection = %#v", first)
	}

	state.Units = state.Units[1:]
	state.Tiles[2].Terrain = engine.TerrainPlains
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	stale, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(stale.Units) != 0 || len(stale.Cities) != 1 || !stale.Tiles[2].Stale || stale.Tiles[2].Terrain != engine.TerrainForest {
		t.Fatalf("stale projection = %#v", stale)
	}

	projectionBytes, err := json.Marshal(stale)
	if err != nil {
		t.Fatal(err)
	}
	projector := NewProjector("a")
	snapshot, err := projector.ProjectSnapshot(events.Viewer{GameID: 1, PlayerID: 1}, mustStateJSON(t, state))
	if err != nil {
		t.Fatal(err)
	}
	event, err := projector.ProjectEvent(events.Viewer{GameID: 1, PlayerID: 1}, events.Event{Type: "combat", Sequence: 8, PayloadJSON: []byte(`{"target":"HIDDEN_UNIT_TOKEN","terrain":"SECRET_TERRAIN_TOKEN"}`)})
	if err != nil {
		t.Fatal(err)
	}
	ai, err := AIInput(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range [][]byte{projectionBytes, snapshot, event, ai, SafeError(errors.New("target HIDDEN_UNIT_TOKEN at SECRET_TERRAIN_TOKEN"))} {
		if bytes.Contains(raw, []byte("HIDDEN_UNIT_TOKEN")) || bytes.Contains(raw, []byte("SECRET_TERRAIN_TOKEN")) {
			t.Fatalf("hidden data leaked: %s", raw)
		}
	}
	if !bytes.Equal(ai, projectionBytes) {
		t.Fatalf("AI input differs from controller projection: %s != %s", ai, projectionBytes)
	}
}

func TestTask9ProjectionExposesOnlyViewerDiplomacy(t *testing.T) {
	state := fixtureState()
	state.Nations = []engine.Nation{{ID: "a", GovernmentID: "GOVERNMENT_SYNTHETIC_REPUBLIC"}, {ID: "b"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "trade", Trade: true}}
	state.Proposals = []engine.DiplomaticProposal{{ID: "p1", FromPlayerID: "b", ToPlayerID: "a", Kind: "union"}}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projection.Diplomacy == nil || projection.Diplomacy.GovernmentID != "GOVERNMENT_SYNTHETIC_REPUBLIC" || len(projection.Diplomacy.Relations) != 1 || projection.Diplomacy.Relations[0].TargetPlayerID != "b" || len(projection.Diplomacy.Proposals) != 1 {
		t.Fatalf("diplomacy projection = %#v", projection.Diplomacy)
	}
}

func TestTask8ProjectionShowsOnlyVisiblePollution(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true, Pollution: 1}, {ID: "9,9", X: 9, Y: 9, Terrain: engine.TerrainForest, Open: true, Pollution: 9}}
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "0,0", VisibilityRange: 0}}
	state.Cities = nil
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Tiles) != 1 || projection.Tiles[0].ID != "0,0" || projection.Tiles[0].Pollution != 1 {
		t.Fatalf("pollution projection = %#v", projection.Tiles)
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("9,9")) || bytes.Contains(encoded, []byte("9")) && bytes.Contains(encoded, []byte("pollution")) {
		t.Fatalf("hidden pollution leaked: %s", encoded)
	}
}

func TestTask7ProjectionExposesOwnedCityEconomyOnly(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "0,0", VisibilityRange: 2}}
	state.Cities = []engine.City{
		{ID: "city-a", OwnerID: "a", TileID: "0,0", Population: 2, StoredFood: 7, StoredProduction: 2, Queue: []engine.ProductionItem{{ID: "granary", Kind: engine.ProductionBuilding, Cost: 3}}, Buildings: []engine.Building{{ID: "shrine", Effects: []engine.Effect{{ID: "faith", Kind: engine.EffectYieldGold, Value: 1}}}}, Yields: engine.Yield{Food: 4, Production: 3, Gold: 2}},
		{ID: "city-b", OwnerID: "b", TileID: "1,0", Population: 9, StoredFood: 99, StoredProduction: 99, Queue: []engine.ProductionItem{{ID: "secret", Kind: engine.ProductionWonder, Cost: 9}}, Buildings: []engine.Building{{ID: "secret-building"}}, Yields: engine.Yield{Food: 9, Production: 9, Gold: 9}},
	}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	cities := map[string]CityProjection{}
	for _, city := range projection.Cities {
		cities[city.ID] = city
	}
	own := cities["city-a"]
	if own.Yields == nil || own.Yields.Food != 4 || own.Population != 2 || own.StoredFood != 7 || own.StoredProduction != 2 || len(own.Queue) != 1 || len(own.Buildings) != 1 {
		t.Fatalf("owned city projection = %#v", own)
	}
	enemy := cities["city-b"]
	if enemy.Yields != nil || enemy.Population != 0 || enemy.StoredFood != 0 || enemy.StoredProduction != 0 || len(enemy.Queue) != 0 || len(enemy.Buildings) != 0 {
		t.Fatalf("enemy economy leaked: %#v", enemy)
	}
}

func TestTask6ProjectionExposesOwnedResearchOnly(t *testing.T) {
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = engine.RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
	state.Nations = []engine.Nation{
		{ID: "a", Advances: []string{"ADVANCE_SYNTHETIC_ROOT"}, ResearchTargetID: "ADVANCE_SYNTHETIC_UNLOCK", ResearchProgress: 1},
		{ID: "b", Advances: []string{"ADVANCE_SECRET"}, ResearchTargetID: "ADVANCE_SECRET_NEXT", ResearchProgress: 99},
	}
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projection.Research == nil || projection.Research.TargetID != "ADVANCE_SYNTHETIC_UNLOCK" || projection.Research.Progress != 1 || !reflect.DeepEqual(projection.Research.Advances, []string{"ADVANCE_SYNTHETIC_ROOT"}) || !reflect.DeepEqual(projection.Research.Available, []string{"ADVANCE_SYNTHETIC_UNLOCK"}) {
		t.Fatalf("research projection = %#v", projection.Research)
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("ADVANCE_SECRET")) {
		t.Fatalf("opponent research leaked: %s", encoded)
	}
}

func TestTask5ProjectionExposesOwnedUnitCapabilitiesOnly(t *testing.T) {
	state := fixtureState()
	state.Units = []engine.Unit{
		{ID: "owned", OwnerID: "a", TileID: "0,0", Domain: engine.DomainLand, Movement: 3, CargoCapacity: 2, CargoDomains: []string{engine.DomainLand}, ZOC: true, Attack: 7, Defense: 5, Strength: 6, Armor: 2, Health: 8, MaxHealth: 10, VisibilityRange: 3, Effects: []engine.Effect{{ID: "settler", Kind: engine.EffectSettler}}},
		{ID: "visible-enemy", OwnerID: "b", TileID: "1,0", Domain: engine.DomainSea, Movement: 9, Attack: 99, Defense: 99, Strength: 99, Armor: 9, Health: 9, MaxHealth: 9},
	}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Units) != 2 {
		t.Fatalf("projection units = %#v", projection.Units)
	}
	owned := projection.Units[0]
	if owned.ID != "owned" || owned.Domain != engine.DomainLand || owned.Movement != 3 || owned.CargoCapacity != 2 || !owned.ZOC || owned.Attack != 7 || owned.Defense != 5 || owned.Strength != 6 || owned.Armor != 2 || owned.Health != 8 || owned.MaxHealth != 10 || owned.VisibilityRange != 3 || len(owned.Effects) != 1 {
		t.Fatalf("owned projection capabilities = %#v", owned)
	}
	enemy := projection.Units[1]
	if enemy.ID != "visible-enemy" || enemy.Domain != "" || enemy.Movement != 0 || enemy.Attack != 0 || enemy.Health != 0 || len(enemy.Effects) != 0 {
		t.Fatalf("enemy projection leaked capabilities = %#v", enemy)
	}
}

func TestTask5UnitVisibilityRangeExpandsOwnedSight(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "2,0", X: 2, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "3,0", X: 3, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "0,0", VisibilityRange: 3}}
	state.Cities = []engine.City{{ID: "far-city", OwnerID: "b", TileID: "3,0"}}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Tiles) != 4 || len(projection.Cities) != 1 || projection.Cities[0].ID != "far-city" {
		t.Fatalf("visibility range projection = %#v", projection)
	}
}

func TestVisibilityBlocksSightAndCanonicalFogSurvivesRestart(t *testing.T) {
	state := fixtureState()
	state.Tiles[1].Terrain = engine.TerrainForest
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Tiles) != 2 || len(projection.Units) != 1 {
		t.Fatalf("blocking projection = %#v", projection)
	}
	encoded := mustStateJSON(t, state)
	restarted, err := engine.DecodeCanonicalStateJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := restarted.Fog, state.Fog; !reflect.DeepEqual(got, want) {
		t.Fatalf("fog restart mismatch: %#v != %#v", got, want)
	}
}

func TestSSEResyncAndTerminalUseViewerProjection(t *testing.T) {
	state, err := Update(fixtureState())
	if err != nil {
		t.Fatal(err)
	}
	state.Units = state.Units[1:]
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	state.Terminal = engine.TerminalState{Ended: true, Reason: "complete", WinningPlayerID: "a"}
	state.Revision = 8
	canonical := mustStateJSON(t, state)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stream := events.NewStore(sqlx.NewDb(db, "sqlmock"))
	projector := NewProjector("a")
	viewer := events.Viewer{GameID: 9, PlayerID: 1}
	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(5))
	mock.ExpectQuery("SELECT revision, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "canonical_state_json"}).AddRow(8, canonical))
	items, err := stream.Resume(context.Background(), viewer, 1, projector)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "resync" || bytes.Contains(items[0].Data, []byte("HIDDEN_UNIT_TOKEN")) {
		t.Fatalf("unsafe resync: %#v", items)
	}
	var resync struct {
		CurrentRevision int64           `json:"current_revision"`
		Snapshot        json.RawMessage `json:"snapshot"`
	}
	if err := json.Unmarshal(items[0].Data, &resync); err != nil {
		t.Fatal(err)
	}
	var snapshot Projection
	if err := json.Unmarshal(resync.Snapshot, &snapshot); err != nil || resync.CurrentRevision != snapshot.Revision || snapshot.Revision != state.Revision {
		t.Fatalf("terminal resync snapshot = %#v, projection = %#v, err = %v", resync, snapshot, err)
	}
	mock.ExpectQuery("SELECT last_event_sequence, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(8, canonical))
	terminal, ended, err := stream.Terminal(context.Background(), viewer, projector)
	if err != nil || !ended || bytes.Contains(terminal.Data, []byte("HIDDEN_UNIT_TOKEN")) {
		t.Fatalf("unsafe terminal: %#v %v %v", terminal, ended, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTask12ProjectionIncludesRulesAndLegalActionMetadata(t *testing.T) {
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = engine.RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
	state.Tiles = []engine.Tile{
		{ID: "home", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true},
		{ID: "site", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true},
		{ID: "forest", X: 0, Y: 1, Terrain: engine.TerrainForest, Open: true, Pollution: 1},
		{ID: "enemy", X: 1, Y: 1, Terrain: engine.TerrainPlains, Open: true},
	}
	state.Nations = []engine.Nation{{ID: "a", Advances: []string{"ADVANCE_SYNTHETIC_ROOT"}, PublicWorks: 5}, {ID: "b"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "peace"}}
	state.Cities = []engine.City{{ID: "city-a", OwnerID: "a", TileID: "home"}, {ID: "city-b", OwnerID: "b", TileID: "enemy"}}
	state.Armies = []engine.Army{{ID: "army-a", OwnerID: "a", UnitIDs: []string{"attacker"}}, {ID: "army-b", OwnerID: "b", UnitIDs: []string{"defender"}}}
	state.Units = []engine.Unit{
		{ID: "settler", OwnerID: "a", TileID: "site", Domain: engine.DomainLand, Movement: 2, Health: 1, MaxHealth: 1, Effects: []engine.Effect{{ID: "settle", Kind: engine.EffectSettler}}},
		{ID: "worker", OwnerID: "a", TileID: "home", Domain: engine.DomainLand, Movement: 2, Health: 1, MaxHealth: 1, Effects: []engine.Effect{{ID: "terraform", Kind: engine.EffectSpecialTerraform, Value: 1}, {ID: "turns", Kind: engine.EffectSpecialTerraformTurns, Value: 1}}},
		{ID: "diplomat", OwnerID: "a", TileID: "home", Domain: engine.DomainLand, Movement: 2, Health: 1, MaxHealth: 1, Effects: []engine.Effect{{ID: "embassy", Kind: engine.EffectSpecialEmbassy}, {ID: "steal", Kind: engine.EffectSpecialSteal}}},
		{ID: "cargo", OwnerID: "a", TileID: "home", Domain: engine.DomainLand, Movement: 2, Health: 1, MaxHealth: 1},
		{ID: "ship", OwnerID: "a", TileID: "home", Domain: engine.DomainSea, Movement: 2, Health: 1, MaxHealth: 1, CargoCapacity: 1, CargoDomains: []string{engine.DomainLand}},
		{ID: "attacker", OwnerID: "a", TileID: "home", ArmyID: "army-a", Domain: engine.DomainLand, Movement: 2, Attack: 2, Health: 1, MaxHealth: 1},
		{ID: "defender", OwnerID: "b", TileID: "enemy", ArmyID: "army-b", Domain: engine.DomainLand, Movement: 2, Attack: 2, Health: 1, MaxHealth: 1},
	}
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Phase.ActivePlayerIndex = 0
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projection.Rules == nil || len(projection.Rules.Units) == 0 || len(projection.Rules.Production) == 0 || len(projection.Rules.Advances) == 0 || len(projection.Rules.Governments) == 0 || len(projection.Rules.TerrainImprovements) == 0 || projection.Rules.Movement == nil {
		t.Fatalf("rules projection incomplete: %#v", projection.Rules)
	}
	for _, command := range []string{engine.CommandPass, engine.CommandSetResearch, engine.CommandSetGovernment, engine.CommandProposeAgreement, engine.CommandQueueProduction, engine.CommandFoundCity, engine.CommandArmyCreate, engine.CommandMoveUnit, engine.CommandEmbark, engine.CommandCombat, engine.CommandEstablishEmbassy, engine.CommandStealTechnology, engine.CommandTerraform, engine.CommandBuildImprovement} {
		if !hasAction(projection.LegalActions, command) {
			t.Fatalf("legal action %q missing from %#v", command, projection.LegalActions)
		}
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("HIDDEN_UNIT_TOKEN")) || bytes.Contains(encoded, []byte("SECRET_TERRAIN_TOKEN")) {
		t.Fatalf("projection leaked hidden data: %s", encoded)
	}
	state.Phase.ActivePlayerIndex = 1
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.LegalActions) != 0 {
		t.Fatalf("wrong-turn projection exposed actions: %#v", projection.LegalActions)
	}
}

func TestTask12ProjectEventExposesStableEnvelopeOnly(t *testing.T) {
	projector := NewProjector("a")
	event, err := projector.ProjectEvent(events.Viewer{GameID: 1, PlayerID: 1}, events.Event{Type: engine.EventCombatResolved, Sequence: 4, Revision: 3, Ordinal: 2, PayloadJSON: []byte(`{"target":"HIDDEN_UNIT_TOKEN","terrain":"SECRET_TERRAIN_TOKEN"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(event, []byte("HIDDEN_UNIT_TOKEN")) || bytes.Contains(event, []byte("SECRET_TERRAIN_TOKEN")) {
		t.Fatalf("event projection leaked payload: %s", event)
	}
	var decoded struct {
		Type     string `json:"type"`
		Sequence int64  `json:"sequence"`
		Revision int64  `json:"revision"`
		Ordinal  int    `json:"ordinal"`
	}
	if err := json.Unmarshal(event, &decoded); err != nil || decoded.Type != engine.EventCombatResolved || decoded.Sequence != 4 || decoded.Revision != 3 || decoded.Ordinal != 2 {
		t.Fatalf("event projection = %#v err=%v raw=%s", decoded, err, event)
	}
}

func TestRetailUnitTargetActionsUsePublicTargetMarkers(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "a", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "b", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Cities = nil
	state.Units = []engine.Unit{
		{ID: "lawyer", OwnerID: "a", TileID: "a", Movement: 1, Effects: []engine.Effect{{ID: "sue", Kind: engine.EffectSpecialSue}, {ID: "expel", Kind: engine.EffectSpecialExpel}}},
		{ID: "guard", OwnerID: "b", TileID: "b"},
	}
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Phase.ActivePlayerIndex = 0
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if hasAction(projection.LegalActions, engine.CommandSue) || hasAction(projection.LegalActions, engine.CommandExpel) {
		t.Fatalf("unit target action leaked without target markers: %#v", projection.LegalActions)
	}

	state.Units[1].Effects = []engine.Effect{{ID: "sued", Kind: engine.EffectCanBeSued}, {ID: "expelled", Kind: engine.EffectCanBeExpelled}}
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !hasAction(projection.LegalActions, engine.CommandSue) || !hasAction(projection.LegalActions, engine.CommandExpel) {
		t.Fatalf("unit target actions missing with public markers: %#v", projection.LegalActions)
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte("can_be_sued")) || !bytes.Contains(encoded, []byte("can_be_expelled")) {
		t.Fatalf("public target markers not projected: %s", encoded)
	}
}

func TestRetailSueFranchiseActionProjectsForOwnFranchisedCity(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "a", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "b", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Cities = []engine.City{{ID: "city-a", OwnerID: "a", TileID: "b", Effects: []engine.Effect{{ID: "franchise:b", Kind: engine.EffectFranchise}}}}
	state.Units = []engine.Unit{{ID: "lawyer", OwnerID: "a", TileID: "a", Movement: 1, Effects: []engine.Effect{{ID: "sue", Kind: engine.EffectSpecialSue}}}}
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Phase.ActivePlayerIndex = 0
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !hasAction(projection.LegalActions, engine.CommandSue) {
		t.Fatalf("sue franchise action missing from %#v", projection.LegalActions)
	}
}

func TestTerrainActionsUseRetailAIImprovementListOrder(t *testing.T) {
	projection := Projection{
		Diplomacy: &DiplomacyProjection{PublicWorks: 10},
		Tiles: []TileProjection{
			{ID: "home", X: 0, Y: 0},
			{ID: "work", X: 1, Y: 0, Terrain: engine.TerrainPlains},
		},
		Rules: &RulesProjection{
			TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{
				{ID: "TILEIMP_ADVANCED_FARMS", BuildTurns: 1, PublicWorksCost: 1},
				{ID: "TILEIMP_FARMS", BuildTurns: 1, PublicWorksCost: 1},
			},
			AI: &rules.RuntimeAI{ImprovementLists: []rules.RuntimeAIList{{ID: "IMPROVEMENT_LIST_GROWTH", Items: []string{"TILEIMP_FARMS"}}}},
		},
	}
	unit := UnitProjection{ID: "worker", OwnerID: "a", TileID: "home", Movement: 1, Effects: []engine.Effect{{ID: "terraform", Kind: engine.EffectSpecialTerraform}}}
	actions := terrainActions(projection, projectedTiles(projection), unit)
	for _, action := range actions {
		if action.Command != engine.CommandBuildImprovement || action.TargetID != "work" {
			continue
		}
		var payload struct {
			ImprovementID string `json:"improvement_id"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.ImprovementID != "TILEIMP_FARMS" {
			t.Fatalf("terrain action payload = %+v, want retail AI list improvement", payload)
		}
		return
	}
	t.Fatalf("missing build-improvement action: %#v", actions)
}

func hasAction(actions []ActionProjection, command string) bool {
	for _, action := range actions {
		if action.Command == command && len(action.Payload) > 0 && action.Route != "" {
			return true
		}
	}
	return false
}

func TestProjectionExposesOnlyCurrentActorPlanningStatusInPlanningModes(t *testing.T) {
	queuedCommand := func(actor string) engine.TurnPolicyCommand {
		return engine.TurnPolicyCommand{Kind: turns.CommandAction, ActorID: actor, CommandID: "command-" + actor, IdempotencyKey: "receipt-" + actor, Payload: json.RawMessage(`{}`)}
	}
	projectionJSON := func(t *testing.T, viewer string, state engine.State) map[string]any {
		t.Helper()
		state, err := Update(state)
		if err != nil {
			t.Fatal(err)
		}
		encoded := mustStateJSON(t, state)
		restarted, err := engine.DecodeCanonicalStateJSON(encoded)
		if err != nil {
			t.Fatal(err)
		}
		projection, err := Project(restarted, viewer)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(projection)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	}
	tests := []struct {
		name    string
		viewer  string
		mutate  func(engine.State) engine.State
		present bool
		want    string
	}{
		{name: "sequential omits status", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeSequential
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}},
		{name: "long-turn omits status", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeLongTurn
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}},
		{name: "simultaneous queued", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}, present: true, want: "queued"},
		{name: "closed beats queued", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}, present: true, want: "closed"},
		{name: "other actors keep eligible viewer open", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("b")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}, present: true, want: "open"},
		{name: "hybrid planning open", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeHybrid
			state.Policy.Stage = turns.StagePlanning
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("b")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}, present: true, want: "open"},
		{name: "hybrid conflict omits status", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeHybrid
			state.Policy.Stage = turns.StageConflict
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}},
		{name: "unknown viewer omits status", viewer: "viewer-is-not-a-player", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("b")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}},
		{name: "terminal omits status", mutate: func(state engine.State) engine.State {
			state.Terminal.Ended = true
			state.Terminal.Reason = "conquest"
			state.Terminal.WinningPlayerID = "a"
			return state
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := fixtureState()
			state.Phase.Mode = turns.ModeSimultaneous
			state.Phase.ModeVersion = turns.ContractVersion
			state.Policy.Stage = turns.StagePlanning
			viewer := tc.viewer
			if viewer == "" {
				viewer = "a"
			}
			decoded := projectionJSON(t, viewer, tc.mutate(state))
			value, ok := decoded["planning_status"]
			if tc.present {
				if !ok {
					t.Fatalf("planning_status omitted from projection JSON: %#v", decoded)
				}
				if got := value.(string); got != tc.want {
					t.Fatalf("planning_status = %q, want %q", got, tc.want)
				}
				return
			}
			if ok {
				t.Fatalf("planning_status = %#v, want omitted; projection = %#v", value, decoded)
			}
		})
	}
}

func mustStateJSON(t *testing.T, state engine.State) []byte {
	t.Helper()
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
