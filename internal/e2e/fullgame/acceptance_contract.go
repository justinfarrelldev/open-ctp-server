package fullgame

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/puppet"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

type AcceptanceHumanPolicy interface {
	Name() string
	validateRestart(AcceptanceRestartPlan, AcceptanceBudgets) error
	execute(context.Context, AcceptanceConfig, *testkit.Scenario) (Artifacts, error)
}

type ConquestHumanPolicy struct{}

func (ConquestHumanPolicy) Name() string { return string(HumanPolicyConquest) }

func (ConquestHumanPolicy) validateRestart(plan AcceptanceRestartPlan, _ AcceptanceBudgets) error {
	if plan.CompletedTurn != 0 {
		return errors.New("conquest restart plan does not accept a completed turn")
	}
	if plan.Point != AcceptanceRestartDisabled && plan.Point != AcceptanceRestartAfterDurableRound {
		return fmt.Errorf("conquest restart point %q is unsupported", plan.Point)
	}
	return nil
}

func (ConquestHumanPolicy) execute(ctx context.Context, config AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	return executeConquestAcceptance(ctx, config, kit)
}

type DiplomacyHumanPolicy struct{}

func (DiplomacyHumanPolicy) Name() string { return string(HumanPolicyDiplomacy) }

func (DiplomacyHumanPolicy) validateRestart(plan AcceptanceRestartPlan, budgets AcceptanceBudgets) error {
	switch plan.Point {
	case AcceptanceRestartDisabled:
		if plan.CompletedTurn != 0 {
			return errors.New("disabled diplomacy restart plan has a completed turn")
		}
	case AcceptanceRestartAfterCompletedTurn:
		if plan.CompletedTurn < 1 || plan.CompletedTurn >= budgets.Turns {
			return errors.New("diplomacy restart completed turn is outside the turn budget")
		}
	default:
		return fmt.Errorf("diplomacy restart point %q is unsupported", plan.Point)
	}
	return nil
}

func (DiplomacyHumanPolicy) execute(ctx context.Context, config AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	return executeDiplomaticAcceptance(ctx, config, kit)
}

type ScriptedHumanPolicy struct {
	Level puppet.TestHarnessPolicyLevel
}

func (policy ScriptedHumanPolicy) Name() string { return string(HumanPolicyScriptedHumans) }

func (policy ScriptedHumanPolicy) validateRestart(plan AcceptanceRestartPlan, budgets AcceptanceBudgets) error {
	switch plan.Point {
	case AcceptanceRestartDisabled:
		if plan.CompletedTurn != 0 {
			return errors.New("disabled scripted-human restart plan has a completed turn")
		}
	case AcceptanceRestartAfterCompletedTurn:
		if plan.CompletedTurn < 1 || plan.CompletedTurn >= budgets.Turns {
			return errors.New("scripted-human restart completed turn is outside the turn budget")
		}
	default:
		return fmt.Errorf("scripted-human restart point %q is unsupported", plan.Point)
	}
	if policy.Level != puppet.TestHarnessPolicyLevelProductionFirst && policy.Level != puppet.TestHarnessPolicyLevelPassOnly {
		return fmt.Errorf("scripted-human test harness level %q is unsupported", policy.Level)
	}
	return nil
}

func (policy ScriptedHumanPolicy) execute(ctx context.Context, config AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	return executeScriptedHumanAcceptance(ctx, config, kit, policy.Level)
}

type ObserveOnlyPolicy struct{}

type ContinuityHumanPolicy struct{}

func (ObserveOnlyPolicy) Name() string { return string(HumanPolicyObserveOnly) }

func (ContinuityHumanPolicy) Name() string { return string(HumanPolicyContinuity) }

func (ObserveOnlyPolicy) validateRestart(plan AcceptanceRestartPlan, budgets AcceptanceBudgets) error {
	switch plan.Point {
	case AcceptanceRestartDisabled:
		return nil
	case AcceptanceRestartAfterCompletedTurn:
		if plan.CompletedTurn < 1 || plan.CompletedTurn >= budgets.Turns {
			return errors.New("observe-only restart completed turn is outside the turn budget")
		}
		return nil
	default:
		return fmt.Errorf("observe-only restart point %q is unsupported", plan.Point)
	}
}

func (ObserveOnlyPolicy) execute(ctx context.Context, config AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	return executeObserveOnlyAcceptance(ctx, config, kit)
}

func (ContinuityHumanPolicy) validateRestart(plan AcceptanceRestartPlan, budgets AcceptanceBudgets) error {
	switch plan.Point {
	case AcceptanceRestartDisabled:
		if plan.CompletedTurn != 0 {
			return errors.New("disabled continuity restart plan has a completed turn")
		}
	case AcceptanceRestartAfterCompletedTurn:
		if plan.CompletedTurn < 1 || plan.CompletedTurn >= budgets.Turns {
			return errors.New("continuity restart completed turn is outside the turn budget")
		}
	default:
		return fmt.Errorf("continuity restart point %q is unsupported", plan.Point)
	}
	return nil
}

func (ContinuityHumanPolicy) execute(ctx context.Context, config AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	return executeContinuityAcceptance(ctx, config, kit)
}

type AcceptanceRestartEvent struct {
	Point         AcceptanceRestartPoint
	CompletedTurn int
}

type AcceptanceConfig struct {
	Name                                                                 string
	Seed                                                                 uint64
	TurnMode                                                             string
	Slots                                                                []SetupSlot
	World                                                                AcceptanceWorldFixture
	Rules                                                                AcceptanceRulesFixture
	HumanPolicy                                                          AcceptanceHumanPolicy
	AutonomousAIPolicy                                                   string
	Budgets                                                              AcceptanceBudgets
	Restart                                                              AcceptanceRestartPlan
	ObserveRestart                                                       func(AcceptanceRestartEvent)
	ExpectedWinner, ExpectedReason, ExpectedStateHash, ExpectedEventHash string
}

func (config AcceptanceConfig) validate() error {
	if config.Name == "" || config.Seed == 0 || config.TurnMode == "" {
		return errors.New("acceptance identity, seed, and turn mode are required")
	}
	if config.HumanPolicy == nil || config.HumanPolicy.Name() == "" {
		return errors.New("acceptance human harness policy is required")
	}
	if config.AutonomousAIPolicy != ProductionAutonomousAI {
		return fmt.Errorf("acceptance AI policy %q is not production autonomous AI", config.AutonomousAIPolicy)
	}
	if config.World.MapSize == "" || config.World.Climate == "" || config.World.Wetness < 0 || config.World.Wetness > 100 {
		return errors.New("acceptance world fixture is invalid")
	}
	if config.Rules.Victory == "" {
		return errors.New("acceptance rules fixture is invalid")
	}
	if config.HumanPolicy.Name() != string(HumanPolicyObserveOnly) && config.Rules.HumanScript == "" {
		return errors.New("acceptance rules fixture is invalid")
	}
	rulesHash, err := acceptanceRulesHash(config.Rules.Victory)
	if err != nil {
		return fmt.Errorf("compile acceptance rules fixture: %w", err)
	}
	if config.Rules.Hash != rulesHash {
		return fmt.Errorf("acceptance rules hash %q does not match production rules pin %q", config.Rules.Hash, rulesHash)
	}
	if config.Budgets.Commands < 1 || config.Budgets.Turns < 1 || config.Budgets.Deadline <= 0 || config.Budgets.StartingSettlers < 1 {
		return errors.New("acceptance command, turn, and time budgets are required")
	}
	if err := config.HumanPolicy.validateRestart(config.Restart, config.Budgets); err != nil {
		return fmt.Errorf("acceptance restart plan is invalid: %w", err)
	}
	selectable, err := campaignPlayableCivilizations()
	if err != nil {
		return err
	}
	if err := validateSetupSlots(config.Slots, selectable); err != nil {
		return err
	}
	humans, ais := 0, 0
	for _, slot := range config.Slots {
		if slot.ControllerKind == engine.ControllerHuman {
			humans++
		}
		if slot.ControllerKind == engine.ControllerAI {
			ais++
		}
	}
	switch config.HumanPolicy.Name() {
	case string(HumanPolicyObserveOnly):
		if humans != 0 || ais == 0 {
			return errors.New("observe-only acceptance requires only autonomous AI slots")
		}
	case string(HumanPolicyContinuity):
		if humans == 0 {
			return errors.New("continuity acceptance requires at least one human slot")
		}
	default:
		if humans == 0 || ais == 0 {
			return errors.New("acceptance requires public human and autonomous AI slots")
		}
	}
	if config.ExpectedWinner != "" && !winnerBelongsToSlots(config.ExpectedWinner, config.Slots) {
		return fmt.Errorf("acceptance winner %q is not present in the declared slot roster", config.ExpectedWinner)
	}
	if config.ExpectedReason != "" && config.ExpectedReason != config.Rules.Victory {
		return fmt.Errorf("acceptance terminal reason %q does not match declared rules victory %q", config.ExpectedReason, config.Rules.Victory)
	}
	return nil
}

func (scenario Scenario) AcceptanceConfig() (AcceptanceConfig, error) {
	policy, err := acceptancePolicyForScenario(scenario)
	if err != nil {
		return AcceptanceConfig{}, err
	}
	return AcceptanceConfig{
		Name:               scenario.Name,
		Seed:               scenario.Seed,
		TurnMode:           scenario.TurnMode,
		Slots:              append([]SetupSlot(nil), scenario.Slots...),
		World:              scenario.World,
		Rules:              scenario.Rules,
		HumanPolicy:        policy,
		AutonomousAIPolicy: scenario.AutonomousAIPolicy,
		Budgets:            scenario.Budgets,
		Restart:            scenario.Restart,
		ExpectedWinner:     scenario.Observed.Winner,
		ExpectedReason:     scenario.Terminal.Reason,
		ExpectedStateHash:  scenario.Observed.StateHash,
		ExpectedEventHash:  scenario.Observed.EventHash,
	}, nil
}

func acceptancePolicyForScenario(s Scenario) (AcceptanceHumanPolicy, error) {
	switch s.HumanPolicy {
	case HumanPolicyConquest:
		return ConquestHumanPolicy{}, nil
	case HumanPolicyDiplomacy:
		return DiplomacyHumanPolicy{}, nil
	case HumanPolicyScriptedHumans:
		return ScriptedHumanPolicy{Level: puppet.TestHarnessPolicyLevelProductionFirst}, nil
	case HumanPolicyObserveOnly:
		return ObserveOnlyPolicy{}, nil
	case HumanPolicyContinuity:
		return ContinuityHumanPolicy{}, nil
	default:
		return nil, fmt.Errorf("scenario %s has unsupported human policy %q", s.ID, s.HumanPolicy)
	}
}

func RowOneConquestAcceptance() AcceptanceConfig {
	return AcceptanceConfig{
		Name:               "fullgame-minimal-public-executor",
		Seed:               seedBase + 3,
		TurnMode:           "sequential",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "temperate", Wetness: 50},
		Rules:              mustRulesFixture("bloodlust_conquest", "scripts/military-zoc-combat.json"),
		HumanPolicy:        ConquestHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 2},
		Restart:            AcceptanceRestartPlan{Point: AcceptanceRestartAfterDurableRound},
		ExpectedWinner:     "account-1",
		ExpectedReason:     "bloodlust_conquest",
		ExpectedStateHash:  "b2d92f8e4421a4a7cea573479e9395657eb0db5bd1415608f37bc7e7f994352a",
		ExpectedEventHash:  "104c65dc8ea937e023b0b6ba8918d5ef75a00f769c9264df09f9227ad4d3b169",
	}
}

func RowTwoDiplomacyAcceptance() AcceptanceConfig {
	return AcceptanceConfig{
		Name:               "fullgame-02-long_turn-diplomatic_world_union",
		Seed:               seedBase + 2,
		TurnMode:           "long_turn",
		Slots:              slotPlan([]string{humanController, aiController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "small", Climate: "cold", Wetness: 50},
		Rules:              mustRulesFixture("diplomatic_world_union", "scripts/city-economy-wonder.json"),
		HumanPolicy:        DiplomacyHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 2},
		Restart:            AcceptanceRestartPlan{Point: AcceptanceRestartAfterCompletedTurn, CompletedTurn: 12},
		ExpectedWinner:     "account-1",
		ExpectedReason:     "diplomatic_world_union",
	}
}
