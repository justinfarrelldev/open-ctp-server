package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailDestroyedEliminateNukesWonderReversesProductionBlock(t *testing.T) {
	state := fixtureState()
	state.Nations = []Nation{{ID: "nation-a", Effects: []Effect{{ID: "legacy-eliminate", Kind: EffectEliminateNukes}}}, {ID: "nation-b"}}
	state.Wonders = []Wonder{{ID: "WONDER_NANITE_DEFUSER", OwnerID: "player-b", Destroyed: true, Effects: []Effect{{ID: "defuse", Kind: EffectEliminateNukes}}}}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Population: 2}}
	runtime := &rules.RuntimeRules{Units: []rules.RuntimeUnitDefinition{{ID: "UNIT_NUKE", Effects: []rules.RuntimeStarterEffect{{ID: "nuke", Kind: EffectNuclearAttack}}}}, Production: []rules.RuntimeProduction{{ID: "UNIT_NUKE", Kind: ProductionUnit, Cost: 1}}}
	payload := QueueProductionPayload{CityID: "city-a", Item: ProductionItem{ID: "UNIT_NUKE", Kind: ProductionUnit, Cost: 1}}

	if err := queueProduction(&state, "player-a", payload, runtime); err != nil {
		t.Fatalf("destroyed eliminate-nukes wonder should not block production: %v", err)
	}
}
