package fullgame

import (
	"encoding/json"
	"errors"

	"github.com/justinfarrelldev/open-ctp-server/internal/game"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func ResolvePublicRole(gameDTO game.Game, accountID int64) (string, error) {
	if accountID < 1 {
		return "", errors.New("public account role is invalid")
	}
	matches := 0
	for _, player := range gameDTO.Players {
		if player.AccountID == accountID {
			matches++
		}
	}
	if matches != 1 {
		return "", errors.New("public account role is absent or ambiguous")
	}
	return "account-" + itoa(accountID), nil
}

func ResolveVisibleArmy(projection []byte, unitID string) (string, error) {
	if unitID == "" {
		return "", errors.New("visible unit identity is required")
	}
	var decoded visibility.Projection
	if err := json.Unmarshal(projection, &decoded); err != nil {
		return "", errors.New("public projection is invalid")
	}
	matches := 0
	armyID := ""
	for _, unit := range decoded.Units {
		if unit.ID == unitID {
			matches++
			armyID = unit.ArmyID
		}
	}
	if matches != 1 || armyID == "" {
		return "", errors.New("visible unit army is absent or ambiguous")
	}
	return armyID, nil
}

func ResolveOwnedCapabilityAndVisibleForeignCity(projection []byte, ownerID, capability string) (string, string, error) {
	if ownerID == "" || capability == "" {
		return "", "", errors.New("public role is invalid")
	}
	var decoded visibility.Projection
	if err := json.Unmarshal(projection, &decoded); err != nil {
		return "", "", errors.New("public projection is invalid")
	}
	unitID := ""
	for _, unit := range decoded.Units {
		if unit.OwnerID != ownerID || !hasPublicEffect(unit.Effects, capability) {
			continue
		}
		if unitID == "" || unit.ID < unitID {
			unitID = unit.ID
		}
	}
	if unitID == "" {
		return "", "", errors.New("owned capability actor is absent")
	}
	cityID := ""
	for _, city := range decoded.Cities {
		if city.OwnerID == ownerID || city.Stale {
			continue
		}
		if cityID == "" || city.ID < cityID {
			cityID = city.ID
		}
	}
	if cityID == "" {
		return "", "", errors.New("visible foreign city target is absent")
	}
	return unitID, cityID, nil
}

func hasPublicEffect(effects []engine.Effect, capability string) bool {
	for _, effect := range effects {
		if effect.Kind == capability {
			return true
		}
	}
	return false
}

func itoa(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var digits [20]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		index--
		digits[index] = '-'
	}
	return string(digits[index:])
}
