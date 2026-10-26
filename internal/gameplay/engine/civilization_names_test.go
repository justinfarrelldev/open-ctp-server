package engine

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"testing"
)

func civilizationNameState(t *testing.T) State {
	t.Helper()
	state := task10State()
	state.Rules = runtimeStartedStateAt(t, 1).Rules
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	runtime.CityNaming = &rules.RuntimeCityNaming{Version: rules.CityNamingVersion, Language: "english", Civilizations: []rules.RuntimeCivilization{{ID: "ROMAN", Aliases: []string{"roman", "romans"}, CityNames: []string{"Rome", "Antium"}}, {ID: "GREEK", Aliases: []string{"greek", "greeks"}, CityNames: []string{"Athens", "Sparta"}}}}
	defs, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Rules = RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: defs}
	state.Nations = []Nation{{ID: "nation-a", CivilizationID: "ROMAN"}, {ID: "nation-b", CivilizationID: "GREEK"}}
	return state
}

func TestDefaultFoundingUsesPinnedCivilizationNamesAndOnlyOwnCollisions(t *testing.T) {
	for _, foreign := range []string{"Rome", "Rome1", "Secret"} {
		state := civilizationNameState(t)
		// Naming collisions must not depend on a foreign city, but the new
		// settlement must still be outside its occupied/adjacent site.
		state.Tiles = append(state.Tiles, Tile{ID: "foreign-site", X: 10, Terrain: TerrainPlains})
		state.Cities = []City{{ID: "foreign", Name: foreign, OwnerID: "player-b", TileID: "foreign-site"}}
		if err := foundCity(&state, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "home"}); err != nil {
			t.Fatal(err)
		}
		city, _ := cityByID(state, "home")
		if city.Name != "Rome" {
			t.Fatalf("civilization or foreign-dependent name: %s", city.Name)
		}
		if err := foundCity(&state, "player-b", FoundCityPayload{UnitID: "settler-b", CityID: "greek-home"}); err != nil {
			t.Fatal(err)
		}
		greek, _ := cityByID(state, "greek-home")
		if greek.Name != "Athens" {
			t.Fatalf("Greek city received %s", greek.Name)
		}
	}
}

func TestRuinCityUsesPinnedCivilizationName(t *testing.T) {
	state := civilizationNameState(t)
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	state.Tiles[0].Ruin = &Ruin{Profile: VillageRuinProfile, TypeRoll: 72, ValueRoll: 0}
	exploreRuin(&state, "player-a", "a", &runtime)
	city, ok := cityByID(state, "ruin-city:a")
	if !ok || city.Name != "Rome" {
		t.Fatalf("ruin bypassed civilization names: %+v", city)
	}
}

func TestCivilizationNameSuggestionExhaustionAndCanonicalRestart(t *testing.T) {
	state := civilizationNameState(t)
	for _, name := range []string{"Rome", "Antium", "Rome1", "Antium1"} {
		state.Cities = append(state.Cities, City{ID: name, Name: name, OwnerID: "player-a", TileID: "a"})
	}
	// The chooser itself needs only owned display names; no global name registry,
	// counters, RNG, command/event IDs, or hidden lookup is involved.
	if got := defaultCityName(state, "player-a"); got != "Rome2" {
		t.Fatalf("exhausted list = %s", got)
	}
	state.Cities = nil
	raw, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	inspector, err := NewInspector(restored)
	if err != nil {
		t.Fatal(err)
	}
	if got := inspector.SuggestedCityName("player-a"); got != "Rome" {
		t.Fatalf("restart changed name: %s", got)
	}
	if err := foundCity(&restored, "player-a", FoundCityPayload{UnitID: "settler-a", CityID: "custom", Name: " My capital "}); err != nil {
		t.Fatal(err)
	}
	city, _ := cityByID(restored, "custom")
	if city.Name != "My capital" {
		t.Fatal("custom name was overwritten")
	}
}
