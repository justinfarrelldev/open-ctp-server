package game

import (
	"encoding/json"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
	"github.com/justinfarrelldev/open-ctp-server/internal/world"
)

func TestInitialGameplayPinsOriginalResourceGenerationPolicy(t *testing.T) {
	t.Setenv(retailRulesModeEnv, "")
	input := world.GenerationInput{GameID: 9, PlayerOrder: []int64{1, 2}, GeneratorVersion: world.OriginalGoodsGeneratorVersion, HostParams: world.GenerationHostParams{Style: "sequential", RulesJSON: "[]", ExclusionsJSON: "{}", StartingSettlers: 1, StartingGold: 100, MapSize: "small", Climate: "temperate"}}
	seed := int64(123)
	wetness := 50
	input.HostParams.Seed = &seed
	input.HostParams.Wetness = &wetness
	snapshot, _, err := world.GenerateWorld(input)
	if err != nil {
		t.Fatal(err)
	}
	pin, err := serverRuntimePin()
	if err != nil {
		t.Fatal(err)
	}
	state, err := initialGameplayState(snapshot, "sequential", pin, []GamePlayer{{AccountID: 1}, {AccountID: 2}})
	if err != nil {
		t.Fatal(err)
	}
	for index, tile := range state.Tiles {
		if tile.Terrain != snapshot.Tiles[index].Terrain || tile.GoodID != snapshot.Tiles[index].GoodID {
			t.Fatalf("handoff changed resource terrain lineage: %+v -> %+v", snapshot.Tiles[index], tile)
		}
	}
	p, err := visibility.Project(state, "account-1")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(p.World)
	var wire map[string]any
	_ = json.Unmarshal(raw, &wire)
	if wire["resource_catalog"] != world.OriginalResourceCatalogVersion || wire["generator_version"] != snapshot.GeneratorVersion {
		t.Fatalf("lost pinned resource policy: %s", raw)
	}
	for _, pair := range [][2]string{{"plains", "TERRAIN_DESERT_GOOD_TWO"}, {"plains", "UNDEFINED"}, {"plains", "TERRAIN_PLAINS_GOOD_FOUR"}, {"unknown", ""}} {
		invalid := snapshot.Clone()
		invalid.Tiles[0].Terrain = pair[0]
		invalid.Tiles[0].GoodID = pair[1]
		if _, err := initialGameplayState(invalid, "sequential", pin, []GamePlayer{{AccountID: 1}, {AccountID: 2}}); err == nil {
			t.Fatalf("invalid initial catalog/terrain accepted: %v", pair)
		}
	}
}
