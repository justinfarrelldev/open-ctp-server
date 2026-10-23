package game

import (
	"errors"
	"fmt"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/world"
)

var errStartingAdvanceSetup = errors.New("starting advance difficulty/age settings are invalid")
var errCivilizationSetup = errors.New("selected country is not available in the pinned retail civilization catalog")

func configureRetailStartingAdvances(runtime *rules.RuntimeRules, options []string) error {
	if runtime.StartingAdvances == nil {
		return fmt.Errorf("retail new games require an explicit starting advance policy")
	}
	seen := map[string]bool{}
	for _, option := range options {
		key, value, ok := strings.Cut(option, ":")
		if !ok || key != "difficulty" && key != "start-age" {
			continue
		}
		if seen[key] || value == "" {
			return errStartingAdvanceSetup
		}
		seen[key] = true
		if key == "difficulty" {
			runtime.StartingAdvances.DifficultyID = value
		} else {
			runtime.StartAgeID = value
		}
	}
	if err := rules.ValidateStartingAdvances(*runtime); err != nil {
		return fmt.Errorf("%w: %v", errStartingAdvanceSetup, err)
	}
	return nil
}

// Startup uses a dedicated per-player stream derived from the server-generated
// world's private RNG seed/state. It cannot perturb terrain/spawns, gameplay RNG
// domains or another player's starting rolls; no seed or roll is projected.
func startingAdvanceRolls(seed uint64) engine.StartingAdvanceRoll {
	streams := map[string]world.DeterministicRNG{}
	return func(playerID string) int {
		stream, ok := streams[playerID]
		if !ok {
			stream = world.NewDeterministicRNG(int64(seed)).Split("starting-advances-v1:" + playerID)
		}
		next, value := stream.Intn(100)
		streams[playerID] = next
		return value
	}
}
