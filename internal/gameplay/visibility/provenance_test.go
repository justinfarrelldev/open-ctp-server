package visibility

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"testing"
)

func TestRulesProvenanceIsPinnedNotGuessedFromCompatibility(t *testing.T) {
	synthetic, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	pin := engine.RulesPin{Version: synthetic.CompatibilityVersion, Definitions: synthetic.CanonicalJSON, ContentHash: synthetic.ContentHash}
	p := rulesProjection(engine.State{Rules: pin})
	wire := metadataWire(t, p)
	if wire["family"] != "synthetic" {
		t.Fatalf("ambiguous provenance: %v", wire["family"])
	}
	legacy := productionOptionsState(t)
	if got := metadataWire(t, rulesProjection(legacy))["family"]; got != "unknown" {
		t.Fatalf("guessed legacy family: %v", got)
	}
	retail := rules.ProjectRetailRuntime(rules.RetailRules{})
	if got := metadataWire(t, retail)["family"]; got != "retail" {
		t.Fatalf("retail pin not labelled: %v", got)
	}
	if metadataWire(t, p)["content_hash"] != pin.ContentHash {
		t.Fatal("lost content identity")
	}
}
