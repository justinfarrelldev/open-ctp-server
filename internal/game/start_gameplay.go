package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	gameplaystore "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/store"
	"github.com/justinfarrelldev/open-ctp-server/internal/world"
)

const startRulesVersion = "ctp2-supported-v2"

const (
	retailRulesModeEnv = "OPEN_CTP_RULES_MODE"
	retailGameDataEnv  = "OPEN_CTP_RETAIL_GAMEDATA"
)

type startClock struct{}

func (startClock) Now() time.Time { return time.Now().UTC() }

type startIDs struct{ next uint64 }

func (s *startIDs) Next() string { s.next++; return fmt.Sprintf("start-%d", s.next) }

type startRNG struct{}

func (startRNG) Next(_ string, state uint64) (uint64, uint64) {
	next := state + 0x9e3779b97f4a7c15
	return next, next
}

func defaultStartGameplaySources() engine.Sources {
	return engine.Sources{Clock: startClock{}, IDs: &startIDs{}, RNG: startRNG{}}
}

func serverRuntimePin() (engine.RulesPin, error) {
	return serverRuntimePinForSetupRules(defaultRulesJSON)
}

func serverRuntimePinForSetupRules(setupRules string) (engine.RulesPin, error) {
	if strings.EqualFold(os.Getenv(retailRulesModeEnv), "retail") {
		return retailRuntimePin(os.Getenv(retailGameDataEnv))
	}
	var options []string
	if err := json.Unmarshal([]byte(setupRules), &options); err != nil {
		return engine.RulesPin{}, fmt.Errorf("decode game rules configuration: %w", err)
	}
	selected := make([]string, 0)
	for _, option := range options {
		if !strings.HasPrefix(option, "victory:") {
			continue
		}
		evaluatorID := strings.TrimPrefix(option, "victory:")
		if evaluatorID == "" {
			return engine.RulesPin{}, fmt.Errorf("game rules configuration contains an empty victory evaluator")
		}
		selected = append(selected, evaluatorID)
	}
	ruleset, err := rules.ServerOwnedSyntheticRuleset(selected...)
	if err != nil {
		return engine.RulesPin{}, err
	}
	return engine.RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: json.RawMessage(append([]byte(nil), ruleset.CanonicalJSON...))}, nil
}

func retailRuntimePin(dataDir string) (engine.RulesPin, error) {
	if strings.TrimSpace(dataDir) == "" {
		return engine.RulesPin{}, fmt.Errorf("%s is required when %s=retail", retailGameDataEnv, retailRulesModeEnv)
	}
	db, err := rules.LoadRetailRulesFromDirectory(dataDir)
	if err != nil {
		return engine.RulesPin{}, err
	}
	runtime := rules.ProjectRetailRuntime(db)
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		return engine.RulesPin{}, err
	}
	if _, err := rules.CompileRuntime(definitions, contentHash); err != nil {
		return engine.RulesPin{}, err
	}
	return engine.RulesPin{Version: runtime.CompatibilityVersion, ContentHash: contentHash, Definitions: json.RawMessage(append([]byte(nil), definitions...))}, nil
}

func loadStartControllers(tx *sqlx.Tx, players []GamePlayer) error {
	if tx == nil {
		return fmt.Errorf("start controller load requires a transaction")
	}
	if len(players) == 0 {
		return fmt.Errorf("start controller load requires at least one persisted player")
	}
	var rows []struct {
		AccountID  int64  `db:"account_id"`
		Controller string `db:"controller_kind"`
	}
	if err := tx.Select(&rows, `SELECT account_id, controller_kind FROM game_players WHERE game_id = $1`, players[0].GameID); err != nil {
		return fmt.Errorf("load start controllers: %w", err)
	}
	controllers := make(map[int64]string, len(rows))
	for _, row := range rows {
		controllers[row.AccountID] = row.Controller
	}
	for index := range players {
		controller, ok := controllers[players[index].AccountID]
		if !ok {
			return fmt.Errorf("missing persisted controller for account %d", players[index].AccountID)
		}
		players[index].Controller = controller
	}
	return nil
}

func initialGameplayState(snapshot world.WorldSnapshot, style string, pin engine.RulesPin, players []GamePlayer) (engine.State, error) {
	if len(players) == 0 || len(snapshot.Spawns) == 0 || len(snapshot.Tiles) == 0 {
		return engine.State{}, fmt.Errorf("generated world must contain players, spawns, and tiles")
	}
	runtime, err := rules.CompileRuntime(pin.Definitions, pin.ContentHash)
	if err != nil {
		return engine.State{}, fmt.Errorf("compile server runtime rules: %w", err)
	}
	spawnTileIDs := make(map[int64]string, len(snapshot.Spawns))
	for _, spawn := range snapshot.Spawns {
		spawnTileIDs[spawn.AccountID] = tileID(spawn.X, spawn.Y)
	}
	state := engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         pin,
		Phase:         engine.PhaseState{Mode: style, ModeVersion: "turn-policy-v1", Turn: 1},
		RNG:           []engine.RNGDomainState{{Domain: engine.RNGDomainCombat, State: 1}, {Domain: engine.RNGDomainSpecialAction, State: 2}, {Domain: engine.RNGDomainAI, State: 3}, {Domain: engine.RNGDomainRiot, State: 4}},
	}
	for _, tile := range snapshot.Tiles {
		state.Tiles = append(state.Tiles, engine.Tile{ID: tileID(tile.X, tile.Y), X: tile.X, Y: tile.Y, Terrain: engineTerrainForRuntime(tile.Terrain, runtime), Open: true})
	}
	for _, player := range players {
		playerID := fmt.Sprintf("account-%d", player.AccountID)
		if _, ok := spawnTileIDs[player.AccountID]; !ok {
			return engine.State{}, fmt.Errorf("missing generated spawn for account %d", player.AccountID)
		}
		controller, err := startController(player.Controller)
		if err != nil {
			return engine.State{}, err
		}
		state.Phase.ActivePlayerOrder = append(state.Phase.ActivePlayerOrder, playerID)
		state.Players = append(state.Players, engine.Player{ID: playerID, NationID: "nation-" + playerID, Lifecycle: engine.PlayerActive, Controller: controller})
		state.Nations = append(state.Nations, engine.Nation{ID: "nation-" + playerID, PublicWorks: snapshot.Config.StartingGold, Gold: snapshot.Config.StartingGold})
	}
	for _, player := range players {
		playerID := fmt.Sprintf("account-%d", player.AccountID)
		tileID := spawnTileIDs[player.AccountID]
		for ordinal := 0; ordinal < snapshot.Config.StartingSettlers; ordinal++ {
			state.Units = append(state.Units, engine.Unit{ID: fmt.Sprintf("settler-%d-%d", player.AccountID, ordinal+1), OwnerID: playerID, TileID: tileID, Domain: runtime.Starter.Domain, Movement: runtime.Starter.Movement, Fuel: runtime.Starter.MaxFuel, MaxFuel: runtime.Starter.MaxFuel, CargoCapacity: runtime.Starter.TransportCapacity, CargoDomains: append([]string(nil), runtime.Starter.CargoDomains...), ZOC: runtime.Starter.ZOC, Effects: starterEffects(runtime.Starter), Attack: runtime.Starter.Attack, Defense: runtime.Starter.Defense, Ranged: runtime.Starter.Ranged, Strength: runtime.Starter.Strength, Armor: runtime.Starter.Armor, Firepower: runtime.Starter.Firepower, Health: runtime.Starter.Health, MaxHealth: runtime.Starter.Health, VisibilityRange: runtime.Starter.VisibilityRange})
		}
	}
	if len(state.Units) != len(players)*snapshot.Config.StartingSettlers {
		return engine.State{}, fmt.Errorf("initial gameplay state has %d starter units for %d players", len(state.Units), len(players))
	}
	for _, player := range state.Players {
		owned := false
		for _, unit := range state.Units {
			owned = owned || unit.OwnerID == player.ID
		}
		if !owned {
			return engine.State{}, fmt.Errorf("initial gameplay state has no starter unit for %s", player.ID)
		}
	}
	sort.Strings(state.Phase.ActivePlayerOrder)
	if err := engine.ValidateState(state); err != nil {
		return engine.State{}, err
	}
	return state, nil
}

func tileID(x, y int) string { return fmt.Sprintf("tile-%d-%d", x, y) }

func engineTerrainForRuntime(terrain string, runtime rules.RuntimeRules) string {
	if _, ok := runtime.Movement.Terrain("TERRAIN_PLAINS"); ok {
		return retailEngineTerrain(terrain, runtime)
	}
	return engineTerrain(terrain)
}

func retailEngineTerrain(terrain string, runtime rules.RuntimeRules) string {
	candidates := map[string]string{
		"plains":    "TERRAIN_PLAINS",
		"grassland": "TERRAIN_GRASSLAND",
		"forest":    "TERRAIN_FOREST",
		"swamp":     "TERRAIN_SWAMP",
		"tundra":    "TERRAIN_TUNDRA",
		"shelf":     "TERRAIN_WATER_SHELF",
		"ocean":     "TERRAIN_WATER_DEEP",
	}
	if candidate, ok := candidates[terrain]; ok {
		if _, exists := runtime.Movement.Terrain(candidate); exists {
			return candidate
		}
	}
	return "TERRAIN_PLAINS"
}

func engineTerrain(terrain string) string {
	switch terrain {
	case engine.TerrainOcean, "shelf":
		return engine.TerrainOcean
	case engine.TerrainForest:
		return engine.TerrainForest
	default:
		return engine.TerrainPlains
	}
}

func starterEffects(starter rules.RuntimeStarter) []engine.Effect {
	effects := make([]engine.Effect, len(starter.Effects))
	for index, effect := range starter.Effects {
		effects[index] = engine.Effect{ID: effect.ID, Kind: effect.Kind, Value: effect.Value}
	}
	return effects
}

func startController(value string) (string, error) {
	switch value {
	case "", engine.ControllerHuman:
		return engine.ControllerHuman, nil
	case engine.ControllerAI:
		return engine.ControllerAI, nil
	default:
		return "", fmt.Errorf("unsupported persisted controller kind %q", value)
	}
}

func createStartedGameplay(tx *sqlx.Tx, snapshot world.WorldSnapshot, game *Game, players []GamePlayer, sources engine.Sources) error {
	if err := loadStartControllers(tx, players); err != nil {
		return err
	}
	pin, err := serverRuntimePinForSetupRules(game.RulesJSON)
	if err != nil {
		return fmt.Errorf("select server runtime rules: %w", err)
	}
	rules := []byte(pin.Definitions)
	state, err := initialGameplayState(snapshot, game.Style, pin, players)
	if err != nil {
		return fmt.Errorf("derive initial gameplay state: %w", err)
	}
	initialJSON, err := engine.CanonicalStateJSON(state)
	if err != nil {
		return fmt.Errorf("canonical initial gameplay state: %w", err)
	}
	initialRoundTrip, err := engine.DecodeCanonicalStateJSON(initialJSON)
	if err != nil || len(initialRoundTrip.Units) != len(state.Units) {
		return fmt.Errorf("initial gameplay canonical round trip lost starter units")
	}
	if err := world.CreateV1InTx(tx, snapshot, world.RulesetPin{CanonicalJSON: rules, ContentHash: pin.ContentHash, Version: pin.Version}, initialJSON); err != nil {
		return err
	}
	next, events, result, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: "game-start", ActorPlayerID: state.Phase.ActivePlayerOrder[0], Type: engine.CommandNoop}, pin, sources)
	if err != nil {
		return fmt.Errorf("initialize gameplay event: %w", err)
	}
	nextJSON, err := engine.CanonicalStateJSON(next)
	if err != nil {
		return err
	}
	nextRoundTrip, err := engine.DecodeCanonicalStateJSON(nextJSON)
	if err != nil || len(nextRoundTrip.Units) != len(next.Units) {
		return fmt.Errorf("game-start canonical round trip lost starter units")
	}
	payload, err := json.Marshal(events[0])
	if err != nil {
		return err
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	commandHash := sha256.Sum256([]byte("game-start"))
	return gameplaystore.PersistCommandInTx(tx, 0, gameplaystore.Aggregate{GameID: snapshot.GameID, StateSchemaVersion: gameplaystore.StateSchemaV1, Revision: next.Revision, LastEventSequence: next.EventSequence, CanonicalStateJSON: nextJSON, Ruleset: gameplaystore.RulesetPin{CanonicalJSON: rules, ContentHash: pin.ContentHash, Version: pin.Version}}, gameplaystore.CommandReceipt{GameID: snapshot.GameID, ActorPlayerID: players[0].ID, IdempotencyKey: "game-start", PayloadSHA256: hex.EncodeToString(commandHash[:]), ResultJSON: resultJSON, ResultingRevision: next.Revision, ResultingEventSequence: next.EventSequence}, []gameplaystore.Event{{GameID: snapshot.GameID, Sequence: events[0].Sequence, Revision: next.Revision, Ordinal: events[0].Ordinal, ID: events[0].ID, Type: events[0].Type, PayloadJSON: payload}})
}
