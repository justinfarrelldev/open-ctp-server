package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
	"github.com/justinfarrelldev/open-ctp-server/internal/world"
)

func startingAdvanceGameDataRoot(t *testing.T) string {
	t.Helper()
	if root := os.Getenv("CTPCLIENT"); root != "" {
		original := filepath.Join(root, "ctp2", "ctp2_data", "default", "gamedata")
		if _, err := os.Stat(filepath.Join(original, "Units.txt")); err == nil {
			return original
		}
	}
	return localRetailGameDataRoot(t)
}

func startingAdvanceSnapshot() world.WorldSnapshot {
	return world.WorldSnapshot{Config: world.WorldConfig{StartingSettlers: 1, StartingGold: 100}, Turn: world.MutableTurnState{RNGState: 987654321}, Tiles: []world.WorldTile{{X: 0, Y: 0, Terrain: "plains"}, {X: 10, Y: 0, Terrain: "plains"}}, Spawns: []world.PlayerSpawn{{AccountID: 7, X: 0, Y: 0}, {AccountID: 8, X: 10, Y: 0}}}
}

func TestRetailLaterAgeStartupAndPinnedSelectionValidation(t *testing.T) {
	t.Setenv(retailRulesModeEnv, "retail")
	t.Setenv(retailGameDataEnv, startingAdvanceGameDataRoot(t))
	defaultPin, err := serverRuntimePin()
	if err != nil {
		t.Fatal(err)
	}
	laterPin, err := serverRuntimePinForSetupRules(`["difficulty:impossible","start-age:AGE_THREE"]`)
	if err != nil {
		t.Fatal(err)
	}
	if laterPin.ContentHash == defaultPin.ContentHash {
		t.Fatal("starting selections didn't change the pin")
	}
	runtime, err := rules.CompileRuntime(laterPin.Definitions, laterPin.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	state, err := initialGameplayState(startingAdvanceSnapshot(), "hybrid", laterPin, []GamePlayer{{AccountID: 7}, {AccountID: 8, Controller: engine.ControllerAI}})
	if err != nil {
		t.Fatal(err)
	}
	start, _ := runtime.Age("AGE_THREE")
	for _, nation := range state.Nations {
		if nation.CurrentAgeID != "AGE_THREE" {
			t.Fatalf("later start age omitted: %+v", nation)
		}
		for _, advance := range runtime.Advances {
			age, _ := runtime.Age(advance.AgeID)
			if slices.Contains(nation.Advances, advance.ID) != (age.Index < start.Index) {
				t.Fatalf("wrong later-age grant %s", advance.ID)
			}
		}
		if nation.ResearchTargetID == "" || !engine.CanResearchAdvance(nation, runtime, nation.ResearchTargetID) {
			t.Fatal("no legal later-age research")
		}
	}
	reordered, err := initialGameplayState(startingAdvanceSnapshot(), "hybrid", laterPin, []GamePlayer{{AccountID: 8, Controller: engine.ControllerAI}, {AccountID: 7}})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := engine.CanonicalStateJSON(state)
	second, _ := engine.CanonicalStateJSON(reordered)
	if string(first) != string(second) {
		t.Fatal("player ordering changed starting state")
	}
	for _, invalid := range []string{`["difficulty:unknown"]`, `["difficulty:"]`, `["difficulty:easy","difficulty:hard"]`, `["start-age:unknown"]`, `["start-age:"]`, `["start-age:AGE_ONE","start-age:AGE_TWO"]`, `[42]`} {
		if _, err := serverRuntimePinForSetupRules(invalid); err == nil {
			t.Fatalf("invalid startup settings accepted: %s", invalid)
		}
	}
	t.Setenv(retailRulesModeEnv, "")
	if _, err := serverRuntimePinForSetupRules(`["difficulty:medium"]`); err == nil {
		t.Fatal("retail difficulty silently applied to synthetic rules")
	}
}

func TestStartingAdvanceStreamsArePlayerLocalAndSeeded(t *testing.T) {
	one, two, other := startingAdvanceRolls(123), startingAdvanceRolls(123), startingAdvanceRolls(124)
	different := false
	for n := 0; n < 24; n++ {
		a := one("a")
		_ = two("b")
		if a != two("a") {
			t.Fatal("other player consumed this player's starting randomness")
		}
		different = different || a != other("a")
	}
	if !different {
		t.Fatal("starting randomness ignores the world seed")
	}
}

func TestRetailActualStartupEnablesFirstWarriorAtEveryDifficulty(t *testing.T) {
	t.Setenv(retailRulesModeEnv, "retail")
	t.Setenv(retailGameDataEnv, startingAdvanceGameDataRoot(t))
	for _, difficulty := range []string{"beginner", "easy", "medium", "hard", "very-hard", "impossible"} {
		t.Run(difficulty, func(t *testing.T) {
			pin, err := serverRuntimePinForSetupRules(fmt.Sprintf(`["difficulty:%s"]`, difficulty))
			if err != nil {
				t.Fatal(err)
			}
			runtime, err := rules.CompileRuntime(pin.Definitions, pin.ContentHash)
			if err != nil {
				t.Fatal(err)
			}
			if runtime.StartingAdvances == nil || runtime.StartingAdvances.DifficultyID != difficulty {
				t.Fatal("selected difficulty was not pinned")
			}
			profile, _ := runtime.StartingAdvances.Difficulty()
			state, err := initialGameplayState(startingAdvanceSnapshot(), "sequential", pin, []GamePlayer{{AccountID: 7}, {AccountID: 8, Controller: engine.ControllerAI}})
			if err != nil {
				t.Fatal(err)
			}
			for _, nation := range state.Nations {
				cap := profile.MaxHumanAdvances
				if nation.ID == "nation-account-8" {
					cap = profile.MaxAIAdvances
				}
				if len(nation.Advances) > cap || len(nation.Advances) >= len(profile.Chances) {
					t.Fatalf("all advances granted instead of source cap/chances: %+v", nation)
				}
				if !slices.Contains(nation.Advances, "ADVANCE_TOOLMAKING") {
					t.Fatalf("%s omitted guaranteed Toolmaking: %+v", difficulty, nation)
				}
				if nation.CurrentAgeID != runtime.StartAgeID || nation.ResearchTargetID == "" || !engine.CanResearchAdvance(nation, runtime, nation.ResearchTargetID) {
					t.Fatalf("missing age/legal initial research: %+v", nation)
				}
			}
			original, err := engine.CanonicalStateJSON(state)
			if err != nil {
				t.Fatal(err)
			}
			repeat, err := initialGameplayState(startingAdvanceSnapshot(), "sequential", pin, []GamePlayer{{AccountID: 8, Controller: engine.ControllerAI}, {AccountID: 7}})
			if err != nil {
				t.Fatal(err)
			}
			repeated, _ := engine.CanonicalStateJSON(repeat)
			if string(original) != string(repeated) {
				t.Fatal("startup is not deterministic")
			}
			state, err = engine.DecodeCanonicalStateJSON(original)
			if err != nil {
				t.Fatal(err)
			}
			sources := testStartGameplaySources()
			apply := func(kind string, payload any) (engine.State, error) {
				body, _ := json.Marshal(payload)
				next, _, _, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: kind, ActorPlayerID: "account-7", Type: kind, Payload: body}, pin, sources)
				return next, err
			}
			state, err = apply(engine.CommandFoundCity, engine.FoundCityPayload{UnitID: "settler-7-1", CityID: "home", Name: "Home"})
			if err != nil {
				t.Fatal(err)
			}
			p, err := visibility.Project(state, "account-7")
			if err != nil {
				t.Fatal(err)
			}
			if p.Rules.StartingDifficulty != difficulty || p.Rules.StartAgeID != "AGE_ONE" {
				t.Fatal("public pin provenance omitted starting settings")
			}
			options := p.Cities[0].ProductionOptions
			if options == nil || options.Status != "complete" {
				t.Fatalf("no owned-city options: %+v", options)
			}
			warrior, settler := false, false
			for _, item := range options.Items {
				warrior = warrior || item.ID == "UNIT_WARRIOR"
				settler = settler || item.ID == "UNIT_SETTLER"
			}
			if !warrior || settler {
				t.Fatalf("population-one city options: warrior=%v settler=%v", warrior, settler)
			}
			if _, err := apply(engine.CommandEditProduction, engine.EditProductionPayload{CityID: "home", Entries: []engine.QueueSelection{{ItemID: "UNIT_SETTLER"}}}); err == nil {
				t.Fatal("population-one Settler accepted")
			}
			built, err := apply(engine.CommandEditProduction, engine.EditProductionPayload{CityID: "home", Entries: []engine.QueueSelection{{ItemID: "UNIT_WARRIOR"}}})
			if err != nil || len(built.Cities[0].Queue) != 1 {
				t.Fatalf("guaranteed Warrior append rejected: %v", err)
			}
			state = built
			// Do not assign Population or StoredFood. Let the real pinned city
			// economy grow the first city across the one-to-two threshold.
			for round := 0; round < 64 && state.Cities[0].Population == 1; round++ {
				for _, actor := range state.Phase.ActivePlayerOrder {
					next, _, _, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: fmt.Sprintf("grow-%d-%s", state.Revision, actor), ActorPlayerID: actor, Type: engine.CommandPass, Payload: []byte(`{}`)}, pin, sources)
					if err != nil {
						t.Fatalf("grow first city: %v", err)
					}
					state = next
				}
			}
			if state.Cities[0].Population != 2 {
				t.Fatalf("pinned economy did not grow the city to pop 2: %+v", state.Cities[0])
			}
			p, err = visibility.Project(state, "account-7")
			if err != nil {
				t.Fatal(err)
			}
			options = p.Cities[0].ProductionOptions
			settler = false
			if options == nil || options.Status != "complete" {
				t.Fatalf("grown city options unavailable: %+v", options)
			}
			for _, item := range options.Items {
				settler = settler || item.ID == "UNIT_SETTLER"
			}
			if !settler {
				t.Fatal("Settler did not become an option at population 2")
			}
			entries := make([]engine.QueueSelection, 0, len(p.Cities[0].Queue)+1)
			for _, item := range p.Cities[0].Queue {
				entries = append(entries, engine.QueueSelection{EntryID: item.EntryID})
			}
			entries = append(entries, engine.QueueSelection{ItemID: "UNIT_SETTLER"})
			queued, err := apply(engine.CommandEditProduction, engine.EditProductionPayload{CityID: "home", Entries: entries})
			if err != nil || queued.Cities[0].Queue[len(queued.Cities[0].Queue)-1].ID != "UNIT_SETTLER" || queued.Cities[0].Population != 2 {
				t.Fatalf("advertised population-two Settler append failed: %v", err)
			}
		})
	}
}
