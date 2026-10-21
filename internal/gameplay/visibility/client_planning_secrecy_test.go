package visibility

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"testing"
)

func TestClientPlanningDoesNotDiscloseUndetectedEnemyZOC(t *testing.T) {
	state := inspectionState(t)
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Movement.ZOCStopsPath = true
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules.ContentHash = hash
	state.Rules.Definitions = definitions
	state.Units = append(state.Units, engine.Unit{ID: "undetected", OwnerID: "b", TileID: "t1", ZOC: true, Effects: []engine.Effect{{ID: "stealth", Kind: engine.EffectStealth}}})
	projection, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if projectionHasUnit(projection, "undetected") {
		t.Fatal("stealth fixture must not be detected")
	}
	tile := metadataTile(t, projection, "t0")
	if tile.Planning == nil || tile.Planning.EnemyZOC {
		t.Fatalf("preview disclosed undetected enemy ZOC: %+v", tile.Planning)
	}
}
