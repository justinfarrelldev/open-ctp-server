package fullgame

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestHumanDiplomacySchedulerAdvancesEachEligibleUnitWithoutBounce(t *testing.T) {
	scheduler := newHumanDiplomacyScheduler(3)
	projection := diplomaticProjection([]visibility.UnitProjection{
		{ID: "unit-a", OwnerID: "human", TileID: "a-origin", Domain: "land", Movement: 1},
		{ID: "unit-b", OwnerID: "human", TileID: "b-origin", Domain: "land", Movement: 1},
	})

	first, ok := scheduler.nextAction(projection, "human")
	if !ok || first.kind != diplomaticMove || first.unitID != "unit-a" || first.path[0] != "a-north" {
		t.Fatalf("first action = %#v, %t", first, ok)
	}
	scheduler.recordSuccess(first)
	projection.Units[0].TileID = "a-north"
	projection.Units[0].Movement = 0

	second, ok := scheduler.nextAction(projection, "human")
	if !ok || second.kind != diplomaticMove || second.unitID != "unit-b" || second.path[0] != "b-north" {
		t.Fatalf("second action = %#v, %t", second, ok)
	}
	scheduler.recordSuccess(second)
	projection.Units[0].Movement = 1
	projection.Units[1].TileID = "b-north"
	projection.Units[1].Movement = 0

	third, ok := scheduler.nextAction(projection, "human")
	if !ok || third.kind != diplomaticMove || third.unitID != "unit-a" || third.path[0] != "a-east" {
		t.Fatalf("third action = %#v, %t", third, ok)
	}
}

func TestEmbassyApproachUsesStableReachableVisibleTile(t *testing.T) {
	projection := visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "approach", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "stale-approach", X: 2, Y: 1, Stale: true, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "blocked-approach", X: 2, Y: -1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: false}},
			{ID: "city-tile", X: 2, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "human", TileID: "origin", Domain: "land", Movement: 2}},
	}
	city := visibility.CityProjection{ID: "foreign-city", OwnerID: "ai", TileID: "city-tile"}
	path, ok := embassyApproach(projection, projection.Units[0], city)
	if !ok || len(path) != 1 || path[0] != "approach" {
		t.Fatalf("embassy approach = %v, ok=%v", path, ok)
	}

	projection.Tiles[1].Stale = true
	if path, ok := embassyApproach(projection, projection.Units[0], city); ok || path != nil {
		t.Fatalf("unreachable embassy approach = %v, ok=%v", path, ok)
	}
}

func diplomaticProjection(units []visibility.UnitProjection) visibility.Projection {
	return visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "a-origin", X: 0, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "a-north", X: 0, Y: -1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "a-east", X: 1, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "b-origin", X: 10, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "b-north", X: 10, Y: -1, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "b-east", X: 11, Y: 0, Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units: units,
	}
}
