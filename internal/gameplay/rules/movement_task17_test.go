package rules

import "testing"

func TestTask17RuntimeMovementModelIsClosedAndSorted(t *testing.T) {
	ruleset, err := ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := CompileRuntime(ruleset.CanonicalJSON, ruleset.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	if !runtime.Movement.OpenAccess || !runtime.Movement.ZOCStopsPath {
		t.Fatalf("synthetic movement policy must expose open access and path-stopping ZOC: %#v", runtime.Movement)
	}
	if got, want := runtime.Movement.Terrains[0].ID, "forest"; got != want {
		t.Fatalf("terrain iteration must be stable: got %q want %q", got, want)
	}
	forest, ok := runtime.Movement.Terrain("forest")
	if !ok || forest.Cost != 2 || !forest.Allows("land") || forest.Allows("sea") {
		t.Fatalf("forest movement definition is invalid: %#v", forest)
	}
	ocean, ok := runtime.Movement.Terrain("ocean")
	if !ok || ocean.Cost != 1 || !ocean.Allows("sea") || ocean.Allows("land") {
		t.Fatalf("ocean movement definition is invalid: %#v", ocean)
	}
}
