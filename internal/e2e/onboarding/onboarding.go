// Package onboarding provides fixed, deterministic public-HTTP onboarding scripts.
package onboarding

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/justinfarrelldev/open-ctp-server/internal/e2e/testkit"
)

var ErrConflict = errors.New("public endpoint conflict")

type Scenario struct {
	ID                    string
	Seed                  int64
	Players               int
	Style                 string
	DuplicateSignup       bool
	DuplicateJoin         bool
	CountryCollision      bool
	OwnerNonOwnerEdits    bool
	DisconnectBeforeStart bool
	ConcurrentStart       bool
	OrderlyTeardown       bool
}

// FixedScenarios is the complete Task 16 onboarding matrix. It intentionally
// contains no generated actors, modes, or country counts.
var FixedScenarios = []Scenario{
	{ID: "onboarding-01-sequential-2", Seed: 0x4354503200000001, Players: 2, Style: "sequential"},
	{ID: "onboarding-02-long-turn-3-duplicate-signup", Seed: 0x4354503200000002, Players: 3, Style: "long_turn", DuplicateSignup: true},
	{ID: "onboarding-03-simultaneous-4-duplicate-join", Seed: 0x4354503200000003, Players: 4, Style: "simultaneous", DuplicateJoin: true},
	{ID: "onboarding-04-hybrid-6-country-collision", Seed: 0x4354503200000004, Players: 6, Style: "hybrid", CountryCollision: true},
	{ID: "onboarding-05-sequential-8-owner-edits", Seed: 0x4354503200000005, Players: 8, Style: "sequential", OwnerNonOwnerEdits: true},
	{ID: "onboarding-06-long-turn-4-stream-rejoin", Seed: 0x4354503200000006, Players: 4, Style: "long_turn", DisconnectBeforeStart: true},
	{ID: "onboarding-07-simultaneous-6-start-race", Seed: 0x4354503200000007, Players: 6, Style: "simultaneous", ConcurrentStart: true},
	{ID: "onboarding-08-hybrid-8-full-teardown", Seed: 0x4354503200000008, Players: 8, Style: "hybrid", OwnerNonOwnerEdits: true, DisconnectBeforeStart: true, ConcurrentStart: true, OrderlyTeardown: true},
}

type Actor struct {
	Name    string
	Email   string
	Session int64
}

type GameSetup struct {
	Style            string
	RulesJSON        string
	ExclusionsJSON   string
	StartingSettlers int
	StartingGold     int
	MapSize          string
	Climate          string
	Wetness          *int
}

// Client is deliberately only the public boundary used by onboarding scripts.
// Implementations must not seed accounts, lobby membership, games, worlds, or
// gameplay aggregates directly.
type Client interface {
	Signup(context.Context, string, string) (Actor, error)
	CreateLobby(context.Context, Actor, string) (int64, error)
	JoinLobby(context.Context, Actor, int64) error
	OpenLobbyStream(context.Context, Actor, int64) (func() error, error)
	CreateGame(context.Context, Actor, int64, string) (int64, error)
	ConfigureGame(context.Context, Actor, int64, GameSetup) error
	SelectCountry(context.Context, Actor, int64, string) error
	StartGame(context.Context, Actor, int64) error
	ReadProjectedWorld(context.Context, Actor, int64) ([]byte, error)
	LeaveLobby(context.Context, Actor, int64) error
	AssertClean(context.Context, string) error
}

type Result struct {
	ScenarioID string
	Namespace  string
	LobbyID    int64
	GameID     int64
	Actors     []Actor
	World      []byte
}

func Run(ctx context.Context, scenario Scenario, client Client) (Result, error) {
	if err := validateScenario(scenario); err != nil {
		return Result{}, err
	}
	kit := testkit.NewScenario(scenario.ID, scenario.Seed)
	defer kit.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-kit.Context.Done():
			cancel()
		case <-ctx.Done():
		}
	}()

	actors, err := signupTogether(ctx, kit, scenario.Players, client)
	if err != nil {
		return Result{}, err
	}
	if scenario.DuplicateSignup {
		if _, err := client.Signup(ctx, actors[0].Name, actors[0].Email); !errors.Is(err, ErrConflict) {
			return Result{}, fmt.Errorf("duplicate signup: got %w, want %w", err, ErrConflict)
		}
	}

	lobbyID, err := client.CreateLobby(ctx, actors[0], kit.IDs.Next())
	if err != nil {
		return Result{}, fmt.Errorf("create lobby: %w", err)
	}
	for _, actor := range actors[1:] {
		if err := client.JoinLobby(ctx, actor, lobbyID); err != nil {
			return Result{}, fmt.Errorf("join lobby for %s: %w", actor.Name, err)
		}
	}
	if scenario.DuplicateJoin {
		if err := client.JoinLobby(ctx, actors[1], lobbyID); !errors.Is(err, ErrConflict) {
			return Result{}, fmt.Errorf("duplicate join: got %w, want %w", err, ErrConflict)
		}
	}

	closeStream, err := client.OpenLobbyStream(ctx, actors[0], lobbyID)
	if err != nil {
		return Result{}, fmt.Errorf("open lobby stream: %w", err)
	}

	gameID, err := client.CreateGame(ctx, actors[0], lobbyID, kit.IDs.Next())
	if err != nil {
		return Result{}, fmt.Errorf("create game: %w", err)
	}
	setup := GameSetup{Style: scenario.Style, RulesJSON: `[]`, ExclusionsJSON: `{}`, StartingSettlers: 2, StartingGold: 250}
	if err := client.ConfigureGame(ctx, actors[0], gameID, setup); err != nil {
		return Result{}, fmt.Errorf("owner setup: %w", err)
	}
	if scenario.OwnerNonOwnerEdits {
		if err := client.ConfigureGame(ctx, actors[1], gameID, setup); !errors.Is(err, ErrConflict) {
			return Result{}, fmt.Errorf("nonowner setup: got %w, want %w", err, ErrConflict)
		}
	}

	countries := countriesFor(scenario.Players)
	for index, actor := range actors {
		if scenario.CountryCollision && index == 1 {
			if err := client.SelectCountry(ctx, actor, gameID, countries[0]); !errors.Is(err, ErrConflict) {
				return Result{}, fmt.Errorf("country collision: got %w, want %w", err, ErrConflict)
			}
		}
		if err := client.SelectCountry(ctx, actor, gameID, countries[index]); err != nil {
			return Result{}, fmt.Errorf("select country for %s: %w", actor.Name, err)
		}
	}
	if scenario.DisconnectBeforeStart {
		if err := closeStream(); err != nil {
			return Result{}, fmt.Errorf("close lobby stream: %w", err)
		}
		closeStream, err = client.OpenLobbyStream(ctx, actors[0], lobbyID)
		if err != nil {
			return Result{}, fmt.Errorf("reopen lobby stream: %w", err)
		}
	}

	if err := start(ctx, scenario.ConcurrentStart, client, actors, gameID); err != nil {
		return Result{}, err
	}
	world, err := client.ReadProjectedWorld(ctx, actors[0], gameID)
	if err != nil || len(world) == 0 {
		return Result{}, fmt.Errorf("read projected world: %w", err)
	}
	if err := closeStream(); err != nil {
		return Result{}, fmt.Errorf("close lobby stream after start: %w", err)
	}
	if scenario.OrderlyTeardown {
		for _, actor := range actors[1:] {
			if err := client.LeaveLobby(ctx, actor, lobbyID); err != nil && !errors.Is(err, ErrConflict) {
				return Result{}, fmt.Errorf("leave lobby for %s: %w", actor.Name, err)
			}
		}
	}
	if err := client.AssertClean(ctx, kit.Namespace); err != nil {
		return Result{}, fmt.Errorf("assert cleanup: %w", err)
	}
	return Result{ScenarioID: scenario.ID, Namespace: kit.Namespace, LobbyID: lobbyID, GameID: gameID, Actors: actors, World: world}, nil
}

func signupTogether(ctx context.Context, kit *testkit.Scenario, count int, client Client) ([]Actor, error) {
	barrier := testkit.NewBarrier(count)
	actors := make([]Actor, count)
	errs := make(chan error, count)
	var wg sync.WaitGroup
	for index := range actors {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := barrier.Wait(ctx); err != nil {
				errs <- err
				return
			}
			name := kit.IDs.Next()
			actor, err := client.Signup(ctx, name, name+"@example.invalid")
			if err != nil {
				errs <- fmt.Errorf("signup for %s: %w", name, err)
				return
			}
			actors[index] = actor
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		return nil, err
	}
	return actors, nil
}

func start(ctx context.Context, concurrent bool, client Client, actors []Actor, gameID int64) error {
	if !concurrent {
		return client.StartGame(ctx, actors[0], gameID)
	}
	barrier := testkit.NewBarrier(2)
	errs := make(chan error, 2)
	for _, actor := range []Actor{actors[0], actors[1]} {
		actor := actor
		go func() {
			if err := barrier.Wait(ctx); err != nil {
				errs <- err
				return
			}
			errs <- client.StartGame(ctx, actor, gameID)
		}()
	}
	first, second := <-errs, <-errs
	if first == nil && errors.Is(second, ErrConflict) || second == nil && errors.Is(first, ErrConflict) {
		return nil
	}
	return fmt.Errorf("start race results %v and %v", first, second)
}

func validateScenario(s Scenario) error {
	if s.ID == "" || s.Seed == 0 || s.Players < 2 || s.Players > 8 || s.Style == "" {
		return fmt.Errorf("invalid onboarding scenario %#v", s)
	}
	return nil
}

func countriesFor(count int) []string {
	all := []string{"Athenians", "Babylonians", "Celts", "Dutch", "Egyptians", "French", "Greeks", "Hittites"}
	countries := append([]string(nil), all[:count]...)
	sort.Strings(countries)
	return countries
}
