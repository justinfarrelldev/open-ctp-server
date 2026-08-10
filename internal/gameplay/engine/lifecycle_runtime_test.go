package engine

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRuntimeRulesNaturalVictoryTransitions(t *testing.T) {
	tests := []struct {
		name  string
		run   func(t *testing.T, state State) State
		want         string
		startingTurn int
	}{
		{
			name: "conquest follows combat elimination",
			want: "bloodlust_conquest",
			run: func(t *testing.T, state State) State {
				state.Units = []Unit{
					{ID: "attacker", OwnerID: "player-a", TileID: "tile-a", ArmyID: "army-a", Domain: DomainLand, Movement: 2, Attack: 100, Defense: 100, Strength: 100, Armor: 100, Health: 10, MaxHealth: 10},
					{ID: "defender", OwnerID: "player-b", TileID: "tile-b", ArmyID: "army-b", Domain: DomainLand, Movement: 2, Attack: 1, Defense: 1, Strength: 1, Armor: 1, Health: 1, MaxHealth: 1},
				}
				state.Armies = []Army{{ID: "army-a", OwnerID: "player-a", UnitIDs: []string{"attacker"}}, {ID: "army-b", OwnerID: "player-b", UnitIDs: []string{"defender"}}}
				state.Fog = []FogObservation{{PlayerID: "player-a", TileID: "tile-b"}}
				state = runtimeApply(t, state, "player-a", CommandCombat, combatPayload{AttackerArmyID: "army-a", DefenderArmyID: "army-b"})
				return runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
			},
		},
		{
			name: "world union follows reciprocal embassies",
			want: "diplomatic_world_union",
			run: func(t *testing.T, state State) State {
				state.Units = []Unit{
					{ID: "envoy-a", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "embassy", Kind: EffectSpecialEmbassy}}},
					{ID: "envoy-b", OwnerID: "player-b", TileID: "tile-b", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "embassy", Kind: EffectSpecialEmbassy}}},
				}
				state.Cities = []City{{ID: "city-a", Name: "A", OwnerID: "player-a", TileID: "tile-a"}, {ID: "city-b", Name: "B", OwnerID: "player-b", TileID: "tile-b"}}
				state = runtimeApply(t, state, "player-a", CommandEstablishEmbassy, SpecialActionPayload{UnitID: "envoy-a", CityID: "city-b"})
				state = runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
				state = runtimeApply(t, state, "player-b", CommandEstablishEmbassy, SpecialActionPayload{UnitID: "envoy-b", CityID: "city-a"})
				return runtimeApply(t, state, "player-b", CommandRoundComplete, struct{}{})
			},
		},
		{
			name:         "calendar score follows founded cities",
			want:         "score_2300",
			startingTurn: 46,
			run: func(t *testing.T, state State) State {
				state.Units = []Unit{
					{ID: "settler-a", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "settler", Kind: EffectSettler}}},
					{ID: "settler-b", OwnerID: "player-b", TileID: "tile-b", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "settler", Kind: EffectSettler}}},
				}
				state = runtimeApply(t, state, "player-a", CommandFoundCity, FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "A"})
				state = runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
				state = runtimeApply(t, state, "player-b", CommandFoundCity, FoundCityPayload{UnitID: "settler-b", CityID: "city-b", Name: "B"})
				return runtimeApply(t, state, "player-b", CommandRoundComplete, struct{}{})
			},
		},
		{
			name: "Gaia follows science theft",
			want: "science_gaia_controller",
			run: func(t *testing.T, state State) State {
				state.Units = []Unit{{ID: "scientist", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Movement: 2, Effects: []Effect{{ID: "steal", Kind: EffectSpecialSteal, Value: 100}}}}
				state.Cities = []City{{ID: "city-b", Name: "B", OwnerID: "player-b", TileID: "tile-b"}}
				state.Nations[1].Advances = []string{"gaia-science"}
				state = runtimeApply(t, state, "player-a", CommandStealTechnology, SpecialActionPayload{UnitID: "scientist", CityID: "city-b"})
				return runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startingTurn := tt.startingTurn
			if startingTurn == 0 {
				startingTurn = 1
			}
			next := tt.run(t, runtimeStartedStateAt(t, startingTurn))
			if !next.Terminal.Ended || next.Terminal.Reason != tt.want || next.Victory.Reason != tt.want {
				t.Fatalf("terminal victory = %#v / %#v, want %q", next.Terminal, next.Victory, tt.want)
			}
		})
	}
}

func runtimeStartedState(t *testing.T) State {
	return runtimeStartedStateAt(t, 1)
}

func runtimeStartedStateAt(t *testing.T, turn int) State {
	t.Helper()
	ruleset, err := rules.Import(rules.ImportConfig{Directory: filepath.Join("..", "rules", "testdata", "synthetic-ctp2")})
	if err != nil {
		t.Fatal(err)
	}
	return State{
		SchemaVersion: StateSchemaV1,
		Rules: RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON},
		Phase: PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: turn, ActivePlayerOrder: []string{"player-a", "player-b"}},
		Tiles: []Tile{{ID: "tile-a", X: 0, Y: 0, Terrain: TerrainPlains, Open: true}, {ID: "tile-b", X: 1, Y: 0, Terrain: TerrainForest, Open: true}},
		Players: []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}, {ID: "player-b", NationID: "nation-b", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Nations: []Nation{{ID: "nation-a"}, {ID: "nation-b"}},
		RNG: []RNGDomainState{{Domain: RNGDomainCombat, State: 1}, {Domain: RNGDomainSpecialAction, State: 2}, {Domain: RNGDomainAI, State: 3}},
	}
}

func runtimeApply(t *testing.T, state State, actor, commandType string, payload any) State {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: commandType + "-" + actor, ActorPlayerID: actor, Type: commandType, Payload: encoded}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatalf("Apply(%s): %v", commandType, err)
	}
	return next
}
