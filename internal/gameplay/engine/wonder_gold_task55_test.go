package engine

import "testing"

func TestRetailBonusGoldWonderAddsTreasuryOncePerOwner(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}},
		Nations: []Nation{{ID: "nation-a"}},
		Cities: []City{
			{ID: "city-a", OwnerID: "player-a", Yields: Yield{Gold: 0}},
			{ID: "city-b", OwnerID: "player-a", Yields: Yield{Gold: 0}},
		},
		Wonders: []Wonder{{ID: "WONDER_PYRAMIDS", OwnerID: "player-a", Effects: []Effect{{ID: "bonus", Kind: EffectBonusGold, Value: 100}}}},
	}

	advanceTreasury(&state)

	if got := state.Nations[0].Gold; got != 100 {
		t.Fatalf("bonus-gold treasury = %d, want one owner-level 100 gold payment", got)
	}
}

func TestRetailTradeRouteWonderGoldIsOwnerTreasuryForAllRoutes(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}, {ID: "player-c", NationID: "nation-c"}},
		Nations: []Nation{{ID: "nation-a"}, {ID: "nation-b"}, {ID: "nation-c"}},
		Cities: []City{
			{ID: "city-a", OwnerID: "player-a"},
			{ID: "city-b", OwnerID: "player-b"},
			{ID: "city-c", OwnerID: "player-c"},
		},
		Wonders:     []Wonder{{ID: "WONDER_EAST_INDIA_COMPANY", OwnerID: "player-a", Effects: []Effect{{ID: "water", Kind: EffectGoldPerWaterTradeRoute, Value: 15}, {ID: "international", Kind: EffectGoldPerInternationalTradeRoute, Value: 25}}}},
		TradeRoutes: []TradeRoute{{ID: "foreign-route", SourceCityID: "city-b", DestinationCityID: "city-c", CrossesWater: true}},
	}

	advanceTreasury(&state)

	if got := state.Nations[0].Gold; got != 40 {
		t.Fatalf("trade-route wonder treasury = %d, want water + international gold from global route", got)
	}
}

func TestRetailWonderGoldPerTelevisionCountsForeignTelevisionPopulation(t *testing.T) {
	state := State{
		Players: []Player{{ID: "player-a", NationID: "nation-a"}, {ID: "player-b", NationID: "nation-b"}, {ID: "player-c", NationID: "nation-c"}},
		Nations: []Nation{{ID: "nation-a"}, {ID: "nation-b"}, {ID: "nation-c"}},
		Cities: []City{
			{ID: "own-tv", OwnerID: "player-a", Population: 10, Buildings: []Building{{ID: "IMPROVE_TELEVISION", Effects: []Effect{{ID: "tv", Kind: EffectTelevision}}}}},
			{ID: "foreign-tv", OwnerID: "player-b", Population: 6, Buildings: []Building{{ID: "IMPROVE_TELEVISION", Effects: []Effect{{ID: "tv", Kind: EffectTelevision}}}}},
			{ID: "foreign-no-tv", OwnerID: "player-c", Population: 8},
		},
		Wonders: []Wonder{{ID: "WONDER_MEDIA_INSTALLATION", OwnerID: "player-a", Effects: []Effect{{ID: "tv-gold", Kind: EffectGoldPerTelevision, Value: 7}}}},
	}

	advanceTreasury(&state)

	if got := state.Nations[0].Gold; got != 42 {
		t.Fatalf("television wonder treasury = %d, want 7 * foreign television population 6", got)
	}
}
