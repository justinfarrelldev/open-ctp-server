package game

import (
	"errors"
	"fmt"

	world "github.com/justinfarrelldev/open-ctp-server/internal/world"
)

var ErrWorldStartContractsUnimplemented = errors.New("world start contracts are not implemented yet")

func NormalizeWorldGenerationInput(game *Game, players []GamePlayer) (world.GenerationInput, error) {
	input := world.GenerationInput{}
	if game != nil {
		input.GameID = game.ID
		input.HostParams = world.GenerationHostParams{
			Style:            game.Style,
			RulesJSON:        game.RulesJSON,
			ExclusionsJSON:   game.ExclusionsJSON,
			StartingSettlers: game.StartingSettlers,
			StartingGold:     game.StartingGold,
		}
	}
	for _, player := range players {
		input.PlayerOrder = append(input.PlayerOrder, player.AccountID)
	}
	return input, ErrWorldStartContractsUnimplemented
}

func ValidateWorldStartContract(game *Game, hasExistingWorld bool) error {
	if game == nil {
		return ErrWorldStartContractsUnimplemented
	}
	if hasExistingWorld {
		return fmt.Errorf("world already exists for game %d", game.ID)
	}
	return ErrWorldStartContractsUnimplemented
}
