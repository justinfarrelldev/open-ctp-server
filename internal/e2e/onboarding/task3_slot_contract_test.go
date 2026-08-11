package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	e2esetup "github.com/justinfarrelldev/open-ctp-server/internal/e2e/setup"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func orderedSlots(controllers []string) []e2esetup.Slot {
	slots := make([]e2esetup.Slot, len(controllers))
	for index, controller := range controllers {
		slots[index] = e2esetup.Slot{Slot: index + 1, CountryRecordID: syntheticCivilizationID(index + 1), ControllerKind: controller}
	}
	return slots
}

func importedPlayableRuleset(t *testing.T, count int) rules.Ruleset {
	t.Helper()
	directory := t.TempDir()
	var records strings.Builder
	for slot := 1; slot <= count; slot++ {
		fmt.Fprintf(&records, "%s { Playable true Special false Barbarian false }\n", syntheticCivilizationID(slot))
	}
	if err := os.WriteFile(filepath.Join(directory, "Civilisation.ctp2synthetic"), []byte(records.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	ruleset, err := rules.Import(rules.ImportConfig{Directory: directory})
	if err != nil {
		t.Fatal(err)
	}
	return ruleset
}

func syntheticCivilizationID(slot int) string {
	return fmt.Sprintf("CIVILIZATION_SYNTHETIC_%02d", slot)
}
