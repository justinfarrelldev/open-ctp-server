package onboarding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestPublicRetailAIQueuesAndBuildsUnitsWithOriginalCityNamesAndTerritory(t *testing.T) {
	root := filepath.Join("..", "..", "..", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata")
	if client := os.Getenv("CTPCLIENT"); client != "" {
		candidate := filepath.Join(client, "ctp2", "ctp2_data", "default", "gamedata")
		if _, err := os.Stat(filepath.Join(candidate, "civilisation.txt")); err == nil {
			root = candidate
		}
	}
	if _, err := os.Stat(filepath.Join(root, "civilisation.txt")); err != nil {
		t.Skip("released game data unavailable")
	}
	t.Setenv("OPEN_CTP_RULES_MODE", "retail")
	t.Setenv("OPEN_CTP_RETAIL_GAMEDATA", root)
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
	owner, err := client.Signup(ctx, "city-developer", "city-developer@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(ctx, owner, 0, "AI original city development")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(ctx, owner, gameID, GameSetup{Style: "sequential", RulesJSON: `["difficulty:medium"]`, MapSize: "small", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(ctx, owner, gameID, "Romans"); err != nil {
		t.Fatal(err)
	}
	if err := client.AddManagedAI(ctx, owner, gameID, "Greeks", "Greek AI"); err != nil {
		t.Fatal(err)
	}
	if err := client.StartGame(ctx, owner, gameID); err != nil {
		t.Fatal(err)
	}
	read := func() engine.State {
		t.Helper()
		var raw string
		if err := db.GetContext(ctx, &raw, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, gameID); err != nil {
			t.Fatal(err)
		}
		state, err := engine.DecodeCanonicalStateJSON([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		return state
	}
	project := func() visibility.Projection {
		t.Helper()
		raw, err := client.ReadProjectedWorld(ctx, owner, gameID)
		if err != nil {
			t.Fatal(err)
		}
		var p visibility.Projection
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	p := project()
	humanID := p.ViewerPlayerID
	var aiID string
	for _, identity := range p.Players {
		if identity.ID != humanID {
			aiID = identity.ID
		}
		if identity.ColorIndex == nil {
			t.Fatal("new game omitted authoritative palette slot")
		}
	}
	if humanID == "" || aiID == "" {
		t.Fatal("missing stable identities")
	}
	var founding visibility.ActionProjection
	for _, action := range p.LegalActions {
		if action.Command == engine.CommandFoundCity {
			founding = action
			break
		}
	}
	if len(founding.Payload) == 0 {
		t.Fatal("no legal founding action")
	}
	var payload map[string]any
	if err := json.Unmarshal(founding.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["name"] != "Rome" {
		t.Fatalf("Roman founding suggestion = %v", payload["name"])
	}
	if _, err := client.GameplayCommand(ctx, owner, gameID, founding.Route, p.Revision, "roman-found", payload); err != nil {
		t.Fatal(err)
	}
	p = project()
	if len(p.Cities) != 1 || p.Cities[0].Name != "Rome" {
		t.Fatalf("Roman city = %+v", p.Cities)
	}
	national, influence := 0, 0
	for _, tile := range p.Tiles {
		if tile.OwnerID != nil && *tile.OwnerID == humanID {
			national++
		}
		if tile.InfluenceCityID != nil && *tile.InfluenceCityID == p.Cities[0].ID {
			influence++
		}
	}
	if national < 2 || influence < 2 {
		t.Fatalf("founding did not publish territory/influence: %d/%d", national, influence)
	}
	sawQueue, produced := false, false
	for turn := 0; turn < 48 && !produced; turn++ {
		state := read()
		if _, err := client.GameplayCommand(ctx, owner, gameID, "pass", state.Revision, fmt.Sprintf("city-human-pass-%d", turn), map[string]any{}); err != nil {
			t.Fatal(err)
		}
		if err := runtime.RunAutonomousAI(ctx); err != nil {
			t.Fatalf("actual one-action AI worker: %v", err)
		}
		state = read()
		for _, city := range state.Cities {
			if city.OwnerID == aiID {
				if city.Name != "Athens" {
					t.Fatalf("Greek capital named %s", city.Name)
				}
				sawQueue = sawQueue || len(city.Queue) > 0
			}
		}
		for _, unit := range state.Units {
			if unit.OwnerID == aiID && unit.TypeID != "UNIT_SETTLER" {
				for _, city := range state.Cities {
					if city.OwnerID == aiID && strings.HasPrefix(unit.ID, "unit:"+city.ID+":") {
						produced = true
					}
				}
			}
		}
		if turn == 4 { // Reconstruct the real runtime while the worker is developing.
			server.Close()
			runtime = NewPublicRuntime(db, sources, seed)
			server = httptest.NewServer(runtime.Handler)
			defer server.Close()
			client.BaseURL = server.URL
		}
	}
	if !sawQueue || !produced {
		t.Fatalf("AI never queued/completed ordinary production: queue=%v produced=%v state=%+v", sawQueue, produced, read().Cities)
	}
	p = project()
	if p.ViewerPlayerID != humanID {
		t.Fatal("identity changed after settlers disappeared/runtime restart")
	}
	before, _ := engine.CanonicalStateJSON(read())
	replayed, err := client.GameplayCommand(ctx, owner, gameID, founding.Route, 1, "roman-found", payload)
	if err != nil || replayed != 2 {
		t.Fatalf("founding receipt replay after worker/restart: %d %v", replayed, err)
	}
	after, _ := engine.CanonicalStateJSON(read())
	if !bytes.Equal(before, after) {
		t.Fatal("founding receipt replay changed names, borders or influence")
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unexpected SQL: %v", telemetry.Unknown)
	}
}
