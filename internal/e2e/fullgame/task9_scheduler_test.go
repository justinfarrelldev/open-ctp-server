package fullgame

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/onboarding"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/puppet"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestPublicExecutorContinuitySurvivorRepetitionCompletesSequentialMixedControllerCampaign(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x903,
		TurnMode:           "sequential",
		Slots:              slotPlan([]string{humanController, humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "temperate", Wetness: 50},
		Rules:              mustRulesFixture("bloodlust_conquest", "scripts/resignation-continuity.json"),
		HumanPolicy:        ContinuityHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 2, Deadline: 2 * time.Minute, StartingSettlers: 1},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	viewingPlayer := fixture.setup.Players[len(fixture.setup.Players)-1]
	leavePlayerID := "account-" + strconv.FormatInt(fixture.setup.Players[0].AccountID, 10)
	survivorPlayerID := "account-" + strconv.FormatInt(fixture.setup.Players[1].AccountID, 10)
	scheduler := mixedControllerScheduler{
		scenario:         config,
		client:           fixture.client,
		setup:            fixture.setup,
		viewer:           viewingPlayer.Actor,
		level:            puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:          &fixture.runtime,
		server:           &fixture.server,
		db:               fixture.db,
		sources:          fixture.sources,
		seed:             fixture.seed,
		transport:        fixture.transport,
		transcript:       fixture.transcript,
		observeAI:        fixture.observeAI,
		histories:        map[string][]puppet.PublicExplorationStep{},
		scripted:         map[string]bool{},
		scriptProgress:   map[string]int{},
		runTakeovers:     true,
		humanScriptModes: map[string]humanScriptMode{leavePlayerID: humanScriptOnly},
	}

	if _, err := runMixedControllerCampaign(fixture.ctx, &scheduler); err != nil {
		t.Fatal(err)
	}
	if !scheduler.scripted[leavePlayerID] {
		t.Fatalf("leave actor %s did not execute the continuity leave script", leavePlayerID)
	}
	if scheduler.scripted[survivorPlayerID] {
		t.Fatalf("surviving human %s unexpectedly replayed the leave script", survivorPlayerID)
	}
	if len(*fixture.transcript) == 0 {
		t.Fatal("mixed-controller continuity scheduler produced no AI transcript")
	}
}

func TestRunAIOnlyGameFailsClosedWhenHumanControllersRemain(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x904,
		TurnMode:           "sequential",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "temperate", Wetness: 50},
		Rules:              mustRulesFixture("bloodlust_conquest", "scripts/military-zoc-combat.json"),
		HumanPolicy:        ConquestHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            defaultAcceptanceBudgets(),
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	_, err := runAIOnlyGame(fixture.ctx, config, fixture.client, fixture.setup, fixture.setup.Owner, fixture.db, fixture.sources, fixture.seed, &fixture.runtime, &fixture.server, fixture.transport, fixture.transcript, fixture.observeAI)
	if err == nil || !strings.Contains(err.Error(), "human controllers remain") {
		t.Fatalf("runAIOnlyGame error = %v, want fail-closed human-controller guard", err)
	}
}

func TestMixedControllerSchedulerFailsFastOnSameTurnCommandBudgetExhaustion(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x906,
		TurnMode:           "simultaneous",
		Slots:              slotPlan([]string{aiController, aiController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "small", Climate: "temperate", Wetness: 10},
		Rules:              mustRulesFixture("science_gaia_controller", ""),
		HumanPolicy:        ObserveOnlyPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 3, Turns: 1, Deadline: 30 * time.Second, StartingSettlers: 1},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	initialTurn, err := decodeProjectionTurn(fixture.transport.InitialProjection)
	if err != nil {
		t.Fatal(err)
	}
	scheduler := mixedControllerScheduler{
		scenario:       config,
		client:         fixture.client,
		setup:          fixture.setup,
		viewer:         fixture.setup.Owner,
		level:          puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:        &fixture.runtime,
		server:         &fixture.server,
		db:             fixture.db,
		sources:        fixture.sources,
		seed:           fixture.seed,
		transport:      fixture.transport,
		transcript:     fixture.transcript,
		observeAI:      fixture.observeAI,
		histories:      map[string][]puppet.PublicExplorationStep{},
		scripted:       map[string]bool{},
		scriptProgress: map[string]int{},
	}
	_, err = runMixedControllerCampaign(fixture.ctx, &scheduler)
	if err == nil || !strings.Contains(err.Error(), "mixed-controller command budget exhausted") {
		t.Fatalf("same-turn command budget error = %v", err)
	}
	if strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("same-turn command budget error leaked deadline timeout: %v", err)
	}
	currentTurn, turnErr := decodeProjectionTurn(fixture.transport.InitialProjection)
	if turnErr != nil {
		t.Fatal(turnErr)
	}
	if currentTurn != initialTurn {
		t.Fatalf("same-turn budget test advanced turn from %d to %d", initialTurn, currentTurn)
	}
}

func TestScriptedHumanStagesDependentCityScriptAcrossHybridPlanningBoundaries(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x908,
		TurnMode:           "hybrid",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "small", Climate: "tropical", Wetness: 90},
		Rules:              mustRulesFixture("science_gaia_controller", "scripts/city-economy-wonder.json"),
		HumanPolicy:        ScriptedHumanPolicy{Level: puppet.TestHarnessPolicyLevelProductionFirst},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 40, Turns: 2, Deadline: 2 * time.Minute, StartingSettlers: 1},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	human := fixture.setup.Players[0]
	humanID := "account-" + strconv.FormatInt(human.AccountID, 10)
	histories := map[string][]puppet.PublicExplorationStep{}
	scripted := map[string]bool{}
	scriptProgress := map[string]int{}
	accepted, err := driveScriptedHuman(fixture.ctx, fixture.client, config, fixture.setup.GameID, human, humanID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptOnly, histories, scripted, scriptProgress)
	if err != nil {
		t.Fatal(err)
	}
	if accepted != 1 {
		t.Fatalf("dependent hybrid city script accepted %d commands in one planning phase, want only the first eligible command before resolution", accepted)
	}
	if got := scriptProgress[humanID]; got != 1 {
		t.Fatalf("script progress after first hybrid planning command = %d, want 1", got)
	}
	accepted, err = driveScriptedHuman(fixture.ctx, fixture.client, config, fixture.setup.GameID, human, humanID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptThenPuppet, histories, scripted, scriptProgress)
	if err != nil {
		t.Fatal(err)
	}
	if accepted != 1 {
		t.Fatalf("queued hybrid closeout accepted %d commands, want round-complete only", accepted)
	}
	if got := scriptProgress[humanID]; got != 1 {
		t.Fatalf("script progress advanced to %d while planning status was queued, want it held at 1 until resolution", got)
	}
}

func TestScriptedHumanFailedSubmissionDoesNotConsumeProgress(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x909,
		TurnMode:           "hybrid",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "small", Climate: "tropical", Wetness: 90},
		Rules:              mustRulesFixture("science_gaia_controller", "scripts/city-economy-wonder.json"),
		HumanPolicy:        ScriptedHumanPolicy{Level: puppet.TestHarnessPolicyLevelProductionFirst},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 40, Turns: 2, Deadline: 2 * time.Minute, StartingSettlers: 1},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	human := fixture.setup.Players[0]
	humanID := "account-" + strconv.FormatInt(human.AccountID, 10)
	histories := map[string][]puppet.PublicExplorationStep{}
	scripted := map[string]bool{}
	scriptProgress := map[string]int{}
	brokenClient := fixture.client
	brokenClient.BaseURL = "http://127.0.0.1:1"
	accepted, err := driveScriptedHuman(fixture.ctx, brokenClient, config, fixture.setup.GameID, human, humanID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptOnly, histories, scripted, scriptProgress)
	if err == nil {
		t.Fatal("expected scripted-human submission failure")
	}
	if accepted != 0 {
		t.Fatalf("failed scripted-human submission accepted %d commands, want 0", accepted)
	}
	if got := scriptProgress[humanID]; got != 0 {
		t.Fatalf("script progress after failed submission = %d, want 0", got)
	}
	accepted, err = driveScriptedHuman(fixture.ctx, fixture.client, config, fixture.setup.GameID, human, humanID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptOnly, histories, scripted, scriptProgress)
	if err != nil {
		t.Fatal(err)
	}
	if accepted != 1 {
		t.Fatalf("retried scripted-human submission accepted %d commands, want 1", accepted)
	}
	if got := scriptProgress[humanID]; got != 1 {
		t.Fatalf("script progress after successful retry = %d, want 1", got)
	}
}

func TestProductionAIDriveClosesQueuedPlanningInSimultaneousObserveOnly(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x907,
		TurnMode:           "simultaneous",
		Slots:              slotPlan([]string{aiController, aiController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "small", Climate: "temperate", Wetness: 10},
		Rules:              mustRulesFixture("science_gaia_controller", ""),
		HumanPolicy:        ObserveOnlyPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 6, Turns: 1, Deadline: 30 * time.Second, StartingSettlers: 1},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	scheduler := mixedControllerScheduler{
		scenario:   config,
		client:     fixture.client,
		setup:      fixture.setup,
		viewer:     fixture.setup.Owner,
		level:      puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:    &fixture.runtime,
		server:     &fixture.server,
		db:         fixture.db,
		sources:    fixture.sources,
		seed:       fixture.seed,
		transport:  fixture.transport,
		transcript: fixture.transcript,
		observeAI:  fixture.observeAI,
		histories:  map[string][]puppet.PublicExplorationStep{},
		scripted:   map[string]bool{},
	}
	if _, err := runMixedControllerCampaign(fixture.ctx, &scheduler); err != nil {
		t.Fatalf("simultaneous planning closeout error = %v", err)
	}
	transcript := *fixture.transcript
	if len(transcript) < 4 {
		t.Fatalf("simultaneous planning closeout transcript len = %d, want at least 4", len(transcript))
	}
	actedIndex := -1
	for index, entry := range transcript {
		if entry.Type == engine.CommandFoundCity {
			actedIndex = index
			break
		}
	}
	if actedIndex < 0 {
		t.Fatalf("expected at least one tactical AI action before closeout; transcript=%+v", transcript)
	}
	acted := transcript[actedIndex]
	closeoutIndex := -1
	for index := actedIndex + 1; index < len(transcript); index++ {
		entry := transcript[index]
		if entry.ActorPlayerID != acted.ActorPlayerID {
			continue
		}
		if entry.Type == engine.CommandFoundCity && string(entry.Payload) == string(acted.Payload) {
			t.Fatalf("actor %s repeated found-city instead of closing planning: acted=%+v repeated=%+v", acted.ActorPlayerID, acted, entry)
		}
		if entry.Type == engine.CommandRoundComplete {
			closeoutIndex = index
			break
		}
	}
	if closeoutIndex < 0 {
		t.Fatalf("actor %s never closed queued planning after acting; transcript=%+v", acted.ActorPlayerID, transcript)
	}
}

func TestPublicExecutorModeSpecificCloseoutCompletesObserveOnlyPlanningModes(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			config := AcceptanceConfig{
				Name:               t.Name(),
				Seed:               seedBase + 0x905,
				TurnMode:           mode,
				Slots:              slotPlan([]string{aiController, aiController, aiController}),
				World:              AcceptanceWorldFixture{MapSize: "small", Climate: "temperate", Wetness: 10},
				Rules:              mustRulesFixture("science_gaia_controller", ""),
				HumanPolicy:        ObserveOnlyPolicy{},
				AutonomousAIPolicy: ProductionAutonomousAI,
				Budgets:            AcceptanceBudgets{Commands: 12, Turns: 1, Deadline: 30 * time.Second, StartingSettlers: 1},
				Restart:            disabledRestart(),
			}
			fixture := newTask9MixedControllerFixture(t, config)
			defer fixture.Close()

			initialTurn, err := decodeProjectionTurn(fixture.transport.InitialProjection)
			if err != nil {
				t.Fatal(err)
			}
			scheduler := mixedControllerScheduler{
				scenario:       config,
				client:         fixture.client,
				setup:          fixture.setup,
				viewer:         fixture.setup.Owner,
				level:          puppet.TestHarnessPolicyLevelProductionFirst,
				runtime:        &fixture.runtime,
				server:         &fixture.server,
				db:             fixture.db,
				sources:        fixture.sources,
				seed:           fixture.seed,
				transport:      fixture.transport,
				transcript:     fixture.transcript,
				observeAI:      fixture.observeAI,
				histories:      map[string][]puppet.PublicExplorationStep{},
				scripted:       map[string]bool{},
				scriptProgress: map[string]int{},
			}
			if _, err := runMixedControllerCampaign(fixture.ctx, &scheduler); err != nil {
				t.Fatal(err)
			}
			currentTurn, err := decodeProjectionTurn(fixture.transport.InitialProjection)
			if err != nil {
				t.Fatal(err)
			}
			if currentTurn <= initialTurn {
				t.Fatalf("mode %s projection turn = %d, want advancement beyond %d", mode, currentTurn, initialTurn)
			}
			if len(*fixture.transcript) == 0 {
				t.Fatalf("mode %s produced no AI transcript", mode)
			}
		})
	}
}

func TestRowSevenScienceScriptReachesCompletedTurnRestartCheckpointBeforeTerminal(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x90a,
		TurnMode:           "simultaneous",
		Slots:              slotPlan([]string{humanController, aiController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "small", Climate: "temperate", Wetness: 10},
		Rules:              mustRulesFixture("science_gaia_controller", "scripts/row07-delayed-city-economy-wonder.json"),
		HumanPolicy:        ScriptedHumanPolicy{Level: puppet.TestHarnessPolicyLevelProductionFirst},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 1},
		Restart:            completedTurnRestart(12),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	restartEvents := make([]AcceptanceRestartEvent, 0, 1)
	config.ObserveRestart = func(event AcceptanceRestartEvent) {
		restartEvents = append(restartEvents, event)
	}
	scheduler := mixedControllerScheduler{
		scenario:       config,
		client:         fixture.client,
		setup:          fixture.setup,
		viewer:         fixture.setup.Owner,
		level:          puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:        &fixture.runtime,
		server:         &fixture.server,
		db:             fixture.db,
		sources:        fixture.sources,
		seed:           fixture.seed,
		transport:      fixture.transport,
		transcript:     fixture.transcript,
		observeAI:      fixture.observeAI,
		histories:      map[string][]puppet.PublicExplorationStep{},
		scripted:       map[string]bool{},
		scriptProgress: map[string]int{},
		humanScriptModes: map[string]humanScriptMode{
			"account-1": humanScriptThenPuppet,
		},
	}
	if _, err := runMixedControllerCampaign(fixture.ctx, &scheduler); err != nil {
		t.Fatal(err)
	}
	projection, err := fixture.transport.CurrentProjection(fixture.ctx)
	if err != nil {
		t.Fatalf("refresh terminal projection after scheduler exit: %v", err)
	}
	currentTurn, err := decodeProjectionTurn(projection)
	if err != nil {
		t.Fatal(err)
	}
	want := AcceptanceRestartEvent{Point: AcceptanceRestartAfterCompletedTurn, CompletedTurn: 12}
	if currentTurn < 12 || len(restartEvents) != 1 || restartEvents[0] != want {
		t.Fatalf("immediate science script terminalized before completed-turn restart accounting: turn=%d restart_events=%+v want=%+v", currentTurn, restartEvents, want)
	}
}

type task9MixedControllerFixture struct {
	ctx        context.Context
	kit        *testkit.Scenario
	db         *sqlx.DB
	telemetry  *memorysql.Telemetry
	runtime    *onboarding.PublicRuntime
	server     *httptest.Server
	client     onboarding.HTTPClient
	transport  *HTTPExplorationTransport
	setup      publicGameSetup
	sources    engine.Sources
	seed       func() (int64, error)
	transcript *[]aiTranscriptEntry
	observeAI  func(ai.Submission)
}

func TestMixedControllerSchedulerReusesRestartedPublicBaseURLForContinuityTakeover(t *testing.T) {
	config := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x904,
		TurnMode:           "simultaneous",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "temperate", Wetness: 50},
		Rules:              mustRulesFixture("score_2300", "scripts/resignation-continuity.json"),
		HumanPolicy:        ContinuityHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 40, Turns: 2, Deadline: 2 * time.Minute, StartingSettlers: 1},
		Restart:            completedTurnRestart(1),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	viewingPlayer := fixture.setup.Players[len(fixture.setup.Players)-1]
	leavePlayerID := "account-" + strconv.FormatInt(fixture.setup.Players[0].AccountID, 10)
	scheduler := mixedControllerScheduler{
		scenario:         config,
		client:           fixture.client,
		setup:            fixture.setup,
		viewer:           viewingPlayer.Actor,
		level:            puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:          &fixture.runtime,
		server:           &fixture.server,
		db:               fixture.db,
		sources:          fixture.sources,
		seed:             fixture.seed,
		transport:        fixture.transport,
		transcript:       fixture.transcript,
		observeAI:        fixture.observeAI,
		histories:        map[string][]puppet.PublicExplorationStep{},
		scripted:         map[string]bool{},
		scriptProgress:   map[string]int{},
		runTakeovers:     true,
		humanScriptModes: map[string]humanScriptMode{leavePlayerID: humanScriptOnly},
	}

	if _, err := runMixedControllerCampaign(fixture.ctx, &scheduler); err != nil {
		t.Fatal(err)
	}
	if !scheduler.scripted[leavePlayerID] {
		t.Fatalf("leave actor %s did not execute the continuity leave script", leavePlayerID)
	}
	if _, _, err := scheduler.currentControllers(fixture.ctx); err != nil {
		t.Fatalf("post-restart controller audit reused stale public base URL: %v", err)
	}
}

func TestExecuteContinuityAcceptanceCompletesLongTurnThreeCityContinuityOpener(t *testing.T) {
	// The old catalog seed depended on adjacent founding. With legal spacing
	// it reaches a score tie, not a terminal within this bounded test. Keep a
	// deterministic non-tied fixture without weakening terminal assertions.
	config := AcceptanceConfig{
		Name:               "fullgame-06-long_turn-score_2300",
		Seed:               seedBase + 0x906,
		TurnMode:           "long_turn",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "cold", Wetness: 90},
		Rules:              mustRulesFixture("score_2300", "scripts/score-resignation-continuity.json"),
		HumanPolicy:        ContinuityHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 3},
		Restart:            disabledRestart(),
	}
	kit := testkit.NewScenario(t.Name(), int64(config.Seed))
	defer kit.Close()

	artifacts, err := executeContinuityAcceptance(kit.Context, config, kit)
	if err != nil {
		t.Fatalf("full continuity executor opener failed: %v", err)
	}
	if artifacts.TerminalReason != "score_2300" {
		t.Fatalf("terminal reason = %q, want score_2300", artifacts.TerminalReason)
	}
	if artifacts.TerminalWinner == "" || artifacts.StateHash == "" || artifacts.EventHash == "" {
		t.Fatalf("artifacts must be populated: %+v", artifacts)
	}
}

func TestLongTurnThreeCityContinuityCandidateOwnerBootstrapShowsThreeOwnedUnits(t *testing.T) {
	config := AcceptanceConfig{
		Name:               "fullgame-06-long_turn-score_2300",
		Seed:               seedBase + 6,
		TurnMode:           "long_turn",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "cold", Wetness: 90},
		Rules:              mustRulesFixture("score_2300", "scripts/score-resignation-continuity.json"),
		HumanPolicy:        ContinuityHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 3},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	transport := &HTTPExplorationTransport{BaseURL: fixture.server.URL, Client: fixture.client.Client, GameID: fixture.setup.GameID, SessionID: fixture.setup.Owner.Session, Namespace: t.Name()}
	_, projection, err := transport.Bootstrap(fixture.ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	commands, preparing, err := resolveContinuityPreparationCommands(config, projection, "account-"+strconv.FormatInt(fixture.setup.Players[0].AccountID, 10))
	if err != nil {
		t.Fatal(err)
	}
	if !preparing || len(commands) != 2 {
		t.Fatalf("preparing=%v commands=%d, want two continuity spread moves", preparing, len(commands))
	}
}

func TestLongTurnThreeCityContinuityCandidateOwnerProjectionSurvivesObserverBootstrapAndWorker(t *testing.T) {
	config := AcceptanceConfig{
		Name:               "fullgame-06-long_turn-score_2300",
		Seed:               seedBase + 6,
		TurnMode:           "long_turn",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "cold", Wetness: 90},
		Rules:              mustRulesFixture("score_2300", "scripts/score-resignation-continuity.json"),
		HumanPolicy:        ContinuityHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 3},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	ownerTransport := &HTTPExplorationTransport{BaseURL: fixture.server.URL, Client: fixture.client.Client, GameID: fixture.setup.GameID, SessionID: fixture.setup.Owner.Session, Namespace: t.Name() + "-owner"}
	if _, err := ownerTransport.CurrentProjection(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	viewer := fixture.setup.Players[len(fixture.setup.Players)-1]
	viewerTransport := &HTTPExplorationTransport{BaseURL: fixture.server.URL, Client: fixture.client.Client, GameID: fixture.setup.GameID, SessionID: viewer.Actor.Session, Namespace: t.Name() + "-viewer"}
	if _, err := viewerTransport.CurrentProjection(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	if err := fixture.runtime.RunPendingTakeovers(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	projection, err := ownerTransport.CurrentProjection(fixture.ctx)
	if err != nil {
		t.Fatal(err)
	}
	commands, preparing, err := resolveContinuityPreparationCommands(config, projection, "account-"+strconv.FormatInt(fixture.setup.Players[0].AccountID, 10))
	if err != nil {
		t.Fatal(err)
	}
	if !preparing || len(commands) != 2 {
		t.Fatalf("preparing=%v commands=%d, want two continuity spread moves after observer bootstrap/worker", preparing, len(commands))
	}
}

func TestLongTurnThreeCityContinuitySchedulerCachedTransportPreservesOpener(t *testing.T) {
	config := AcceptanceConfig{
		Name:               "fullgame-06-long_turn-score_2300",
		Seed:               seedBase + 6,
		TurnMode:           "long_turn",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "cold", Wetness: 90},
		Rules:              mustRulesFixture("score_2300", "scripts/score-resignation-continuity.json"),
		HumanPolicy:        ContinuityHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 3},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	leavePlayerID := "account-" + strconv.FormatInt(fixture.setup.Players[0].AccountID, 10)
	cached := &HTTPExplorationTransport{BaseURL: fixture.server.URL, Client: fixture.client.Client, GameID: fixture.setup.GameID, SessionID: fixture.setup.Owner.Session, Namespace: t.Name() + "-owner"}
	if _, err := cached.CurrentProjection(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	viewer := fixture.setup.Players[len(fixture.setup.Players)-1]
	scheduler := mixedControllerScheduler{
		scenario:                config,
		client:                  fixture.client,
		setup:                   fixture.setup,
		viewer:                  viewer.Actor,
		level:                   puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:                 &fixture.runtime,
		server:                  &fixture.server,
		db:                      fixture.db,
		sources:                 fixture.sources,
		seed:                    fixture.seed,
		transport:               fixture.transport,
		transcript:              fixture.transcript,
		observeAI:               fixture.observeAI,
		histories:               map[string][]puppet.PublicExplorationStep{},
		longTurnContinuityState: map[string]*HTTPExplorationTransport{leavePlayerID: cached},
		scripted:                map[string]bool{},
		scriptProgress:          map[string]int{},
		runTakeovers:            true,
		humanScriptModes:        map[string]humanScriptMode{leavePlayerID: humanScriptOnly},
	}
	if err := fixture.runtime.RunPendingTakeovers(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	humans, _, err := scheduler.currentControllers(fixture.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(humans) != 1 {
		t.Fatalf("human controllers = %d, want 1", len(humans))
	}
	accepted, err := driveScriptedHuman(fixture.ctx, fixture.client, config, fixture.setup.GameID, humans[0], leavePlayerID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, scheduler.longTurnContinuityTransport(leavePlayerID), humanScriptOnly, scheduler.histories, scheduler.scripted, scheduler.scriptProgress)
	if err != nil {
		t.Fatal(err)
	}
	if accepted < 1 {
		t.Fatalf("accepted=%d, want opener progress", accepted)
	}
}

func TestLongTurnThreeCityContinuitySchedulerCachedTransportSurvivesSecondCycle(t *testing.T) {
	config := AcceptanceConfig{
		Name:               "fullgame-06-long_turn-score_2300",
		Seed:               seedBase + 6,
		TurnMode:           "long_turn",
		Slots:              slotPlan([]string{humanController, aiController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "cold", Wetness: 90},
		Rules:              mustRulesFixture("score_2300", "scripts/score-resignation-continuity.json"),
		HumanPolicy:        ContinuityHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            AcceptanceBudgets{Commands: 240, Turns: 48, Deadline: 2 * time.Minute, StartingSettlers: 3},
		Restart:            disabledRestart(),
	}
	fixture := newTask9MixedControllerFixture(t, config)
	defer fixture.Close()

	leavePlayerID := "account-" + strconv.FormatInt(fixture.setup.Players[0].AccountID, 10)
	cached := &HTTPExplorationTransport{BaseURL: fixture.server.URL, Client: fixture.client.Client, GameID: fixture.setup.GameID, SessionID: fixture.setup.Owner.Session, Namespace: t.Name() + "-owner"}
	if _, err := cached.CurrentProjection(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	viewer := fixture.setup.Players[len(fixture.setup.Players)-1]
	scheduler := mixedControllerScheduler{
		scenario:                config,
		client:                  fixture.client,
		setup:                   fixture.setup,
		viewer:                  viewer.Actor,
		level:                   puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:                 &fixture.runtime,
		server:                  &fixture.server,
		db:                      fixture.db,
		sources:                 fixture.sources,
		seed:                    fixture.seed,
		transport:               fixture.transport,
		transcript:              fixture.transcript,
		observeAI:               fixture.observeAI,
		histories:               map[string][]puppet.PublicExplorationStep{},
		longTurnContinuityState: map[string]*HTTPExplorationTransport{leavePlayerID: cached},
		scripted:                map[string]bool{},
		scriptProgress:          map[string]int{},
		runTakeovers:            true,
		humanScriptModes:        map[string]humanScriptMode{leavePlayerID: humanScriptOnly},
	}
	if err := fixture.runtime.RunPendingTakeovers(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	humans, _, err := scheduler.currentControllers(fixture.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(humans) != 1 {
		t.Fatalf("human controllers = %d, want 1", len(humans))
	}
	accepted, err := driveScriptedHuman(fixture.ctx, fixture.client, config, fixture.setup.GameID, humans[0], leavePlayerID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, scheduler.longTurnContinuityTransport(leavePlayerID), humanScriptOnly, scheduler.histories, scheduler.scripted, scheduler.scriptProgress)
	if err != nil {
		t.Fatal(err)
	}
	if accepted < 1 {
		t.Fatalf("first cycle accepted=%d, want opener progress", accepted)
	}
	accepted, err = driveScriptedHuman(fixture.ctx, fixture.client, config, fixture.setup.GameID, humans[0], leavePlayerID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, scheduler.longTurnContinuityTransport(leavePlayerID), humanScriptOnly, scheduler.histories, scheduler.scripted, scheduler.scriptProgress)
	if err != nil {
		t.Fatal(err)
	}
	if accepted < 1 {
		t.Fatalf("second cycle accepted=%d, want found-city/leave progress", accepted)
	}
}

func newTask9MixedControllerFixture(t *testing.T, config AcceptanceConfig) *task9MixedControllerFixture {
	t.Helper()
	kit := testkit.NewScenario(t.Name(), int64(config.Seed))
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		kit.Close()
		t.Fatal(err)
	}
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(config.Name), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(config.Seed), nil }
	transcript := make([]aiTranscriptEntry, 0)
	observeAI := appendAITranscript(&transcript)
	runtime := onboarding.NewPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
	server := httptest.NewServer(runtime.Handler)
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	setup, err := setupCampaignGame(kit.Context, client, config)
	if err != nil {
		server.Close()
		_ = db.Close()
		database.Close()
		kit.Close()
		t.Fatal(err)
	}
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: setup.GameID, SessionID: setup.Owner.Session, Namespace: config.Name}
	if _, err := transport.CurrentProjection(kit.Context); err != nil {
		server.Close()
		_ = db.Close()
		database.Close()
		kit.Close()
		t.Fatalf("bootstrap fixture projection: %v", err)
	}
	return &task9MixedControllerFixture{
		ctx:        kit.Context,
		kit:        kit,
		db:         db,
		telemetry:  telemetry,
		runtime:    runtime,
		server:     server,
		client:     client,
		transport:  transport,
		setup:      setup,
		sources:    sources,
		seed:       seed,
		transcript: &transcript,
		observeAI:  observeAI,
	}
}

func (f *task9MixedControllerFixture) Close() {
	if f == nil {
		return
	}
	if f.server != nil {
		f.server.Close()
	}
	if f.db != nil {
		_ = f.db.Close()
	}
	if f.kit != nil {
		memorysql.Reset(f.kit.Namespace)
		f.kit.Close()
	}
}
