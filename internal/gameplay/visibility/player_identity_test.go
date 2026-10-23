package visibility

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"testing"
)

func TestViewerIdentityAndPersistentColorSlotsDoNotDependOnUnits(t *testing.T) {
	state := fixtureState()
	state.Units = nil
	state.Nations = []engine.Nation{{ID: "a", CivilizationID: "ROMAN"}, {ID: "b", CivilizationID: "GREEK"}}
	zero, one := 0, 1
	state.Players[0].ColorIndex = &zero
	state.Players[1].ColorIndex = &one
	p, err := Project(state, "a")
	if err != nil {
		t.Fatal(err)
	}
	if p.ViewerPlayerID != "a" || len(p.Players) != 2 {
		t.Fatalf("missing authoritative identity: %+v", p)
	}
	for _, player := range p.Players {
		if player.ID == "a" && (player.ColorIndex == nil || *player.ColorIndex != 0 || player.CivilizationID != "ROMAN") {
			t.Fatalf("lost identity after last settler: %+v", player)
		}
	}
	raw, err := engine.CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := engine.DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	p, err = Project(restored, "a")
	if err != nil {
		t.Fatal(err)
	}
	if p.Players[1].ColorIndex == nil || *p.Players[1].ColorIndex != 1 {
		t.Fatal("restart changed color slot")
	}
}
