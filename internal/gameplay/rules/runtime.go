package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const calendarFinalTurn = 48

type RuntimeRules struct {
	CompatibilityVersion string              `json:"compatibility_version"`
	ContentHash          string              `json:"content_hash"`
	Victory              []RuntimeVictory    `json:"victory"`
	Calendar             RuntimeCalendar     `json:"calendar"`
	Gaia                 RuntimeGaia         `json:"gaia"`
	Production           []RuntimeProduction `json:"production"`
	Effects              []RuntimeEffect     `json:"effects"`
	Starter              RuntimeStarter      `json:"starter"`
	Movement             RuntimeMovement     `json:"movement"`
}

type RuntimeMovement struct {
	Configured       bool             `json:"configured"`
	PolicyConfigured bool             `json:"-"`
	OpenAccess       bool             `json:"open_access"`
	ZOCStopsPath     bool             `json:"zoc_stops_path"`
	Terrains         []RuntimeTerrain `json:"terrains"`
}

type RuntimeTerrain struct {
	ID      string   `json:"id"`
	Cost    int      `json:"cost"`
	Domains []string `json:"domains"`
}

func (movement RuntimeMovement) Terrain(id string) (RuntimeTerrain, bool) {
	index := sort.Search(len(movement.Terrains), func(index int) bool { return movement.Terrains[index].ID >= id })
	if index == len(movement.Terrains) || movement.Terrains[index].ID != id {
		return RuntimeTerrain{}, false
	}
	return movement.Terrains[index], true
}

func (terrain RuntimeTerrain) Allows(domain string) bool {
	index := sort.SearchStrings(terrain.Domains, domain)
	return index < len(terrain.Domains) && terrain.Domains[index] == domain
}

type RuntimeStarter struct {
	Movement int                    `json:"movement"`
	Attack   int                    `json:"attack"`
	Defense  int                    `json:"defense"`
	Strength int                    `json:"strength"`
	Armor    int                    `json:"armor"`
	Health   int                    `json:"health"`
	Effects  []RuntimeStarterEffect `json:"effects"`
}

type RuntimeStarterEffect struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}

type RuntimeVictory struct {
	ID string `json:"id"`
}

type RuntimeCalendar struct {
	FinalTurn int `json:"final_turn"`
}

type RuntimeGaia struct {
	ActivationTurns int `json:"activation_turns"`
}

type RuntimeProduction struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type RuntimeEffect struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

func (runtime RuntimeRules) VictoryEvaluatorIDs() []string {
	ids := make([]string, len(runtime.Victory))
	for i, evaluator := range runtime.Victory {
		ids[i] = evaluator.ID
	}
	return ids
}

func containsVictoryEvaluator(evaluators []VictoryEvaluator, id string) bool {
	for _, evaluator := range evaluators {
		if evaluator.ID == id {
			return true
		}
	}
	return false
}

func (runtime RuntimeRules) ProductionIDs() string {
	ids := make([]string, len(runtime.Production))
	for i, production := range runtime.Production {
		ids[i] = production.ID
	}
	return strings.Join(ids, ",")
}

func (runtime RuntimeRules) EffectIDs() string {
	ids := make([]string, len(runtime.Effects))
	for i, effect := range runtime.Effects {
		ids[i] = effect.ID
	}
	return strings.Join(ids, ",")
}

func CanonicalDefinitions(definitions []byte) ([]byte, error) {
	if len(definitions) == 0 || !json.Valid(definitions) {
		return nil, errors.New("rules definitions are malformed")
	}
	var value any
	if err := json.Unmarshal(definitions, &value); err != nil {
		return nil, errors.New("rules definitions are malformed")
	}
	return json.Marshal(value)
}

func CompileRuntime(definitions []byte, contentHash string) (RuntimeRules, error) {
	canonical, err := CanonicalDefinitions(definitions)
	if err != nil {
		return RuntimeRules{}, err
	}
	digest := sha256.Sum256(canonical)
	if hex.EncodeToString(digest[:]) != contentHash {
		return RuntimeRules{}, errors.New("rules content hash mismatch")
	}
	var definition struct {
		CompatibilityVersion string                 `json:"compatibility_version"`
		Manifest                    SupportedRulesManifest `json:"manifest"`
		Records                     []Record               `json:"records"`
		EnabledVictoryEvaluatorIDs []string               `json:"enabled_victory_evaluator_ids,omitempty"`
	}
	if err := json.Unmarshal(definitions, &definition); err != nil {
		return RuntimeRules{}, errors.New("rules definitions are malformed")
	}
	if definition.CompatibilityVersion != CTP2CompatibilityVersion || definition.Manifest.CompatibilityVersion != CTP2CompatibilityVersion {
		return RuntimeRules{}, errors.New("rules compatibility version is unsupported")
	}
	if err := definition.Manifest.Validate(); err != nil {
		return RuntimeRules{}, fmt.Errorf("rules manifest is invalid: %w", err)
	}
	if len(definition.Records) == 0 {
		return RuntimeRules{}, errors.New("rules definitions contain no records")
	}

	runtime := RuntimeRules{
		CompatibilityVersion: definition.CompatibilityVersion,
		ContentHash:          contentHash,
		Calendar:             RuntimeCalendar{FinalTurn: calendarFinalTurn},
	}
	enabledVictories := make(map[string]struct{}, len(definition.Manifest.VictoryEvaluators))
	for _, evaluator := range definition.Manifest.VictoryEvaluators {
		enabledVictories[evaluator.ID] = struct{}{}
	}
	if len(definition.EnabledVictoryEvaluatorIDs) > 0 {
		enabledVictories = make(map[string]struct{}, len(definition.EnabledVictoryEvaluatorIDs))
		for _, evaluatorID := range definition.EnabledVictoryEvaluatorIDs {
			if _, selected := enabledVictories[evaluatorID]; selected {
				return RuntimeRules{}, fmt.Errorf("victory evaluator %q is duplicated", evaluatorID)
			}
			if !containsVictoryEvaluator(definition.Manifest.VictoryEvaluators, evaluatorID) {
				return RuntimeRules{}, fmt.Errorf("victory evaluator %q is unsupported", evaluatorID)
			}
			enabledVictories[evaluatorID] = struct{}{}
		}
	}
	for evaluatorID := range enabledVictories {
		runtime.Victory = append(runtime.Victory, RuntimeVictory{ID: evaluatorID})
	}

	seen := make(map[string]struct{}, len(definition.Records))
	for _, record := range definition.Records {
		if record.ID == "" || record.Kind == "" || record.Tags == nil {
			return RuntimeRules{}, errors.New("rules definition record is malformed")
		}
		if _, ok := seen[record.ID]; ok {
			return RuntimeRules{}, fmt.Errorf("rules definition record %q is duplicated", record.ID)
		}
		seen[record.ID] = struct{}{}
		if err := compileRecord(&runtime, record); err != nil {
			return RuntimeRules{}, err
		}
	}
	if runtime.Gaia.ActivationTurns < 1 {
		return RuntimeRules{}, errors.New("rules definitions require one Gaia endgame record")
	}
	if !containsRuntimeProduction(runtime.Production, "UNIT_SYNTHETIC") {
		return RuntimeRules{}, errors.New("rules definitions require the synthetic starter unit")
	}
	if runtime.Movement.Configured && (!runtime.Movement.PolicyConfigured || len(runtime.Movement.Terrains) == 0) {
		return RuntimeRules{}, errors.New("rules definitions require complete synthetic movement definitions")
	}
	runtime.Starter = RuntimeStarter{
		Movement: 2, Attack: 4, Defense: 4, Strength: 4, Armor: 2, Health: 10,
		Effects: []RuntimeStarterEffect{
			{ID: "settler", Kind: "settler"},
			{ID: "embassy", Kind: "special_embassy"},
			{ID: "steal", Kind: "special_steal", Value: 100},
			{ID: "capture", Kind: "special_capture", Value: 0},
			{ID: "terraform", Kind: "special_terraform", Value: 1},
			{ID: "terraform-turns", Kind: "special_terraform_turns", Value: 1},
		},
	}
	sort.Slice(runtime.Victory, func(i, j int) bool { return runtime.Victory[i].ID < runtime.Victory[j].ID })
	sort.Slice(runtime.Production, func(i, j int) bool { return runtime.Production[i].ID < runtime.Production[j].ID })
	sort.Slice(runtime.Effects, func(i, j int) bool { return runtime.Effects[i].ID < runtime.Effects[j].ID })
	sort.Slice(runtime.Movement.Terrains, func(i, j int) bool { return runtime.Movement.Terrains[i].ID < runtime.Movement.Terrains[j].ID })
	return runtime, nil
}

// ServerOwnedSyntheticRuleset is the fixed original synthetic catalog used when
// a server starts a gameplay aggregate. It is intentionally independent of
// client-visible setup options and contains no retail data.
func ServerOwnedSyntheticRuleset(victoryEvaluatorIDs ...string) (Ruleset, error) {
	manifest := SupportedManifest()
	records := []Record{
		{Kind: "endgame", ID: "ENDGAME_GAIA_VECTOR", Tags: map[string]string{"TurnsToActivate": "1"}},
		{Kind: "endgame", ID: "ENDGAME_UNION_VECTOR", Tags: map[string]string{"TurnsToActivate": "0"}},
		{Kind: "building", ID: "IMPROVE_SYNTHETIC", Tags: map[string]string{"Description": "IMPROVE_SYNTHETIC"}},
		{Kind: "terrain", ID: "forest", Tags: map[string]string{"MovementCost": "2", "MovementDomains": "amphibious,land"}},
		{Kind: "terrain", ID: "ocean", Tags: map[string]string{"MovementCost": "1", "MovementDomains": "amphibious,sea"}},
		{Kind: "terrain", ID: "plains", Tags: map[string]string{"MovementCost": "1", "MovementDomains": "amphibious,land"}},
		{Kind: "movement", ID: "MOVEMENT_SYNTHETIC", Tags: map[string]string{"OpenAccess": "true", "ZOCStopsPath": "true"}},
		{Kind: "unit", ID: "UNIT_SYNTHETIC", Tags: map[string]string{"Description": "UNIT_SYNTHETIC", "MaxMovePoints": "1"}},
		{Kind: "wonder", ID: "WONDER_SYNTHETIC", Tags: map[string]string{"Description": "WONDER_SYNTHETIC"}},
	}
	enabled := append([]string(nil), victoryEvaluatorIDs...)
	sort.Strings(enabled)
	canonical, err := json.Marshal(struct {
		CompatibilityVersion       string                 `json:"compatibility_version"`
		Manifest                   SupportedRulesManifest `json:"manifest"`
		Records                    []Record               `json:"records"`
		EnabledVictoryEvaluatorIDs []string               `json:"enabled_victory_evaluator_ids,omitempty"`
	}{CompatibilityVersion: CTP2CompatibilityVersion, Manifest: manifest, Records: records, EnabledVictoryEvaluatorIDs: enabled})
	if err != nil {
		return Ruleset{}, err
	}
	canonical, err = CanonicalDefinitions(canonical)
	if err != nil {
		return Ruleset{}, err
	}
	digest := sha256.Sum256(canonical)
	ruleset := Ruleset{CompatibilityVersion: CTP2CompatibilityVersion, Manifest: manifest, Records: records, CanonicalJSON: canonical, ContentHash: hex.EncodeToString(digest[:])}
	if _, err := CompileRuntime(ruleset.CanonicalJSON, ruleset.ContentHash); err != nil {
		return Ruleset{}, err
	}
	return ruleset, nil
}

func compileRecord(runtime *RuntimeRules, record Record) error {
	switch record.Kind {
	case "unit":
		if err := requirePositiveInteger(record, "MaxMovePoints"); err != nil {
			return err
		}
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: record.ID, Kind: "unit"})
	case "building":
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: record.ID, Kind: "building"})
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: record.ID, Kind: "building"})
	case "wonder":
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: record.ID, Kind: "wonder"})
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: record.ID, Kind: "wonder"})
	case "terrain":
		if _, ok := record.Tags["MovementCost"]; !ok {
			return nil
		}
		cost, err := positiveInteger(record, "MovementCost")
		if err != nil {
			return err
		}
		domains, err := movementDomains(record)
		if err != nil {
			return err
		}
		runtime.Movement.Configured = true
		runtime.Movement.Terrains = append(runtime.Movement.Terrains, RuntimeTerrain{ID: record.ID, Cost: cost, Domains: domains})
	case "movement":
		if record.ID != "MOVEMENT_SYNTHETIC" || runtime.Movement.PolicyConfigured {
			return fmt.Errorf("rules definition record %q has invalid movement policy", record.ID)
		}
		openAccess, err := boolean(record, "OpenAccess")
		if err != nil {
			return err
		}
		zocStopsPath, err := boolean(record, "ZOCStopsPath")
		if err != nil {
			return err
		}
		runtime.Movement.Configured = true
		runtime.Movement.PolicyConfigured = true
		runtime.Movement.OpenAccess = openAccess
		runtime.Movement.ZOCStopsPath = zocStopsPath
	case "endgame":
		turns, err := nonNegativeInteger(record, "TurnsToActivate")
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(record.ID), "gaia") {
			if runtime.Gaia.ActivationTurns != 0 {
				return errors.New("rules definitions contain multiple Gaia endgame records")
			}
			if turns < 1 {
				return fmt.Errorf("rules definition record %q has invalid TurnsToActivate", record.ID)
			}
			runtime.Gaia = RuntimeGaia{ActivationTurns: turns}
		}
	default:
		return fmt.Errorf("rules definition record %q has unknown kind %q", record.ID, record.Kind)
	}
	return nil
}

func containsRuntimeProduction(production []RuntimeProduction, id string) bool {
	for _, item := range production {
		if item.ID == id {
			return true
		}
	}
	return false
}

func requirePositiveInteger(record Record, tag string) error {
	_, err := positiveInteger(record, tag)
	return err
}

func positiveInteger(record Record, tag string) (int, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return 0, fmt.Errorf("rules definition record %q is missing %s", record.ID, tag)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}

func nonNegativeInteger(record Record, tag string) (int, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return 0, fmt.Errorf("rules definition record %q is missing %s", record.ID, tag)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}

func movementDomains(record Record) ([]string, error) {
	value, ok := record.Tags["MovementDomains"]
	if !ok || value == "" {
		return nil, fmt.Errorf("rules definition record %q is missing MovementDomains", record.ID)
	}
	domains := strings.Split(value, ",")
	for _, domain := range domains {
		if domain != "amphibious" && domain != "land" && domain != "sea" {
			return nil, fmt.Errorf("rules definition record %q has invalid MovementDomains", record.ID)
		}
	}
	sort.Strings(domains)
	for index := 1; index < len(domains); index++ {
		if domains[index] == domains[index-1] {
			return nil, fmt.Errorf("rules definition record %q has duplicate MovementDomains", record.ID)
		}
	}
	return domains, nil
}

func boolean(record Record, tag string) (bool, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return false, fmt.Errorf("rules definition record %q is missing %s", record.ID, tag)
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}
