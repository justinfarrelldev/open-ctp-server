package engine

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestProductionInspectionMatchesAppendValidation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		allowed bool
		change  func(*State, *rules.RuntimeRules)
	}{
		{"ordinary", true, func(*State, *rules.RuntimeRules) {}},
		{"cant-build", false, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].Effects = []rules.RuntimeStarterEffect{{Kind: EffectCantBuild}}
		}},
		{"special-pop", false, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].Effects = []rules.RuntimeStarterEffect{{Kind: EffectHasPopAndCanBuild}}
		}},
		{"population", false, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].Effects = []rules.RuntimeStarterEffect{{Kind: EffectBuildingRemovesPopulation}}
			s.Cities[0].Population = 1
		}},
		{"advance", false, func(s *State, r *rules.RuntimeRules) { r.Units[0].Prerequisites = []string{"locked"} }},
		{"obsolete", false, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].ObsoleteBy = []string{"known"}
			s.Nations[0].Advances = []string{"known"}
		}},
		{"government", false, func(s *State, r *rules.RuntimeRules) { r.Units[0].RequiredGovernmentID = "other" }},
		{"missing-good", false, func(s *State, r *rules.RuntimeRules) { r.Units[0].NeedsCityGoodID = "coffee" }},
		{"imported-good", true, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].NeedsCityGoodID = "coffee"
			s.Cities = append(s.Cities, City{ID: "foreign", OwnerID: "player-b", TileID: "c"})
			s.TradeRoutes = []TradeRoute{{ID: "import", SourceCityID: "foreign", DestinationCityID: "home", GoodID: "coffee"}}
		}},
		{"park-locked", false, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].Effects = []rules.RuntimeStarterEffect{{Kind: EffectSpecialCreatePark}}
		}},
		{"park-unlocked", true, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].Effects = []rules.RuntimeStarterEffect{{Kind: EffectSpecialCreatePark}}
			s.Wonders = []Wonder{{ID: "enable", OwnerID: "player-b", Effects: []Effect{{Kind: EffectEnableParkRangers}}}}
		}},
		{"global-slavery", false, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].Effects = []rules.RuntimeStarterEffect{{Kind: EffectSpecialSlaveRaid}}
			s.Wonders = []Wonder{{ID: "free", OwnerID: "player-b", Effects: []Effect{{Kind: EffectFreeSlaves}}}}
		}},
		{"nuclear-unit", false, func(s *State, r *rules.RuntimeRules) {
			r.Units[0].Effects = []rules.RuntimeStarterEffect{{Kind: EffectNuclearAttack}}
			s.Wonders = []Wonder{{ID: "ban", OwnerID: "player-b", Effects: []Effect{{Kind: EffectEliminateNukes}}}}
		}},
		{"effective-building", true, func(s *State, r *rules.RuntimeRules) {
			r.Production = []rules.RuntimeProduction{{ID: "candidate", Kind: ProductionBuilding, Cost: 30, RequiredBuildings: []string{"shrine"}}}
			r.Units = nil
			s.Wonders = []Wonder{{ID: "global-building", OwnerID: "player-a", Effects: []Effect{{ID: "shrine", Kind: EffectBuildingEverywhere}}}}
		}},
		{"existing-wonder", false, func(s *State, r *rules.RuntimeRules) {
			r.Units = nil
			r.Production = []rules.RuntimeProduction{{ID: "candidate", Kind: ProductionWonder, Cost: 30}}
			s.Wonders = []Wonder{{ID: "candidate", OwnerID: "player-b"}}
		}},
		{"nuclear-plant", false, func(s *State, r *rules.RuntimeRules) {
			r.Units = nil
			r.Production = []rules.RuntimeProduction{{ID: "candidate", Kind: ProductionBuilding, Cost: 30, Effects: []rules.RuntimeStarterEffect{{Kind: EffectNuclearPlant}}}}
			s.Wonders = []Wonder{{ID: "ban", OwnerID: "player-b", Effects: []Effect{{Kind: EffectEliminateNukes}}}}
		}},
		{"duplicate-building", false, func(s *State, r *rules.RuntimeRules) {
			r.Units = nil
			r.Production = []rules.RuntimeProduction{{ID: "candidate", Kind: ProductionBuilding, Cost: 30}}
			s.Cities[0].Queue = []ProductionItem{{ID: "candidate", Kind: ProductionBuilding, Cost: 30}}
		}},
		{"duplicate-unit", true, func(s *State, r *rules.RuntimeRules) {
			s.Cities[0].Queue = []ProductionItem{{ID: "candidate", Kind: ProductionUnit, Cost: 30}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := task10State()
			state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 3}}
			state.Nations = []Nation{{ID: nationForPlayer(state, "player-a")}}
			runtime := rules.RuntimeRules{CompatibilityVersion: rules.CTP2CompatibilityVersion, Starter: rules.RuntimeStarter{Domain: DomainLand, Movement: 1, Health: 1}, Units: []rules.RuntimeUnitDefinition{{ID: "candidate", Domain: DomainLand, Movement: 1, Health: 1, ProductionCost: 30}}}
			tc.change(&state, &runtime)
			if len(runtime.Units) == 0 {
				runtime.Units = []rules.RuntimeUnitDefinition{{ID: "dummy", Domain: DomainLand, Movement: 1, Health: 1}}
			}
			if len(runtime.Production) == 0 {
				runtime.Production = []rules.RuntimeProduction{{ID: "candidate", Kind: ProductionUnit, Cost: 30}}
			}
			for n := range state.Wonders {
				state.Wonders[n].CityID = "home"
				for e := range state.Wonders[n].Effects {
					if state.Wonders[n].Effects[e].ID == "" {
						state.Wonders[n].Effects[e].ID = "effect"
					}
				}
			}
			raw, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
			if err != nil {
				t.Fatal(err)
			}
			state.Rules = RulesPin{Version: rules.CTP2CompatibilityVersion, Definitions: raw, ContentHash: hash}
			before, err := CanonicalStateJSON(state)
			if err != nil {
				t.Fatal(err)
			}
			inspector, err := NewInspector(state)
			if err != nil {
				t.Fatal(err)
			}
			options := inspector.ProductionOptions("home", "player-a", true)
			found := false
			for _, item := range options.Items {
				if item.ID == "candidate" {
					found = true
				}
			}
			if found != tc.allowed {
				t.Fatalf("inspection allowed=%v expected=%v: %+v", found, tc.allowed, options)
			}
			after, _ := CanonicalStateJSON(state)
			if !bytes.Equal(before, after) {
				t.Fatal("query mutated state")
			}
			err = queueProduction(&state, "player-a", QueueProductionPayload{CityID: "home", Item: ProductionItem{ID: "candidate"}}, inspector.runtime)
			if (err == nil) != tc.allowed {
				t.Fatalf("validation disagrees: %v", err)
			}
		})
	}
}

func TestProductionOptionsCapacityAndInvalidRetainedEntries(t *testing.T) {
	state := task10State()
	runtime, err := rules.ServerOwnedSyntheticRuleset()
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: runtime.CompatibilityVersion, Definitions: runtime.CanonicalJSON, ContentHash: runtime.ContentHash}
	inspector, err := NewInspector(state)
	if err != nil {
		t.Fatal(err)
	}
	id := inspector.runtime.Units[0].ID
	state.Cities = []City{{ID: "home", OwnerID: "player-a", TileID: "a", Population: 3}}
	for n := 0; n < MaxProductionQueueEntries; n++ {
		state.Cities[0].Queue = append(state.Cities[0].Queue, ProductionItem{ID: id, Kind: ProductionUnit, Cost: 1, EntryID: fmt.Sprintf("home:queue:%d", n+1)})
	}
	inspector, _ = NewInspector(state)
	options := inspector.ProductionOptions("home", "player-a", true)
	if options.Status != "complete" || len(options.Items) != 0 {
		t.Fatalf("full queue: %+v", options)
	}
	if err := queueProduction(&state, "player-a", QueueProductionPayload{CityID: "home", Item: ProductionItem{ID: id}}, inspector.runtime); err == nil {
		t.Fatal("257th entry accepted")
	}
	state.Cities[0].Queue = state.Cities[0].Queue[:1]
	state.Cities[0].Queue[0].ID = "no-longer-valid"
	inspector, _ = NewInspector(state)
	options = inspector.ProductionOptions("home", "player-a", true)
	if options.Status != "unavailable" {
		t.Fatal("invalid retained entry advertised appends")
	}
	if inspector.ProductionOptions("home", "player-b", true) != nil {
		t.Fatal("foreign inspection")
	}
}
