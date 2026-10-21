package onboarding

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/memorysql"
	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestTask13PublicCrossSystemScenariosAreDeterministic(t *testing.T) {
	scenarios := []struct {
		name string
		seed int64
		run  func(*task13Harness) string
	}{
		{name: "advance-production-movement", seed: 13, run: task13AdvanceProductionMovementScenario},
		{name: "terrain-public-works-pollution", seed: 14, run: task13TerrainPublicWorksScenario},
		{name: "diplomacy-government-ai", seed: 15, run: task13DiplomacyGovernmentAIScenario},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			cleanA := task13RunScenario(t, scenario.name, scenario.seed, "clean-a", false, scenario.run)
			cleanB := task13RunScenario(t, scenario.name, scenario.seed, "clean-b", false, scenario.run)
			restarted := task13RunScenario(t, scenario.name, scenario.seed, "restart", true, scenario.run)
			if cleanA != cleanB || cleanA != restarted {
				t.Fatalf("scenario %s is not deterministic:\ncleanA=%s\ncleanB=%s\nrestart=%s", scenario.name, cleanA, cleanB, restarted)
			}
		})
	}
}

type task13Harness struct {
	t         *testing.T
	ctx       context.Context
	database  *sql.DB
	db        *sqlx.DB
	sources   engine.Sources
	seed      func() (int64, error)
	runtime   *PublicRuntime
	server    *httptest.Server
	client    HTTPClient
	owner     Actor
	guest     Actor
	gameID    int64
	ownerID   string
	aiID      string
	revision  int64
	telemetry *memorysql.Telemetry
}

func task13RunScenario(t *testing.T, scenario string, seed int64, runLabel string, restart bool, run func(*task13Harness) string) string {
	t.Helper()
	h := task13NewHarness(t, scenario, runLabel, seed)
	defer h.Close()
	objective := run(h)
	if restart {
		h.Restart()
	}
	stateHash, eventHash, projectionHash := h.Hashes()
	return fmt.Sprintf("objective=%s state=%s events=%s projection=%s", objective, stateHash, eventHash, projectionHash)
}

func task13NewHarness(t *testing.T, scenario, runLabel string, seedValue int64) *task13Harness {
	t.Helper()
	database, telemetry, err := memorysql.Open("task13-" + scenario + "-" + runLabel)
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(database, memorysql.DriverName)
	sources := task13Sources(t, scenario, seedValue)
	seed := func() (int64, error) { return seedValue, nil }
	runtime := NewPublicRuntime(db, sources, seed)
	server := httptest.NewServer(runtime.Handler)
	h := &task13Harness{t: t, ctx: context.Background(), database: database, db: db, sources: sources, seed: seed, runtime: runtime, server: server, client: HTTPClient{BaseURL: server.URL, Client: &http.Client{}}, telemetry: telemetry}
	h.setupGame(scenario)
	return h
}

func task13Sources(t *testing.T, scenario string, seed int64) engine.Sources {
	t.Helper()
	kit := testkit.NewScenario("task13-"+scenario, seed)
	t.Cleanup(kit.Close)
	return engine.Sources{Clock: kit.Clock, IDs: kit.IDs, RNG: deterministicRNG{}}
}

func (h *task13Harness) setupGame(scenario string) {
	var err error
	h.owner, err = h.client.Signup(h.ctx, scenario+"-owner", scenario+"-owner@example.invalid")
	if err != nil {
		h.t.Fatal(err)
	}
	h.guest, err = h.client.Signup(h.ctx, scenario+"-ai", scenario+"-ai@example.invalid")
	if err != nil {
		h.t.Fatal(err)
	}
	lobbyID, err := h.client.CreateLobby(h.ctx, h.owner, scenario+"-lobby")
	if err != nil {
		h.t.Fatal(err)
	}
	if err := h.client.JoinLobby(h.ctx, h.guest, lobbyID); err != nil {
		h.t.Fatal(err)
	}
	h.gameID, err = h.client.CreateGame(h.ctx, h.owner, lobbyID, scenario+"-game")
	if err != nil {
		h.t.Fatal(err)
	}
	wetness := 50
	if err := h.client.ConfigureGame(h.ctx, h.owner, h.gameID, GameSetup{Style: "sequential", MapSize: "duel", Climate: "temperate", Wetness: &wetness}); err != nil {
		h.t.Fatal(err)
	}
	if err := h.client.SelectCountry(h.ctx, h.owner, h.gameID, "america"); err != nil {
		h.t.Fatal(err)
	}
	if err := h.client.SelectCountry(h.ctx, h.guest, h.gameID, "aztec"); err != nil {
		h.t.Fatal(err)
	}
	gameDTO, err := h.client.ReadGame(h.ctx, h.owner, h.gameID)
	if err != nil {
		h.t.Fatal(err)
	}
	for _, player := range gameDTO.Players {
		playerID := "account-" + strconv.FormatInt(player.AccountID, 10)
		switch player.Country {
		case "america":
			h.ownerID = playerID
		case "aztec":
			h.aiID = playerID
			if err := h.client.AddAI(h.ctx, h.owner, h.gameID, player.AccountID); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	if h.ownerID == "" || h.aiID == "" {
		h.t.Fatalf("missing player IDs owner=%q ai=%q", h.ownerID, h.aiID)
	}
	if err := h.client.StartGame(h.ctx, h.owner, h.gameID); err != nil {
		h.t.Fatal(err)
	}
	h.revision = h.OwnerProjection().Revision
}

func (h *task13Harness) Close() {
	if h.server != nil {
		h.server.Close()
	}
	if h.telemetry != nil && h.telemetry.Unknown != nil {
		h.t.Fatalf("unknown SQL = %v", h.telemetry.Unknown)
	}
	if h.database != nil {
		h.database.Close()
	}
}

func (h *task13Harness) Restart() {
	h.server.Close()
	h.runtime = RestartPublicRuntime(h.db, h.sources, h.seed)
	h.server = httptest.NewServer(h.runtime.Handler)
	h.client.BaseURL = h.server.URL
	restarted := h.OwnerProjection()
	if restarted.Revision < h.revision {
		h.t.Fatalf("restart revision regressed from %d to %d", h.revision, restarted.Revision)
	}
}

func (h *task13Harness) OwnerProjection() visibility.Projection {
	h.t.Helper()
	return simultaneousProjection(h.t, h.server.URL, h.gameID, h.owner.Session)
}

func (h *task13Harness) SubmitAction(actor Actor, key string, action visibility.ActionProjection) {
	h.t.Helper()
	payload := map[string]any{}
	if len(action.Payload) > 0 && string(action.Payload) != "null" {
		if err := json.Unmarshal(action.Payload, &payload); err != nil {
			h.t.Fatalf("decode legal action payload for %s: %v", action.Command, err)
		}
	}
	h.Command(actor, action.Route, key, payload)
}

func (h *task13Harness) Command(actor Actor, route, key string, payload map[string]any) {
	h.t.Helper()
	next, err := h.client.GameplayCommand(h.ctx, actor, h.gameID, route, h.revision, key, payload)
	if err != nil {
		h.t.Fatalf("%s at revision %d: %v", route, h.revision, err)
	}
	if next <= h.revision {
		h.t.Fatalf("%s did not advance revision: before=%d after=%d", route, h.revision, next)
	}
	h.revision = next
}

func (h *task13Harness) PassOwner(key string) { h.Command(h.owner, "pass", key, map[string]any{}) }

func (h *task13Harness) DriveAI(label string) {
	h.t.Helper()
	result, err := h.runtime.Driver.Drive(h.ctx, h.gameID, h.aiID, h.revision)
	if err != nil {
		h.t.Fatalf("drive AI %s: %v", label, err)
	}
	if result.Commands == 0 {
		h.t.Fatalf("drive AI %s submitted no public gameplay command", label)
	}
	h.revision = result.Revision
}

func (h *task13Harness) CompleteRound(label string) {
	h.PassOwner(label + "-owner-pass")
	h.DriveAI(label + "-ai")
}

func (h *task13Harness) Hashes() (string, string, string) {
	h.t.Helper()
	var stateJSON string
	if err := h.db.GetContext(h.ctx, &stateJSON, `SELECT canonical_state_json::text FROM world_turn_states WHERE game_id = $1 AND state_schema_version = 1`, h.gameID); err != nil {
		h.t.Fatal(err)
	}
	type eventRow struct {
		EventID string `db:"event_id"`
		Payload string `db:"payload_json"`
	}
	var rows []eventRow
	if err := h.db.SelectContext(h.ctx, &rows, `SELECT event_id, payload_json::text FROM gameplay_events WHERE game_id = $1 ORDER BY sequence, ordinal`, h.gameID); err != nil {
		h.t.Fatal(err)
	}
	eventsJSON, _ := json.Marshal(rows)
	projectionJSON, _ := json.Marshal(h.OwnerProjection())
	return task13Hash([]byte(stateJSON)), task13Hash(eventsJSON), task13Hash(projectionJSON)
}

func task13Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func task13AdvanceProductionMovementScenario(h *task13Harness) string {
	projection := h.OwnerProjection()
	if projection.Rules == nil || projection.Rules.ContentHash == "" || len(projection.Rules.Units) == 0 || len(projection.Rules.Advances) == 0 || len(projection.LegalActions) == 0 {
		h.t.Fatalf("projection is not client-ready: rules=%+v legal=%d", projection.Rules, len(projection.LegalActions))
	}
	h.SubmitAction(h.owner, "task13-research-root", task13FindAction(h.t, projection, engine.CommandSetResearch, nil))
	projection = h.OwnerProjection()
	h.SubmitAction(h.owner, "task13-move", task13FindAction(h.t, projection, engine.CommandMoveUnit, nil))
	projection = h.OwnerProjection()
	h.SubmitAction(h.owner, "task13-found", task13FindAction(h.t, projection, engine.CommandFoundCity, nil))
	for cycles := 0; !task13HasAdvance(h.OwnerProjection(), "ADVANCE_SYNTHETIC_UNLOCK") && cycles < 4; cycles++ {
		h.CompleteRound(fmt.Sprintf("task13-unlock-%d", cycles))
	}
	projection = h.OwnerProjection()
	if !task13HasAdvance(projection, "ADVANCE_SYNTHETIC_UNLOCK") {
		h.t.Fatalf("owner research did not unlock synthetic content: %+v", projection.Research)
	}
	city := task13OwnedCity(h.t, projection, h.ownerID)
	// A public client consumes pinned production, never defines its own lab
	// cost, prerequisite or research effect inside an executable request.
	building := task13FindAction(h.t, projection, engine.CommandQueueProduction, func(action visibility.ActionProjection) bool {
		return action.EntityID == city.ID && action.TargetID == "IMPROVE_SYNTHETIC"
	})
	h.SubmitAction(h.owner, "task13-queue-building", building)
	h.CompleteRound("task13-produce-lab")
	projection = h.OwnerProjection()
	task13RequireBuilding(h.t, projection, h.ownerID, "IMPROVE_SYNTHETIC")
	return "advance-progression-and-production-building"
}

func task13TerrainPublicWorksScenario(h *task13Harness) string {
	projection := h.OwnerProjection()
	move := task13FindAction(h.t, projection, engine.CommandMoveUnit, nil)
	h.SubmitAction(h.owner, "task13-terrain-move", move)
	h.CompleteRound("task13-terrain-move-refresh")
	projection = h.OwnerProjection()
	beforeWorks := projection.Diplomacy.PublicWorks
	cleanup := task13FindAction(h.t, projection, engine.CommandBuildImprovement, func(action visibility.ActionProjection) bool {
		return task13PayloadString(action, "improvement_id") == "TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP"
	})
	h.SubmitAction(h.owner, "task13-cleanup", cleanup)
	h.CompleteRound("task13-cleanup-complete")
	projection = h.OwnerProjection()
	if projection.Diplomacy == nil || projection.Diplomacy.PublicWorks > beforeWorks-1 {
		h.t.Fatalf("public works did not fund cleanup: before=%d after=%v", beforeWorks, projection.Diplomacy)
	}
	if pollution := task13TilePollution(projection, cleanup.TargetID); pollution != 0 {
		h.t.Fatalf("cleanup/public pollution clamp = %d, want 0", pollution)
	}
	return "terrain-cleanup-public-works-move"
}

func task13DiplomacyGovernmentAIScenario(h *task13Harness) string {
	h.Command(h.owner, "establish-contact", "task13-contact", map[string]any{"target_player_id": h.aiID})
	projection := h.OwnerProjection()
	h.Command(h.owner, "propose-agreement", "task13-propose-trade", map[string]any{"proposal_id": "task13-trade", "target_player_id": h.aiID, "kind": "trade"})
	for cycles := 0; !task13RelationTrade(h.OwnerProjection(), h.aiID) && cycles < 4; cycles++ {
		h.CompleteRound(fmt.Sprintf("task13-diplomacy-response-%d", cycles))
	}
	projection = h.OwnerProjection()
	if !task13RelationTrade(projection, h.aiID) {
		h.t.Fatalf("AI did not accept public diplomacy proposal: %+v", projection.Diplomacy)
	}
	if projection.Research == nil || projection.Research.TargetID == "" {
		research := task13FindAction(h.t, projection, engine.CommandSetResearch, nil)
		h.SubmitAction(h.owner, "task13-government-research", research)
	}
	for cycles := 0; !task13HasAdvance(h.OwnerProjection(), "ADVANCE_SYNTHETIC_ROOT") && cycles < 3; cycles++ {
		h.CompleteRound(fmt.Sprintf("task13-government-unlock-%d", cycles))
	}
	projection = h.OwnerProjection()
	if !task13HasAdvance(projection, "ADVANCE_SYNTHETIC_ROOT") {
		h.t.Fatalf("owner did not research government prerequisite: %+v", projection.Research)
	}
	government := task13FindAction(h.t, projection, engine.CommandSetGovernment, func(action visibility.ActionProjection) bool {
		return action.TargetID == "GOVERNMENT_SYNTHETIC_REPUBLIC"
	})
	h.SubmitAction(h.owner, "task13-government", government)
	projection = h.OwnerProjection()
	if projection.Diplomacy == nil || projection.Diplomacy.GovernmentID != "GOVERNMENT_SYNTHETIC_REPUBLIC" {
		h.t.Fatalf("government did not change from public command: %+v", projection.Diplomacy)
	}
	return "diplomacy-union-government-republic"
}

func task13FindAction(t *testing.T, projection visibility.Projection, command string, match func(visibility.ActionProjection) bool) visibility.ActionProjection {
	t.Helper()
	for _, action := range projection.LegalActions {
		if action.Command == command && (match == nil || match(action)) {
			return action
		}
	}
	commands := make([]string, 0, len(projection.LegalActions))
	for _, action := range projection.LegalActions {
		commands = append(commands, action.Command+":"+action.TargetID+":"+string(action.Payload))
	}
	sort.Strings(commands)
	t.Fatalf("legal action %q not found; have %v", command, commands)
	return visibility.ActionProjection{}
}

func task13PayloadString(action visibility.ActionProjection, field string) string {
	payload := map[string]any{}
	_ = json.Unmarshal(action.Payload, &payload)
	if value, ok := payload[field].(string); ok {
		return value
	}
	return ""
}

func task13TilePollution(projection visibility.Projection, tileID string) int {
	for _, tile := range projection.Tiles {
		if tile.ID == tileID {
			return tile.Pollution
		}
	}
	return -1
}

func task13HasAdvance(projection visibility.Projection, advanceID string) bool {
	if projection.Research == nil {
		return false
	}
	for _, existing := range projection.Research.Advances {
		if existing == advanceID {
			return true
		}
	}
	return false
}

func task13RelationTrade(projection visibility.Projection, targetID string) bool {
	if projection.Diplomacy == nil {
		return false
	}
	for _, relation := range projection.Diplomacy.Relations {
		if relation.TargetPlayerID == targetID && relation.Trade && relation.Treaty == "trade" {
			return true
		}
	}
	return false
}

func task13OwnedCity(t *testing.T, projection visibility.Projection, ownerID string) visibility.CityProjection {
	t.Helper()
	for _, city := range projection.Cities {
		if city.OwnerID == ownerID && !city.Stale {
			return city
		}
	}
	t.Fatalf("projection has no owned city for %s: %+v", ownerID, projection.Cities)
	return visibility.CityProjection{}
}

func task13RequireBuilding(t *testing.T, projection visibility.Projection, ownerID, buildingID string) {
	t.Helper()
	for _, city := range projection.Cities {
		if city.OwnerID != ownerID || city.Stale {
			continue
		}
		for _, building := range city.Buildings {
			if building.ID == buildingID {
				return
			}
		}
	}
	var cities []string
	for _, city := range projection.Cities {
		if city.OwnerID == ownerID {
			ids := make([]string, 0, len(city.Buildings))
			for _, building := range city.Buildings {
				ids = append(ids, building.ID)
			}
			cities = append(cities, fmt.Sprintf("%s:%v", city.ID, ids))
		}
	}
	t.Fatalf("building %q not visible for %s; cities=%s", buildingID, ownerID, strings.Join(cities, ","))
}
