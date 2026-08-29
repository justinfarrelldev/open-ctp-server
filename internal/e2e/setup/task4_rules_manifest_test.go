package setup

import "testing"

func TestTask4ImportedRulesetCarriesMechanicMatrix(t *testing.T) {
	ruleset := importedRuleset(t, 10)
	if err := ruleset.Manifest.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got, want := len(ruleset.Manifest.MechanicMatrix), 10; got != want {
		t.Fatalf("mechanic matrix rows = %d, want %d", got, want)
	}
	if got, want := len(ruleset.Manifest.SyntheticContracts), 10; got != want {
		t.Fatalf("synthetic contracts = %d, want %d", got, want)
	}
}
