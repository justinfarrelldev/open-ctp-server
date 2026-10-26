package ai

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestExplorationUsesReleasedFirstMoveAllowance(t *testing.T) {
	ready := true
	for _, remaining := range []int{0, 100} {
		p := visibility.Projection{Units: []visibility.UnitProjection{{ID: "settler", OwnerID: "ai", TileID: "origin", Domain: engine.DomainLand, Movement: remaining, FirstMove: &ready}}, Tiles: []visibility.TileProjection{{ID: "origin", Planning: &visibility.TilePlanning{Cost: 100, Accessible: true, Domains: []string{"land"}}}, {ID: "forest", X: 1, Planning: &visibility.TilePlanning{Cost: 200, Accessible: true, Domains: []string{"land"}}}}}
		if got := explorationCandidates(p, "ai", nil); len(got) != 1 || got[0].Type != engine.CommandMoveUnit {
			t.Fatalf("remaining=%d: %+v", remaining, got)
		}
	}
}
