package fullgame

import (
	"context"
	"errors"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

func TestCaptureAfterCombatHonorsPublicCapturedCityResult(t *testing.T) {
	fallbackCalled := false
	err := captureAfterCombat(context.Background(), CombatTransition{
		Result: &engine.CombatResult{CapturedCityID: "captured-city"},
	}, func(context.Context) error {
		fallbackCalled = true
		return errors.New("redundant city search")
	})
	if err != nil {
		t.Fatal(err)
	}
	if fallbackCalled {
		t.Fatal("captured combat result triggered a redundant city search")
	}
}

func TestConquestAttackerArmyKeepsCoLocatedReservePubliclyAvailable(t *testing.T) {
	projection := []byte(`{"units":[{"id":"selected","owner_id":"human","tile_id":"spawn","domain":"land"},{"id":"reserve","owner_id":"human","tile_id":"spawn","domain":"land"}]}`)
	unitIDs, err := visibleAttackerArmyUnitIDs(projection, "selected")
	if err != nil {
		t.Fatal(err)
	}
	if len(unitIDs) != 1 || unitIDs[0] != "selected" {
		t.Fatalf("attacker army units = %v, want only selected unit so reserve can capture after combat", unitIDs)
	}
}

func TestVisibleAttackerArmyUnitIDsAllowsAlreadyAssignedSelectedUnit(t *testing.T) {
	projection := []byte(`{"units":[{"id":"selected","owner_id":"human","tile_id":"spawn","army_id":"army","domain":"land"},{"id":"reserve","owner_id":"human","tile_id":"spawn","domain":"land"}]}`)
	unitIDs, err := visibleAttackerArmyUnitIDs(projection, "selected")
	if err != nil {
		t.Fatal(err)
	}
	if len(unitIDs) != 1 || unitIDs[0] != "selected" {
		t.Fatalf("attacker army units = %v, want only already assigned selected unit", unitIDs)
	}
}

func TestVisibleArmyUnitIDsIncludesEligiblePublicCoLocatedUnits(t *testing.T) {
	projection := []byte(`{"units":[{"id":"unit-b","owner_id":"human","tile_id":"spawn","domain":"land"},{"id":"unit-a","owner_id":"human","tile_id":"spawn","domain":"land"},{"id":"assigned","owner_id":"human","tile_id":"spawn","army_id":"other","domain":"land"},{"id":"away","owner_id":"human","tile_id":"elsewhere","domain":"land"},{"id":"enemy","owner_id":"ai","tile_id":"spawn","domain":"land"}]}`)
	unitIDs, err := visibleArmyUnitIDs(projection, "unit-b")
	if err != nil {
		t.Fatal(err)
	}
	if len(unitIDs) != 2 || unitIDs[0] != "unit-a" || unitIDs[1] != "unit-b" {
		t.Fatalf("visible army units = %v, want [unit-a unit-b]", unitIDs)
	}
}

func TestVisibleBacktrackUsesPublicTrailWithoutStaleTarget(t *testing.T) {
	projection := []byte(`{"revision":4,"turn":2,"tiles":[{"id":"a","x":0,"y":0,"terrain":"plains","planning":{"cost":1,"domains":["land"],"accessible":true}},{"id":"b","x":1,"y":0,"terrain":"plains","planning":{"cost":1,"domains":["land"],"accessible":true}},{"id":"c","x":2,"y":0,"terrain":"plains","planning":{"cost":1,"domains":["land"],"accessible":true}}],"units":[{"id":"unit","owner_id":"human","tile_id":"c","domain":"land","movement":2}]}`)
	path, trail, ok := visibleBacktrack(projection, "unit", []string{"a", "b"})
	if !ok || len(path) != 1 || path[0] != "b" {
		t.Fatalf("backtrack path = %v, ok=%v, want [b]", path, ok)
	}
	if len(trail) != 1 || trail[0] != "a" {
		t.Fatalf("remaining trail = %v, want [a]", trail)
	}
}

func TestFarthestKnownLandTileUsesPublicTerrainAndStableTieBreak(t *testing.T) {
	projection := visibility.Projection{Tiles: []visibility.TileProjection{
		{ID: "origin", X: 0, Y: 0, Terrain: "plains"},
		{ID: "far-b", X: 4, Y: 2, Terrain: "forest"},
		{ID: "far-a", X: 5, Y: 1, Terrain: "plains"},
		{ID: "ocean", X: 9, Y: 9, Terrain: "ocean"},
	}}
	tile, ok := farthestKnownLandTile(projection, visibility.TileProjection{ID: "origin", X: 0, Y: 0})
	if !ok || tile.ID != "far-a" {
		t.Fatalf("farthest land tile = %#v, ok=%v, want far-a", tile, ok)
	}
}

func TestFarthestKnownLandTileExcludingSkipsVisitedEdgeToAvoidOscillation(t *testing.T) {
	projection := visibility.Projection{Tiles: []visibility.TileProjection{
		{ID: "origin", X: 0, Y: 0, Terrain: "plains"},
		{ID: "visited-edge", X: 5, Y: 0, Terrain: "plains"},
		{ID: "next-edge", X: 4, Y: 0, Terrain: "plains"},
		{ID: "ocean", X: 6, Y: 0, Terrain: "ocean"},
	}}
	tile, ok := farthestKnownLandTileExcluding(projection, projection.Tiles[0], map[string]bool{"visited-edge": true})
	if !ok || tile.ID != "next-edge" {
		t.Fatalf("farthest unvisited land tile = %#v, ok=%v, want next-edge", tile, ok)
	}
}

func TestCaptureMoverSwitchesToMobileCoLocatedArmyMember(t *testing.T) {
	projection := visibility.Projection{Units: []visibility.UnitProjection{
		{ID: "spent", OwnerID: "account-1", TileID: "origin", Domain: "land", Movement: 0},
		{ID: "mobile", OwnerID: "account-1", TileID: "origin", Domain: "land", Movement: 1},
		{ID: "elsewhere", OwnerID: "account-1", TileID: "other", Domain: "land", Movement: 2},
	}}
	unitID, mover, ok := captureMoverUnit(projection, "spent")
	if !ok || unitID != "mobile" || mover.ID != "mobile" {
		t.Fatalf("capture mover = %q/%#v ok=%v, want mobile", unitID, mover, ok)
	}
}

func TestCaptureMoverSwitchesToVisibleLegalReserveWhenPreferredArmyCannotMove(t *testing.T) {
	projection := visibility.Projection{
		Units: []visibility.UnitProjection{
			{ID: "army-bound", OwnerID: "account-1", TileID: "origin", Domain: "land", Movement: 0},
			{ID: "reserve", OwnerID: "account-1", TileID: "reserve-origin", Domain: "land", Movement: 2},
			{ID: "enemy", OwnerID: "account-2", TileID: "reserve-origin", Domain: "land", Movement: 2},
		},
		LegalActions: []visibility.ActionProjection{
			{Command: engine.CommandMoveUnit, EntityID: "reserve", Payload: []byte(`{"unit_id":"reserve","path":["next"]}`)},
		},
	}
	unitID, mover, ok := captureMoverUnit(projection, "army-bound")
	if !ok || unitID != "reserve" || mover.ID != "reserve" {
		t.Fatalf("capture mover = %q/%#v ok=%v, want visible reserve", unitID, mover, ok)
	}
}

func TestCaptureFrontierUsesOnlyProjectedLegalMoveSteps(t *testing.T) {
	projection := visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}},
			{ID: "blocked-by-movement", X: 1, Y: 0, Terrain: "forest", Planning: &visibility.TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}},
			{ID: "legal-step", X: 0, Y: 1, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}},
		},
		Units: []visibility.UnitProjection{{ID: "mover", OwnerID: "human", TileID: "origin", Domain: "land", Movement: 1}},
		LegalActions: []visibility.ActionProjection{
			{Command: engine.CommandMoveUnit, EntityID: "mover", Payload: []byte(`{"unit_id":"mover","path":["legal-step"]}`)},
		},
	}
	_, path, ok := nextReachableFrontierToward(projection, projection.Units[0], map[string]bool{"origin": true}, "mover", visibility.TileProjection{ID: "target", X: 2, Y: 0})
	if !ok || len(path) == 0 || path[0] != "legal-step" {
		t.Fatalf("frontier path = %v ok=%v, want legal-step", path, ok)
	}
}

func TestCaptureKnownTargetStepUsesPublicMemoryRoute(t *testing.T) {
	projection := visibility.Projection{
		Tiles: []visibility.TileProjection{
			{ID: "origin", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}},
			{ID: "east", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}},
			{ID: "local", X: 0, Y: 1, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Accessible: true, Domains: []string{"land"}}},
			{ID: "remembered-target", X: 2, Y: 0, Terrain: "plains", Stale: true},
		},
		Units: []visibility.UnitProjection{{ID: "mover", OwnerID: "human", TileID: "origin", Domain: "land", Movement: 1}},
		LegalActions: []visibility.ActionProjection{
			{Command: engine.CommandMoveUnit, EntityID: "mover", Payload: []byte(`{"unit_id":"mover","path":["east"]}`)},
			{Command: engine.CommandMoveUnit, EntityID: "mover", Payload: []byte(`{"unit_id":"mover","path":["local"]}`)},
		},
	}
	path, ok := knownCaptureTargetStep(projection, "mover", visibility.TileProjection{ID: "remembered-target"})
	if !ok || len(path) != 1 || path[0] != "east" {
		t.Fatalf("known target step = %v ok=%v, want east", path, ok)
	}
}

func TestRecoverCaptureFrontierResetsStaleVisitedDeadEnd(t *testing.T) {
	projection := visibility.Projection{Tiles: []visibility.TileProjection{
		{ID: "origin", X: 0, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		{ID: "east", X: 1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
		{ID: "west", X: -1, Y: 0, Terrain: "plains", Planning: &visibility.TilePlanning{Cost: 1, Domains: []string{"land"}, Accessible: true}},
	}, Units: []visibility.UnitProjection{{ID: "scout", OwnerID: "account-1", TileID: "origin", Domain: "land", Movement: 1}}}
	mover := projection.Units[0]
	blockedVisited := map[string]bool{"origin": true, "east": true, "west": true}
	if _, _, ok := nextReachableFrontierToward(projection, mover, blockedVisited, "scout", projection.Tiles[1]); ok {
		t.Fatal("test setup unexpectedly had a route before visited reset")
	}
	path, ok := recoverCaptureFrontier(projection, mover, "scout", &projection.Tiles[1])
	if !ok || len(path) != 1 || path[0] != "east" {
		t.Fatalf("recovered capture frontier = %v, ok=%v, want [east]", path, ok)
	}
}

func TestCaptureAfterCombatRetainsProjectionFallbackWithoutCapturedCity(t *testing.T) {
	want := errors.New("projection fallback")
	fallbackCalled := false
	err := captureAfterCombat(context.Background(), CombatTransition{
		Result: &engine.CombatResult{},
	}, func(context.Context) error {
		fallbackCalled = true
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("capture fallback error = %v, want %v", err, want)
	}
	if !fallbackCalled {
		t.Fatal("empty captured city result skipped projection fallback")
	}
}

func TestMoveKnownCityStepRequiresFinalRouteStep(t *testing.T) {
	projection := explorationProjection(t, "origin", false)
	transport := &explorationTransport{projections: [][]byte{projection, projection, projection}}
	budget := &conquestBudget{actions: 2}

	reached, err := moveKnownCityStep(context.Background(), transport, budget, "scout", []string{"middle", "city"})
	if err != nil {
		t.Fatal(err)
	}
	if reached {
		t.Fatal("multi-step city route reported capture after its first step")
	}
	reached, err = moveKnownCityStep(context.Background(), transport, budget, "scout", []string{"city"})
	if err != nil {
		t.Fatal(err)
	}
	if !reached {
		t.Fatal("final city route step did not report arrival")
	}
	if len(transport.moves) != 2 || transport.moves[0][0] != "middle" || transport.moves[1][0] != "city" || budget.actions != 0 {
		t.Fatalf("city route moves = %v, remaining actions = %d", transport.moves, budget.actions)
	}
}

func TestConquestPhaseBudgetIsSharedAcrossMovesAndTurns(t *testing.T) {
	projection := explorationProjection(t, "origin", false)
	transport := &explorationTransport{projections: [][]byte{projection, projection}}
	budget := &conquestBudget{actions: 1, turns: 1}

	if _, err := budget.move(context.Background(), transport, "scout", []string{"first"}); err != nil {
		t.Fatal(err)
	}
	if _, err := budget.move(context.Background(), transport, "scout", []string{"second"}); !errors.Is(err, ErrExplorationBudget) {
		t.Fatalf("second phase move error = %v, want %v", err, ErrExplorationBudget)
	}
	advances := 0
	advance := func(context.Context) error {
		advances++
		return nil
	}
	if err := budget.advance(context.Background(), advance); err != nil {
		t.Fatal(err)
	}
	if err := budget.advance(context.Background(), advance); !errors.Is(err, errConquestTurnBudget) {
		t.Fatalf("second phase advance error = %v, want %v", err, errConquestTurnBudget)
	}
	if len(transport.moves) != 1 || advances != 1 {
		t.Fatalf("budget allowed moves=%d advances=%d, want 1/1", len(transport.moves), advances)
	}
}

func TestValidateTerminalAgreementRejectsSSEMismatchAndUnexpectedReason(t *testing.T) {
	audit := AuditArtifacts{TerminalWinner: "account-1", TerminalReason: "bloodlust_conquest"}
	if err := validateTerminalAgreement(audit, TerminalFrame{Winner: "account-1", Reason: "bloodlust_conquest"}, "bloodlust_conquest"); err != nil {
		t.Fatal(err)
	}
	if err := validateTerminalAgreement(audit, TerminalFrame{Winner: "account-1", Reason: "score_2300"}, "bloodlust_conquest"); err == nil {
		t.Fatal("mismatched SSE terminal reason was accepted")
	}
	wrongExpected := AuditArtifacts{TerminalWinner: "account-1", TerminalReason: "score_2300"}
	if err := validateTerminalAgreement(wrongExpected, TerminalFrame{Winner: "account-1", Reason: "score_2300"}, "bloodlust_conquest"); err == nil {
		t.Fatal("unexpected terminal reason was accepted")
	}
}
