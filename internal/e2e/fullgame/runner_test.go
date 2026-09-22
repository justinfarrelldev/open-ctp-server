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
