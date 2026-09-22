package fullgame

import (
	"encoding/json"
	"errors"
)

type TerminalFrame struct {
	Sequence int64
	Winner   string
	Reason   string
}

func ParseTerminalSSE(sequence int64, name string, data []byte) (TerminalFrame, error) {
	if sequence < 1 || name != "terminal" {
		return TerminalFrame{}, errors.New("terminal SSE frame is invalid")
	}
	var envelope struct {
		Terminal struct {
			WinningPlayerID string `json:"winning_player_id"`
			Reason          string `json:"reason"`
		} `json:"terminal"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.Terminal.WinningPlayerID == "" || envelope.Terminal.Reason == "" {
		return TerminalFrame{}, errors.New("terminal SSE envelope is invalid")
	}
	return TerminalFrame{Sequence: sequence, Winner: envelope.Terminal.WinningPlayerID, Reason: envelope.Terminal.Reason}, nil
}
