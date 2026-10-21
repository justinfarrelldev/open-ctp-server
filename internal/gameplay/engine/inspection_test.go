package engine

import (
	"bytes"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestProductionForecastUsesEffectiveEconomyAndPublicWorksWithoutMutation(t *testing.T) {
	state := task10State()
	works := 50
	state.Nations = []Nation{{ID: "nation-a", Gold: 1000, PublicWorksPercent: &works}}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2, StoredProduction: 4, Queue: []ProductionItem{{ID: "warrior", Kind: ProductionUnit, Cost: 10}}}}
	before, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	inspector, err := NewInspector(state)
	if err != nil {
		t.Fatal(err)
	}
	forecast := inspector.Production("home", "player-a")
	yield, err := cityYieldWithRuntime(state, state.Cities[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	rate := yield.Production - yield.Production*works/100
	if forecast == nil || forecast.Rate == nil || *forecast.Rate != rate || forecast.Turns == nil || *forecast.Turns != (6+rate-1)/rate || forecast.Status != "estimated" {
		t.Fatalf("forecast=%+v rate=%d", forecast, rate)
	}
	if inspector.Production("home", "player-b") != nil {
		t.Fatal("foreign economy quote disclosed")
	}
	after, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("inspection mutated authority")
	}
	state.Cities[0].Effects = []Effect{{ID: "injunction", Kind: EffectInjunction, Turns: 2}}
	inspector, err = NewInspector(state)
	if err != nil {
		t.Fatal(err)
	}
	forecast = inspector.Production("home", "player-a")
	if forecast.Status != "stalled" || forecast.Turns != nil || forecast.Rate == nil || *forecast.Rate != 0 {
		t.Fatalf("stalled forecast invented completion: %+v", forecast)
	}
}

func TestResearchForecastRecomputesOwnedRatesWithoutMutation(t *testing.T) {
	state := task10State()
	ruleset, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: ruleset.CompatibilityVersion, ContentHash: ruleset.ContentHash, Definitions: ruleset.CanonicalJSON}
	state.Nations = []Nation{{ID: "nation-a", Advances: []string{"ADVANCE_SYNTHETIC_ROOT"}, ResearchTargetID: "ADVANCE_SYNTHETIC_UNLOCK", ResearchProgress: 1}}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2, Yields: Yield{Gold: 999}}}
	before, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	inspector, err := NewInspector(state)
	if err != nil {
		t.Fatal(err)
	}
	forecast := inspector.Research("player-a")
	evaluated := state
	evaluated.Cities = append([]City(nil), state.Cities...)
	evaluated.Cities[0].Yields, err = cityYieldWithRuntime(evaluated, evaluated.Cities[0], inspector.runtime)
	if err != nil {
		t.Fatal(err)
	}
	rate := researchRate(evaluated, state.Nations[0])
	if forecast == nil || forecast.Rate == nil || *forecast.Rate != rate || forecast.Turns == nil {
		t.Fatalf("research forecast=%+v want rate=%d", forecast, rate)
	}
	after, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("research inspection mutated authority")
	}
}

func TestRushQuoteMatchesExecutionAndDistinguishesStatuses(t *testing.T) {
	state := task10State()
	state.Nations = []Nation{{ID: "nation-a", Gold: 1000}}
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 2, StoredProduction: 4, Queue: []ProductionItem{{ID: "warrior", Kind: ProductionUnit, Cost: 10}}}}
	inspector, err := NewInspector(state)
	if err != nil {
		t.Fatal(err)
	}
	quote := inspector.RushBuy("home", "player-a")
	if quote == nil || quote.Cost == nil || quote.Status != "available" {
		t.Fatalf("quote=%+v", quote)
	}
	next := task10Apply(t, state, "player-a", CommandRushBuy, RushBuyPayload{CityID: "home"})
	if state.Nations[0].Gold-next.Nations[0].Gold != *quote.Cost {
		t.Fatal("display quote and execution diverged")
	}
	for _, tc := range []struct {
		name         string
		gold, stored int
		kind         string
		status       string
		hasCost      bool
	}{
		{"poor", 0, 4, ProductionUnit, "insufficient-gold", true},
		{"complete", 1000, 10, ProductionUnit, "completed", true},
		{"continuous", 1000, 4, ProductionCapitalization, "not-rushable", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state.Nations[0].Gold = tc.gold
			state.Cities[0].StoredProduction = tc.stored
			state.Cities[0].Queue[0].Kind = tc.kind
			inspector, err := NewInspector(state)
			if err != nil {
				t.Fatal(err)
			}
			quote := inspector.RushBuy("home", "player-a")
			if quote.Status != tc.status || (quote.Cost != nil) != tc.hasCost {
				t.Fatalf("quote=%+v", quote)
			}
		})
	}
}
