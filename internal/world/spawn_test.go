package world

import (
	"errors"
	"fmt"
	"testing"

	fixtures "github.com/justinfarrelldev/open-ctp-server/internal/world/fixtures"
)

func TestPlayerSpawnsAreValid(t *testing.T) {
	for _, definition := range fixtures.Inventory {
		definition := definition
		t.Run(definition.Name, func(t *testing.T) {
			response, err := GenerateWorldFromFixture(definition)
			if err != nil {
				t.Fatalf("GenerateWorldFromFixture(%s) error = %v", definition.Name, err)
			}

			profile, ok := supportedMapProfiles[definition.HostParams.MapSize]
			if !ok {
				t.Fatalf("unsupported map size %q in fixture", definition.HostParams.MapSize)
			}

			if len(response.Spawns) != len(definition.PlayerOrder) {
				t.Fatalf("spawn count=%d want %d", len(response.Spawns), len(definition.PlayerOrder))
			}

			continents := continentsFromTiles(response.Tiles, profile)
			tilesByCoordinate := worldTilesByCoordinate(response.Tiles)
			seenCoordinates := make(map[[2]int]int64, len(response.Spawns))
			minDistance := 3

			for i, spawn := range response.Spawns {
				if spawn.AccountID != definition.PlayerOrder[i] {
					t.Fatalf("spawn[%d].account_id=%d want %d", i, spawn.AccountID, definition.PlayerOrder[i])
				}
				if spawn.X < 0 || spawn.X >= profile.width || spawn.Y < 0 || spawn.Y >= profile.height {
					t.Fatalf("spawn[%d]=(%d,%d) out of bounds for %dx%d map", i, spawn.X, spawn.Y, profile.width, profile.height)
				}

				coordinate := [2]int{spawn.X, spawn.Y}
				tile, ok := tilesByCoordinate[coordinate]
				if !ok {
					t.Fatalf("spawn[%d]=(%d,%d) has no backing tile", i, spawn.X, spawn.Y)
				}
				if !isSpawnableTerrain(tile.Terrain) {
					t.Fatalf("spawn[%d] terrain=%q is not spawnable", i, tile.Terrain)
				}
				if continents[spawn.Y][spawn.X] == 0 {
					t.Fatalf("spawn[%d]=(%d,%d) must be on a land continent", i, spawn.X, spawn.Y)
				}
				if previousAccount, exists := seenCoordinates[coordinate]; exists {
					t.Fatalf("spawn[%d] duplicates coordinate (%d,%d) already used by account %d", i, spawn.X, spawn.Y, previousAccount)
				}
				seenCoordinates[coordinate] = spawn.AccountID
			}

			for i := 0; i < len(response.Spawns); i++ {
				for j := i + 1; j < len(response.Spawns); j++ {
					left := response.Spawns[i]
					right := response.Spawns[j]
					if got := wrappedManhattanDistance(left.X, left.Y, right.X, right.Y, profile.width, profile.height, profile.wrapX, profile.wrapY); got < minDistance {
						t.Fatalf("spawn distance between account %d and %d = %d want >= %d", left.AccountID, right.AccountID, got, minDistance)
					}
				}
			}
		})
	}

	t.Run("horizontal wrap distance blocks edge-overlapping starts", func(t *testing.T) {
		profile := mapProfile{width: 24, height: 1, wrapX: true, wrapY: false, baseWaterTarget: 50, baseSetupRadius: 2}
		input := fixtureGenerationInput()
		input.PlayerOrder = []int64{101, 202}
		tiles := []WorldTile{
			{X: 0, Y: 0, Terrain: "grassland"},
			{X: 12, Y: 0, Terrain: "grassland"},
			{X: 23, Y: 0, Terrain: "grassland"},
		}
		continents := [][]int{{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}}

		spawns, _, err := generateSpawns(input, profile, tiles, continents, NewDeterministicRNG(99))
		if err != nil {
			t.Fatalf("generateSpawns() error = %v", err)
		}
		if len(spawns) != len(input.PlayerOrder) {
			t.Fatalf("spawn count=%d want %d", len(spawns), len(input.PlayerOrder))
		}
		if usesCoordinate(spawns, 0, 0) && usesCoordinate(spawns, 23, 0) {
			t.Fatalf("wrap-aware selection chose overlapping edge starts: %+v", spawns)
		}
	})
}

func TestPlayerSpawnsAreValidOnLargerProfiles(t *testing.T) {
	tests := []struct {
		name        string
		mapSize     string
		climate     string
		playerCount int
		minDistance int
	}{
		{name: "medium sixteen players", mapSize: "medium", climate: "temperate", playerCount: 16, minDistance: 5},
		{name: "large eighteen players", mapSize: "large", climate: "temperate", playerCount: 18, minDistance: 5},
		{name: "huge twenty players", mapSize: "huge", climate: "temperate", playerCount: 20, minDistance: 7},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := cloneGenerationInput(fixtureGenerationInput())
			input.HostParams.MapSize = tc.mapSize
			input.HostParams.Climate = tc.climate
			input.PlayerOrder = make([]int64, tc.playerCount)
			for i := range input.PlayerOrder {
				input.PlayerOrder[i] = int64(5000 + i)
			}

			snapshot, _, err := GenerateWorld(input)
			if err != nil {
				t.Fatalf("GenerateWorld(%s) error = %v", tc.name, err)
			}

			profile, ok := supportedMapProfiles[tc.mapSize]
			if !ok {
				t.Fatalf("unsupported map size %q", tc.mapSize)
			}
			assertSpawnSetValid(t, snapshot.Spawns, snapshot.Tiles, input.PlayerOrder, profile, tc.minDistance)
		})
	}
}

func TestPlayerSpawnsDeterministic(t *testing.T) {
	for _, definition := range fixtures.Inventory {
		definition := definition
		t.Run(definition.Name, func(t *testing.T) {
			first, err := GenerateWorldFromFixture(definition)
			if err != nil {
				t.Fatalf("GenerateWorldFromFixture(%s) first error = %v", definition.Name, err)
			}
			wantSpawns := MustCanonicalString(t, first.Spawns)
			wantWorld := MustCanonicalString(t, first)

			for iteration := 0; iteration < 8; iteration++ {
				again, err := GenerateWorldFromFixture(definition)
				if err != nil {
					t.Fatalf("GenerateWorldFromFixture(%s) iteration %d error = %v", definition.Name, iteration, err)
				}
				if got := MustCanonicalString(t, again.Spawns); got != wantSpawns {
					t.Fatalf("spawn output changed on iteration %d\ngot:  %s\nwant: %s", iteration, got, wantSpawns)
				}
				if got := MustCanonicalString(t, again); got != wantWorld {
					t.Fatalf("fixture world output changed on iteration %d", iteration)
				}
			}
		})
	}
}

func TestStartPlacementFailsWhenNoValidSpawn(t *testing.T) {
	profile := mapProfile{width: 5, height: 1, wrapX: false, wrapY: false, baseWaterTarget: 50, baseSetupRadius: 2}
	input := fixtureGenerationInput()
	input.PlayerOrder = []int64{101, 202}
	tiles := []WorldTile{
		{X: 0, Y: 0, Terrain: "grassland"},
		{X: 1, Y: 0, Terrain: "plains"},
	}
	continents := [][]int{{1, 1, 0, 0, 0}}

	spawns, _, err := generateSpawns(input, profile, tiles, continents, NewDeterministicRNG(77))
	if err == nil {
		t.Fatal("generateSpawns() error = nil, want deterministic retry-cap failure")
	}
	if len(spawns) != 0 {
		t.Fatalf("generateSpawns() returned partial spawns: %+v", spawns)
	}
	if !errors.Is(err, ErrWorldGenerationRetryCap) {
		t.Fatalf("generateSpawns() error = %v, want retry cap", err)
	}

	want := fmt.Sprintf("%s: placed 1 of 2 spawns after %d attempts", ErrWorldGenerationRetryCap.Error(), maxSpawnPlacementAttempts)
	if err.Error() != want {
		t.Fatalf("generateSpawns() error = %q want %q", err.Error(), want)
	}
}

func assertSpawnSetValid(t *testing.T, spawns []PlayerSpawn, tiles []WorldTile, playerOrder []int64, profile mapProfile, minDistance int) {
	t.Helper()

	if len(spawns) != len(playerOrder) {
		t.Fatalf("spawn count=%d want %d", len(spawns), len(playerOrder))
	}

	continents := continentsFromTiles(tiles, profile)
	tilesByCoordinate := worldTilesByCoordinate(tiles)
	seenCoordinates := make(map[[2]int]int64, len(spawns))

	for i, spawn := range spawns {
		if spawn.AccountID != playerOrder[i] {
			t.Fatalf("spawn[%d].account_id=%d want %d", i, spawn.AccountID, playerOrder[i])
		}
		if spawn.X < 0 || spawn.X >= profile.width || spawn.Y < 0 || spawn.Y >= profile.height {
			t.Fatalf("spawn[%d]=(%d,%d) out of bounds for %dx%d map", i, spawn.X, spawn.Y, profile.width, profile.height)
		}

		coordinate := [2]int{spawn.X, spawn.Y}
		tile, ok := tilesByCoordinate[coordinate]
		if !ok {
			t.Fatalf("spawn[%d]=(%d,%d) has no backing tile", i, spawn.X, spawn.Y)
		}
		if !isSpawnableTerrain(tile.Terrain) {
			t.Fatalf("spawn[%d] terrain=%q is not spawnable", i, tile.Terrain)
		}
		if continents[spawn.Y][spawn.X] == 0 {
			t.Fatalf("spawn[%d]=(%d,%d) must be on a land continent", i, spawn.X, spawn.Y)
		}
		if previousAccount, exists := seenCoordinates[coordinate]; exists {
			t.Fatalf("spawn[%d] duplicates coordinate (%d,%d) already used by account %d", i, spawn.X, spawn.Y, previousAccount)
		}
		seenCoordinates[coordinate] = spawn.AccountID
	}

	for i := 0; i < len(spawns); i++ {
		for j := i + 1; j < len(spawns); j++ {
			left := spawns[i]
			right := spawns[j]
			if got := wrappedManhattanDistance(left.X, left.Y, right.X, right.Y, profile.width, profile.height, profile.wrapX, profile.wrapY); got < minDistance {
				t.Fatalf("spawn distance between account %d and %d = %d want >= %d", left.AccountID, right.AccountID, got, minDistance)
			}
		}
	}
}

func wrappedManhattanDistance(ax, ay, bx, by, width, height int, wrapX, wrapY bool) int {
	dx := absInt(ax - bx)
	if wrapX {
		wrapped := width - dx
		if wrapped < dx {
			dx = wrapped
		}
	}
	dy := absInt(ay - by)
	if wrapY {
		wrapped := height - dy
		if wrapped < dy {
			dy = wrapped
		}
	}
	return dx + dy
}

func continentsFromTiles(tiles []WorldTile, profile mapProfile) [][]int {
	continents := make([][]int, profile.height)
	for y := range continents {
		continents[y] = make([]int, profile.width)
	}
	mergeContinents(tiles, continents, profile)
	return continents
}

func worldTilesByCoordinate(tiles []WorldTile) map[[2]int]WorldTile {
	indexed := make(map[[2]int]WorldTile, len(tiles))
	for _, tile := range tiles {
		indexed[[2]int{tile.X, tile.Y}] = tile
	}
	return indexed
}

func usesCoordinate(spawns []PlayerSpawn, x, y int) bool {
	for _, spawn := range spawns {
		if spawn.X == x && spawn.Y == y {
			return true
		}
	}
	return false
}
