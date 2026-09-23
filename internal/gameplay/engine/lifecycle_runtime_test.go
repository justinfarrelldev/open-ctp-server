package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRuntimeRulesNaturalVictoryTransitions(t *testing.T) {
	tests := []struct {
		name         string
		run          func(t *testing.T, state State) State
		want         string
		startingTurn int
	}{
		{
			name: "conquest follows combat elimination",
			want: "bloodlust_conquest",
			run: func(t *testing.T, state State) State {
				state.Units = []Unit{
					{ID: "attacker", OwnerID: "player-a", TileID: "tile-a", ArmyID: "army-a", Domain: DomainLand, Movement: 2, Attack: 100, Defense: 100, Strength: 100, Armor: 100, Health: 10, MaxHealth: 10},
					{ID: "defender", OwnerID: "player-b", TileID: "tile-b", ArmyID: "army-b", Domain: DomainLand, Movement: 2, Attack: 1, Defense: 1, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1},
				}
				state.Armies = []Army{{ID: "army-a", OwnerID: "player-a", UnitIDs: []string{"attacker"}}, {ID: "army-b", OwnerID: "player-b", UnitIDs: []string{"defender"}}}
				state.Fog = []FogObservation{{PlayerID: "player-a", TileID: "tile-b"}}
				state = runtimeApply(t, state, "player-a", CommandCombat, combatPayload{AttackerArmyID: "army-a", DefenderArmyID: "army-b"})
				return runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
			},
		},
		{
			name: "world union follows reciprocal embassies",
			want: "diplomatic_world_union",
			run: func(t *testing.T, state State) State {
				state.Units = []Unit{
					{ID: "envoy-a", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "embassy", Kind: EffectSpecialEmbassy}}},
					{ID: "envoy-b", OwnerID: "player-b", TileID: "tile-b", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "embassy", Kind: EffectSpecialEmbassy}}},
				}
				state.Cities = []City{{ID: "city-a", Name: "A", OwnerID: "player-a", TileID: "tile-a"}, {ID: "city-b", Name: "B", OwnerID: "player-b", TileID: "tile-b"}}
				state = runtimeApply(t, state, "player-a", CommandEstablishEmbassy, SpecialActionPayload{UnitID: "envoy-a", CityID: "city-b"})
				state = runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
				state = runtimeApply(t, state, "player-b", CommandEstablishEmbassy, SpecialActionPayload{UnitID: "envoy-b", CityID: "city-a"})
				return runtimeApply(t, state, "player-b", CommandRoundComplete, struct{}{})
			},
		},
		{
			name:         "calendar score follows founded cities",
			want:         "score_2300",
			startingTurn: 46,
			run: func(t *testing.T, state State) State {
				state.Units = []Unit{
					{ID: "settler-a", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "settler", Kind: EffectSettler}}},
					{ID: "settler-b", OwnerID: "player-b", TileID: "tile-b", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "settler", Kind: EffectSettler}}},
				}
				state = runtimeApply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "A"})
				state = runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
				state = runtimeApply(t, state, "player-b", CommandFoundCity, FoundCityPayload{UnitID: "settler-b", CityID: "city-b", Name: "B"})
				return runtimeApply(t, state, "player-b", CommandRoundComplete, struct{}{})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startingTurn := tt.startingTurn
			if startingTurn == 0 {
				startingTurn = 1
			}
			next := tt.run(t, runtimeStartedStateAt(t, startingTurn))
			if !next.Terminal.Ended || next.Terminal.Reason != tt.want || next.Victory.Reason != tt.want {
				t.Fatalf("terminal victory = %#v / %#v, want %q", next.Terminal, next.Victory, tt.want)
			}
		})
	}
}

func TestTask6ResearchCompletesAdvancesAndUnlocksProduction(t *testing.T) {
	state := runtimeStartedState(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].Effects = []Effect{{ID: "research-lab", Kind: EffectResearch, Value: 1}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}, {ID: "city-b", OwnerID: "player-b", TileID: "tile-b"}}
	selected := runtimeApply(t, state, "player-a", CommandSetResearch, map[string]any{"advance_id": "ADVANCE_SYNTHETIC_ROOT"})
	nation := nationForID(&selected, "nation-a")
	if nation == nil || nation.ResearchTargetID != "ADVANCE_SYNTHETIC_ROOT" || nation.ResearchProgress != 0 {
		t.Fatalf("selected research nation = %#v", nation)
	}
	if err := runtimeApplyExpectError(t, selected, "player-a", CommandQueueProduction, map[string]any{"city_id": "city-a", "item": ProductionItem{ID: "UNIT_UNLOCKED", Kind: ProductionUnit, Cost: 1, Prerequisite: "ADVANCE_SYNTHETIC_ROOT"}}); err == nil || !strings.Contains(err.Error(), "prerequisite") {
		t.Fatalf("production before research error = %v, want prerequisite rejection", err)
	}
	runtimeGate := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_RUNTIME_LOCKED", Prerequisites: []string{"ADVANCE_SYNTHETIC_ROOT"}}}}
	locked := selected
	if err := queueProduction(&locked, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_RUNTIME_LOCKED", Kind: ProductionUnit, Cost: 1}}, runtimeGate); err == nil || !strings.Contains(err.Error(), "prerequisite") {
		t.Fatalf("runtime production before research error = %v, want prerequisite rejection", err)
	}
	afterRound := runtimeApply(t, selected, "player-a", CommandRoundComplete, map[string]any{})
	nation = nationForID(&afterRound, "nation-a")
	if nation == nil || !containsString(nation.Advances, "ADVANCE_SYNTHETIC_ROOT") || nation.ResearchTargetID != "ADVANCE_SYNTHETIC_UNLOCK" {
		t.Fatalf("researched nation = %#v", nation)
	}
	unlocked := afterRound
	if err := queueProduction(&unlocked, "player-a", QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_RUNTIME_LOCKED", Kind: ProductionUnit, Cost: 1}}, runtimeGate); err != nil {
		t.Fatalf("runtime production after research error = %v", err)
	}
	queued := runtimeApply(t, afterRound, "player-a", CommandQueueProduction, map[string]any{"city_id": "city-a", "item": ProductionItem{ID: "UNIT_UNLOCKED", Kind: ProductionUnit, Cost: 1, Prerequisite: "ADVANCE_SYNTHETIC_ROOT"}})
	city, ok := cityByID(queued, "city-a")
	if !ok || !queuedItem(city.Queue, "UNIT_UNLOCKED") {
		t.Fatalf("queued city = %#v", city)
	}
}

func TestRuntimeRulesRejectInvalidResearchSelection(t *testing.T) {
	state := runtimeStartedState(t)
	for _, tc := range []struct {
		name      string
		advanceID string
		want      string
	}{
		{name: "unknown", advanceID: "ADVANCE_UNKNOWN", want: "unsupported"},
		{name: "missing prerequisite", advanceID: "ADVANCE_SYNTHETIC_UNLOCK", want: "prerequisite"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := runtimeApplyExpectError(t, state, "player-a", CommandSetResearch, map[string]any{"advance_id": tc.advanceID})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("set research error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestRuntimeRulesGaiaRequiresCountdownInsteadOfAdvances(t *testing.T) {
	state := runtimeStartedStateWithGaiaTurns(t, 1, 2)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].Advances = []string{"alpha", "beta", "gaia-science"}

	next := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if next.Terminal.Ended || next.Victory.Reason != "" {
		t.Fatalf("advances-only state triggered Gaia victory: %#v / %#v", next.Terminal, next.Victory)
	}
	if got := next.Nations[0].GaiaCompletionTurn; got != 0 {
		t.Fatalf("advances-only countdown = %d, want 0", got)
	}
}

func TestRuntimeRulesGaiaPrerequisitesCanComeFromRetailRuntime(t *testing.T) {
	runtime := &rules.RuntimeRules{Gaia: rules.RuntimeGaia{ActivationTurns: 10, BuildingID: "IMPROVE_GAIA_COMPUTER", WonderID: "WONDER_THE_SOLARIS_PROJECT"}}
	buildingID, wonderID, ok := gaiaPrerequisiteIDs(runtime)
	if !ok || buildingID != "IMPROVE_GAIA_COMPUTER" || wonderID != "WONDER_THE_SOLARIS_PROJECT" {
		t.Fatalf("Gaia prerequisites = %q/%q/%v", buildingID, wonderID, ok)
	}
	state := runtimeStartedState(t)
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_GAIA_COMPUTER"}}}}
	state.Wonders = []Wonder{{ID: "WONDER_THE_SOLARIS_PROJECT", OwnerID: "player-a", CityID: "city-a"}}
	updateGaiaCountdown(&state, runtime, 5)
	if got := state.Nations[0].GaiaCompletionTurn; got != 15 {
		t.Fatalf("retail Gaia completion turn = %d, want 15", got)
	}
}

func TestRuntimeRulesGaiaStartsFromSyntheticBuildingAndWonder(t *testing.T) {
	state := runtimeStartedStateWithGaiaTurns(t, 1, 2)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Cities = []City{{
		ID:      "city-a",
		Name:    "A",
		OwnerID: "player-a",
		TileID:  "tile-a",
		Queue: []ProductionItem{
			{ID: "IMPROVE_SYNTHETIC", Kind: ProductionBuilding, Cost: 1},
			{ID: "WONDER_SYNTHETIC", Kind: ProductionWonder, Scope: "global", Cost: 1},
		},
	}, {
		ID:      "city-b",
		Name:    "B",
		OwnerID: "player-b",
		TileID:  "tile-b",
	}}

	afterBuilding := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if afterBuilding.Terminal.Ended {
		t.Fatalf("building-only step ended game: %#v", afterBuilding.Terminal)
	}
	if got := afterBuilding.Nations[0].GaiaCompletionTurn; got != 0 {
		t.Fatalf("building-only countdown = %d, want 0", got)
	}

	afterWonder := runtimeApply(t, afterBuilding, "player-a", CommandRoundComplete, struct{}{})
	if afterWonder.Terminal.Ended {
		t.Fatalf("countdown start ended game early: %#v", afterWonder.Terminal)
	}
	if got := afterWonder.Nations[0].GaiaCompletionTurn; got != 4 {
		t.Fatalf("countdown completion turn = %d, want 4", got)
	}

	complete := runtimeApply(t, afterWonder, "player-a", CommandRoundComplete, struct{}{})
	if !complete.Terminal.Ended || complete.Terminal.Reason != "science_gaia_controller" || complete.Victory.Reason != "science_gaia_controller" {
		t.Fatalf("Gaia completion victory = %#v / %#v", complete.Terminal, complete.Victory)
	}
}

func TestRuntimeRulesGaiaCountdownCancelsAndRestartsWhenPrerequisitesBreak(t *testing.T) {
	state := runtimeStartedStateWithGaiaTurns(t, 1, 2)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Cities = []City{{ID: "city-a", Name: "A", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_SYNTHETIC"}}}, {ID: "city-b", Name: "B", OwnerID: "player-b", TileID: "tile-b"}}
	state.Wonders = []Wonder{{ID: "WONDER_SYNTHETIC", OwnerID: "player-a", CityID: "city-a", Scope: "global"}}

	counting := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if counting.Terminal.Ended {
		t.Fatalf("countdown start ended game early: %#v", counting.Terminal)
	}
	if got := counting.Nations[0].GaiaCompletionTurn; got != 3 {
		t.Fatalf("initial completion turn = %d, want 3", got)
	}

	counting.Wonders[0].Destroyed = true
	cancelled := runtimeApply(t, counting, "player-a", CommandRoundComplete, struct{}{})
	if cancelled.Terminal.Ended {
		t.Fatalf("destroyed wonder still ended game: %#v", cancelled.Terminal)
	}
	if got := cancelled.Nations[0].GaiaCompletionTurn; got != 0 {
		t.Fatalf("cancelled countdown = %d, want 0", got)
	}

	cancelled.Wonders[0].Destroyed = false
	restarted := runtimeApply(t, cancelled, "player-a", CommandRoundComplete, struct{}{})
	if restarted.Terminal.Ended {
		t.Fatalf("restarted countdown ended game early: %#v", restarted.Terminal)
	}
	if got := restarted.Nations[0].GaiaCompletionTurn; got != 5 {
		t.Fatalf("restarted completion turn = %d, want 5", got)
	}

	complete := runtimeApply(t, restarted, "player-a", CommandRoundComplete, struct{}{})
	if !complete.Terminal.Ended || complete.Terminal.Reason != "science_gaia_controller" {
		t.Fatalf("restarted countdown did not complete Gaia: %#v / %#v", complete.Terminal, complete.Victory)
	}
}

func TestRuntimeRulesGaiaCountdownSurvivesCanonicalRestart(t *testing.T) {
	state := runtimeStartedStateWithGaiaTurns(t, 1, 2)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Cities = []City{{ID: "city-a", Name: "A", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_SYNTHETIC"}}}, {ID: "city-b", Name: "B", OwnerID: "player-b", TileID: "tile-b"}}
	state.Wonders = []Wonder{{ID: "WONDER_SYNTHETIC", OwnerID: "player-a", CityID: "city-a", Scope: "global"}}

	midCountdown := runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if got := midCountdown.Nations[0].GaiaCompletionTurn; got != 3 {
		t.Fatalf("mid-countdown completion turn = %d, want 3", got)
	}

	encoded, err := CanonicalStateJSON(midCountdown)
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := DecodeCanonicalStateJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}

	uninterrupted := runtimeApply(t, midCountdown, "player-a", CommandRoundComplete, struct{}{})
	afterRestart := runtimeApply(t, restarted, "player-a", CommandRoundComplete, struct{}{})
	if got, want := canonicalStateHashForTest(t, afterRestart), canonicalStateHashForTest(t, uninterrupted); got != want {
		t.Fatalf("restarted countdown hash = %s, want %s", got, want)
	}
	if !afterRestart.Terminal.Ended || afterRestart.Terminal.Reason != "science_gaia_controller" {
		t.Fatalf("restarted countdown did not complete Gaia: %#v / %#v", afterRestart.Terminal, afterRestart.Victory)
	}
}

func runtimeStartedStateWithGaiaTurns(t *testing.T, turn, gaiaTurns int) State {
	t.Helper()
	state := runtimeStartedStateAt(t, turn)
	var definition struct {
		CompatibilityVersion       string                       `json:"compatibility_version"`
		Manifest                   rules.SupportedRulesManifest `json:"manifest"`
		Records                    []rules.Record               `json:"records"`
		EnabledVictoryEvaluatorIDs []string                     `json:"enabled_victory_evaluator_ids,omitempty"`
	}
	if err := json.Unmarshal(state.Rules.Definitions, &definition); err != nil {
		t.Fatal(err)
	}
	for i := range definition.Records {
		if definition.Records[i].ID == "ENDGAME_GAIA_VECTOR" {
			definition.Records[i].Tags["TurnsToActivate"] = strconv.Itoa(gaiaTurns)
		}
	}
	canonical, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err = rules.CanonicalDefinitions(canonical)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	state.Rules.Definitions = canonical
	state.Rules.ContentHash = hex.EncodeToString(digest[:])
	return state
}

func canonicalStateHashForTest(t *testing.T, state State) string {
	t.Helper()
	hash, err := CanonicalStateHash(state)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func runtimeStartedState(t *testing.T) State {
	return runtimeStartedStateAt(t, 1)
}

func runtimeStartedStateAt(t *testing.T, turn int) State {
	t.Helper()
	ruleset, err := rules.Import(rules.ImportConfig{Directory: filepath.Join("..", "rules", "testdata", "synthetic-ctp2")})
	if err != nil {
		t.Fatal(err)
	}
	return State{
		SchemaVersion: StateSchemaV1,
		Rules:         RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON},
		Phase:         PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: turn, ActivePlayerOrder: []string{"player-a", "player-b"}},
		Tiles:         []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, Open: true}},
		Players:       []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}, {ID: "player-b", NationID: "nation-b", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Nations:       []Nation{{ID: "nation-a"}, {ID: "nation-b"}},
		RNG:           []RNGDomainState{{Domain: RNGDomainCombat, State: 1}, {Domain: RNGDomainSpecialAction, State: 2}, {Domain: RNGDomainAI, State: 3}},
	}
}

func queuedItem(queue []ProductionItem, id string) bool {
	for _, item := range queue {
		if item.ID == id {
			return true
		}
	}
	return false
}

func runtimeApply(t *testing.T, state State, actor, commandType string, payload any) State {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: commandType + "-" + actor, ActorPlayerID: actor, Type: commandType, Payload: encoded}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Apply(%s): %v", commandType, err)
	}
	return next
}

func runtimeApplyExpectError(t *testing.T, state State, actor, commandType string, payload any) error {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: commandType + "-" + actor, ActorPlayerID: actor, Type: commandType, Payload: encoded}, state.Rules, fixtureSources())
	return err
}
