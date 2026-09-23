package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/world"
)

func TestInitialGameplayStateDerivesEntitiesFromSnapshot(t *testing.T) {
	snapshot := world.WorldSnapshot{
		Config: world.WorldConfig{StartingSettlers: 2, StartingGold: 75},
		Tiles:  []world.WorldTile{{X: 0, Y: 0, Terrain: "plains"}, {X: 1, Y: 0, Terrain: "shelf"}, {X: 2, Y: 0, Terrain: "forest"}},
		Spawns: []world.PlayerSpawn{{AccountID: 7, X: 0, Y: 0}, {AccountID: 9, X: 2, Y: 0}},
	}
	players := []GamePlayer{{AccountID: 7}, {AccountID: 9}}
	pin, err := serverRuntimePin()
	if err != nil {
		t.Fatal(err)
	}
	first, err := initialGameplayState(snapshot, "sequential", pin, players)
	if err != nil {
		t.Fatal(err)
	}
	second, err := initialGameplayState(snapshot, "sequential", pin, players)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := engine.CanonicalStateJSON(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := engine.CanonicalStateJSON(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("initial state is nondeterministic\n%s\n%s", firstJSON, secondJSON)
	}
	if len(first.Tiles) != 3 || first.Tiles[1].Terrain != engine.TerrainOcean {
		t.Fatalf("tiles = %#v", first.Tiles)
	}
	if len(first.Players) != 2 || len(first.Nations) != 2 || len(first.Units) != 4 {
		t.Fatalf("entities players=%d nations=%d units=%d", len(first.Players), len(first.Nations), len(first.Units))
	}
	if first.Units[0].Strength < 1 || first.Units[0].Armor < 1 || !hasStarterEffect(first.Units[0].Effects, engine.EffectSpecialEmbassy) || !hasStarterEffect(first.Units[0].Effects, engine.EffectSpecialSteal) {
		t.Fatalf("starter capabilities = %#v", first.Units[0])
	}
	if first.Units[0].TileID != "tile-0-0" || first.Units[2].TileID != "tile-2-0" {
		t.Fatalf("units = %#v", first.Units)
	}
	if err := engine.CheckInvariants(first); err != nil {
		t.Fatalf("invariants: %v", err)
	}
}

func TestLoadStartControllersRejectsEmptyPlayers(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	defer sqlxDB.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx := sqlxDB.MustBegin()

	if err := loadStartControllers(tx, nil); err == nil {
		t.Fatal("loadStartControllers accepted an empty persisted player list")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback transaction: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL execution: %v", err)
	}
}

func TestInitialGameplayStatePreservesPersistedAIController(t *testing.T) {
	pin, err := serverRuntimePin()
	if err != nil {
		t.Fatal(err)
	}
	snapshot := world.WorldSnapshot{Config: world.WorldConfig{StartingSettlers: 1}, Tiles: []world.WorldTile{{X: 0, Y: 0, Terrain: "plains"}}, Spawns: []world.PlayerSpawn{{AccountID: 7, X: 0, Y: 0}}}
	state, err := initialGameplayState(snapshot, "sequential", pin, []GamePlayer{{AccountID: 7, Controller: engine.ControllerAI}})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].Controller != engine.ControllerAI {
		t.Fatalf("controller = %q, want AI", state.Players[0].Controller)
	}
	if _, err := initialGameplayState(snapshot, "sequential", pin, []GamePlayer{{AccountID: 7, Controller: "invalid"}}); err == nil {
		t.Fatal("initialGameplayState accepted invalid controller")
	}
}

func hasStarterEffect(effects []engine.Effect, kind string) bool {
	for _, effect := range effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func TestServerRuntimePinIsIndependentOfSetupRules(t *testing.T) {
	pin, err := serverRuntimePin()
	if err != nil {
		t.Fatal(err)
	}
	if string(pin.Definitions) == defaultRulesJSON || pin.Version != startRulesVersion {
		t.Fatalf("server pin = %#v", pin)
	}
	if _, err := rules.CompileRuntime(pin.Definitions, pin.ContentHash); err != nil {
		t.Fatalf("CompileRuntime(server pin): %v", err)
	}
}

func TestServerRuntimePinSelectsConfiguredVictoryEvaluators(t *testing.T) {
	pin, err := serverRuntimePinForSetupRules(`["victory:score_2300","victory:bloodlust_conquest"]`)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := rules.CompileRuntime(pin.Definitions, pin.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	if got := runtime.VictoryEvaluatorIDs(); len(got) != 2 || got[0] != "bloodlust_conquest" || got[1] != "score_2300" {
		t.Fatalf("selected victory evaluators = %#v", got)
	}
	if _, err := serverRuntimePinForSetupRules(`["victory:unknown"]`); err == nil {
		t.Fatal("unsupported configured victory evaluator was accepted")
	}
}

func TestServerRuntimePinCanUseRetailProjectionFromEnvironment(t *testing.T) {
	root := localRetailGameDataRoot(t)
	t.Setenv(retailRulesModeEnv, "retail")
	t.Setenv(retailGameDataEnv, root)
	pin, err := serverRuntimePinForSetupRules(`[]`)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := rules.CompileRuntime(pin.Definitions, pin.ContentHash)
	if err != nil {
		t.Fatalf("CompileRuntime(retail pin): %v", err)
	}
	settler, ok := runtime.Unit("UNIT_SETTLER")
	if !ok || len(runtime.Units) != 74 || runtime.Starter.Movement != settler.Movement {
		t.Fatalf("retail runtime = %+v settler=%+v ok=%v", runtime, settler, ok)
	}
	snapshot := world.WorldSnapshot{Config: world.WorldConfig{StartingSettlers: 1}, Tiles: []world.WorldTile{{X: 0, Y: 0, Terrain: "plains"}, {X: 1, Y: 0, Terrain: "shelf"}, {X: 2, Y: 0, Terrain: "forest"}}, Spawns: []world.PlayerSpawn{{AccountID: 7, X: 0, Y: 0}}}
	state, err := initialGameplayState(snapshot, "sequential", pin, []GamePlayer{{AccountID: 7}})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Units) != 1 || state.Units[0].Movement != runtime.Starter.Movement || state.Units[0].Health != runtime.Starter.Health {
		t.Fatalf("starter from retail runtime = %#v want %+v", state.Units, runtime.Starter)
	}
	if got := []string{state.Tiles[0].Terrain, state.Tiles[1].Terrain, state.Tiles[2].Terrain}; got[0] != "TERRAIN_PLAINS" || got[1] != "TERRAIN_WATER_SHELF" || got[2] != "TERRAIN_FOREST" {
		t.Fatalf("retail startup terrains = %#v", got)
	}
}

func TestRetailRuntimeModeRequiresDataDirectory(t *testing.T) {
	t.Setenv(retailRulesModeEnv, "retail")
	t.Setenv(retailGameDataEnv, "")
	if _, err := serverRuntimePinForSetupRules(`[]`); err == nil {
		t.Fatalf("serverRuntimePinForSetupRules accepted missing %s", retailGameDataEnv)
	}
}

func TestInitialGameplayStateRejectsMissingSpawn(t *testing.T) {
	_, err := initialGameplayState(world.WorldSnapshot{Tiles: []world.WorldTile{{X: 0, Y: 0, Terrain: "plains"}}, Spawns: []world.PlayerSpawn{{AccountID: 7, X: 0, Y: 0}}}, "sequential", engine.RulesPin{Version: startRulesVersion, ContentHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, []GamePlayer{{AccountID: 8}})
	if err == nil {
		t.Fatal("initialGameplayState accepted missing player spawn")
	}
}

func localRetailGameDataRoot(t *testing.T) string {
	t.Helper()
	candidates := []string{
		filepath.Join("..", "..", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"),
		filepath.Join("..", "..", "open-ctp-server", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"),
		"/home/justi/Projects/open-ctp-server/ctp2/ctp2_released_game/ctp2_data/default/gamedata",
	}
	for _, candidate := range candidates {
		if stat, err := os.Stat(filepath.Join(candidate, "Units.txt")); err == nil && !stat.IsDir() {
			return candidate
		}
	}
	t.Skip("local CTP2 released-game data is not present")
	return ""
}
