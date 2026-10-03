package visibility

import (
	"bytes"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestGiftUnitLegalActionUsesKnownDiplomacyTargets(t *testing.T) {
	projection := Projection{
		Units:     []UnitProjection{{ID: "elephant", OwnerID: "a", TileID: "0,0", Effects: []engine.Effect{{ID: "gift", Kind: engine.EffectCanBeGifted}}}},
		Diplomacy: &DiplomacyProjection{Relations: []DiplomacyRelationProjection{{TargetPlayerID: "b"}, {TargetPlayerID: "a"}}},
	}

	actions := unitActions(projection, "a")
	for _, action := range actions {
		if action.Command == engine.CommandGiftUnit && action.EntityID == "elephant" && action.TargetID == "b" && bytes.Contains(action.Payload, []byte(`"target_player_id":"b"`)) {
			return
		}
	}
	t.Fatalf("gift action missing from %#v", actions)
}

func TestGiftUnitLegalActionRequiresCanBeGifted(t *testing.T) {
	projection := Projection{
		Units:     []UnitProjection{{ID: "warrior", OwnerID: "a", TileID: "0,0"}},
		Diplomacy: &DiplomacyProjection{Relations: []DiplomacyRelationProjection{{TargetPlayerID: "b"}}},
	}

	for _, action := range unitActions(projection, "a") {
		if action.Command == engine.CommandGiftUnit {
			t.Fatalf("gift action exposed without CanBeGifted: %#v", action)
		}
	}
}
