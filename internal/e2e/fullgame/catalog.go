// Package fullgame defines the checked-in deterministic complete-game catalog.
package fullgame

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	gameplayrules "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	ExpectedCases          = 60
	seedBase               = uint64(0x4354503200000000)
	ProductionAutonomousAI = "production-autonomous"
)

type AcceptanceWorldFixture struct {
	MapSize, Climate string
	Wetness          int
}

type AcceptanceRulesFixture struct {
	Hash        string
	Victory     string
	HumanScript string
}

type AcceptanceBudgets struct {
	Commands         int
	Turns            int
	Deadline         time.Duration
	StartingSettlers int
}

type AcceptanceRestartPoint string

const (
	AcceptanceRestartDisabled           AcceptanceRestartPoint = "disabled"
	AcceptanceRestartAfterDurableRound  AcceptanceRestartPoint = "after-durable-round"
	AcceptanceRestartAfterCompletedTurn AcceptanceRestartPoint = "after-completed-turn"
)

type AcceptanceRestartPlan struct {
	Point         AcceptanceRestartPoint
	CompletedTurn int
}

type HumanPolicyKind string

const (
	HumanPolicyConquest       HumanPolicyKind = "public-conquest"
	HumanPolicyDiplomacy      HumanPolicyKind = "public-diplomacy"
	HumanPolicyScriptedHumans HumanPolicyKind = "scripted-humans"
	HumanPolicyObserveOnly    HumanPolicyKind = "observe-only"
	HumanPolicyContinuity     HumanPolicyKind = "leave-continuity"
)

type TerminalOracle struct {
	Reason string
}

type ObservedArtifacts struct {
	Winner    string
	StateHash string
	EventHash string
}

type Scenario struct {
	ID                 string
	Name               string
	Seed               uint64
	TurnMode           string
	Slots              []SetupSlot
	World              AcceptanceWorldFixture
	Rules              AcceptanceRulesFixture
	HumanPolicy        HumanPolicyKind
	AutonomousAIPolicy string
	Restart            AcceptanceRestartPlan
	Budgets            AcceptanceBudgets
	Terminal           TerminalOracle
	Observed           ObservedArtifacts
}

var modes = []string{"sequential", "long_turn", "simultaneous", "hybrid"}
var mapSizes = []string{"duel", "small", "medium", "large", "huge"}
var climates = []string{"arid", "cold", "temperate", "tropical"}
var wetnesses = []int{10, 50, 90}
var countryCounts = []int{2, 3, 4, 6, 8}
var victories = []string{"bloodlust_conquest", "diplomatic_world_union", "science_gaia_controller", "score_2300"}
var humanPolicies = []HumanPolicyKind{HumanPolicyConquest, HumanPolicyDiplomacy, HumanPolicyScriptedHumans, HumanPolicyObserveOnly, HumanPolicyContinuity}

const (
	humanController = engine.ControllerHuman
	aiController    = engine.ControllerAI
)

// Catalog is intentionally written as fixed values: it is not sampled, shuffled, or generated at test time.
var Catalog = assembleCatalog()

func assembleCatalog() []Scenario {
	chunks := [][]Scenario{
		catalogRows01_12(),
		catalogRows13_24(),
		catalogRows25_36(),
		catalogRows37_48(),
		catalogRows49_60(),
	}
	catalog := make([]Scenario, 0, ExpectedCases)
	for _, chunk := range chunks {
		catalog = append(catalog, chunk...)
	}
	return catalog
}

func CatalogRows01_12() []Scenario { return append([]Scenario(nil), catalogRows01_12()...) }
func CatalogRows13_24() []Scenario { return append([]Scenario(nil), catalogRows13_24()...) }
func CatalogRows25_36() []Scenario { return append([]Scenario(nil), catalogRows25_36()...) }
func CatalogRows37_48() []Scenario { return append([]Scenario(nil), catalogRows37_48()...) }
func CatalogRows49_60() []Scenario { return append([]Scenario(nil), catalogRows49_60()...) }

func acceptanceRulesJSON(fixture AcceptanceRulesFixture) string {
	options, _ := json.Marshal([]string{"victory:" + fixture.Victory})
	return string(options)
}

func acceptanceRulesHash(victory string) (string, error) {
	ruleset, err := gameplayrules.ServerOwnedSyntheticRuleset(victory)
	if err != nil {
		return "", err
	}
	return ruleset.ContentHash, nil
}

func mustRulesFixture(victory, humanScript string) AcceptanceRulesFixture {
	hash, err := acceptanceRulesHash(victory)
	if err != nil {
		panic(err)
	}
	return AcceptanceRulesFixture{Hash: hash, Victory: victory, HumanScript: humanScript}
}

func defaultAcceptanceBudgets() AcceptanceBudgets {
	return AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 2}
}

func durableRoundRestart() AcceptanceRestartPlan {
	return AcceptanceRestartPlan{Point: AcceptanceRestartAfterDurableRound}
}

func completedTurnRestart(turn int) AcceptanceRestartPlan {
	return AcceptanceRestartPlan{Point: AcceptanceRestartAfterCompletedTurn, CompletedTurn: turn}
}

func disabledRestart() AcceptanceRestartPlan {
	return AcceptanceRestartPlan{Point: AcceptanceRestartDisabled}
}

func observedArtifacts(winner, stateHash, eventHash string) ObservedArtifacts {
	return ObservedArtifacts{Winner: winner, StateHash: stateHash, EventHash: eventHash}
}

func scenarioSeed(number int) uint64 {
	switch number {
	case 1:
		return seedBase + 3
	case 2:
		return seedBase + 2
	case 3:
		return seedBase + 1
	default:
		return seedBase + uint64(number)
	}
}

func catalogScenario(number int, turnMode string, world AcceptanceWorldFixture, slots []SetupSlot, humanPolicy HumanPolicyKind, victory, humanScript string, restart AcceptanceRestartPlan, observed ObservedArtifacts) Scenario {
	id := fmt.Sprintf("fullgame-%02d-%s-%s", number, turnMode, victory)
	return Scenario{
		ID:                 id,
		Name:               id,
		Seed:               scenarioSeed(number),
		TurnMode:           turnMode,
		Slots:              slots,
		World:              world,
		Rules:              mustRulesFixture(victory, humanScript),
		HumanPolicy:        humanPolicy,
		AutonomousAIPolicy: ProductionAutonomousAI,
		Restart:            restart,
		Budgets:            defaultAcceptanceBudgets(),
		Terminal:           TerminalOracle{Reason: victory},
		Observed:           observed,
	}
}

func sorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func known(value string, values []string) bool {
	sortedValues := sorted(values)
	index := sort.SearchStrings(sortedValues, value)
	return strings.TrimSpace(value) != "" && index < len(sortedValues) && sortedValues[index] == value
}

func knownHumanPolicy(value HumanPolicyKind) bool {
	for _, want := range humanPolicies {
		if value == want {
			return true
		}
	}
	return false
}

func winnerBelongsToSlots(winner string, slots []SetupSlot) bool {
	if !strings.HasPrefix(winner, "account-") {
		return false
	}
	for index := range slots {
		if winner == fmt.Sprintf("account-%d", index+1) {
			return true
		}
	}
	return false
}

func (s Scenario) hasObservedArtifacts() bool {
	return strings.TrimSpace(s.Observed.Winner) != "" || strings.TrimSpace(s.Observed.StateHash) != "" || strings.TrimSpace(s.Observed.EventHash) != ""
}

func (s Scenario) hasCompleteObservedArtifacts() bool {
	return strings.TrimSpace(s.Observed.Winner) != "" && strings.TrimSpace(s.Observed.StateHash) != "" && strings.TrimSpace(s.Observed.EventHash) != ""
}
