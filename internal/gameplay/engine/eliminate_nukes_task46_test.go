package engine

import "testing"

func TestRetailEliminateNukesRemovesNuclearPlantsAndQueuedPlants(t *testing.T) {
	state := runtimeStartedState(t)
	state.Cities = []City{
		{ID: "city-a", OwnerID: "player-a", TileID: "tile-a", Buildings: []Building{{ID: "IMPROVE_NUCLEAR_PLANT", Effects: []Effect{{ID: "plant", Kind: EffectNuclearPlant}}}}, Queue: []ProductionItem{{ID: "IMPROVE_NUCLEAR_PLANT_2", Kind: ProductionBuilding, Cost: 1, Effects: []Effect{{ID: "plant", Kind: EffectNuclearPlant}}}, {ID: "IMPROVE_GRANARY", Kind: ProductionBuilding, Cost: 1}}},
		{ID: "city-b", OwnerID: "player-b", TileID: "tile-b", Buildings: []Building{{ID: "IMPROVE_NUCLEAR_PLANT", Effects: []Effect{{ID: "plant", Kind: EffectNuclearPlant}}}}},
	}

	eliminateNukes(&state)
	for _, city := range state.Cities {
		for _, building := range city.Buildings {
			if hasEffect(building.Effects, EffectNuclearPlant) {
				t.Fatalf("nuclear plant building survived in city %s: %#v", city.ID, city.Buildings)
			}
		}
	}
	if len(state.Cities[0].Queue) != 1 || state.Cities[0].Queue[0].ID != "IMPROVE_GRANARY" {
		t.Fatalf("illegal nuclear plant queue items after eliminate nukes = %#v", state.Cities[0].Queue)
	}
}
