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
