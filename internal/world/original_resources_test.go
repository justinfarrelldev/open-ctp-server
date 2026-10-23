package world

import (
	"reflect"
	"testing"
)

// Original release terrain.txt Resources, not the four-resource GOG extension.
func originalTestGoods(terrain string) []string {
	families := map[string]string{"forest": "FOREST", "plains": "PLAINS", "grassland": "GRASSLAND", "desert": "DESERT", "mountain": "MOUNTAIN", "ocean": "WATER_DEEP"}
	if terrain == "swamp" {
		return []string{"TERRAIN_SWAMP_GOOD_ONE"}
	}
	if terrain == "tundra" {
		return []string{"TERRAIN_TUNDRA_GOOD_ONE"}
	}
	family, ok := families[terrain]
	if !ok {
		return nil
	}
	return []string{"TERRAIN_" + family + "_GOOD_ONE", "TERRAIN_" + family + "_GOOD_TWO"}
}

func TestOriginalResourceGenerationNeverInventsCatalogIDs(t *testing.T) {
	count := 0
	seen := map[string]bool{}
	for seed := int64(1); seed <= 100; seed++ {
		input := fixtureGenerationInput()
		input.GeneratorVersion = "ctp2-original-goods-v4"
		input.HostParams.GeneratorVersion = input.GeneratorVersion
		input.HostParams.Seed = &seed
		snapshot, _, err := GenerateWorld(input)
		if err != nil {
			t.Fatal(err)
		}
		for _, tile := range snapshot.Tiles {
			if tile.GoodID != "" {
				count++
				seen[tile.Terrain] = true
				valid := false
				for _, good := range originalTestGoods(tile.Terrain) {
					if tile.GoodID == good {
						valid = true
					}
				}
				if !valid {
					t.Fatalf("undefined original terrain/good pair %s/%s", tile.Terrain, tile.GoodID)
				}
			}
		}
	}
	if count == 0 || len(seen) < 5 {
		t.Fatalf("no meaningful generated resources: %d %v", count, seen)
	}
}

func TestOriginalResourceVersionRetainsOtherGenerationLineage(t *testing.T) {
	input := fixtureGenerationInput()
	input.GeneratorVersion = RuinGeneratorVersion
	input.HostParams.GeneratorVersion = input.GeneratorVersion
	old, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	input.GeneratorVersion = "ctp2-original-goods-v4"
	input.HostParams.GeneratorVersion = input.GeneratorVersion
	current, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(old.Spawns, current.Spawns) {
		t.Fatal("resource correction changed spawns")
	}
	for n, tile := range old.Tiles {
		other := current.Tiles[n]
		tile.GoodID = ""
		other.GoodID = ""
		if tile != other {
			t.Fatalf("terrain/river/ruin lineage changed at %d: %+v %+v", n, tile, other)
		}
	}
	repeat, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	AssertCanonicalJSONEqual(t, current, repeat)
}
