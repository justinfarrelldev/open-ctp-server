package game

import (
	"errors"
	"fmt"
	"strings"

	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

const (
	defaultWorldStartGeneratorVersion = "ctp2-first-v1"
	defaultWorldStartMapSize          = "small"
	defaultWorldStartClimate          = "temperate"
	defaultWorldStartWetness          = 50
)

func NormalizeWorldGenerationInput(game *Game, players []GamePlayer) (world.GenerationInput, error) {
	if game == nil {
		return world.GenerationInput{}, errors.New("game must be provided")
	}

	seed := game.ID
	if seed <= 0 {
		seed = 1
	}
	wetness := defaultWorldStartWetness
	config, err := world.NormalizeWorldConfig(world.WorldConfig{
		Style:            game.Style,
		RulesJSON:        game.RulesJSON,
		ExclusionsJSON:   game.ExclusionsJSON,
		StartingSettlers: game.StartingSettlers,
		StartingGold:     game.StartingGold,
		Seed:             &seed,
		Wetness:          &wetness,
		MapSize:          defaultWorldStartMapSize,
		Climate:          defaultWorldStartClimate,
		GeneratorVersion: defaultWorldStartGeneratorVersion,
	})
	if err != nil {
		return world.GenerationInput{}, err
	}

	input := world.GenerationInput{
		GameID:           game.ID,
		GeneratorVersion: config.GeneratorVersion,
		HostParams:       config.ToHostParams(),
	}
	for _, player := range players {
		input.PlayerOrder = append(input.PlayerOrder, player.AccountID)
	}
	return input, nil
}

func ApplyWorldSpawnCountries(snapshot world.WorldSnapshot, players []GamePlayer) world.WorldSnapshot {
	clone := snapshot.Clone()
	countriesByAccountID := make(map[int64]string, len(players))
	for _, player := range players {
		countriesByAccountID[player.AccountID] = strings.TrimSpace(player.Country)
	}
	for i := range clone.Spawns {
		if country, ok := countriesByAccountID[clone.Spawns[i].AccountID]; ok {
			clone.Spawns[i].Country = country
		}
	}
	clone.Turn = snapshot.Turn.Clone()
	return clone
}

func ValidateWorldStartContract(game *Game, hasExistingWorld bool) error {
	if game == nil {
		return errors.New("game must be provided")
	}
	if game.Status != DefaultGameStatus {
		return fmt.Errorf("game %d is not in setup", game.ID)
	}
	if hasExistingWorld {
		return fmt.Errorf("world already exists for game %d", game.ID)
	}
	return nil
}
