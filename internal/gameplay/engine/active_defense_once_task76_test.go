package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailActiveDefenseFiresOncePerTurn(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "middle", X: 1, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "target", X: 2, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "bomber-a", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}},
		{ID: "bomber-b", OwnerID: "player-a", TileID: "origin", Domain: DomainLand, Movement: 1, Ranged: 100, BombardRange: 2, BombRounds: 1, Strength: 1, Armor: 1, Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "bombard-land", Kind: EffectCanBombardLand}}},
		{ID: "interceptor", OwnerID: "player-b", TileID: "target", Domain: DomainLand, Defense: 0, Ranged: 100, ActiveDefenseRange: 2, Strength: 3, Armor: 1, Health: 20, MaxHealth: 20, Effects: []Effect{{ID: "defend-land", Kind: EffectDefendLand}}},
	}

	for _, unitID := range []string{"bomber-a", "bomber-b"} {
		payload, _ := json.Marshal(bombardPayload{UnitID: unitID, TargetTileID: "target"})
		if _, _, err := applyBombard(&state, Command{Type: CommandBombard, ActorPlayerID: "player-a", Payload: payload}, fixtureSources(), nil); err != nil {
			t.Fatalf("bombard %s err=%v", unitID, err)
		}
	}

	first, _ := unitByID(state, "bomber-a")
	second, _ := unitByID(state, "bomber-b")
	if first.Health != 7 || second.Health != 10 {
		t.Fatalf("active defense should fire once per turn; first=%d second=%d", first.Health, second.Health)
	}
}

func TestRetailActiveDefenseUseResetsOnTurnRefresh(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "origin", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{{ID: "interceptor", OwnerID: "player-b", TileID: "origin", Domain: DomainLand, Movement: 0, MaxMovement: 1, Health: 10, MaxHealth: 10, UsedActiveDefense: true}}

	advanceRound(&state, &rules.RuntimeRules{Starter: rules.RuntimeStarter{Movement: 1}})

	unit, _ := unitByID(state, "interceptor")
	if unit.UsedActiveDefense {
		t.Fatalf("active-defense use flag did not reset on turn refresh: %+v", unit)
	}
}
