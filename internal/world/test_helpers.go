package world

import (
	"strings"
	"testing"
	"time"
)

// FixtureWorldConfig returns a representative normalized config for tests.
func FixtureWorldConfig() WorldConfig {
	seed := int64(12345)
	wetness := 7
	config := WorldConfig{
		Style:            "standard_ctp2",
		RulesJSON:        `{"fast":true}`,
		ExclusionsJSON:   `{"units":["settler"]}`,
		StartingSettlers: 2,
		StartingGold:     250,
		Seed:             &seed,
		Wetness:          &wetness,
		MapSize:          "small",
		Climate:          "temperate",
		GeneratorVersion: "v1",
	}
	return config
}

// FixtureMutableTurnState returns a representative mutable turn state for tests.
func FixtureMutableTurnState() MutableTurnState {
	return MutableTurnState{
		TurnNumber:         1,
		CurrentPlayerIndex: 0,
		ActivePlayerOrder:  []int64{7, 8},
		RecalculationCount: 1,
		RNGState:           99,
		Ratings:            WorldRatings{Overall: 10, Science: 3, Military: 2, Economic: 4, Pollution: 1},
		Pollution:          WorldPollution{GlobalPollution: 5, TilePollutionCount: 2, UpdatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)},
		LastRecalculatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
}

// FixtureWorldSnapshot returns a representative immutable world snapshot for tests.
func FixtureWorldSnapshot() WorldSnapshot {
	return WorldSnapshot{
		GameID:           44,
		GameStatus:       "started",
		GeneratorVersion: "v1",
		Config:           FixtureWorldConfig(),
		PlayerOrder:      []int64{7, 8},
		Spawns: []PlayerSpawn{
			{AccountID: 7, Country: "Greeks", X: 3, Y: 4},
			{AccountID: 8, Country: "Romans", X: 9, Y: 10},
		},
		Tiles: []WorldTile{
			{X: 0, Y: 0, Terrain: "ocean"},
			{X: 1, Y: 0, Terrain: "grassland"},
		},
		Turn: FixtureMutableTurnState(),
	}
}

// AssertCanonicalJSONStable verifies that repeated serialization yields identical bytes.
func AssertCanonicalJSONStable(t testing.TB, value any) {
	t.Helper()

	first := MustCanonicalJSON(t, value)
	second := MustCanonicalJSON(t, value)
	if string(first) != string(second) {
		t.Fatalf("canonical JSON changed between runs\nfirst:  %s\nsecond: %s", first, second)
	}
}

// AssertContainsAll checks that every needle appears in the haystack.
func AssertContainsAll(t testing.TB, haystack string, needles ...string) {
	t.Helper()

	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			t.Fatalf("expected %q to contain %q", haystack, needle)
		}
	}
}
