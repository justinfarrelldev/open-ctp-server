package engine

import (
	"path/filepath"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

func TestReleasedCatalogExecutesGovernorAndNomadRewards(t *testing.T) {
	db, err := rules.LoadRetailRulesFromDirectory(filepath.Join("..", "..", "..", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"))
	if err != nil {
		t.Fatal(err)
	}
	runtime := rules.ProjectRetailRuntime(db)
	if len(GovernorPriorities(&runtime)) != 7 {
		t.Fatal("released focuses not executable")
	}
	defs, hash, err := rules.CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state := journeyState(t)
	state.Rules = RulesPin{Version: runtime.CompatibilityVersion, ContentHash: hash, Definitions: defs}
	for index := range state.Tiles {
		state.Tiles[index].Terrain = "TERRAIN_PLAINS"
	}
	state.Units[0].Movement = runtime.Starter.Movement
	state.Units[0].MaxMovement = runtime.Starter.Movement
	state.Cities = append(state.Cities, City{ID: "home", OwnerID: "player-a", TileID: "t0", Population: 3})
	nation := nationForID(&state, "nation-a")
	for _, advance := range runtime.Advances {
		nation.Advances = append(nation.Advances, advance.ID)
	}
	state = runtimeApply(t, state, "player-a", CommandSetGovernor, map[string]any{"city_id": "home", "enabled": true, "priority": "production"})
	city, _ := cityByID(state, "home")
	if len(city.Queue) != 1 || city.Queue[0].EntryID == "" {
		t.Fatal("released production governor did not choose executable content")
	}
	if _, ok := runtime.ProductionDefinition(city.Queue[0].ID); !ok {
		t.Fatal("governor invented content")
	}
	state = withTestRuin(t, state, "t1", 75, 9999)
	state = runtimeApply(t, state, "player-a", CommandMoveUnit, map[string]any{"unit_id": "traveller", "path": []string{"t1"}})
	unit, ok := unitByID(state, "ruin-unit:t1")
	if !ok || unit.TypeID != "UNIT_SETTLER" || !hasEffect(unit.Effects, EffectSettler) {
		t.Fatalf("released nomads not original land settlers: %+v", unit)
	}
}
