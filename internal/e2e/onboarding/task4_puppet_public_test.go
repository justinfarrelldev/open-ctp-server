package onboarding

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/puppet"
)

func TestPuppetCommandTraversesAuthenticatedHumanHTTPPath(t *testing.T) {
	namespace := t.Name()
	memorysql.Reset(namespace)
	defer memorysql.Reset(namespace)
	database, telemetry, err := memorysql.Open(namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	db := sqlx.NewDb(database, memorysql.DriverName)
	runtime := publicRuntime(db, orderedSetupSources(namespace), func() (int64, error) { return 2, nil })
	server := httptest.NewServer(runtime.Handler)
	defer server.Close()
	setupClient := slotHTTPClient{baseURL: server.URL, client: server.Client()}
	ctx := context.Background()
	owner, err := setupClient.signup(ctx, "puppet-owner", "puppet-owner@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	other, err := setupClient.signup(ctx, "puppet-other", "puppet-other@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	lobbyID, err := setupClient.createLobby(ctx, owner.session, "puppet-public-path")
	if err != nil {
		t.Fatal(err)
	}
	if err := setupClient.post(ctx, "/lobby/join_lobby", map[string]any{"session_id": other.session, "lobby_id": lobbyID}, nil); err != nil {
		t.Fatal(err)
	}
	gameID, err := setupClient.createGame(ctx, owner.session, lobbyID)
	if err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string]map[string]any{
		"/game/update_style":             {"session_id": owner.session, "game_id": gameID, "style": "sequential", "map_size": "huge", "climate": "temperate", "wetness": 50},
		"/game/update_rules":             {"session_id": owner.session, "game_id": gameID, "rules_json": json.RawMessage(`["victory:bloodlust_conquest"]`)},
		"/game/update_starting_settlers": {"session_id": owner.session, "game_id": gameID, "starting_settlers": 1},
	} {
		if err := setupClient.post(ctx, path, body, nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := setupClient.post(ctx, "/game/select_country", map[string]any{"session_id": owner.session, "game_id": gameID, "country": syntheticCivilizationID(1)}, nil); err != nil {
		t.Fatal(err)
	}
	if err := setupClient.post(ctx, "/game/select_country", map[string]any{"session_id": other.session, "game_id": gameID, "country": syntheticCivilizationID(2)}, nil); err != nil {
		t.Fatal(err)
	}
	gameDTO, err := setupClient.readGame(ctx, owner.session, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if len(gameDTO.Players) != 2 {
		t.Fatalf("setup players = %d, want 2", len(gameDTO.Players))
	}
	ownerPlayerID := "account-" + strconv.FormatInt(gameDTO.Players[0].AccountID, 10)
	if err := setupClient.post(ctx, "/game/start_game", map[string]any{"session_id": owner.session, "game_id": gameID}, nil); err != nil {
		t.Fatal(err)
	}

	publicClient := puppet.Client{BaseURL: server.URL, HTTPClient: server.Client()}
	puppeteer := puppet.Puppeteer{Client: publicClient}
	input := puppet.PlanningInput{ActorPlayerID: ownerPlayerID, ExpectedRevision: 1, PublicProjectionJSON: []byte(`{}`), PublicHistoryJSON: []byte(`[]`), TestHarnessPolicyLevel: puppet.TestHarnessPolicyLevelPassOnly}
	_, err = puppeteer.SelectAndSubmit(ctx, puppet.HumanActor{GameID: gameID, SessionID: other.session}, input, "wrong-human-session")
	var publicError *puppet.HTTPError
	if !errors.As(err, &publicError) || publicError.Status < http.StatusBadRequest {
		t.Fatalf("wrong human actor did not fail through public path: %v", err)
	}
	result, err := puppeteer.SelectAndSubmit(ctx, puppet.HumanActor{GameID: gameID, SessionID: owner.session}, input, "authorized-human-session")
	if err != nil {
		t.Fatal(err)
	}
	if result.Revision != 2 {
		t.Fatalf("ordinary human handler revision = %d, want 2", result.Revision)
	}
	if telemetry.Unknown != nil {
		t.Fatalf("puppet public path used unknown SQL: %v", telemetry.Unknown)
	}
}
