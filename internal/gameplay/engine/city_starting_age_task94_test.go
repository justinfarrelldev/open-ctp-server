package engine

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestRetailStartingAgeSeedsEarlierAgeCityBuildings(t *testing.T) {
	runtime := rules.RuntimeRules{
		CompatibilityVersion: rules.CTP2CompatibilityVersion,
		Starter:              rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1},
		Units:                []rules.RuntimeUnitDefinition{{ID: "UNIT_SETTLER", Domain: DomainLand, Movement: 1, Health: 1, Effects: []rules.RuntimeStarterEffect{{ID: "settler", Kind: EffectSettler}}}},
		Ages:                 []rules.RuntimeAgeDefinition{{ID: "AGE_ONE", Index: 0}, {ID: "AGE_TWO", Index: 1}},
		StartAgeID:           "AGE_TWO",
		Advances:             []rules.RuntimeAdvanceDefinition{{ID: "ADVANCE_OLD", Cost: 1, AgeID: "AGE_ONE"}, {ID: "ADVANCE_CURRENT", Cost: 1, AgeID: "AGE_TWO"}},
		Production: []rules.RuntimeProduction{
			{ID: "IMPROVE_CAPITOL", Kind: ProductionBuilding, Cost: 1, Prerequisites: []string{"ADVANCE_OLD"}, Effects: []rules.RuntimeStarterEffect{{ID: "capitol", Kind: EffectCapitol}}},
			{ID: "IMPROVE_GRANARY", Kind: ProductionBuilding, Cost: 1, Upkeep: 2, Prerequisites: []string{"ADVANCE_OLD"}, Effects: []rules.RuntimeStarterEffect{{ID: "food", Kind: EffectYieldFoodPercent, Value: 10}}},
			{ID: "IMPROVE_MODERN", Kind: ProductionBuilding, Cost: 1, Prerequisites: []string{"ADVANCE_CURRENT"}},
			{ID: "UNIT_SETTLER", Kind: ProductionUnit, Cost: 1},
		},
	}
	definitions, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := State{
		SchemaVersion: StateSchemaV1,
		Rules:         RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: definitions},
		Phase:         PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", ActivePlayerOrder: []string{"player-a"}},
		Players:       []Player{{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman}},
		Nations:       []Nation{{ID: "nation-a", CurrentAgeID: "AGE_TWO", Advances: []string{"ADVANCE_OLD"}}},
		Tiles:         []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true, OwnerID: "player-a"}},
		Units:         []Unit{{ID: "settler-a", OwnerID: "player-a", TypeID: "UNIT_SETTLER", TileID: "tile-a", Health: 10, MaxHealth: 10, Effects: []Effect{{ID: "settler", Kind: EffectSettler}}}},
		RNG:           []RNGDomainState{{Domain: RNGDomainCombat}, {Domain: RNGDomainSpecialAction}, {Domain: RNGDomainAI}, {Domain: RNGDomainRiot}},
	}
	payload, _ := json.Marshal(FoundCityPayload{UnitID: "settler-a", CityID: "city-a", Name: "Age Seed"})

	next, _, _, err := Apply(state, Command{SchemaVersion: CommandSchemaV1, ID: "found", ActorPlayerID: "player-a", Type: CommandFoundCity, Payload: payload}, state.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	city, ok := cityByID(next, "city-a")
	if !ok {
		t.Fatal("founded city missing")
	}
	if !cityHasBuilding(*city, "IMPROVE_GRANARY") || cityHasBuilding(*city, "IMPROVE_CAPITOL") || cityHasBuilding(*city, "IMPROVE_MODERN") {
		t.Fatalf("city buildings = %+v, want old non-capitol building only", city.Buildings)
	}
}
