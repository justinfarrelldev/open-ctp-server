package engine

import (
	"bytes"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// Visibility tests exercise the authorization boundary; these exercise the
// engine's exported inspection API and its read-only/unknown-value guarantees.
func TestClientInspectorPublicFactsAreReadOnlyAndExplicit(t *testing.T) {
	state := task10State()
	pin, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: pin.CompatibilityVersion, ContentHash: pin.ContentHash, Definitions: pin.CanonicalJSON}
	state.Nations = []Nation{{ID: "nation-a", CurrentAgeID: "ancient"}}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2, Effects: []Effect{{ID: "capital", Kind: EffectCapitol}}}}
	state.Tiles[0].Pollution = 7
	state.Overlays = []TileOverlay{{ID: "z", TileID: "b", OwnerID: "player-a", Kind: "improvement", ImprovementID: "road", Remaining: 2}, {ID: "a", TileID: "b", OwnerID: "player-a", Kind: "improvement", ImprovementID: "farm"}}
	before, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	inspector, err := NewInspector(state)
	if err != nil {
		t.Fatal(err)
	}
	facts := inspector.Tile("a")
	if facts == nil || facts.Yields == nil || facts.YieldContext != "city-center" || facts.Pollution != 7 || facts.MovementCost == nil {
		t.Fatalf("effective city facts absent: %+v", facts)
	}
	terrain := inspector.Tile("b")
	if terrain == nil || terrain.Yields == nil || terrain.YieldContext != "terrain" || len(terrain.Overlays) != 2 || terrain.Overlays[0].ImprovementID != "farm" {
		t.Fatalf("terrain facts/ordering absent: %+v", terrain)
	}
	art := inspector.CityArt("home")
	if art == nil || art.AgeID != "ancient" || !art.Capital {
		t.Fatalf("semantic city art absent: %+v", art)
	}
	if inspector.Tile("missing") != nil || inspector.MovementCost("missing") != nil || inspector.CityArt("missing") != nil || inspector.Calendar() != nil || inspector.CanDisband("home", "player-b") {
		t.Fatal("unavailable facts were fabricated")
	}
	after, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("inspection mutated authority")
	}
}

func TestClientInspectorUsesPinnedCalendarAndRejectsInvalidRules(t *testing.T) {
	year := -4000
	inspector := Inspector{state: State{Phase: PhaseState{Turn: 3}}, runtime: &rules.RuntimeRules{Calendar: rules.RuntimeCalendar{StartYear: &year, EpochTurn: 1, Periods: []rules.RuntimeCalendarPeriod{{StartTurn: 0, YearsPerTurn: 50}}}}}
	if display := inspector.Calendar(); display == nil || display.Year != -3900 || display.Turn != 3 {
		t.Fatalf("pinned calendar was not used: %+v", display)
	}
	state := task10State()
	state.Rules.Definitions = []byte(`{"compatibility_version":"invalid"}`)
	if _, err := NewInspector(state); err == nil {
		t.Fatal("invalid rules acquired a display fallback")
	}
}
