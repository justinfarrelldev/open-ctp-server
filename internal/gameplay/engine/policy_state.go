package engine

import "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"

// AsTurnPolicyState projects the persisted canonical policy fields into the existing resolver contract.
func (state State) AsTurnPolicyState() turns.State {
	policy := turns.State{Mode: state.Phase.Mode, Version: state.Phase.ModeVersion, Turn: state.Phase.Turn, ActiveOrder: append([]string(nil), state.Phase.ActivePlayerOrder...), ActiveIndex: state.Phase.ActivePlayerIndex, ResolutionLocked: state.Phase.ResolutionLocked, Stage: state.Policy.Stage, ComponentIndex: state.Policy.ComponentIndex, Recalculations: state.Policy.Recalculations, EarlyEnds: append([]string(nil), state.Policy.EarlyEnds...)}
	if policy.Stage == "" {
		policy.Stage = turns.StagePlanning
	}
	for _, player := range state.Players {
		policy.Players = append(policy.Players, turns.Player{ID: player.ID, Eligible: player.Lifecycle == PlayerActive})
	}
	for _, command := range state.Policy.Queue {
		policy.Queue = append(policy.Queue, turns.Command{Kind: command.Kind, ActorID: command.ActorID, ActionClass: command.ActionClass, UnitOrArmyID: command.UnitOrArmyID, ReceiptSequence: command.ReceiptSequence, Hostile: command.Hostile, TargetPlayerID: command.TargetPlayerID, CommandID: command.CommandID, IdempotencyKey: command.IdempotencyKey, Payload: append([]byte(nil), command.Payload...)})
	}
	if state.Policy.Deadline != nil {
		policy.Deadline = &turns.Deadline{PlayerID: state.Policy.Deadline.PlayerID, AtUnix: state.Policy.Deadline.AtUnix}
	}
	for _, war := range state.Policy.Wars {
		policy.Wars = append(policy.Wars, turns.War{A: war.A, B: war.B})
	}
	policy.Components = make([][]string, len(state.Policy.Components))
	for index := range state.Policy.Components {
		policy.Components[index] = append([]string(nil), state.Policy.Components[index]...)
	}
	return policy
}

// WithTurnPolicyState applies resolver output to canonical authority for transactional persistence.
func (state State) WithTurnPolicyState(policy turns.State) State {
	state.Phase.Turn = policy.Turn
	state.Phase.ActivePlayerOrder = append([]string(nil), policy.ActiveOrder...)
	state.Phase.ActivePlayerIndex = policy.ActiveIndex
	state.Phase.ResolutionLocked = policy.ResolutionLocked
	state.Policy.Stage = policy.Stage
	state.Policy.ComponentIndex = policy.ComponentIndex
	state.Policy.Recalculations = policy.Recalculations
	state.Policy.EarlyEnds = append([]string(nil), policy.EarlyEnds...)
	state.Policy.Queue = make([]TurnPolicyCommand, 0, len(policy.Queue))
	for _, command := range policy.Queue {
		state.Policy.Queue = append(state.Policy.Queue, TurnPolicyCommand{Kind: command.Kind, ActorID: command.ActorID, ActionClass: command.ActionClass, UnitOrArmyID: command.UnitOrArmyID, ReceiptSequence: command.ReceiptSequence, Hostile: command.Hostile, TargetPlayerID: command.TargetPlayerID, CommandID: command.CommandID, IdempotencyKey: command.IdempotencyKey, Payload: append([]byte(nil), command.Payload...)})
	}
	state.Policy.Wars = make([]TurnPolicyWar, 0, len(policy.Wars))
	for _, war := range policy.Wars {
		state.Policy.Wars = append(state.Policy.Wars, TurnPolicyWar{A: war.A, B: war.B})
	}
	state.Policy.Components = make([][]string, len(policy.Components))
	for index := range policy.Components {
		state.Policy.Components[index] = append([]string(nil), policy.Components[index]...)
	}
	if policy.Deadline == nil {
		state.Policy.Deadline = nil
	} else {
		state.Policy.Deadline = &TurnPolicyDeadline{PlayerID: policy.Deadline.PlayerID, AtUnix: policy.Deadline.AtUnix}
	}
	return state
}
