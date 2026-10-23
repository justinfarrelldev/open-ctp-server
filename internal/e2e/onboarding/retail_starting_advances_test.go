package onboarding

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestRetailPublicStartupPersistsStartingGrantsBeforeFirstFounding(t *testing.T) {
	root := filepath.Join("..", "..", "..", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata")
	if _, err := os.Stat(filepath.Join(root, "DiffDB.txt")); err != nil {
		t.Skip("local released CTP2 data is unavailable")
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
	owner, err := client.Signup(ctx, "retail-founder", "retail-founder@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	gameID, err := client.CreateGame(ctx, owner, 0, "retail initialized start")
	if err != nil {
		t.Fatal(err)
	}
	wetness := 50
	if err := client.ConfigureGame(ctx, owner, gameID, GameSetup{Style: "sequential", RulesJSON: `["difficulty:hard","start-age:AGE_ONE"]`, MapSize: "small", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.SelectCountry(ctx, owner, gameID, "america"); err != nil {
		t.Fatal(err)
	}
	if err := client.AddManagedAI(ctx, owner, gameID, "aztec", "Retail AI"); err != nil {
		t.Fatal(err)
	}
	if err := client.ConfigureGame(ctx, owner, gameID, GameSetup{Style: "sequential", RulesJSON: `["difficulty:invalid"]`, MapSize: "small", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	badStart, _ := json.Marshal(map[string]any{"session_id": owner.Session, "game_id": gameID})
	response, err := client.Client.Post(fmt.Sprintf("%s/game/start_game", client.BaseURL), "application/json", bytes.NewReader(badStart))
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad difficulty should reject as 400, got %d", response.StatusCode)
	}
	var absent string
	if err := db.GetContext(ctx, &absent, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, gameID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed startup persisted state: %v", err)
	}
	if err := client.ConfigureGame(ctx, owner, gameID, GameSetup{Style: "sequential", RulesJSON: `["difficulty:hard","start-age:AGE_ONE"]`, MapSize: "small", Climate: "temperate", Wetness: &wetness}); err != nil {
		t.Fatal(err)
	}
	if err := client.StartGame(ctx, owner, gameID); err != nil {
		t.Fatal(err)
	}
	readCanonical := func() []byte {
		t.Helper()
		var raw string
		if err := db.GetContext(ctx, &raw, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, gameID); err != nil {
			t.Fatal(err)
		}
		return []byte(raw)
	}
	initial := readCanonical()
	state, err := engine.DecodeCanonicalStateJSON(initial)
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	if state.Revision != 1 || !state.StartingAdvancesInitialized || compiled.StartingAdvances == nil || compiled.StartingAdvances.DifficultyID != "hard" {
		t.Fatal("transaction persisted an uninitialized retail game")
	}
	for _, nation := range state.Nations {
		if !slices.Contains(nation.Advances, "ADVANCE_TOOLMAKING") || nation.CurrentAgeID != "AGE_ONE" || nation.ResearchTargetID == "" || !engine.CanResearchAdvance(nation, compiled, nation.ResearchTargetID) {
			t.Fatalf("persisted startup missing grants/research: %+v", nation)
		}
	}
	raw, err := client.ReadProjectedWorld(ctx, owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	var p visibility.Projection
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	var founding visibility.ActionProjection
	for _, action := range p.LegalActions {
		if action.Command == engine.CommandFoundCity {
			founding = action
			break
		}
	}
	if len(founding.Payload) == 0 {
		t.Fatal("initial founder has no advertised found action")
	}
	payload := map[string]any{}
	if err := json.Unmarshal(founding.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	revision, err := client.GameplayCommand(ctx, owner, gameID, founding.Route, p.Revision, "retail-found", payload)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = client.ReadProjectedWorld(ctx, owner, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Cities) != 1 || p.Cities[0].Population != 1 || p.Cities[0].ProductionOptions == nil || p.Cities[0].ProductionOptions.Status != "complete" {
		t.Fatalf("first city has no complete choices: %+v", p.Cities)
	}
	city := p.Cities[0]
	warrior := false
	for _, item := range city.ProductionOptions.Items {
		if item.ID == "UNIT_WARRIOR" {
			warrior = true
		}
		if item.ID == "UNIT_SETTLER" {
			t.Fatal("population-one settler advertised")
		}
	}
	if !warrior {
		t.Fatal("first founding receipt/current view omits Warrior")
	}
	entries := []engine.QueueSelection{{ItemID: "UNIT_WARRIOR"}}
	edit := map[string]any{"city_id": city.ID, "entries": entries}
	accepted, err := client.GameplayCommand(ctx, owner, gameID, "edit-production", revision, "retail-warrior", edit)
	if err != nil {
		t.Fatal(err)
	}
	beforeRestart := readCanonical()
	server.Close()
	runtime = NewPublicRuntime(db, sources, seed)
	server = httptest.NewServer(runtime.Handler)
	defer server.Close()
	client.BaseURL = server.URL
	if err := client.StartGame(ctx, owner, gameID); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate start not rejected: %v", err)
	}
	replay, err := client.GameplayCommand(ctx, owner, gameID, "edit-production", revision, "retail-warrior", edit)
	if err != nil || replay != accepted {
		t.Fatalf("immutable replay failed after restart: %d %v", replay, err)
	}
	if string(beforeRestart) != string(readCanonical()) {
		t.Fatal("restart/duplicate start/receipt replay rerolled or changed grants")
	}
	if _, err := client.GameplayCommand(ctx, owner, gameID, "edit-production", accepted, "retail-reject-settler", map[string]any{"city_id": city.ID, "entries": []engine.QueueSelection{{ItemID: "UNIT_SETTLER"}}}); err == nil {
		t.Fatal("population-one settler accepted through HTTP")
	}
	if telemetry.Unknown != nil {
		t.Fatalf("unexpected persistence SQL: %v", telemetry.Unknown)
	}
}
