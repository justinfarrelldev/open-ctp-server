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
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
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

func mustStateJSON(t *testing.T, state engine.State) []byte {
	t.Helper()
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
