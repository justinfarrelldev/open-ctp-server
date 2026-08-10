package world

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	fixtures "github.com/justinfarrelldev/open-ctp-server/internal/world/fixtures"
)

func TestWorldGenerationDeterministic(t *testing.T) {
	input := fixtureGenerationInput()

	firstSnapshot, firstTurn, err := GenerateWorld(input)
	if err != nil {
		t.Fatalf("GenerateWorld() first error = %v", err)
	}
	secondSnapshot, secondTurn, err := GenerateWorld(input)
	if err != nil {
		t.Fatalf("GenerateWorld() second error = %v", err)
	}

	AssertCanonicalJSONEqual(t, firstSnapshot, secondSnapshot)
	AssertCanonicalJSONEqual(t, firstTurn, secondTurn)
	if len(firstSnapshot.Tiles) == 0 {
		t.Fatal("expected deterministic world tiles")
	}
	if len(firstSnapshot.Spawns) != len(input.PlayerOrder) {
		t.Fatalf("spawn count=%d want %d", len(firstSnapshot.Spawns), len(input.PlayerOrder))
	}
	if firstTurn.TurnNumber != 1 || firstTurn.CurrentPlayerIndex != 0 {
		t.Fatalf("unexpected initial turn state: %+v", firstTurn)
	}
}

func TestVictorySelectorDoesNotChangeWorldGeneration(t *testing.T) {
	base := fixtureGenerationInput()
	selected := cloneGenerationInput(base)
	selected.HostParams.RulesJSON = `["victory:bloodlust_conquest"]`

	baseSeed, err := deriveGeneratorSeed(base)
	if err != nil {
		t.Fatal(err)
	}
	selectedSeed, err := deriveGeneratorSeed(selected)
	if err != nil {
		t.Fatal(err)
	}
	if selectedSeed != baseSeed {
		t.Fatalf("victory selector changed generator seed from %d to %d", baseSeed, selectedSeed)
	}
}

func TestGenerationConsumesAllHostParams(t *testing.T) {
	base := fixtureGenerationInput()
	baseSnapshot, _, err := GenerateWorld(base)
	if err != nil {
		t.Fatalf("GenerateWorld(base) error = %v", err)
	}
	baseKey := MustCanonicalString(t, baseSnapshot)

	tests := []struct {
		name   string
		mutate func(*GenerationInput)
	}{
		{name: "style", mutate: func(input *GenerationInput) { input.HostParams.Style = "scenario_ctp2" }},
		{name: "rules_json", mutate: func(input *GenerationInput) { input.HostParams.RulesJSON = `{"fast":true}` }},
		{name: "exclusions_json", mutate: func(input *GenerationInput) { input.HostParams.ExclusionsJSON = `{"units":["settler"]}` }},
		{name: "starting_settlers", mutate: func(input *GenerationInput) { input.HostParams.StartingSettlers = 2 }},
		{name: "starting_gold", mutate: func(input *GenerationInput) { input.HostParams.StartingGold = 250 }},
		{name: "seed", mutate: func(input *GenerationInput) { seed := int64(99123); input.HostParams.Seed = &seed }},
		{name: "wetness", mutate: func(input *GenerationInput) { wetness := 80; input.HostParams.Wetness = &wetness }},
		{name: "map_size", mutate: func(input *GenerationInput) { input.HostParams.MapSize = "medium" }},
		{name: "climate", mutate: func(input *GenerationInput) { input.HostParams.Climate = "tropical" }},
		{name: "generator_version", mutate: func(input *GenerationInput) {
			input.GeneratorVersion = "ctp2-first-v2"
			input.HostParams.GeneratorVersion = "ctp2-first-v2"
		}},
		{name: "player_order", mutate: func(input *GenerationInput) { input.PlayerOrder = []int64{303, 202, 101} }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mutated := cloneGenerationInput(base)
			tc.mutate(&mutated)

			snapshot, _, err := GenerateWorld(mutated)
			if err != nil {
				t.Fatalf("GenerateWorld(%s) error = %v", tc.name, err)
			}
			if got := MustCanonicalString(t, snapshot); got == baseKey {
				t.Fatalf("changing %s did not alter canonical snapshot", tc.name)
			}
		})
	}

	wetFixture, _ := fixtures.ByName("small-wet")
	dryFixture, _ := fixtures.ByName("small-dry")
	wetResponse, err := GenerateWorldFromFixture(wetFixture)
	if err != nil {
		t.Fatalf("GenerateWorldFromFixture(wet) error = %v", err)
	}
	dryResponse, err := GenerateWorldFromFixture(dryFixture)
	if err != nil {
		t.Fatalf("GenerateWorldFromFixture(dry) error = %v", err)
	}
	if countTerrain(wetResponse.Tiles, "desert") >= countTerrain(dryResponse.Tiles, "desert") {
		t.Fatalf("wet fixture should produce fewer desert tiles than dry fixture")
	}
}

func TestWorldGenerationRejectsInvalidParams(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*GenerationInput)
		needle string
	}{
		{name: "missing seed", mutate: func(input *GenerationInput) { input.HostParams.Seed = nil }, needle: "seed must be specified"},
		{name: "missing wetness", mutate: func(input *GenerationInput) { input.HostParams.Wetness = nil }, needle: "wetness must be specified"},
		{name: "wetness range", mutate: func(input *GenerationInput) { wetness := 101; input.HostParams.Wetness = &wetness }, needle: "wetness must be between 0 and 100"},
		{name: "missing generator version", mutate: func(input *GenerationInput) { input.GeneratorVersion = "" }, needle: "generator_version must be specified"},
		{name: "missing map size", mutate: func(input *GenerationInput) { input.HostParams.MapSize = "" }, needle: "map_size must be specified"},
		{name: "unsupported climate", mutate: func(input *GenerationInput) { input.HostParams.Climate = "volcanic" }, needle: "unsupported climate"},
		{name: "duplicate player", mutate: func(input *GenerationInput) { input.PlayerOrder = []int64{11, 11, 12} }, needle: "duplicate account_id"},
		{name: "negative settlers", mutate: func(input *GenerationInput) { input.HostParams.StartingSettlers = -1 }, needle: "starting_settlers must be non-negative"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := cloneGenerationInput(fixtureGenerationInput())
			tc.mutate(&input)

			_, _, err := GenerateWorld(input)
			if err == nil {
				t.Fatalf("GenerateWorld(%s) error = nil", tc.name)
			}
			if !errors.Is(err, ErrWorldGenerationInvalidInput) {
				t.Fatalf("GenerateWorld(%s) error = %v, want invalid input", tc.name, err)
			}
			if !strings.Contains(err.Error(), tc.needle) {
				t.Fatalf("GenerateWorld(%s) error = %q, want substring %q", tc.name, err.Error(), tc.needle)
			}
		})
	}
}

func TestWorldGenerationCapsRetries(t *testing.T) {
	input := fixtureGenerationInput()
	input.PlayerOrder = make([]int64, 64)
	for i := range input.PlayerOrder {
		input.PlayerOrder[i] = int64(1000 + i)
	}

	_, _, err := GenerateWorld(input)
	if err == nil {
		t.Fatal("GenerateWorld() error = nil, want capped retry failure")
	}
	if !errors.Is(err, ErrWorldGenerationRetryCap) {
		t.Fatalf("GenerateWorld() error = %v, want retry cap", err)
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("after %d attempts", maxSpawnPlacementAttempts)) {
		t.Fatalf("retry cap error %q must mention deterministic attempt cap", err.Error())
	}
}

func TestWorldGenerationSupportsLargerDeterministicMaps(t *testing.T) {
	tests := []struct {
		name      string
		mapSize   string
		climate   string
		playerCnt int
	}{
		{name: "medium sixteen players", mapSize: "medium", climate: "temperate", playerCnt: 16},
		{name: "large eighteen players", mapSize: "large", climate: "temperate", playerCnt: 18},
		{name: "huge twenty players", mapSize: "huge", climate: "temperate", playerCnt: 20},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := cloneGenerationInput(fixtureGenerationInput())
			input.HostParams.MapSize = tc.mapSize
			input.HostParams.Climate = tc.climate
			input.PlayerOrder = make([]int64, tc.playerCnt)
			for i := range input.PlayerOrder {
				input.PlayerOrder[i] = int64(5000 + i)
			}

			snapshot, _, err := GenerateWorld(input)
			if err != nil {
				t.Fatalf("GenerateWorld(%s) error = %v", tc.name, err)
			}
			if len(snapshot.Spawns) != tc.playerCnt {
				t.Fatalf("spawn count=%d want %d", len(snapshot.Spawns), tc.playerCnt)
			}
		})
	}
}

func TestMergeContinentsAssignsBridgedLandToSingleContinent(t *testing.T) {
	profile := mapProfile{width: 3, height: 2, wrapX: false, wrapY: false}
	tiles := []WorldTile{
		{X: 0, Y: 0, Terrain: "plains"},
		{X: 1, Y: 0, Terrain: "shelf"},
		{X: 2, Y: 0, Terrain: "plains"},
		{X: 0, Y: 1, Terrain: "plains"},
		{X: 1, Y: 1, Terrain: "plains"},
		{X: 2, Y: 1, Terrain: "plains"},
	}
	continents := make([][]int, profile.height)
	for y := range continents {
		continents[y] = make([]int, profile.width)
	}

	mergeContinents(tiles, continents, profile)

	want := continents[0][0]
	if want == 0 {
		t.Fatal("expected land tile to receive a continent id")
	}
	for _, coordinate := range [][2]int{{2, 0}, {0, 1}, {1, 1}, {2, 1}} {
		if got := continents[coordinate[1]][coordinate[0]]; got != want {
			t.Fatalf("continent at (%d,%d)=%d want %d", coordinate[0], coordinate[1], got, want)
		}
	}
	if got := continents[0][1]; got != 0 {
		t.Fatalf("water tile continent=%d want 0", got)
	}
}

func fixtureGenerationInput() GenerationInput {
	seed := int64(41041)
	wetness := 50
	return GenerationInput{
		GameID:           44,
		PlayerOrder:      []int64{101, 202, 303},
		GeneratorVersion: "ctp2-first-v1",
		HostParams: GenerationHostParams{
			Style:            "standard_ctp2",
			RulesJSON:        "[]",
			ExclusionsJSON:   "{}",
			StartingSettlers: 1,
			StartingGold:     100,
			Seed:             &seed,
			Wetness:          &wetness,
			MapSize:          "small",
			Climate:          "temperate",
			GeneratorVersion: "ctp2-first-v1",
		},
	}
}

func cloneGenerationInput(input GenerationInput) GenerationInput {
	clone := input
	clone.PlayerOrder = append([]int64(nil), input.PlayerOrder...)
	if input.HostParams.Seed != nil {
		seed := *input.HostParams.Seed
		clone.HostParams.Seed = &seed
	}
	if input.HostParams.Wetness != nil {
		wetness := *input.HostParams.Wetness
		clone.HostParams.Wetness = &wetness
	}
	return clone
}

func countTerrain(tiles []WorldTile, terrain string) int {
	count := 0
	for _, tile := range tiles {
		if tile.Terrain == terrain {
			count++
		}
	}
	return count
}
