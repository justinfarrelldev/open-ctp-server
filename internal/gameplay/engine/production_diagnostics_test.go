package engine

import (
	"bytes"
	"errors"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"testing"
)

func TestAtomicProductionBatchLimitsAndTypedRollback(t *testing.T) {
	state := task10State()
	pin, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := rules.CompileRuntime(pin.CanonicalJSON, pin.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 3, StoredProduction: 40}}
	entries := make([]QueueSelection, MaxProductionQueueEntries)
	for n := range entries {
		entries[n] = QueueSelection{ItemID: runtime.Units[0].ID}
	}
	for _, tc := range []struct {
		entries []QueueSelection
		code    string
	}{
		{append(append([]QueueSelection(nil), entries...), entries[0]), "queue_capacity"},
		{[]QueueSelection{{ItemID: runtime.Units[0].ID}, {ItemID: "unknown"}}, "content_unavailable"},
		{[]QueueSelection{{ItemID: runtime.Units[0].ID}, {EntryID: "foreign"}}, "queue_identity"},
		{[]QueueSelection{{ItemID: runtime.Units[0].ID, EntryID: "both"}}, "queue_shape"},
		{nil, "queue_shape"},
	} {
		before, _ := CanonicalStateJSON(state)
		err := editProduction(&state, "player-a", EditProductionPayload{CityID: "home", Entries: tc.entries}, &runtime)
		var rejected *ProductionError
		if !errors.As(err, &rejected) || rejected.Code != tc.code {
			t.Fatalf("expected %s: %v", tc.code, err)
		}
		after, _ := CanonicalStateJSON(state)
		if !bytes.Equal(before, after) {
			t.Fatal("invalid batch consumed queue identity/funding")
		}
	}
	if err := editProduction(&state, "player-a", EditProductionPayload{CityID: "home", Entries: entries}, &runtime); err != nil {
		t.Fatal(err)
	}
	if len(state.Cities[0].Queue) != MaxProductionQueueEntries || state.Cities[0].StoredProduction != 40 {
		t.Fatal("valid batch lost entries/funding")
	}
	seen := map[string]bool{}
	for _, entry := range state.Cities[0].Queue {
		if entry.EntryID == "" || seen[entry.EntryID] {
			t.Fatal("nonindependent batch identities")
		}
		seen[entry.EntryID] = true
	}
}

func TestProductionPrerequisitesKeepDistinctPrivateDiagnosticBranches(t *testing.T) {
	state := task10State()
	city := City{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2}
	messages := map[string]bool{}
	for _, effect := range []string{EffectCantBuild, EffectHasPopAndCanBuild, EffectSpecialCreatePark} {
		runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "candidate", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{Kind: effect}}}}}
		_, err := validateProductionAppend(state, "player-a", city, ProductionItem{ID: "candidate"}, runtime)
		var rejected *ProductionError
		if !errors.As(err, &rejected) || rejected.Code != "production_prerequisite" {
			t.Fatalf("wrong public category: %v", err)
		}
		if messages[err.Error()] {
			t.Fatalf("distinct private branches discarded: %s", err)
		}
		messages[err.Error()] = true
	}
}
