package fullgame

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestRetailRuntimeScoreVictoryCanReachTerminalFinalTurn(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := rules.LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	runtime := rules.ProjectRetailRuntime(db)
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatalf("CanonicalRuntimeDefinitions() error = %v", err)
	}
	state := engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: runtime.CompatibilityVersion, ContentHash: contentHash, Definitions: definitions},
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: runtime.Calendar.FinalTurn - 1, ActivePlayerOrder: []string{"ai-1", "ai-2"}},
		Players: []engine.Player{
			{ID: "ai-1", NationID: "nation-ai-1", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
			{ID: "ai-2", NationID: "nation-ai-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
		},
		Nations: []engine.Nation{{ID: "nation-ai-1"}, {ID: "nation-ai-2"}},
		Tiles:   []engine.Tile{{ID: "tile-1", Terrain: "TERRAIN_PLAINS", Open: true}, {ID: "tile-2", Terrain: "TERRAIN_PLAINS", Open: true}},
		Cities: []engine.City{
			{ID: "city-1", OwnerID: "ai-1", TileID: "tile-1", Effects: []engine.Effect{{ID: "score-1", Kind: engine.EffectVictory, Value: 2}}},
			{ID: "city-2", OwnerID: "ai-2", TileID: "tile-2", Effects: []engine.Effect{{ID: "score-2", Kind: engine.EffectVictory, Value: 1}}},
		},
		RNG: []engine.RNGDomainState{{Domain: engine.RNGDomainCombat, State: 1}, {Domain: engine.RNGDomainSpecialAction, State: 2}, {Domain: engine.RNGDomainAI, State: 3}, {Domain: engine.RNGDomainRiot, State: 4}},
	}
	if err := engine.ValidateState(state); err != nil {
		t.Fatalf("terminal proof state invalid: %v", err)
	}
	next, events, _, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: "retail-final-turn", ActorPlayerID: "ai-1", Type: engine.CommandRoundComplete, Payload: json.RawMessage(`{}`)}, state.Rules, retailAISources())
	if err != nil {
		t.Fatalf("round complete at retail final turn: %v", err)
	}
	if !next.Terminal.Ended || next.Terminal.Reason != "score_2300" || next.Terminal.WinningPlayerID != "ai-1" || len(events) != 1 || events[0].Type != engine.EventTerminal {
		t.Fatalf("retail final turn terminal = %#v events=%#v", next.Terminal, events)
	}
}

func TestRetailAIAutoPlayRecordsAtLeastHundredTurns(t *testing.T) {
	root := localRetailGameDataRoot(t)
	transcript := filepath.Join(t.TempDir(), "retail-ai-100-turns.jsonl")
	state, records := runRetailAIAutoPlay(t, root, 100, transcript)
	if state.Phase.Turn < 100 {
		t.Fatalf("autoplay reached turn %d, want at least 100", state.Phase.Turn)
	}
	if len(records) < 100 {
		t.Fatalf("transcript records = %d, want at least 100", len(records))
	}
	if stat, err := os.Stat(transcript); err != nil || stat.Size() == 0 {
		t.Fatalf("transcript %q stat=%+v err=%v", transcript, stat, err)
	}
	if explicit := os.Getenv("OPEN_CTP_RECORD_RETAIL_AI_TURNS"); explicit != "" {
		writeRetailAITranscript(t, explicit, records)
	}
}

type retailAITurnRecord struct {
	Turn             int      `json:"turn"`
	ActivePlayerID   string   `json:"active_player_id"`
	Commands         []string `json:"commands"`
	Cities           int      `json:"cities"`
	Units            int      `json:"units"`
	Advances         int      `json:"advances"`
	Wonders          int      `json:"wonders"`
	TilePollution    int      `json:"tile_pollution"`
	DiplomacyRecords int      `json:"diplomacy_records"`
	StateHash        string   `json:"state_hash"`
}

func runRetailAIAutoPlay(t *testing.T, root string, targetTurn int, transcript string) (engine.State, []retailAITurnRecord) {
	t.Helper()
	state := retailAIInitialState(t, root)
	records := make([]retailAITurnRecord, 0, targetTurn)
	for len(records) < targetTurn && !state.Terminal.Ended {
		actor := state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex]
		commands := make([]string, 0, 8)
		for action := 0; action < 8 && !state.Terminal.Ended; action++ {
			candidate, ok := nextRetailAICandidate(t, state, actor)
			if !ok {
				break
			}
			next, ok := applyRetailAICommand(t, state, actor, candidate, len(records), action)
			if !ok {
				break
			}
			state = next
			commands = append(commands, candidate.Type)
		}
		state = applyRetailAIRoundComplete(t, state, actor, len(records))
		commands = append(commands, engine.CommandRoundComplete)
		records = append(records, retailAITurnSummary(t, state, actor, commands))
	}
	writeRetailAITranscript(t, transcript, records)
	return state, records
}

func nextRetailAICandidate(t *testing.T, state engine.State, actor string) (ai.Candidate, bool) {
	t.Helper()
	projection, err := visibility.Project(state, actor)
	if err != nil {
		t.Fatalf("Project(%s): %v", actor, err)
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := ai.Plan(encoded, actor)
	if err != nil {
		t.Fatalf("Plan(%s): %v", actor, err)
	}
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandPass || candidate.Type == engine.CommandRoundComplete || candidate.Type == engine.CommandFortify {
			continue
		}
		return candidate, true
	}
	return ai.Candidate{}, false
}

func applyRetailAICommand(t *testing.T, state engine.State, actor string, candidate ai.Candidate, turnIndex, actionIndex int) (engine.State, bool) {
	t.Helper()
	command := engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: fmt.Sprintf("retail-ai-%03d-%02d-%s", turnIndex, actionIndex, candidate.Type), ActorPlayerID: actor, Type: candidate.Type, Payload: candidate.Payload}
	next, _, _, err := engine.Apply(state, command, state.Rules, retailAISources())
	if err != nil {
		return state, false
	}
	return next, true
}

func applyRetailAIRoundComplete(t *testing.T, state engine.State, actor string, turnIndex int) engine.State {
	t.Helper()
	payload := json.RawMessage(`{}`)
	command := engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: fmt.Sprintf("retail-ai-%03d-round", turnIndex), ActorPlayerID: actor, Type: engine.CommandRoundComplete, Payload: payload}
	next, _, _, err := engine.Apply(state, command, state.Rules, retailAISources())
	if err != nil {
		t.Fatalf("round complete turn=%d actor=%s: %v", state.Phase.Turn, actor, err)
	}
	return next
}

func retailAIInitialState(t *testing.T, root string) engine.State {
	t.Helper()
	db, err := rules.LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	runtime := rules.ProjectRetailRuntime(db)
	definitions, contentHash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatalf("CanonicalRuntimeDefinitions() error = %v", err)
	}
	pin := engine.RulesPin{Version: runtime.CompatibilityVersion, ContentHash: contentHash, Definitions: definitions}
	players := []string{"ai-1", "ai-2", "ai-3", "ai-4"}
	state := engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         pin,
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerOrder: players},
		Players: []engine.Player{
			{ID: "ai-1", NationID: "nation-ai-1", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
			{ID: "ai-2", NationID: "nation-ai-2", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
			{ID: "ai-3", NationID: "nation-ai-3", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
			{ID: "ai-4", NationID: "nation-ai-4", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI},
		},
		Nations: []engine.Nation{{ID: "nation-ai-1", PublicWorks: 10000}, {ID: "nation-ai-2", PublicWorks: 10000}, {ID: "nation-ai-3", PublicWorks: 10000}, {ID: "nation-ai-4", PublicWorks: 10000}},
		RNG:     []engine.RNGDomainState{{Domain: engine.RNGDomainCombat, State: 1}, {Domain: engine.RNGDomainSpecialAction, State: 2}, {Domain: engine.RNGDomainAI, State: 3}, {Domain: engine.RNGDomainRiot, State: 4}},
	}
	for y := 0; y < 6; y++ {
		for x := 0; x < 6; x++ {
			terrain := "TERRAIN_PLAINS"
			if (x+y)%5 == 0 {
				terrain = "TERRAIN_FOREST"
			}
			state.Tiles = append(state.Tiles, engine.Tile{ID: fmt.Sprintf("tile-%d-%d", x, y), X: x, Y: y, Terrain: terrain, Open: true})
		}
	}
	spawns := []struct{ player, tile string }{{"ai-1", "tile-0-0"}, {"ai-2", "tile-5-0"}, {"ai-3", "tile-0-5"}, {"ai-4", "tile-5-5"}}
	for index, spawn := range spawns {
		state.Units = append(state.Units, engine.Unit{ID: fmt.Sprintf("settler-%d", index+1), OwnerID: spawn.player, TileID: spawn.tile, Domain: runtime.Starter.Domain, Movement: runtime.Starter.Movement, Fuel: runtime.Starter.MaxFuel, MaxFuel: runtime.Starter.MaxFuel, CargoCapacity: runtime.Starter.TransportCapacity, CargoDomains: append([]string(nil), runtime.Starter.CargoDomains...), ZOC: runtime.Starter.ZOC, Effects: starterEffects(runtime.Starter), Attack: runtime.Starter.Attack, Defense: runtime.Starter.Defense, Ranged: runtime.Starter.Ranged, Strength: runtime.Starter.Strength, Armor: runtime.Starter.Armor, Health: runtime.Starter.Health, MaxHealth: runtime.Starter.Health, VisibilityRange: runtime.Starter.VisibilityRange})
	}
	if err := engine.ValidateState(state); err != nil {
		t.Fatalf("initial state invalid: %v", err)
	}
	return state
}

func retailAITurnSummary(t *testing.T, state engine.State, actor string, commands []string) retailAITurnRecord {
	t.Helper()
	hash, err := engine.CanonicalStateHash(state)
	if err != nil {
		t.Fatal(err)
	}
	advanceCount := 0
	for _, nation := range state.Nations {
		advanceCount += len(nation.Advances)
	}
	pollution := 0
	for _, tile := range state.Tiles {
		pollution += tile.Pollution
	}
	return retailAITurnRecord{Turn: state.Phase.Turn, ActivePlayerID: actor, Commands: append([]string(nil), commands...), Cities: len(state.Cities), Units: len(state.Units), Advances: advanceCount, Wonders: len(state.Wonders), TilePollution: pollution, DiplomacyRecords: len(state.Diplomacy), StateHash: hash}
}

func writeRetailAITranscript(t *testing.T, path string, records []retailAITurnRecord) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("retail AI transcript: %s", path)
}

func starterEffects(starter rules.RuntimeStarter) []engine.Effect {
	effects := make([]engine.Effect, len(starter.Effects))
	for index, effect := range starter.Effects {
		effects[index] = engine.Effect{ID: effect.ID, Kind: effect.Kind, Value: effect.Value}
	}
	return effects
}

func localRetailGameDataRoot(t *testing.T) string {
	t.Helper()
	candidates := []string{
		filepath.Join("..", "..", "..", "..", "open-ctp-server", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"),
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

func retailAISources() engine.Sources {
	return engine.Sources{Clock: retailAIClock{}, IDs: &retailAIIDs{}, RNG: retailAIRNG{}}
}

type retailAIClock struct{}
type retailAIIDs struct{}
type retailAIRNG struct{}

func (retailAIClock) Now() time.Time { return time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC) }
func (*retailAIIDs) Next() string    { return "retail-ai-event" }
func (retailAIRNG) Next(_ string, state uint64) (uint64, uint64) {
	next := state + 0x9e3779b97f4a7c15
	return next, next
}
