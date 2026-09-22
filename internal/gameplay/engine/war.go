package engine

import (
	"encoding/json"
	"errors"
)

type warStatePayload struct {
	TargetPlayerID string `json:"target_player_id"`
	AtWar          bool   `json:"at_war"`
}

func applyWarStateCommand(state *State, command Command) (string, json.RawMessage, error) {
	var payload warStatePayload
	if err := json.Unmarshal(command.Payload, &payload); err != nil || payload.TargetPlayerID == "" || payload.TargetPlayerID == command.ActorPlayerID {
		return "", nil, errors.New("war state payload is invalid")
	}
	if _, exists := playerByIDValue(*state, command.ActorPlayerID); !exists {
		return "", nil, errors.New("war actor is invalid")
	}
	if _, exists := playerByIDValue(*state, payload.TargetPlayerID); !exists {
		return "", nil, errors.New("war target is invalid")
	}
	left, right := command.ActorPlayerID, payload.TargetPlayerID
	if right < left {
		left, right = right, left
	}
	filtered := make([]TurnPolicyWar, 0, len(state.Policy.Wars)+1)
	for _, edge := range state.Policy.Wars {
		if edge.A != left || edge.B != right {
			filtered = append(filtered, edge)
		}
	}
	relation := ensureRelation(state, command.ActorPlayerID, payload.TargetPlayerID)
	relation.Contact = true
	if relation.EstablishedAt == 0 {
		relation.EstablishedAt = state.Phase.Turn
	}
	if payload.AtWar {
		filtered = append(filtered, TurnPolicyWar{A: left, B: right})
		relation.Treaty = ""
		relation.Trade = false
	} else if relation.Treaty == "" {
		relation.Treaty = "peace"
	}
	state.Policy.Wars = filtered
	encoded, err := json.Marshal(warStatePayload{TargetPlayerID: payload.TargetPlayerID, AtWar: payload.AtWar})
	if err != nil {
		return "", nil, err
	}
	return EventWarStateChanged, encoded, nil
}
