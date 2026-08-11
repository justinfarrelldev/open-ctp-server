package onboarding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	e2esetup "github.com/justinfarrelldev/open-ctp-server/internal/e2e/setup"
	"github.com/justinfarrelldev/open-ctp-server/internal/game"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	serverruntime "github.com/justinfarrelldev/open-ctp-server/internal/runtime"
)

func TestPublicOrderedSetupStartsAndRestartsForEveryCount(t *testing.T) {
	for count := 2; count <= 8; count++ {
		t.Run(fmt.Sprintf("slots-%d", count), func(t *testing.T) {
			controllers := make([]string, count)
			for index := range controllers {
				controllers[index] = engine.ControllerHuman
				if index%2 == 1 {
					controllers[index] = engine.ControllerAI
				}
			}
			testPublicOrderedSetup(t, orderedSlots(controllers), 8)
		})
	}
}

func TestPublicOrderedSetupPersistsEveryImportedPlayableCivilization(t *testing.T) {
	const importedCount = 10
	for index := 1; index <= importedCount; index++ {
		t.Run(syntheticCivilizationID(index), func(t *testing.T) {
			other := index%importedCount + 1
			slots := []e2esetup.Slot{
				{Slot: 1, CountryRecordID: syntheticCivilizationID(index), ControllerKind: engine.ControllerHuman},
				{Slot: 2, CountryRecordID: syntheticCivilizationID(other), ControllerKind: engine.ControllerAI},
			}
			testPublicOrderedSetup(t, slots, importedCount)
		})
	}
}

func testPublicOrderedSetup(t *testing.T, slots []e2esetup.Slot, importedCount int) {
	t.Helper()
	if err := e2esetup.Validate(importedPlayableRuleset(t, importedCount), slots); err != nil {
		t.Fatal(err)
	}
	namespace := t.Name()
	memorysql.Reset(namespace)
	defer memorysql.Reset(namespace)
	database, telemetry, err := memorysql.Open(namespace)
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := orderedSetupSources(t.Name())
	seed := func() (int64, error) { return int64(len(slots)), nil }
	runtime := publicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	client := slotHTTPClient{baseURL: server.URL, client: server.Client()}
	ctx := context.Background()

	players := make([]slotActor, len(slots))
	for index, slot := range slots {
		actor, err := client.signup(ctx, fmt.Sprintf("slot-%d", slot.Slot), fmt.Sprintf("slot-%d@example.invalid", slot.Slot))
		if err != nil {
			t.Fatal(err)
		}
		players[index] = slotActor{session: actor.session, country: slot.CountryRecordID, controller: slot.ControllerKind}
	}
	lobbyID, err := client.createLobby(ctx, players[0].session, "ordered-slots")
	if err != nil {
		t.Fatal(err)
	}
	for _, player := range players[1:] {
		if err := client.post(ctx, "/lobby/join_lobby", map[string]any{"session_id": player.session, "lobby_id": lobbyID}, nil); err != nil {
			t.Fatal(err)
		}
	}
	gameID, err := client.createGame(ctx, players[0].session, lobbyID)
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.post(ctx, "/game/update_style", map[string]any{"session_id": players[0].session, "game_id": gameID, "style": "sequential", "map_size": "huge", "climate": "temperate", "wetness": wetness}, nil); err != nil {
		t.Fatal(err)
	}
	if err := client.post(ctx, "/game/update_rules", map[string]any{"session_id": players[0].session, "game_id": gameID, "rules_json": json.RawMessage(`["victory:bloodlust_conquest"]`)}, nil); err != nil {
		t.Fatal(err)
	}
	if err := client.post(ctx, "/game/update_starting_settlers", map[string]any{"session_id": players[0].session, "game_id": gameID, "starting_settlers": 2}, nil); err != nil {
		t.Fatal(err)
	}
	for _, player := range players {
		if err := client.post(ctx, "/game/select_country", map[string]any{"session_id": player.session, "game_id": gameID, "country": player.country}, nil); err != nil {
			t.Fatal(err)
		}
	}
	setup, err := client.readGame(ctx, players[0].session, gameID)
	if err != nil {
		t.Fatal(err)
	}
	for index := range players {
		for _, player := range setup.Players {
			if player.Country == players[index].country {
				players[index].accountID = player.AccountID
			}
		}
		if players[index].accountID == 0 {
			t.Fatalf("missing account for %q", players[index].country)
		}
		if players[index].controller == engine.ControllerAI {
			if err := client.post(ctx, "/game/add_ai", map[string]any{"session_id": players[0].session, "game_id": gameID, "target_account_id": players[index].accountID}, nil); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := client.post(ctx, "/game/start_game", map[string]any{"session_id": players[0].session, "game_id": gameID}, nil); err != nil {
		t.Fatal(err)
	}
	assertOrderedPublicRoster(t, client, players[0].session, gameID, players)
	if telemetry.Unknown != nil {
		t.Fatalf("unknown SQL before restart: %v", telemetry.Unknown)
	}

	server.Close()
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	restartedDatabase, restartedTelemetry, err := memorysql.Open(namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer restartedDatabase.Close()
	restartedRuntime := publicRuntime(sqlx.NewDb(restartedDatabase, memorysql.DriverName), sources, seed)
	restartedServer := httptest.NewServer(restartedRuntime.Handler)
	defer restartedServer.Close()
	client.baseURL = restartedServer.URL
	assertOrderedPublicRoster(t, client, players[0].session, gameID, players)
	if restartedTelemetry.Unknown != nil {
		t.Fatalf("unknown SQL after restart: %v", restartedTelemetry.Unknown)
	}
}

type slotActor struct {
	session    int64
	accountID  int64
	country    string
	controller string
}

type slotHTTPClient struct {
	baseURL string
	client  *http.Client
}

func (c slotHTTPClient) signup(ctx context.Context, name, email string) (slotActor, error) {
	var session int64
	err := c.post(ctx, "/account/create_account", map[string]any{"account": map[string]any{"name": name, "info": "", "location": "", "email": email, "experience_level": 1}, "password": "ordered-slots-password"}, &session)
	return slotActor{session: session}, err
}

func (c slotHTTPClient) createLobby(ctx context.Context, session int64, name string) (int64, error) {
	var result struct {
		ID int64 `json:"id"`
	}
	err := c.post(ctx, "/lobby/create_lobby", map[string]any{"session_id": session, "password": "ordered-slots-password", "lobby": map[string]any{"name": name, "is_public": true, "is_muted": false}}, &result)
	return result.ID, err
}

func (c slotHTTPClient) createGame(ctx context.Context, session, lobbyID int64) (int64, error) {
	var result struct {
		ID int64 `json:"id"`
	}
	err := c.post(ctx, "/game/create_game", map[string]any{"session_id": session, "name": "ordered-slots", "lobby_id": lobbyID, "password_protected": false, "password": ""}, &result)
	return result.ID, err
}

func (c slotHTTPClient) readGame(ctx context.Context, session, gameID int64) (game.Game, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/game/get_game?game_id="+strconv.FormatInt(gameID, 10)+"&session_id="+strconv.FormatInt(session, 10), nil)
	if err != nil {
		return game.Game{}, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return game.Game{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return game.Game{}, fmt.Errorf("GET game status %d", response.StatusCode)
	}
	var result game.Game
	return result, json.NewDecoder(response.Body).Decode(&result)
}

func (c slotHTTPClient) post(ctx context.Context, path string, body any, result any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("POST %s status %d", path, response.StatusCode)
	}
	if result != nil {
		return json.NewDecoder(response.Body).Decode(result)
	}
	return nil
}

func assertOrderedPublicRoster(t *testing.T, client slotHTTPClient, owner, gameID int64, players []slotActor) {
	t.Helper()
	gameDTO, err := client.readGame(context.Background(), owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if gameDTO.Status != "started" || gameDTO.StartingSettlers != 2 || len(gameDTO.Players) != len(players) {
		t.Fatalf("started game = %+v", gameDTO)
	}
	for index, player := range gameDTO.Players {
		want := players[index]
		if player.Country != want.country || player.Controller != want.controller || player.AccountID != want.accountID {
			t.Fatalf("player %d = country/controller/account %q/%q/%d, want %q/%q/%d", index, player.Country, player.Controller, player.AccountID, want.country, want.controller, want.accountID)
		}
	}
}

type orderedClock struct{}

func (orderedClock) Now() time.Time { return time.Unix(0, 0).UTC() }

type orderedIDs struct {
	mu sync.Mutex
	n  int
}

func (ids *orderedIDs) Next() string {
	ids.mu.Lock()
	defer ids.mu.Unlock()
	ids.n++
	return fmt.Sprintf("ordered-slot-%d", ids.n)
}

type orderedRNG struct{}

func (orderedRNG) Next(_ string, state uint64) (uint64, uint64) { return state + 1, state + 1 }

func orderedSetupSources(_ string) engine.Sources {
	return engine.Sources{Clock: orderedClock{}, IDs: &orderedIDs{}, RNG: orderedRNG{}}
}

func publicRuntime(db *sqlx.DB, sources engine.Sources, seed func() (int64, error)) *serverruntime.Runtime {
	return serverruntime.New(db, serverruntime.Config{Sources: sources, StartSeed: seed, NewLimiter: func(string) *limiter.Limiter { return tollbooth.NewLimiter(1000000, nil) }})
}
