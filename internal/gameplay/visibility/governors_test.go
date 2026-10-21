package visibility

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestGovernorProjectionAdvertisesPinnedPrioritiesAndKeepsSettingsPrivate(t *testing.T) {
	state := inspectionState(t)
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.AI = &rules.RuntimeAI{BuildingBuildLists: []rules.RuntimeAIList{{ID: "BUILDING_BUILD_LIST_SCIENCE", Items: []string{"LIBRARY"}}}}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = engine.RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: definitions}
	state.Cities = []engine.City{{ID: "home", OwnerID: "a", TileID: "t0", Population: 2, Governor: &engine.GovernorState{Enabled: true, Priority: "science"}}}
	own, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if own.Contract.Features["governor"] != true {
		t.Fatal("governor capability absent")
	}
	wire, _ := json.Marshal(own)
	if !bytes.Contains(wire, []byte(`"governor":{"enabled":true,"priority":"science"}`)) {
		t.Fatal("governor reconnect state absent")
	}
	enable, disable := false, false
	for _, action := range own.LegalActions {
		if action.Command == engine.CommandSetGovernor {
			var payload struct{ Enabled bool }
			_ = json.Unmarshal(action.Payload, &payload)
			if payload.Enabled {
				enable = true
			} else {
				disable = true
			}
		}
	}
	if !enable || !disable {
		t.Fatal("governor legal actions missing")
	}
	foreign, err := Project(state, "b")
	if err != nil {
		t.Fatal(err)
	}
	wire, _ = json.Marshal(foreign)
	if bytes.Contains(wire, []byte(`"governor":{`)) {
		t.Fatal("foreign city exposed governor settings")
	}
}
