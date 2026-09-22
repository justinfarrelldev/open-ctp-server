package fullgame

import (
	"context"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
)

type Artifacts struct {
	TerminalWinner string
	TerminalReason string
	StateHash      string
	EventHash      string
	AICommands     int
	AITranscript   string
}

// Executor is implemented by the public HTTP/SSE gateway runner. It must create a new namespace
// from the supplied testkit scenario, start through public setup APIs, execute fixed human scripts
// and actual AI commands, restart at RestartCheckpoint, then return audit-free authoritative hashes.
type Executor interface {
	Execute(context.Context, Scenario, *testkit.Scenario) (Artifacts, error)
}

type Result struct {
	ScenarioID string
	First      Artifacts
	Second     Artifacts
}
