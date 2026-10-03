package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailTelevangelistConversionNeedsDoubleTelevangelistsBuilding(t *testing.T) {
	zeroScience := 0
	state := State{
		Nations: []Nation{{ID: "nation-a", SciencePercent: &zeroScience}, {ID: "nation-b", SciencePercent: &zeroScience}},
		Players: []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}, {ID: "player-b", NationID: "nation-b", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Cities:  []City{{ID: "converted", OwnerID: "player-b", TileID: "tile-b", ConvertedToPlayerID: "player-a", ConvertedBy: "televangelist", Yields: Yield{Gold: 10}}},
	}
	runtime := &rules.RuntimeRules{Constants: rules.RuntimeConstants{ClericConversionPercent: 20, TelevangelistConversionPercent: 40}}

	advanceTreasury(&state, runtime)
	if state.Nations[0].Gold != 2 || state.Nations[1].Gold != 8 {
		t.Fatalf("televangelist conversion without double building = %+v, want cleric-rate siphon", state.Nations)
	}

	state.Nations[0].Gold, state.Nations[1].Gold = 0, 0
	state.Cities[0].Buildings = []Building{{ID: "IMPROVE_TELEVISION", Effects: []Effect{{ID: "double", Kind: EffectDoubleTelevangelists}}}}
	advanceTreasury(&state, runtime)
	if state.Nations[0].Gold != 4 || state.Nations[1].Gold != 6 {
		t.Fatalf("televangelist conversion with double building = %+v, want televangelist-rate siphon", state.Nations)
	}
}
