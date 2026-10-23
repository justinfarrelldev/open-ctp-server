package visibility

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
	"testing"
)

func TestTerminalTurnContextDropsActingIdentity(t *testing.T) {
	state := clientMetadataState()
	state.Terminal = engine.TerminalState{Ended: true, Reason: "bloodlust_conquest", WinningPlayerID: "a"}
	p, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if p.TurnContext.Status != "ended" || p.TurnContext.ActivePlayerID != "" || len(p.LegalActions) != 0 {
		t.Fatalf("terminal actor/action leaked: %+v", p.TurnContext)
	}
}

func TestTurnContextIsPublicRevisionBoundAndDoesNotInventResolutionProgress(t *testing.T) {
	for _, tc := range []struct {
		mode, stage, status, actor string
		locked                     bool
	}{
		{turns.ModeSequential, "", "acting", "b", false},
		{turns.ModeLongTurn, "", "acting", "b", false},
		{turns.ModeSimultaneous, turns.StagePlanning, "planning", "", false},
		{turns.ModeHybrid, turns.StagePlanning, "planning", "", false},
		{turns.ModeHybrid, turns.StageConflict, "conflict", "", false},
		{turns.ModeSequential, "", "resolving", "", true},
	} {
		t.Run(tc.mode+tc.status, func(t *testing.T) {
			state := clientMetadataState()
			state.Phase.Mode = tc.mode
			state.Policy.Stage = tc.stage
			state.Phase.ActivePlayerIndex = 1
			state.Phase.ResolutionLocked = tc.locked
			p, err := Project(state, "a")
			if err != nil {
				t.Fatal(err)
			}
			context, ok := metadataWire(t, p)["turn_context"].(map[string]any)
			if !ok {
				t.Fatal("no public turn context")
			}
			actor, _ := context["active_player_id"].(string)
			if context["revision"] != float64(state.Revision) || context["turn"] != float64(state.Phase.Turn) || context["status"] != tc.status || actor != tc.actor || context["live_progress"] != false {
				t.Fatalf("invented/unbound progress: %v", context)
			}
			if context["queue"] != nil || context["components"] != nil {
				t.Fatal("private planning leaked")
			}
			if p.Contract.Features["unit_route_preview"] || p.Contract.Features["live_resolution_progress"] {
				t.Fatal("unsupported preview/progress advertised")
			}
		})
	}
}
