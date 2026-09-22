package fixtures

import "testing"

func TestInventoryHasStableValidatedWorldDefinitions(t *testing.T) {
	wantNames := []string{"small-default", "small-wet", "small-dry"}
	if len(Inventory) != len(wantNames) {
		t.Fatalf("inventory length = %d, want %d", len(Inventory), len(wantNames))
	}
	seeds := make(map[int64]struct{}, len(Inventory))
	for index, definition := range Inventory {
		if definition.Name != wantNames[index] {
			t.Fatalf("inventory[%d].Name = %q, want %q", index, definition.Name, wantNames[index])
		}
		if definition.GeneratorVersion != GeneratorVersion || definition.HostParams.MapSize != "small" || definition.HostParams.Climate != "temperate" {
			t.Fatalf("inventory[%d] has unstable generator inputs: %#v", index, definition)
		}
		if definition.MapParams.PlayerCount != len(definition.PlayerOrder) || !definition.Expected.SpawnCountMatchesPlayerOrder || !definition.Expected.UniquePlayerStarts || !definition.Expected.RequiresLandContinents {
			t.Fatalf("inventory[%d] has invalid player/start invariants: %#v", index, definition)
		}
		if len(definition.SourceReferences) == 0 || definition.Expected.TerrainBias == "" || definition.Expected.StartDistanceRule == "" {
			t.Fatalf("inventory[%d] lacks source-backed expectations", index)
		}
		if _, duplicate := seeds[definition.HostParams.Seed]; duplicate {
			t.Fatalf("duplicate fixture seed %d", definition.HostParams.Seed)
		}
		seeds[definition.HostParams.Seed] = struct{}{}
	}
	if Inventory[1].HostParams.Wetness <= Inventory[0].HostParams.Wetness || Inventory[2].HostParams.Wetness >= Inventory[0].HostParams.Wetness {
		t.Fatalf("wetness ordering = wet:%d default:%d dry:%d", Inventory[1].HostParams.Wetness, Inventory[0].HostParams.Wetness, Inventory[2].HostParams.Wetness)
	}
}

func TestByNameReturnsDefinitionCopyAndRejectsUnknownName(t *testing.T) {
	definition, ok := ByName("small-wet")
	if !ok || definition.Name != "small-wet" || definition.HostParams.Wetness != 80 {
		t.Fatalf("small-wet lookup = %#v, ok=%v", definition, ok)
	}
	definition.Name = "mutated-copy"
	again, ok := ByName("small-wet")
	if !ok || again.Name != "small-wet" {
		t.Fatalf("lookup was mutated through returned value: %#v, ok=%v", again, ok)
	}
	if missing, ok := ByName("not-a-fixture"); ok || missing.Name != "" {
		t.Fatalf("unknown lookup = %#v, ok=%v", missing, ok)
	}
}
