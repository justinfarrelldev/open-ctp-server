package onboarding

import (
	"bufio"
	"bytes"
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
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type deterministicRNG struct{}

func (deterministicRNG) Next(_ string, state uint64) (uint64, uint64) {
	next := state + 0x9e3779b97f4a7c15
	value := next
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return next, value ^ (value >> 31)
}

func deterministicSources(t *testing.T) engine.Sources {
	kit := testkit.NewScenario(t.Name(), 1)
	t.Cleanup(kit.Close)
	return engine.Sources{Clock: kit.Clock, IDs: kit.IDs, RNG: deterministicRNG{}}
}

func TestPublicRuntimeStartsOneHumanSevenServerManagedAI(t *testing.T) {
	database, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := deterministicSources(t)
	seed := func() (int64, error) { return 808, nil }
	runtime := NewPublicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	ctx := context.Background()

	owner, err := client.Signup(ctx, "local-human", "local-human@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(ctx, owner, 0, "localhost singleplayer")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(ctx, owner, gameID, GameSetup{Style: "sequential", MapSize: "medium", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(ctx, owner, gameID, "america"); err != nil {
		t.Fatal(err)
	}
	for index, country := range []string{"aztec", "babylon", "canadian", "chinese", "egyptian", "english", "french"} {
		if err := client.AddManagedAI(ctx, owner, gameID, country, fmt.Sprintf("Local AI %d", index+1)); err != nil {
			t.Fatalf("add managed AI %s: %v", country, err)
		}
	}
	if err := client.StartGame(ctx, owner, gameID); err != nil {
		t.Fatal(err)
	}
	started, err := client.ReadGame(ctx, owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if len(started.Players) != 8 {
		t.Fatalf("started players = %d, want 8: %+v", len(started.Players), started.Players)
	}
	aiPlayers := 0
	for _, player := range started.Players {
		if player.Controller == engine.ControllerAI {
			aiPlayers++
		}
	}
	if aiPlayers != 7 {
		t.Fatalf("AI players = %d, want 7: %+v", aiPlayers, started.Players)
	}
	var raw string
	if err := db.GetContext(ctx, &raw, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, gameID); err != nil {
		t.Fatal(err)
	}
	state, err := engine.DecodeCanonicalStateJSON([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Players) != 8 || len(state.Units) != 8 {
		t.Fatalf("canonical players=%d units=%d, want 8/8", len(state.Players), len(state.Units))
	}
	canonicalAI := 0
	for _, player := range state.Players {
		if player.Controller == engine.ControllerAI {
			canonicalAI++
		}
	}
	if canonicalAI != 7 {
		t.Fatalf("canonical AI players = %d, want 7: %+v", canonicalAI, state.Players)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unknown SQL = %v", telemetry.Unknown)
	}
}

func TestPublicRuntimeAutonomousAIWorkerDrivesActiveMultiplayerAI(t *testing.T) {
	database, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := deterministicSources(t)
	seed := func() (int64, error) { return 22, nil }
	runtime := NewPublicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	ctx := context.Background()

	owner, err := client.Signup(ctx, "worker-human", "worker-human@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(ctx, owner, 0, "worker multiplayer")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(ctx, owner, gameID, GameSetup{Style: "sequential", MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(ctx, owner, gameID, "america"); err != nil {
		t.Fatal(err)
	}
	if err := client.AddManagedAI(ctx, owner, gameID, "aztec", "Worker AI"); err != nil {
		t.Fatal(err)
	}
	if err := client.StartGame(ctx, owner, gameID); err != nil {
		t.Fatal(err)
	}
	if revision, err := client.GameplayCommand(ctx, owner, gameID, "pass", 1, "worker-human-pass", map[string]any{}); err != nil || revision != 2 {
		t.Fatalf("human pass revision=%d err=%v", revision, err)
	}
	if err := runtime.RunAutonomousAI(ctx); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := db.GetContext(ctx, &raw, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, gameID); err != nil {
		t.Fatal(err)
	}
	state, err := engine.DecodeCanonicalStateJSON([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if state.Revision < 4 {
		t.Fatalf("revision after autonomous AI = %d, want at least 4", state.Revision)
	}
	foundAICity := false
	for _, city := range state.Cities {
		if strings.HasPrefix(city.OwnerID, "account-") && city.OwnerID != "account-1" {
			foundAICity = true
		}
	}
	if !foundAICity {
		t.Fatalf("autonomous AI did not found a city: cities=%+v players=%+v", state.Cities, state.Players)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unknown SQL = %v", telemetry.Unknown)
	}
}

func TestPublicRuntimeUsesProductionAccountHandlerWithNamedMemoryStore(t *testing.T) {
	db, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	runtime := NewPublicRuntime(sqlx.NewDb(db, memorysql.DriverName), deterministicSources(t), func() (int64, error) { return 1, nil })
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	actor, err := client.Signup(context.Background(), "owner", "owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if actor.Session == 0 {
		t.Fatal("production account handler did not issue a session")
	}
	if telemetry.Unknown != nil {
		t.Fatalf("memory driver accepted no unknown SQL: %v", telemetry.Unknown)
	}
}

func TestPublicRuntimeCreatesLobbyWithNamedMemoryStore(t *testing.T) {
	db, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	runtime := NewPublicRuntime(sqlx.NewDb(db, memorysql.DriverName), deterministicSources(t), func() (int64, error) { return 1, nil })
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	actor, err := client.Signup(context.Background(), "owner", "owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	lobbyID, err := client.CreateLobby(context.Background(), actor, "public-e2e")
	if err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	guest, err := client.Signup(context.Background(), "guest", "guest@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.JoinLobby(context.Background(), guest, lobbyID); err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	closeStream, err := client.OpenLobbyStream(context.Background(), actor, lobbyID)
	if err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	defer closeStream()
	gameID, err := client.CreateGame(context.Background(), actor, lobbyID, "public-e2e-game")
	if err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	wetness := 50
	if err := client.ConfigureGame(context.Background(), actor, gameID, GameSetup{Style: "sequential", MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	if err := client.SelectCountry(context.Background(), actor, gameID, "america"); err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	if err := client.SelectCountry(context.Background(), guest, gameID, "aztec"); err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	if err := client.StartGame(context.Background(), actor, gameID); err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	if _, err := client.ReadProjectedWorld(context.Background(), actor, gameID); err != nil {
		t.Fatalf("%v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	if err := client.StartGame(context.Background(), actor, gameID); !errors.Is(err, ErrConflict) {
		t.Fatalf("repeated start error = %v, want conflict; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	restartedDB, _, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer restartedDB.Close()
	restartedRuntime := NewPublicRuntime(sqlx.NewDb(restartedDB, memorysql.DriverName), deterministicSources(t), func() (int64, error) { return 1, nil })
	restartedServer := httptest.NewServer(restartedRuntime.Handler)
	defer restartedServer.Close()
	restartedClient := HTTPClient{BaseURL: restartedServer.URL, Client: &http.Client{}}
	world, err := restartedClient.ReadProjectedWorld(context.Background(), actor, gameID)
	if err != nil || len(world) == 0 {
		t.Fatalf("restart read projected world: %v; driver: %v; queries: %q", err, telemetry.Unknown, telemetry.Queries)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("memory driver accepted no unknown SQL: %v", telemetry.Unknown)
	}
}

func TestPublicRuntimeExecutesGameplayCommandAndResumesSSEFromNamedMemoryStore(t *testing.T) {
	db, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	runtime := NewPublicRuntime(sqlx.NewDb(db, memorysql.DriverName), deterministicSources(t), func() (int64, error) { return 1, nil })
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	owner, err := client.Signup(context.Background(), "command-owner", "command-owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	lobbyID, err := client.CreateLobby(context.Background(), owner, "command-lobby")
	if err != nil {
		t.Fatal(err)
	}
	guest, err := client.Signup(context.Background(), "command-guest", "command-guest@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.JoinLobby(context.Background(), guest, lobbyID); err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(context.Background(), owner, lobbyID, "command-game")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(context.Background(), owner, gameID, GameSetup{Style: "sequential", MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(context.Background(), owner, gameID, "america"); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(context.Background(), guest, gameID, "aztec"); err != nil {
		t.Fatal(err)
	}
	if err := client.StartGame(context.Background(), owner, gameID); err != nil {
		t.Fatal(err)
	}
	body := []byte(fmt.Sprintf(`{"session_id":%d,"expected_revision":1,"idempotency_key":"pass-1","command_id":"pass-1"}`, owner.Session))
	response, err := http.Post(fmt.Sprintf("%s/game/%d/pass", server.URL, gameID), "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("public pass status = %d; unknown SQL = %v", response.StatusCode, telemetry.Unknown)
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/game/%d/events?session_id=%d", server.URL, gameID, owner.Session), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Last-Event-ID", "1")
	stream, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Body.Close()
	if stream.StatusCode != http.StatusOK {
		t.Fatalf("SSE status = %d; unknown SQL = %v", stream.StatusCode, telemetry.Unknown)
	}
	scanner := bufio.NewScanner(stream.Body)
	foundPass := false
	for scanner.Scan() {
		if scanner.Text() == "id: 2" {
			foundPass = true
			break
		}
	}
	if err := scanner.Err(); err != nil || !foundPass || telemetry.Unknown != nil {
		t.Fatalf("SSE resume pass=%t err=%v unknown=%v", foundPass, err, telemetry.Unknown)
	}
}

func TestPublicRuntimeDispatchesDueLongTurnPassToSSE(t *testing.T) {
	db, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	kit := testkit.NewScenario(t.Name(), 1)
	t.Cleanup(kit.Close)
	clock := testkit.NewFakeClock(time.Unix(1, 0))
	runtime := NewPublicRuntime(sqlx.NewDb(db, memorysql.DriverName), engine.Sources{Clock: clock, IDs: kit.IDs, RNG: deterministicRNG{}}, func() (int64, error) { return 1, nil })
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	owner, err := client.Signup(context.Background(), "deadline-owner", "deadline-owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	lobbyID, err := client.CreateLobby(context.Background(), owner, "deadline-lobby")
	if err != nil {
		t.Fatal(err)
	}
	guest, err := client.Signup(context.Background(), "deadline-guest", "deadline-guest@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.JoinLobby(context.Background(), guest, lobbyID); err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(context.Background(), owner, lobbyID, "deadline-game")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(context.Background(), owner, gameID, GameSetup{Style: "long_turn", MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(context.Background(), owner, gameID, "america"); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(context.Background(), guest, gameID, "aztec"); err != nil {
		t.Fatal(err)
	}
	if err := client.StartGame(context.Background(), owner, gameID); err != nil {
		t.Fatal(err)
	}
	clock.Advance(time.Second)
	if err := runtime.RunDueDeadlines(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RunDueDeadlines(context.Background()); err != nil {
		t.Fatal(err)
	}
	streamContext, cancelStream := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelStream()
	request, err := http.NewRequestWithContext(streamContext, http.MethodGet, fmt.Sprintf("%s/game/%d/events?session_id=%d", server.URL, gameID, owner.Session), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Last-Event-ID", "1")
	stream, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Body.Close()
	if stream.StatusCode != http.StatusOK {
		t.Fatalf("SSE status = %d; unknown SQL = %v", stream.StatusCode, telemetry.Unknown)
	}
	scanner := bufio.NewScanner(stream.Body)
	if !scanner.Scan() || scanner.Text() != "id: 2" {
		t.Fatalf("deadline SSE event = %q; unknown SQL = %v", scanner.Text(), telemetry.Unknown)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unknown SQL = %v", telemetry.Unknown)
	}
}

func TestPublicRuntimeResolvesSimultaneousQueueAfterRestart(t *testing.T) {
	db, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	sources := deterministicSources(t)
	seed := func() (int64, error) { return 1, nil }
	runtime := NewPublicRuntime(sqlx.NewDb(db, memorysql.DriverName), sources, seed)
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	owner, err := client.Signup(context.Background(), "simultaneous-owner", "simultaneous-owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	lobbyID, err := client.CreateLobby(context.Background(), owner, "simultaneous-lobby")
	if err != nil {
		t.Fatal(err)
	}
	guest, err := client.Signup(context.Background(), "simultaneous-guest", "simultaneous-guest@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.JoinLobby(context.Background(), guest, lobbyID); err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(context.Background(), owner, lobbyID, "simultaneous-game")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(context.Background(), owner, gameID, GameSetup{Style: "hybrid", MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	for _, selection := range []struct {
		actor   Actor
		country string
	}{{owner, "america"}, {guest, "aztec"}} {
		if err := client.SelectCountry(context.Background(), selection.actor, gameID, selection.country); err != nil {
			t.Fatal(err)
		}
	}
	if err := client.StartGame(context.Background(), owner, gameID); err != nil {
		t.Fatal(err)
	}
	outsider, err := client.Signup(context.Background(), "hybrid-outsider", "hybrid-outsider@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GameplayCommand(context.Background(), outsider, gameID, "set-war", 1, "outsider-war", map[string]any{"target_player_id": "account-1", "at_war": true}); err == nil {
		t.Fatal("nonmember set-war accepted")
	}
	gameDTO, err := client.ReadGame(context.Background(), owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	playerID := func(country string) string {
		for _, player := range gameDTO.Players {
			if player.Country == country {
				return "account-" + strconv.FormatInt(player.AccountID, 10)
			}
		}
		t.Fatalf("public game has no %s player", country)
		return ""
	}
	ownerPlayer, guestPlayer := playerID("america"), playerID("aztec")
	if revision, err := client.GameplayCommand(context.Background(), owner, gameID, "set-war", 1, "war-on", map[string]any{"target_player_id": guestPlayer, "at_war": true}); err != nil || revision != 2 {
		t.Fatalf("set war revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), owner, gameID, "set-war", 1, "war-on", map[string]any{"target_player_id": guestPlayer, "at_war": true}); err != nil || revision != 2 {
		t.Fatalf("set war replay revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), guest, gameID, "set-war", 2, "war-off", map[string]any{"target_player_id": ownerPlayer, "at_war": false}); err != nil || revision != 3 {
		t.Fatalf("clear war revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), owner, gameID, "set-war", 3, "war-on-again", map[string]any{"target_player_id": guestPlayer, "at_war": true}); err != nil || revision != 4 {
		t.Fatalf("restore war revision=%d err=%v", revision, err)
	}
	ownerUnit := simultaneousVisibleUnit(t, server.URL, gameID, owner.Session, ownerPlayer)
	guestUnit := simultaneousVisibleUnit(t, server.URL, gameID, guest.Session, guestPlayer)
	if revision, err := client.GameplayCommand(context.Background(), owner, gameID, "army-create", 4, "queue-owner", map[string]any{"army_id": "owner-army", "unit_ids": []string{ownerUnit}}); err != nil || revision != 5 {
		t.Fatalf("owner queue revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), guest, gameID, "army-create", 5, "queue-guest", map[string]any{"army_id": "guest-army", "unit_ids": []string{guestUnit}}); err != nil || revision != 6 {
		t.Fatalf("guest queue revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), owner, gameID, "round-complete", 6, "early-owner", map[string]any{}); err != nil || revision != 7 {
		t.Fatalf("owner early end revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), guest, gameID, "round-complete", 7, "early-guest", map[string]any{}); err != nil || revision != 8 {
		t.Fatalf("hybrid planning resolution revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), owner, gameID, "pass", 8, "conflict-owner", map[string]any{}); err != nil || revision != 9 {
		t.Fatalf("owner conflict pass revision=%d err=%v", revision, err)
	}
	if revision, err := client.GameplayCommand(context.Background(), guest, gameID, "pass", 9, "conflict-guest", map[string]any{}); err != nil || revision != 10 {
		t.Fatalf("guest conflict pass revision=%d err=%v", revision, err)
	}
	server.Close()
	runtime = RestartPublicRuntime(sqlx.NewDb(db, memorysql.DriverName), sources, seed)
	server = httptest.NewServer(runtime.Handler)
	client.BaseURL = server.URL
	streamContext, cancelStream := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelStream()
	request, err := http.NewRequestWithContext(streamContext, http.MethodGet, fmt.Sprintf("%s/game/%d/events?session_id=%d", server.URL, gameID, owner.Session), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Last-Event-ID", "9")
	stream, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Body.Close()
	scanner := bufio.NewScanner(stream.Body)
	if !scanner.Scan() || scanner.Text() != "id: 10" {
		t.Fatalf("resolved SSE event = %q; unknown SQL = %v", scanner.Text(), telemetry.Unknown)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unknown SQL = %v", telemetry.Unknown)
	}
}

func TestHybridQueuedAICloseoutSurvivesRuntimeReconstruction(t *testing.T) {
	database, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := deterministicSources(t)
	seed := func() (int64, error) { return 1, nil }
	runtime := NewPublicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	ctx := context.Background()

	owner, err := client.Signup(ctx, "queued-owner", "queued-owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	guest, err := client.Signup(ctx, "queued-guest", "queued-guest@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	lobbyID, err := client.CreateLobby(ctx, owner, "queued-lobby")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.JoinLobby(ctx, guest, lobbyID); err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(ctx, owner, lobbyID, "queued-hybrid")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(ctx, owner, gameID, GameSetup{Style: "hybrid", MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	for _, selection := range []struct {
		actor   Actor
		country string
	}{{owner, "america"}, {guest, "aztec"}} {
		if err := client.SelectCountry(ctx, selection.actor, gameID, selection.country); err != nil {
			t.Fatal(err)
		}
	}
	gameDTO, err := client.ReadGame(ctx, owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	aiPlayerID := ""
	for _, player := range gameDTO.Players {
		if player.Country == "aztec" {
			aiPlayerID = "account-" + strconv.FormatInt(player.AccountID, 10)
			if err := client.AddAI(ctx, owner, gameID, player.AccountID); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if aiPlayerID == "" {
		t.Fatal("queued hybrid game has no aztec player")
	}
	if err := client.StartGame(ctx, owner, gameID); err != nil {
		t.Fatal(err)
	}
	result, err := runtime.Driver.Drive(ctx, gameID, aiPlayerID, 1)
	if err != nil {
		t.Fatalf("initial AI queued action: %v", err)
	}
	if result.Commands != 2 || result.Revision != 3 {
		t.Fatalf("initial queued AI result=%+v", result)
	}
	projection := simultaneousProjection(t, server.URL, gameID, guest.Session)
	if projection.PlanningStatus != visibility.PlanningStatusClosed {
		t.Fatalf("post-action planning status = %q, want closed", projection.PlanningStatus)
	}
	server.Close()
	runtime = RestartPublicRuntime(db, sources, seed)
	server = httptest.NewServer(runtime.Handler)
	client.BaseURL = server.URL
	restartedProjection := simultaneousProjection(t, server.URL, gameID, guest.Session)
	if restartedProjection.PlanningStatus != visibility.PlanningStatusClosed {
		t.Fatalf("restarted planning status = %q, want closed", restartedProjection.PlanningStatus)
	}
	result, err = runtime.Driver.Drive(ctx, gameID, aiPlayerID, 3)
	if err != nil {
		t.Fatalf("restarted closed AI drive: %v", err)
	}
	if result.Commands != 0 || result.Revision != 3 {
		t.Fatalf("restarted closed result=%+v", result)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unknown SQL = %v", telemetry.Unknown)
	}
}

func TestPublicRuntimePersistsThreePlayerOwnedUnitsAcrossRestart(t *testing.T) {
	database, telemetry, err := memorysql.Open(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := deterministicSources(t)
	seed := func() (int64, error) { return 3, nil }
	runtime := NewPublicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	client := HTTPClient{BaseURL: server.URL, Client: &http.Client{}}
	ctx := context.Background()

	players := make([]struct {
		actor   Actor
		country string
		player  string
	}, 3)
	for index, country := range []string{"america", "aztec", "babylon"} {
		actor, err := client.Signup(ctx, fmt.Sprintf("three-player-%d", index+1), fmt.Sprintf("three-player-%d@example.invalid", index+1))
		if err != nil {
			t.Fatal(err)
		}
		players[index].actor, players[index].country = actor, country
	}
	lobbyID, err := client.CreateLobby(ctx, players[0].actor, "three-player-lobby")
	if err != nil {
		t.Fatal(err)
	}
	for _, player := range players[1:] {
		if err := client.JoinLobby(ctx, player.actor, lobbyID); err != nil {
			t.Fatal(err)
		}
	}
	gameID, err := client.CreateGame(ctx, players[0].actor, lobbyID, "three-player-long-turn")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(ctx, players[0].actor, gameID, GameSetup{Style: "long_turn", MapSize: "small", Climate: "cold", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	for _, player := range players {
		if err := client.SelectCountry(ctx, player.actor, gameID, player.country); err != nil {
			t.Fatal(err)
		}
	}
	gameDTO, err := client.ReadGame(ctx, players[0].actor, gameID)
	if err != nil {
		t.Fatal(err)
	}
	for index := range players {
		for _, gamePlayer := range gameDTO.Players {
			if gamePlayer.Country == players[index].country {
				players[index].player = "account-" + strconv.FormatInt(gamePlayer.AccountID, 10)
			}
		}
		if players[index].player == "" {
			t.Fatalf("missing game player for %s", players[index].country)
		}
		if index > 0 {
			accountID, _ := strconv.ParseInt(strings.TrimPrefix(players[index].player, "account-"), 10, 64)
			if err := client.AddAI(ctx, players[0].actor, gameID, accountID); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := client.StartGame(ctx, players[0].actor, gameID); err != nil {
		t.Fatal(err)
	}

	assertOwnedUnits := func() {
		var raw string
		if err := db.GetContext(ctx, &raw, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, gameID); err != nil {
			t.Fatal(err)
		}
		state, err := engine.DecodeCanonicalStateJSON([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		for _, player := range players {
			owned := 0
			for _, unit := range state.Units {
				if unit.OwnerID == player.player {
					owned++
				}
			}
			if owned == 0 {
				owners := make([]string, 0, len(state.Units))
				for _, unit := range state.Units {
					owners = append(owners, unit.OwnerID)
				}
				cityOwners := make([]string, 0, len(state.Cities))
				for _, city := range state.Cities {
					cityOwners = append(cityOwners, city.OwnerID)
				}
				t.Fatalf("canonical state has no units for %s; players=%+v unit_owners=%v city_owners=%v", player.player, state.Players, owners, cityOwners)
			}
			projection, err := visibility.Project(state, player.player)
			if err != nil {
				t.Fatal(err)
			}
			visibleOwned := 0
			for _, unit := range projection.Units {
				if unit.OwnerID == player.player {
					visibleOwned++
				}
			}
			if visibleOwned == 0 {
				t.Fatalf("projection has no owned units for %s", player.player)
			}
		}
	}
	assertOwnedUnits()
	server.Close()
	runtime = RestartPublicRuntime(db, sources, seed)
	server = httptest.NewServer(runtime.Handler)
	client.BaseURL = server.URL
	assertOwnedUnits()
	if revision, err := client.GameplayCommand(ctx, players[0].actor, gameID, "pass", 1, "three-player-owner-pass", map[string]any{}); err != nil || revision != 2 {
		t.Fatalf("owner pass revision=%d err=%v", revision, err)
	}
	result, err := runtime.Driver.Drive(ctx, gameID, players[1].player, 2)
	if err != nil {
		t.Fatalf("restarted AI command: %v", err)
	}
	if result.Commands != 2 || result.Revision != 4 {
		t.Fatalf("restarted AI result=%+v", result)
	}
	var raw string
	if err := db.GetContext(ctx, &raw, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, gameID); err != nil {
		t.Fatal(err)
	}
	state, err := engine.DecodeCanonicalStateJSON([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	foundCity := false
	for _, city := range state.Cities {
		foundCity = foundCity || city.OwnerID == players[1].player
	}
	if !foundCity {
		t.Fatalf("AI command did not found a city for %s", players[1].player)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unknown SQL = %v", telemetry.Unknown)
	}
}

func simultaneousProjection(t *testing.T, baseURL string, gameID, sessionID int64) visibility.Projection {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/game/%d/events?session_id=%d", baseURL, gameID, sessionID), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Last-Event-ID", "0")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var event, data string
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case line == "" && event == "resync":
			var frame struct {
				Snapshot json.RawMessage `json:"snapshot"`
			}
			var projection visibility.Projection
			if err := json.Unmarshal([]byte(data), &frame); err != nil || json.Unmarshal(frame.Snapshot, &projection) != nil {
				t.Fatalf("public simultaneous resync is invalid: %v", err)
			}
			return projection
		}
	}
	t.Fatalf("public simultaneous resync did not provide a projection: %v", scanner.Err())
	return visibility.Projection{}
}

func simultaneousVisibleUnit(t *testing.T, baseURL string, gameID, sessionID int64, playerID string) string {
	t.Helper()
	projection := simultaneousProjection(t, baseURL, gameID, sessionID)
	for _, unit := range projection.Units {
		if unit.OwnerID == playerID {
			return unit.ID
		}
	}
	t.Fatalf("public simultaneous resync has no unit for %s", playerID)
	return ""
}
