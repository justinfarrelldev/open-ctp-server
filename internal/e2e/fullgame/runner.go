package fullgame

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type PublicConquestRunner struct {
	Transport    *HTTPExplorationTransport
	Database     *sqlx.DB
	ActionBudget int
	TurnBudget   int
	Advance      func(context.Context) error
	Restart      func() error
	Resume       func(context.Context, int64) (int64, []byte, error)
	Terminal     func(context.Context) (int64, string, []byte, error)
}

type PublicConquestResult struct {
	Terminal TerminalFrame
	Audit    AuditArtifacts
	Cursor   int64
}

var errConquestTurnBudget = errors.New("public conquest turn budget exhausted")

type conquestBudget struct {
	actions int
	turns   int
}

func (b *conquestBudget) move(ctx context.Context, transport ExplorationTransport, unitID string, path []string) ([]byte, error) {
	if b == nil || b.actions < 1 {
		return nil, ErrExplorationBudget
	}
	projection, err := transport.Move(ctx, unitID, path)
	if err != nil {
		return nil, err
	}
	b.actions--
	return projection, nil
}

func (b *conquestBudget) advance(ctx context.Context, advance func(context.Context) error) error {
	if b == nil || b.turns < 1 {
		return errConquestTurnBudget
	}
	if err := advance(ctx); err != nil {
		return err
	}
	b.turns--
	return nil
}

type budgetedExplorationTransport struct {
	transport ExplorationTransport
	budget    *conquestBudget
}

func (t budgetedExplorationTransport) Projection(ctx context.Context) ([]byte, error) {
	return t.transport.Projection(ctx)
}

func (t budgetedExplorationTransport) Move(ctx context.Context, unitID string, path []string) ([]byte, error) {
	return t.budget.move(ctx, t.transport, unitID, path)
}

func (r PublicConquestRunner) Execute(ctx context.Context, unitID, attackerArmyID string) (PublicConquestResult, error) {
	if r.Transport == nil || r.Database == nil || r.Advance == nil || r.Restart == nil || r.Resume == nil || r.Terminal == nil || r.ActionBudget < 1 || r.TurnBudget < 1 || unitID == "" || attackerArmyID == "" {
		return PublicConquestResult{}, fmt.Errorf("public conquest setup is invalid")
	}
	cursor, bootstrapProjection, err := r.Transport.Bootstrap(ctx, r.Transport.Cursor)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("bootstrap: %w", err)
	}
	unitIDs, err := visibleArmyUnitIDs(bootstrapProjection, unitID)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("army units: %w", err)
	}
	unitIDsJSON, err := json.Marshal(unitIDs)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("army units: %w", err)
	}
	if _, err := r.Transport.Command(ctx, "army-create", map[string]json.RawMessage{"army_id": json.RawMessage(fmt.Sprintf("%q", attackerArmyID)), "unit_ids": unitIDsJSON}); err != nil {
		return PublicConquestResult{}, fmt.Errorf("army-create: %w", err)
	}
	projection, err := r.Transport.Projection(ctx)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("attacker projection: %w", err)
	}
	attackerArmyID, err = ResolveVisibleArmy(projection, unitID)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("attacker army: %w", err)
	}
	if _, err := r.Transport.Command(ctx, "pass", map[string]json.RawMessage{}); err != nil {
		return PublicConquestResult{}, fmt.Errorf("pass before restart: %w", err)
	}
	if err := r.Restart(); err != nil {
		return PublicConquestResult{}, fmt.Errorf("restart: %w", err)
	}
	resumedCursor, _, err := r.Resume(ctx, cursor)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("resume: %w", err)
	}
	if resumedCursor <= cursor {
		return PublicConquestResult{}, fmt.Errorf("resume cursor %d did not advance beyond %d", resumedCursor, cursor)
	}
	explorationBudget := &conquestBudget{actions: r.ActionBudget, turns: r.TurnBudget}
	boundedTransport := budgetedExplorationTransport{transport: r.Transport, budget: explorationBudget}
	var discovery ExplorationDiscovery
	explorer := FixedFrontierExplorer(r.ActionBudget + 1)
	for {
		discovery, err = explorer.Explore(ctx, boundedTransport, unitID)
		if err == nil && discovery.EnemyUnitID != "" {
			break
		}
		if errors.Is(err, ErrExplorationBudget) {
			return PublicConquestResult{}, fmt.Errorf("exploration action budget exhausted; safe projection=%s", r.Transport.InitialProjection)
		}
		if err != nil && !errors.Is(err, ErrNoVisibleFrontier) {
			return PublicConquestResult{}, fmt.Errorf("exploration: %w; safe projection=%s", err, r.Transport.InitialProjection)
		}
		if err := explorationBudget.advance(ctx, r.Advance); err != nil {
			return PublicConquestResult{}, fmt.Errorf("exploration turn budget exhausted: %w; safe projection=%s", err, r.Transport.InitialProjection)
		}
	}
	if discovery.EnemyUnitID == "" {
		return PublicConquestResult{}, fmt.Errorf("exploration: no visible enemy")
	}
	defenderArmyID, err := ResolveVisibleArmy(discovery.Projection, discovery.EnemyUnitID)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("defender army: %w", err)
	}
	combat, err := r.Transport.Combat(ctx, attackerArmyID, defenderArmyID)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("combat: %w", err)
	}
	if !combat.Terminal {
		if err := r.Advance(ctx); err != nil {
			return PublicConquestResult{}, fmt.Errorf("advance after combat: %w", err)
		}
		if !r.Transport.Terminal {
			captureBudget := &conquestBudget{actions: r.ActionBudget, turns: r.TurnBudget}
			if err := captureAfterCombat(ctx, combat, func(ctx context.Context) error {
				return r.captureVisibleForeignCity(ctx, unitID, captureBudget)
			}); err != nil {
				return PublicConquestResult{}, fmt.Errorf("capture city: %w", err)
			}
		}
		progressionBudget := &conquestBudget{turns: r.TurnBudget}
		for !r.Transport.Terminal {
			if err := progressionBudget.advance(ctx, r.Advance); err != nil {
				return PublicConquestResult{}, fmt.Errorf("conquest progression completed without a terminal result: %w", err)
			}
		}
	}
	return r.finishTerminal(ctx, resumedCursor)
}

func (r PublicConquestRunner) finishTerminal(ctx context.Context, cursor int64) (PublicConquestResult, error) {
	sequence, name, data, err := r.Terminal(ctx)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("terminal: %w", err)
	}
	terminal, err := ParseTerminalSSE(sequence, name, data)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("terminal: %w", err)
	}
	audit, err := AuditTerminal(ctx, r.Database, r.Transport.GameID)
	if err != nil {
		return PublicConquestResult{}, fmt.Errorf("audit: %w", err)
	}
	if err := validateTerminalAgreement(audit, terminal, ""); err != nil {
		return PublicConquestResult{}, err
	}
	if sequence <= cursor {
		return PublicConquestResult{}, fmt.Errorf("terminal cursor %d did not advance beyond resumed cursor %d", sequence, cursor)
	}
	return PublicConquestResult{Terminal: terminal, Audit: audit, Cursor: sequence}, nil
}

func visibleArmyUnitIDs(raw []byte, selectedUnitID string) ([]string, error) {
	projection, selected, err := explorationInputs(raw, selectedUnitID)
	if err != nil {
		return nil, err
	}
	unitIDs := make([]string, 0)
	for _, unit := range projection.Units {
		if unit.OwnerID == selected.OwnerID && unit.TileID == selected.TileID && unit.ArmyID == "" && unit.Domain != "" {
			unitIDs = append(unitIDs, unit.ID)
		}
	}
	sort.Strings(unitIDs)
	if len(unitIDs) == 0 {
		return nil, ErrNoVisibleFrontier
	}
	return unitIDs, nil
}

func captureAfterCombat(ctx context.Context, combat CombatTransition, fallback func(context.Context) error) error {
	if combat.Result != nil && combat.Result.CapturedCityID != "" {
		return nil
	}
	return fallback(ctx)
}

func farthestKnownLandTile(projection visibility.Projection, origin visibility.TileProjection) (visibility.TileProjection, bool) {
	candidates := make([]visibility.TileProjection, 0, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		if tile.Terrain != "ocean" {
			candidates = append(candidates, tile)
		}
	}
	if len(candidates) == 0 {
		return visibility.TileProjection{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		left := abs(candidates[i].X-origin.X) + abs(candidates[i].Y-origin.Y)
		right := abs(candidates[j].X-origin.X) + abs(candidates[j].Y-origin.Y)
		if left != right {
			return left > right
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0], true
}

func visibleBacktrack(raw []byte, unitID string, trail []string) ([]string, []string, bool) {
	var projection visibility.Projection
	if err := json.Unmarshal(raw, &projection); err != nil {
		return nil, trail, false
	}
	for index := len(trail) - 1; index >= 0; index-- {
		path, err := visibility.Plan(projection, unitID, trail[index])
		if err == nil && len(path) > 0 {
			remaining := append([]string(nil), trail[:index]...)
			return path[:1], remaining, true
		}
	}
	return nil, trail, false
}

func validateTerminalAgreement(audit AuditArtifacts, terminal TerminalFrame, expectedReason string) error {
	if audit.TerminalWinner != terminal.Winner || audit.TerminalReason != terminal.Reason {
		return fmt.Errorf("audit terminal mismatch: audit=%s/%s sse=%s/%s", audit.TerminalWinner, audit.TerminalReason, terminal.Winner, terminal.Reason)
	}
	if expectedReason != "" && terminal.Reason != expectedReason {
		return fmt.Errorf("terminal reason %q does not match expected %q", terminal.Reason, expectedReason)
	}
	return nil
}

func moveKnownCityStep(ctx context.Context, transport ExplorationTransport, budget *conquestBudget, unitID string, path []string) (bool, error) {
	if len(path) == 0 {
		return false, ErrNoVisibleFrontier
	}
	if _, err := budget.move(ctx, transport, unitID, path[:1]); err != nil {
		return false, err
	}
	return len(path) == 1, nil
}

func (r PublicConquestRunner) captureVisibleForeignCity(ctx context.Context, unitID string, budget *conquestBudget) error {
	if budget == nil {
		return errors.New("public conquest budget is required")
	}
	visited := make(map[string]bool)
	trail := make([]string, 0)
	var origin visibility.TileProjection
	hasOrigin := false
	for {
		for budget.actions > 0 {
			projection, err := r.Transport.Projection(ctx)
			if err != nil {
				return err
			}
			decoded, mover, err := explorationInputs(projection, unitID)
			if err != nil {
				return err
			}
			if !hasOrigin {
				for _, tile := range decoded.Tiles {
					if tile.ID == mover.TileID {
						origin, hasOrigin = tile, true
						break
					}
				}
			}
			cities := append([]visibility.CityProjection(nil), decoded.Cities...)
			sort.Slice(cities, func(i, j int) bool { return cities[i].ID < cities[j].ID })
			var cityTarget *visibility.TileProjection
			for _, city := range cities {
				if city.OwnerID == "" || city.OwnerID == mover.OwnerID {
					continue
				}
				if !city.Stale {
					path, err := visibility.PlanKnown(decoded, unitID, city.TileID)
					if err == nil && len(path) > 0 {
						reached, err := moveKnownCityStep(ctx, r.Transport, budget, unitID, path)
						if err != nil {
							return err
						}
						if reached {
							return nil
						}
						continue
					}
				}
				if cityTarget != nil {
					continue
				}
				for _, tile := range decoded.Tiles {
					if tile.ID == city.TileID {
						copy := tile
						cityTarget = &copy
						break
					}
				}
			}
			visited[mover.TileID] = true
			var path []string
			var ok bool
			searchTarget := cityTarget
			if searchTarget == nil && hasOrigin {
				if target, found := farthestKnownLandTile(decoded, origin); found {
					searchTarget = &target
				}
			}
			if searchTarget != nil {
				_, path, ok = nextReachableFrontierToward(decoded, mover, visited, unitID, *searchTarget)
			} else {
				_, path, ok = nextReachableFrontier(decoded, mover, visited, unitID)
			}
			if ok {
				trail = append(trail, mover.TileID)
			} else {
				path, trail, ok = visibleBacktrack(projection, unitID, trail)
			}
			if !ok {
				break
			}
			if _, err := budget.move(ctx, r.Transport, unitID, path[:1]); err != nil {
				return err
			}
		}
		if budget.actions < 1 {
			return fmt.Errorf("visible city exploration action budget exhausted: %w; safe projection=%s", ErrExplorationBudget, r.Transport.InitialProjection)
		}
		if err := budget.advance(ctx, r.Advance); err != nil {
			return fmt.Errorf("no visible foreign city capture route: %w; safe projection=%s", err, r.Transport.InitialProjection)
		}
	}
}
