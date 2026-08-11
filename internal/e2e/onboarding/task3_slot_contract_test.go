package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

type orderedSetupSlot struct {
	Slot            int
	CountryRecordID string
	ControllerKind  string
}

func TestOrderedSetupSlotsCoverEveryControllerVector(t *testing.T) {
	selectable := importedPlayableCivilizations(t)
	for count := 2; count <= 8; count++ {
		for vector := 0; vector < 1<<count; vector++ {
			controllers := make([]string, count)
			for index := range controllers {
				controllers[index] = engine.ControllerHuman
				if vector&(1<<index) != 0 {
					controllers[index] = engine.ControllerAI
				}
			}
			slots := orderedSlots(controllers)
			if err := validateOrderedSlots(slots, selectable); err != nil {
				t.Fatalf("count %d vector %0*b rejected: %v", count, count, vector, err)
			}
		}
	}
}

func TestOrderedSetupSlotsFailClosed(t *testing.T) {
	selectable := importedPlayableCivilizations(t)
	valid := orderedSlots([]string{engine.ControllerHuman, engine.ControllerAI})
	cases := map[string][]orderedSetupSlot{
		"one slot":           {{Slot: 1, CountryRecordID: valid[0].CountryRecordID, ControllerKind: engine.ControllerHuman}},
		"nine slots":         append(orderedSlots([]string{engine.ControllerHuman, engine.ControllerAI, engine.ControllerHuman, engine.ControllerAI, engine.ControllerHuman, engine.ControllerAI, engine.ControllerHuman, engine.ControllerAI}), orderedSetupSlot{Slot: 9, CountryRecordID: "CIVILIZATION_SYNTHETIC_09", ControllerKind: engine.ControllerHuman}),
		"duplicate slot":     {{Slot: 1, CountryRecordID: valid[0].CountryRecordID, ControllerKind: engine.ControllerHuman}, {Slot: 1, CountryRecordID: valid[1].CountryRecordID, ControllerKind: engine.ControllerAI}},
		"missing slot":       {{Slot: 1, CountryRecordID: valid[0].CountryRecordID, ControllerKind: engine.ControllerHuman}, {Slot: 3, CountryRecordID: valid[1].CountryRecordID, ControllerKind: engine.ControllerAI}},
		"duplicate country":  {{Slot: 1, CountryRecordID: valid[0].CountryRecordID, ControllerKind: engine.ControllerHuman}, {Slot: 2, CountryRecordID: valid[0].CountryRecordID, ControllerKind: engine.ControllerAI}},
		"unknown country":    {{Slot: 1, CountryRecordID: valid[0].CountryRecordID, ControllerKind: engine.ControllerHuman}, {Slot: 2, CountryRecordID: "CIVILIZATION_UNKNOWN", ControllerKind: engine.ControllerAI}},
		"unknown controller": {{Slot: 1, CountryRecordID: valid[0].CountryRecordID, ControllerKind: engine.ControllerHuman}, {Slot: 2, CountryRecordID: valid[1].CountryRecordID, ControllerKind: "robot"}},
	}
	for name, slots := range cases {
		t.Run(name, func(t *testing.T) {
			if err := validateOrderedSlots(slots, selectable); err == nil {
				t.Fatal("invalid ordered slots accepted")
			}
		})
	}
}

func orderedSlots(controllers []string) []orderedSetupSlot {
	slots := make([]orderedSetupSlot, len(controllers))
	for index, controller := range controllers {
		slots[index] = orderedSetupSlot{Slot: index + 1, CountryRecordID: syntheticCivilizationID(index + 1), ControllerKind: controller}
	}
	return slots
}

func importedPlayableCivilizations(t *testing.T) []string {
	t.Helper()
	directory := t.TempDir()
	var records strings.Builder
	for slot := 1; slot <= 8; slot++ {
		fmt.Fprintf(&records, "%s { Playable true Special false Barbarian false }\n", syntheticCivilizationID(slot))
	}
	if err := os.WriteFile(filepath.Join(directory, "Civilisation.ctp2synthetic"), []byte(records.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	ruleset, err := rules.Import(rules.ImportConfig{Directory: directory})
	if err != nil {
		t.Fatal(err)
	}
	ids, err := ruleset.PlayableCivilizationIDs()
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

func syntheticCivilizationID(slot int) string {
	return fmt.Sprintf("CIVILIZATION_SYNTHETIC_%02d", slot)
}

func validateOrderedSlots(slots []orderedSetupSlot, selectable []string) error {
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
