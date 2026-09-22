package memorysql_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/onboarding"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

type deterministicRNG struct{}

func (deterministicRNG) Next(_ string, state uint64) (uint64, uint64) {
	next := state + 0x9e3779b97f4a7c15
	value := next
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return next, value ^ (value >> 31)
}

func TestProductionRuntimePersistsPublicJourneyAndReceiptReplay(t *testing.T) {
	kit := testkit.NewScenario(t.Name(), 707)
	defer kit.Close()
	defer memorysql.Reset(kit.Namespace)

	database, telemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	sources := engine.Sources{Clock: kit.Clock, IDs: kit.IDs, RNG: deterministicRNG{}}
	publicRuntime := onboarding.NewPublicRuntime(sqlx.NewDb(database, memorysql.DriverName), sources, func() (int64, error) { return 707, nil })
	server := httptest.NewServer(publicRuntime.Handler)
	defer server.Close()
	client := onboarding.HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	ctx := kit.Context

	owner, err := client.Signup(ctx, "coverage-owner", "coverage-owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	guest, err := client.Signup(ctx, "coverage-guest", "coverage-guest@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	lobbyID, err := client.CreateLobby(ctx, owner, "coverage-lobby")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.JoinLobby(ctx, guest, lobbyID); err != nil {
		t.Fatal(err)
	}
	closeStream, err := client.OpenLobbyStream(ctx, owner, lobbyID)
	if err != nil {
		t.Fatal(err)
	}
	if err := closeStream(); err != nil {
		t.Fatal(err)
	}

	gameID, err := client.CreateGame(ctx, owner, lobbyID, "coverage-game")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(ctx, owner, gameID, onboarding.GameSetup{Style: "sequential", RulesJSON: `[]`, MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateStartingSettlers(ctx, owner, gameID, 2); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(ctx, owner, gameID, "america"); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(ctx, guest, gameID, "aztec"); err != nil {
		t.Fatal(err)
	}
	setupGame, err := client.ReadGame(ctx, owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if setupGame.OwnerAccountID == nil {
		t.Fatal("setup game has no owner account")
	}
	var guestAccountID int64
	for _, player := range setupGame.Players {
		if player.AccountID != *setupGame.OwnerAccountID {
			guestAccountID = player.AccountID
			break
		}
	}
	if guestAccountID == 0 {
		t.Fatalf("guest account missing from setup roster: %#v", setupGame.Players)
	}
	if err := client.AddAI(ctx, owner, gameID, guestAccountID); err != nil {
		t.Fatal(err)
	}
	if err := client.StartGame(ctx, owner, gameID); err != nil {
		t.Fatal(err)
	}
	started, err := client.ReadGame(ctx, owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if started.ID != gameID || started.Status != "started" || len(started.Players) != 2 || started.StartingSettlers != 2 {
		t.Fatalf("unexpected started game: %#v", started)
	}
	if started.RulesJSON != "[]" || started.Players[1].Controller != engine.ControllerAI {
		t.Fatalf("rules/controller did not persist at start: rules=%q players=%#v", started.RulesJSON, started.Players)
	}
	world, err := client.ReadProjectedWorld(ctx, owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if len(world) == 0 {
		t.Fatal("projected world response is empty")
	}

	revision, err := client.GameplayCommand(ctx, owner, gameID, "pass", 1, "coverage-pass", nil)
	if err != nil {
		t.Fatal(err)
	}
	if revision != 2 {
		t.Fatalf("pass revision = %d, want 2", revision)
	}
	replayedRevision, err := client.GameplayCommand(ctx, owner, gameID, "pass", 1, "coverage-pass", nil)
	if err != nil {
		t.Fatal(err)
	}
	if replayedRevision != revision {
		t.Fatalf("replayed revision = %d, want %d", replayedRevision, revision)
	}

	restartedDatabase, restartedTelemetry, err := memorysql.Open(kit.Namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer restartedDatabase.Close()
	restartedRuntime := onboarding.RestartPublicRuntime(sqlx.NewDb(restartedDatabase, memorysql.DriverName), sources, func() (int64, error) { return 707, nil })
	restartedServer := httptest.NewServer(restartedRuntime.Handler)
	defer restartedServer.Close()
	restartedClient := onboarding.HTTPClient{BaseURL: restartedServer.URL, Client: &http.Client{}}
	restartedGame, err := restartedClient.ReadGame(context.Background(), owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if restartedGame.ID != gameID || restartedGame.Status != "started" || len(restartedGame.Players) != 2 {
		t.Fatalf("unexpected restarted game: %#v", restartedGame)
	}
	if restartedGame.RulesJSON != "[]" || restartedGame.Players[1].Controller != engine.ControllerAI {
		t.Fatalf("rules/controller did not survive restart: rules=%q players=%#v", restartedGame.RulesJSON, restartedGame.Players)
	}
	if _, err := restartedClient.ReadProjectedWorld(context.Background(), owner, gameID); err != nil {
		t.Fatal(err)
	}
	if telemetry.Unknown != nil || restartedTelemetry.Unknown != nil {
		t.Fatalf("production journey used unknown SQL: first=%v restart=%v", telemetry.Unknown, restartedTelemetry.Unknown)
	}
}
