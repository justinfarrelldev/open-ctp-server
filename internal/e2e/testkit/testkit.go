// Package testkit provides deterministic, dependency-free primitives for gameplay tests.
package testkit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const ScenarioBudget = 2 * time.Minute

// Scenario owns all deterministic dependencies for one isolated test case.
type Scenario struct {
	ID        string
	Seed      int64
	Namespace string
	Context   context.Context
	cancel    context.CancelFunc
	Clock     *FakeClock
	IDs       *IDSource
	RNG       *RNG
}

func NewScenario(id string, seed int64) *Scenario {
	if strings.TrimSpace(id) == "" {
		id = "unnamed"
	}
	namespaceHash := sha256.Sum256([]byte(id + "\x00" + strconv.FormatInt(seed, 10)))
	namespace := "gameplay_" + hex.EncodeToString(namespaceHash[:8])
	context, cancel := context.WithTimeout(context.Background(), ScenarioBudget)
	return &Scenario{
		ID:        id,
		Seed:      seed,
		Namespace: namespace,
		Context:   context,
		cancel:    cancel,
		Clock:     NewFakeClock(time.Unix(0, 0).UTC()),
		IDs:       NewIDSource(namespace),
		RNG:       NewRNG(uint64(seed)),
	}
}

func (s *Scenario) Deadline() time.Time {
	deadline, _ := s.Context.Deadline()
	return deadline
}

func (s *Scenario) Close() { s.cancel() }

// FakeClock is safe for concurrent readers and advances only when a test requests it.
type FakeClock struct {
	mu  sync.RWMutex
	now time.Time
}

func NewFakeClock(start time.Time) *FakeClock { return &FakeClock{now: start.UTC()} }

func (c *FakeClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

func (c *FakeClock) Advance(by time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(by)
	return c.now
}

// IDSource produces stable, scenario-namespaced IDs without package-global state.
type IDSource struct {
	mu        sync.Mutex
	namespace string
	next      uint64
}

func NewIDSource(namespace string) *IDSource { return &IDSource{namespace: namespace} }

func (s *IDSource) Next() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	return fmt.Sprintf("%s_%06d", s.namespace, s.next)
}

// RNG is a deterministic SplitMix64 stream protected for concurrent test helpers.
type RNG struct {
	mu    sync.Mutex
	state uint64
}

func NewRNG(seed uint64) *RNG { return &RNG{state: seed} }

func (r *RNG) Uint64() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (r *RNG) Int63() int64 { return int64(r.Uint64() >> 1) }

// Barrier waits for exactly n participants without time-based ordering.
type Barrier struct {
	mu      sync.Mutex
	target  int
	arrived int
	done    chan struct{}
}

func NewBarrier(participants int) *Barrier {
	if participants < 1 {
		panic("barrier requires at least one participant")
	}
	return &Barrier{target: participants, done: make(chan struct{})}
}

func (b *Barrier) Wait(context context.Context) error {
	b.mu.Lock()
	if b.arrived < b.target {
		b.arrived++
		if b.arrived == b.target {
			close(b.done)
		}
	}
	done := b.done
	b.mu.Unlock()
	select {
	case <-done:
		return nil
	case <-context.Done():
		return context.Err()
	}
}

// FaultPlan permits deterministic command-index fault injection.
type FaultPlan struct {
	AtCommand int
	Err       error
}

func (p FaultPlan) Check(commandIndex int) error {
	if p.Err != nil && p.AtCommand == commandIndex {
		return p.Err
	}
	return nil
}

// FailureEvidence contains only identifiers, hashes, and safe artifact paths.
type FailureEvidence struct {
	ScenarioID     string `json:"scenario_id"`
	Seed           int64  `json:"seed"`
	CommandIndex   int    `json:"command_index"`
	TranscriptPath string `json:"transcript_path"`
	EventPath      string `json:"event_path"`
	ProjectionPath string `json:"projection_path"`
	TranscriptHash string `json:"transcript_hash"`
	EventHash      string `json:"event_hash"`
	ProjectionHash string `json:"projection_hash"`
}

func (s *Scenario) CaptureFailure(root string, commandIndex int, transcript, events, projections any) (FailureEvidence, error) {
	if commandIndex < 0 {
		return FailureEvidence{}, errors.New("command index must not be negative")
	}
	directory := filepath.Join(root, s.Namespace)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return FailureEvidence{}, fmt.Errorf("create evidence directory: %w", err)
	}
	transcriptPath, transcriptHash, err := writeCanonicalArtifact(directory, "transcript.json", transcript)
	if err != nil {
		return FailureEvidence{}, err
	}
	eventPath, eventHash, err := writeCanonicalArtifact(directory, "events.json", events)
	if err != nil {
		return FailureEvidence{}, err
	}
	projectionPath, projectionHash, err := writeCanonicalArtifact(directory, "projections.json", projections)
	if err != nil {
		return FailureEvidence{}, err
	}
	return FailureEvidence{
		ScenarioID: s.ID, Seed: s.Seed, CommandIndex: commandIndex,
		TranscriptPath: transcriptPath, EventPath: eventPath, ProjectionPath: projectionPath,
		TranscriptHash: transcriptHash, EventHash: eventHash, ProjectionHash: projectionHash,
	}, nil
}

func writeCanonicalArtifact(directory, name string, value any) (string, string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", "", fmt.Errorf("encode %s: %w", name, err)
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		return "", "", fmt.Errorf("write %s: %w", name, err)
	}
	hash := sha256.Sum256(encoded)
	return path, hex.EncodeToString(hash[:]), nil
}
