package rules

import "testing"

func TestRetailTerritoryPinsSeparateBorderAndPopulationInfluenceRadii(t *testing.T) {
	db, err := LoadRetailRulesFromDirectory(localRetailGameDataRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	runtime := ProjectRetailRuntime(db)
	if runtime.CityTerritory == nil || runtime.CityTerritory.BorderIntRadius != 4 || runtime.CityTerritory.BorderSquaredRadius != 20 {
		t.Fatalf("missing source national radii: %+v", runtime.CityTerritory)
	}
	one := runtime.CityTerritory.SizeForPopulation(1)
	seven := runtime.CityTerritory.SizeForPopulation(7)
	if one.Population != 6 || one.IntRadius != 1 || one.SquaredRadius != 2 || seven.Population != 18 || seven.SquaredRadius != 5 {
		t.Fatalf("wrong population thresholds: %+v / %+v", one, seven)
	}
	defs, hash, err := CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := CompileRuntime(defs, hash)
	if err != nil {
		t.Fatal(err)
	}
	if restored.CityTerritory.Version != CityTerritoryVersion {
		t.Fatal("territory policy not pinned")
	}
	runtime.CityTerritory.BorderSquaredRadius = -1
	if _, _, err := CanonicalRuntimeDefinitions(runtime); err != nil {
		t.Fatal(err)
	} // canonicalization hashes; compilation validates.
	defs, hash, _ = CanonicalRuntimeDefinitions(runtime)
	if _, err := CompileRuntime(defs, hash); err == nil {
		t.Fatal("invalid territory policy compiled")
	}
}
