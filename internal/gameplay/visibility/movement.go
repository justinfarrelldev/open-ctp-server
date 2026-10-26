package visibility

import "github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"

func projectedDomainAllows(domains []string, domain string) bool {
	return containsString(domains, domain) || (domain == engine.DomainAmphibious && (containsString(domains, engine.DomainLand) || containsString(domains, engine.DomainSea)))
}

func projectedMovers(p Projection, leader UnitProjection) []UnitProjection {
	if leader.ArmyID == "" {
		return []UnitProjection{leader}
	}
	members := []UnitProjection{}
	for _, unit := range p.Units {
		if unit.ArmyID == leader.ArmyID && unit.OwnerID == leader.OwnerID {
			members = append(members, unit)
		}
	}
	return members
}

func projectedCanEnter(p Projection, leader UnitProjection, tile TileProjection) bool {
	members := projectedMovers(p, leader)
	for _, member := range members {
		if !projectedDomainAllows(tile.Planning.Domains, member.Domain) {
			return false
		}
	}
	if p.Rules == nil || p.Rules.Movement == nil || p.Rules.Movement.RetailMovementVersion != 1 {
		return true
	}
	for _, member := range members {
		for _, definition := range p.Rules.Units {
			if definition.ID == member.TypeID && definition.Movement == 0 {
				return false
			}
		}
	}
	count := 0
	for _, unit := range p.Units {
		if unit.TileID == tile.ID && unit.TransportID == "" {
			count++
		}
	}
	if count+len(members) > 12 {
		return false
	}
	for _, city := range p.Cities {
		if city.TileID == tile.ID && city.OwnerID != leader.OwnerID {
			for _, member := range members {
				if hasEffect(member.Effects, engine.EffectCantCaptureCity) {
					return false
				}
			}
		}
	}
	return true
}

func projectedStepCost(p Projection, unit UnitProjection, cost int) int {
	version := 0
	if p.Rules != nil && p.Rules.Movement != nil {
		version = p.Rules.Movement.RetailMovementVersion
	}
	return engine.UnitStepCost(version, unit.Domain, cost)
}

// Each army member carries its own budget and first-move fact. Never infer an
// unused allowance from a full budget or from the selected leader's state.
func projectedStepAffordable(p Projection, leader UnitProjection, cost int, path []string) bool {
	members := projectedMovers(p, leader)
	for _, member := range members {
		remaining, first := member.Movement, member.FirstMove
		spent := false
		for _, id := range path {
			found := false
			for _, tile := range p.Tiles {
				if tile.ID == id && tile.Planning != nil {
					var ok bool
					remaining, ok = engine.MovementRemaining(remaining, projectedStepCost(p, member, tile.Planning.Cost), first)
					if !ok {
						return false
					}
					first = &spent
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		if _, ok := engine.MovementRemaining(remaining, projectedStepCost(p, member, cost), first); !ok {
			return false
		}
	}
	return len(members) > 0
}
