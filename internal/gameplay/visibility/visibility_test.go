package visibility

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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

func TestTreasuryGoldProjectsOnlyToOwner(t *testing.T) {
	state := fixtureState()
	state.Nations = []engine.Nation{{ID: "a", Gold: 42}, {ID: "b", Gold: 99}}
	owner, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if owner.Diplomacy == nil || owner.Diplomacy.Gold != 42 {
		t.Fatalf("owner diplomacy gold = %#v", owner.Diplomacy)
	}
	viewer, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	if viewer.Diplomacy == nil || viewer.Diplomacy.Gold != 99 {
		t.Fatalf("viewer saw wrong treasury gold = %#v", viewer.Diplomacy)
	}
}

func TestSocialPolicyProjectsOwnSliderValuesAndLegalAction(t *testing.T) {
	workday, wages, rations := 1, -1, 0
	projection := Projection{Rules: &RulesProjection{}, Diplomacy: &DiplomacyProjection{Workday: &workday, Wages: &wages, Rations: &rations}}
	actions := governmentActions(projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandSetSocialPolicy && action.EntityID == "a" {
			if !bytes.Contains(action.Payload, []byte(`"workday":1`)) || !bytes.Contains(action.Payload, []byte(`"wages":-1`)) || !bytes.Contains(action.Payload, []byte(`"rations":0`)) {
				t.Fatalf("social policy payload = %s", action.Payload)
			}
			return
		}
	}
	t.Fatalf("social policy action missing: %#v", actions)
}

func TestTradeRouteActionsSuppressStaleCityMemory(t *testing.T) {
	projection := Projection{
		Cities: []CityProjection{
			{ID: "old-source", OwnerID: "a", TileID: "old-tile", Stale: true, Observed: 1},
			{ID: "fresh-source", OwnerID: "a", TileID: "fresh-tile", Observed: 3},
			{ID: "old-destination", OwnerID: "b", TileID: "old-dest", Stale: true, Observed: 1},
			{ID: "fresh-destination", OwnerID: "b", TileID: "fresh-dest", Observed: 3},
		},
		Diplomacy: &DiplomacyProjection{Relations: []DiplomacyRelationProjection{{TargetPlayerID: "b", Trade: true}}},
	}
	actions := tradeRouteActions(engine.State{}, projection, "a")
	if len(actions) != 1 {
		t.Fatalf("trade route actions = %#v, want only fresh source/destination", actions)
	}
	if actions[0].EntityID != "fresh-source" || actions[0].TargetID != "fresh-destination" {
		t.Fatalf("stale city memory leaked into trade route action: %#v", actions[0])
	}
}

func TestTerrainImprovementVisionRangeRevealsTiles(t *testing.T) {
	state := engine.State{
		Players:  []engine.Player{{ID: "a", NationID: "na", Lifecycle: engine.PlayerActive}},
		Tiles:    []engine.Tile{{ID: "center", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "near", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "mid", X: 2, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "far", X: 3, Y: 0, Terrain: engine.TerrainPlains, Open: true}},
		Overlays: []engine.TileOverlay{{ID: "watch", TileID: "center", Kind: "improvement", OwnerID: "a", ImprovementID: "TILEIMP_WATCH", VisionRange: 3}},
	}
	visible := visibleTiles(state)
	if !visible["a"]["far"] {
		t.Fatalf("vision-range improvement did not reveal far tile: %#v", visible["a"])
	}
}

func TestVisibleCitiesProjectCreateTradeRouteAction(t *testing.T) {
	state := fixtureState()
	state.Players = []engine.Player{{ID: "a", NationID: "na", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "nb", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Nations = []engine.Nation{{ID: "na"}, {ID: "nb"}}
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = nil
	state.Cities = []engine.City{{ID: "home", OwnerID: "a", TileID: "0,0"}, {ID: "foreign", OwnerID: "b", TileID: "1,0"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "trade", Trade: true}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandCreateTradeRoute && action.EntityID == "home" && action.TargetID == "foreign" && bytes.Contains(action.Payload, []byte(`"source_city_id":"home"`)) {
			return
		}
	}
	t.Fatalf("trade route action missing: %#v", projection.LegalActions)
}

func TestVisibleCitiesSuppressForeignTradeRouteActionWithoutTradeAgreement(t *testing.T) {
	state := fixtureState()
	state.Players = []engine.Player{{ID: "a", NationID: "na", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "nb", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Nations = []engine.Nation{{ID: "na"}, {ID: "nb"}}
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = nil
	state.Cities = []engine.City{{ID: "home", OwnerID: "a", TileID: "0,0"}, {ID: "foreign", OwnerID: "b", TileID: "1,0"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "peace"}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandCreateTradeRoute && action.TargetID == "foreign" {
			t.Fatalf("foreign trade route action leaked without trade agreement: %#v", action)
		}
	}
}

func TestVisibleCitiesSuppressTradeRouteActionsAtGovernmentCaps(t *testing.T) {
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: engine.DomainLand, Movement: 1, Health: 1}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: engine.ProductionUnit, Cost: 1}}, Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_LIMITED_TRADE", MaxOutgoingTrade: 1, MaxIncomingTrade: 1}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = engine.RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Players = []engine.Player{{ID: "a", NationID: "na", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"a"}
	state.Nations = []engine.Nation{{ID: "na", GovernmentID: "GOVERNMENT_LIMITED_TRADE"}}
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "2,0", X: 2, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = nil
	state.Cities = []engine.City{{ID: "home", OwnerID: "a", TileID: "0,0"}, {ID: "target", OwnerID: "a", TileID: "1,0"}, {ID: "other", OwnerID: "a", TileID: "2,0"}}
	state.TradeRoutes = []engine.TradeRoute{{ID: "existing", SourceCityID: "home", DestinationCityID: "other"}}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandCreateTradeRoute && action.EntityID == "home" && action.TargetID == "target" {
			t.Fatalf("trade route action leaked at outgoing cap: %#v", action)
		}
	}
}

func TestRushBuyActionUsesGovernmentModifier(t *testing.T) {
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: engine.DomainLand, Movement: 1, Health: 1}}, Production: []rules.RuntimeProduction{{ID: "UNIT_SETTLER", Kind: engine.ProductionUnit, Cost: 1}}, Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_RETAIL", BuildingRushModifier: 4}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = engine.RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Players = []engine.Player{{ID: "a", NationID: "na", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"a"}
	state.Nations = []engine.Nation{{ID: "na", GovernmentID: "GOVERNMENT_RETAIL", Gold: 383}}
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = nil
	state.Cities = []engine.City{{ID: "city", OwnerID: "a", TileID: "0,0", StoredProduction: 40, Queue: []engine.ProductionItem{{ID: "IMPROVE_SYNTHETIC", Kind: engine.ProductionBuilding, Cost: 100}}}}

	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandRushBuy && action.EntityID == "city" {
			t.Fatalf("rush-buy action leaked below government-modified cost: %#v", action)
		}
	}
}

func TestProjectionSuppressesCantBuildRuntimeUnitProduction(t *testing.T) {
	state := fixtureState()
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_OK", Kind: engine.ProductionUnit}, {ID: "UNIT_CITY", Kind: engine.ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_OK", Domain: engine.DomainLand, Movement: 1, Health: 1}, {ID: "UNIT_CITY", Domain: engine.DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "cant-build", Kind: engine.EffectCantBuild}}}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = engine.RulesPin{Version: rules.CTP2CompatibilityVersion, Definitions: definitions, ContentHash: hash}
	state.Cities = []engine.City{{ID: "city-a", OwnerID: "a", TileID: "0,0", Population: 2}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	seenOK := false
	for _, action := range projection.LegalActions {
		if action.Command != engine.CommandQueueProduction {
			continue
		}
		if action.TargetID == "UNIT_CITY" {
			t.Fatalf("cant-build unit production action leaked: %#v", action)
		}
		if action.TargetID == "UNIT_OK" {
			seenOK = true
		}
	}
	if !seenOK {
		t.Fatalf("normal unit production action missing: %#v", projection.LegalActions)
	}
}

func TestProjectionSuppressesNuclearAttackProductionAfterEliminateNukes(t *testing.T) {
	state := projectionStateWithRuntime(t, rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_NUKE", Kind: engine.ProductionUnit}, {ID: "UNIT_WORKER", Kind: engine.ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{
		{ID: "UNIT_NUKE", Domain: engine.DomainAir, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "nuke", Kind: engine.EffectNuclearAttack}}},
		{ID: "UNIT_WORKER", Domain: engine.DomainLand, Movement: 1, Health: 1},
	}})
	state.Nations = []engine.Nation{{ID: "a", Effects: []engine.Effect{{ID: "eliminate", Kind: engine.EffectEliminateNukes}}}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_NUKE") {
		t.Fatalf("nuclear-attack production leaked after eliminate-nukes: %#v", projection.LegalActions)
	}
	if !productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_WORKER") {
		t.Fatalf("fallback worker production missing after nuke suppression: %#v", projection.LegalActions)
	}
}

func TestProjectionSuppressesWrongGovernmentUnitProduction(t *testing.T) {
	state := projectionStateWithRuntime(t, rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_CLERIC", Kind: engine.ProductionUnit}, {ID: "UNIT_WORKER", Kind: engine.ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{
		{ID: "UNIT_CLERIC", Domain: engine.DomainLand, Movement: 1, Health: 1, RequiredGovernmentID: "GOVERNMENT_THEOCRACY"},
		{ID: "UNIT_WORKER", Domain: engine.DomainLand, Movement: 1, Health: 1},
	}})
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_CLERIC") {
		t.Fatalf("wrong-government unit production leaked: %#v", projection.LegalActions)
	}
	if !productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_WORKER") {
		t.Fatalf("fallback worker production missing after government suppression: %#v", projection.LegalActions)
	}

	state.Nations = []engine.Nation{{ID: "a", GovernmentID: "GOVERNMENT_THEOCRACY"}}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_CLERIC") {
		t.Fatalf("matching-government unit production missing: %#v", projection.LegalActions)
	}
}

func TestProjectionSuppressesNoSlavesAndSlaverExclusiveProduction(t *testing.T) {
	state := projectionStateWithRuntime(t, rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_SLAVER", Kind: engine.ProductionUnit}, {ID: "UNIT_WORKER", Kind: engine.ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{
		{ID: "UNIT_SLAVER", Domain: engine.DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "slave", Kind: engine.EffectSpecialSlaveRaid}}},
		{ID: "UNIT_WORKER", Domain: engine.DomainLand, Movement: 1, Health: 1},
	}})
	state.Units = []engine.Unit{{ID: "abolitionist", OwnerID: "a", TileID: "0,0", Effects: []engine.Effect{{ID: "no-slaves", Kind: engine.EffectNoSlaves}}}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_SLAVER") {
		t.Fatalf("slaver production leaked while no-slaves unit exists: %#v", projection.LegalActions)
	}
	if !productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_WORKER") {
		t.Fatalf("fallback worker production missing after slaver suppression: %#v", projection.LegalActions)
	}

	state = projectionStateWithRuntime(t, rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_ABOLITIONIST", Kind: engine.ProductionUnit}, {ID: "UNIT_WORKER", Kind: engine.ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{
		{ID: "UNIT_ABOLITIONIST", Domain: engine.DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "no-slaves", Kind: engine.EffectNoSlaves}}},
		{ID: "UNIT_WORKER", Domain: engine.DomainLand, Movement: 1, Health: 1},
	}})
	state.Units = []engine.Unit{{ID: "slaver", OwnerID: "a", TileID: "0,0", Effects: []engine.Effect{{ID: "slave", Kind: engine.EffectSpecialSlaveRaid}}}}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_ABOLITIONIST") {
		t.Fatalf("no-slaves production leaked while slaver exists: %#v", projection.LegalActions)
	}
	if !productionActionTarget(projection, engine.CommandQueueProduction, "UNIT_WORKER") {
		t.Fatalf("fallback worker production missing after no-slaves suppression: %#v", projection.LegalActions)
	}
}

func projectionStateWithRuntime(t *testing.T, runtime rules.RuntimeRules) engine.State {
	t.Helper()
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = engine.RulesPin{Version: rules.CTP2CompatibilityVersion, Definitions: definitions, ContentHash: hash}
	state.Cities = []engine.City{{ID: "city-a", OwnerID: "a", TileID: "0,0", Population: 2}}
	return state
}

func productionActionTarget(projection Projection, command, targetID string) bool {
	for _, action := range projection.LegalActions {
		if action.Command == command && action.TargetID == targetID {
			return true
		}
	}
	return false
}

func TestProjectionExposesRetailUpgradeActionForOwnedUnit(t *testing.T) {
	state := fixtureState()
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_OLD", Kind: engine.ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_OLD", Domain: engine.DomainLand, Movement: 1, Health: 10, UpgradeTo: []string{"UNIT_NEW"}}, {ID: "UNIT_NEW", Domain: engine.DomainLand, Movement: 2, Health: 20, Prerequisites: []string{"ADVANCE_UPGRADE"}}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = engine.RulesPin{Version: rules.CTP2CompatibilityVersion, Definitions: definitions, ContentHash: hash}
	state.Nations = []engine.Nation{{ID: "a", Advances: []string{"ADVANCE_UPGRADE"}}, {ID: "b"}}
	state.Units = []engine.Unit{{ID: "unit-old", TypeID: "UNIT_OLD", OwnerID: "a", TileID: "0,0", Domain: engine.DomainLand, Movement: 1, MaxMovement: 1, Health: 10, MaxHealth: 10}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandUpgradeUnit && action.EntityID == "unit-old" && action.TargetID == "UNIT_NEW" && bytes.Contains(action.Payload, []byte(`"upgrade_to":"UNIT_NEW"`)) {
			return
		}
	}
	t.Fatalf("upgrade action missing: %#v", projection.LegalActions)
}

func TestProjectionSuppressesWrongGovernmentUpgradeAction(t *testing.T) {
	state := fixtureState()
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: engine.DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_OLD", Kind: engine.ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_OLD", Domain: engine.DomainLand, Movement: 1, Health: 10, UpgradeTo: []string{"UNIT_FASCIST"}}, {ID: "UNIT_FASCIST", Domain: engine.DomainLand, Movement: 2, Health: 20, RequiredGovernmentID: "GOVERNMENT_FASCISM"}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = engine.RulesPin{Version: rules.CTP2CompatibilityVersion, Definitions: definitions, ContentHash: hash}
	state.Nations = []engine.Nation{{ID: "a"}, {ID: "b"}}
	state.Units = []engine.Unit{{ID: "unit-old", TypeID: "UNIT_OLD", OwnerID: "a", TileID: "0,0", Domain: engine.DomainLand, Movement: 1, MaxMovement: 1, Health: 10, MaxHealth: 10}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if productionActionTarget(projection, engine.CommandUpgradeUnit, "UNIT_FASCIST") {
		t.Fatalf("wrong-government upgrade action leaked: %#v", projection.LegalActions)
	}
	state.Nations = []engine.Nation{{ID: "a", GovernmentID: "GOVERNMENT_FASCISM"}, {ID: "b"}}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !productionActionTarget(projection, engine.CommandUpgradeUnit, "UNIT_FASCIST") {
		t.Fatalf("matching-government upgrade action missing: %#v", projection.LegalActions)
	}
}

func TestDiplomacyProjectsPeaceProposalAtWar(t *testing.T) {
	state := fixtureState()
	state.Players = []engine.Player{{ID: "a", NationID: "na", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "nb", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Nations = []engine.Nation{{ID: "na"}, {ID: "nb"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true}}
	state.Policy.Wars = []engine.TurnPolicyWar{{A: "a", B: "b"}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandProposeAgreement && action.TargetID == "b" && bytes.Contains(action.Payload, []byte(`"kind":"peace"`)) {
			return
		}
	}
	t.Fatalf("peace proposal action missing: %#v", projection.LegalActions)
}

func TestDiplomacyProjectsTradeProposalBeforeForeignTradeRoutes(t *testing.T) {
	state := fixtureState()
	state.Players = []engine.Player{{ID: "a", NationID: "na", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "nb", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Nations = []engine.Nation{{ID: "na"}, {ID: "nb"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "peace"}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandProposeAgreement && action.TargetID == "b" && bytes.Contains(action.Payload, []byte(`"kind":"trade"`)) {
			return
		}
	}
	t.Fatalf("trade proposal action missing: %#v", projection.LegalActions)
}

func TestSpecialistsProjectOwnerAssignmentsAndLegalAction(t *testing.T) {
	projection := Projection{Cities: []CityProjection{{ID: "city-a", OwnerID: "a", Population: 2, Specialists: map[string]int{"POP_SCIENTIST": 1}}}}
	actions := specialistActions(projection, "a")
	if len(actions) != 1 || actions[0].Command != engine.CommandSetSpecialists || !bytes.Contains(actions[0].Payload, []byte(`"POP_SCIENTIST":1`)) {
		t.Fatalf("specialist actions = %#v", actions)
	}
}

func TestReadinessProjectsOwnValueAndLegalActions(t *testing.T) {
	projection := Projection{Rules: &RulesProjection{}, Diplomacy: &DiplomacyProjection{Readiness: "peace"}}
	actions := governmentActions(projection, "a")
	if !hasAction(actions, engine.CommandSetReadiness) {
		t.Fatalf("readiness action missing: %#v", actions)
	}
	for _, action := range actions {
		if action.Command == engine.CommandSetReadiness && action.TargetID == "peace" {
			t.Fatalf("current readiness was projected redundantly: %#v", actions)
		}
		if action.Command == engine.CommandSetReadiness && action.TargetID == "war" && !bytes.Contains(action.Payload, []byte(`"readiness":"war"`)) {
			t.Fatalf("readiness payload = %s", action.Payload)
		}
	}
}

func TestScienceRateLegalActionsTreatUnsetAsDefaultHundred(t *testing.T) {
	projection := Projection{Rules: &RulesProjection{}, Diplomacy: &DiplomacyProjection{}}
	actions := governmentActions(projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandSetScienceRate && action.TargetID == "100" {
			t.Fatalf("default science percent should already be 100; legal actions=%#v", actions)
		}
	}
}

func TestProductionActionsSuppressAlreadyFundedRushBuy(t *testing.T) {
	projection := Projection{
		Diplomacy: &DiplomacyProjection{Gold: 100},
		Cities:    []CityProjection{{ID: "city-a", OwnerID: "a", StoredProduction: 100, Queue: []engine.ProductionItem{{ID: "IMPROVE_SYNTHETIC", Kind: engine.ProductionBuilding, Cost: 100}}}},
	}
	for _, action := range productionActions(engine.State{}, projection, "a") {
		if action.Command == engine.CommandRushBuy {
			t.Fatalf("already funded queue should not expose rush buy: %#v", action)
		}
	}
}

func TestProductionActionsIncludeAffordableRushBuy(t *testing.T) {
	projection := Projection{
		Diplomacy: &DiplomacyProjection{Gold: 96},
		Cities:    []CityProjection{{ID: "city-a", OwnerID: "a", StoredProduction: 40, Queue: []engine.ProductionItem{{ID: "IMPROVE_SYNTHETIC", Kind: engine.ProductionBuilding, Cost: 100}}}},
	}
	actions := productionActions(engine.State{}, projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandRushBuy && action.EntityID == "city-a" && action.TargetID == "IMPROVE_SYNTHETIC" && bytes.Contains(action.Payload, []byte(`"city_id":"city-a"`)) {
			return
		}
	}
	t.Fatalf("rush-buy action missing: %#v", actions)
}

func TestGovernmentActionsIncludeRetailAIStrategyPublicWorksPercent(t *testing.T) {
	projection := Projection{
		Rules:     &RulesProjection{AI: &rules.RuntimeAI{Strategies: []rules.RuntimeAIStrategy{{ID: "STRATEGY_DEFAULT", PublicWorksPercent: 30}}}},
		Diplomacy: &DiplomacyProjection{},
	}
	actions := governmentActions(projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandSetPublicWorksRate && action.TargetID == "30" && bytes.Contains(action.Payload, []byte(`"public_works_percent":30`)) {
			return
		}
	}
	t.Fatalf("strategy public works percent action missing: %#v", actions)
}

func TestGovernmentActionsIncludeRetailAIStrategySciencePercent(t *testing.T) {
	projection := Projection{
		Rules:     &RulesProjection{AI: &rules.RuntimeAI{Strategies: []rules.RuntimeAIStrategy{{ID: "STRATEGY_DEFAULT", SciencePercent: 60}}}},
		Diplomacy: &DiplomacyProjection{},
	}
	actions := governmentActions(projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandSetScienceRate && action.TargetID == "60" && bytes.Contains(action.Payload, []byte(`"science_percent":60`)) {
			return
		}
	}
	t.Fatalf("strategy science percent action missing: %#v", actions)
}

func TestGovernmentActionsIncludeRetailAIStrategySocialPolicy(t *testing.T) {
	projection := Projection{
		Rules: &RulesProjection{
			AI:          &rules.RuntimeAI{Strategies: []rules.RuntimeAIStrategy{{ID: "STRATEGY_DEFAULT", SliderElements: []rules.RuntimeAISliderElement{{Kind: "Food", Delta: 3}}}}},
			Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_DEMOCRACY", WorkdayExpectation: 0, WagesExpectation: 1, RationsExpectation: 0}},
		},
		Diplomacy: &DiplomacyProjection{GovernmentID: "GOVERNMENT_DEMOCRACY"},
		Cities:    []CityProjection{{ID: "city-a", OwnerID: "a", Population: 1, StoredFood: 20, Yields: &engine.Yield{Food: 20, Production: 20, Gold: 20}}},
	}
	actions := governmentActions(projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandSetSocialPolicy && action.TargetID == "social-policy" && bytes.Contains(action.Payload, []byte(`"workday":2`)) && bytes.Contains(action.Payload, []byte(`"wages":2`)) && bytes.Contains(action.Payload, []byte(`"rations":-2`)) {
			return
		}
	}
	t.Fatalf("strategy social policy action missing: %#v", actions)
}

func TestGovernmentActionsKeepCurrentSocialPolicyWithoutRetailSliderCityFit(t *testing.T) {
	projection := Projection{
		Rules: &RulesProjection{
			AI:          &rules.RuntimeAI{Strategies: []rules.RuntimeAIStrategy{{ID: "STRATEGY_DEFAULT", SliderElements: []rules.RuntimeAISliderElement{{Kind: "Food", Delta: 3}}}}},
			Governments: []rules.RuntimeGovernmentDefinition{{ID: "GOVERNMENT_DEMOCRACY", WorkdayExpectation: 0, WagesExpectation: 1, RationsExpectation: 0}},
		},
		Diplomacy: &DiplomacyProjection{GovernmentID: "GOVERNMENT_DEMOCRACY"},
	}
	actions := governmentActions(projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandSetSocialPolicy && action.TargetID == "social-policy" {
			if bytes.Contains(action.Payload, []byte(`"workday":2`)) || bytes.Contains(action.Payload, []byte(`"wages":2`)) || bytes.Contains(action.Payload, []byte(`"rations":-2`)) {
				t.Fatalf("unsafe strategy social policy applied without city fit evidence: %s", action.Payload)
			}
			return
		}
	}
	t.Fatalf("social policy action missing: %#v", actions)
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

func TestDiplomacyProjectionIncludesRetailRegard(t *testing.T) {
	state := fixtureState()
	state.Players = []engine.Player{{ID: "a", NationID: "nation-a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "nation-b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}, {ID: "c", NationID: "nation-c", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI}}
	state.Nations = []engine.Nation{{ID: "nation-a"}, {ID: "nation-b"}, {ID: "nation-c"}}
	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true, Treaty: "union"}, {ID: "a:c", PlayerAID: "a", PlayerBID: "c", Contact: true}}
	state.Policy.Wars = []engine.TurnPolicyWar{{A: "a", B: "c"}}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projection.Diplomacy == nil || len(projection.Diplomacy.Relations) != 2 {
		t.Fatalf("relations = %#v", projection.Diplomacy)
	}
	if projection.Diplomacy.Relations[0].TargetPlayerID != "b" || projection.Diplomacy.Relations[0].Regard != "love" {
		t.Fatalf("union relation = %#v, want love regard", projection.Diplomacy.Relations[0])
	}
	if projection.Diplomacy.Relations[1].TargetPlayerID != "c" || projection.Diplomacy.Relations[1].Regard != "hotwar" {
		t.Fatalf("war relation = %#v, want hotwar regard", projection.Diplomacy.Relations[1])
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
		{ID: "owned", OwnerID: "a", TileID: "0,0", Domain: engine.DomainLand, Movement: 3, CargoCapacity: 2, CargoDomains: []string{engine.DomainLand}, Upkeep: 4, ShieldHunger: 5, FoodHunger: 6, ZOC: true, Attack: 7, Defense: 5, Strength: 6, Armor: 2, Health: 8, MaxHealth: 10, VisibilityRange: 3, SettleCityTypeID: "UNIT_CITY", Effects: []engine.Effect{{ID: "settler", Kind: engine.EffectSettler}}},
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
	if owned.ID != "owned" || owned.Domain != engine.DomainLand || owned.Movement != 3 || owned.CargoCapacity != 2 || owned.Upkeep != 4 || owned.ShieldHunger != 5 || owned.FoodHunger != 6 || owned.SettleCityTypeID != "UNIT_CITY" || !owned.ZOC || owned.Attack != 7 || owned.Defense != 5 || owned.Strength != 6 || owned.Armor != 2 || owned.Health != 8 || owned.MaxHealth != 10 || owned.VisibilityRange != 3 || len(owned.Effects) != 1 {
		t.Fatalf("owned projection capabilities = %#v", owned)
	}
	enemy := projection.Units[1]
	if enemy.ID != "visible-enemy" || enemy.Domain != "" || enemy.Movement != 0 || enemy.Attack != 0 || enemy.Health != 0 || enemy.Upkeep != 0 || enemy.ShieldHunger != 0 || enemy.FoodHunger != 0 || enemy.SettleCityTypeID != "" || len(enemy.Effects) != 0 {
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
		{ID: "diplomat", OwnerID: "a", TileID: "home", Domain: engine.DomainLand, Movement: 2, Health: 1, MaxHealth: 1, Effects: []engine.Effect{{ID: "embassy", Kind: engine.EffectSpecialEmbassy}, {ID: "steal", Kind: engine.EffectSpecialSteal}, {ID: "cause", Kind: engine.EffectSpecialCauseUnhappiness}, {ID: "incite", Kind: engine.EffectSpecialInciteRevolution}, {ID: "indulgence", Kind: engine.EffectSpecialIndulgenceSales}, {ID: "uprising", Kind: engine.EffectSpecialSlaveUprising}}},
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
	for _, command := range []string{engine.CommandPass, engine.CommandSetResearch, engine.CommandSetGovernment, engine.CommandProposeAgreement, engine.CommandQueueProduction, engine.CommandFoundCity, engine.CommandArmyCreate, engine.CommandMoveUnit, engine.CommandEmbark, engine.CommandCombat, engine.CommandEstablishEmbassy, engine.CommandStealTechnology, engine.CommandCauseUnhappiness, engine.CommandInciteRevolution, engine.CommandIndulgenceSales, engine.CommandSlaveUprising, engine.CommandTerraform, engine.CommandBuildImprovement} {
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

func TestRetailVisibleForeignOwnerProjectsEstablishContact(t *testing.T) {
	state := fixtureState()
	state.Cities = nil
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "0,0", VisibilityRange: 2}, {ID: "foreign", OwnerID: "b", TileID: "1,0"}}
	state.Nations = []engine.Nation{{ID: "a"}, {ID: "b"}}
	state.Diplomacy = nil
	state.Phase.ActivePlayerOrder = []string{"a", "b"}
	state.Phase.ActivePlayerIndex = 0
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !hasAction(projection.LegalActions, engine.CommandEstablishContact) {
		t.Fatalf("visible foreign owner did not project contact action: %#v", projection.LegalActions)
	}

	state.Diplomacy = []engine.DiplomaticRelation{{ID: "a:b", PlayerAID: "a", PlayerBID: "b", Contact: true}}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if hasAction(projection.LegalActions, engine.CommandEstablishContact) {
		t.Fatalf("existing contact still projected establish action: %#v", projection.LegalActions)
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

func TestPublicWorksTerrainActionsDoNotRequireTerraformUnit(t *testing.T) {
	projection := Projection{
		Diplomacy: &DiplomacyProjection{PublicWorks: 10},
		Cities:    []CityProjection{{ID: "city-a", OwnerID: "a", TileID: "home"}},
		Tiles: []TileProjection{
			{ID: "home", X: 0, Y: 0, Terrain: engine.TerrainPlains},
			{ID: "work", X: 1, Y: 0, Terrain: engine.TerrainPlains},
		},
		Rules: &RulesProjection{TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_FARMS", BuildTurns: 1, PublicWorksCost: 1, AllowedTerrains: []string{engine.TerrainPlains}}}},
	}
	actions := publicWorksTerrainActions(projection, "a")
	for _, action := range actions {
		if action.Command != engine.CommandBuildImprovement || action.EntityID != "city-a" || action.TargetID != "work" {
			continue
		}
		if strings.Contains(string(action.Payload), "unit_id") {
			t.Fatalf("unitless public works action payload should omit unit_id: %s", action.Payload)
		}
		return
	}
	t.Fatalf("missing public works terrain action: %#v", actions)
}

func TestPublicWorksTerrainActionsSuppressPendingImprovement(t *testing.T) {
	projection := Projection{
		Diplomacy: &DiplomacyProjection{PublicWorks: 10},
		Cities:    []CityProjection{{ID: "city-a", OwnerID: "a", TileID: "home"}},
		Tiles: []TileProjection{
			{ID: "home", X: 0, Y: 0, Terrain: engine.TerrainPlains},
			{ID: "work", X: 1, Y: 0, Terrain: engine.TerrainPlains, Overlays: []TileOverlayProjection{{Kind: "improvement", ImprovementID: "TILEIMP_FARMS", Remaining: 1}}},
		},
		Rules: &RulesProjection{TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_FARMS", BuildTurns: 1, PublicWorksCost: 1, AllowedTerrains: []string{engine.TerrainPlains}}}},
	}
	for _, action := range publicWorksTerrainActions(projection, "a") {
		if action.TargetID == "work" {
			t.Fatalf("pending improvement tile was still actionable: %#v", action)
		}
	}
}

func TestTerrainActionsSuppressDisallowedTerrainImprovements(t *testing.T) {
	projection := Projection{
		Diplomacy: &DiplomacyProjection{PublicWorks: 10},
		Tiles: []TileProjection{
			{ID: "home", X: 0, Y: 0},
			{ID: "work", X: 1, Y: 0, Terrain: engine.TerrainForest},
		},
		Rules: &RulesProjection{TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_FARMS", BuildTurns: 1, PublicWorksCost: 1, AllowedTerrains: []string{"TERRAIN_PLAINS"}}}},
	}
	unit := UnitProjection{ID: "worker", OwnerID: "a", TileID: "home", Movement: 1, Effects: []engine.Effect{{ID: "terraform", Kind: engine.EffectSpecialTerraform}}}
	if actions := terrainActions(projection, projectedTiles(projection), unit); len(actions) != 0 {
		t.Fatalf("disallowed-terrain improvement action projected: %#v", actions)
	}
	projection.Tiles[1].Terrain = "TERRAIN_PLAINS"
	if actions := terrainActions(projection, projectedTiles(projection), unit); !hasAction(actions, engine.CommandBuildImprovement) {
		t.Fatalf("allowed-terrain improvement action missing: %#v", actions)
	}
}

func TestTerrainActionsSuppressAdvanceGatedImprovements(t *testing.T) {
	projection := Projection{
		Diplomacy: &DiplomacyProjection{PublicWorks: 10},
		Research:  &ResearchProjection{},
		Tiles: []TileProjection{
			{ID: "home", X: 0, Y: 0},
			{ID: "work", X: 1, Y: 0, Terrain: engine.TerrainPlains},
		},
		Rules: &RulesProjection{TerrainImprovements: []rules.RuntimeTerrainImprovementDefinition{{ID: "TILEIMP_FARMS", BuildTurns: 1, PublicWorksCost: 1, RequiredAdvanceID: "ADVANCE_AGRICULTURE"}}},
	}
	unit := UnitProjection{ID: "worker", OwnerID: "a", TileID: "home", Movement: 1, Effects: []engine.Effect{{ID: "terraform", Kind: engine.EffectSpecialTerraform}}}
	if actions := terrainActions(projection, projectedTiles(projection), unit); len(actions) != 0 {
		t.Fatalf("advance-gated terrain actions projected before advance: %#v", actions)
	}
	projection.Research.Advances = []string{"ADVANCE_AGRICULTURE"}
	if actions := terrainActions(projection, projectedTiles(projection), unit); !hasAction(actions, engine.CommandBuildImprovement) {
		t.Fatalf("advance-gated terrain action missing after advance: %#v", actions)
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
