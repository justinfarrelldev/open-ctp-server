package fullgame

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestCatalogSchemaAssemblesExactlySixtyStableCases(t *testing.T) {
	if len(CatalogRows01_12()) != 12 || len(CatalogRows13_24()) != 12 || len(CatalogRows25_36()) != 12 || len(CatalogRows37_48()) != 12 || len(CatalogRows49_60()) != 12 {
		t.Fatalf("catalog split lengths = %d/%d/%d/%d/%d, want 12 each", len(CatalogRows01_12()), len(CatalogRows13_24()), len(CatalogRows25_36()), len(CatalogRows37_48()), len(CatalogRows49_60()))
	}
	if err := ValidateCatalogSchema(Catalog); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogRowOneMatchesAcceptedFixture(t *testing.T) {
	catalogRow := Catalog[0]
	accepted := RowOneConquestAcceptance()
	if catalogRow.ID != "fullgame-01-sequential-bloodlust_conquest" {
		t.Fatalf("catalog row 1 id = %q", catalogRow.ID)
	}
	if catalogRow.Name != accepted.Name || catalogRow.Seed != accepted.Seed || catalogRow.TurnMode != accepted.TurnMode {
		t.Fatalf("catalog row 1 identity drift: row=%#v acceptance=%#v", catalogRow, accepted)
	}
	if len(catalogRow.Slots) != len(accepted.Slots) {
		t.Fatalf("catalog row 1 slot count = %d, want %d", len(catalogRow.Slots), len(accepted.Slots))
	}
	for index := range catalogRow.Slots {
		if catalogRow.Slots[index] != accepted.Slots[index] {
			t.Fatalf("catalog row 1 slot %d = %#v, want %#v", index, catalogRow.Slots[index], accepted.Slots[index])
		}
	}
	if catalogRow.World != accepted.World || catalogRow.Rules != accepted.Rules || catalogRow.HumanPolicy != HumanPolicyConquest || catalogRow.Restart != accepted.Restart || catalogRow.Budgets != accepted.Budgets || catalogRow.Terminal.Reason != accepted.ExpectedReason || catalogRow.Observed.Winner != accepted.ExpectedWinner || catalogRow.Observed.StateHash != accepted.ExpectedStateHash || catalogRow.Observed.EventHash != accepted.ExpectedEventHash {
		t.Fatalf("catalog row 1 drifted from the pinned acceptance fixture: %#v vs %#v", catalogRow, accepted)
	}
}

func TestCatalogRowThreeDeclaresTwoSlotContinuityScoreReachability(t *testing.T) {
	rows := CatalogRows01_12()
	if len(rows) < 3 {
		t.Fatalf("catalog rows 01-12 length = %d, want at least 3", len(rows))
	}
	row := rows[2]
	if row.ID != "fullgame-03-simultaneous-score_2300" {
		t.Fatalf("row 3 id = %q", row.ID)
	}
	if row.TurnMode != "simultaneous" {
		t.Fatalf("row 3 turn mode = %q, want simultaneous", row.TurnMode)
	}
	if row.Rules.Victory != "score_2300" {
		t.Fatalf("row 3 victory = %q, want score_2300", row.Rules.Victory)
	}
	if row.HumanPolicy != HumanPolicyContinuity {
		t.Fatalf("row 3 human policy = %q, want %q", row.HumanPolicy, HumanPolicyContinuity)
	}
	if row.Rules.HumanScript != "scripts/score-resignation-continuity.json" {
		t.Fatalf("row 3 human script = %q, want scripts/score-resignation-continuity.json", row.Rules.HumanScript)
	}
	if row.Budgets.StartingSettlers != 3 || row.Budgets.Commands != 240 || row.Budgets.Turns != 48 {
		t.Fatalf("row 3 budgets = %+v, want starting settlers 3 with commands 240 and turns 48", row.Budgets)
	}
	if row.Restart.Point != AcceptanceRestartAfterCompletedTurn || row.Restart.CompletedTurn != 12 {
		t.Fatalf("row 3 restart = %+v, want completed turn 12", row.Restart)
	}
	if len(row.Slots) != 2 {
		t.Fatalf("row 3 slot count = %d, want 2", len(row.Slots))
	}
	if row.Slots[0].ControllerKind != humanController || row.Slots[1].ControllerKind != aiController {
		t.Fatalf("row 3 controllers = [%q %q], want [%q %q]", row.Slots[0].ControllerKind, row.Slots[1].ControllerKind, humanController, aiController)
	}
}

func TestCatalogRowFourMatchesAcceptedTaskNineArtifacts(t *testing.T) {
	row := CatalogRows01_12()[3]
	if row.ID != "fullgame-04-hybrid-science_gaia_controller" || row.Name != "fullgame-04-hybrid-science_gaia_controller" {
		t.Fatalf("catalog row 4 identity drift: %#v", row)
	}
	if row.TurnMode != "hybrid" || row.World != (AcceptanceWorldFixture{MapSize: "small", Climate: "tropical", Wetness: 90}) {
		t.Fatalf("catalog row 4 world drift: %#v", row)
	}
	if len(row.Slots) != 6 || row.Slots[0].ControllerKind != humanController {
		t.Fatalf("catalog row 4 slots drift: %#v", row.Slots)
	}
	for index := 1; index < len(row.Slots); index++ {
		if row.Slots[index].ControllerKind != aiController {
			t.Fatalf("catalog row 4 slot %d controller = %q, want %q", index, row.Slots[index].ControllerKind, aiController)
		}
	}
	if row.Budgets.StartingSettlers != 1 {
		t.Fatalf("catalog row 4 starting settlers = %d, want 1", row.Budgets.StartingSettlers)
	}
	if row.Budgets.Commands != 240 || row.Budgets.Turns != 48 || row.Budgets.Deadline != 2*time.Minute {
		t.Fatalf("catalog row 4 budgets = %+v, want default acceptance envelope", row.Budgets)
	}
	if row.HumanPolicy != HumanPolicyScriptedHumans || row.Rules.Victory != "science_gaia_controller" || row.Rules.HumanScript != "scripts/city-economy-wonder.json" {
		t.Fatalf("catalog row 4 rules drift: %#v", row)
	}
	if row.Restart.Point != AcceptanceRestartAfterCompletedTurn || row.Restart.CompletedTurn != 12 {
		t.Fatalf("catalog row 4 restart = %#v, want completed turn 12", row.Restart)
	}
	if row.Observed.Winner != "account-1" || row.Observed.StateHash != "eaf7533a03a1d7efee31f1e0f2ad960f0d1d1f79f09b09bae0597be243812fa9" || row.Observed.EventHash != "4b0c590d1b1acce527e4e405f8f122b3e0302adf7f29f963ef0d8179204963a2" {
		t.Fatalf("catalog row 4 observed artifacts drift: %#v", row.Observed)
	}
}

func TestCatalogRowFiveDeclaresSequentialDiplomacyReachability(t *testing.T) {
	row := CatalogRows01_12()[4]
	if row.ID != "fullgame-05-sequential-diplomatic_world_union" || row.Name != "fullgame-05-sequential-diplomatic_world_union" {
		t.Fatalf("catalog row 5 identity drift: %#v", row)
	}
	if row.TurnMode != "sequential" || row.World != (AcceptanceWorldFixture{MapSize: "small", Climate: "cold", Wetness: 50}) {
		t.Fatalf("catalog row 5 world drift: %#v", row)
	}
	if len(row.Slots) != 3 {
		t.Fatalf("catalog row 5 slot count = %d, want 3", len(row.Slots))
	}
	if row.Slots[0].ControllerKind != humanController {
		t.Fatalf("catalog row 5 slot 0 controller = %q, want %q", row.Slots[0].ControllerKind, humanController)
	}
	for index := 1; index < len(row.Slots); index++ {
		if row.Slots[index].ControllerKind != aiController {
			t.Fatalf("catalog row 5 slot %d controller = %q, want %q", index, row.Slots[index].ControllerKind, aiController)
		}
	}
	if row.HumanPolicy != HumanPolicyDiplomacy || row.Rules.Victory != "diplomatic_world_union" || row.Rules.HumanScript != "scripts/city-economy-wonder.json" {
		t.Fatalf("catalog row 5 rules drift: %#v", row)
	}
	if row.Budgets.StartingSettlers != 2 || row.Budgets.Commands != 240 || row.Budgets.Turns != 48 || row.Budgets.Deadline != 2*time.Minute {
		t.Fatalf("catalog row 5 budgets = %+v, want default acceptance envelope", row.Budgets)
	}
	if row.Restart.Point != AcceptanceRestartAfterCompletedTurn || row.Restart.CompletedTurn != 12 {
		t.Fatalf("catalog row 5 restart = %#v, want completed turn 12", row.Restart)
	}
	if row.Observed.Winner != "account-1" || row.Observed.StateHash != "5b4deeb4f0a5106ed1351a6656f7e6db2fdca1de1571772ae78be3572dd861e8" || row.Observed.EventHash != "e3509ea099a3c7bf025683a4378539e40fe4769f615c5d31f8978cf597a44cf9" {
		t.Fatalf("catalog row 5 observed artifacts drift: %#v", row.Observed)
	}
}

func TestCatalogRowSixDeclaresLongTurnContinuityScoreReachability(t *testing.T) {
	row := CatalogRows01_12()[5]
	if row.ID != "fullgame-06-long_turn-score_2300" || row.Name != "fullgame-06-long_turn-score_2300" {
		t.Fatalf("catalog row 6 identity drift: %#v", row)
	}
	if row.TurnMode != "long_turn" || row.World != (AcceptanceWorldFixture{MapSize: "duel", Climate: "cold", Wetness: 90}) {
		t.Fatalf("catalog row 6 world drift: %#v", row)
	}
	if len(row.Slots) != 2 {
		t.Fatalf("catalog row 6 slot count = %d, want 2", len(row.Slots))
	}
	if row.Slots[0].ControllerKind != humanController || row.Slots[1].ControllerKind != aiController {
		t.Fatalf("catalog row 6 controllers = [%q %q], want [%q %q]", row.Slots[0].ControllerKind, row.Slots[1].ControllerKind, humanController, aiController)
	}
	if row.HumanPolicy != HumanPolicyContinuity || row.Rules.Victory != "score_2300" || row.Rules.HumanScript != "scripts/score-resignation-continuity.json" {
		t.Fatalf("catalog row 6 rules drift: %#v", row)
	}
	if row.Budgets.StartingSettlers != 3 || row.Budgets.Commands != 240 || row.Budgets.Turns != 48 || row.Budgets.Deadline != 2*time.Minute {
		t.Fatalf("catalog row 6 budgets = %+v, want starting settlers 3 with default acceptance envelope", row.Budgets)
	}
	if row.Restart.Point != AcceptanceRestartAfterCompletedTurn || row.Restart.CompletedTurn != 12 {
		t.Fatalf("catalog row 6 restart = %#v, want completed turn 12", row.Restart)
	}
	if row.Observed.Winner != "account-1" || row.Observed.StateHash != "5292b679fa07ec631929e4f4eb36805d89abe2ba00c2ccb2dbe0430c616477b5" || row.Observed.EventHash != "b8e31801117572d3cdb441fc16eb3d4eb42d409a953fe74761637573a5017282" {
		t.Fatalf("catalog row 6 observed artifacts drift: %#v", row.Observed)
	}
}

func TestCatalogRowSevenDeclaresDelayedSimultaneousScienceGaiaReachability(t *testing.T) {
	row := CatalogRows01_12()[6]
	if row.ID != "fullgame-07-simultaneous-science_gaia_controller" || row.Name != "fullgame-07-simultaneous-science_gaia_controller" {
		t.Fatalf("catalog row 7 identity drift: %#v", row)
	}
	if row.TurnMode != "simultaneous" || row.World != (AcceptanceWorldFixture{MapSize: "small", Climate: "temperate", Wetness: 10}) {
		t.Fatalf("catalog row 7 world drift: %#v", row)
	}
	if len(row.Slots) != 3 {
		t.Fatalf("catalog row 7 slot count = %d, want 3", len(row.Slots))
	}
	if row.Slots[0].ControllerKind != humanController || row.Slots[1].ControllerKind != aiController || row.Slots[2].ControllerKind != aiController {
		t.Fatalf("catalog row 7 controllers = [%q %q %q], want [%q %q %q]", row.Slots[0].ControllerKind, row.Slots[1].ControllerKind, row.Slots[2].ControllerKind, humanController, aiController, aiController)
	}
	if row.HumanPolicy != HumanPolicyScriptedHumans || row.Rules.Victory != "science_gaia_controller" || row.Rules.HumanScript != "scripts/row07-delayed-city-economy-wonder.json" {
		t.Fatalf("catalog row 7 rules drift: %#v", row)
	}
	if row.Budgets.StartingSettlers != 1 || row.Budgets.Commands != 240 || row.Budgets.Turns != 48 || row.Budgets.Deadline != 2*time.Minute {
		t.Fatalf("catalog row 7 budgets = %+v, want starting settlers 1 with default acceptance envelope", row.Budgets)
	}
	if row.Restart.Point != AcceptanceRestartAfterCompletedTurn || row.Restart.CompletedTurn != 12 {
		t.Fatalf("catalog row 7 restart = %#v, want completed turn 12", row.Restart)
	}
	if row.Observed.Winner != "account-1" || row.Observed.StateHash != "56fce18df944759630c17f22a1e8fab9b6178d7eada079371982db34d9e22018" || row.Observed.EventHash != "242730d9ee70eece9caaae4a9245b3d104ea5a0ad483058c5cffd0ca5a16b278" {
		t.Fatalf("catalog row 7 observed artifacts drift: %#v", row.Observed)
	}
}

func TestCatalogRowNineDeclaresSequentialScoreReachability(t *testing.T) {
	row := CatalogRows01_12()[8]
	if row.ID != "fullgame-09-sequential-score_2300" || row.Name != "fullgame-09-sequential-score_2300" {
		t.Fatalf("catalog row 9 identity drift: %#v", row)
	}
	if row.TurnMode != "sequential" || row.World != (AcceptanceWorldFixture{MapSize: "small", Climate: "cold", Wetness: 50}) {
		t.Fatalf("catalog row 9 world drift: %#v", row)
	}
	if len(row.Slots) != 3 {
		t.Fatalf("catalog row 9 slot count = %d, want 3", len(row.Slots))
	}
	if row.Slots[0].ControllerKind != humanController {
		t.Fatalf("catalog row 9 slot 0 controller = %q, want %q", row.Slots[0].ControllerKind, humanController)
	}
	for index := 1; index < len(row.Slots); index++ {
		if row.Slots[index].ControllerKind != aiController {
			t.Fatalf("catalog row 9 slot %d controller = %q, want %q", index, row.Slots[index].ControllerKind, aiController)
		}
	}
	if row.HumanPolicy != HumanPolicyDiplomacy || row.Rules.Victory != "score_2300" || row.Rules.HumanScript != "scripts/city-economy-wonder.json" {
		t.Fatalf("catalog row 9 rules drift: %#v", row)
	}
	if row.Budgets.StartingSettlers != 2 || row.Budgets.Commands != 240 || row.Budgets.Turns != 48 || row.Budgets.Deadline != 2*time.Minute {
		t.Fatalf("catalog row 9 budgets = %+v, want default acceptance envelope", row.Budgets)
	}
	if row.Restart.Point != AcceptanceRestartAfterCompletedTurn || row.Restart.CompletedTurn != 12 {
		t.Fatalf("catalog row 9 restart = %#v, want completed turn 12", row.Restart)
	}
	if row.Observed != (ObservedArtifacts{}) {
		t.Fatalf("catalog row 9 observed artifacts = %#v, want provisional zero value", row.Observed)
	}
}

func TestCatalogRows01_12CoverTaskNineRosterAndPolicyEnvelope(t *testing.T) {
	rows := CatalogRows01_12()
	modesSeen := map[string]bool{}
	countsSeen := map[int]bool{}
	policiesSeen := map[HumanPolicyKind]bool{}
	oneHumanRestAI := false
	allAIObserveOnly := false
	alternatingMixed := false
	scriptedHumans := false
	continuityWithSurvivor := false
	for _, row := range rows {
		modesSeen[row.TurnMode] = true
		countsSeen[len(row.Slots)] = true
		policiesSeen[row.HumanPolicy] = true
		humans, ais := 0, 0
		alternating := true
		for index, slot := range row.Slots {
			if slot.ControllerKind == humanController {
				humans++
			}
			if slot.ControllerKind == aiController {
				ais++
			}
			want := humanController
			if index%2 == 1 {
				want = aiController
			}
			if slot.ControllerKind != want {
				alternating = false
			}
		}
		if humans == 1 && ais == len(row.Slots)-1 {
			oneHumanRestAI = true
		}
		if humans == 0 && ais == len(row.Slots) && row.HumanPolicy == HumanPolicyObserveOnly {
			allAIObserveOnly = true
		}
		if humans > 0 && ais > 0 && alternating {
			alternatingMixed = true
		}
		if row.HumanPolicy == HumanPolicyScriptedHumans {
			scriptedHumans = true
		}
		if row.HumanPolicy == HumanPolicyContinuity && humans >= 2 {
			continuityWithSurvivor = true
		}
		parts := strings.SplitN(row.ID, "-", 4)
		if len(parts) != 4 {
			t.Fatalf("row id %q is malformed", row.ID)
		}
		if row.Rules.Victory != parts[3] {
			t.Fatalf("row %s rules victory = %q, want stable id suffix %q", row.ID, row.Rules.Victory, parts[3])
		}
	}
	for _, mode := range modes {
		if !modesSeen[mode] {
			t.Fatalf("rows 01-12 missing mode %q", mode)
		}
	}
	for _, count := range countryCounts {
		if !countsSeen[count] {
			t.Fatalf("rows 01-12 missing roster count %d", count)
		}
	}
	for _, policy := range []HumanPolicyKind{HumanPolicyConquest, HumanPolicyDiplomacy, HumanPolicyScriptedHumans, HumanPolicyObserveOnly, HumanPolicyContinuity} {
		if !policiesSeen[policy] {
			t.Fatalf("rows 01-12 missing policy %q", policy)
		}
	}
	if !oneHumanRestAI || !allAIObserveOnly || !alternatingMixed || !scriptedHumans || !continuityWithSurvivor {
		t.Fatalf("rows 01-12 families oneHumanRestAI=%v allAIObserveOnly=%v alternatingMixed=%v scriptedHumans=%v continuityWithSurvivor=%v", oneHumanRestAI, allAIObserveOnly, alternatingMixed, scriptedHumans, continuityWithSurvivor)
	}
}

func TestPublicExecutorCompletesMinimalScenarioThroughProductionRuntime(t *testing.T) {
	if _, err := RunAcceptance(context.Background(), RowOneConquestAcceptance(), PublicExecutor{}); err != nil {
		t.Fatal(err)
	}
}

func TestPublicExecutorCompletesRowTwoThroughProductionRuntime(t *testing.T) {
	config := RowTwoDiplomacyAcceptance()
	restartEvents := make([]AcceptanceRestartEvent, 0, 1)
	config.ObserveRestart = func(event AcceptanceRestartEvent) {
		restartEvents = append(restartEvents, event)
	}
	if _, err := RunAcceptance(context.Background(), config, PublicExecutor{}); err != nil {
		t.Fatal(err)
	}
	want := AcceptanceRestartEvent{Point: AcceptanceRestartAfterCompletedTurn, CompletedTurn: 12}
	if len(restartEvents) != 1 || restartEvents[0] != want {
		t.Fatalf("diplomacy restart events = %+v, want %+v", restartEvents, want)
	}
}

func TestCatalogVictoryIDsMatchSupportedRuntime(t *testing.T) {
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := rules.CompileRuntime(ruleset.CanonicalJSON, ruleset.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(victories, ","), strings.Join(runtime.VictoryEvaluatorIDs(), ","); got != want {
		t.Fatalf("catalog victory IDs = %s, want runtime IDs %s", got, want)
	}
}

func TestAcceptedScenarioRegistryTracksRows01Through07Only(t *testing.T) {
	want := []string{
		"fullgame-01-sequential-bloodlust_conquest",
		"fullgame-02-long_turn-diplomatic_world_union",
		"fullgame-03-simultaneous-score_2300",
		"fullgame-04-hybrid-science_gaia_controller",
		"fullgame-05-sequential-diplomatic_world_union",
		"fullgame-06-long_turn-score_2300",
		"fullgame-07-simultaneous-science_gaia_controller",
	}
	if got := strings.Join(acceptedScenarioIDs, ","); got != strings.Join(want, ",") {
		t.Fatalf("acceptedScenarioIDs = %q, want %q", got, strings.Join(want, ","))
	}
}

func TestCatalogRejectsProvisionalScenarios(t *testing.T) {
	if err := ValidateFinalAll(Catalog); err == nil || !strings.Contains(err.Error(), "scenario fullgame-08-hybrid-bloodlust_conquest has non-observed immutable artifacts") {
		t.Fatalf("ValidateFinalAll(Catalog) error = %v, want row 8 provisional artifact rejection", err)
	}
}

func TestValidateObservationScenarioAcceptsProvisionalRow(t *testing.T) {
	if err := ValidateObservationScenario(Catalog[1]); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAcceptedRows(Catalog); err != nil {
		t.Fatalf("ValidateAcceptedRows(Catalog) error = %v, want accepted rows 01-07 to pass", err)
	}
	if err := ValidateFinalAll(Catalog); err == nil {
		t.Fatal("final-all catalog validation accepted provisional artifacts")
	}
}

func TestValidateAcceptedRowsRejectsRegistryDrift(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func([]Scenario) []Scenario
		acceptedIDs []string
		want        string
	}{
		{
			name: "missing row 01 artifact",
			mutate: func(rows []Scenario) []Scenario {
				rows[0].Observed.StateHash = ""
				return rows
			},
			acceptedIDs: acceptedScenarioIDs,
			want:        "scenario fullgame-01-sequential-bloodlust_conquest has non-observed immutable artifacts",
		},
		{
			name: "corrupt row 02 winner",
			mutate: func(rows []Scenario) []Scenario {
				rows[1].Observed.Winner = "account-missing"
				return rows
			},
			acceptedIDs: acceptedScenarioIDs,
			want:        "scenario fullgame-02-long_turn-diplomatic_world_union observed winner \"account-missing\" is not present in the declared slots",
		},
		{
			name:        "row 09 added to accepted registry without artifacts",
			mutate:      func(rows []Scenario) []Scenario { return rows },
			acceptedIDs: append(append([]string(nil), acceptedScenarioIDs...), "fullgame-09-sequential-score_2300"),
			want:        "scenario fullgame-09-sequential-score_2300 has non-observed immutable artifacts",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := cloneCatalogRows(Catalog)
			rows = test.mutate(rows)
			if err := validateAcceptedRows(rows, test.acceptedIDs); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validateAcceptedRows error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestValidateFinalAllRequiresAcceptedRowsToStayObserved(t *testing.T) {
	rows := cloneCatalogRows(Catalog)
	rows[0].Observed = ObservedArtifacts{}
	if err := ValidateFinalAll(rows); err == nil || !strings.Contains(err.Error(), "scenario fullgame-01-sequential-bloodlust_conquest has non-observed immutable artifacts") {
		t.Fatalf("ValidateFinalAll(rows) error = %v, want accepted row 01 artifact rejection", err)
	}
}

func TestValidateRequiresDimensionValueCoverage(t *testing.T) {
	rows := cloneCatalogRows(Catalog)
	for index := range rows {
		rows[index].World.Climate = "arid"
		rows[index].Observed = observedArtifacts("account-1", strings.Repeat("b", 64), strings.Repeat("c", 64))
	}
	if err := ValidateCatalogSchema(rows); err == nil || !strings.Contains(err.Error(), `climate value "cold"`) {
		t.Fatalf("ValidateCatalogSchema value coverage error = %v, want missing climate", err)
	}
}

func TestValidateFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]Scenario) []Scenario
		want   string
	}{
		{"count", func(rows []Scenario) []Scenario { return rows[:ExpectedCases-1] }, "want exactly"},
		{"duplicate id", func(rows []Scenario) []Scenario { rows[1].ID = rows[0].ID; return rows }, "duplicate scenario ID"},
		{"partial observed", func(rows []Scenario) []Scenario { rows[7].Observed.Winner = "account-1"; return rows }, "provisional observed artifacts"},
		{"slot order", func(rows []Scenario) []Scenario { rows[0].Slots[1].Slot = 3; return rows }, "out of order"},
		{"continuity missing script", func(rows []Scenario) []Scenario {
			rows[2].Rules.HumanScript = ""
			return rows
		}, "has non-fixed human script"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := cloneCatalogRows(Catalog)
			rows = test.mutate(rows)
			if err := ValidateCatalogSchema(rows); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateCatalogSchema error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func cloneCatalogRows(rows []Scenario) []Scenario {
	cloned := append([]Scenario(nil), rows...)
	for index := range cloned {
		cloned[index].Slots = append([]SetupSlot(nil), cloned[index].Slots...)
	}
	return cloned
}

func TestValidatePublicOnlyPayloadRejectsHiddenTargetingAndScriptedAI(t *testing.T) {
	for name, command := range map[string]ScriptCommand{
		"hidden tile": {Command: "move", ActorRole: "human-1", Payload: json.RawMessage(`{"unit_id":{"role":"human-1.unit-1"},"path":["tile-hidden"]}`)},
		"ai actor":    {Command: "fortify", ActorRole: "ai-1", Payload: json.RawMessage(`{"unit_id":{"role":"human-1.unit-1"}}`)},
	} {
		t.Run(name, func(t *testing.T) {
			if name == "ai actor" {
				if err := validateCheckedInHumanScript("scripts/military-zoc-combat.json"); err != nil {
					t.Fatal(err)
				}
				if isHumanScriptActor(command.ActorRole) {
					t.Fatal("AI actor unexpectedly classified as human")
				}
				return
			}
			if err := validatePublicOnlyPayload(command); err == nil {
				t.Fatal("hidden target payload was accepted")
			}
		})
	}
}
