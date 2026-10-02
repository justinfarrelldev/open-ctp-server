package engine

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestNationFeatBoatMovementRefreshesSeaUnits(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].Effects = []Effect{{ID: "feat:sail", Kind: EffectIncreaseBoatMovement, Value: 100, Turns: 25}}
	state.Units = []Unit{{ID: "ship", OwnerID: "player-a", TileID: "tile-a", Domain: DomainSea, Movement: 0, MaxMovement: 400, Health: 1, MaxHealth: 1}}

	refreshStarterMovement(&state, &rules.RuntimeRules{Starter: rules.RuntimeStarter{Movement: 400}})

	if got := state.Units[0].Movement; got != 500 {
		t.Fatalf("sea unit movement with feat boat bonus = %d, want 500", got)
	}
}

func TestNationFeatBoatMovementDoesNotAffectLandUnits(t *testing.T) {
	state := runtimeStartedState(t)
	state.Nations[0].Effects = []Effect{{ID: "feat:sail", Kind: EffectIncreaseBoatMovement, Value: 100, Turns: 25}}
	state.Units = []Unit{{ID: "infantry", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 0, MaxMovement: 100, Health: 1, MaxHealth: 1}}

	refreshStarterMovement(&state, &rules.RuntimeRules{Starter: rules.RuntimeStarter{Movement: 100}})

	if got := state.Units[0].Movement; got != 100 {
		t.Fatalf("land unit movement with feat boat bonus = %d, want 100", got)
	}
}
