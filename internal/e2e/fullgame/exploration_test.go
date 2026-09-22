package fullgame

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type explorationTransport struct {
	projections [][]byte
	moves       [][]string
}

func (t *explorationTransport) Projection(context.Context) ([]byte, error) {
	if len(t.projections) == 0 {
		return nil, errors.New("no projection")
	}
	return t.projections[0], nil
}

func (t *explorationTransport) Move(_ context.Context, _ string, path []string) ([]byte, error) {
	t.moves = append(t.moves, append([]string(nil), path...))
	if len(t.projections) < 2 {
		return nil, errors.New("no refreshed projection")
	}
	t.projections = t.projections[1:]
	return t.projections[0], nil
}

func explorationProjection(t *testing.T, unitTile string, enemy bool) []byte {
	t.Helper()
	projection := visibility.Projection{Tiles: []visibility.TileProjection{
		{ID: "origin", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		{ID: "east", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		{ID: "north", X: 0, Y: -1, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
	}, Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "account-1", TileID: unitTile, Domain: "land", Movement: 2}}}
	if enemy {
		projection.Units = append(projection.Units, visibility.UnitProjection{ID: "enemy", OwnerID: "account-2", TileID: "east"})
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func TestFrontierExplorerUsesStableVisibleCoordinateTieBreak(t *testing.T) {
	template := FixedFrontierExplorer(2)
	transport := &explorationTransport{projections: [][]byte{explorationProjection(t, "origin", false), explorationProjection(t, "north", true)}}
	discovery, err := template.Explore(context.Background(), transport, "scout")
	if err != nil {
		t.Fatal(err)
	}
	if discovery.EnemyUnitID != "enemy" || discovery.Moves != 1 {
		t.Fatalf("discovery = %#v", discovery)
	}
	if len(transport.moves) != 1 || len(transport.moves[0]) != 1 || transport.moves[0][0] != "north" {
		t.Fatalf("frontier moves = %#v", transport.moves)
	}
}

func TestFrontierExplorerFailsClosedWithoutVisibleLegalFrontier(t *testing.T) {
	projection := visibility.Projection{Tiles: []visibility.TileProjection{{ID: "origin", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}}}, Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "account-1", TileID: "origin", Domain: "land", Movement: 2}}}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	_, err = FixedFrontierExplorer(1).Explore(context.Background(), &explorationTransport{projections: [][]byte{encoded}}, "scout")
	if !errors.Is(err, ErrNoVisibleFrontier) {
		t.Fatalf("Explore() error = %v, want %v", err, ErrNoVisibleFrontier)
	}
}

func TestFrontierExplorerSkipsUnaffordableVisibleFrontier(t *testing.T) {
	projection := visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
			{ID: "forest", X: 0, Y: -1, Terrain: "forest", Planning: &visibility.TilePlanning{Cost: 2, Domains: []string{"land"}, Accessible: true}},
			{ID: "east", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		},
		Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "account-1", TileID: "origin", Domain: "land", Movement: 1}},
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	transport := &explorationTransport{projections: [][]byte{encoded, explorationProjection(t, "north", true)}}
	if _, err := FixedFrontierExplorer(2).Explore(context.Background(), transport, "scout"); err != nil {
		t.Fatal(err)
	}
	if len(transport.moves) != 1 || transport.moves[0][0] != "east" {
		t.Fatalf("reachable frontier move = %#v", transport.moves)
	}
}

func TestCityDirectedFrontierUsesReachableVisibleStep(t *testing.T) {
	projection := visibility.Projection{Tiles: []visibility.TileProjection{
		{ID: "origin", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		{ID: "toward-city", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		{ID: "away", X: 0, Y: 1, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		{ID: "city", X: 2, Y: 0, Terrain: "plains"},
	}, Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "account-1", TileID: "origin", Domain: "land", Movement: 1}}}
	mover := projection.Units[0]
	target := projection.Tiles[3]
	_, path, ok := nextReachableFrontierToward(projection, mover, map[string]bool{"origin": true}, "scout", target)
	if !ok || len(path) != 1 || path[0] != "toward-city" {
		t.Fatalf("city-directed frontier = %#v, ok=%v", path, ok)
	}
}

