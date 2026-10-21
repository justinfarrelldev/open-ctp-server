package ai

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

// emergencyQueueCandidates uses the atomic editor, never cancel/re-add, when
// newly researched defense would finish sooner than an undefended city's settler.
// The original settler remains queued with its stable identity.
func emergencyQueueCandidates(projection visibility.Projection, playerID string) []Candidate {
	if !shouldProduceEmergencyMilitaryUnit(projection, playerID) || projection.Rules == nil {
		return nil
	}
	result := []Candidate{}
	known := ownedAdvances(projection)
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.Stale || len(city.Queue) == 0 || len(city.Queue) >= engine.MaxProductionQueueEntries {
			continue
		}
		advertised := false
		for _, action := range projection.LegalActions {
			if action.Command == engine.CommandEditProduction && action.EntityID == city.ID {
				advertised = true
				break
			}
		}
		if !advertised {
			continue
		}
		head := city.Queue[0]
		settler := false
		for _, unit := range projection.Rules.Units {
			if unit.ID == head.ID && runtimeUnitHasEffect(unit, engine.EffectSettler) {
				settler = true
				break
			}
		}
		if !settler {
			continue
		}
		bestID, bestCost := "", 0
		for _, unit := range projection.Rules.Units {
			if !simpleEmergencyDefender(unit) || !runtimeUnitAvailable(unit, known) {
				continue
			}
			cost := productionCost(unit)
			if bestID == "" || cost < bestCost || cost == bestCost && unit.ID < bestID {
				bestID, bestCost = unit.ID, cost
			}
		}
		if bestID == "" || bestCost >= head.Cost-city.StoredProduction {
			continue
		}
		entries := []engine.QueueSelection{{ItemID: bestID}}
		valid := true
		for _, item := range city.Queue {
			if item.EntryID == "" {
				valid = false
				break
			}
			entries = append(entries, engine.QueueSelection{EntryID: item.EntryID})
		}
		if valid {
			result = append(result, candidate(engine.CommandEditProduction, city.ID, bestID, 740, engine.EditProductionPayload{CityID: city.ID, Entries: entries}))
		}
	}
	return result
}

// emergencyDefenderResearch prevents an undefended civilization from committing
// all early production to expensive expansion while a cheap defender is one
// available advance away. It uses only projected catalogs and own state.
func emergencyDefenderResearch(projection visibility.Projection, playerID string) (string, bool) {
	if !shouldProduceEmergencyMilitaryUnit(projection, playerID) || projection.Research == nil || projection.Rules == nil {
		return "", false
	}
	if _, ok := bestProjectedCombatUnitProduction(projection); ok {
		return "", false
	}
	known := ownedAdvances(projection)
	costs := map[string]int{}
	for _, advance := range projection.Rules.Advances {
		costs[advance.ID] = advance.Cost
	}
	bestAdvance, bestUnit := "", ""
	bestProduction, bestResearch := 0, 0
	for _, advance := range projection.Research.Available {
		future := make(map[string]bool, len(known)+1)
		for id := range known {
			future[id] = true
		}
		future[advance] = true
		for _, unit := range projection.Rules.Units {
			if !simpleEmergencyDefender(unit) || !runtimeUnitAvailable(unit, future) {
				continue
			}
			production, research := productionCost(unit), costs[advance]
			earlierID := advance < bestAdvance || advance == bestAdvance && unit.ID < bestUnit
			betterResearch := research < bestResearch || research == bestResearch && earlierID
			if bestAdvance == "" || production < bestProduction || production == bestProduction && betterResearch {
				bestAdvance, bestUnit, bestProduction, bestResearch = advance, unit.ID, production, research
			}
		}
	}
	return bestAdvance, bestAdvance != ""
}

// simpleEmergencyDefender deliberately excludes population-consuming and
// specialized prerequisite paths instead of guessing city/government eligibility.
// Execution still validates all remaining rules through the normal command gate.
func simpleEmergencyDefender(unit rules.RuntimeUnitDefinition) bool {
	if unit.Domain != engine.DomainLand && unit.Domain != engine.DomainAmphibious {
		return false
	}
	if unit.Attack <= 0 && unit.Defense <= 0 {
		return false
	}
	if unit.NeedsCityGoodID != "" || unit.RequiredGovernmentID != "" || len(unit.RequiredGovernmentIDs) > 0 {
		return false
	}
	for _, blocked := range []string{engine.EffectSettler, engine.EffectCantBuild, engine.EffectHasPopAndCanBuild, engine.EffectBuildingRemovesPopulation} {
		if runtimeUnitHasEffect(unit, blocked) {
			return false
		}
	}
	return true
}
