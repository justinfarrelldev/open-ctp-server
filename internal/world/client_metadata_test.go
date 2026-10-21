package world

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionedRiverGoodsGenerationPersistsAndIsDeterministic(t *testing.T) {
	input := fixtureGenerationInput()
	input.GeneratorVersion = "ctp2-rivers-goods-v2"
	input.HostParams.GeneratorVersion = input.GeneratorVersion
	first, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	AssertCanonicalJSONEqual(t, first, second)
	raw, err := first.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var restored WorldSnapshot
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	AssertCanonicalJSONEqual(t, first, restored)
	var tiles []map[string]any
	tileRaw, err := json.Marshal(restored.Tiles)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(tileRaw, &tiles); err != nil {
		t.Fatal(err)
	}
	rivers, goods := 0, 0
	for _, tile := range tiles {
		if tile["river"] == true {
			rivers++
			if tile["terrain"] == "ocean" || tile["terrain"] == "shelf" {
				t.Error("river generated on water")
			}
		}
		if good, ok := tile["good_id"].(string); ok && good != "" {
			goods++
			if !strings.HasPrefix(good, "TERRAIN_") {
				t.Errorf("noncanonical good ID %q", good)
			}
		}
	}
	if rivers == 0 || goods == 0 {
		t.Fatalf("versioned generation omitted independent metadata: rivers=%d goods=%d", rivers, goods)
	}
	profile := supportedMapProfiles[input.HostParams.MapSize]
	for source, tile := range restored.Tiles {
		if !tile.River {
			continue
		}
		seen := map[int]bool{source: true}
		queue := []int{source}
		reachesWater := false
		for head := 0; head < len(queue); head++ {
			for _, next := range cardinalNeighbors(queue[head], profile.width, profile.height) {
				if !isLandTerrain(restored.Tiles[next].Terrain) {
					reachesWater = true
				}
				if restored.Tiles[next].River && !seen[next] {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
		if !reachesWater {
			t.Errorf("river at %d never reaches water", source)
		}
	}
}

func TestRiverGoodsVersionPreservesBaseTerrainAndSpawnLineage(t *testing.T) {
	input := fixtureGenerationInput()
	legacy, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	input.GeneratorVersion = "ctp2-rivers-goods-v2"
	input.HostParams.GeneratorVersion = input.GeneratorVersion
	current, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(current.Spawns) != len(legacy.Spawns) || len(current.Tiles) != len(legacy.Tiles) {
		t.Fatal("metadata version changed base world shape")
	}
	for index, tile := range current.Tiles {
		if tile.X != legacy.Tiles[index].X || tile.Y != legacy.Tiles[index].Y || tile.Terrain != legacy.Tiles[index].Terrain {
			t.Fatalf("metadata-only version changed base terrain at %d", index)
		}
	}
	for index, spawn := range current.Spawns {
		if spawn != legacy.Spawns[index] {
			t.Fatalf("metadata-only version changed spawn %d", index)
		}
	}
}

func TestLegacyGeneratorDoesNotAcquireNewMetadata(t *testing.T) {
	snapshot, _, err := GenerateWorld(fixtureGenerationInput())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := snapshot.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"river"`) || strings.Contains(string(raw), `"good_id"`) {
		t.Fatalf("legacy generation changed: %s", raw)
	}
}
