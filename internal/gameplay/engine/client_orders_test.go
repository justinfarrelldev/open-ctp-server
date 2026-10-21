package engine

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestClientRenameCityValidatesIdentityAndName(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "home", Name: "Old", OwnerID: "player-a", TileID: "a", Population: 2}}
	renamed := task10Apply(t, state, "player-a", CommandRenameCity, RenameCityPayload{CityID: "home", Name: " New Memphis "})
	if renamed.Cities[0].Name != "New Memphis" {
		t.Fatal("authoritative rename not retained")
	}
	for _, payload := range []RenameCityPayload{{CityID: "home", Name: " "}, {CityID: "home", Name: "bad\nname"}, {CityID: "missing", Name: "Name"}} {
		before, err := CanonicalStateJSON(state)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "bad-rename", ActorPlayerID: "player-a", Type: CommandRenameCity, Payload: raw}, state.Rules, fixtureSources()); err == nil {
			t.Fatal("invalid rename accepted")
		}
		after, err := CanonicalStateJSON(state)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Fatal("invalid rename changed state")
		}
	}
}

func TestClientSleepIsDistinctFromFortifyAndPersistsUntilWake(t *testing.T) {
	state := task10State()
	asleep := task10Apply(t, state, "player-a", CommandSleep, unitPayload{UnitID: "settler-a"})
	if !asleep.Units[0].Asleep || asleep.Units[0].Fortified {
		t.Fatalf("sleep was substituted for fortify: %+v", asleep.Units[0])
	}
	raw, err := CanonicalStateJSON(asleep)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	awake := task10Apply(t, restored, "player-a", CommandWake, unitPayload{UnitID: "settler-a"})
	if awake.Units[0].Asleep {
		t.Fatal("wake did not clear retained sleep order")
	}
}

func TestClientDisbandCityCreatesExhaustedSettlerAndEnforcesSourceRestrictions(t *testing.T) {
	state := task10State()
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2}, {ID: "other", OwnerID: "player-a", TileID: "b", Population: 2}}
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	zero := 0
	runtime.Constants.MaxDisbandSize = &zero
	if _, ok := disbandCitySettler(state, "player-a", "home", &runtime); ok {
		t.Fatal("explicit zero size limit was replaced by a legacy default")
	}
	if err := disbandCity(&state, "player-a", CityIdentityPayload{CityID: "home"}, &rules.RuntimeRules{}); err == nil {
		t.Fatal("missing settler must reject instead of destroying the city")
	}
	next := task10Apply(t, state, "player-a", CommandDisbandCity, CityIdentityPayload{CityID: "home"})
	if len(next.Cities) != 1 || next.Cities[0].ID != "other" || len(next.Units) != len(state.Units)+1 {
		t.Fatalf("city disband not atomic: %+v", next.Cities)
	}
	found := false
	for _, unit := range next.Units {
		if unit.ID != state.Units[0].ID && unit.ID != state.Units[1].ID && unit.TileID == "a" && hasEffect(unit.Effects, EffectSettler) && unit.Movement == 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("disband did not create a valid exhausted settler")
	}
	for _, invalid := range []State{next, func() State {
		copy := state
		copy.Cities = append([]City(nil), state.Cities...)
		copy.Cities[0].Population = 4
		return copy
	}()} {
		cityID := invalid.Cities[0].ID
		raw, err := json.Marshal(CityIdentityPayload{CityID: cityID})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := Apply(invalid, Command{SchemaVersion: CommandSchemaV1, ID: "bad-disband", ActorPlayerID: "player-a", Type: CommandDisbandCity, Payload: raw}, invalid.Rules, fixtureSources()); err == nil {
			t.Fatal("last/oversized city disband accepted")
		}
	}
}
