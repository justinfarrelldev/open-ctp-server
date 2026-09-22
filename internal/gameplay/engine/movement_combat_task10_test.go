package engine

import (
	"strings"
	"testing"
)

func TestTask10EmbassyCreatesDurableContactRelation(t *testing.T) {
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{{ID: "a", X: 0, Y: 0, Terrain: TerrainPlains}, {ID: "b", X: 1, Y: 0, Terrain: TerrainPlains}}
	state.Units = []Unit{{ID: "diplomat", OwnerID: "player-a", TileID: "a", Domain: DomainLand, Movement: 1, Effects: []Effect{{ID: "embassy", Kind: EffectSpecialEmbassy, Value: 1}}}}
	state.Cities = []City{{ID: "foreign", OwnerID: "player-b", TileID: "b"}}
	state = runtimeApply(t, state, "player-a", CommandEstablishEmbassy, SpecialActionPayload{UnitID: "diplomat", CityID: "foreign"})
	if len(state.Diplomacy) != 1 || !state.Diplomacy[0].Contact || state.Diplomacy[0].ID != "player-a:player-b" {
		t.Fatalf("embassy did not create contact relation: %#v", state.Diplomacy)
	}
	if !containsString(state.Nations[0].Embassies, "nation-b") {
		t.Fatalf("embassy nation state = %#v", state.Nations[0])
	}
}

func TestTask10WarTransitionsUpdateTreatiesAndTrade(t *testing.T) {
	state := runtimeStartedState(t)
	state = runtimeApply(t, state, "player-a", CommandEstablishContact, contactPayload{TargetPlayerID: "player-b"})
	state = runtimeApply(t, state, "player-a", CommandProposeAgreement, proposeAgreementPayload{ProposalID: "trade-1", TargetPlayerID: "player-b", Kind: "trade"})
	state = runtimeApply(t, state, "player-b", CommandRespondAgreement, respondAgreementPayload{ProposalID: "trade-1", Accept: true})
	state = runtimeApply(t, state, "player-a", CommandSetWar, warStatePayload{TargetPlayerID: "player-b", AtWar: true})
	if len(state.Policy.Wars) != 1 || len(state.Diplomacy) != 1 || state.Diplomacy[0].Treaty != "" || state.Diplomacy[0].Trade {
		t.Fatalf("war did not clear treaty/trade: wars=%#v diplomacy=%#v", state.Policy.Wars, state.Diplomacy)
	}
	state = runtimeApply(t, state, "player-a", CommandSetWar, warStatePayload{TargetPlayerID: "player-b", AtWar: false})
	if len(state.Policy.Wars) != 0 || state.Diplomacy[0].Treaty != "peace" {
		t.Fatalf("peace transition failed: wars=%#v diplomacy=%#v", state.Policy.Wars, state.Diplomacy)
	}
}

func TestTask10CombatRejectsIncompatibleDomainsAndNonCombatArmies(t *testing.T) {
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Tiles = []Tile{{ID: "land", X: 0, Y: 0, Terrain: TerrainPlains}, {ID: "sea", X: 1, Y: 0, Terrain: TerrainOcean}}
	state.Fog = append(state.Fog, FogObservation{PlayerID: "player-a", TileID: "sea"})
	state.Units = []Unit{
		{ID: "worker", OwnerID: "player-a", TileID: "land", ArmyID: "army-a", Domain: DomainLand, Attack: 0, Defense: 1, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1},
		{ID: "ship", OwnerID: "player-b", TileID: "sea", ArmyID: "army-b", Domain: DomainSea, Attack: 1, Defense: 1, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1},
	}
	state.Armies = []Army{{ID: "army-a", OwnerID: "player-a", UnitIDs: []string{"worker"}}, {ID: "army-b", OwnerID: "player-b", UnitIDs: []string{"ship"}}}
	if err := runtimeApplyExpectError(t, state, "player-a", CommandCombat, combatPayload{AttackerArmyID: "army-a", DefenderArmyID: "army-b"}); err == nil || !strings.Contains(err.Error(), "combat target") {
		t.Fatalf("non-combat army error = %v", err)
	}
	state.Units[0].Attack = 1
	if err := runtimeApplyExpectError(t, state, "player-a", CommandCombat, combatPayload{AttackerArmyID: "army-a", DefenderArmyID: "army-b"}); err == nil || !strings.Contains(err.Error(), "combat target") {
		t.Fatalf("incompatible domain error = %v", err)
	}
}

func TestTask10DisembarkRejectsEnemyOccupiedTile(t *testing.T) {
	state := task8State()
	next, _, err := applyTask8(t, state, CommandEmbark, map[string]string{"unit_id": "cargo", "transport_id": "ship"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := applyTask8(t, next, CommandDisembark, map[string]string{"unit_id": "cargo", "tile_id": "d"}); err == nil || !strings.Contains(err.Error(), ErrObstructed.Error()) {
		t.Fatalf("enemy-occupied disembark error = %v", err)
	}
}
