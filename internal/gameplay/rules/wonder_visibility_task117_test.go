package rules

import "testing"

func TestRetailActualBuildingEverywhereProjectsEffectiveBuilding(t *testing.T) {
	effects := retailWonderRuntimeEffects(RetailWonder{ID: "WONDER_TEST", RawFields: []RetailRawField{{Key: "ActualBuildingEverywhere", Values: []string{"IMPROVE_TEMPLE"}}}}, nil)
	for _, effect := range effects {
		if effect.Kind == "building_everywhere" && effect.ID == "IMPROVE_TEMPLE" {
			return
		}
	}
	t.Fatalf("effects = %#v, want ActualBuildingEverywhere as building_everywhere", effects)
}
