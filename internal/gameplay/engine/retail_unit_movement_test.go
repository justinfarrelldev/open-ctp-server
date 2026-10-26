package engine

import (
	"path/filepath"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// One subtest per released unit, with expectations taken from the original
// Units.txt/Terrain.txt records rather than the projected runtime's values.
func TestEveryReleasedUnitMovement(t *testing.T) {
	db, err := rules.LoadRetailRulesFromDirectory(filepath.Join("..", "..", "..", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"))
	if err != nil {
		t.Fatal(err)
	}
	runtime := rules.ProjectRetailRuntime(db)
	if len(db.Units) != 74 {
		t.Fatalf("released unit coverage changed: %d", len(db.Units))
	}
	for _, original := range db.Units {
		t.Run(original.ID, func(t *testing.T) {
			definition, ok := runtime.Unit(original.ID)
			if !ok {
				t.Fatal("unit missing from runtime")
			}
			if definition.Movement != original.MaxMovePoints {
				t.Fatalf("movement %d != original %d", definition.Movement, original.MaxMovePoints)
			}
			trader := false
			for _, flag := range original.Flags {
				if flag == "IsTrader" {
					trader = true
				}
			}
			if trader {
				// Player::CreateUnit adds MaxMovePoints as trade transport points;
				// traders are not mobile pieces on the world map.
				s := journeyState(t)
				city := City{ID: "home", OwnerID: "player-a", TileID: "t0", Population: 1}
				beforeUnits := len(s.Units)
				beforePoints := nationForID(&s, "nation-a").TradeTransportPoints
				if err := completeUnitProduction(&s, &city, ProductionItem{ID: original.ID, Kind: ProductionUnit}, &runtime); err != nil {
					t.Fatal(err)
				}
				if len(s.Units) != beforeUnits || nationForID(&s, "nation-a").TradeTransportPoints-beforePoints != original.MaxMovePoints {
					t.Fatal("trader allowance did not become original transport points")
				}
				return
			}
			// Prefer expensive terrain to exercise land first-move and fixed-cost air
			// movement; ships and space units use their own released legal environment.
			terrainID := ""
			cost := 0
			air := false
			for _, kind := range original.MovementTypes {
				if kind == "Air" {
					air = true
				}
			}
			for _, terrain := range db.Terrains {
				canEnter := false
				for _, environment := range terrain.MovementTypes {
					for _, kind := range original.MovementTypes {
						if environment == kind {
							canEnter = true
						}
					}
				}
				if canEnter && terrain.EnvBase.Movement > cost {
					terrainID, cost = terrain.ID, terrain.EnvBase.Movement
				}
			}
			if terrainID == "" {
				t.Fatalf("no released terrain for domain %s", definition.Domain)
			}
			if air {
				cost = 100
			} // k_MOVE_AIR_COST, ArmyData::DeductMoveCost
			s := journeyState(t)
			s.Tiles = []Tile{{ID: "origin", Terrain: terrainID, Open: true}, {ID: "destination", X: 1, Terrain: terrainID, Open: true}}
			s.Cities = nil
			unit := Unit{ID: original.ID, OwnerID: "player-a", TileID: "origin"}
			applyUnitDefinition(&unit, definition)
			unit.Movement = refreshedUnitMovement(s, unit, definition.Movement)
			unit.FirstMove = NewFirstMove(&runtime)
			s.Units = []Unit{unit}
			if unit.Movement != original.MaxMovePoints {
				t.Fatalf("fresh allowance %d != original %d", unit.Movement, original.MaxMovePoints)
			}
			err := moveUnit(&s, "player-a", moveUnitPayload{UnitID: unit.ID, Path: []string{"destination"}}, &runtime)
			if original.MaxMovePoints == 0 {
				if err == nil {
					t.Fatal("stationary city unit moved")
				}
				return
			}
			if err != nil {
				t.Fatalf("original legal first step: %v", err)
			}
			got := s.Units[0]
			if got.TileID != "destination" || got.Movement != max(0, original.MaxMovePoints-cost) || got.FirstMove == nil || *got.FirstMove {
				t.Fatalf("step result %+v; expected remaining %d", got, max(0, original.MaxMovePoints-cost))
			}
		})
	}
}
