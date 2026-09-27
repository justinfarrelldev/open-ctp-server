package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

func combatState() State {
	state := fixtureState()
	state.Tiles = []Tile{
		{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, OwnerID: "player-a", Open: true},
		{ID: "b", X: 1, Y: 0, Terrain: TerrainForest, OwnerID: "player-b", Open: true},
		{ID: "retreat", X: 2, Y: 0, Terrain: TerrainPlains, OwnerID: "player-b", Open: true},
	}
	state.Units = []Unit{
		{ID: "attacker-melee", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Attack: 8, Defense: 3, Strength: 4, Armor: 1, Health: 8, MaxHealth: 8},
		{ID: "attacker-ranged", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Attack: 2, Defense: 2, Ranged: 9, Strength: 3, Armor: 1, Health: 8, MaxHealth: 8},
		{ID: "defender", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Attack: 2, Defense: 4, Strength: 2, Armor: 1, Health: 4, MaxHealth: 4, Fortified: true},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"attacker-melee", "attacker-ranged"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"defender"}}}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "b", Effects: []Effect{{ID: "wall", Kind: EffectCombatDefense, Value: 1}}}}
	state.Nations = []Nation{{ID: "nation-a", Effects: []Effect{{ID: "banner", Kind: EffectCombatAttack, Value: 1}}}}
	state.Fog = []FogObservation{{PlayerID: "player-a", TileID: "a", Turn: 1}, {PlayerID: "player-a", TileID: "b", Turn: 1}}
	return state
}

func combatCommand() Command {
	payload, _ := json.Marshal(map[string]string{"attacker_army_id": "attackers", "defender_army_id": "defenders"})
	return Command{SchemaVersion: CommandSchemaV1, ID: "combat-1", ActorPlayerID: "player-a", Type: CommandCombat, Payload: payload}
}

func TestRetailFirepowerScalesCombatDamage(t *testing.T) {
	state := fixtureState()
	attacker := &Unit{ID: "attacker", OwnerID: "player-a", Attack: 100, Strength: 2, Armor: 1, Firepower: 3, Health: 10, MaxHealth: 10}
	defender := &Unit{ID: "defender", OwnerID: "player-b", Defense: 0, Strength: 1, Armor: 2, Health: 10, MaxHealth: 10}
	combatStrike(&state, fixtureSources(), []*Unit{attacker}, []*Unit{defender}, false, &CombatResult{})
	if got, want := defender.Health, 7; got != want {
		t.Fatalf("defender health after firepower strike = %d, want %d", got, want)
	}
}

func TestRetailBombardUsesCanBombardDomains(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 10, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "sea", Kind: EffectCanBombardSea}}},
		{ID: "land-first", OwnerID: "player-b", TileID: "retreat", Domain: DomainLand, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
		{ID: "sea-second", OwnerID: "player-b", TileID: "retreat", Domain: DomainSea, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
	}
	state.Armies = nil
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bombard-domain", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if land, ok := unitByID(next, "land-first"); !ok || land.Health != 5 {
		t.Fatalf("land target was damaged by sea-only bombard: unit=%+v ok=%v", land, ok)
	}
	if _, ok := unitByID(next, "sea-second"); ok {
		t.Fatalf("sea target survived sea-only bombard: %+v", next.Units)
	}
}

func TestRetailActiveDefenseRequiresCarrierEnabler(t *testing.T) {
	base := combatState()
	base.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Domain: DomainAir, Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "target", OwnerID: "player-b", TileID: "retreat", Defense: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "carrier-defense", OwnerID: "player-b", TileID: "retreat", Domain: DomainSea, Ranged: 100, ActiveDefenseRange: 3, Strength: 3, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "air", Kind: EffectDefendAir}, {ID: "requires-enabler", Kind: EffectActiveDefenseRequiresEnabler}}},
	}
	base.Armies = nil
	base.Cities = nil
	base.Fog = append(base.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})
	next, _, _, err := Apply(base, Command{SchemaVersion: CommandSchemaV1, ID: "no-enabler", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, base.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	bombarder, _ := unitByID(next, "bombarder")
	if bombarder.Health != 10 {
		t.Fatalf("carrier defense without enabler hit bombarder: %+v", bombarder)
	}

	withEnabler := base
	withEnabler.Units = append(append([]Unit(nil), base.Units...), Unit{ID: "enabler", OwnerID: "player-b", TransportID: "carrier-defense", Domain: DomainAir, Health: 1, MaxHealth: 1, Effects: []Effect{{ID: "enable", Kind: EffectEnableCarrierDefenses}}})
	next, _, _, err = Apply(withEnabler, Command{SchemaVersion: CommandSchemaV1, ID: "with-enabler", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, withEnabler.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	bombarder, _ = unitByID(next, "bombarder")
	if bombarder.Health >= 10 {
		t.Fatalf("carrier defense with enabler did not hit bombarder: %+v", bombarder)
	}
}

func TestRetailActiveDefenseHonorsDefendDomains(t *testing.T) {
	base := combatState()
	base.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "target", OwnerID: "player-b", TileID: "retreat", Defense: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
	}
	base.Armies = nil
	base.Cities = nil
	base.Fog = append(base.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})

	airOnly := base
	airOnly.Units = append(append([]Unit(nil), base.Units...), Unit{ID: "air-defense", OwnerID: "player-b", TileID: "retreat", Domain: DomainLand, Ranged: 100, ActiveDefenseRange: 3, Strength: 3, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "air", Kind: EffectDefendAir}}})
	next, _, _, err := Apply(airOnly, Command{SchemaVersion: CommandSchemaV1, ID: "air-defense", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, airOnly.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	bombarder, _ := unitByID(next, "bombarder")
	if bombarder.Health != 10 {
		t.Fatalf("air-only active defense hit land bombarder: %+v", bombarder)
	}

	landDefense := base
	landDefense.Units = append(append([]Unit(nil), base.Units...), Unit{ID: "land-defense", OwnerID: "player-b", TileID: "retreat", Domain: DomainLand, Ranged: 100, ActiveDefenseRange: 3, Strength: 3, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "land", Kind: EffectDefendLand}}})
	next, _, _, err = Apply(landDefense, Command{SchemaVersion: CommandSchemaV1, ID: "land-defense", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, landDefense.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	bombarder, _ = unitByID(next, "bombarder")
	if bombarder.Health >= 10 {
		t.Fatalf("land active defense did not hit land bombarder: %+v", bombarder)
	}
}

func TestRetailSingleUseBombarderIsConsumedAfterBombardment(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "b", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "missile", OwnerID: "player-a", TileID: "a", Domain: DomainAir, Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Defense: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "single-use", Kind: EffectSingleUse}, {ID: "can-bombard-land", Kind: EffectCanBombardLand}}},
		{ID: "target", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1, Defense: 100, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
	}
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "b", Turn: 1})
	payload := []byte(`{"unit_id":"missile","target_tile_id":"b"}`)
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "single-use-bombard", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Bombard() error = %v", err)
	}
	if _, ok := unitByID(next, "missile"); ok {
		t.Fatalf("single-use bombarder survived: %#v", next.Units)
	}
}

func TestRetailSingleUseAttackerIsConsumedAfterCombat(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "missile", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Attack: 100, Defense: 1, Strength: 2, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "single-use", Kind: EffectSingleUse}}},
		{ID: "defender", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Defense: 0, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"missile"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"defender"}}}
	next, _, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Combat() error = %v", err)
	}
	if _, ok := unitByID(next, "missile"); ok {
		t.Fatalf("single-use attacker survived combat: %#v", next.Units)
	}
}

func TestRetailCanCounterBombardReturnsFireAfterBombardment(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "b", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 1, Ranged: 1, BombardRange: 2, BombRounds: 1, Defense: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "counter", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 1, BombRounds: 1, Defense: 100, Strength: 10, Armor: 1, Firepower: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "counter", Kind: EffectCanCounterBombard}}},
	}
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "b", Turn: 1})
	payload := []byte(`{"unit_id":"bombarder","target_tile_id":"b"}`)
	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "counter-bombard", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Bombard() error = %v", err)
	}
	if _, ok := unitByID(next, "bombarder"); ok {
		t.Fatalf("counter-bombard should destroy bombarder: %#v", next.Units)
	}
	if _, ok := unitByID(next, "counter"); !ok {
		t.Fatalf("counter-bombarder should survive failed incoming bombard: %#v", next.Units)
	}
	if len(events) != 1 || events[0].Type != EventBombarded || !strings.Contains(string(events[0].Payload), "counter_bombard_hits") {
		t.Fatalf("counter-bombard event = %#v", events)
	}
}

func TestRetailFirepowerScalesBombardDamage(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 2, Armor: 1, Firepower: 3, Health: 10, MaxHealth: 10},
		{ID: "target", OwnerID: "player-b", TileID: "retreat", Defense: 0, Strength: 1, Armor: 2, Health: 10, MaxHealth: 10},
	}
	state.Armies = nil
	state.Cities = nil
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bombard-firepower", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	target, ok := unitByID(next, "target")
	if !ok || target.Health != 7 {
		t.Fatalf("target after firepower bombard = %+v ok=%v, want health 7", target, ok)
	}
}

func TestRetailCombatPromotesSurvivingWinnersToVeteran(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "attacker", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Domain: DomainLand, Attack: 100, Defense: 10, Strength: 20, Armor: 1, Health: 10, MaxHealth: 10},
		{ID: "defender", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Domain: DomainLand, Attack: 0, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"attacker"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"defender"}}}
	next, _, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	attacker, ok := unitByID(next, "attacker")
	if !ok || !attacker.Veteran {
		t.Fatalf("surviving winner was not promoted: unit=%+v ok=%v", attacker, ok)
	}
}

func TestRetailVictoryEnslavementAddsPopulationToNearestWinnerCity(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "slaver", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Domain: DomainLand, Attack: 100, Defense: 10, Strength: 20, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "victory", Kind: EffectVictoryEnslavement}}},
		{ID: "defender", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Domain: DomainLand, Attack: 0, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"slaver"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"defender"}}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "a", Population: 2}, {ID: "city-b", OwnerID: "player-b", TileID: "b", Population: 2}}
	next, events, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	city, ok := cityByID(next, "city-a")
	if !ok || city.Population != 3 {
		t.Fatalf("winner slave city = %+v ok=%v", city, ok)
	}
	captured, ok := cityByID(next, "city-b")
	if !ok || captured.Population != 2 {
		t.Fatalf("captured city population should not receive victory slave: %+v ok=%v", captured, ok)
	}
	if len(events) != 1 || !strings.Contains(string(events[0].Payload), "enslaved_population") {
		t.Fatalf("combat event did not expose aggregate enslavement count safely: %#v", events)
	}
}

func TestRetailCombatDomainMasksSelectEligibleTargets(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "attacker", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Domain: DomainLand, Attack: 100, Defense: 10, Strength: 10, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "land", Kind: EffectCanAttackLand}}},
		{ID: "air-first", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Domain: DomainAir, Attack: 0, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
		{ID: "land-second", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Domain: DomainLand, Attack: 0, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"attacker"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"air-first", "land-second"}}}
	next, _, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if air, ok := unitByID(next, "air-first"); !ok || air.Health != 5 {
		t.Fatalf("ineligible air target was damaged: unit=%+v ok=%v", air, ok)
	}
	if _, ok := unitByID(next, "land-second"); ok {
		t.Fatalf("eligible land target survived: %+v", next.Units)
	}
}

func TestRetailBombardDomainMasksSelectEligibleTargets(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 10, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "sea", Kind: EffectCanAttackSea}}},
		{ID: "land-first", OwnerID: "player-b", TileID: "retreat", Domain: DomainLand, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
		{ID: "sea-second", OwnerID: "player-b", TileID: "retreat", Domain: DomainSea, Defense: 1, Strength: 1, Armor: 1, Health: 5, MaxHealth: 5},
	}
	state.Armies = nil
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bombard-domain", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if land, ok := unitByID(next, "land-first"); !ok || land.Health != 5 {
		t.Fatalf("ineligible land bombard target was damaged: unit=%+v ok=%v", land, ok)
	}
	if _, ok := unitByID(next, "sea-second"); ok {
		t.Fatalf("eligible sea bombard target survived: %+v", next.Units)
	}
}

func TestRetailSubmarineRequiresUnderwaterAttackCapability(t *testing.T) {
	target := Unit{Domain: DomainSea, Effects: []Effect{{ID: "sub", Kind: EffectSubmarine}}}
	if attackerCanAttackTarget(Unit{Effects: []Effect{{ID: "sea", Kind: EffectCanAttackSea}}}, target) {
		t.Fatal("sea-only attacker could attack submarine")
	}
	if !attackerCanAttackTarget(Unit{Effects: []Effect{{ID: "underwater", Kind: EffectCanAttackUnderwater}}}, target) {
		t.Fatal("underwater-capable attacker could not attack submarine")
	}
}

func TestRetailProtectFromBarbariansBlocksBarbarianCombat(t *testing.T) {
	state := combatState()
	state.Players = []Player{{ID: "barbarian", NationID: "CIVILIZATION_BARBARIAN", Lifecycle: PlayerActive, Controller: ControllerAI}, {ID: "player-b", NationID: "nation-b", Lifecycle: PlayerActive, Controller: ControllerHuman}}
	state.Phase.ActivePlayerOrder = []string{"barbarian", "player-b"}
	state.Phase.ActivePlayerIndex = 0
	state.Nations = []Nation{{ID: "CIVILIZATION_BARBARIAN"}, {ID: "nation-b"}}
	state.Units[0].OwnerID, state.Units[1].OwnerID = "barbarian", "barbarian"
	state.Armies[0].OwnerID = "barbarian"
	state.Fog = []FogObservation{{PlayerID: "barbarian", TileID: "a", Turn: 1}, {PlayerID: "barbarian", TileID: "b", Turn: 1}}
	state.Wonders = []Wonder{{ID: "WONDER_LABYRINTH", OwnerID: "player-b", CityID: "city-b", Effects: []Effect{{ID: "barbarians", Kind: EffectProtectFromBarbarians}}}}
	payload, _ := json.Marshal(map[string]string{"attacker_army_id": "attackers", "defender_army_id": "defenders"})
	_, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "barbarian-attack", ActorPlayerID: "barbarian", Type: CommandCombat, Payload: payload}, state.Rules, fixtureSources())
	if err == nil || !strings.Contains(err.Error(), "combat target") {
		t.Fatalf("barbarian combat was not blocked err=%v", err)
	}
}

func TestRetailCityDefenseBuildingEffectsModifyDefenders(t *testing.T) {
	state := combatState()
	state.Cities[0].Effects = nil
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_CITY_WALLS", Effects: []Effect{{ID: "city-walls", Kind: EffectCombatDefense, Value: 15}}}}
	defender := state.Units[2]
	if got := combatModifier(state, defender, false); got != 17 { // fortified + forest + city walls
		t.Fatalf("defender combat modifier = %d, want fortified + forest + city walls", got)
	}
}

func TestRetailTerrainImprovementDefenseBonusModifiesDefenders(t *testing.T) {
	state := combatState()
	state.Tiles = []Tile{{ID: "b", Terrain: TerrainPlains}}
	state.Cities = nil
	state.Overlays = []TileOverlay{{ID: "fort", TileID: "b", Kind: "improvement", OwnerID: "player-b", Remaining: 0, ImprovementID: "TILEIMP_FORTIFICATIONS", DefenseBonusPercent: 50}}
	defender := Unit{ID: "defender", OwnerID: "player-b", TileID: "b", Defense: 10}
	if got := combatDefenseModifier(state, defender, Unit{ID: "attacker", OwnerID: "player-a", Domain: DomainLand}); got != 5 {
		t.Fatalf("defender terrain-improvement modifier = %d, want 50%% of defense", got)
	}
}

func TestRetailMultipleActiveDefensesCanInterceptBombardment(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 4, Armor: 1, Health: 2, MaxHealth: 2},
		{ID: "target", OwnerID: "player-b", TileID: "retreat", Domain: DomainLand, Defense: 1, Strength: 1, Armor: 2, Health: 4, MaxHealth: 4},
		{ID: "guardian-a", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Ranged: 100, ActiveDefenseRange: 1, Strength: 1, Armor: 1, Health: 4, MaxHealth: 4},
		{ID: "guardian-b", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Ranged: 100, ActiveDefenseRange: 1, Strength: 1, Armor: 1, Health: 4, MaxHealth: 4},
	}
	state.Armies = nil
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})

	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "multi-active-defense", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := unitByID(next, "bombarder"); ok {
		t.Fatalf("multiple active defenses did not destroy bombarder: %+v events=%+v", next.Units, events)
	}
	if target, ok := unitByID(next, "target"); !ok || target.Health != 4 {
		t.Fatalf("bombard continued after multiple interceptions: target=%+v ok=%v", target, ok)
	}
}

func TestRetailActiveDefenseInterceptsBombardment(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 4, Armor: 1, Health: 2, MaxHealth: 2},
		{ID: "target", OwnerID: "player-b", TileID: "retreat", Domain: DomainLand, Defense: 1, Strength: 1, Armor: 2, Health: 4, MaxHealth: 4},
		{ID: "guardian", OwnerID: "player-b", TileID: "b", Domain: DomainLand, Ranged: 100, ActiveDefenseRange: 1, Strength: 3, Armor: 1, Health: 4, MaxHealth: 4},
	}
	state.Armies = nil
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})

	next, events, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "active-defense", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventBombarded {
		t.Fatalf("active defense events = %+v", events)
	}
	if _, ok := unitByID(next, "bombarder"); ok {
		t.Fatalf("active defense did not destroy bombarder: %+v", next.Units)
	}
	if target, ok := unitByID(next, "target"); !ok || target.Health != 4 {
		t.Fatalf("bombard continued after interception: target=%+v ok=%v", target, ok)
	}
}

func TestRetailBombardDamagesVisibleTargetAtRange(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "bombarder", OwnerID: "player-a", TileID: "a", Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 4, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "can-land", Kind: EffectCanAttackLand}}},
		{ID: "target", OwnerID: "player-b", TileID: "retreat", Domain: DomainLand, Defense: 1, Strength: 1, Armor: 2, Health: 4, MaxHealth: 4},
	}
	state.Armies = nil
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "retreat", Turn: 1})
	payload, _ := json.Marshal(map[string]string{"unit_id": "bombarder", "target_tile_id": "retreat"})

	next, events, result, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bombard", ActorPlayerID: "player-a", Type: CommandBombard, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventBombarded || result.Combat != nil {
		t.Fatalf("bombard transition = events:%+v result:%+v", events, result)
	}
	target, ok := unitByID(next, "target")
	if !ok || target.Health >= 4 {
		t.Fatalf("target after bombard = %+v ok=%v", target, ok)
	}
	bombarder, ok := unitByID(next, "bombarder")
	if !ok || bombarder.Movement != 0 {
		t.Fatalf("bombarder movement after bombard = %+v ok=%v", bombarder, ok)
	}
}

func TestRetailCanAttackEffectsRestrictCombatTargets(t *testing.T) {
	state := combatState()
	state.Units[0].Effects = []Effect{{ID: "air-only", Kind: EffectCanAttackAir}}
	state.Units[2].Domain = DomainLand
	if canArmyFight(state, []*Unit{&state.Units[0]}, []*Unit{&state.Units[2]}) {
		t.Fatal("air-only attacker could fight land defender")
	}
	state.Units[2].Domain = DomainAir
	if !canArmyFight(state, []*Unit{&state.Units[0]}, []*Unit{&state.Units[2]}) {
		t.Fatal("air-only attacker could not fight air defender")
	}
	state.Units[0].Effects = []Effect{{ID: "underwater-only", Kind: EffectCanAttackUnderwater}}
	state.Units[2].Domain = DomainSea
	state.Units[2].Effects = []Effect{{ID: "sub", Kind: EffectSubmarine}}
	if !canArmyFight(state, []*Unit{&state.Units[0]}, []*Unit{&state.Units[2]}) {
		t.Fatal("underwater attacker could not fight submarine defender")
	}
}

func TestRetailAirDefenseBonusModifiesDefenderAgainstAirAttackers(t *testing.T) {
	state := combatState()
	attacker := state.Units[0]
	attacker.Domain = DomainAir
	defender := state.Units[2]
	defender.Defense = 10
	defender.Effects = []Effect{{ID: "air-defense", Kind: EffectCombatDefenseVsAirPercent, Value: 100}}
	if got := combatDefenseModifier(state, defender, attacker); got != 13 { // fortified + forest + city wall + 100% of defense
		t.Fatalf("air defense modifier = %d, want terrain/city plus air defense", got)
	}
	attacker.Domain = DomainLand
	if got := combatDefenseModifier(state, defender, attacker); got != 3 {
		t.Fatalf("land attacker modifier = %d, want only terrain/city", got)
	}
}

func TestRetailIgnoreCityWallsRemovesOnlyWallDefense(t *testing.T) {
	state := combatState()
	state.Cities[0].Effects = nil
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_CITY_WALLS", Effects: []Effect{{ID: "city-walls-defense", Kind: EffectCombatDefense, Value: 15}, {ID: "city-walls", Kind: EffectCityWalls, Value: 15}}}, {ID: "IMPROVE_BUNKER", Effects: []Effect{{ID: "bunker", Kind: EffectCombatDefense, Value: 4}}}}
	attacker := state.Units[0]
	defender := state.Units[2]
	if got := combatDefenseModifier(state, defender, attacker); got != 21 { // fortified + forest + city walls + bunker
		t.Fatalf("defender modifier without ignore = %d, want 21", got)
	}
	attacker.Effects = []Effect{{ID: "ignore-walls", Kind: EffectIgnoreCityWalls}}
	if got := combatDefenseModifier(state, defender, attacker); got != 6 { // fortified + forest + bunker only
		t.Fatalf("defender modifier with ignore = %d, want 6", got)
	}
}

func TestRetailMountedDefenseBonusModifiesDefenderAgainstMountedAttackers(t *testing.T) {
	state := combatState()
	attacker := state.Units[0]
	attacker.Effects = []Effect{{ID: "mounted", Kind: EffectMounted}}
	defender := state.Units[2]
	defender.Defense = 10
	defender.Effects = []Effect{{ID: "pike", Kind: EffectCombatDefenseVsMountedPercent, Value: 50}}
	if got := combatDefenseModifier(state, defender, attacker); got != 8 { // fortified + forest + city wall + 50% of defense
		t.Fatalf("mounted defense modifier = %d, want terrain/city plus mounted bonus", got)
	}
	attacker.Effects = nil
	if got := combatDefenseModifier(state, defender, attacker); got != 3 {
		t.Fatalf("non-mounted defense modifier = %d, want only terrain/city", got)
	}
}

func TestRetailWoodenShipAttackBonusAppliesAgainstWoodenShips(t *testing.T) {
	state := combatState()
	attacker := state.Units[0]
	attacker.Attack = 20
	attacker.Effects = []Effect{{ID: "fire", Kind: EffectCombatAttackVsWoodenShipPercent, Value: 25}}
	defender := state.Units[2]
	defender.Effects = []Effect{{ID: "wood", Kind: EffectWoodenShip}}
	if got := combatAttackModifier(state, attacker, defender); got != 6 { // nation banner + 25% of attack
		t.Fatalf("wooden ship attack modifier = %d, want 6", got)
	}
	defender.Effects = nil
	if got := combatAttackModifier(state, attacker, defender); got != 1 {
		t.Fatalf("non-wooden target modifier = %d, want only nation banner", got)
	}
}

func TestRetailCityOffenseBuildingEffectsModifyAttackersByDefenderDomain(t *testing.T) {
	state := combatState()
	state.Cities = append(state.Cities, City{ID: "city-a", OwnerID: "player-a", TileID: "a", Buildings: []Building{{ID: "IMPROVE_BALLISTA_TOWERS", Effects: []Effect{{ID: "ballista", Kind: EffectCombatAttackLand, Value: 20}}}}})
	attacker := state.Units[0]
	defender := state.Units[2]
	defender.Domain = DomainLand
	if got := combatAttackModifier(state, attacker, defender); got != 21 { // nation banner + ballista
		t.Fatalf("land attack modifier = %d, want nation + city offense", got)
	}
	defender.Domain = DomainAir
	if got := combatAttackModifier(state, attacker, defender); got != 1 {
		t.Fatalf("air attack modifier = %d, want only nation banner", got)
	}
}

func TestRetailCantCaptureCityPreventsCityCapture(t *testing.T) {
	state := combatState()
	state.Units = []Unit{
		{ID: "attacker", OwnerID: "player-a", TileID: "a", ArmyID: "attackers", Attack: 100, Defense: 1, Strength: 10, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "cant-capture", Kind: EffectCantCaptureCity}}},
		{ID: "defender", OwnerID: "player-b", TileID: "b", ArmyID: "defenders", Attack: 0, Defense: 1, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1},
	}
	state.Armies = []Army{{ID: "attackers", OwnerID: "player-a", UnitIDs: []string{"attacker"}}, {ID: "defenders", OwnerID: "player-b", UnitIDs: []string{"defender"}}}

	next, _, result, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if result.Combat.CapturedCityID != "" || next.Cities[0].OwnerID != "player-b" {
		t.Fatalf("city captured by non-capturing unit: result=%+v city=%+v", result.Combat, next.Cities[0])
	}
	attacker, ok := unitByID(next, "attacker")
	if !ok || attacker.TileID != "a" {
		t.Fatalf("non-capturing attacker moved into city: %+v ok=%v", attacker, ok)
	}
}

func TestCombatVectorsResolveAtomicallyWithSourceOrderAndCapture(t *testing.T) {
	state := combatState()
	next, events, result, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventCombatResolved || result.Combat == nil {
		t.Fatalf("unexpected combat transition: %#v %#v", events, result)
	}
	if result.Combat.Outcome != CombatOutcomeAttackerWon || result.Combat.DefenderCasualties != 1 || result.Combat.CapturedCityID != "city-b" {
		t.Fatalf("unexpected combat result: %#v", result.Combat)
	}
	nextCombatRNG, _ := next.RNGState(RNGDomainCombat)
	stateCombatRNG, _ := state.RNGState(RNGDomainCombat)
	if next.Revision != state.Revision+1 || next.EventSequence != state.EventSequence+1 || nextCombatRNG.State == stateCombatRNG.State {
		t.Fatalf("combat was not committed atomically: %#v", next)
	}
	if _, ok := unitByID(next, "defender"); ok {
		t.Fatal("defender survived a lethal combat vector")
	}
	attacker, _ := unitByID(next, "attacker-melee")
	if attacker.TileID != "b" {
		t.Fatalf("victorious army did not capture the target tile: %#v", attacker)
	}
	if city, ok := cityByID(next, "city-b"); !ok || city.OwnerID != "player-a" {
		t.Fatalf("city was not captured: %#v", city)
	}
}

func TestCombatRejectsHiddenAndInvalidTargetsWithoutRNG(t *testing.T) {
	state := combatState()
	state.Fog = []FogObservation{{PlayerID: "player-a", TileID: "a", Turn: 1}}
	_, _, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err == nil || !strings.Contains(err.Error(), "combat target is unavailable") {
		t.Fatalf("hidden combat target error leaked or succeeded: %v", err)
	}
	combatRNG, _ := state.RNGState(RNGDomainCombat)
	if combatRNG.State != 1 {
		t.Fatal("rejected combat consumed RNG")
	}
	bad := combatCommand()
	bad.Payload = json.RawMessage(`{"attacker_army_id":"attackers","defender_army_id":"missing"}`)
	if _, _, _, err := Apply(combatState(), bad, state.Rules, fixtureSources()); err == nil || strings.Contains(err.Error(), "missing") {
		t.Fatalf("invalid target was not safely rejected: %v", err)
	}
}

func TestCombatRetreatRetainsBothArmies(t *testing.T) {
	state := combatState()
	for i := range state.Units {
		state.Units[i].Attack, state.Units[i].Ranged, state.Units[i].Defense, state.Units[i].Health, state.Units[i].MaxHealth = 1, 0, 100, 100, 100
	}
	next, _, result, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if result.Combat.Outcome != CombatOutcomeRetreated || len(next.Units) != len(state.Units) {
		t.Fatalf("nonlethal battle did not retreat safely: %#v %#v", result.Combat, next.Units)
	}
}

func TestCombatReplayIsDeterministicAndMoveNeverResolvesCombat(t *testing.T) {
	state := combatState()
	var firstStateHash, firstEventsHash string
	for i := 0; i < 1000; i++ {
		next, events, _, err := Apply(state, combatCommand(), state.Rules, fixtureSources())
		if err != nil {
			t.Fatal(err)
		}
		stateHash, err := CanonicalStateHash(next)
		if err != nil {
			t.Fatal(err)
		}
		eventsHash, err := CanonicalEventsHash(events)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			firstStateHash, firstEventsHash = stateHash, eventsHash
		} else if stateHash != firstStateHash || eventsHash != firstEventsHash {
			t.Fatalf("combat replay %d was not deterministic", i)
		}
	}
	move := Command{SchemaVersion: CommandSchemaV1, ID: "move-1", ActorPlayerID: "player-a", Type: CommandMoveUnit, Payload: json.RawMessage(`{"unit_id":"attacker-melee","path":["b"]}`)}
	if _, _, _, err := Apply(combatState(), move, state.Rules, fixtureSources()); err == nil {
		t.Fatal("move resolved an occupied combat tile")
	}
}
