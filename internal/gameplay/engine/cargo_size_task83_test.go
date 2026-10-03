package engine

import "testing"

func TestRetailCargoRequiresMatchingSizeClass(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "carrier", OwnerID: "player-a", TileID: "tile-a", Domain: DomainSea, CargoCapacity: 1, CargoDomains: []string{DomainLand}, CargoClasses: []string{"SmallLand"}},
		{ID: "tank", OwnerID: "player-a", TileID: "tile-a", Domain: DomainLand, Size: "Large"},
	}

	err := embark(&state, "player-a", embarkPayload{UnitID: "tank", TransportID: "carrier"})
	if err == nil {
		t.Fatalf("large land unit embarked on SmallLand-only carrier: %+v", state.Units)
	}
}

func TestRetailCargoAllowsAmphibiousUnitsToMatchLandOrWaterClass(t *testing.T) {
	state := fixtureState()
	state.Tiles = []Tile{{ID: "tile-a", Terrain: TerrainPlains, Open: true}}
	state.Units = []Unit{
		{ID: "carrier", OwnerID: "player-a", TileID: "tile-a", Domain: DomainSea, CargoCapacity: 1, CargoDomains: []string{DomainLand}, CargoClasses: []string{"SmallLand"}},
		{ID: "marine", OwnerID: "player-a", TileID: "tile-a", Domain: DomainAmphibious, Size: "Small"},
	}

	if err := embark(&state, "player-a", embarkPayload{UnitID: "marine", TransportID: "carrier"}); err != nil {
		t.Fatalf("amphibious small unit should match retail SmallLand carry class: %v", err)
	}
}
