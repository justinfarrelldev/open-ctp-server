package fullgame

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func TestPublicLeaveTransitionsToAutonomousAIWithoutPublicTakeoverCall(t *testing.T) {
	fixture := newLeaveContinuityFixture(t)
	defer fixture.Close()

	if _, err := fixture.transport.CurrentProjection(fixture.ctx); err != nil {
		t.Fatalf("bootstrap guest projection: %v", err)
	}
	leftRevision, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "leave", fixture.transport.Revision, t.Name()+"-leave", map[string]any{})
	if err != nil {
		t.Fatalf("public leave failed: %v", err)
	}
	if _, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "takeover-ai", leftRevision, t.Name()+"-takeover", map[string]any{}); err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("public takeover-ai error = %v, want HTTP 404 because the public gameplay gateway exposes no takeover route", err)
	}
	if controller := fixture.readController(t, fixture.viewer.Actor); controller != engine.ControllerHuman {
		t.Fatalf("controller after leave = %q, want pending human before worker", controller)
	}
	cursor, event, payload := readGameplayEvent(t, fixture.server.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 1)
	if cursor != 2 || event != engine.EventPlayerDisconnected || payload.Type != engine.EventPlayerDisconnected || payload.Sequence != 2 {
		t.Fatalf("leave SSE = cursor %d event %q payload %#v", cursor, event, payload)
	}
	if err := fixture.runtime.RunPendingTakeovers(fixture.ctx); err != nil {
		t.Fatalf("autonomous takeover worker: %v", err)
	}
	if controller := fixture.readController(t, fixture.viewer.Actor); controller != engine.ControllerAI {
		t.Fatalf("controller after worker = %q, want ai autonomous takeover after public leave", controller)
	}
	cursor, event, payload = readGameplayEvent(t, fixture.server.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 2)
	if cursor != 3 || event != engine.EventControllerChanged || payload.Type != engine.EventControllerChanged || payload.Sequence != 3 {
		t.Fatalf("takeover SSE = cursor %d event %q payload %#v", cursor, event, payload)
	}
	if _, err := fixture.runtime.Driver.Drive(fixture.ctx, fixture.setup.GameID, fixture.leftPlayerID, 3); err != nil && !errors.Is(err, ai.ErrTerminal) {
		t.Fatalf("AI driver after autonomous takeover = %v", err)
	}
	if err := fixture.runtime.RunPendingTakeovers(fixture.ctx); err != nil {
		t.Fatalf("duplicate worker cycle: %v", err)
	}
	if fixture.telemetry.Unknown != nil {
		t.Fatalf("memorysql rejected production query: %v", fixture.telemetry.Unknown)
	}
}

func TestPublicLeavePendingTakeoverSurvivesRestartUntilWorkerCompletesIt(t *testing.T) {
	fixture := newLeaveContinuityFixture(t)
	defer fixture.Close()

	if _, err := fixture.transport.CurrentProjection(fixture.ctx); err != nil {
		t.Fatalf("bootstrap guest projection: %v", err)
	}
	leftRevision, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "leave", fixture.transport.Revision, t.Name()+"-leave", map[string]any{})
	if err != nil {
		t.Fatalf("public leave failed: %v", err)
	}
	if _, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "takeover-ai", leftRevision, t.Name()+"-takeover", map[string]any{}); err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("public takeover-ai error = %v, want HTTP 404 because the public gameplay gateway exposes no takeover route", err)
	}
	if controller := fixture.readController(t, fixture.viewer.Actor); controller != engine.ControllerHuman {
		t.Fatalf("controller before restart worker = %q, want pending human", controller)
	}

	restartedDatabase, restartedTelemetry, err := memorysql.Open(fixture.kit.Namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer restartedDatabase.Close()
	restartedRuntime := onboarding.RestartPublicRuntime(sqlx.NewDb(restartedDatabase, memorysql.DriverName), fixture.sources, fixture.seed)
	restartedServer := httptest.NewServer(restartedRuntime.Handler)
	defer restartedServer.Close()
	restartedClient := onboarding.HTTPClient{BaseURL: restartedServer.URL, Client: &http.Client{}}

	if err := restartedRuntime.RunPendingTakeovers(fixture.ctx); err != nil {
		t.Fatalf("restart autonomous takeover worker: %v", err)
	}
	if controller := fixture.readControllerWithClient(t, restartedClient, fixture.viewer.Actor); controller != engine.ControllerAI {
		t.Fatalf("controller after restart worker = %q, want ai autonomous takeover after public leave", controller)
	}
	cursor, event, payload := readGameplayEvent(t, restartedServer.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 2)
	if cursor != 3 || event != engine.EventControllerChanged || payload.Type != engine.EventControllerChanged || payload.Sequence != 3 {
		t.Fatalf("restart takeover SSE = cursor %d event %q payload %#v", cursor, event, payload)
	}
	if _, err := restartedRuntime.Driver.Drive(fixture.ctx, fixture.setup.GameID, fixture.leftPlayerID, 3); err != nil && !errors.Is(err, ai.ErrTerminal) {
		t.Fatalf("restarted AI driver after autonomous takeover = %v", err)
	}
	if err := restartedRuntime.RunPendingTakeovers(fixture.ctx); err != nil {
		t.Fatalf("restart duplicate worker cycle: %v", err)
	}
	if fixture.telemetry.Unknown != nil || restartedTelemetry.Unknown != nil {
		t.Fatalf("memorysql rejected production query: first=%v restart=%v", fixture.telemetry.Unknown, restartedTelemetry.Unknown)
	}
}

func TestPublicPlanningLeaveCloseoutTransitionsToAutonomousAIWithoutDepartedActorRequest(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			fixture := newPlanningLeaveContinuityFixture(t, mode)
			defer fixture.Close()

			if _, err := fixture.transport.CurrentProjection(fixture.ctx); err != nil {
				t.Fatalf("bootstrap guest projection: %v", err)
			}
			leftRevision, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "leave", fixture.transport.Revision, t.Name()+"-leave", map[string]any{})
			if err != nil {
				t.Fatalf("public leave failed: %v", err)
			}
			if _, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "takeover-ai", leftRevision, t.Name()+"-takeover", map[string]any{}); err == nil || !strings.Contains(err.Error(), "status 404") {
				t.Fatalf("public takeover-ai error = %v, want HTTP 404 because the public gameplay gateway exposes no takeover route", err)
			}
			cursor, event, payload := readGameplayEvent(t, fixture.server.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 1)
			if cursor != 2 || event != "command.queued" || payload.Type != "command.queued" || payload.Sequence != 2 {
				t.Fatalf("queued leave SSE = cursor %d event %q payload %#v", cursor, event, payload)
			}

			survivorID := "account-" + strconv.FormatInt(fixture.viewer.AccountID, 10)
			histories := map[string][]puppet.PublicExplorationStep{}
			scripted := map[string]bool{}
			scriptProgress := map[string]int{}
			accepted, err := driveScriptedHuman(fixture.ctx, fixture.client, fixture.scenario, fixture.setup.GameID, fixture.viewer, survivorID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptThenPuppet, histories, scripted, scriptProgress)
			if err != nil {
				t.Fatalf("survivor planning action: %v", err)
			}
			if accepted != 1 {
				t.Fatalf("survivor accepted commands after leave = %d, want exactly one planning action", accepted)
			}
			cursor, event, payload = readGameplayEvent(t, fixture.server.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 2)
			if cursor != 3 || event != "command.queued" || payload.Type != "command.queued" || payload.Sequence != 3 {
				t.Fatalf("survivor action SSE = cursor %d event %q payload %#v", cursor, event, payload)
			}

			accepted, err = driveScriptedHuman(fixture.ctx, fixture.client, fixture.scenario, fixture.setup.GameID, fixture.viewer, survivorID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptThenPuppet, histories, scripted, scriptProgress)
			if err != nil {
				t.Fatalf("survivor planning closeout: %v", err)
			}
			if accepted != 1 {
				t.Fatalf("survivor closeout accepted commands = %d, want round-complete only", accepted)
			}
			cursor, event, payload = readGameplayEvent(t, fixture.server.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 3)
			if cursor != 4 || event != engine.EventPlayerDisconnected || payload.Type != engine.EventPlayerDisconnected || payload.Sequence != 4 {
				t.Fatalf("planning closeout SSE = cursor %d event %q payload %#v", cursor, event, payload)
			}
			if controller := fixture.readController(t, fixture.viewer.Actor); controller != engine.ControllerHuman {
				t.Fatalf("controller after planning closeout before worker = %q, want pending human", controller)
			}
			if err := fixture.runtime.RunPendingTakeovers(fixture.ctx); err != nil {
				t.Fatalf("autonomous takeover worker after planning closeout: %v", err)
			}
			if controller := fixture.readController(t, fixture.viewer.Actor); controller != engine.ControllerAI {
				t.Fatalf("controller after planning worker = %q, want ai autonomous takeover after public leave", controller)
			}
			cursor, event, payload = readGameplayEvent(t, fixture.server.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 4)
			if cursor != 5 || event != engine.EventControllerChanged || payload.Type != engine.EventControllerChanged || payload.Sequence != 5 {
				t.Fatalf("planning takeover SSE = cursor %d event %q payload %#v", cursor, event, payload)
			}
		})
	}
}

func TestPublicPlanningLeavePendingTakeoverSurvivesRestartUntilWorkerCompletesIt(t *testing.T) {
	for _, mode := range []string{"simultaneous", "hybrid"} {
		t.Run(mode, func(t *testing.T) {
			fixture := newPlanningLeaveContinuityFixture(t, mode)
			defer fixture.Close()

			if _, err := fixture.transport.CurrentProjection(fixture.ctx); err != nil {
				t.Fatalf("bootstrap guest projection: %v", err)
			}
			leftRevision, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "leave", fixture.transport.Revision, t.Name()+"-leave", map[string]any{})
			if err != nil {
				t.Fatalf("public leave failed: %v", err)
			}
			if _, err := fixture.client.GameplayCommand(fixture.ctx, fixture.setup.Owner, fixture.setup.GameID, "takeover-ai", leftRevision, t.Name()+"-takeover", map[string]any{}); err == nil || !strings.Contains(err.Error(), "status 404") {
				t.Fatalf("public takeover-ai error = %v, want HTTP 404 because the public gameplay gateway exposes no takeover route", err)
			}
			survivorID := "account-" + strconv.FormatInt(fixture.viewer.AccountID, 10)
			histories := map[string][]puppet.PublicExplorationStep{}
			scripted := map[string]bool{}
			scriptProgress := map[string]int{}
			if accepted, err := driveScriptedHuman(fixture.ctx, fixture.client, fixture.scenario, fixture.setup.GameID, fixture.viewer, survivorID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptThenPuppet, histories, scripted, scriptProgress); err != nil || accepted != 1 {
				t.Fatalf("survivor planning action = %d, %v", accepted, err)
			}
			if accepted, err := driveScriptedHuman(fixture.ctx, fixture.client, fixture.scenario, fixture.setup.GameID, fixture.viewer, survivorID, fixture.server.URL, puppet.TestHarnessPolicyLevelProductionFirst, nil, humanScriptThenPuppet, histories, scripted, scriptProgress); err != nil || accepted != 1 {
				t.Fatalf("survivor planning closeout = %d, %v", accepted, err)
			}
			cursor, event, payload := readGameplayEvent(t, fixture.server.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 3)
			if cursor != 4 || event != engine.EventPlayerDisconnected || payload.Type != engine.EventPlayerDisconnected || payload.Sequence != 4 {
				t.Fatalf("planning closeout SSE before restart = cursor %d event %q payload %#v", cursor, event, payload)
			}
			if controller := fixture.readController(t, fixture.viewer.Actor); controller != engine.ControllerHuman {
				t.Fatalf("controller before restart worker = %q, want pending human", controller)
			}

			restartedDatabase, restartedTelemetry, err := memorysql.Open(fixture.kit.Namespace)
			if err != nil {
				t.Fatal(err)
			}
			defer restartedDatabase.Close()
			restartedRuntime := onboarding.RestartPublicRuntime(sqlx.NewDb(restartedDatabase, memorysql.DriverName), fixture.sources, fixture.seed)
			restartedServer := httptest.NewServer(restartedRuntime.Handler)
			defer restartedServer.Close()
			restartedClient := onboarding.HTTPClient{BaseURL: restartedServer.URL, Client: &http.Client{}}

			if err := restartedRuntime.RunPendingTakeovers(fixture.ctx); err != nil {
				t.Fatalf("restart autonomous takeover worker after planning closeout: %v", err)
			}
			if controller := fixture.readControllerWithClient(t, restartedClient, fixture.viewer.Actor); controller != engine.ControllerAI {
				t.Fatalf("controller after restart planning worker = %q, want ai autonomous takeover after public leave", controller)
			}
			cursor, event, payload = readGameplayEvent(t, restartedServer.URL, fixture.setup.GameID, fixture.viewer.Actor.Session, 4)
			if cursor != 5 || event != engine.EventControllerChanged || payload.Type != engine.EventControllerChanged || payload.Sequence != 5 {
				t.Fatalf("restart planning takeover SSE = cursor %d event %q payload %#v", cursor, event, payload)
			}
			if fixture.telemetry.Unknown != nil || restartedTelemetry.Unknown != nil {
				t.Fatalf("memorysql rejected production query: first=%v restart=%v", fixture.telemetry.Unknown, restartedTelemetry.Unknown)
			}
		})
	}
}

func newPlanningLeaveContinuityFixture(t *testing.T, mode string) *leaveContinuityFixture {
	t.Helper()
	kit := testkit.NewScenario(t.Name(), int64(seedBase+0x92))
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		kit.Close()
		t.Fatal(err)
	}
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(t.Name()), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(seedBase + 0x92), nil }
	runtime := onboarding.NewPublicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	scenario := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x92,
		TurnMode:           mode,
		Slots:              slotPlan([]string{humanController, humanController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "temperate", Wetness: 50},
		Rules:              mustRulesFixture("bloodlust_conquest", ""),
		HumanPolicy:        ConquestHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            defaultAcceptanceBudgets(),
		Restart:            disabledRestart(),
	}
	setup, err := setupCampaignGame(kit.Context, client, scenario)
	if err != nil {
		server.Close()
		_ = db.Close()
		kit.Close()
		t.Fatal(err)
	}
	viewer := setup.Players[1]
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: setup.GameID, SessionID: viewer.Actor.Session, Namespace: t.Name()}
	return &leaveContinuityFixture{
		ctx:          kit.Context,
		kit:          kit,
		database:     db,
		telemetry:    telemetry,
		runtime:      runtime,
		server:       server,
		client:       client,
		setup:        setup,
		viewer:       viewer,
		transport:    transport,
		leftPlayerID: "account-" + strconv.FormatInt(setup.Players[0].AccountID, 10),
		scenario:     scenario,
		sources:      sources,
		seed:         seed,
	}
}

type leaveContinuityFixture struct {
	ctx          context.Context
	kit          *testkit.Scenario
	database     *sqlx.DB
	telemetry    *memorysql.Telemetry
	runtime      *onboarding.PublicRuntime
	server       *httptest.Server
	client       onboarding.HTTPClient
	setup        publicGameSetup
	viewer       publicSetupPlayer
	transport    *HTTPExplorationTransport
	leftPlayerID string
	scenario     AcceptanceConfig
	sources      engine.Sources
	seed         func() (int64, error)
}

func newLeaveContinuityFixture(t *testing.T) *leaveContinuityFixture {
	t.Helper()
	kit := testkit.NewScenario(t.Name(), int64(seedBase+0x91))
	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		kit.Close()
		t.Fatal(err)
	}
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := engine.Sources{Clock: kit.Clock, IDs: testkit.NewIDSource(t.Name()), RNG: executorRNG{}}
	seed := func() (int64, error) { return int64(seedBase + 0x91), nil }
	runtime := onboarding.NewPublicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	scenario := AcceptanceConfig{
		Name:               t.Name(),
		Seed:               seedBase + 0x91,
		TurnMode:           "sequential",
		Slots:              slotPlan([]string{humanController, humanController}),
		World:              AcceptanceWorldFixture{MapSize: "duel", Climate: "temperate", Wetness: 50},
		Rules:              mustRulesFixture("bloodlust_conquest", "scripts/military-zoc-combat.json"),
		HumanPolicy:        ConquestHumanPolicy{},
		AutonomousAIPolicy: ProductionAutonomousAI,
		Budgets:            defaultAcceptanceBudgets(),
		Restart:            disabledRestart(),
	}
	setup, err := setupCampaignGame(kit.Context, client, scenario)
	if err != nil {
		server.Close()
		_ = db.Close()
		kit.Close()
		t.Fatal(err)
	}
	viewer := setup.Players[1]
	transport := &HTTPExplorationTransport{BaseURL: server.URL, Client: client.Client, GameID: setup.GameID, SessionID: viewer.Actor.Session, Namespace: t.Name()}
	return &leaveContinuityFixture{
		ctx:          kit.Context,
		kit:          kit,
		database:     db,
		telemetry:    telemetry,
		runtime:      runtime,
		server:       server,
		client:       client,
		setup:        setup,
		viewer:       viewer,
		transport:    transport,
		leftPlayerID: "account-" + strconv.FormatInt(setup.Players[0].AccountID, 10),
		scenario:     scenario,
		sources:      sources,
		seed:         seed,
	}
}

func (f *leaveContinuityFixture) Close() {
	if f == nil {
		return
	}
	if f.server != nil {
		f.server.Close()
	}
	if f.database != nil {
		_ = f.database.Close()
	}
	if f.kit != nil {
		memorysql.Reset(f.kit.Namespace)
		f.kit.Close()
	}
}

func (f *leaveContinuityFixture) readController(t *testing.T, actor onboarding.Actor) string {
	t.Helper()
	return f.readControllerWithClient(t, f.client, actor)
}

func (f *leaveContinuityFixture) readControllerWithClient(t *testing.T, client onboarding.HTTPClient, actor onboarding.Actor) string {
	t.Helper()
	gameDTO, err := client.ReadGame(f.ctx, actor, f.setup.GameID)
	if err != nil {
		t.Fatal(err)
	}
	for _, player := range gameDTO.Players {
		if player.AccountID == f.setup.Players[0].AccountID {
			return player.Controller
		}
	}
	t.Fatalf("game %d does not expose player %d controller", f.setup.GameID, f.setup.Players[0].AccountID)
	return ""
}

type gameplayEventPayload struct {
	Type     string `json:"type"`
	Sequence int64  `json:"sequence"`
}

func readGameplayEvent(t *testing.T, baseURL string, gameID int64, sessionID int64, cursor int64) (int64, string, gameplayEventPayload) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/game/%d/events?session_id=%d", baseURL, gameID, sessionID), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Last-Event-ID", strconv.FormatInt(cursor, 10))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("public SSE status = %d", response.StatusCode)
	}
	var (
		id      int64
		event   string
		payload gameplayEventPayload
	)
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "id: "):
			id, err = strconv.ParseInt(strings.TrimPrefix(line, "id: "), 10, 64)
			if err != nil {
				t.Fatal(err)
			}
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &payload); err != nil {
				t.Fatal(err)
			}
		case line == "" && id > cursor:
			return id, event, payload
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	t.Fatalf("public SSE after cursor %d did not yield a committed event", cursor)
	return 0, "", gameplayEventPayload{}
}
