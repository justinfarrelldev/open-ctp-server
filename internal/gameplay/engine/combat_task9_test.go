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
		state.Units[i].Attack, state.Units[i].Ranged = 0, 0
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
