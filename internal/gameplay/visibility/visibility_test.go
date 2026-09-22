package visibility

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func fixtureState() engine.State {
	return engine.State{
		SchemaVersion: engine.StateSchemaV1,
		Rules:         engine.RulesPin{Version: "ctp2-supported-v2", ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		Revision:      7,
		Players:       []engine.Player{{ID: "a", NationID: "a", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}, {ID: "b", NationID: "b", Lifecycle: engine.PlayerActive, Controller: engine.ControllerHuman}},
		Phase:         engine.PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 4, ActivePlayerOrder: []string{"a", "b"}},
		RNG:           []engine.RNGDomainState{{Domain: engine.RNGDomainAI}, {Domain: engine.RNGDomainCombat}, {Domain: engine.RNGDomainSpecialAction}},
		Tiles:         []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "2,0", X: 2, Y: 0, Terrain: engine.TerrainForest, Open: true}},
		Units:         []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "0,0"}, {ID: "HIDDEN_UNIT_TOKEN", OwnerID: "b", TileID: "2,0"}},
		Cities:        []engine.City{{ID: "HIDDEN_CITY_TOKEN", OwnerID: "b", TileID: "2,0"}},
	}
}

func TestObservationMemoryProjectionAndRawLeakBoundaries(t *testing.T) {
	state, err := Update(fixtureState())
	if err != nil {
		t.Fatal(err)
	}
	first, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Tiles) != 3 || len(first.Units) != 2 || len(first.Cities) != 1 {
		t.Fatalf("visible projection = %#v", first)
	}

	state.Units = state.Units[1:]
	state.Tiles[2].Terrain = engine.TerrainPlains
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	stale, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(stale.Units) != 0 || len(stale.Cities) != 1 || !stale.Tiles[2].Stale || stale.Tiles[2].Terrain != engine.TerrainForest {
		t.Fatalf("stale projection = %#v", stale)
	}

	projectionBytes, err := json.Marshal(stale)
	if err != nil {
		t.Fatal(err)
	}
	projector := NewProjector("a")
	snapshot, err := projector.ProjectSnapshot(events.Viewer{GameID: 1, PlayerID: 1}, mustStateJSON(t, state))
	if err != nil {
		t.Fatal(err)
	}
	event, err := projector.ProjectEvent(events.Viewer{GameID: 1, PlayerID: 1}, events.Event{Type: "combat", Sequence: 8, PayloadJSON: []byte(`{"target":"HIDDEN_UNIT_TOKEN","terrain":"SECRET_TERRAIN_TOKEN"}`)})
	if err != nil {
		t.Fatal(err)
	}
	ai, err := AIInput(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range [][]byte{projectionBytes, snapshot, event, ai, SafeError(errors.New("target HIDDEN_UNIT_TOKEN at SECRET_TERRAIN_TOKEN"))} {
		if bytes.Contains(raw, []byte("HIDDEN_UNIT_TOKEN")) || bytes.Contains(raw, []byte("SECRET_TERRAIN_TOKEN")) {
			t.Fatalf("hidden data leaked: %s", raw)
		}
	}
	if !bytes.Equal(ai, projectionBytes) {
		t.Fatalf("AI input differs from controller projection: %s != %s", ai, projectionBytes)
	}
}

func TestTask6ProjectionExposesOwnedResearchOnly(t *testing.T) {
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state := fixtureState()
	state.Rules = engine.RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
	state.Nations = []engine.Nation{
		{ID: "a", Advances: []string{"ADVANCE_SYNTHETIC_ROOT"}, ResearchTargetID: "ADVANCE_SYNTHETIC_UNLOCK", ResearchProgress: 1},
		{ID: "b", Advances: []string{"ADVANCE_SECRET"}, ResearchTargetID: "ADVANCE_SECRET_NEXT", ResearchProgress: 99},
	}
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projection.Research == nil || projection.Research.TargetID != "ADVANCE_SYNTHETIC_UNLOCK" || projection.Research.Progress != 1 || !reflect.DeepEqual(projection.Research.Advances, []string{"ADVANCE_SYNTHETIC_ROOT"}) || !reflect.DeepEqual(projection.Research.Available, []string{"ADVANCE_SYNTHETIC_UNLOCK"}) {
		t.Fatalf("research projection = %#v", projection.Research)
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("ADVANCE_SECRET")) {
		t.Fatalf("opponent research leaked: %s", encoded)
	}
}

func TestTask5ProjectionExposesOwnedUnitCapabilitiesOnly(t *testing.T) {
	state := fixtureState()
	state.Units = []engine.Unit{
		{ID: "owned", OwnerID: "a", TileID: "0,0", Domain: engine.DomainLand, Movement: 3, CargoCapacity: 2, CargoDomains: []string{engine.DomainLand}, ZOC: true, Attack: 7, Defense: 5, Strength: 6, Armor: 2, Health: 8, MaxHealth: 10, VisibilityRange: 3, Effects: []engine.Effect{{ID: "settler", Kind: engine.EffectSettler}}},
		{ID: "visible-enemy", OwnerID: "b", TileID: "1,0", Domain: engine.DomainSea, Movement: 9, Attack: 99, Defense: 99, Strength: 99, Armor: 9, Health: 9, MaxHealth: 9},
	}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Units) != 2 {
		t.Fatalf("projection units = %#v", projection.Units)
	}
	owned := projection.Units[0]
	if owned.ID != "owned" || owned.Domain != engine.DomainLand || owned.Movement != 3 || owned.CargoCapacity != 2 || !owned.ZOC || owned.Attack != 7 || owned.Defense != 5 || owned.Strength != 6 || owned.Armor != 2 || owned.Health != 8 || owned.MaxHealth != 10 || owned.VisibilityRange != 3 || len(owned.Effects) != 1 {
		t.Fatalf("owned projection capabilities = %#v", owned)
	}
	enemy := projection.Units[1]
	if enemy.ID != "visible-enemy" || enemy.Domain != "" || enemy.Movement != 0 || enemy.Attack != 0 || enemy.Health != 0 || len(enemy.Effects) != 0 {
		t.Fatalf("enemy projection leaked capabilities = %#v", enemy)
	}
}

func TestTask5UnitVisibilityRangeExpandsOwnedSight(t *testing.T) {
	state := fixtureState()
	state.Tiles = []engine.Tile{{ID: "0,0", X: 0, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "1,0", X: 1, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "2,0", X: 2, Y: 0, Terrain: engine.TerrainPlains, Open: true}, {ID: "3,0", X: 3, Y: 0, Terrain: engine.TerrainPlains, Open: true}}
	state.Units = []engine.Unit{{ID: "scout", OwnerID: "a", TileID: "0,0", VisibilityRange: 3}}
	state.Cities = []engine.City{{ID: "far-city", OwnerID: "b", TileID: "3,0"}}
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Tiles) != 4 || len(projection.Cities) != 1 || projection.Cities[0].ID != "far-city" {
		t.Fatalf("visibility range projection = %#v", projection)
	}
}

func TestVisibilityBlocksSightAndCanonicalFogSurvivesRestart(t *testing.T) {
	state := fixtureState()
	state.Tiles[1].Terrain = engine.TerrainForest
	state, err := Update(state)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Tiles) != 2 || len(projection.Units) != 1 {
		t.Fatalf("blocking projection = %#v", projection)
	}
	encoded := mustStateJSON(t, state)
	restarted, err := engine.DecodeCanonicalStateJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := restarted.Fog, state.Fog; !reflect.DeepEqual(got, want) {
		t.Fatalf("fog restart mismatch: %#v != %#v", got, want)
	}
}

func TestSSEResyncAndTerminalUseViewerProjection(t *testing.T) {
	state, err := Update(fixtureState())
	if err != nil {
		t.Fatal(err)
	}
	state.Units = state.Units[1:]
	state, err = Update(state)
	if err != nil {
		t.Fatal(err)
	}
	state.Terminal = engine.TerminalState{Ended: true, Reason: "complete", WinningPlayerID: "a"}
	state.Revision = 8
	canonical := mustStateJSON(t, state)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stream := events.NewStore(sqlx.NewDb(db, "sqlmock"))
	projector := NewProjector("a")
	viewer := events.Viewer{GameID: 9, PlayerID: 1}
	mock.ExpectQuery("SELECT min\\(sequence\\)").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"min"}).AddRow(5))
	mock.ExpectQuery("SELECT revision, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"revision", "canonical_state_json"}).AddRow(8, canonical))
	items, err := stream.Resume(context.Background(), viewer, 1, projector)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "resync" || bytes.Contains(items[0].Data, []byte("HIDDEN_UNIT_TOKEN")) {
		t.Fatalf("unsafe resync: %#v", items)
	}
	var resync struct {
		CurrentRevision int64           `json:"current_revision"`
		Snapshot        json.RawMessage `json:"snapshot"`
	}
	if err := json.Unmarshal(items[0].Data, &resync); err != nil {
		t.Fatal(err)
	}
	var snapshot Projection
	if err := json.Unmarshal(resync.Snapshot, &snapshot); err != nil || resync.CurrentRevision != snapshot.Revision || snapshot.Revision != state.Revision {
		t.Fatalf("terminal resync snapshot = %#v, projection = %#v, err = %v", resync, snapshot, err)
	}
	mock.ExpectQuery("SELECT last_event_sequence, canonical_state_json").WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"last_event_sequence", "canonical_state_json"}).AddRow(8, canonical))
	terminal, ended, err := stream.Terminal(context.Background(), viewer, projector)
	if err != nil || !ended || bytes.Contains(terminal.Data, []byte("HIDDEN_UNIT_TOKEN")) {
		t.Fatalf("unsafe terminal: %#v %v %v", terminal, ended, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectionExposesOnlyCurrentActorPlanningStatusInPlanningModes(t *testing.T) {
	queuedCommand := func(actor string) engine.TurnPolicyCommand {
		return engine.TurnPolicyCommand{Kind: turns.CommandAction, ActorID: actor, CommandID: "command-" + actor, IdempotencyKey: "receipt-" + actor, Payload: json.RawMessage(`{}`)}
	}
	projectionJSON := func(t *testing.T, viewer string, state engine.State) map[string]any {
		t.Helper()
		state, err := Update(state)
		if err != nil {
			t.Fatal(err)
		}
		encoded := mustStateJSON(t, state)
		restarted, err := engine.DecodeCanonicalStateJSON(encoded)
		if err != nil {
			t.Fatal(err)
		}
		projection, err := Project(restarted, viewer)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(projection)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	}
	tests := []struct {
		name    string
		viewer  string
		mutate  func(engine.State) engine.State
		present bool
		want    string
	}{
		{name: "sequential omits status", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeSequential
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}},
		{name: "long-turn omits status", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeLongTurn
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}},
		{name: "simultaneous queued", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}, present: true, want: "queued"},
		{name: "closed beats queued", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}, present: true, want: "closed"},
		{name: "other actors keep eligible viewer open", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("b")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}, present: true, want: "open"},
		{name: "hybrid planning open", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeHybrid
			state.Policy.Stage = turns.StagePlanning
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("b")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}, present: true, want: "open"},
		{name: "hybrid conflict omits status", mutate: func(state engine.State) engine.State {
			state.Phase.Mode = turns.ModeHybrid
			state.Policy.Stage = turns.StageConflict
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("a")}
			state.Policy.EarlyEnds = []string{"a"}
			return state
		}},
		{name: "unknown viewer omits status", viewer: "viewer-is-not-a-player", mutate: func(state engine.State) engine.State {
			state.Policy.Queue = []engine.TurnPolicyCommand{queuedCommand("b")}
			state.Policy.EarlyEnds = []string{"b"}
			return state
		}},
		{name: "terminal omits status", mutate: func(state engine.State) engine.State {
			state.Terminal.Ended = true
			state.Terminal.Reason = "conquest"
			state.Terminal.WinningPlayerID = "a"
			return state
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := fixtureState()
			state.Phase.Mode = turns.ModeSimultaneous
			state.Phase.ModeVersion = turns.ContractVersion
			state.Policy.Stage = turns.StagePlanning
			viewer := tc.viewer
			if viewer == "" {
				viewer = "a"
			}
			decoded := projectionJSON(t, viewer, tc.mutate(state))
			value, ok := decoded["planning_status"]
			if tc.present {
				if !ok {
					t.Fatalf("planning_status omitted from projection JSON: %#v", decoded)
				}
				if got := value.(string); got != tc.want {
					t.Fatalf("planning_status = %q, want %q", got, tc.want)
				}
				return
			}
			if ok {
				t.Fatalf("planning_status = %#v, want omitted; projection = %#v", value, decoded)
			}
		})
	}
}

func mustStateJSON(t *testing.T, state engine.State) []byte {
	t.Helper()
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
