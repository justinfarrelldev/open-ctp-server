package rules

import "testing"

func TestRetailRuntimeProjectionCanBePinnedAndCompiled(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	definitions, contentHash, err := CanonicalRuntimeDefinitions(ProjectRetailRuntime(db))
	if err != nil {
		t.Fatalf("CanonicalRuntimeDefinitions() error = %v", err)
	}
	runtime, err := CompileRuntime(definitions, contentHash)
	if err != nil {
		t.Fatalf("CompileRuntime(retail projection) error = %v", err)
	}
	if runtime.ContentHash != contentHash {
		t.Fatalf("compiled hash = %q, want %q", runtime.ContentHash, contentHash)
	}
	if got, want := len(runtime.Units), 74; got != want {
		t.Fatalf("len(runtime.Units) = %d, want %d", got, want)
	}
	settler, ok := runtime.Unit("UNIT_SETTLER")
	if !ok || runtime.Starter.Domain != settler.Domain || runtime.Starter.Movement != settler.Movement {
		t.Fatalf("runtime starter = %+v settler=%+v ok=%v", runtime.Starter, settler, ok)
	}
}
