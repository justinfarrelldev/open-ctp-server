package onboarding

import (
	"context"
	"net/http"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/ai"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	serverruntime "github.com/justinfarrelldev/open-ctp-server/internal/runtime"
)

type PublicRuntime struct {
	Handler             http.Handler
	Driver              *ai.Driver
	RunDueDeadlines     func(context.Context) error
	RunPendingTakeovers func(context.Context) error
	RunAutonomousAI     func(context.Context) error
}

func NewPublicRuntime(db *sqlx.DB, sources engine.Sources, seed func() (int64, error)) *PublicRuntime {
	return newPublicRuntime(db, sources, seed, nil)
}

// NewPublicRuntimeWithAIObserver uses the ordinary production composition and records only AI
// submissions that the production driver has successfully sent through its gateway.
func NewPublicRuntimeWithAIObserver(db *sqlx.DB, sources engine.Sources, seed func() (int64, error), observe func(ai.Submission)) *PublicRuntime {
	return newPublicRuntime(db, sources, seed, observe)
}

func RestartPublicRuntime(db *sqlx.DB, sources engine.Sources, seed func() (int64, error)) *PublicRuntime {
	return NewPublicRuntime(db, sources, seed)
}

func RestartPublicRuntimeWithAIObserver(db *sqlx.DB, sources engine.Sources, seed func() (int64, error), observe func(ai.Submission)) *PublicRuntime {
	return NewPublicRuntimeWithAIObserver(db, sources, seed, observe)
}

func newPublicRuntime(db *sqlx.DB, sources engine.Sources, seed func() (int64, error), observe func(ai.Submission)) *PublicRuntime {
	runtime := serverruntime.New(db, serverruntime.Config{Sources: sources, StartSeed: seed, NewLimiter: func(string) *limiter.Limiter { return tollbooth.NewLimiter(1000000, nil) }, ObserveAICommand: observe})
	return &PublicRuntime{Handler: runtime.Handler, Driver: runtime.AIDriver, RunDueDeadlines: runtime.Deadlines.RunOnce, RunPendingTakeovers: runtime.PendingTakeovers.RunOnce, RunAutonomousAI: runtime.AIWorker.RunOnce}
}
