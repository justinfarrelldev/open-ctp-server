package visibility

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
)

// TurnContext is committed public scheduler context, not an in-flight AI progress
// feed. No policy queues, closed players, conflict components or future orders are
// exposed. Hybrid acting participants are deliberately unavailable to spectators.
type TurnContext struct {
	Version        string `json:"version"`
	Revision       int64  `json:"revision"`
	Turn           int    `json:"turn"`
	Mode           string `json:"mode"`
	Status         string `json:"status"`
	ActivePlayerID string `json:"active_player_id,omitempty"`
	LiveProgress   bool   `json:"live_progress"`
}

func turnContext(state engine.State) *TurnContext {
	result := &TurnContext{Version: "committed-turn-context-v1", Revision: state.Revision, Turn: state.Phase.Turn, Mode: state.Phase.Mode, Status: "unavailable"}
	switch {
	case state.Terminal.Ended:
		result.Status = "ended"
	case state.Phase.ResolutionLocked:
		result.Status = "resolving"
	case state.Phase.Mode == turns.ModeSimultaneous || state.Phase.Mode == turns.ModeHybrid:
		if state.Policy.Stage == turns.StageConflict {
			result.Status = "conflict"
		} else {
			result.Status = "planning"
		}
	case state.Phase.Mode == turns.ModeSequential || state.Phase.Mode == turns.ModeLongTurn:
		if state.Phase.ActivePlayerIndex >= 0 && state.Phase.ActivePlayerIndex < len(state.Phase.ActivePlayerOrder) {
			result.Status = "acting"
			result.ActivePlayerID = state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex]
		}
	}
	return result
}
