package engine

// MovementUnitVisible is shared by projection and server routing so an
// undetected occupant cannot act as an omniscient route-selection oracle.
func MovementUnitVisible(state State, playerID string, visible map[string]bool, unit Unit) bool {
	if unit.OwnerID == playerID {
		return true
	}
	if !visible[unit.TileID] {
		return false
	}
	if hasEffect(unit.Effects, EffectInvisibleForGlobeSat) && HasActiveWonderEffect(state, playerID, EffectGlobalRadar) && !MovementSight(state, playerID, false)[unit.TileID] {
		return false
	}
	submarine, stealth := hasEffect(unit.Effects, EffectSubmarine), hasEffect(unit.Effects, EffectStealth)
	if !submarine && !stealth {
		return true
	}
	return (submarine && movementDetectorInRange(state, playerID, unit.TileID, EffectCanAttackUnderwater)) || (stealth && movementDetectorInRange(state, playerID, unit.TileID, EffectCanSeeStealth))
}

func movementDetectorInRange(state State, playerID, tileID, kind string) bool {
	target, ok := tileByID(state, tileID)
	if !ok {
		return false
	}
	for _, unit := range state.Units {
		if unit.OwnerID != playerID || unit.TileID == "" || unit.TransportID != "" || !hasEffect(unit.Effects, kind) {
			continue
		}
		origin, ok := tileByID(state, unit.TileID)
		if ok && tileDistance(origin, target) <= max(2, unit.VisibilityRange) {
			return true
		}
	}
	for _, overlay := range state.Overlays {
		if overlay.OwnerID != playerID || overlay.TileID == "" || overlay.Remaining > 0 || overlay.Kind != "improvement" {
			continue
		}
		visionClass := "Stealth"
		if kind == EffectCanAttackUnderwater {
			visionClass = "Underwater"
		}
		if !containsString(overlay.CanSee, visionClass) {
			continue
		}
		origin, ok := tileByID(state, overlay.TileID)
		if ok && tileDistance(origin, target) <= max(2, max(overlay.VisionRange, overlay.RadarRange)) {
			return true
		}
	}
	return false
}
