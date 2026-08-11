package setup

import (
	"fmt"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// Slot is one ordered public-game setup choice. Slot numbers are one-based.
type Slot struct {
	Slot            int
	CountryRecordID string
	ControllerKind  string
}

// Validate checks an ordered slot setup against imported selectable civilizations.
func Validate(ruleset rules.Ruleset, slots []Slot) error {
	if len(slots) < 2 || len(slots) > 8 {
		return fmt.Errorf("setup requires 2 to 8 slots, got %d", len(slots))
	}
	selectable, err := ruleset.PlayableCivilizationIDs()
	if err != nil {
		return fmt.Errorf("resolve selectable civilizations: %w", err)
	}
	allowed := make(map[string]struct{}, len(selectable))
	for _, id := range selectable {
		allowed[id] = struct{}{}
	}
	countries := make(map[string]struct{}, len(slots))
	for index, slot := range slots {
		if slot.Slot != index+1 {
			return fmt.Errorf("setup slot %d is out of order", slot.Slot)
		}
		if slot.ControllerKind != engine.ControllerHuman && slot.ControllerKind != engine.ControllerAI {
			return fmt.Errorf("setup slot %d has invalid controller %q", slot.Slot, slot.ControllerKind)
		}
		if _, ok := allowed[slot.CountryRecordID]; !ok {
			return fmt.Errorf("setup slot %d has unselectable civilization %q", slot.Slot, slot.CountryRecordID)
		}
		if _, duplicate := countries[slot.CountryRecordID]; duplicate {
			return fmt.Errorf("setup slot %d duplicates civilization %q", slot.Slot, slot.CountryRecordID)
		}
		countries[slot.CountryRecordID] = struct{}{}
	}
	return nil
}
