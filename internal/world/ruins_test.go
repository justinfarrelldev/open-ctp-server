package world

import (
	"encoding/json"
	"testing"
)

func TestRuinGeneratorVersionPreservesV2AndAddsDeterministicLandRuins(t *testing.T) {
	input := fixtureGenerationInput()
	input.GeneratorVersion = RiverGoodsGeneratorVersion
	input.HostParams.GeneratorVersion = input.GeneratorVersion
	old, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	input.GeneratorVersion = "ctp2-ruins-v3"
	input.HostParams.GeneratorVersion = input.GeneratorVersion
	current, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	repeated, _, err := GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	AssertCanonicalJSONEqual(t, current, repeated)
	raw, _ := current.CanonicalJSON()
	var restored WorldSnapshot
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	AssertCanonicalJSONEqual(t, current, restored)
	raw, _ = json.Marshal(current.Tiles)
	var tiles []map[string]any
	_ = json.Unmarshal(raw, &tiles)
	count := 0
	boxes := map[[2]int]bool{}
	for i, tile := range current.Tiles {
		prior := old.Tiles[i]
		if tile.X != prior.X || tile.Y != prior.Y || tile.Terrain != prior.Terrain || tile.River != prior.River || tile.GoodID != prior.GoodID {
			t.Fatal("new generator changed v2 terrain/metadata")
		}
		if tiles[i]["ruin"] != true {
			continue
		}
		count++
		if !isLandTerrain(tile.Terrain) {
			t.Fatal("water ruin")
		}
		box := [2]int{tile.X / 6, tile.Y / 6}
		if boxes[box] {
			t.Fatal("multiple ruins per source box")
		}
		boxes[box] = true
		for _, spawn := range current.Spawns {
			if spawn.X == tile.X && spawn.Y == tile.Y {
				t.Fatal("ruin under starting unit")
			}
		}
	}
	if count == 0 {
		t.Fatal("no ruins generated")
	}
	for i, spawn := range current.Spawns {
		if spawn != old.Spawns[i] {
			t.Fatal("ruins changed spawn stream")
		}
	}
	raw, _ = json.Marshal(old.Tiles)
	tiles = nil
	_ = json.Unmarshal(raw, &tiles)
	for _, tile := range tiles {
		if tile["ruin"] == true {
			t.Fatal("legacy v2 acquired ruins")
		}
	}
}
