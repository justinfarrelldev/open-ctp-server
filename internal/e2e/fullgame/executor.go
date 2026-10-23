package fullgame

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/onboarding"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/puppet"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
	"github.com/justinfarrelldev/open-ctp-server/internal/game"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

// PublicExecutor runs a complete game only through production HTTP, SSE, and AI gateways.
type PublicExecutor struct{}

type executorRNG struct{}

type aiTranscriptEntry struct {
	GameID         int64           `json:"game_id"`
	ActorPlayerID  string          `json:"actor_player_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	CommandID      string          `json:"command_id"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
}

func appendAITranscript(transcript *[]aiTranscriptEntry) func(ai.Submission) {
	return func(submission ai.Submission) {
		*transcript = append(*transcript, aiTranscriptEntry{
			GameID:         submission.GameID,
			ActorPlayerID:  submission.ActorPlayerID,
			IdempotencyKey: submission.IdempotencyKey,
			CommandID:      submission.CommandID,
			Type:           submission.Type,
			Payload:        append(json.RawMessage(nil), submission.Payload...),
		})
	}
}

func validateAITranscript(transcript []aiTranscriptEntry, gameID int64, players []string, commandCount int) error {
	allowed := make(map[string]struct{}, len(players))
	for _, player := range players {
		allowed[player] = struct{}{}
	}
	if commandCount < 1 || len(transcript) != commandCount {
		return fmt.Errorf("production AI gateway audit has %d entries for %d counted commands", len(transcript), commandCount)
	}
	for index, entry := range transcript {
		if entry.GameID != gameID {
			return fmt.Errorf("production AI gateway audit entry %d has game %d, want %d", index, entry.GameID, gameID)
		}
		if _, ok := allowed[entry.ActorPlayerID]; !ok {
			return fmt.Errorf("production AI gateway audit entry %d has unexpected actor %q", index, entry.ActorPlayerID)
		}
		if entry.IdempotencyKey == "" || entry.CommandID == "" || entry.Type == "" || !json.Valid(entry.Payload) {
			return fmt.Errorf("production AI gateway audit entry %d is incomplete", index)
		}
	}
	return nil
}

func recentAITranscript(transcript []aiTranscriptEntry, limit int) string {
	if len(transcript) == 0 {
		return "[]"
	}
	if limit < 1 || limit > len(transcript) {
		limit = len(transcript)
	}
	recent, err := json.Marshal(transcript[len(transcript)-limit:])
	if err != nil {
		return "[\"encode-error\"]"
	}
	return string(recent)
}

func commandBudgetError(mode string, turn int, revision int64, actor string, accepted, budget int, transcript []aiTranscriptEntry) error {
	if actor == "" {
		actor = "unknown"
	}
	return fmt.Errorf("mixed-controller command budget exhausted in mode %s at turn %d revision %d after %d accepted commands (budget %d); actor=%s; recent AI transcript=%s", mode, turn, revision, accepted, budget, actor, recentAITranscript(transcript, 8))
}

func validateAIObservation(commandCount, observed int) error {
	if commandCount < 0 {
		return errors.New("production AI reported a negative command count")
	}
	if observed != commandCount {
		return fmt.Errorf("production AI submitted %d commands but observer recorded %d", commandCount, observed)
	}
	return nil
}

func (executorRNG) Next(_ string, state uint64) (uint64, uint64) {
	next := state + 0x9e3779b97f4a7c15
	value := next
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return next, value ^ (value >> 31)
}

func (PublicExecutor) Execute(ctx context.Context, config AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	if kit == nil {
		return Artifacts{}, errors.New("public executor requires a scenario kit")
	}
	if err := config.validate(); err != nil {
		return Artifacts{}, fmt.Errorf("invalid acceptance config: %w", err)
	}
	memorysql.Reset(kit.Namespace)
	defer memorysql.Reset(kit.Namespace)
	runCtx, cancel := context.WithTimeout(ctx, config.Budgets.Deadline)
	defer cancel()
	artifacts, err := config.HumanPolicy.execute(runCtx, config, kit)
	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return Artifacts{}, fmt.Errorf("acceptance time budget exhausted after %s: %w", config.Budgets.Deadline, err)
		}
		return Artifacts{}, err
	}
	if err := runCtx.Err(); err != nil {
		return Artifacts{}, fmt.Errorf("acceptance time budget exhausted after %s: %w", config.Budgets.Deadline, err)
	}
	return artifacts, nil
}

func executeConquestAcceptance(ctx context.Context, scenario AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		return Artifacts{}, err
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(scenario.Name), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(scenario.Seed), nil }
	transcript := make([]aiTranscriptEntry, 0)
	observeAI := appendAITranscript(&transcript)
	runtime := onboarding.NewPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
	server := httptest.NewServer(runtime.Handler)
	defer func() { server.Close() }()
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}

	owner, gameID, humanPlayer, aiPlayer, err := setupPublicGame(ctx, client, scenario)
	if err != nil {
		return Artifacts{}, err
	}
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: gameID, SessionID: owner.Session, Namespace: scenario.Name}
	_, projection, err := transport.Bootstrap(ctx, 0)
	if err != nil {
		return Artifacts{}, fmt.Errorf("bootstrap public projection: %w", err)
	}
	unitID, armyID, err := resolveConquestRoles(scenario.Rules.HumanScript, projection, humanPlayer)
	if err != nil {
		return Artifacts{}, err
	}
	aiCommands := 0
	driveAI := func() error {
		expectedRevision := transport.Revision
		for cycle := 0; cycle < scenario.Budgets.Commands; cycle++ {
			transcriptStart := len(transcript)
			result, err := runtime.Driver.Drive(ctx, gameID, aiPlayer, expectedRevision)
			terminal := errors.Is(err, ai.ErrTerminal)
			if terminal {
				transport.Terminal = true
			}
			if err != nil && !terminal {
				return fmt.Errorf("production AI drive: %w", err)
			}
			observed := len(transcript) - transcriptStart
			if err := validateAIObservation(result.Commands, observed); err != nil {
				return err
			}
			aiCommands += result.Commands
			expectedRevision = result.Revision
			if result.Commands == 0 {
				break
			}
			if terminal || transcript[len(transcript)-1].Type == engine.CommandPass {
				break
			}
			if cycle+1 == scenario.Budgets.Commands {
				transcriptJSON, _ := json.Marshal(transcript)
				return fmt.Errorf("production AI command budget exhausted after %d cycles; AI transcript=%s", scenario.Budgets.Commands, transcriptJSON)
			}
		}
		_, projection, err := transport.Bootstrap(ctx, transport.Cursor)
		if err != nil {
			return fmt.Errorf("refresh SSE projection after AI: %w", err)
		}
		if transport.Revision != expectedRevision {
			return fmt.Errorf("AI drive returned revision %d but SSE projection is %d", expectedRevision, transport.Revision)
		}
		transport.InitialProjection = projection
		return nil
	}
	restart := func() error {
		if scenario.Restart.Point != AcceptanceRestartAfterDurableRound {
			return driveAI()
		}
		server.Close()
		runtime = onboarding.RestartPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
		server = httptest.NewServer(runtime.Handler)
		transport.BaseURL = server.URL
		if scenario.ObserveRestart != nil {
			scenario.ObserveRestart(AcceptanceRestartEvent{Point: scenario.Restart.Point})
		}
		return driveAI()
	}
	runner := PublicConquestRunner{
		Transport:    transport,
		Database:     db,
		ActionBudget: scenario.Budgets.Commands,
		TurnBudget:   scenario.Budgets.Turns,
		Advance: func(ctx context.Context) error {
			if _, err := transport.Command(ctx, "pass", map[string]json.RawMessage{}); err != nil {
				return fmt.Errorf("human pass: %w", err)
			}
			if transport.Terminal {
				return nil
			}
			return driveAI()
		},
		Restart: restart,
		Resume: func(ctx context.Context, cursor int64) (int64, []byte, error) {
			return resumePublicEvents(ctx, transport, cursor)
		},
		Terminal: func(ctx context.Context) (int64, string, []byte, error) {
			return terminalPublicEvent(ctx, transport)
		},
	}
	result, err := runner.Execute(ctx, unitID, armyID)
	if err != nil {
		return Artifacts{}, err
	}
	if telemetry.Unknown != nil {
		return Artifacts{}, fmt.Errorf("memorysql rejected production query: %w", telemetry.Unknown)
	}
	if err := validateAITranscript(transcript, gameID, []string{aiPlayer}, aiCommands); err != nil {
		return Artifacts{}, err
	}
	transcriptJSON, err := json.Marshal(transcript)
	if err != nil {
		return Artifacts{}, fmt.Errorf("encode AI transcript: %w", err)
	}
	return Artifacts{TerminalWinner: result.Audit.TerminalWinner, TerminalReason: result.Audit.TerminalReason, StateHash: result.Audit.StateHash, EventHash: result.Audit.EventHash, AICommands: aiCommands, AITranscript: string(transcriptJSON)}, nil
}

func executeDiplomaticAcceptance(ctx context.Context, scenario AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		return Artifacts{}, err
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(scenario.Name), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(scenario.Seed), nil }
	transcript := make([]aiTranscriptEntry, 0)
	observeAI := appendAITranscript(&transcript)
	runtime := onboarding.NewPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
	server := httptest.NewServer(runtime.Handler)
	defer func() { server.Close() }()
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}

	owner, gameID, humanPlayer, aiPlayers, err := setupDiplomaticGame(ctx, client, scenario)
	if err != nil {
		return Artifacts{}, err
	}
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: gameID, SessionID: owner.Session, Namespace: scenario.Name}
	scheduler := newHumanDiplomacyScheduler(scenario.Budgets.Commands)
	if projection, err := transport.CurrentProjection(ctx); err != nil {
		return Artifacts{}, fmt.Errorf("bootstrap public projection: %w", err)
	} else {
		raw, err := os.ReadFile(scenario.Rules.HumanScript)
		if err != nil {
			return Artifacts{}, err
		}
		var script Script
		if err := json.Unmarshal(raw, &script); err != nil {
			return Artifacts{}, err
		}
		commands, err := ResolveScript(script, projection, map[string]string{"human-1": humanPlayer})
		if err != nil {
			return Artifacts{}, fmt.Errorf("resolve public city script: %w", err)
		}
		for _, command := range commands {
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(command.Payload, &payload); err != nil {
				return Artifacts{}, err
			}
			if _, err := transport.Command(ctx, command.Command, payload); err != nil {
				return Artifacts{}, fmt.Errorf("public city script %s: %w", command.Command, err)
			}
			scheduler.recordSuccess(diplomaticAction{})
		}
	}

	aiCommands := 0
	driveAI := func(player string) error {
		for cycle := 0; cycle < scenario.Budgets.Commands; cycle++ {
			transcriptStart := len(transcript)
			result, err := runtime.Driver.Drive(ctx, gameID, player, transport.Revision)
			if errors.Is(err, ai.ErrTerminal) {
				observed := len(transcript) - transcriptStart
				if observationErr := validateAIObservation(result.Commands, observed); observationErr != nil {
					return observationErr
				}
				aiCommands += result.Commands
				return err
			}
			if err != nil {
				return fmt.Errorf("production AI drive for %s: %w", player, err)
			}
			observed := len(transcript) - transcriptStart
			if err := validateAIObservation(result.Commands, observed); err != nil {
				return err
			}
			aiCommands += result.Commands
			if _, err := transport.CurrentProjection(ctx); err != nil {
				return fmt.Errorf("refresh public projection after production AI drive for %s: %w", player, err)
			}
			if transport.Revision != result.Revision {
				return fmt.Errorf("production AI drive for %s returned revision %d but current public projection is %d", player, result.Revision, transport.Revision)
			}
			if result.Commands == 0 {
				return nil
			}
			if transcript[len(transcript)-1].Type == engine.CommandPass {
				return nil
			}
		}
		transcriptJSON, _ := json.Marshal(transcript)
		return fmt.Errorf("production AI command budget exhausted after %d cycles; AI transcript=%s", scenario.Budgets.Commands, transcriptJSON)
	}
	terminalReached := false
	for turn := 0; turn < scenario.Budgets.Turns; turn++ {
		if err := advanceHumanDiplomacy(ctx, transport, humanPlayer, scheduler); err != nil {
			return Artifacts{}, fmt.Errorf("human diplomatic action: %w", err)
		}
		if _, err := transport.Command(ctx, "pass", map[string]json.RawMessage{}); err != nil {
			return Artifacts{}, fmt.Errorf("human pass to AI: %w", err)
		}
		scheduler.recordSuccess(diplomaticAction{})
		for _, player := range aiPlayers {
			if err := driveAI(player); err != nil {
				if errors.Is(err, ai.ErrTerminal) {
					terminalReached = true
					break
				}
				return Artifacts{}, err
			}
		}
		if terminalReached {
			break
		}
		if scenario.Restart.Point == AcceptanceRestartAfterCompletedTurn && turn+1 == scenario.Restart.CompletedTurn {
			restartCursor := transport.Cursor
			server.Close()
			runtime = onboarding.RestartPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
			server = httptest.NewServer(runtime.Handler)
			transport.BaseURL = server.URL
			if scenario.ObserveRestart != nil {
				scenario.ObserveRestart(AcceptanceRestartEvent{Point: scenario.Restart.Point, CompletedTurn: turn + 1})
			}
			if _, err := transport.CurrentProjection(ctx); err != nil {
				return Artifacts{}, fmt.Errorf("resume current public projection after restart: %w", err)
			}
			if transport.Cursor < restartCursor {
				return Artifacts{}, fmt.Errorf("restart cursor regressed from %d to %d", restartCursor, transport.Cursor)
			}
		}
	}
	terminalCtx, cancelTerminal := context.WithTimeout(ctx, time.Second)
	defer cancelTerminal()
	sequence, name, data, err := terminalPublicEvent(terminalCtx, transport)
	if err != nil {
		safeProjection, projectionErr := transport.CurrentProjection(ctx)
		if projectionErr != nil {
			return Artifacts{}, fmt.Errorf("diplomatic turn budget exhausted; projection unavailable: %w", projectionErr)
		}
		transcriptJSON, transcriptErr := json.Marshal(transcript)
		if transcriptErr != nil {
			return Artifacts{}, transcriptErr
		}
		return Artifacts{}, fmt.Errorf("diplomatic turn budget exhausted after %d turns; AI transcript=%s; safe projection=%s", scenario.Budgets.Turns, strings.ReplaceAll(string(transcriptJSON), "},{", "},\n{"), safeProjection)
	}
	audit, err := AuditTerminal(ctx, db, gameID)
	if err != nil {
		return Artifacts{}, err
	}
	terminal, err := ParseTerminalSSE(sequence, name, data)
	if err != nil {
		return Artifacts{}, err
	}
	if err := validateTerminalAgreement(audit, terminal, scenario.Rules.Victory); err != nil {
		return Artifacts{}, fmt.Errorf("diplomatic terminal: %w", err)
	}
	if telemetry.Unknown != nil {
		return Artifacts{}, fmt.Errorf("memorysql rejected production query: %w", telemetry.Unknown)
	}
	if err := validateAITranscript(transcript, gameID, aiPlayers, aiCommands); err != nil {
		return Artifacts{}, err
	}
	transcriptJSON, err := json.Marshal(transcript)
	if err != nil {
		return Artifacts{}, fmt.Errorf("encode AI transcript: %w", err)
	}
	return Artifacts{TerminalWinner: audit.TerminalWinner, TerminalReason: audit.TerminalReason, StateHash: audit.StateHash, EventHash: audit.EventHash, AICommands: aiCommands, AITranscript: string(transcriptJSON)}, nil
}

func advanceHumanDiplomacy(ctx context.Context, transport *HTTPExplorationTransport, humanPlayer string, scheduler *humanDiplomacyScheduler) error {
	if scheduler == nil || !scheduler.remaining() {
		return errors.New("diplomatic human action budget exhausted")
	}
	projectionBytes, err := transport.CurrentProjection(ctx)
	if err != nil {
		return err
	}
	for scheduler.remaining() {
		var projection visibility.Projection
		if err := json.Unmarshal(projectionBytes, &projection); err != nil {
			return err
		}
		action, ok := scheduler.nextAction(projection, humanPlayer)
		if !ok {
			return nil
		}
		switch action.kind {
		case diplomaticMove:
			if _, err := transport.Move(ctx, action.unitID, action.path); err != nil {
				return err
			}
		case diplomaticEmbassy:
			if _, err := transport.Command(ctx, "establish-embassy", map[string]json.RawMessage{
				"unit_id": json.RawMessage(fmt.Sprintf("%q", action.unitID)),
				"city_id": json.RawMessage(fmt.Sprintf("%q", action.cityID)),
			}); err != nil {
				return err
			}
		case diplomaticRespond:
			if _, err := transport.Command(ctx, "respond-agreement", map[string]json.RawMessage{
				"proposal_id": json.RawMessage(fmt.Sprintf("%q", action.proposalID)),
				"accept":      json.RawMessage(`true`),
			}); err != nil {
				return err
			}
		default:
			return errors.New("diplomatic scheduler selected an unknown action")
		}
		scheduler.recordSuccess(action)
		if projectionBytes, err = transport.CurrentProjection(ctx); err != nil {
			return fmt.Errorf("refresh public projection after %s: %w", action.kind, err)
		}
	}
	return errors.New("diplomatic human action budget exhausted")
}

func executeContinuityAcceptance(ctx context.Context, scenario AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		return Artifacts{}, err
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(scenario.Name), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(scenario.Seed), nil }
	transcript := make([]aiTranscriptEntry, 0)
	observeAI := appendAITranscript(&transcript)
	runtime := onboarding.NewPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
	server := httptest.NewServer(runtime.Handler)
	defer func() { server.Close() }()
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}

	setup, err := setupCampaignGame(ctx, client, scenario)
	if err != nil {
		return Artifacts{}, err
	}
	humanPlayers := publicHumanPlayers(setup)
	if len(humanPlayers) == 0 {
		return Artifacts{}, fmt.Errorf("continuity row %s has no public human actor", scenario.Name)
	}
	leavePlayerID := "account-" + strconv.FormatInt(humanPlayers[0].AccountID, 10)
	longTurnContinuityState := map[string]*HTTPExplorationTransport{}
	if scenario.TurnMode == "long_turn" {
		leaveTransport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: setup.GameID, SessionID: humanPlayers[0].Actor.Session, Namespace: scenario.Name + "-" + leavePlayerID}
		if _, err := leaveTransport.CurrentProjection(ctx); err != nil {
			return Artifacts{}, fmt.Errorf("bootstrap continuity leave projection: %w", err)
		}
		if err := ensureNoHiddenCanary(leaveTransport.InitialProjection); err != nil {
			return Artifacts{}, fmt.Errorf("continuity leave projection: %w", err)
		}
		longTurnContinuityState[leavePlayerID] = leaveTransport
	}
	viewer := setup.Players[len(setup.Players)-1]
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: setup.GameID, SessionID: viewer.Actor.Session, Namespace: scenario.Name}
	if _, err := transport.CurrentProjection(ctx); err != nil {
		return Artifacts{}, fmt.Errorf("bootstrap continuity projection: %w", err)
	}
	if err := ensureNoHiddenCanary(transport.InitialProjection); err != nil {
		return Artifacts{}, err
	}
	humanScriptModes := map[string]humanScriptMode{leavePlayerID: humanScriptOnly}
	scheduler := mixedControllerScheduler{
		scenario:                scenario,
		client:                  client,
		setup:                   setup,
		viewer:                  viewer.Actor,
		level:                   puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:                 &runtime,
		server:                  &server,
		db:                      db,
		sources:                 sources,
		seed:                    seed,
		transport:               transport,
		transcript:              &transcript,
		observeAI:               observeAI,
		histories:               map[string][]puppet.PublicExplorationStep{},
		longTurnContinuityState: longTurnContinuityState,
		scripted:                map[string]bool{},
		scriptProgress:          map[string]int{},
		runTakeovers:            true,
		humanScriptModes:        humanScriptModes,
	}
	aiCommands, err := runMixedControllerCampaign(ctx, &scheduler)
	if err != nil {
		return Artifacts{}, err
	}
	if telemetry.Unknown != nil {
		return Artifacts{}, fmt.Errorf("memorysql rejected production query: %w", telemetry.Unknown)
	}
	_, aiPlayers, err := scheduler.currentControllers(ctx)
	if err != nil {
		return Artifacts{}, err
	}
	if err := validateAITranscript(transcript, setup.GameID, aiPlayers, aiCommands); err != nil {
		return Artifacts{}, err
	}
	transcriptJSON, err := json.Marshal(transcript)
	if err != nil {
		return Artifacts{}, fmt.Errorf("encode AI transcript: %w", err)
	}
	audit, terminal, err := finishObservedTerminal(ctx, db, transport, scenario.Rules.Victory)
	if err != nil {
		return Artifacts{}, err
	}
	return Artifacts{TerminalWinner: terminal.Winner, TerminalReason: terminal.Reason, StateHash: audit.StateHash, EventHash: audit.EventHash, AICommands: aiCommands, AITranscript: string(transcriptJSON)}, nil
}

func executeObserveOnlyAcceptance(ctx context.Context, scenario AcceptanceConfig, kit *testkit.Scenario) (Artifacts, error) {
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		return Artifacts{}, err
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(scenario.Name), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(scenario.Seed), nil }
	transcript := make([]aiTranscriptEntry, 0)
	observeAI := appendAITranscript(&transcript)
	runtime := onboarding.NewPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
	server := httptest.NewServer(runtime.Handler)
	defer func() { server.Close() }()
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}

	setup, err := setupCampaignGame(ctx, client, scenario)
	if err != nil {
		return Artifacts{}, err
	}
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: setup.GameID, SessionID: setup.Owner.Session, Namespace: scenario.Name}
	if _, err := transport.CurrentProjection(ctx); err != nil {
		return Artifacts{}, fmt.Errorf("bootstrap public projection: %w", err)
	}
	if err := ensureNoHiddenCanary(transport.InitialProjection); err != nil {
		return Artifacts{}, err
	}
	aiCommands, err := runMixedControllerCampaign(ctx, &mixedControllerScheduler{
		scenario:       scenario,
		client:         client,
		setup:          setup,
		viewer:         setup.Owner,
		level:          puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:        &runtime,
		server:         &server,
		db:             db,
		sources:        sources,
		seed:           seed,
		transport:      transport,
		transcript:     &transcript,
		observeAI:      observeAI,
		histories:      map[string][]puppet.PublicExplorationStep{},
		scripted:       map[string]bool{},
		scriptProgress: map[string]int{},
	})
	if err != nil {
		return Artifacts{}, err
	}
	if telemetry.Unknown != nil {
		return Artifacts{}, fmt.Errorf("memorysql rejected production query: %w", telemetry.Unknown)
	}
	aiPlayers := publicAIPlayers(setup)
	if err := validateAITranscript(transcript, setup.GameID, aiPlayers, aiCommands); err != nil {
		return Artifacts{}, err
	}
	transcriptJSON, err := json.Marshal(transcript)
	if err != nil {
		return Artifacts{}, fmt.Errorf("encode AI transcript: %w", err)
	}
	audit, terminal, err := finishObservedTerminal(ctx, db, transport, scenario.Rules.Victory)
	if err != nil {
		return Artifacts{}, err
	}
	return Artifacts{TerminalWinner: terminal.Winner, TerminalReason: terminal.Reason, StateHash: audit.StateHash, EventHash: audit.EventHash, AICommands: aiCommands, AITranscript: string(transcriptJSON)}, nil
}

func executeScriptedHumanAcceptance(ctx context.Context, scenario AcceptanceConfig, kit *testkit.Scenario, level puppet.TestHarnessPolicyLevel) (Artifacts, error) {
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		return Artifacts{}, err
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(scenario.Name), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(scenario.Seed), nil }
	transcript := make([]aiTranscriptEntry, 0)
	observeAI := appendAITranscript(&transcript)
	runtime := onboarding.NewPublicRuntimeWithAIObserver(db, sources, seed, observeAI)
	server := httptest.NewServer(runtime.Handler)
	defer func() { server.Close() }()
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}

	setup, err := setupCampaignGame(ctx, client, scenario)
	if err != nil {
		return Artifacts{}, err
	}
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: setup.GameID, SessionID: setup.Owner.Session, Namespace: scenario.Name}
	if _, err := transport.CurrentProjection(ctx); err != nil {
		return Artifacts{}, fmt.Errorf("bootstrap public projection: %w", err)
	}
	if err := ensureNoHiddenCanary(transport.InitialProjection); err != nil {
		return Artifacts{}, err
	}
	humanScriptModes := map[string]humanScriptMode{}
	for _, player := range publicHumanPlayers(setup) {
		humanScriptModes["account-"+strconv.FormatInt(player.AccountID, 10)] = humanScriptThenPuppet
	}
	aiCommands, err := runMixedControllerCampaign(ctx, &mixedControllerScheduler{
		scenario:         scenario,
		client:           client,
		setup:            setup,
		viewer:           setup.Owner,
		level:            level,
		runtime:          &runtime,
		server:           &server,
		db:               db,
		sources:          sources,
		seed:             seed,
		transport:        transport,
		transcript:       &transcript,
		observeAI:        observeAI,
		histories:        map[string][]puppet.PublicExplorationStep{},
		scripted:         map[string]bool{},
		scriptProgress:   map[string]int{},
		humanScriptModes: humanScriptModes,
	})
	if err != nil {
		return Artifacts{}, err
	}
	if telemetry.Unknown != nil {
		return Artifacts{}, fmt.Errorf("memorysql rejected production query: %w", telemetry.Unknown)
	}
	aiPlayers := publicAIPlayers(setup)
	if err := validateAITranscript(transcript, setup.GameID, aiPlayers, aiCommands); err != nil {
		return Artifacts{}, err
	}
	transcriptJSON, err := json.Marshal(transcript)
	if err != nil {
		return Artifacts{}, fmt.Errorf("encode AI transcript: %w", err)
	}
	audit, terminal, err := finishObservedTerminal(ctx, db, transport, scenario.Rules.Victory)
	if err != nil {
		return Artifacts{}, err
	}
	return Artifacts{TerminalWinner: terminal.Winner, TerminalReason: terminal.Reason, StateHash: audit.StateHash, EventHash: audit.EventHash, AICommands: aiCommands, AITranscript: string(transcriptJSON)}, nil
}

func finishObservedTerminal(ctx context.Context, db *sqlx.DB, transport *HTTPExplorationTransport, expectedReason string) (AuditArtifacts, TerminalFrame, error) {
	terminalCtx, cancelTerminal := context.WithTimeout(ctx, time.Second)
	defer cancelTerminal()
	sequence, name, data, err := terminalPublicEvent(terminalCtx, transport)
	if err != nil {
		return AuditArtifacts{}, TerminalFrame{}, err
	}
	audit, err := AuditTerminal(ctx, db, transport.GameID)
	if err != nil {
		return AuditArtifacts{}, TerminalFrame{}, err
	}
	terminal, err := ParseTerminalSSE(sequence, name, data)
	if err != nil {
		return AuditArtifacts{}, TerminalFrame{}, err
	}
	if err := validateTerminalAgreement(audit, terminal, expectedReason); err != nil {
		return AuditArtifacts{}, TerminalFrame{}, err
	}
	return audit, terminal, nil
}

func publicHumanPlayers(setup publicGameSetup) []publicSetupPlayer {
	players := make([]publicSetupPlayer, 0, len(setup.Players))
	for _, player := range setup.Players {
		if player.Controller == engine.ControllerHuman {
			players = append(players, player)
		}
	}
	return players
}

func containsPlayerID(players []string, playerID string) bool {
	for _, existing := range players {
		if existing == playerID {
			return true
		}
	}
	return false
}

func currentPublicController(ctx context.Context, client onboarding.HTTPClient, actor onboarding.Actor, gameID, accountID int64) (string, error) {
	gameDTO, err := client.ReadGame(ctx, actor, gameID)
	if err != nil {
		return "", err
	}
	for _, player := range gameDTO.Players {
		if player.AccountID == accountID {
			return player.Controller, nil
		}
	}
	return "", fmt.Errorf("public game %d does not expose controller for account %d", gameID, accountID)
}

func publicAIPlayers(setup publicGameSetup) []string {
	players := make([]string, 0, len(setup.Players))
	for _, player := range setup.Players {
		if player.Controller == engine.ControllerAI {
			players = append(players, "account-"+strconv.FormatInt(player.AccountID, 10))
		}
	}
	return players
}

func ensureNoHiddenCanary(data []byte) error {
	if strings.Contains(string(data), "HIDDEN_CANARY") {
		return errors.New("public projection leaked hidden canary")
	}
	return nil
}

func decodeProjectionRevision(data []byte) (int64, error) {
	var projection visibility.Projection
	if err := json.Unmarshal(data, &projection); err != nil {
		return 0, err
	}
	if projection.Revision < 1 {
		return 0, errors.New("public projection revision is invalid")
	}
	return projection.Revision, nil
}

func decodeCommandPayload(raw json.RawMessage) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func resolveContinuityPreparationCommands(scenario AcceptanceConfig, projection []byte, playerID string) ([]ScriptCommand, bool, error) {
	if scenario.Rules.HumanScript != "scripts/score-resignation-continuity.json" {
		return nil, false, nil
	}
	var public visibility.Projection
	if err := json.Unmarshal(projection, &public); err != nil {
		return nil, false, fmt.Errorf("decode public projection: %w", err)
	}
	units := make([]visibility.UnitProjection, 0)
	for _, unit := range public.Units {
		if unit.OwnerID == playerID {
			units = append(units, unit)
		}
	}
	sort.Slice(units, func(i, j int) bool { return units[i].ID < units[j].ID })
	if len(units) < 3 {
		return nil, false, fmt.Errorf("row 03 continuity script requires three visible owned units, got %d", len(units))
	}
	occupiedTiles := make(map[string]bool, len(units))
	for _, unit := range units {
		occupiedTiles[unit.TileID] = true
	}
	if len(occupiedTiles) >= 3 {
		return nil, false, nil
	}
	candidateTiles := make([]visibility.TileProjection, 0, len(public.Tiles))
	for _, tile := range public.Tiles {
		if tile.Stale || tile.Planning == nil || !tile.Planning.Accessible || occupiedTiles[tile.ID] {
			continue
		}
		candidateTiles = append(candidateTiles, tile)
	}
	sort.Slice(candidateTiles, func(i, j int) bool { return candidateTiles[i].ID < candidateTiles[j].ID })
	usedTiles := make(map[string]bool, len(occupiedTiles))
	for tileID := range occupiedTiles {
		usedTiles[tileID] = true
	}
	keptTile := make(map[string]bool, len(occupiedTiles))
	commands := make([]ScriptCommand, 0, len(units)-1)
	for _, unit := range units {
		if len(usedTiles) >= 3 {
			break
		}
		if !keptTile[unit.TileID] {
			keptTile[unit.TileID] = true
			continue
		}
		firstStep := ""
		for _, tile := range candidateTiles {
			path, err := visibility.Plan(public, unit.ID, tile.ID)
			if err != nil || len(path) == 0 {
				continue
			}
			if usedTiles[path[0]] {
				continue
			}
			firstStep = path[0]
			break
		}
		if firstStep == "" {
			return nil, false, errors.New("row 03 continuity preparation could not find distinct accessible city sites")
		}
		payload, err := json.Marshal(map[string]any{"unit_id": unit.ID, "path": []string{firstStep}})
		if err != nil {
			return nil, false, err
		}
		commands = append(commands, ScriptCommand{Command: "move", ActorRole: "human-1", Payload: payload})
		usedTiles[firstStep] = true
	}
	if len(usedTiles) < 3 {
		return nil, false, errors.New("row 03 continuity preparation did not produce three distinct owned city sites")
	}
	return commands, len(commands) > 0, nil
}

func resolveContinuityLeaveCommands(path string, projection []byte, playerID string) ([]ScriptCommand, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var script Script
	if err := json.Unmarshal(raw, &script); err != nil {
		return nil, err
	}
	commands, err := ResolveScript(script, projection, map[string]string{"human-1": playerID})
	if err != nil {
		return nil, err
	}
	if len(commands) == 0 || commands[len(commands)-1].Command != "leave" {
		return nil, errors.New("continuity human script must end with public leave")
	}
	return commands, nil
}

func resolveScriptedHumanCommands(scenario AcceptanceConfig, projection []byte, playerID string) ([]ScriptCommand, bool, error) {
	if scenario.HumanPolicy.Name() == string(HumanPolicyContinuity) {
		commands, preparing, err := resolveContinuityPreparationCommands(scenario, projection, playerID)
		if err != nil {
			return nil, false, err
		}
		if preparing {
			return commands, true, nil
		}
		commands, err = resolveContinuityLeaveCommands(scenario.Rules.HumanScript, projection, playerID)
		if err != nil {
			return nil, false, err
		}
		return commands, false, nil
	}
	commands, _, err := resolveScriptedHumanCommandAt(scenario.Rules.HumanScript, projection, playerID, 0)
	if err != nil {
		return nil, false, err
	}
	return commands, false, nil
}

func resolveScriptedHumanCommandAt(path string, projection []byte, playerID string, index int) ([]ScriptCommand, int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	var script Script
	if err := json.Unmarshal(raw, &script); err != nil {
		return nil, 0, err
	}
	total := len(script.Commands)
	if index >= total {
		return nil, total, nil
	}
	script.Commands = script.Commands[index : index+1]
	commands, err := ResolveScript(script, projection, map[string]string{"human-1": playerID})
	if err != nil {
		return nil, total, err
	}
	return commands, total, nil
}

func recordHumanMove(history map[string][]puppet.PublicExplorationStep, playerID string, candidate puppet.Result) error {
	if candidate.Candidate.Type != engine.CommandMoveUnit {
		return nil
	}
	var payload struct {
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(candidate.Candidate.Payload, &payload); err != nil {
		return err
	}
	if len(payload.Path) == 0 {
		return nil
	}
	history[playerID] = append(history[playerID], puppet.PublicExplorationStep{CommandType: engine.CommandMoveUnit, DestinationTileID: payload.Path[0], Sequence: int64(len(history[playerID]) + 1)})
	return nil
}

type humanScriptMode int

const (
	humanScriptNone humanScriptMode = iota
	humanScriptThenPuppet
	humanScriptOnly
)

type mixedControllerScheduler struct {
	scenario                AcceptanceConfig
	client                  onboarding.HTTPClient
	setup                   publicGameSetup
	viewer                  onboarding.Actor
	level                   puppet.TestHarnessPolicyLevel
	runtime                 **onboarding.PublicRuntime
	server                  **httptest.Server
	db                      *sqlx.DB
	sources                 engine.Sources
	seed                    func() (int64, error)
	transport               *HTTPExplorationTransport
	transcript              *[]aiTranscriptEntry
	observeAI               func(ai.Submission)
	histories               map[string][]puppet.PublicExplorationStep
	longTurnContinuityState map[string]*HTTPExplorationTransport
	scripted                map[string]bool
	scriptProgress          map[string]int
	runTakeovers            bool
	requireAIOnly           bool
	humanScriptModes        map[string]humanScriptMode
}

func (scheduler mixedControllerScheduler) scriptModeFor(playerID string) humanScriptMode {
	if scheduler.humanScriptModes == nil {
		return humanScriptNone
	}
	return scheduler.humanScriptModes[playerID]
}

func (scheduler mixedControllerScheduler) longTurnContinuityTransport(playerID string) *HTTPExplorationTransport {
	if scheduler.scenario.TurnMode != "long_turn" || scheduler.scenario.HumanPolicy.Name() != string(HumanPolicyContinuity) {
		return nil
	}
	if scheduler.longTurnContinuityState == nil {
		return nil
	}
	return scheduler.longTurnContinuityState[playerID]
}

func (scheduler mixedControllerScheduler) currentControllers(ctx context.Context) ([]publicSetupPlayer, []string, error) {
	gameDTO, err := scheduler.client.ReadGame(ctx, scheduler.viewer, scheduler.setup.GameID)
	if err != nil {
		return nil, nil, err
	}
	playersByAccount := make(map[int64]publicSetupPlayer, len(scheduler.setup.Players))
	for _, player := range scheduler.setup.Players {
		playersByAccount[player.AccountID] = player
	}
	humans := make([]publicSetupPlayer, 0, len(gameDTO.Players))
	aiPlayers := make([]string, 0, len(gameDTO.Players))
	for _, player := range gameDTO.Players {
		setupPlayer, ok := playersByAccount[player.AccountID]
		if !ok {
			return nil, nil, fmt.Errorf("public game %d exposed unknown account %d", scheduler.setup.GameID, player.AccountID)
		}
		setupPlayer.Controller = player.Controller
		switch player.Controller {
		case engine.ControllerHuman:
			humans = append(humans, setupPlayer)
		case engine.ControllerAI:
			aiPlayers = append(aiPlayers, "account-"+strconv.FormatInt(player.AccountID, 10))
		}
	}
	return humans, aiPlayers, nil
}

func decodeProjectionTurn(data []byte) (int, error) {
	var projection visibility.Projection
	if err := json.Unmarshal(data, &projection); err != nil {
		return 0, err
	}
	if projection.Turn < 1 {
		return 0, errors.New("public projection turn is invalid")
	}
	return projection.Turn, nil
}

func (scheduler mixedControllerScheduler) refreshObserverProjection(ctx context.Context, contextLabel string) error {
	if _, err := scheduler.transport.CurrentProjection(ctx); err != nil {
		return fmt.Errorf("refresh public projection after %s: %w", contextLabel, err)
	}
	return ensureNoHiddenCanary(scheduler.transport.InitialProjection)
}

func (scheduler *mixedControllerScheduler) maybeRestartAfterCompletedTurn(ctx context.Context, completedTurns int, restarted *bool) error {
	if *restarted || scheduler.scenario.Restart.Point != AcceptanceRestartAfterCompletedTurn || completedTurns < scheduler.scenario.Restart.CompletedTurn {
		return nil
	}
	restartCursor := scheduler.transport.Cursor
	(*scheduler.server).Close()
	*scheduler.runtime = onboarding.RestartPublicRuntimeWithAIObserver(scheduler.db, scheduler.sources, scheduler.seed, scheduler.observeAI)
	*scheduler.server = httptest.NewServer((*scheduler.runtime).Handler)
	scheduler.client.BaseURL = (*scheduler.server).URL
	scheduler.transport.BaseURL = (*scheduler.server).URL
	for _, transport := range scheduler.longTurnContinuityState {
		if transport == nil {
			continue
		}
		transport.BaseURL = (*scheduler.server).URL
	}
	if scheduler.scenario.ObserveRestart != nil {
		scheduler.scenario.ObserveRestart(AcceptanceRestartEvent{Point: scheduler.scenario.Restart.Point, CompletedTurn: completedTurns})
	}
	if _, err := scheduler.transport.CurrentProjection(ctx); err != nil {
		return fmt.Errorf("resume current public projection after restart: %w", err)
	}
	if scheduler.transport.Cursor < restartCursor {
		return fmt.Errorf("restart cursor regressed from %d to %d", restartCursor, scheduler.transport.Cursor)
	}
	if err := ensureNoHiddenCanary(scheduler.transport.InitialProjection); err != nil {
		return err
	}
	*restarted = true
	return nil
}

func publicPlanningStatus(data []byte) (string, error) {
	var projection visibility.Projection
	if err := json.Unmarshal(data, &projection); err != nil {
		return "", err
	}
	return projection.PlanningStatus, nil
}

func hybridBloodlustTargetPriority(scenario AcceptanceConfig, playerID string) int {
	if !strings.HasPrefix(playerID, "account-") {
		return 2
	}
	accountIndex, err := strconv.Atoi(strings.TrimPrefix(playerID, "account-"))
	if err != nil || accountIndex < 1 || accountIndex > len(scenario.Slots) {
		return 2
	}
	if scenario.Slots[accountIndex-1].ControllerKind == engine.ControllerAI {
		return 0
	}
	if scenario.Slots[accountIndex-1].ControllerKind == engine.ControllerHuman {
		return 1
	}
	return 2
}

func sortHybridBloodlustTargets(scenario AcceptanceConfig, playerIDs []string) {
	sort.Slice(playerIDs, func(i, j int) bool {
		leftPriority := hybridBloodlustTargetPriority(scenario, playerIDs[i])
		rightPriority := hybridBloodlustTargetPriority(scenario, playerIDs[j])
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return playerIDs[i] < playerIDs[j]
	})
}

func hybridBloodlustDeclaredTargetStage(progress int) int {
	if progress >= 0 {
		return 0
	}
	if progress == -1 {
		return 1
	}
	magnitude := -progress
	if magnitude < 1000 {
		return 0
	}
	stage := magnitude/1000 - 1
	if stage < 0 {
		return 0
	}
	return stage
}

func conquestVisited(history []puppet.PublicExplorationStep) map[string]bool {
	visited := make(map[string]bool, len(history))
	for _, step := range history {
		if step.DestinationTileID == "" {
			continue
		}
		visited[step.DestinationTileID] = true
	}
	return visited
}

func isBenignPublicConflictMiss(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, onboarding.ErrConflict) {
		return true
	}
	message := err.Error()
	return strings.Contains(message, `public endpoint status 422:{"error":"gameplay request rejected"}`) || strings.Contains(message, `public endpoint status 422: {"error":"gameplay request rejected"}`) || strings.Contains(message, "public exploration combat status 422") || strings.Contains(message, "public exploration move status 422")
}

type bloodlustHybridAction struct {
	Kind           string
	DefenderArmyID string
	Path           []string
}

func chooseBloodlustHybridAction(projection []byte, unitID string, visited map[string]bool) (bloodlustHybridAction, error) {
	decoded, mover, err := explorationInputs(projection, unitID)
	if err != nil {
		return bloodlustHybridAction{}, err
	}
	if enemy, ok := visibleEnemy(decoded, mover); ok {
		defenderArmyID, err := ResolveVisibleArmy(projection, enemy.ID)
		if err == nil {
			return bloodlustHybridAction{Kind: "combat", DefenderArmyID: defenderArmyID}, nil
		}
	}
	var origin *visibility.TileProjection
	for _, tile := range decoded.Tiles {
		if tile.ID == mover.TileID {
			copy := tile
			origin = &copy
			break
		}
	}
	tileByID := make(map[string]visibility.TileProjection, len(decoded.Tiles))
	for _, tile := range decoded.Tiles {
		tileByID[tile.ID] = tile
	}
	var target *visibility.TileProjection
	bestDistance := 0
	targetIsCity := false
	consider := func(tile visibility.TileProjection, isCity bool) {
		distance := 0
		if origin != nil {
			distance = abs(tile.X-origin.X) + abs(tile.Y-origin.Y)
		}
		if target == nil || distance < bestDistance || (distance == bestDistance && isCity && !targetIsCity) {
			copy := tile
			target = &copy
			bestDistance = distance
			targetIsCity = isCity
		}
	}
	cities := append([]visibility.CityProjection(nil), decoded.Cities...)
	sort.Slice(cities, func(i, j int) bool { return cities[i].ID < cities[j].ID })
	for _, city := range cities {
		if city.OwnerID == "" || city.OwnerID == mover.OwnerID {
			continue
		}
		tile, ok := tileByID[city.TileID]
		if !ok {
			continue
		}
		consider(tile, true)
	}
	units := append([]visibility.UnitProjection(nil), decoded.Units...)
	sort.Slice(units, func(i, j int) bool { return units[i].ID < units[j].ID })
	for _, unit := range units {
		if unit.OwnerID == "" || unit.OwnerID == mover.OwnerID {
			continue
		}
		tile, ok := tileByID[unit.TileID]
		if !ok {
			continue
		}
		consider(tile, false)
	}
	if target == nil && origin != nil {
		if farthest, ok := farthestKnownLandTile(decoded, *origin); ok {
			target = &farthest
		}
	}
	var path []string
	var ok bool
	if target != nil {
		_, path, ok = nextReachableFrontierToward(decoded, mover, visited, unitID, *target)
		if !ok {
			_, path, ok = nextReachableFrontierToward(decoded, mover, nil, unitID, *target)
		}
	} else {
		_, path, ok = nextReachableFrontier(decoded, mover, visited, unitID)
		if !ok {
			_, path, ok = nextReachableFrontier(decoded, mover, nil, unitID)
		}
	}
	if ok {
		if mover.Movement > 0 && len(path) > mover.Movement {
			path = path[:mover.Movement]
		}
		return bloodlustHybridAction{Kind: "move", Path: path}, nil
	}
	return bloodlustHybridAction{Kind: "pass"}, nil
}

func driveScriptedHuman(ctx context.Context, client onboarding.HTTPClient, scenario AcceptanceConfig, gameID int64, player publicSetupPlayer, playerID, baseURL string, level puppet.TestHarnessPolicyLevel, transport *HTTPExplorationTransport, scriptMode humanScriptMode, histories map[string][]puppet.PublicExplorationStep, scripted map[string]bool, scriptProgress map[string]int) (int, error) {
	if scriptMode == humanScriptOnly && scripted[playerID] {
		return 0, nil
	}
	if transport == nil {
		transport = &HTTPExplorationTransport{BaseURL: baseURL, Client: client.Client, GameID: gameID, SessionID: player.Actor.Session, Namespace: scenario.Name + "-" + playerID}
		if _, _, err := transport.Bootstrap(ctx, 0); err != nil {
			return 0, fmt.Errorf("bootstrap public projection for %s: %w", playerID, err)
		}
	} else {
		transport.BaseURL = baseURL
		transport.Client = client.Client
	}
	projection, err := transport.CurrentProjection(ctx)
	if err != nil {
		return 0, fmt.Errorf("bootstrap public projection for %s: %w", playerID, err)
	}
	if err := ensureNoHiddenCanary(projection); err != nil {
		return 0, fmt.Errorf("%s: %w", playerID, err)
	}
	revision := transport.Revision
	accepted := 0
	var planning visibility.Projection
	if err := json.Unmarshal(projection, &planning); err != nil {
		return 0, fmt.Errorf("decode public projection for %s: %w", playerID, err)
	}
	if scriptMode != humanScriptNone && scenario.HumanPolicy.Name() != string(HumanPolicyContinuity) && (scriptProgress[playerID] > 0 || scripted[playerID]) {
		switch planning.PlanningStatus {
		case visibility.PlanningStatusClosed:
			return 0, nil
		case visibility.PlanningStatusQueued:
			if scriptMode == humanScriptOnly {
				return 0, nil
			}
			nextRevision, commandErr := client.GameplayCommand(ctx, player.Actor, gameID, "round-complete", revision, fmt.Sprintf("%s-%s-round-complete-%d", scenario.Name, playerID, revision), map[string]any{})
			if commandErr != nil {
				return 0, fmt.Errorf("scripted-human %s command round-complete: %w", playerID, commandErr)
			}
			revision = nextRevision
			return 1, nil
		}
	}
	if scriptMode != humanScriptNone && !scripted[playerID] && scenario.Rules.HumanScript != "" {
		planningMode := scenario.TurnMode == "simultaneous" || scenario.TurnMode == "hybrid"
		preparing := false
		var commands []ScriptCommand
		totalCommands := 0
		if planningMode && scenario.HumanPolicy.Name() != string(HumanPolicyContinuity) {
			next := scriptProgress[playerID]
			commands, totalCommands, err = resolveScriptedHumanCommandAt(scenario.Rules.HumanScript, projection, playerID, next)
			if err != nil {
				return 0, fmt.Errorf("resolve scripted-human commands for %s: %w", playerID, err)
			}
			if next >= totalCommands {
				scripted[playerID] = true
				commands = nil
			}
		} else {
			commands, preparing, err = resolveScriptedHumanCommands(scenario, projection, playerID)
			if err != nil {
				if scenario.HumanPolicy.Name() == string(HumanPolicyContinuity) {
					return 0, fmt.Errorf("prepare scripted-human commands for %s: %w", playerID, err)
				}
				return 0, fmt.Errorf("resolve scripted-human commands for %s: %w", playerID, err)
			}
		}
		for index, command := range commands {
			payload, err := decodeCommandPayload(command.Payload)
			if err != nil {
				return 0, err
			}
			nextRevision, commandErr := client.GameplayCommand(ctx, player.Actor, gameID, command.Command, revision, fmt.Sprintf("%s-%s-script-%d-%d", scenario.Name, playerID, index, revision), payload)
			if commandErr != nil {
				return 0, fmt.Errorf("scripted-human %s command %s: %w", playerID, command.Command, commandErr)
			}
			revision = nextRevision
			accepted++
			if planningMode && scenario.HumanPolicy.Name() != string(HumanPolicyContinuity) {
				scriptProgress[playerID]++
				if scriptProgress[playerID] >= totalCommands {
					scripted[playerID] = true
				}
			}
		}
		if scenario.HumanPolicy.Name() == string(HumanPolicyContinuity) && scenario.TurnMode == "long_turn" && preparing {
			projection, err = transport.CurrentProjection(ctx)
			if err != nil {
				return 0, fmt.Errorf("refresh public projection for %s continuity stage: %w", playerID, err)
			}
			if err := ensureNoHiddenCanary(projection); err != nil {
				return 0, fmt.Errorf("%s continuity stage projection: %w", playerID, err)
			}
			return accepted, nil
		}
		if !preparing {
			if !(planningMode && scenario.HumanPolicy.Name() != string(HumanPolicyContinuity) && !scripted[playerID]) {
				scripted[playerID] = true
			}
			if scriptMode == humanScriptOnly {
				return accepted, nil
			}
		}
	}
	projection, err = transport.CurrentProjection(ctx)
	if err != nil {
		return 0, fmt.Errorf("refresh public projection for %s: %w", playerID, err)
	}
	if err := ensureNoHiddenCanary(projection); err != nil {
		return 0, fmt.Errorf("%s refreshed projection: %w", playerID, err)
	}
	revision = transport.Revision
	if err := json.Unmarshal(projection, &planning); err != nil {
		return 0, fmt.Errorf("decode public projection for %s: %w", playerID, err)
	}
	switch planning.PlanningStatus {
	case visibility.PlanningStatusClosed:
		return accepted, nil
	case visibility.PlanningStatusQueued:
		nextRevision, commandErr := client.GameplayCommand(ctx, player.Actor, gameID, "round-complete", revision, fmt.Sprintf("%s-%s-round-complete-%d", scenario.Name, playerID, revision), map[string]any{})
		if commandErr != nil {
			return 0, fmt.Errorf("scripted-human %s command round-complete: %w", playerID, commandErr)
		}
		revision = nextRevision
		return accepted + 1, nil
	}
	if scenario.TurnMode == "hybrid" && scenario.Rules.Victory == "bloodlust_conquest" && scripted[playerID] {
		visibleTargets := make(map[string]bool)
		for _, city := range planning.Cities {
			if city.OwnerID != "" && city.OwnerID != playerID {
				visibleTargets[city.OwnerID] = true
			}
		}
		for _, unit := range planning.Units {
			if unit.OwnerID != "" && unit.OwnerID != playerID {
				visibleTargets[unit.OwnerID] = true
			}
		}
		targetIDs := make([]string, 0, len(visibleTargets))
		for targetID := range visibleTargets {
			targetIDs = append(targetIDs, targetID)
		}
		sortHybridBloodlustTargets(scenario, targetIDs)
		declaredStage := hybridBloodlustDeclaredTargetStage(scriptProgress[playerID])
		if declaredStage < len(targetIDs) {
			targetID := targetIDs[declaredStage]
			_, commandErr := client.GameplayCommand(ctx, player.Actor, gameID, "set-war", revision, fmt.Sprintf("%s-%s-set-war-%d-%s", scenario.Name, playerID, revision, targetID), map[string]any{"target_player_id": targetID, "at_war": true})
			scriptProgress[playerID] = -((declaredStage + 2) * 1000) - planning.Turn
			if commandErr == nil {
				return accepted + 1, nil
			}
			if !isBenignPublicConflictMiss(commandErr) {
				return 0, fmt.Errorf("scripted-human %s command set-war: %w", playerID, commandErr)
			}
		}
		candidateUnits := make([]visibility.UnitProjection, 0, len(planning.Units))
		for _, unit := range planning.Units {
			if unit.OwnerID != playerID || unit.Movement <= 0 {
				continue
			}
			if unit.ArmyID == "" {
				continue
			}
			candidateUnits = append(candidateUnits, unit)
		}
		sort.Slice(candidateUnits, func(i, j int) bool { return candidateUnits[i].ID < candidateUnits[j].ID })
		visited := conquestVisited(histories[playerID])
		for _, unit := range candidateUnits {
			action, actionErr := chooseBloodlustHybridAction(projection, unit.ID, visited)
			if actionErr != nil {
				continue
			}
			switch action.Kind {
			case "combat":
				attackerArmyID := unit.ArmyID
				if attackerArmyID == "" {
					attackerArmyID, actionErr = ResolveVisibleArmy(projection, unit.ID)
					if actionErr != nil {
						continue
					}
				}
				transition, combatErr := transport.Combat(ctx, attackerArmyID, action.DefenderArmyID)
				if combatErr != nil {
					if isBenignPublicConflictMiss(combatErr) {
						continue
					}
					return 0, fmt.Errorf("scripted-human %s combat: %w", playerID, combatErr)
				}
				if err := ensureNoHiddenCanary(transition.Projection); err != nil {
					return 0, fmt.Errorf("%s combat projection: %w", playerID, err)
				}
				return accepted + 1, nil
			case "move":
				if len(action.Path) == 0 {
					continue
				}
				nextProjection, moveErr := transport.Move(ctx, unit.ID, action.Path[:1])
				if moveErr != nil {
					if isBenignPublicConflictMiss(moveErr) {
						continue
					}
					return 0, fmt.Errorf("scripted-human %s move-unit: %w", playerID, moveErr)
				}
				if err := ensureNoHiddenCanary(nextProjection); err != nil {
					return 0, fmt.Errorf("%s movement projection: %w", playerID, err)
				}
				histories[playerID] = append(histories[playerID], puppet.PublicExplorationStep{CommandType: engine.CommandMoveUnit, DestinationTileID: action.Path[0], Sequence: int64(len(histories[playerID]) + 1)})
				return accepted + 1, nil
			}
		}
	}
	historyJSON, err := json.Marshal(histories[playerID])
	if err != nil {
		return 0, err
	}
	var observed puppet.RawPlanningInput
	result, err := (puppet.Puppeteer{Client: puppet.Client{BaseURL: baseURL, HTTPClient: client.Client}, ObservePlanningInput: func(input puppet.RawPlanningInput) { observed = input }}).SelectAndSubmit(ctx, puppet.HumanActor{GameID: gameID, SessionID: player.Actor.Session}, puppet.PlanningInput{ActorPlayerID: playerID, ExpectedRevision: revision, PublicProjectionJSON: projection, PublicHistoryJSON: historyJSON, TestHarnessPolicyLevel: level}, fmt.Sprintf("%s-%s-human-%d", scenario.Name, playerID, revision))
	if err != nil {
		var httpErr *puppet.HTTPError
		if errors.As(err, &httpErr) && (httpErr.Status == http.StatusConflict || httpErr.Status == http.StatusUnprocessableEntity || httpErr.Status == http.StatusForbidden) {
			return accepted, nil
		}
		return 0, fmt.Errorf("puppet drive for %s: %w (session=%d revision=%d actor=%q projection_json=%t history_json=%t)", playerID, err, player.Actor.Session, revision, playerID, json.Valid(projection), json.Valid(historyJSON))
	}
	if err := ensureNoHiddenCanary(observed.PublicProjectionJSON); err != nil {
		return 0, fmt.Errorf("%s planning projection: %w", playerID, err)
	}
	if err := ensureNoHiddenCanary(observed.PublicHistoryJSON); err != nil {
		return 0, fmt.Errorf("%s planning history: %w", playerID, err)
	}
	if err := recordHumanMove(histories, playerID, result); err != nil {
		return 0, err
	}
	return accepted + 1, nil
}

func driveAIPlayers(ctx context.Context, runtime *onboarding.PublicRuntime, transport *HTTPExplorationTransport, gameID int64, aiPlayers []string, transcript *[]aiTranscriptEntry, maxCommands int) (int, string, error) {
	commands := 0
	progressed := false
	expectedRevision := transport.Revision
	lastActor := ""
	zeroCommandPlayers := make([]string, 0, len(aiPlayers))
	ignoredErrors := make([]string, 0, len(aiPlayers))
	for _, player := range aiPlayers {
		if maxCommands > 0 && commands >= maxCommands {
			break
		}
		transcriptStart := len(*transcript)
		result, err := runtime.Driver.Drive(ctx, gameID, player, expectedRevision)
		if errors.Is(err, ai.ErrTerminal) {
			observed := len(*transcript) - transcriptStart
			if observationErr := validateAIObservation(result.Commands, observed); observationErr != nil {
				return commands, player, observationErr
			}
			commands += result.Commands
			transport.Terminal = true
			return commands, player, ai.ErrTerminal
		}
		if err != nil {
			message := err.Error()
			if strings.Contains(message, "command is not allowed in this phase") || strings.Contains(message, "player controller is not authorized") {
				ignoredErrors = append(ignoredErrors, player+":"+message)
				continue
			}
			return commands, player, fmt.Errorf("production AI drive for %s: %w; recent AI transcript=%s", player, err, recentAITranscript(*transcript, 8))
		}
		observed := len(*transcript) - transcriptStart
		if err := validateAIObservation(result.Commands, observed); err != nil {
			return commands, player, err
		}
		if result.Commands == 0 {
			zeroCommandPlayers = append(zeroCommandPlayers, player)
			continue
		}
		progressed = true
		commands += result.Commands
		expectedRevision = result.Revision
		lastActor = player
	}
	if !progressed && !transport.Terminal {
		return commands, lastActor, fmt.Errorf("production AI could not make progress at the current policy boundary; ai_players=%v zero_command_players=%v ignored_errors=%v", aiPlayers, zeroCommandPlayers, ignoredErrors)
	}
	if progressed && !transport.Terminal {
		if _, err := transport.CurrentProjection(ctx); err != nil {
			return commands, lastActor, fmt.Errorf("refresh public projection after production AI drive batch: %w", err)
		}
		if transport.Revision != expectedRevision {
			return commands, lastActor, fmt.Errorf("production AI drive batch returned revision %d but current public projection is %d", expectedRevision, transport.Revision)
		}
	}
	return commands, lastActor, nil
}

func runMixedControllerCampaign(ctx context.Context, scheduler *mixedControllerScheduler) (int, error) {
	currentTurn, err := decodeProjectionTurn(scheduler.transport.InitialProjection)
	if err != nil {
		return 0, err
	}
	completedTurns := 0
	aiCommands := 0
	totalAccepted := 0
	restarted := false
	lastActor := ""
	pendingTakeoverWaits := 0
	humanDriveDebug := []string(nil)
	for completedTurns < scheduler.scenario.Budgets.Turns && !scheduler.transport.Terminal {
		if totalAccepted >= scheduler.scenario.Budgets.Commands {
			err := commandBudgetError(scheduler.scenario.TurnMode, currentTurn, scheduler.transport.Revision, lastActor, totalAccepted, scheduler.scenario.Budgets.Commands, *scheduler.transcript)
			if scheduler.scenario.TurnMode == "hybrid" {
				return aiCommands, fmt.Errorf("%w; human_loop=%v", err, humanDriveDebug)
			}
			return aiCommands, err
		}
		pendingTakeover := false
		if scheduler.runTakeovers {
			if err := (*scheduler.runtime).RunPendingTakeovers(ctx); err != nil {
				return aiCommands, fmt.Errorf("run pending takeovers: %w", err)
			}
			if err := scheduler.refreshObserverProjection(ctx, "pending takeover worker"); err != nil {
				return aiCommands, err
			}
			for playerID, mode := range scheduler.humanScriptModes {
				if mode != humanScriptOnly || !scheduler.scripted[playerID] {
					continue
				}
				accountID, parseErr := strconv.ParseInt(strings.TrimPrefix(playerID, "account-"), 10, 64)
				if parseErr != nil || accountID < 1 {
					return aiCommands, fmt.Errorf("continuity leave player id %q is invalid", playerID)
				}
				controller, err := currentPublicController(ctx, scheduler.client, scheduler.viewer, scheduler.setup.GameID, accountID)
				if err != nil {
					return aiCommands, fmt.Errorf("read continuity takeover controller for %s: %w", playerID, err)
				}
				if controller == engine.ControllerHuman {
					pendingTakeover = true
				}
			}
		}
		humans, aiPlayers, err := scheduler.currentControllers(ctx)
		if err != nil {
			return aiCommands, err
		}
		filteredHumans := humans[:0]
		for _, player := range humans {
			playerID := "account-" + strconv.FormatInt(player.AccountID, 10)
			if scheduler.humanScriptModes[playerID] == humanScriptOnly && scheduler.scripted[playerID] {
				continue
			}
			filteredHumans = append(filteredHumans, player)
		}
		humans = filteredHumans
		if scheduler.requireAIOnly && len(humans) > 0 {
			return aiCommands, fmt.Errorf("human controllers remain at AI-only closeout boundary")
		}
		progressed := false
		humanPlanningPending := false
		humanDriveDebug = humanDriveDebug[:0]
		planningMode := scheduler.scenario.TurnMode == "simultaneous"
		if scheduler.scenario.TurnMode == "hybrid" {
			status, statusErr := publicPlanningStatus(scheduler.transport.InitialProjection)
			if statusErr != nil {
				return aiCommands, statusErr
			}
			planningMode = status == visibility.PlanningStatusOpen || status == visibility.PlanningStatusQueued
		}
		for _, player := range humans {
			playerID := "account-" + strconv.FormatInt(player.AccountID, 10)
			transport := scheduler.longTurnContinuityTransport(playerID)
			accepted, err := driveScriptedHuman(ctx, scheduler.client, scheduler.scenario, scheduler.setup.GameID, player, playerID, (*scheduler.server).URL, scheduler.level, transport, scheduler.scriptModeFor(playerID), scheduler.histories, scheduler.scripted, scheduler.scriptProgress)
			debugStatus := ""
			debugRevision := int64(0)
			debugTransport := transport
			if debugTransport == nil {
				debugTransport = &HTTPExplorationTransport{BaseURL: (*scheduler.server).URL, Client: scheduler.client.Client, GameID: scheduler.setup.GameID, SessionID: player.Actor.Session, Namespace: scheduler.scenario.Name + "-debug-" + playerID}
			}
			if projection, projectionErr := debugTransport.CurrentProjection(ctx); projectionErr == nil {
				debugRevision = debugTransport.Revision
				if status, statusErr := publicPlanningStatus(projection); statusErr == nil {
					debugStatus = status
					if scheduler.scenario.TurnMode == "hybrid" && (status == visibility.PlanningStatusOpen || status == visibility.PlanningStatusQueued) {
						humanPlanningPending = true
					}
				}
			}
			humanDriveDebug = append(humanDriveDebug, fmt.Sprintf("%s:accepted=%d status=%q revision=%d progress=%d err=%v", playerID, accepted, debugStatus, debugRevision, scheduler.scriptProgress[playerID], err))
			if err != nil {
				return aiCommands, err
			}
			if accepted == 0 {
				continue
			}
			lastActor = playerID
			totalAccepted += accepted
			progressed = true
			if err := scheduler.refreshObserverProjection(ctx, "human drive"); err != nil {
				return aiCommands, err
			}
			if scheduler.scenario.TurnMode == "hybrid" {
				break
			}
			if !planningMode {
				break
			}
		}
		if totalAccepted >= scheduler.scenario.Budgets.Commands && !scheduler.transport.Terminal {
			err := commandBudgetError(scheduler.scenario.TurnMode, currentTurn, scheduler.transport.Revision, lastActor, totalAccepted, scheduler.scenario.Budgets.Commands, *scheduler.transcript)
			if scheduler.scenario.TurnMode == "hybrid" {
				return aiCommands, fmt.Errorf("%w; human_loop=%v", err, humanDriveDebug)
			}
			return aiCommands, err
		}
		if scheduler.scenario.TurnMode == "hybrid" && progressed {
			nextTurn, err := decodeProjectionTurn(scheduler.transport.InitialProjection)
			if err != nil {
				return aiCommands, err
			}
			if nextTurn < currentTurn {
				return aiCommands, fmt.Errorf("public projection turn regressed from %d to %d", currentTurn, nextTurn)
			}
			completedTurns += nextTurn - currentTurn
			currentTurn = nextTurn
			if err := scheduler.maybeRestartAfterCompletedTurn(ctx, completedTurns, &restarted); err != nil {
				return aiCommands, err
			}
			pendingTakeoverWaits = 0
			continue
		}
		if scheduler.scenario.TurnMode == "hybrid" && humanPlanningPending {
			pendingTakeoverWaits = 0
			continue
		}
		if len(aiPlayers) > 0 && (planningMode || !progressed) {
			commands, actor, err := driveAIPlayers(ctx, *scheduler.runtime, scheduler.transport, scheduler.setup.GameID, aiPlayers, scheduler.transcript, scheduler.scenario.Budgets.Commands-totalAccepted)
			aiCommands += commands
			totalAccepted += commands
			if actor != "" {
				lastActor = actor
			}
			if err != nil {
				if errors.Is(err, ai.ErrTerminal) {
					scheduler.transport.Terminal = true
					break
				}
				if strings.Contains(err.Error(), "production AI could not make progress at the current policy boundary") {
					if pendingTakeover {
						pendingTakeoverWaits++
						if pendingTakeoverWaits > 4 {
							return aiCommands, fmt.Errorf("pending takeover worker did not convert the departed human after %d idle scheduler cycles: %w", pendingTakeoverWaits, err)
						}
						continue
					}
					if scheduler.scenario.TurnMode == "hybrid" {
						return aiCommands, fmt.Errorf("%w; human_loop=%v", err, humanDriveDebug)
					}
					return aiCommands, err
				}
				return aiCommands, err
			}
			progressed = progressed || commands > 0
			if progressed {
				if err := ensureNoHiddenCanary(scheduler.transport.InitialProjection); err != nil {
					return aiCommands, err
				}
			}
		}
		if scheduler.transport.Terminal {
			break
		}
		if totalAccepted >= scheduler.scenario.Budgets.Commands {
			err := commandBudgetError(scheduler.scenario.TurnMode, currentTurn, scheduler.transport.Revision, lastActor, totalAccepted, scheduler.scenario.Budgets.Commands, *scheduler.transcript)
			if scheduler.scenario.TurnMode == "hybrid" {
				return aiCommands, fmt.Errorf("%w; human_loop=%v", err, humanDriveDebug)
			}
			return aiCommands, err
		}
		if !progressed {
			if pendingTakeover {
				pendingTakeoverWaits++
				if pendingTakeoverWaits > 4 {
					return aiCommands, fmt.Errorf("pending takeover worker did not convert the departed human after %d idle scheduler cycles", pendingTakeoverWaits)
				}
				continue
			}
			return aiCommands, errors.New("mixed-controller campaign could not make progress at the current mode boundary")
		}
		pendingTakeoverWaits = 0
		nextTurn, err := decodeProjectionTurn(scheduler.transport.InitialProjection)
		if err != nil {
			return aiCommands, err
		}
		if nextTurn < currentTurn {
			return aiCommands, fmt.Errorf("public projection turn regressed from %d to %d", currentTurn, nextTurn)
		}
		completedTurns += nextTurn - currentTurn
		currentTurn = nextTurn
		if err := scheduler.maybeRestartAfterCompletedTurn(ctx, completedTurns, &restarted); err != nil {
			return aiCommands, err
		}
	}
	return aiCommands, nil
}

func runAIOnlyGame(ctx context.Context, scenario AcceptanceConfig, client onboarding.HTTPClient, setup publicGameSetup, viewer onboarding.Actor, db *sqlx.DB, sources engine.Sources, seed func() (int64, error), runtime **onboarding.PublicRuntime, server **httptest.Server, transport *HTTPExplorationTransport, transcript *[]aiTranscriptEntry, observeAI func(ai.Submission)) (int, error) {
	scheduler := mixedControllerScheduler{
		scenario:       scenario,
		client:         client,
		setup:          setup,
		viewer:         viewer,
		level:          puppet.TestHarnessPolicyLevelProductionFirst,
		runtime:        runtime,
		server:         server,
		db:             db,
		sources:        sources,
		seed:           seed,
		transport:      transport,
		transcript:     transcript,
		observeAI:      observeAI,
		histories:      map[string][]puppet.PublicExplorationStep{},
		scripted:       map[string]bool{},
		scriptProgress: map[string]int{},
		requireAIOnly:  true,
	}
	return runMixedControllerCampaign(ctx, &scheduler)
}

type publicSetupPlayer struct {
	Actor      onboarding.Actor
	AccountID  int64
	Country    string
	Controller string
}

type publicGameSetup struct {
	Owner   onboarding.Actor
	GameID  int64
	Players []publicSetupPlayer
}

func setupCampaignGame(ctx context.Context, client onboarding.HTTPClient, scenario AcceptanceConfig) (publicGameSetup, error) {
	selectable, err := campaignPlayableCivilizations()
	if err != nil {
		return publicGameSetup{}, err
	}
	if err := validateSetupSlots(scenario.Slots, selectable); err != nil {
		return publicGameSetup{}, err
	}
	players := make([]publicSetupPlayer, len(scenario.Slots))
	for index, setup := range scenario.Slots {
		actor, err := client.Signup(ctx, fmt.Sprintf("%s-player-%d", scenario.Name, setup.Slot), fmt.Sprintf("%s-player-%d@example.invalid", scenario.Name, setup.Slot))
		if err != nil {
			return publicGameSetup{}, err
		}
		players[index] = publicSetupPlayer{Actor: actor, Country: setup.CountryRecordID, Controller: setup.ControllerKind}
	}
	owner := players[0].Actor
	lobbyID, err := client.CreateLobby(ctx, owner, scenario.Name+"-lobby")
	if err != nil {
		return publicGameSetup{}, err
	}
	for _, player := range players[1:] {
		if err := client.JoinLobby(ctx, player.Actor, lobbyID); err != nil {
			return publicGameSetup{}, err
		}
	}
	gameID, err := client.CreateGame(ctx, owner, lobbyID, scenario.Name)
	if err != nil {
		return publicGameSetup{}, err
	}
	if err := client.ConfigureGame(ctx, owner, gameID, onboarding.GameSetup{Style: scenario.TurnMode, RulesJSON: acceptanceRulesJSON(scenario.Rules), MapSize: scenario.World.MapSize, Climate: scenario.World.Climate, Wetness: &scenario.World.Wetness}); err != nil {
		return publicGameSetup{}, err
	}
	if err := client.UpdateStartingSettlers(ctx, owner, gameID, scenario.Budgets.StartingSettlers); err != nil {
		return publicGameSetup{}, err
	}
	for _, player := range players {
		if err := client.SelectCountry(ctx, player.Actor, gameID, player.Country); err != nil {
			return publicGameSetup{}, err
		}
	}
	gameDTO, err := client.ReadGame(ctx, owner, gameID)
	if err != nil {
		return publicGameSetup{}, err
	}
	for index := range players {
		accountID, err := accountIDForCountry(gameDTO, players[index].Country)
		if err != nil {
			return publicGameSetup{}, err
		}
		players[index].AccountID = accountID
	}
	for _, player := range players {
		if player.Controller != engine.ControllerAI {
			continue
		}
		if err := client.AddAI(ctx, owner, gameID, player.AccountID); err != nil {
			return publicGameSetup{}, err
		}
	}
	if err := client.StartGame(ctx, owner, gameID); err != nil {
		return publicGameSetup{}, err
	}
	return publicGameSetup{Owner: owner, GameID: gameID, Players: players}, nil
}

func setupDiplomaticGame(ctx context.Context, client onboarding.HTTPClient, scenario AcceptanceConfig) (onboarding.Actor, int64, string, []string, error) {
	setup, err := setupCampaignGame(ctx, client, scenario)
	if err != nil {
		return onboarding.Actor{}, 0, "", nil, err
	}
	aiPlayers := make([]string, 0, len(setup.Players))
	for _, player := range setup.Players {
		if player.Controller != engine.ControllerAI {
			continue
		}
		aiPlayers = append(aiPlayers, "account-"+strconv.FormatInt(player.AccountID, 10))
	}
	return setup.Owner, setup.GameID, "account-" + strconv.FormatInt(setup.Players[0].AccountID, 10), aiPlayers, nil
}

func setupPublicGame(ctx context.Context, client onboarding.HTTPClient, scenario AcceptanceConfig) (onboarding.Actor, int64, string, string, error) {
	setup, err := setupCampaignGame(ctx, client, scenario)
	if err != nil {
		return onboarding.Actor{}, 0, "", "", err
	}
	if len(setup.Players) != 2 {
		return onboarding.Actor{}, 0, "", "", errors.New("public setup requires two players")
	}
	return setup.Owner, setup.GameID, "account-" + strconv.FormatInt(setup.Players[0].AccountID, 10), "account-" + strconv.FormatInt(setup.Players[1].AccountID, 10), nil
}

func accountIDForCountry(gameDTO game.Game, country string) (int64, error) {
	for _, player := range gameDTO.Players {
		if player.Country == country && player.AccountID > 0 {
			return player.AccountID, nil
		}
	}
	return 0, errors.New("public game does not expose the guest account role")
}

func resolveConquestRoles(path string, projection []byte, humanPlayer string) (string, string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	var script Script
	if err := json.Unmarshal(raw, &script); err != nil {
		return "", "", err
	}
	commands, err := ResolveScript(script, projection, map[string]string{"human-1": humanPlayer})
	if err != nil || len(commands) == 0 || commands[0].Command != "army-create" {
		return "", "", errors.New("military script did not resolve one public army command")
	}
	var payload struct {
		ArmyID  string   `json:"army_id"`
		UnitIDs []string `json:"unit_ids"`
	}
	if err := json.Unmarshal(commands[0].Payload, &payload); err != nil || payload.ArmyID == "" || len(payload.UnitIDs) != 1 {
		return "", "", errors.New("military script resolved invalid public roles")
	}
	return payload.UnitIDs[0], payload.ArmyID, nil
}

func resumePublicEvents(ctx context.Context, transport *HTTPExplorationTransport, cursor int64) (int64, []byte, error) {
	return scanPublicEvents(ctx, transport, cursor, false)
}

func terminalPublicEvent(ctx context.Context, transport *HTTPExplorationTransport) (int64, string, []byte, error) {
	cursor, frame, err := scanPublicEvents(ctx, transport, 0, true)
	return cursor, "terminal", frame, err
}

func scanPublicEvents(ctx context.Context, transport *HTTPExplorationTransport, cursor int64, terminal bool) (int64, []byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/game/%d/events?session_id=%d", transport.BaseURL, transport.GameID, transport.SessionID), nil)
	if err != nil {
		return 0, nil, err
	}
	request.Header.Set("Last-Event-ID", strconv.FormatInt(cursor, 10))
	response, err := transport.client().Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("public SSE status %d", response.StatusCode)
	}
	var id int64
	var event string
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), maxPublicSSELineBytes)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "id: "):
			id, err = strconv.ParseInt(strings.TrimPrefix(line, "id: "), 10, 64)
			if err != nil {
				return 0, nil, err
			}
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data := []byte(strings.TrimPrefix(line, "data: "))
			if terminal && event == "terminal" {
				if !json.Valid(data) {
					return 0, nil, errors.New("terminal SSE frame is invalid")
				}
				return id, data, nil
			}
			if !terminal && id > cursor {
				return id, data, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, nil, err
	}
	return 0, nil, errors.New("public SSE did not yield requested event")
}
