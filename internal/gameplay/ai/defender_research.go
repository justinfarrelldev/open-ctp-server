package ai

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
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
			if !runtimeUnitAvailable(unit, known) || unit.Attack <= 0 && unit.Defense <= 0 || unit.Domain != engine.DomainLand && unit.Domain != engine.DomainAmphibious || runtimeUnitHasEffect(unit, engine.EffectSettler) || runtimeUnitHasEffect(unit, engine.EffectCantBuild) || runtimeUnitHasEffect(unit, engine.EffectHasPopAndCanBuild) || unit.NeedsCityGoodID != "" || unit.RequiredGovernmentID != "" || len(unit.RequiredGovernmentIDs) > 0 {
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
// available advance away. It uses only projected catalogs, goods and own state.
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
			if unit.Domain != engine.DomainLand && unit.Domain != engine.DomainAmphibious || unit.Attack <= 0 && unit.Defense <= 0 || runtimeUnitHasEffect(unit, engine.EffectSettler) || runtimeUnitHasEffect(unit, engine.EffectCantBuild) || runtimeUnitHasEffect(unit, engine.EffectHasPopAndCanBuild) || !runtimeUnitAvailable(unit, future) {
				continue
			}
			if unit.NeedsCityGoodID != "" || unit.RequiredGovernmentID != "" || len(unit.RequiredGovernmentIDs) > 0 {
				continue
			}
			production, research := productionCost(unit), costs[advance]
			if bestAdvance == "" || production < bestProduction || production == bestProduction && (research < bestResearch || research == bestResearch && (advance < bestAdvance || advance == bestAdvance && unit.ID < bestUnit)) {
				bestAdvance, bestUnit, bestProduction, bestResearch = advance, unit.ID, production, research
			}
		}
	}
	return bestAdvance, bestAdvance != ""
}
