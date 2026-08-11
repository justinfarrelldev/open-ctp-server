package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestValidateAcceptsEveryOrderedControllerVector(t *testing.T) {
	ruleset := importedRuleset(t, 10)
	for count := 2; count <= 8; count++ {
		for vector := 0; vector < 1<<count; vector++ {
			slots := make([]Slot, count)
			for index := range slots {
				controller := engine.ControllerHuman
				if vector&(1<<index) != 0 {
					controller = engine.ControllerAI
				}
				slots[index] = Slot{Slot: index + 1, CountryRecordID: civilizationID(index + 1), ControllerKind: controller}
			}
			if err := Validate(ruleset, slots); err != nil {
				t.Fatalf("count %d vector %0*b rejected: %v", count, count, vector, err)
			}
		}
	}
}

func TestValidateAcceptsEveryImportedPlayableCivilization(t *testing.T) {
	ruleset := importedRuleset(t, 10)
	ids, err := ruleset.PlayableCivilizationIDs()
	if err != nil {
		t.Fatal(err)
	}
	for index, id := range ids {
		other := ids[(index+1)%len(ids)]
		slots := []Slot{
			{Slot: 1, CountryRecordID: id, ControllerKind: engine.ControllerHuman},
			{Slot: 2, CountryRecordID: other, ControllerKind: engine.ControllerAI},
		}
		if err := Validate(ruleset, slots); err != nil {
			t.Fatalf("playable civilization %q rejected: %v", id, err)
		}
	}
}

func TestValidateRejectsInvalidSlots(t *testing.T) {
	ruleset := importedRuleset(t, 9)
	valid := []Slot{
		{Slot: 1, CountryRecordID: civilizationID(1), ControllerKind: engine.ControllerHuman},
		{Slot: 2, CountryRecordID: civilizationID(2), ControllerKind: engine.ControllerAI},
	}
	nine := make([]Slot, 9)
	for index := range nine {
		nine[index] = Slot{Slot: index + 1, CountryRecordID: civilizationID(index + 1), ControllerKind: engine.ControllerHuman}
	}
	cases := map[string][]Slot{
		"one slot":           {valid[0]},
		"nine slots":         nine,
		"duplicate slot":     {{Slot: 1, CountryRecordID: civilizationID(1), ControllerKind: engine.ControllerHuman}, {Slot: 1, CountryRecordID: civilizationID(2), ControllerKind: engine.ControllerAI}},
		"missing slot":       {{Slot: 1, CountryRecordID: civilizationID(1), ControllerKind: engine.ControllerHuman}, {Slot: 3, CountryRecordID: civilizationID(2), ControllerKind: engine.ControllerAI}},
		"duplicate country":  {{Slot: 1, CountryRecordID: civilizationID(1), ControllerKind: engine.ControllerHuman}, {Slot: 2, CountryRecordID: civilizationID(1), ControllerKind: engine.ControllerAI}},
		"unknown country":    {{Slot: 1, CountryRecordID: civilizationID(1), ControllerKind: engine.ControllerHuman}, {Slot: 2, CountryRecordID: "CIVILIZATION_UNKNOWN", ControllerKind: engine.ControllerAI}},
		"unknown controller": {{Slot: 1, CountryRecordID: civilizationID(1), ControllerKind: engine.ControllerHuman}, {Slot: 2, CountryRecordID: civilizationID(2), ControllerKind: "robot"}},
	}
	for name, slots := range cases {
		t.Run(name, func(t *testing.T) {
			if err := Validate(ruleset, slots); err == nil {
				t.Fatal("invalid ordered slots accepted")
			}
		})
	}
}

func importedRuleset(t *testing.T, playable int) rules.Ruleset {
	t.Helper()
	directory := t.TempDir()
	file, err := os.Create(filepath.Join(directory, "Civilisation.ctp2synthetic"))
	if err != nil {
		t.Fatal(err)
	}
	for index := 1; index <= playable; index++ {
		if _, err := fmt.Fprintf(file, "%s { Playable true Special false Barbarian false }\n", civilizationID(index)); err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	if _, err := fmt.Fprintln(file, "CIVILIZATION_SPECIAL { Playable true Special true Barbarian false }"); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if _, err := fmt.Fprintln(file, "CIVILIZATION_BARBARIAN { Playable true Special false Barbarian true }"); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	ruleset, err := rules.Import(rules.ImportConfig{Directory: directory})
	if err != nil {
		t.Fatal(err)
	}
	return ruleset
}

func civilizationID(index int) string {
	return fmt.Sprintf("CIVILIZATION_SYNTHETIC_%02d", index)
}
