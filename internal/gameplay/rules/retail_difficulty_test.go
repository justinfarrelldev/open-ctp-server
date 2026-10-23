package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartingChancesMatchOriginalReleaseNotInstallationExtensions(t *testing.T) {
	root := os.Getenv("CTPCLIENT")
	if root == "" {
		t.Skip("original client installation is unavailable")
	}
	data, err := os.ReadFile(filepath.Join(root, "ctp2", "ctp2_data", "default", "gamedata", "DiffDB.txt"))
	if err != nil {
		t.Skip("original DiffDB is unavailable")
	}
	original, err := ParseRetailStartingAdvanceDifficulties(data)
	if err != nil {
		t.Fatal(err)
	}
	installed, err := LoadRetailRulesFromDirectory(localRetailGameDataRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(original) != fmt.Sprint(installed.StartingAdvanceDifficulties) {
		t.Fatal("configured starting grants differ from original release")
	}
}

func TestRetailStartingDifficultyImportsAllSixOriginalProfiles(t *testing.T) {
	db, err := LoadRetailRulesFromDirectory(localRetailGameDataRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	runtime := ProjectRetailRuntime(db)
	if runtime.StartingAdvances == nil || runtime.StartingAdvances.DifficultyID != "medium" || len(runtime.StartingAdvances.Difficulties) != 6 {
		t.Fatalf("missing explicit starting policy: %+v", runtime.StartingAdvances)
	}
	caps := [][2]int{{6, 2}, {6, 3}, {4, 4}, {3, 4}, {3, 6}, {3, 8}}
	for n, profile := range runtime.StartingAdvances.Difficulties {
		if profile.MaxHumanAdvances != caps[n][0] || profile.MaxAIAdvances != caps[n][1] || len(profile.Chances) != 12 {
			t.Fatalf("wrong source profile %d: %+v", n, profile)
		}
		first := profile.Chances[0]
		if first.AdvanceID != "ADVANCE_TOOLMAKING" || first.HumanChance != 100 || first.AIChance != 100 {
			t.Fatalf("lost guaranteed Toolmaking: %+v", first)
		}
	}
	impossible := runtime.StartingAdvances.Difficulties[5].Chances[1]
	if impossible.AdvanceID != "ADVANCE_AGRICULTURE" || impossible.HumanChance != 100 || impossible.AIChance != 100 {
		t.Fatalf("lost Impossible guaranteed Agriculture: %+v", impossible)
	}
	raw, hash, err := CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := CompileRuntime(raw, hash)
	if err != nil {
		t.Fatal(err)
	}
	for n, profile := range restored.StartingAdvances.Difficulties {
		if fmt.Sprint(profile.Chances) != fmt.Sprint(runtime.StartingAdvances.Difficulties[n].Chances) {
			t.Fatal("canonicalization changed source order")
		}
	}
}

func TestPinnedStartingPolicyFailsClosedForUnsupportedOrMalformedChoices(t *testing.T) {
	db, err := LoadRetailRulesFromDirectory(localRetailGameDataRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*RuntimeRules){
		func(r *RuntimeRules) { r.StartingAdvances.Version = "future" },
		func(r *RuntimeRules) { r.StartingAdvances.DifficultyID = "unknown" },
		func(r *RuntimeRules) { r.StartAgeID = "unknown" },
		func(r *RuntimeRules) { r.StartAgeID = "AGE_THREE"; r.EndAgeID = "AGE_ONE" },
		func(r *RuntimeRules) { r.StartingAdvances.Difficulties[0].MaxHumanAdvances = -1 },
		func(r *RuntimeRules) { r.StartingAdvances.Difficulties[0].Chances[0].AdvanceID = "unknown" },
		func(r *RuntimeRules) { r.StartingAdvances.Difficulties[0].Chances[0].AIChance = -1 },
		func(r *RuntimeRules) {
			r.StartingAdvances.Difficulties[0].Chances = append(r.StartingAdvances.Difficulties[0].Chances, r.StartingAdvances.Difficulties[0].Chances[0])
		},
	} {
		runtime := ProjectRetailRuntime(db)
		mutate(&runtime)
		raw, hash, err := CanonicalRuntimeDefinitions(runtime)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := CompileRuntime(raw, hash); err == nil {
			t.Fatal("malformed starting policy accepted by runtime pin")
		}
	}
}

func TestStartingDifficultyParserRejectsMalformedDataAndPreservesOrder(t *testing.T) {
	block := "{\nMAX_HUMAN_ADVANCES 1\nMAX_AI_ADVANCES 0\nADVANCE_CHANCES {\nADVANCE_Z 25 100\nADVANCE_A 100 0\n}\n}\n"
	input := strings.Repeat(block, 6)
	profiles, err := ParseRetailStartingAdvanceDifficulties([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if profiles[0].Chances[0].AdvanceID != "ADVANCE_Z" {
		t.Fatal("source order was alphabetized")
	}
	for _, invalid := range []string{
		strings.Repeat(block, 5), strings.Repeat(block, 7),
		strings.Replace(input, "MAX_AI_ADVANCES 0", "MAX_AI_ADVANCES -1", 1),
		strings.Replace(input, "MAX_AI_ADVANCES 0\n", "", 1),
		strings.Replace(input, "25 100", "oops 100", 1),
		strings.Replace(input, "25 100", "25", 1),
		strings.Replace(input, "25 100", "25 100 100", 1),
		strings.Replace(input, "25 100", "-1 100", 1),
		strings.Replace(input, "ADVANCE_A 100 0", "ADVANCE_Z 100 0", 1),
		input[:len(input)-2],
	} {
		if _, err := ParseRetailStartingAdvanceDifficulties([]byte(invalid)); err == nil {
			t.Fatalf("malformed difficulty accepted: %q", invalid)
		}
	}
}
