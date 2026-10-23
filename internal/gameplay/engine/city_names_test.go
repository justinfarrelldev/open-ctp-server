package engine

import (
	"strings"
	"testing"
)

func TestFoundCityDefaultDoesNotUseTechnicalIdentitiesOrForeignNames(t *testing.T) {
	for _, foreign := range []string{"City 1", "Secret foreign name"} {
		state := task10State()
		state.Cities = []City{{ID: "foreign", Name: foreign, OwnerID: "player-b", TileID: "c"}}
		if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "technical-unit-settler-city"}); err != nil {
			t.Fatal(err)
		}
		city, _ := cityByID(state, "technical-unit-settler-city")
		if city.Name != "City 1" {
			t.Fatalf("technical/foreign-dependent default: %q", city.Name)
		}
	}
}

func TestFoundCityNamesAreValidatedNotGlobalIdentities(t *testing.T) {
	state := task10State()
	state.Cities = []City{{ID: "foreign", Name: "Memphis", OwnerID: "player-b", TileID: "c"}}
	if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home", Name: " Memphis "}); err != nil {
		t.Fatal(err)
	}
	city, _ := cityByID(state, "home")
	if city.Name != "Memphis" {
		t.Fatalf("name not normalized: %q", city.Name)
	}
	for _, name := range []string{"in\nvalid", strings.Repeat("a", 129)} {
		state = task10State()
		if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home", Name: name}); err == nil {
			t.Fatalf("invalid name accepted %q", name)
		}
		if len(state.Units) != 2 || len(state.Cities) != 0 {
			t.Fatal("invalid name consumed settler")
		}
	}
}
