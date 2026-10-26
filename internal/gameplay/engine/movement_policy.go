package engine

import "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"

// NewFirstMove initializes explicit movement state for corrected retail pins.
// Legacy saves have no such fact: they gain it only at a real turn refresh.
func NewFirstMove(runtime *rules.RuntimeRules) *bool {
	if runtime == nil || runtime.Movement.RetailMovementVersion != 1 {
		return nil
	}
	ready := true
	return &ready
}

// HasMovementAllowance does not infer first-move state from the budget.
func HasMovementAllowance(remaining int, firstMove *bool) bool {
	return remaining > 0 || (remaining == 0 && firstMove != nil && *firstMove)
}

// MovementRemaining is shared by execution and projected route planning. Unknown
// first-move state never grants an allowance. Fuel is checked independently.
func MovementRemaining(remaining, cost int, firstMove *bool) (int, bool) {
	if cost < 1 || remaining < 0 {
		return remaining, false
	}
	if remaining < cost && (firstMove == nil || !*firstMove) {
		return remaining, false
	}
	return max(0, remaining-cost), true
}

// UnitStepCost preserves the released fixed air cost (k_MOVE_AIR_COST).
func UnitStepCost(retailVersion int, domain string, terrainCost int) int {
	if retailVersion == 1 && domain == DomainAir {
		return 100
	}
	return terrainCost
}

func consumeFirstMove(unit *Unit) {
	if unit.FirstMove != nil {
		spent := false
		unit.FirstMove = &spent
	}
}

// CanAccessTerrain is a terrain policy only; cities, occupants and stack size
// are deliberately independent checks, even on open foreign land.
func CanAccessTerrain(runtime *rules.RuntimeRules, actorID string, tile Tile) bool {
	return runtime != nil && runtime.Movement.Configured && (runtime.Movement.RetailMovementVersion == 1 || tile.OwnerID == "" || tile.OwnerID == actorID || (runtime.Movement.OpenAccess && tile.Open))
}

// CanAccessMovementTerrain adds the unit-independent city/treaty restriction.
func CanAccessMovementTerrain(state State, runtime *rules.RuntimeRules, actorID string, tile Tile) bool {
	if !CanAccessTerrain(runtime, actorID, tile) {
		return false
	}
	if runtime.Movement.RetailMovementVersion == 1 {
		if city, ok := cityByTile(state, tile.ID); ok && city.OwnerID != actorID && !playersAtWar(state, actorID, city.OwnerID) {
			return false
		}
	}
	return true
}

func canEnterMovementTile(state State, runtime *rules.RuntimeRules, actorID string, movers []*Unit, domains []string, to Tile) bool {
	if !CanAccessMovementTerrain(state, runtime, actorID, to) || !domainCanEnter(runtime, domains, EffectiveTerrain(state, to.ID)) || occupiedByEnemy(state, actorID, to.ID) {
		return false
	}
	// Domain eligibility is per member, not equality of domain labels: air and
	// amphibious members may legally accompany land units on ordinary land.
	for _, member := range movers {
		if !domainCanEnter(runtime, []string{member.Domain}, EffectiveTerrain(state, to.ID)) {
			return false
		}
	}
	if runtime.Movement.RetailMovementVersion != 1 {
		return true
	}
	for _, member := range movers {
		if definition, ok := runtime.Unit(member.TypeID); ok && definition.Movement == 0 {
			return false
		}
	}
	if stackCount(state, to.ID)+len(movers) > maxArmySize {
		return false
	}
	if city, ok := cityByTile(state, to.ID); ok && city.OwnerID != actorID {
		if !playersAtWar(state, actorID, city.OwnerID) {
			return false
		}
		for _, mover := range movers {
			if hasEffect(mover.Effects, EffectCantCaptureCity) {
				return false
			}
		}
	}
	return true
}
