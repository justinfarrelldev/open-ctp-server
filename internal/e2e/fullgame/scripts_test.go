package fullgame

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/puppet"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestResolveScriptUsesOnlyPublicProjectionRoles(t *testing.T) {
	script := Script{Version: 1, Commands: []ScriptCommand{{Command: "fortify", ActorRole: "human-1", Payload: json.RawMessage(`{"unit_id":{"role":"human-1.unit-1"}}`)}}}
	projection, err := json.Marshal(visibility.Projection{Units: []visibility.UnitProjection{{ID: "settler-7-1", OwnerID: "account-7", TileID: "tile-0-0"}}})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveScript(script, projection, map[string]string{"human-1": "account-7"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(resolved[0].Payload), `{"unit_id":"settler-7-1"}`; got != want {
		t.Fatalf("payload = %s, want %s", got, want)
	}
}

func TestResolveScriptRejectsAIActor(t *testing.T) {
	script := Script{Version: 1, Commands: []ScriptCommand{{Command: "found-city", ActorRole: "ai-1", Payload: json.RawMessage(`{"unit_id":"visible-unit"}`)}}}
	projection, err := json.Marshal(visibility.Projection{Units: []visibility.UnitProjection{{ID: "visible-unit", OwnerID: "account-8"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveScript(script, projection, map[string]string{"ai-1": "account-8"}); err == nil {
		t.Fatal("ResolveScript accepted an AI actor")
	}
}

func TestResolveScriptBindsOnlyCurrentlyVisibleForeignCity(t *testing.T) {
	script := Script{Version: 1, Commands: []ScriptCommand{{Command: "establish-embassy", ActorRole: "human-1", Payload: json.RawMessage(`{"city_id":{"role":"human-1.target-city-1"}}`)}}}
	projection, _ := json.Marshal(visibility.Projection{Cities: []visibility.CityProjection{{ID: "visible-city", OwnerID: "account-8"}, {ID: "stale-city", OwnerID: "account-8", Stale: true}}})
	resolved, err := ResolveScript(script, projection, map[string]string{"human-1": "account-7"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(resolved[0].Payload), `{"city_id":"visible-city"}`; got != want {
		t.Fatalf("payload = %s, want %s", got, want)
	}
}

func TestResolveScriptRejectsUnknownOrHiddenRoles(t *testing.T) {
	script := Script{Version: 1, Commands: []ScriptCommand{{Command: "fortify", ActorRole: "human-1", Payload: json.RawMessage(`{"unit_id":{"role":"human-1.unit-2"}}`)}}}
	projection, _ := json.Marshal(visibility.Projection{Units: []visibility.UnitProjection{{ID: "visible", OwnerID: "account-7"}}})
	if _, err := ResolveScript(script, projection, map[string]string{"human-1": "account-7"}); err == nil {
		t.Fatal("ResolveScript accepted hidden role")
	}
}

func TestResolveScriptBindsVisibleTileRoles(t *testing.T) {
	script := Script{Version: 1, Commands: []ScriptCommand{{Command: "terraform", ActorRole: "human-1", Payload: json.RawMessage(`{"tile_id":{"role":"human-1.tile-1"},"unit_id":{"role":"human-1.unit-1"}}`)}}}
	projection, _ := json.Marshal(visibility.Projection{Tiles: []visibility.TileProjection{{ID: "tile-0-0", X: 0, Y: 0}, {ID: "tile-1-0", X: 1, Y: 0}}, Units: []visibility.UnitProjection{{ID: "settler", OwnerID: "account-7"}}})
	resolved, err := ResolveScript(script, projection, map[string]string{"human-1": "account-7"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(resolved[0].Payload), `{"tile_id":"tile-0-0","unit_id":"settler"}`; got != want {
		t.Fatalf("payload = %s, want %s", got, want)
	}
}

func TestMilitaryScriptRebindsArmyIDPerBoundHuman(t *testing.T) {
	body, err := os.ReadFile("scripts/military-zoc-combat.json")
	if err != nil {
		t.Fatal(err)
	}
	var script Script
	if err := json.Unmarshal(body, &script); err != nil {
		t.Fatal(err)
	}
	resolveArmy := func(playerID, unitID string) string {
		t.Helper()
		projection, err := json.Marshal(visibility.Projection{Units: []visibility.UnitProjection{{ID: unitID, OwnerID: playerID, TileID: "tile-0-0"}}})
		if err != nil {
			t.Fatal(err)
		}
		resolved, err := ResolveScript(script, projection, map[string]string{"human-1": playerID})
		if err != nil {
			t.Fatal(err)
		}
		var payload struct {
			ArmyID string   `json:"army_id"`
			UnitIDs []string `json:"unit_ids"`
		}
		if err := json.Unmarshal(resolved[0].Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.UnitIDs) != 1 || payload.UnitIDs[0] != unitID {
			t.Fatalf("resolved unit ids = %#v, want [%q]", payload.UnitIDs, unitID)
		}
		return payload.ArmyID
	}
	if got, want := resolveArmy("account-1", "unit-1"), "fullgame-army-human-1"; got != want {
		t.Fatalf("account-1 army id = %q, want %q", got, want)
	}
	if got, want := resolveArmy("account-2", "unit-2"), "fullgame-army-human-2"; got != want {
		t.Fatalf("account-2 army id = %q, want %q", got, want)
	}
}

func TestResolveContinuityPreparationCommandsSpreadsStackedRowThreeSettlers(t *testing.T) {
	scenario := AcceptanceConfig{Rules: AcceptanceRulesFixture{HumanScript: "scripts/score-resignation-continuity.json"}}
	projection, err := json.Marshal(visibility.Projection{
		Units: []visibility.UnitProjection{
			{ID: "settler-a", OwnerID: "account-7", TileID: "tile-2", Domain: "land", Movement: 2},
			{ID: "settler-b", OwnerID: "account-7", TileID: "tile-2", Domain: "land", Movement: 2},
			{ID: "settler-c", OwnerID: "account-7", TileID: "tile-2", Domain: "land", Movement: 2},
		},
		Tiles: []visibility.TileProjection{
			{ID: "tile-1", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "tile-2", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "tile-3", X: 2, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "tile-10", X: 3, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "tile-20", X: 4, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "tile-water", X: 1, Y: 1, Terrain: "water", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"sea"}, Accessible: true}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	commands, preparing, err := resolveContinuityPreparationCommands(scenario, projection, "account-7")
	if err != nil {
		t.Fatal(err)
	}
	if !preparing || len(commands) != 2 {
		t.Fatalf("preparation = %v commands=%#v, want two spread moves", preparing, commands)
	}
	want := []string{`{"path":["tile-1"],"unit_id":"settler-b"}`, `{"path":["tile-3"],"unit_id":"settler-c"}`}
	for index, command := range commands {
		if command.Command != "move" {
			t.Fatalf("command %d = %q, want move", index, command.Command)
		}
		if string(command.Payload) != want[index] {
			t.Fatalf("command %d payload = %s, want %s", index, command.Payload, want[index])
		}
	}
	if strings.Contains(string(commands[1].Payload), `"tile-10"`) {
		t.Fatalf("non-adjacent accessible target leaked as a direct hop: %s", commands[1].Payload)
	}
}

func TestResolveContinuityLeaveCommandsQueuesThreeDistinctCitiesBeforeLeave(t *testing.T) {
	projection, err := json.Marshal(visibility.Projection{Units: []visibility.UnitProjection{{ID: "settler-a", OwnerID: "account-7", TileID: "tile-0-0"}, {ID: "settler-b", OwnerID: "account-7", TileID: "tile-1-0"}, {ID: "settler-c", OwnerID: "account-7", TileID: "tile-2-0"}}})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveContinuityLeaveCommands("scripts/score-resignation-continuity.json", projection, "account-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 4 {
		t.Fatalf("resolved command count = %d, want 4", len(resolved))
	}
	wantUnits := []string{"settler-a", "settler-b", "settler-c"}
	wantCityIDs := []string{"fullgame-row03-city-human-1", "fullgame-row03-city-human-2", "fullgame-row03-city-human-3"}
	wantNames := []string{"Row03 City 1", "Row03 City 2", "Row03 City 3"}
	for index := 0; index < 3; index++ {
		if resolved[index].Command != "found-city" {
			t.Fatalf("command %d = %q, want found-city", index, resolved[index].Command)
		}
		var payload struct {
			UnitID string `json:"unit_id"`
			CityID string `json:"city_id"`
			Name   string `json:"name"`
		}
		if err := json.Unmarshal(resolved[index].Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.UnitID != wantUnits[index] || payload.CityID != wantCityIDs[index] || payload.Name != wantNames[index] {
			t.Fatalf("command %d payload = %+v, want unit %q city %q name %q", index, payload, wantUnits[index], wantCityIDs[index], wantNames[index])
		}
	}
	if resolved[3].Command != "leave" || string(resolved[3].Payload) != "{}" {
		t.Fatalf("last command = %+v, want leave with empty payload", resolved[3])
	}
}

func TestResolveScriptedHumanCommandsUsesDirectScriptForNonContinuityPolicy(t *testing.T) {
	scenario := AcceptanceConfig{
		HumanPolicy: ScriptedHumanPolicy{Level: puppet.TestHarnessPolicyLevelProductionFirst},
		Rules:       AcceptanceRulesFixture{HumanScript: "scripts/fog-exploration-lifecycle.json"},
	}
	projection, err := json.Marshal(visibility.Projection{
		Units: []visibility.UnitProjection{{ID: "settler-7-1", OwnerID: "account-7", TileID: "tile-0-0"}},
		Tiles: []visibility.TileProjection{{ID: "tile-0-0", X: 0, Y: 0}, {ID: "tile-1-0", X: 1, Y: 0}},
	})
	if err != nil {
		t.Fatal(err)
	}
	resolved, preparing, err := resolveScriptedHumanCommands(scenario, projection, "account-7")
	if err != nil {
		t.Fatal(err)
	}
	if preparing {
		t.Fatal("non-continuity scripted-human command resolution entered continuity preparation")
	}
	if len(resolved) != 1 || resolved[0].Command != "move" {
		t.Fatalf("resolved commands = %#v, want only the first directly resolved move command", resolved)
	}
}

func TestRowSevenDelayedScienceScriptStagesTwelveRoundClosesBeforeScienceSequence(t *testing.T) {
	delayedBody, err := os.ReadFile("scripts/row07-delayed-city-economy-wonder.json")
	if err != nil {
		t.Fatal(err)
	}
	immediateBody, err := os.ReadFile("scripts/city-economy-wonder.json")
	if err != nil {
		t.Fatal(err)
	}
	var delayed Script
	if err := json.Unmarshal(delayedBody, &delayed); err != nil {
		t.Fatal(err)
	}
	var immediate Script
	if err := json.Unmarshal(immediateBody, &immediate); err != nil {
		t.Fatal(err)
	}
	if got, want := len(delayed.Commands), len(immediate.Commands)+12; got != want {
		t.Fatalf("delayed row 7 command count = %d, want %d", got, want)
	}
	for index := 0; index < 12; index++ {
		command := delayed.Commands[index]
		if command.Command != "round-complete" || command.ActorRole != "human-1" || string(command.Payload) != "{}" {
			t.Fatalf("delayed row 7 prefix command %d = %+v, want public round-complete with empty payload", index, command)
		}
	}
	for index, command := range immediate.Commands {
		delayedCommand := delayed.Commands[index+12]
		if delayedCommand.Command != command.Command || delayedCommand.ActorRole != command.ActorRole || string(delayedCommand.Payload) != string(command.Payload) {
			t.Fatalf("delayed row 7 science command %d = %+v, want %+v", index, delayedCommand, command)
		}
	}
}

func TestMechanicScriptsContainTheirDeclaredPublicCommand(t *testing.T) {
	want := map[string]string{
		"scripts/military-zoc-combat.json":               "army-create",
		"scripts/city-economy-wonder.json":               "found-city",
		"scripts/row07-delayed-city-economy-wonder.json": "round-complete",
		"scripts/score-resignation-continuity.json":      "found-city",
		"scripts/special-terraform.json":                 "terraform",
		"scripts/fog-exploration-lifecycle.json":         "move",
	}
	for path, command := range want {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var script Script
		if err := json.Unmarshal(body, &script); err != nil {
			t.Fatal(err)
		}
		if len(script.Commands) < 2 || script.Commands[0].Command != command || script.Commands[1].Command == "" {
			t.Fatalf("%s command sequence = %#v, want multi-command script beginning with %q", path, script.Commands, command)
		}
	}
}

func TestCheckedInScriptsTargetOnlyHumanRoles(t *testing.T) {
	paths := []string{
		"scripts/military-zoc-combat.json",
		"scripts/city-economy-wonder.json",
		"scripts/row07-delayed-city-economy-wonder.json",
		"scripts/score-resignation-continuity.json",
		"scripts/special-terraform.json",
		"scripts/fog-exploration-lifecycle.json",
	}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var script Script
		if err := json.Unmarshal(body, &script); err != nil {
			t.Fatal(err)
		}
		for _, command := range script.Commands {
			if !isHumanScriptActor(command.ActorRole) {
				t.Fatalf("%s targets non-human actor %q", path, command.ActorRole)
			}
		}
	}
}

func TestCampaignExecutorCannotDirectAICommands(t *testing.T) {
	body, err := os.ReadFile("executor.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{".Driver.Pass(", ".Gateway.Submit(", "ai.Submission{"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("campaign executor contains forbidden AI control %q", forbidden)
		}
	}
}
