package game

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

const (
	defaultWorldStartGeneratorVersion = world.RuinGeneratorVersion
	defaultWorldStartMapSize          = "small"
	defaultWorldStartClimate          = "temperate"
	defaultWorldStartWetness          = 50
)

var generateStartWorldSeed = func() (int64, error) {
	var seedBytes [8]byte
	if _, err := cryptorand.Read(seedBytes[:]); err != nil {
		return 0, fmt.Errorf("generate start world seed: %w", err)
	}
	seed := int64(binary.BigEndian.Uint64(seedBytes[:]) >> 1)
	if seed == 0 {
		seed = 1
	}
	return seed, nil
}

func NormalizeWorldGenerationInput(game *Game, players []GamePlayer, seed int64) (world.GenerationInput, error) {
	if game == nil {
		return world.GenerationInput{}, errors.New("game must be provided")
	}
	if seed == 0 {
		return world.GenerationInput{}, errors.New("seed must be non-zero")
	}

	wetness := defaultWorldStartWetness
	if game.Wetness != nil {
		wetness = *game.Wetness
	}
	mapSize, climate := defaultWorldStartMapSize, defaultWorldStartClimate
	if game.MapSize != nil && strings.TrimSpace(*game.MapSize) != "" {
		mapSize = strings.TrimSpace(*game.MapSize)
	}
	if game.Climate != nil && strings.TrimSpace(*game.Climate) != "" {
		climate = strings.TrimSpace(*game.Climate)
	}
	config, err := world.NormalizeWorldConfig(world.WorldConfig{
		Style:            game.Style,
		RulesJSON:        game.RulesJSON,
		ExclusionsJSON:   game.ExclusionsJSON,
		StartingSettlers: game.StartingSettlers,
		StartingGold:     game.StartingGold,
		Seed:             &seed,
		Wetness:          &wetness,
		MapSize:          mapSize,
		Climate:          climate,
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
	style := strings.TrimSpace(game.Style)
	if style != "" && !supportedGameStyle(style) {
		return fmt.Errorf("unsupported game style %q", game.Style)
	}
	return nil
}
