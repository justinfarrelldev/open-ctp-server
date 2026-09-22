package fullgame

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// SetupSlot is one deterministic public-game setup choice. Slots are one-based
// so their order remains explicit at every public setup boundary.
type SetupSlot struct {
	Slot            int
	CountryRecordID string
	ControllerKind  string
}

func catalogSlotPlan(count, caseNumber int) []SetupSlot {
	controllers := make([]string, count)
	for index := range controllers {
		controllers[index] = engine.ControllerAI
		if caseNumber > 2 && caseNumber&(1<<index) == 0 {
			controllers[index] = engine.ControllerHuman
		}
	}
	controllers[0] = engine.ControllerHuman
	return slotPlan(controllers)
}

func slotPlan(controllers []string) []SetupSlot {
	slots := make([]SetupSlot, len(controllers))
	for index, controller := range controllers {
		slots[index] = SetupSlot{Slot: index + 1, CountryRecordID: campaignCivilizationID(index + 1), ControllerKind: controller}
	}
	return slots
}

func campaignCivilizationID(slot int) string {
	return fmt.Sprintf("CIVILIZATION_SYNTHETIC_%02d", slot)
}

func campaignPlayableCivilizations() ([]string, error) {
	directory, err := os.MkdirTemp("", "open-ctp-civilizations-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(directory)

	var records strings.Builder
	for slot := 1; slot <= 8; slot++ {
		fmt.Fprintf(&records, "%s { Playable true Special false Barbarian false }\n", campaignCivilizationID(slot))
	}
	rulesPath := filepath.Join(directory, "Civilisation.ctp2synthetic")
	if err := os.WriteFile(rulesPath, []byte(records.String()), 0o600); err != nil {
		return nil, err
	}
	ruleset, err := rules.Import(rules.ImportConfig{Directory: directory})
	if err != nil {
		return nil, err
	}
	return ruleset.PlayableCivilizationIDs()
}

func validateSetupSlots(slots []SetupSlot, selectable []string) error {
	if len(slots) < 2 || len(slots) > 8 {
		return fmt.Errorf("setup requires 2 to 8 slots, got %d", len(slots))
	}
	allowed := make(map[string]struct{}, len(selectable))
	for _, id := range selectable {
		allowed[id] = struct{}{}
	}
	countries := make(map[string]struct{}, len(slots))
	for index, setup := range slots {
		if setup.Slot != index+1 {
			return fmt.Errorf("setup slot %d is out of order", setup.Slot)
		}
		if setup.ControllerKind != engine.ControllerHuman && setup.ControllerKind != engine.ControllerAI {
			return fmt.Errorf("setup slot %d has invalid controller %q", setup.Slot, setup.ControllerKind)
		}
		if _, ok := allowed[setup.CountryRecordID]; !ok {
			return fmt.Errorf("setup slot %d has unselectable civilization %q", setup.Slot, setup.CountryRecordID)
		}
		if _, duplicate := countries[setup.CountryRecordID]; duplicate {
			return fmt.Errorf("setup slot %d duplicates civilization %q", setup.Slot, setup.CountryRecordID)
		}
		countries[setup.CountryRecordID] = struct{}{}
	}
	return nil
}
