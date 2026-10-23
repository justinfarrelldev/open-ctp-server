package service

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"
)

// CommandTiming is private, bounded command telemetry. No payload, session,
// account, unit/city IDs, command text or planning contents are included. Durations
// never enter canonical state or receipts. Configure ObserveCommand before use;
// observers must be quick, concurrency-safe and must not mutate gameplay state.
type CommandTiming struct {
	GameID        int64
	CommandSHA256 string
	Outcome       string // accepted, replayed, or an allowlisted rejection code
	Total         time.Duration
	Admission     time.Duration // transaction start, identity/receipt lookup, game lock/load
	Engine        time.Duration // including synchronous planning/barrier resolution
	Visibility    time.Duration // pre/post fog and viewer projection
	Persistence   time.Duration // serialization, writes and commit
}

type commandMeter struct {
	now    func() time.Time
	last   time.Time
	stage  *time.Duration
	metric CommandTiming
}

func newCommandMeter(now func() time.Time, request CommandRequest) *commandMeter {
	if now == nil {
		now = time.Now
	}
	hash := sha256.Sum256([]byte(request.CommandID))
	meter := &commandMeter{now: now, metric: CommandTiming{GameID: request.GameID, CommandSHA256: hex.EncodeToString(hash[:])}}
	meter.last = now()
	meter.stage = &meter.metric.Admission
	return meter
}

func (m *commandMeter) step(next *time.Duration) {
	now := m.now()
	elapsed := now.Sub(m.last)
	if elapsed < 0 {
		elapsed = 0
	} // defensive for injected clocks
	*m.stage += elapsed
	m.last = now
	m.stage = next
}

func (m *commandMeter) finish(outcome string) CommandTiming {
	m.step(m.stage)
	m.metric.Outcome = outcome
	m.metric.Total = m.metric.Admission + m.metric.Engine + m.metric.Visibility + m.metric.Persistence
	return m.metric
}

func logCommandTiming(metric CommandTiming) {
	slog.Debug("gameplay command timing", "game_id", metric.GameID, "command_sha256", metric.CommandSHA256, "outcome", metric.Outcome, "total_ms", metric.Total.Milliseconds(), "admission_ms", metric.Admission.Milliseconds(), "engine_ms", metric.Engine.Milliseconds(), "visibility_ms", metric.Visibility.Milliseconds(), "persistence_ms", metric.Persistence.Milliseconds())
}
