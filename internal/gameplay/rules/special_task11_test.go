package rules

import "testing"

func TestTask11ManifestSpecialActionsRequireExecutableVectors(t *testing.T) {
	manifest := SupportedManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	manifest.SpecialActionVectors = manifest.SpecialActionVectors[:3]
	if err := manifest.Validate(); err == nil {
		t.Fatal("missing special-action vector was accepted")
	}
	manifest = SupportedManifest()
	manifest.SpecialActionEvaluators[0].Source.Symbol = ""
	if err := manifest.Validate(); err == nil {
		t.Fatal("unproven special-action evaluator was accepted")
	}
}
