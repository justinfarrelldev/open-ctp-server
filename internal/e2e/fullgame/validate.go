package fullgame

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var sha256Hex = regexp.MustCompile(`^[a-f0-9]{64}$`)

// ValidateObservationScenario validates a provisional row's fixed inputs before an artifact-observation run.
// It intentionally does not accept a catalog or relax Validate's final all-row requirements.
func ValidateObservationScenario(s Scenario) error {
	if strings.TrimSpace(s.ID) == "" || !strings.HasPrefix(s.ID, "fullgame-") {
		return fmt.Errorf("scenario has invalid fixed ID %q", s.ID)
	}
	if s.Seed == 0 {
		return fmt.Errorf("scenario %s has invalid fixed seed", s.ID)
	}
	if !known(s.TurnMode, modes) || !known(s.World.MapSize, mapSizes) || !known(s.World.Climate, climates) || !known(s.Rules.Victory, victories) {
		return fmt.Errorf("scenario %s contains unsupported coverage value", s.ID)
	}
	if !containsInt(wetnesses, s.World.Wetness) || !containsInt(countryCounts, len(s.Slots)) {
		return fmt.Errorf("scenario %s contains unsupported numeric coverage value", s.ID)
	}
	if !knownHumanPolicy(s.HumanPolicy) {
		return fmt.Errorf("scenario %s has unsupported human policy %q", s.ID, s.HumanPolicy)
	}
	if s.AutonomousAIPolicy != ProductionAutonomousAI {
		return fmt.Errorf("scenario %s has unsupported autonomous AI policy %q", s.ID, s.AutonomousAIPolicy)
	}
	selectable, err := campaignPlayableCivilizations()
	if err != nil {
		return err
	}
	if err := validateSetupSlots(s.Slots, selectable); err != nil {
		return err
	}
	if err := validateRulesFixture(s); err != nil {
		return err
	}
	if err := validateRestartPlanForCatalogScenario(s); err != nil {
		return err
	}
	if s.Terminal.Reason != s.Rules.Victory {
		return fmt.Errorf("scenario %s terminal reason %q does not match rules victory %q", s.ID, s.Terminal.Reason, s.Rules.Victory)
	}
	if s.hasObservedArtifacts() {
		if !s.hasCompleteObservedArtifacts() {
			return fmt.Errorf("scenario %s has provisional observed artifacts", s.ID)
		}
		if !winnerBelongsToSlots(s.Observed.Winner, s.Slots) {
			return fmt.Errorf("scenario %s observed winner %q is not present in the declared slots", s.ID, s.Observed.Winner)
		}
		for _, hash := range []string{s.Observed.StateHash, s.Observed.EventHash} {
			if !sha256Hex.MatchString(hash) {
				return fmt.Errorf("scenario %s has invalid observed SHA-256 artifact %q", s.ID, hash)
			}
		}
	}
	return nil
}

func ValidateCatalogSchema(catalog []Scenario) error {
	if len(catalog) != ExpectedCases {
		return fmt.Errorf("fixed full-game catalog has %d cases, want exactly %d", len(catalog), ExpectedCases)
	}
	dimensions := []struct {
		name   string
		values []string
	}{
		{"mode", modes},
		{"map", mapSizes},
		{"climate", climates},
		{"wetness", []string{"10", "50", "90"}},
		{"countries", []string{"2", "3", "4", "6", "8"}},
		{"victory", victories},
	}
	ids := make(map[string]struct{}, len(catalog))
	seeds := make(map[uint64]struct{}, len(catalog))
	seen := make(map[string]map[string]bool)
	modeVictory := make(map[string]bool)
	for index, scenario := range catalog {
		if strings.TrimSpace(scenario.ID) == "" || !strings.HasPrefix(scenario.ID, "fullgame-") {
			return fmt.Errorf("scenario %d has invalid fixed ID %q", index+1, scenario.ID)
		}
		if _, ok := ids[scenario.ID]; ok {
			return fmt.Errorf("duplicate scenario ID %q", scenario.ID)
		}
		ids[scenario.ID] = struct{}{}
		if scenario.Seed != scenarioSeed(index+1) {
			return fmt.Errorf("scenario %s seed %#x is not fixed catalog seed %#x", scenario.ID, scenario.Seed, scenarioSeed(index+1))
		}
		if _, ok := seeds[scenario.Seed]; ok {
			return fmt.Errorf("duplicate scenario seed %#x", scenario.Seed)
		}
		seeds[scenario.Seed] = struct{}{}
		if err := ValidateObservationScenario(scenario); err != nil {
			return err
		}
		values := map[string]string{
			"mode":      scenario.TurnMode,
			"map":       scenario.World.MapSize,
			"climate":   scenario.World.Climate,
			"wetness":   fmt.Sprint(scenario.World.Wetness),
			"countries": fmt.Sprint(len(scenario.Slots)),
			"victory":   scenario.Rules.Victory,
		}
		for _, dimension := range dimensions {
			observe(seen, dimension.name, values[dimension.name])
		}
		modeVictory[scenario.TurnMode+"/"+scenario.Rules.Victory] = true
	}
	for _, dimension := range dimensions {
		for _, value := range dimension.values {
			if !seen[dimension.name][value] {
				return fmt.Errorf("catalog misses required %s value %q", dimension.name, value)
			}
		}
	}
	for _, mode := range modes {
		for _, victory := range victories {
			if !modeVictory[mode+"/"+victory] {
				return fmt.Errorf("catalog misses required mode/victory pair %s/%s", mode, victory)
			}
		}
	}
	return nil
}

var acceptedScenarioIDs = []string{
	"fullgame-01-sequential-bloodlust_conquest",
	"fullgame-02-long_turn-diplomatic_world_union",
	"fullgame-03-simultaneous-score_2300",
	"fullgame-04-hybrid-science_gaia_controller",
	"fullgame-05-sequential-diplomatic_world_union",
	"fullgame-06-long_turn-score_2300",
	"fullgame-07-simultaneous-science_gaia_controller",
}

func ValidateRows01_12Final(rows []Scenario) error {
	return validateAcceptedRows(rows, acceptedScenarioIDs)
}

func ValidateAcceptedRows(catalog []Scenario) error {
	if err := ValidateCatalogSchema(catalog); err != nil {
		return err
	}
	return validateAcceptedRows(catalog, acceptedScenarioIDs)
}

func ValidateFinalAll(catalog []Scenario) error {
	if err := ValidateAcceptedRows(catalog); err != nil {
		return err
	}
	for _, scenario := range catalog {
		if err := validateScenarioRequiresObservedArtifacts(scenario); err != nil {
			return err
		}
	}
	return nil
}

func Validate(catalog []Scenario) error {
	return ValidateFinalAll(catalog)
}

func validateAcceptedRows(catalog []Scenario, acceptedIDs []string) error {
	if _, err := validateAcceptedRowRegistry(catalog, acceptedIDs); err != nil {
		return err
	}
	for _, acceptedID := range acceptedIDs {
		scenario, err := scenarioByID(catalog, acceptedID)
		if err != nil {
			return err
		}
		if err := validateScenarioRequiresObservedArtifacts(scenario); err != nil {
			return err
		}
	}
	return nil
}

func validateAcceptedRowRegistry(catalog []Scenario, acceptedIDs []string) (map[string]struct{}, error) {
	accepted := make(map[string]struct{}, len(acceptedIDs))
	for _, acceptedID := range acceptedIDs {
		if _, ok := accepted[acceptedID]; ok {
			return nil, fmt.Errorf("accepted scenario %s is registered more than once", acceptedID)
		}
		if _, err := scenarioByID(catalog, acceptedID); err != nil {
			return nil, fmt.Errorf("accepted scenario %s is missing from catalog", acceptedID)
		}
		accepted[acceptedID] = struct{}{}
	}
	return accepted, nil
}

func scenarioByID(catalog []Scenario, id string) (Scenario, error) {
	for _, scenario := range catalog {
		if scenario.ID == id {
			return scenario, nil
		}
	}
	return Scenario{}, fmt.Errorf("scenario %s is missing from catalog", id)
}

func validateScenarioRequiresObservedArtifacts(scenario Scenario) error {
	if err := ValidateObservationScenario(scenario); err != nil {
		if !scenario.hasCompleteObservedArtifacts() {
			return fmt.Errorf("scenario %s has non-observed immutable artifacts", scenario.ID)
		}
		return err
	}
	if !scenario.hasCompleteObservedArtifacts() {
		return fmt.Errorf("scenario %s has non-observed immutable artifacts", scenario.ID)
	}
	return nil
}

func validateRulesFixture(s Scenario) error {
	if strings.TrimSpace(s.Rules.Victory) == "" {
		return fmt.Errorf("scenario %s has invalid rules fixture", s.ID)
	}
	rulesHash, err := acceptanceRulesHash(s.Rules.Victory)
	if err != nil {
		return fmt.Errorf("scenario %s rules fixture: %w", s.ID, err)
	}
	if s.Rules.Hash != rulesHash {
		return fmt.Errorf("scenario %s rules hash %q does not match production rules pin %q", s.ID, s.Rules.Hash, rulesHash)
	}
	if s.HumanPolicy == HumanPolicyObserveOnly {
		if s.Rules.HumanScript != "" {
			return fmt.Errorf("scenario %s observe-only policy must not declare a human script", s.ID)
		}
		return nil
	}
	if !strings.HasPrefix(s.Rules.HumanScript, "scripts/") || !strings.HasSuffix(s.Rules.HumanScript, ".json") {
		return fmt.Errorf("scenario %s has non-fixed human script %q", s.ID, s.Rules.HumanScript)
	}
	return validateCheckedInHumanScript(s.Rules.HumanScript)
}

func validateRestartPlanForCatalogScenario(s Scenario) error {
	if s.Budgets.Commands <= 0 || s.Budgets.Turns <= 0 || s.Budgets.Deadline <= 0 {
		return fmt.Errorf("scenario %s has invalid execution bounds", s.ID)
	}
	switch s.HumanPolicy {
	case HumanPolicyConquest:
		if s.Restart.CompletedTurn != 0 {
			return fmt.Errorf("scenario %s conquest restart plan does not accept a completed turn", s.ID)
		}
		if s.Restart.Point != AcceptanceRestartDisabled && s.Restart.Point != AcceptanceRestartAfterDurableRound {
			return fmt.Errorf("scenario %s conquest restart point %q is unsupported", s.ID, s.Restart.Point)
		}
	case HumanPolicyDiplomacy, HumanPolicyScriptedHumans, HumanPolicyObserveOnly, HumanPolicyContinuity:
		switch s.Restart.Point {
		case AcceptanceRestartDisabled:
			if s.Restart.CompletedTurn != 0 {
				return fmt.Errorf("scenario %s disabled restart plan has a completed turn", s.ID)
			}
		case AcceptanceRestartAfterCompletedTurn:
			if s.Restart.CompletedTurn < 1 || s.Restart.CompletedTurn >= s.Budgets.Turns {
				return fmt.Errorf("scenario %s completed-turn restart is outside the turn budget", s.ID)
			}
		default:
			return fmt.Errorf("scenario %s restart point %q is unsupported", s.ID, s.Restart.Point)
		}
	}
	return nil
}

func validateCheckedInHumanScript(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var script Script
	if err := json.Unmarshal(body, &script); err != nil {
		return err
	}
	if script.Version != 1 || len(script.Commands) == 0 {
		return fmt.Errorf("script %s is invalid", path)
	}
	for _, command := range script.Commands {
		if !isHumanScriptActor(command.ActorRole) {
			return fmt.Errorf("script %s directs non-human actor %q", path, command.ActorRole)
		}
		if err := validatePublicOnlyPayload(command); err != nil {
			return fmt.Errorf("script %s %s: %w", path, command.Command, err)
		}
	}
	return nil
}

func validatePublicOnlyPayload(command ScriptCommand) error {
	if !json.Valid(command.Payload) {
		return fmt.Errorf("payload is not valid JSON")
	}
	if command.Command == "leave" || command.Command == "pass" || command.Command == "round-complete" {
		var payload map[string]any
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return err
		}
		if len(payload) != 0 {
			return fmt.Errorf("%s payload must be empty", command.Command)
		}
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(command.Payload, &payload); err != nil {
		return err
	}
	switch command.Command {
	case "army-create":
		if !payloadRoleRefArray(payload["unit_ids"]) {
			return fmt.Errorf("unit_ids must use public role references")
		}
	case "fortify", "found-city", "move", "terraform", "establish-embassy", "steal-technology":
		if !payloadRoleRef(payload["unit_id"]) {
			return fmt.Errorf("unit_id must use a public role reference")
		}
	}
	switch command.Command {
	case "move":
		if !payloadRoleRefArray(payload["path"]) {
			return fmt.Errorf("path must use public role references")
		}
	case "terraform", "cancel-terraform":
		if !payloadRoleRef(payload["tile_id"]) {
			return fmt.Errorf("tile_id must use a public role reference")
		}
	case "establish-embassy", "steal-technology":
		if !payloadRoleRef(payload["city_id"]) {
			return fmt.Errorf("city_id must use a public role reference")
		}
	}
	return nil
}

func payloadRoleRef(value any) bool {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 1 {
		return false
	}
	role, ok := object["role"].(string)
	return ok && isHumanScriptActor(strings.Split(role, ".")[0])
}

func payloadRoleRefArray(value any) bool {
	array, ok := value.([]any)
	if !ok || len(array) == 0 {
		return false
	}
	for _, item := range array {
		if !payloadRoleRef(item) {
			return false
		}
	}
	return true
}

func observe(seen map[string]map[string]bool, dimension, value string) {
	if seen[dimension] == nil {
		seen[dimension] = make(map[string]bool)
	}
	seen[dimension][value] = true
}

func containsInt(values []int, value int) bool {
	sortedValues := append([]int(nil), values...)
	sort.Ints(sortedValues)
	index := sort.SearchInts(sortedValues, value)
	return index < len(sortedValues) && sortedValues[index] == value
}
