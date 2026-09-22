package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeWorldConfig(t *testing.T) {
	config := WorldConfig{
		Style:            "  standard_ctp2  ",
		RulesJSON:        `  {"fast":true,"wet":false}  `,
		ExclusionsJSON:   `  {"units":["settler"]}  `,
		StartingSettlers: 2,
		StartingGold:     250,
		MapSize:          "  small  ",
		Climate:          "  temperate  ",
		GeneratorVersion: "  v1  ",
	}

	normalized, err := NormalizeWorldConfig(config)
	if err != nil {
		t.Fatalf("NormalizeWorldConfig() error = %v", err)
	}
	if normalized.Style != "standard_ctp2" {
		t.Fatalf("got style %q want standard_ctp2", normalized.Style)
	}
	if normalized.RulesJSON != `{"fast":true,"wet":false}` {
		t.Fatalf("got rules_json %q", normalized.RulesJSON)
	}
	if normalized.ExclusionsJSON != `{"units":["settler"]}` {
		t.Fatalf("got exclusions_json %q", normalized.ExclusionsJSON)
	}
	if normalized.MapSize != "small" || normalized.Climate != "temperate" || normalized.GeneratorVersion != "v1" {
		t.Fatalf("normalized string fields were not trimmed: %+v", normalized)
	}
}

func TestCanonicalWorldSerializationStable(t *testing.T) {
	snapshot := FixtureWorldSnapshot()
	AssertCanonicalJSONStable(t, snapshot)

	first := MustCanonicalString(t, snapshot)
	second := MustCanonicalString(t, snapshot.Clone())
	if first != second {
		t.Fatalf("snapshot clone changed canonical bytes\nfirst:  %s\nsecond: %s", first, second)
	}

	if !strings.Contains(first, `"game_id":44`) {
		t.Fatalf("expected canonical snapshot to include game_id, got %s", first)
	}
	if !strings.Contains(first, `"turn_state"`) {
		t.Fatalf("expected canonical snapshot to include turn_state, got %s", first)
	}
}

func TestDeterministicRNGSequenceStable(t *testing.T) {
	first := NewDeterministicRNG(12345)
	second := NewDeterministicRNG(12345)

	for i := 0; i < 5; i++ {
		var value1, value2 uint64
		firstNext, nextValue1 := first.Next()
		secondNext, nextValue2 := second.Next()
		first = firstNext.(SplitMix64RNG)
		second = secondNext.(SplitMix64RNG)
		value1 = nextValue1
		value2 = nextValue2
		if value1 != value2 {
			t.Fatalf("seeded sequences diverged at step %d: %d != %d", i, value1, value2)
		}
	}

	split := NewDeterministicRNG(12345).Split("terrain")
	if split.Seed() == 0 {
		t.Fatal("expected split RNG to produce a deterministic derived seed")
	}
}

func TestPureHarnessFilesAvoidForbiddenImports(t *testing.T) {
	files := []string{"config.go", "snapshot.go", "rng.go", "serialization.go", "test_helpers.go"}
	for _, file := range files {
		path := filepath.Join(".", file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("os.ReadFile(%q) error = %v", path, err)
		}
		text := string(content)
		for _, forbidden := range []string{"net/http", "database/sql", "github.com/jmoiron/sqlx", "internal/auth"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s unexpectedly imports %q", file, forbidden)
			}
		}
	}
}
