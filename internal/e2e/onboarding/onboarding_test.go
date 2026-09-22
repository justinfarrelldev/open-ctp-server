package onboarding

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type recordingClient struct {
	mu          sync.Mutex
	signups     map[string]Actor
	owner       int64
	countries   map[string]int64
	joins       map[int64]bool
	started     bool
	streams     int
	leaves      int
	startCalls  int
	nextLobbyID int64
	nextGameID  int64
}

func newRecordingClient() *recordingClient {
	return &recordingClient{signups: make(map[string]Actor), countries: make(map[string]int64), joins: make(map[int64]bool), nextLobbyID: 100, nextGameID: 200}
}

func (c *recordingClient) Signup(_ context.Context, name, email string) (Actor, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.signups[email]; exists {
		return Actor{}, ErrConflict
	}
	actor := Actor{Name: name, Email: email, Session: int64(len(c.signups) + 1)}
	c.signups[email] = actor
	return actor, nil
}
func (c *recordingClient) CreateLobby(_ context.Context, actor Actor, _ string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.owner = actor.Session
	c.nextLobbyID++
	return c.nextLobbyID, nil
}
func (c *recordingClient) JoinLobby(_ context.Context, actor Actor, _ int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if actor.Session == 0 || actor.Session > int64(len(c.signups)) {
		return ErrConflict
	}
	if c.joins[actor.Session] {
		return ErrConflict
	}
	c.joins[actor.Session] = true
	return nil
}
func (c *recordingClient) OpenLobbyStream(_ context.Context, _ Actor, _ int64) (func() error, error) {
	c.mu.Lock()
	c.streams++
	closed := false
	c.mu.Unlock()
	return func() error {
		c.mu.Lock()
		defer c.mu.Unlock()
		if !closed {
			c.streams--
			closed = true
		}
		return nil
	}, nil
}
func (c *recordingClient) CreateGame(_ context.Context, actor Actor, _ int64, _ string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if actor.Session != c.owner {
		return 0, ErrConflict
	}
	c.nextGameID++
	return c.nextGameID, nil
}
func (c *recordingClient) ConfigureGame(_ context.Context, actor Actor, _ int64, _ GameSetup) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if actor.Session != c.owner {
		return ErrConflict
	}
	return nil
}
func (c *recordingClient) SelectCountry(_ context.Context, actor Actor, _ int64, country string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if owner, exists := c.countries[country]; exists && owner != actor.Session {
		return ErrConflict
	}
	c.countries[country] = actor.Session
	return nil
}
func (c *recordingClient) StartGame(_ context.Context, actor Actor, _ int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startCalls++
	if actor.Session != c.owner || c.started {
		return ErrConflict
	}
	c.started = true
	return nil
}
func (c *recordingClient) ReadProjectedWorld(_ context.Context, actor Actor, _ int64) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.started || actor.Session == 0 {
		return nil, fmt.Errorf("world unavailable")
	}
	return []byte(`{"viewer":"projected"}`), nil
}
func (c *recordingClient) LeaveLobby(_ context.Context, _ Actor, _ int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.leaves++
	return nil
}
func (c *recordingClient) AssertClean(_ context.Context, namespace string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if namespace == "" || c.streams != 0 {
		return fmt.Errorf("unclean namespace=%q streams=%d", namespace, c.streams)
	}
	return nil
}

func TestFixedScenariosAreCompleteAndStable(t *testing.T) {
	if got, want := len(FixedScenarios), 8; got != want {
		t.Fatalf("scenario count=%d want=%d", got, want)
	}
	seen := map[string]bool{}
	counts := map[int]bool{}
	styles := map[string]bool{}
	for _, scenario := range FixedScenarios {
		if seen[scenario.ID] {
			t.Fatalf("duplicate scenario %s", scenario.ID)
		}
		seen[scenario.ID] = true
		counts[scenario.Players] = true
		styles[scenario.Style] = true
	}
	for _, count := range []int{2, 3, 4, 6, 8} {
		if !counts[count] {
			t.Fatalf("missing country count %d", count)
		}
	}
	for _, style := range []string{"sequential", "long_turn", "simultaneous", "hybrid"} {
		if !styles[style] {
			t.Fatalf("missing style %s", style)
		}
	}
}

func TestRunExecutesEightFixedOnboardingScriptsWithoutLeaks(t *testing.T) {
	for _, scenario := range FixedScenarios {
		scenario := scenario
		t.Run(scenario.ID, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testScenarioTimeout(t))
			defer cancel()
			client := newRecordingClient()
			result, err := Run(ctx, scenario, client)
			if err != nil {
				t.Fatal(err)
			}
			if result.Namespace == "" || len(result.Actors) != scenario.Players || len(result.World) == 0 {
				t.Fatalf("incomplete result: %#v", result)
			}
			if scenario.OrderlyTeardown && client.leaves != scenario.Players-1 {
				t.Fatalf("leaves=%d want=%d", client.leaves, scenario.Players-1)
			}
			if scenario.ConcurrentStart && client.startCalls != 2 {
				t.Fatalf("start calls=%d want=2", client.startCalls)
			}
		})
	}
}

func testScenarioTimeout(t *testing.T) time.Duration { t.Helper(); return 2 * time.Minute }
