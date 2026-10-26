package fullgame

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestContinuityPreparationRejectsDistinctButAdjacentCitySites(t *testing.T) {
	p := visibility.Projection{Units: []visibility.UnitProjection{
		{ID: "a", OwnerID: "owner", TileID: "tile-2", Domain: engine.DomainLand, Movement: 4},
		{ID: "b", OwnerID: "owner", TileID: "tile-1", Domain: engine.DomainLand, Movement: 4},
		{ID: "c", OwnerID: "owner", TileID: "tile-3", Domain: engine.DomainLand, Movement: 4},
	}}
	for x := 0; x <= 4; x++ {
		p.Tiles = append(p.Tiles, visibility.TileProjection{ID: fmt.Sprintf("tile-%d", x), X: x, Planning: &visibility.TilePlanning{Accessible: true, Cost: 1, Domains: []string{engine.DomainLand}}})
	}
	raw, _ := json.Marshal(p)
	commands, preparing, err := resolveContinuityPreparationCommands(AcceptanceConfig{Rules: AcceptanceRulesFixture{HumanScript: "scripts/score-resignation-continuity.json"}}, raw, "owner")
	if err != nil || !preparing || len(commands) != 2 {
		t.Fatalf("adjacent sites were considered prepared: commands=%+v preparing=%v err=%v", commands, preparing, err)
	}
	for _, command := range commands {
		var move struct {
			UnitID string   `json:"unit_id"`
			Path   []string `json:"path"`
		}
		if err := json.Unmarshal(command.Payload, &move); err != nil {
			t.Fatal(err)
		}
		for i := range p.Units {
			if p.Units[i].ID == move.UnitID {
				p.Units[i].TileID = move.Path[len(move.Path)-1]
			}
		}
	}
	var centers []engine.Tile
	for _, unit := range p.Units {
		for _, tile := range p.Tiles {
			if tile.ID != unit.TileID {
				continue
			}
			center := engine.Tile{ID: tile.ID, X: tile.X, Y: tile.Y}
			if !engine.CitySiteAvailable(center, centers) {
				t.Fatalf("harness still selected adjacent city sites: %+v", p.Units)
			}
			centers = append(centers, center)
		}
	}
}

func TestContinuityPreparationWaitsForRealMovementRefresh(t *testing.T) {
	for _, status := range []string{"", visibility.PlanningStatusOpen, visibility.PlanningStatusQueued} {
		p := visibility.Projection{PlanningStatus: status,
			Tiles: []visibility.TileProjection{{ID: "site"}},
			Units: []visibility.UnitProjection{
				{ID: "a", TileID: "site", Domain: engine.DomainLand},
				{ID: "b", TileID: "site", Domain: engine.DomainLand},
				{ID: "c", TileID: "site", Domain: engine.DomainLand},
			},
		}
		commands, preparing, err := prepareContinuityCitySites(p, p.Units)
		want := "pass"
		if status != "" {
			want = "round-complete"
		}
		if err != nil || !preparing || len(commands) != 1 || commands[0].Command != want {
			t.Fatalf("exhausted preparation status=%q: %+v %v %v", status, commands, preparing, err)
		}
	}
}

func TestContinuityPreparationCanRevealSitesBeyondInitialSight(t *testing.T) {
	p := visibility.Projection{Units: []visibility.UnitProjection{
		{ID: "a", TileID: "tile-0", Domain: engine.DomainLand, Movement: 2},
		{ID: "b", TileID: "tile-0", Domain: engine.DomainLand, Movement: 2},
		{ID: "c", TileID: "tile-0", Domain: engine.DomainLand, Movement: 2},
	}}
	addTile := func(x int) {
		p.Tiles = append(p.Tiles, visibility.TileProjection{ID: fmt.Sprintf("tile-%d", x), X: x, Planning: &visibility.TilePlanning{Accessible: true, Cost: 1, Domains: []string{engine.DomainLand}}})
	}
	for x := -1; x <= 1; x++ {
		addTile(x)
	}
	for stage := 0; stage < 3; stage++ {
		commands, preparing, err := prepareContinuityCitySites(p, p.Units)
		if err != nil {
			t.Fatal(err)
		}
		if stage == 2 {
			if preparing || len(commands) != 0 {
				t.Fatalf("non-adjacent sites still need preparation: %+v", commands)
			}
			return
		}
		if !preparing || len(commands) != 2 {
			t.Fatalf("stage %d: %+v", stage, commands)
		}
		for _, command := range commands {
			var move struct {
				UnitID string   `json:"unit_id"`
				Path   []string `json:"path"`
			}
			if command.Command != "move" {
				t.Fatalf("stage %d closed before moving: %+v", stage, command)
			}
			if err := json.Unmarshal(command.Payload, &move); err != nil {
				t.Fatal(err)
			}
			for i := range p.Units {
				if p.Units[i].ID == move.UnitID {
					p.Units[i].TileID = move.Path[len(move.Path)-1]
					p.Units[i].Movement--
				}
			}
		}
		if stage == 0 {
			addTile(-2)
			addTile(2)
		}
	}
}
