package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailUpgradeUnitAppliesTargetDefinitionFromUpgradeChain(t *testing.T) {
	state := fixtureState()
	state.Nations = []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_UPGRADE"}}}
	state.Units = []Unit{{ID: "unit-old", TypeID: "UNIT_OLD", OwnerID: "player-a", TileID: "0,0", Domain: DomainLand, Movement: 1, MaxMovement: 1, Health: 5, MaxHealth: 10, Attack: 1, Effects: []Effect{{ID: "old", Kind: EffectCanPillage}}}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_OLD", Kind: ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{
		{ID: "UNIT_OLD", Domain: DomainLand, Movement: 1, Health: 10, UpgradeTo: []string{"UNIT_NEW"}},
		{ID: "UNIT_NEW", Domain: DomainLand, Movement: 3, Attack: 7, Defense: 4, Strength: 6, Armor: 2, Firepower: 2, Health: 20, Prerequisites: []string{"ADVANCE_UPGRADE"}, Effects: []rules.RuntimeStarterEffect{{ID: "new", Kind: EffectCanEntrench}}},
	}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	payload, _ := json.Marshal(map[string]string{"unit_id": "unit-old", "upgrade_to": "UNIT_NEW"})
	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "upgrade", ActorPlayerID: "player-a", Type: CommandUpgradeUnit, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("UpgradeUnit() error = %v", err)
	}
	if len(events) != 1 || events[0].Type != EventUnitUpgraded {
		t.Fatalf("events = %#v", events)
	}
	unit, ok := unitByID(next, "unit-old")
	if !ok {
		t.Fatalf("upgraded unit missing: %#v", next.Units)
	}
	if unit.TypeID != "UNIT_NEW" || unit.Attack != 7 || unit.Defense != 4 || unit.MaxMovement != 3 || unit.Movement != 0 || unit.Health != 10 || unit.MaxHealth != 20 || len(unit.Effects) != 1 || unit.Effects[0].Kind != EffectCanEntrench {
		t.Fatalf("upgraded unit = %+v", *unit)
	}
}

func TestRetailUpgradeUnitRejectsUnknownChain(t *testing.T) {
	state := fixtureState()
	state.Units = []Unit{{ID: "unit-old", TypeID: "UNIT_OLD", OwnerID: "player-a", TileID: "0,0", Health: 10, MaxHealth: 10}}
	runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Production: []rules.RuntimeProduction{{ID: "UNIT_OLD", Kind: ProductionUnit}}, Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_OLD", Domain: DomainLand, Movement: 1, Health: 10, UpgradeTo: []string{"UNIT_NEW"}}, {ID: "UNIT_OTHER", Domain: DomainLand, Movement: 1, Health: 10}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	payload, _ := json.Marshal(map[string]string{"unit_id": "unit-old", "upgrade_to": "UNIT_OTHER"})
	if _, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bad-upgrade", ActorPlayerID: "player-a", Type: CommandUpgradeUnit, Payload: payload}, state.Rules, fixtureSources()); err == nil {
		t.Fatal("UpgradeUnit() succeeded for target outside upgrade chain")
	}
}
