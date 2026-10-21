package visibility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
)

func TestClientHybridCloseoutAndWaitingActionsArePhaseCorrect(t *testing.T) {
	for _, actor := range []string{"a", "b"} {
		t.Run(actor, func(t *testing.T) {
			for _, status := range []string{"queued", "closed", "locked", "left"} {
				t.Run(status, func(t *testing.T) {
					state := clientMetadataState()
					state.Phase.Mode = turns.ModeHybrid
					state.Policy.Stage = turns.StagePlanning
					switch status {
					case "queued":
						state.Policy.Queue = []engine.TurnPolicyCommand{{Kind: turns.CommandAction, ActorID: actor, CommandID: "pending", IdempotencyKey: "key", Payload: []byte(`{}`)}}
					case "closed":
						state.Policy.EarlyEnds = []string{actor}
					case "locked":
						state.Phase.ResolutionLocked = true
					case "left":
						for index := range state.Players {
							if state.Players[index].ID == actor {
								state.Players[index].Lifecycle = engine.PlayerLeft
							}
						}
					}
					projection, err := Project(state, actor)
					if err != nil {
						t.Fatal(err)
					}
					if status == "queued" {
						if len(projection.LegalActions) != 1 || projection.LegalActions[0].Command != engine.CommandRoundComplete {
							t.Fatalf("queued actor must only be able to close planning: %+v", projection.LegalActions)
						}
					} else if len(projection.LegalActions) != 0 {
						t.Fatalf("%s actor received executable actions: %+v", status, projection.LegalActions)
					}
				})
			}
		})
	}
}

type topologyClock struct{}

func (topologyClock) Now() time.Time { return time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC) }

type topologyIDs struct{ next int }

func (s *topologyIDs) Next() string { s.next++; return fmt.Sprintf("topology-%d", s.next) }

type topologyRNG struct{}

func (topologyRNG) Next(_ string, state uint64) (uint64, uint64) { return state + 1, state + 1 }

func TestClientHybridPlanningConflictNextRoundAndResync(t *testing.T) {
	state := clientMetadataState()
	state.Phase.Mode = turns.ModeHybrid
	state.Players = append(state.Players, engine.Player{ID: "c", NationID: "c", Lifecycle: engine.PlayerActive, Controller: engine.ControllerAI})
	state.Phase.ActivePlayerOrder = []string{"a", "b", "c"}
	state.Policy.Wars = []engine.TurnPolicyWar{{A: "a", B: "b"}}
	sources := engine.Sources{Clock: topologyClock{}, IDs: &topologyIDs{}, RNG: topologyRNG{}}
	apply := func(actor, kind string) {
		t.Helper()
		next, _, _, err := engine.Apply(state, engine.Command{SchemaVersion: engine.CommandSchemaV1, ID: fmt.Sprintf("%s-%d", actor, state.Revision), ActorPlayerID: actor, Type: kind, Payload: []byte(`{}`)}, state.Rules, sources)
		if err != nil {
			t.Fatalf("%s %s: %v", actor, kind, err)
		}
		if next.Revision != state.Revision+1 {
			t.Fatal("revision is not monotonic")
		}
		state, err = Update(next)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := engine.CanonicalStateJSON(state)
		if err != nil {
			t.Fatal(err)
		}
		state, err = engine.DecodeCanonicalStateJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
	}
	apply("c", engine.CommandNoop)
	projection, err := Project(state, "c")
	if err != nil {
		t.Fatal(err)
	}
	if projection.PlanningStatus != PlanningStatusQueued {
		t.Fatal("AI planning intent was not retained")
	}
	apply("a", engine.CommandRoundComplete)
	apply("c", engine.CommandRoundComplete)
	apply("b", engine.CommandRoundComplete)
	if state.Policy.Stage != turns.StageConflict || state.Phase.Turn != 4 {
		t.Fatalf("hybrid did not enter conflict without advancing turn: %+v", state.Policy)
	}
	for _, actor := range []string{"b", "c"} {
		projection, err := Project(state, actor)
		if err != nil {
			t.Fatal(err)
		}
		if len(projection.LegalActions) != 0 {
			t.Fatalf("inactive/component-external %s can act", actor)
		}
	}
	projection, err = Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !hasAction(projection.LegalActions, engine.CommandPass) || hasAction(projection.LegalActions, engine.CommandRoundComplete) {
		t.Fatal("conflict actor received wrong closeout action")
	}
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	current, err := (Projector{PlayerID: "a"}).ProjectSnapshot(events.Viewer{}, raw)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(current, receipt) {
		t.Fatalf("current resync differs from receipt projection:\n%s\n%s", current, receipt)
	}
	apply("a", engine.CommandPass)
	apply("b", engine.CommandPass)
	if state.Policy.Stage != turns.StagePlanning || state.Phase.Turn != 5 {
		t.Fatalf("hybrid conflict failed to return to planning: %+v", state.Policy)
	}
}
