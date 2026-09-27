// Package ai implements a projection-only, deterministic gameplay controller.
package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

var ErrTerminal = errors.New("game is terminal")

// ProjectionSource is the only state input available to an AI controller.
type ProjectionSource interface {
	Projection(context.Context, int64, string) ([]byte, error)
}

// ExplorationStep is accepted self-movement metadata, not tactical world state.
type ExplorationStep struct {
	CommandType       string
	DestinationTileID string
	Sequence          int64
}

// ExplorationHistorySource returns only one AI actor's accepted movement history.
type ExplorationHistorySource interface {
	ExplorationHistory(context.Context, int64, string) ([]ExplorationStep, error)
}

// Submission is an ordinary gameplay command expressed without an authenticated human session.
type Submission struct {
	GameID           int64
	ExpectedRevision int64
	IdempotencyKey   string
	CommandID        string
	ActorPlayerID    string
	Type             string
	Payload          json.RawMessage
}

type SubmissionResult struct {
	Revision int64
	Replayed bool
	Terminal bool
}

// Gateway persists an AI command through the same receipt/revision authority as a human command.
type Gateway interface {
	Submit(context.Context, Submission) (SubmissionResult, error)
}

type Config struct {
	ActionBudget        int
	History             ExplorationHistorySource
	Observe             func(Submission)
	CloseoutAfterAction bool
}

type Candidate struct {
	Type     string
	EntityID string
	TargetID string
	Score    int
	Payload  json.RawMessage
}

type Result struct {
	Commands int
	Revision int64
	Replayed bool
}

type Driver struct {
	projections         ProjectionSource
	gateway             Gateway
	history             ExplorationHistorySource
	budget              int
	observe             func(Submission)
	closeoutAfterAction bool
	mu                  sync.Mutex
	active              map[int64]bool
}

type closeoutProjection struct {
	Revision       int64  `json:"revision"`
	PlanningStatus string `json:"planning_status,omitempty"`
}

func NewDriver(projections ProjectionSource, gateway Gateway, config Config) *Driver {
	budget := config.ActionBudget
	if budget < 1 {
		budget = 1
	}
	return &Driver{projections: projections, gateway: gateway, history: config.History, budget: budget, observe: config.Observe, closeoutAfterAction: config.CloseoutAfterAction, active: make(map[int64]bool)}
}

func (d *Driver) observeSubmission(submission Submission) {
	if d.observe == nil {
		return
	}
	submission.Payload = append(json.RawMessage(nil), submission.Payload...)
	d.observe(submission)
}

// Plan derives candidates solely from a player-visible projection. Candidates are ranked by
// score, command type, source entity ID, and target ID; those identifier tie-breaks are stable.
type plannedRoute struct {
	unitID       string
	cityID       string
	tile         visibility.TileProjection
	path         []string
	stepID       string
	frontier     bool
	coverage     int
	homeDistance int
}

func visitedDestinations(history []ExplorationStep) map[string]bool {
	visited := make(map[string]bool, len(history))
	for _, step := range history {
		if step.CommandType == engine.CommandMoveUnit && step.DestinationTileID != "" {
			visited[step.DestinationTileID] = true
		}
	}
	return visited
}

// selectRoute ranks the current visible step that the selected command will submit.
func selectRoute(routes []plannedRoute, visited map[string]bool) (plannedRoute, bool) {
	if len(routes) == 0 {
		return plannedRoute{}, false
	}
	sorted := append([]plannedRoute(nil), routes...)
	sort.Slice(sorted, func(i, j int) bool {
		leftNovel, rightNovel := !visited[sorted[i].stepID], !visited[sorted[j].stepID]
		if leftNovel != rightNovel {
			return leftNovel
		}
		if sorted[i].frontier != sorted[j].frontier {
			return sorted[i].frontier
		}
		if sorted[i].coverage != sorted[j].coverage {
			return sorted[i].coverage > sorted[j].coverage
		}
		if sorted[i].homeDistance != sorted[j].homeDistance {
			return sorted[i].homeDistance < sorted[j].homeDistance
		}
		if len(sorted[i].path) != len(sorted[j].path) {
			return len(sorted[i].path) < len(sorted[j].path)
		}
		if sorted[i].tile.X != sorted[j].tile.X {
			return sorted[i].tile.X < sorted[j].tile.X
		}
		if sorted[i].tile.Y != sorted[j].tile.Y {
			return sorted[i].tile.Y < sorted[j].tile.Y
		}
		if sorted[i].cityID != sorted[j].cityID {
			return sorted[i].cityID < sorted[j].cityID
		}
		return sorted[i].tile.ID < sorted[j].tile.ID
	})
	return sorted[0], true
}

func Plan(input []byte, playerID string) ([]Candidate, error) {
	return PlanWithHistory(input, playerID, nil)
}

// PlanWithHistory ranks only legal routes from input and uses self-history solely to avoid
// revisiting a visible destination when another visible legal destination exists.
func PlanWithHistory(input []byte, playerID string, history []ExplorationStep) ([]Candidate, error) {
	var projection visibility.Projection
	if err := json.Unmarshal(input, &projection); err != nil {
		return nil, err
	}
	visited := visitedDestinations(history)
	fallbackType := engine.CommandPass
	if projection.PlanningStatus == visibility.PlanningStatusOpen {
		fallbackType = engine.CommandRoundComplete
	}
	candidates := []Candidate{{Type: fallbackType, EntityID: playerID, Score: 100, Payload: json.RawMessage(`{}`)}}
	candidates = append(candidates, researchCandidates(projection, playerID)...)
	candidates = append(candidates, governmentCandidates(projection, playerID)...)
	candidates = append(candidates, projectedScienceRateCandidates(projection)...)
	candidates = append(candidates, projectedReadinessCandidates(projection)...)
	candidates = append(candidates, projectedSpecialistCandidates(projection, playerID)...)
	candidates = append(candidates, projectedTradeRouteCandidates(projection)...)
	candidates = append(candidates, diplomacyCandidates(projection, playerID)...)
	candidates = append(candidates, projectedUnitSpecialTargetCandidates(projection)...)
	candidates = append(candidates, combatCandidates(projection, playerID)...)
	candidates = append(candidates, transportCandidates(projection, playerID)...)
	candidates = append(candidates, terrainImprovementCandidates(projection, playerID)...)
	embassy := embassyCandidates(projection, playerID, visited)
	candidates = append(candidates, embassy...)
	if len(embassy) == 0 {
		candidates = append(candidates, explorationCandidates(projection, playerID, visited)...)
	}
	cityTiles := ownedCityTiles(projection, playerID)
	ownedCityCount := len(cityTiles)
	expansionEnabled := aiExpansionEnabled(projection)
	for _, unit := range projection.Units {
		if unit.OwnerID != playerID {
			continue
		}
		isSettler := hasEffect(unit.Effects, engine.EffectSettler)
		if unit.ArmyID == "" && unit.TransportID == "" && unit.TileID != "" && isSettler && !cityTiles[unit.TileID] && (ownedCityCount == 0 || expansionEnabled && ownedCityCount < 3) {
			candidates = append(candidates, candidate(engine.CommandFoundCity, unit.ID, unit.TileID, 900, map[string]string{"unit_id": unit.ID, "city_id": "ai-city-" + unit.ID, "name": "AI-" + playerID + "-" + unit.ID}))
		}
		if unit.ArmyID == "" && unit.TransportID == "" && unit.TileID != "" && (!isSettler || !expansionEnabled) {
			candidates = append(candidates, candidate(engine.CommandArmyCreate, unit.ID, "", 180, map[string]any{"army_id": "ai-army-" + unit.ID, "unit_ids": []string{unit.ID}}))
		}
		if unit.TransportID == "" && unit.TileID != "" && projectedUnitCanFortify(unit.Effects) {
			candidates = append(candidates, candidate(engine.CommandFortify, unit.ID, unit.TileID, 20, map[string]string{"unit_id": unit.ID}))
		}
	}
	candidates = append(candidates, productionCandidates(projection, playerID)...)
	sort.Slice(candidates, func(i, j int) bool { return Compare(candidates[i], candidates[j]) < 0 })
	return candidates, nil
}

func projectedUnitCanFortify(effects []engine.Effect) bool {
	if hasEffect(effects, engine.EffectCanEntrench) {
		return true
	}
	for _, effect := range effects {
		switch effect.Kind {
		case engine.EffectCantCaptureCity, engine.EffectNoFuelThenCrash, engine.EffectCanPillage, engine.EffectBuildingRemovesPopulation, engine.EffectIgnoreZOC, engine.EffectCanAttackLand, engine.EffectCanAttackSea, engine.EffectCanAttackAir, engine.EffectCanAttackSpace, engine.EffectCanAttackUnderwater, engine.EffectSubmarine, engine.EffectWoodenShip,
			engine.EffectSpecialEmbassy, engine.EffectSpecialSoothsaying, engine.EffectSpecialCauseUnhappiness, engine.EffectSpecialInciteRevolution, engine.EffectSpecialIndulgenceSales, engine.EffectSpecialSlaveUprising, engine.EffectSpecialSteal, engine.EffectSpecialCapture, engine.EffectSpecialTerraform, engine.EffectSpecialTerraformTurns,
			engine.EffectSpecialConvertCity, engine.EffectSpecialCreateFranchise, engine.EffectSpecialInvestigateCity, engine.EffectSpecialPlantNuke, engine.EffectSpecialBioTerror, engine.EffectSpecialNanoTerror, engine.EffectSpecialPlague,
			engine.EffectSpecialAssassinateRuler, engine.EffectSpecialSlaveRaid, engine.EffectSpecialSettlerSlaveRaid, engine.EffectSpecialUndergroundRailway, engine.EffectSpecialInjoin, engine.EffectSpecialReform,
			engine.EffectSpecialCreatePark, engine.EffectSpecialExpel, engine.EffectSpecialSue:
			return false
		}
	}
	return true
}

func aiExpansionEnabled(projection visibility.Projection) bool {
	if projection.Rules == nil {
		return false
	}
	for _, unit := range projection.Rules.Units {
		if unit.ID == "UNIT_SETTLER" {
			return true
		}
	}
	return false
}

func ownedCityTiles(projection visibility.Projection, playerID string) map[string]bool {
	cityTiles := make(map[string]bool)
	for _, city := range projection.Cities {
		if city.OwnerID == playerID && city.TileID != "" {
			cityTiles[city.TileID] = true
		}
	}
	return cityTiles
}

func firstAIListItem(lists []rules.RuntimeAIList, available map[string]bool) (string, bool) {
	for _, list := range lists {
		for _, id := range list.Items {
			if available[id] {
				return id, true
			}
		}
	}
	return "", false
}

func researchCandidates(projection visibility.Projection, playerID string) []Candidate {
	if projection.Research == nil || projection.Research.TargetID != "" || len(projection.Research.Available) == 0 {
		return nil
	}
	available := append([]string(nil), projection.Research.Available...)
	availableSet := make(map[string]bool, len(available))
	for _, advanceID := range available {
		availableSet[advanceID] = true
	}
	if projection.Rules != nil && projection.Rules.AI != nil {
		if advanceID, ok := firstAIListItem(projection.Rules.AI.AdvanceLists, availableSet); ok {
			return []Candidate{candidate(engine.CommandSetResearch, playerID, advanceID, 800, map[string]string{"advance_id": advanceID})}
		}
	}
	costs := map[string]int{}
	if projection.Rules != nil {
		for _, advance := range projection.Rules.Advances {
			if advance.ID != "" && advance.Cost > 0 {
				costs[advance.ID] = advance.Cost
			}
		}
	}
	sort.Slice(available, func(i, j int) bool {
		leftCost, leftOK := costs[available[i]]
		rightCost, rightOK := costs[available[j]]
		if leftOK != rightOK {
			return leftOK
		}
		if leftOK && rightOK && leftCost != rightCost {
			return leftCost < rightCost
		}
		return available[i] < available[j]
	})
	advanceID := available[0]
	return []Candidate{candidate(engine.CommandSetResearch, playerID, advanceID, 800, map[string]string{"advance_id": advanceID})}
}

func governmentCandidates(projection visibility.Projection, playerID string) []Candidate {
	if projection.Diplomacy == nil || projection.Diplomacy.GovernmentID != "" || projection.Research == nil {
		return nil
	}
	owned := ownedAdvances(projection)
	if projection.Rules != nil && len(projection.Rules.Governments) > 0 {
		best, ok := bestProjectedGovernment(projection.Rules.Governments, owned)
		if ok {
			return []Candidate{candidate(engine.CommandSetGovernment, playerID, best.ID, 760, map[string]string{"government_id": best.ID})}
		}
	}
	if !owned["ADVANCE_SYNTHETIC_ROOT"] {
		return nil
	}
	return []Candidate{candidate(engine.CommandSetGovernment, playerID, "GOVERNMENT_SYNTHETIC_REPUBLIC", 760, map[string]string{"government_id": "GOVERNMENT_SYNTHETIC_REPUBLIC"})}
}

func projectedScienceRateCandidates(projection visibility.Projection) []Candidate {
	candidates := make([]Candidate, 0)
	for _, action := range projection.LegalActions {
		if action.Command != engine.CommandSetScienceRate || len(action.Payload) == 0 {
			continue
		}
		var payload struct {
			SciencePercent int `json:"science_percent"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil || payload.SciencePercent != 100 {
			continue
		}
		candidates = append(candidates, Candidate{Type: engine.CommandSetScienceRate, EntityID: action.EntityID, TargetID: action.TargetID, Score: 755, Payload: append(json.RawMessage(nil), action.Payload...)})
	}
	return candidates
}

func projectedTradeRouteCandidates(projection visibility.Projection) []Candidate {
	candidates := make([]Candidate, 0)
	for _, action := range projection.LegalActions {
		if action.Command != engine.CommandCreateTradeRoute || len(action.Payload) == 0 {
			continue
		}
		var payload struct {
			SourceCityID      string `json:"source_city_id"`
			DestinationCityID string `json:"destination_city_id"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil || payload.SourceCityID == "" || payload.DestinationCityID == "" {
			continue
		}
		candidates = append(candidates, Candidate{Type: engine.CommandCreateTradeRoute, EntityID: action.EntityID, TargetID: action.TargetID, Score: 752, Payload: append(json.RawMessage(nil), action.Payload...)})
	}
	return candidates
}

func projectedSpecialistCandidates(projection visibility.Projection, playerID string) []Candidate {
	legalCities := make(map[string]visibility.ActionProjection)
	for _, action := range projection.LegalActions {
		if action.Command == engine.CommandSetSpecialists && action.EntityID != "" {
			legalCities[action.EntityID] = action
		}
	}
	if len(legalCities) == 0 {
		return nil
	}
	candidates := make([]Candidate, 0)
	for _, city := range projection.Cities {
		action, ok := legalCities[city.ID]
		if !ok || city.OwnerID != playerID || city.Happiness <= 0 || city.Happiness >= 73 || city.Population <= 0 {
			continue
		}
		specialists := make(map[string]int, len(city.Specialists)+1)
		total := 0
		for kind, count := range city.Specialists {
			if count <= 0 {
				continue
			}
			specialists[kind] = count
			total += count
		}
		if total >= city.Population {
			continue
		}
		specialists["POP_ENTERTAINER"]++
		payload, err := json.Marshal(map[string]any{"city_id": city.ID, "specialists": specialists})
		if err != nil {
			continue
		}
		candidates = append(candidates, Candidate{Type: engine.CommandSetSpecialists, EntityID: action.EntityID, TargetID: action.TargetID, Score: 753, Payload: payload})
	}
	return candidates
}

func projectedReadinessCandidates(projection visibility.Projection) []Candidate {
	candidates := make([]Candidate, 0)
	for _, action := range projection.LegalActions {
		if action.Command != engine.CommandSetReadiness || len(action.Payload) == 0 {
			continue
		}
		var payload struct {
			Readiness string `json:"readiness"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil || payload.Readiness != "war" {
			continue
		}
		candidates = append(candidates, Candidate{Type: engine.CommandSetReadiness, EntityID: action.EntityID, TargetID: action.TargetID, Score: 754, Payload: append(json.RawMessage(nil), action.Payload...)})
	}
	return candidates
}

func diplomacyCandidates(projection visibility.Projection, playerID string) []Candidate {
	candidates := projectedEstablishContactCandidates(projection)
	if projection.Diplomacy == nil {
		return candidates
	}
	pending := make(map[string]bool, len(projection.Diplomacy.Proposals))
	for _, proposal := range projection.Diplomacy.Proposals {
		if proposal.Kind != "" && proposal.FromPlayerID != "" && proposal.ToPlayerID != "" {
			pending[proposal.Kind+"\x00"+proposal.FromPlayerID+"\x00"+proposal.ToPlayerID] = true
			pending[proposal.Kind+"\x00"+proposal.ToPlayerID+"\x00"+proposal.FromPlayerID] = true
		}
		if proposal.ToPlayerID == playerID {
			candidates = append(candidates, candidate(engine.CommandRespondAgreement, proposal.ID, proposal.Kind, 740, map[string]any{"proposal_id": proposal.ID, "accept": true}))
		}
	}
	for _, relation := range projection.Diplomacy.Relations {
		if relation.TargetPlayerID == "" || relation.Treaty == "union" || pending["union\x00"+playerID+"\x00"+relation.TargetPlayerID] {
			continue
		}
		proposalID := "ai-union-" + playerID + "-" + relation.TargetPlayerID
		candidates = append(candidates, candidate(engine.CommandProposeAgreement, relation.TargetPlayerID, "union", 430, map[string]string{"proposal_id": proposalID, "target_player_id": relation.TargetPlayerID, "kind": "union"}))
	}
	return candidates
}

func projectedUnitSpecialTargetCandidates(projection visibility.Projection) []Candidate {
	unitTargetScores := map[string]int{engine.CommandExpel: 650, engine.CommandSue: 645}
	cityTargetScores := map[string]int{
		engine.CommandUndergroundRailway: 640,
		engine.CommandSlaveRaid:          635,
		engine.CommandSettlerSlaveRaid:   635,
		engine.CommandBioTerror:          630,
		engine.CommandNanoTerror:         630,
		engine.CommandPlague:             630,
		engine.CommandCreateFranchise:    625,
		engine.CommandConvertCity:        620,
		engine.CommandCauseUnhappiness:   618,
		engine.CommandInciteRevolution:   617,
		engine.CommandIndulgenceSales:    616,
		engine.CommandSlaveUprising:      615,
		engine.CommandInvestigateCity:    610,
		engine.CommandAssassinateRuler:   605,
		engine.CommandInjoin:             600,
		engine.CommandReform:             600,
		engine.CommandPlantNuke:          500,
	}
	candidates := make([]Candidate, 0)
	seen := map[string]bool{}
	for _, action := range projection.LegalActions {
		if score, ok := unitTargetScores[action.Command]; ok {
			candidate, ok := projectedUnitTargetActionCandidate(action, score)
			if ok && !seen[candidate.Type+"\x00"+candidate.EntityID+"\x00"+candidate.TargetID] {
				seen[candidate.Type+"\x00"+candidate.EntityID+"\x00"+candidate.TargetID] = true
				candidates = append(candidates, candidate)
			}
			continue
		}
		if score, ok := cityTargetScores[action.Command]; ok {
			candidate, ok := projectedCityTargetActionCandidate(action, score)
			if ok && !seen[candidate.Type+"\x00"+candidate.EntityID+"\x00"+candidate.TargetID] {
				seen[candidate.Type+"\x00"+candidate.EntityID+"\x00"+candidate.TargetID] = true
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

func projectedUnitTargetActionCandidate(action visibility.ActionProjection, score int) (Candidate, bool) {
	if len(action.Payload) == 0 {
		return Candidate{}, false
	}
	var payload struct {
		UnitID       string `json:"unit_id"`
		TargetUnitID string `json:"target_unit_id"`
	}
	if err := json.Unmarshal(action.Payload, &payload); err != nil || payload.UnitID == "" || payload.TargetUnitID == "" {
		return Candidate{}, false
	}
	return Candidate{Type: action.Command, EntityID: payload.UnitID, TargetID: payload.TargetUnitID, Score: score, Payload: append(json.RawMessage(nil), action.Payload...)}, true
}

func projectedCityTargetActionCandidate(action visibility.ActionProjection, score int) (Candidate, bool) {
	if len(action.Payload) == 0 {
		return Candidate{}, false
	}
	var payload struct {
		UnitID string `json:"unit_id"`
		CityID string `json:"city_id"`
	}
	if err := json.Unmarshal(action.Payload, &payload); err != nil || payload.UnitID == "" || payload.CityID == "" {
		return Candidate{}, false
	}
	return Candidate{Type: action.Command, EntityID: payload.UnitID, TargetID: payload.CityID, Score: score, Payload: append(json.RawMessage(nil), action.Payload...)}, true
}

func projectedEstablishContactCandidates(projection visibility.Projection) []Candidate {
	candidates := make([]Candidate, 0)
	seen := map[string]bool{}
	for _, action := range projection.LegalActions {
		if action.Command != engine.CommandEstablishContact || len(action.Payload) == 0 {
			continue
		}
		var payload struct {
			TargetPlayerID string `json:"target_player_id"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil || payload.TargetPlayerID == "" || seen[payload.TargetPlayerID] {
			continue
		}
		seen[payload.TargetPlayerID] = true
		candidates = append(candidates, candidate(engine.CommandEstablishContact, action.EntityID, payload.TargetPlayerID, 735, payload))
	}
	return candidates
}

func combatCandidates(projection visibility.Projection, playerID string) []Candidate {
	tiles := projectionTiles(projection)
	seen := map[string]bool{}
	candidates := make([]Candidate, 0)
	for _, attacker := range projection.Units {
		if attacker.OwnerID != playerID || attacker.ArmyID == "" || attacker.TransportID != "" || attacker.TileID == "" || attacker.Attack <= 0 {
			continue
		}
		attackerTile, ok := tiles[attacker.TileID]
		if !ok || attackerTile.Stale {
			continue
		}
		for _, defender := range projection.Units {
			if defender.OwnerID == playerID || defender.ArmyID == "" || defender.TileID == "" || defender.TransportID != "" {
				continue
			}
			defenderTile, ok := tiles[defender.TileID]
			if !ok || defenderTile.Stale || !adjacent(attackerTile, defenderTile) || !combatDomainsCompatible(attacker, defender) {
				continue
			}
			key := attacker.ArmyID + "\x00" + defender.ArmyID
			if seen[key] {
				continue
			}
			seen[key] = true
			candidates = append(candidates, candidate(engine.CommandCombat, attacker.ArmyID, defender.ArmyID, 700, map[string]string{"attacker_army_id": attacker.ArmyID, "defender_army_id": defender.ArmyID}))
		}
	}
	return candidates
}

func transportCandidates(projection visibility.Projection, playerID string) []Candidate {
	tiles := projectionTiles(projection)
	units := projectionUnits(projection)
	candidates := make([]Candidate, 0)
	for _, unit := range projection.Units {
		if unit.OwnerID != playerID {
			continue
		}
		if unit.TransportID != "" {
			transport, ok := units[unit.TransportID]
			transportTile, tileOK := tiles[transport.TileID]
			if !ok || !tileOK || unit.Movement < 1 {
				continue
			}
			for _, tile := range projection.Tiles {
				if tile.Stale || !adjacent(transportTile, tile) || tile.Planning == nil || !tile.Planning.Accessible || !containsString(tile.Planning.Domains, unit.Domain) || tile.Planning.EnemyZOC {
					continue
				}
				candidates = append(candidates, candidate(engine.CommandDisembark, unit.ID, tile.ID, 600, map[string]string{"unit_id": unit.ID, "tile_id": tile.ID}))
			}
			continue
		}
		if unit.TileID == "" || unit.ArmyID != "" {
			continue
		}
		for _, transport := range projection.Units {
			if transport.OwnerID != playerID || transport.ID == unit.ID || transport.TileID != unit.TileID || transport.TransportID != "" || transport.CargoCapacity <= cargoCount(projection, transport.ID) || !containsString(transport.CargoDomains, unit.Domain) {
				continue
			}
			candidates = append(candidates, candidate(engine.CommandEmbark, unit.ID, transport.ID, 600, map[string]string{"unit_id": unit.ID, "transport_id": transport.ID}))
		}
	}
	return candidates
}

func terrainImprovementCandidates(projection visibility.Projection, playerID string) []Candidate {
	if projection.Diplomacy == nil || projection.Diplomacy.PublicWorks < 1 {
		return nil
	}
	if candidates := projectedTerrainImprovementActionCandidates(projection); len(candidates) > 0 {
		return candidates
	}
	tiles := projectionTiles(projection)
	candidates := make([]Candidate, 0)
	for _, unit := range projection.Units {
		if unit.OwnerID != playerID || unit.TransportID != "" || unit.TileID == "" || unit.Movement < 1 || !hasEffect(unit.Effects, engine.EffectSpecialTerraform) {
			continue
		}
		origin, ok := tiles[unit.TileID]
		if !ok {
			continue
		}
		for _, tile := range projection.Tiles {
			if tile.Stale || !adjacent(origin, tile) {
				continue
			}
			if tile.Pollution <= 0 {
				continue
			}
			candidates = append(candidates, candidate(engine.CommandBuildImprovement, unit.ID, tile.ID, 620, map[string]string{"unit_id": unit.ID, "tile_id": tile.ID, "improvement_id": "TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP"}))
		}
	}
	return candidates
}

func projectedTerrainImprovementActionCandidates(projection visibility.Projection) []Candidate {
	if len(projection.LegalActions) == 0 {
		return nil
	}
	ranks := retailImprovementRanks(projection)
	tiles := projectionTiles(projection)
	improvements := projectedTerrainImprovements(projection)
	candidates := make([]Candidate, 0)
	for _, action := range projection.LegalActions {
		if action.Command != engine.CommandBuildImprovement || len(action.Payload) == 0 {
			continue
		}
		var payload struct {
			UnitID        string `json:"unit_id"`
			TileID        string `json:"tile_id"`
			ImprovementID string `json:"improvement_id"`
		}
		if err := json.Unmarshal(action.Payload, &payload); err != nil || payload.UnitID == "" || payload.TileID == "" || payload.ImprovementID == "" {
			continue
		}
		improvement := improvements[payload.ImprovementID]
		if improvement.PollutionDelta < 0 {
			if tile, ok := tiles[payload.TileID]; ok && tile.Pollution <= 0 {
				continue
			}
		}
		rank, listed := ranks[payload.ImprovementID]
		if !listed {
			rank = len(ranks) + 100
		}
		candidates = append(candidates, candidate(engine.CommandBuildImprovement, payload.UnitID, payload.TileID, 620-rank, payload))
	}
	return candidates
}

func projectedTerrainImprovements(projection visibility.Projection) map[string]rules.RuntimeTerrainImprovementDefinition {
	result := map[string]rules.RuntimeTerrainImprovementDefinition{}
	if projection.Rules == nil {
		return result
	}
	for _, improvement := range projection.Rules.TerrainImprovements {
		result[improvement.ID] = improvement
	}
	return result
}

func retailImprovementRanks(projection visibility.Projection) map[string]int {
	ranks := map[string]int{}
	if projection.Rules == nil || projection.Rules.AI == nil {
		return ranks
	}
	rank := 0
	for _, list := range projection.Rules.AI.ImprovementLists {
		for _, id := range list.Items {
			if _, seen := ranks[id]; seen {
				continue
			}
			ranks[id] = rank
			rank++
		}
	}
	return ranks
}

func productionCandidates(projection visibility.Projection, playerID string) []Candidate {
	candidates := make([]Candidate, 0)
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.Stale || len(city.Queue) != 0 {
			continue
		}
		item, ok := engine.ProductionItem{}, false
		if shouldProduceEmergencyMilitaryUnit(projection, playerID) {
			item, ok = bestProjectedCombatUnitProduction(projection)
		}
		if !ok && shouldProduceExpansionSettler(projection, playerID) {
			item, ok = bestProjectedSettlerProduction(projection)
		}
		if !ok && shouldProduceMilitaryUnit(projection, playerID) {
			item, ok = bestProjectedCombatUnitProduction(projection)
		}
		if !ok {
			item, ok = bestProjectedCityProduction(projection, city)
		}
		if !ok {
			item, ok = bestProjectedUnitProduction(projection)
		}
		if ok && city.Population <= 1 && productionItemRemovesPopulation(item) {
			item, ok = bestProjectedCityProduction(projection, city)
		}
		if !ok {
			item = engine.ProductionItem{ID: "ai-unit-" + city.ID, Kind: engine.ProductionUnit, Cost: 1}
		}
		if city.Population <= 1 && productionItemRemovesPopulation(item) {
			continue
		}
		candidates = append(candidates, candidate(engine.CommandQueueProduction, city.ID, item.ID, 90, map[string]any{"city_id": city.ID, "item": item}))
	}
	return candidates
}

func productionItemRemovesPopulation(item engine.ProductionItem) bool {
	for _, effect := range item.Effects {
		if effect.Kind == engine.EffectBuildingRemovesPopulation {
			return true
		}
	}
	return false
}

func ownedAdvances(projection visibility.Projection) map[string]bool {
	owned := map[string]bool{}
	if projection.Research == nil {
		return owned
	}
	for _, advance := range projection.Research.Advances {
		owned[advance] = true
	}
	return owned
}

func bestProjectedGovernment(governments []rules.RuntimeGovernmentDefinition, owned map[string]bool) (rules.RuntimeGovernmentDefinition, bool) {
	var best rules.RuntimeGovernmentDefinition
	bestScore := -1
	for _, government := range governments {
		if government.ID == "" || (government.EnableAdvanceID != "" && !owned[government.EnableAdvanceID]) {
			continue
		}
		score := government.ResearchBonus + government.PublicWorksBonus + government.ProductionPercent + government.GoldPercent
		if bestScore == -1 || score > bestScore || score == bestScore && government.ID < best.ID {
			best, bestScore = government, score
		}
	}
	return best, bestScore >= 0
}

func shouldProduceEmergencyMilitaryUnit(projection visibility.Projection, playerID string) bool {
	if !aiExpansionEnabled(projection) {
		return false
	}
	populatedCity := false
	for _, city := range projection.Cities {
		if city.OwnerID == playerID && !city.Stale && city.Population > 0 {
			populatedCity = true
			break
		}
	}
	if !populatedCity {
		return false
	}
	for _, unit := range projection.Units {
		if unit.OwnerID == playerID && !hasEffect(unit.Effects, engine.EffectSettler) {
			return false
		}
	}
	return true
}

func shouldProduceMilitaryUnit(projection visibility.Projection, playerID string) bool {
	if !aiExpansionEnabled(projection) {
		return false
	}
	ownedCities := 0
	for _, city := range projection.Cities {
		if city.OwnerID == playerID && !city.Stale {
			ownedCities++
		}
	}
	if ownedCities == 0 {
		return false
	}
	ownedUnits := 0
	for _, unit := range projection.Units {
		if unit.OwnerID == playerID && !hasEffect(unit.Effects, engine.EffectSettler) {
			ownedUnits++
		}
	}
	if ownedUnits == 0 {
		return true
	}
	if ownedCities < 3 {
		return false
	}
	return ownedUnits < ownedCities
}

func shouldProduceExpansionSettler(projection visibility.Projection, playerID string) bool {
	if !aiExpansionEnabled(projection) {
		return false
	}
	ownedCities := 0
	for _, city := range projection.Cities {
		if city.OwnerID == playerID && !city.Stale {
			ownedCities++
		}
	}
	if ownedCities == 0 || ownedCities >= 3 {
		return false
	}
	for _, unit := range projection.Units {
		if unit.OwnerID == playerID && hasEffect(unit.Effects, engine.EffectSettler) {
			return false
		}
	}
	return true
}

func bestProjectedSettlerProduction(projection visibility.Projection) (engine.ProductionItem, bool) {
	if projection.Rules == nil || len(projection.Rules.Units) == 0 {
		return engine.ProductionItem{}, false
	}
	owned := ownedAdvances(projection)
	if item, ok := bestAIListedUnitProduction(projection, owned, func(unit rules.RuntimeUnitDefinition) bool {
		return runtimeUnitHasEffect(unit, engine.EffectSettler)
	}); ok {
		return item, true
	}
	var best rules.RuntimeUnitDefinition
	bestCost := 0
	for _, unit := range projection.Rules.Units {
		if unit.ID == "" || !runtimeUnitAvailable(unit, owned) || !runtimeUnitHasEffect(unit, engine.EffectSettler) {
			continue
		}
		cost := productionCost(unit)
		if best.ID == "" || cost < bestCost || cost == bestCost && unit.ID < best.ID {
			best, bestCost = unit, cost
		}
	}
	if best.ID == "" {
		return engine.ProductionItem{}, false
	}
	return engine.ProductionItem{ID: best.ID, Kind: engine.ProductionUnit, Cost: bestCost, Upkeep: best.Upkeep, Effects: productionEffects(best)}, true
}

func runtimeUnitHasEffect(unit rules.RuntimeUnitDefinition, kind string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func bestProjectedCityProduction(projection visibility.Projection, city visibility.CityProjection) (engine.ProductionItem, bool) {
	if projection.Rules == nil || len(projection.Rules.Production) == 0 {
		return engine.ProductionItem{}, false
	}
	owned := ownedAdvances(projection)
	built := make(map[string]bool, len(city.Buildings))
	for _, building := range city.Buildings {
		built[building.ID] = true
	}
	if item, ok := bestAIListedBuildingProduction(projection, city, owned, built); ok {
		return item, true
	}
	var best rules.RuntimeProduction
	bestScore := -1
	for _, production := range projection.Rules.Production {
		if production.ID == "" || production.Kind != engine.ProductionBuilding || built[production.ID] || !runtimeProductionAvailable(production, owned, built) || !runtimeProductionPlacementAvailable(projection, city, production) {
			continue
		}
		score := projectedRuntimeProductionScore(production)
		if bestScore == -1 || score > bestScore || score == bestScore && production.Cost < best.Cost || score == bestScore && production.Cost == best.Cost && production.ID < best.ID {
			best, bestScore = production, score
		}
	}
	if bestScore < 0 {
		return engine.ProductionItem{}, false
	}
	return productionItemFromRuntime(best), true
}

func bestAIListedUnitProduction(projection visibility.Projection, owned map[string]bool, include func(rules.RuntimeUnitDefinition) bool) (engine.ProductionItem, bool) {
	if projection.Rules == nil || projection.Rules.AI == nil || len(projection.Rules.AI.UnitBuildLists) == 0 {
		return engine.ProductionItem{}, false
	}
	unitsByID := make(map[string]rules.RuntimeUnitDefinition, len(projection.Rules.Units))
	for _, unit := range projection.Rules.Units {
		if unit.ID != "" && unit.ID != "UNIT_SYNTHETIC" {
			unitsByID[unit.ID] = unit
		}
	}
	for _, list := range projection.Rules.AI.UnitBuildLists {
		for _, id := range list.Items {
			unit, ok := unitsByID[id]
			if !ok || !include(unit) || !runtimeUnitAvailable(unit, owned) {
				continue
			}
			return engine.ProductionItem{ID: unit.ID, Kind: engine.ProductionUnit, Cost: productionCost(unit), Upkeep: unit.Upkeep, Effects: productionEffects(unit)}, true
		}
	}
	return engine.ProductionItem{}, false
}

func bestAIListedBuildingProduction(projection visibility.Projection, city visibility.CityProjection, owned map[string]bool, built map[string]bool) (engine.ProductionItem, bool) {
	if projection.Rules == nil || projection.Rules.AI == nil || len(projection.Rules.AI.BuildingBuildLists) == 0 {
		return engine.ProductionItem{}, false
	}
	productionByID := make(map[string]rules.RuntimeProduction, len(projection.Rules.Production))
	for _, production := range projection.Rules.Production {
		if production.ID != "" && production.Kind == engine.ProductionBuilding {
			productionByID[production.ID] = production
		}
	}
	for _, list := range projection.Rules.AI.BuildingBuildLists {
		for _, id := range list.Items {
			production, ok := productionByID[id]
			if !ok || built[id] || !runtimeProductionAvailable(production, owned, built) || !runtimeProductionPlacementAvailable(projection, city, production) {
				continue
			}
			return productionItemFromRuntime(production), true
		}
	}
	return engine.ProductionItem{}, false
}

func runtimeProductionPlacementAvailable(projection visibility.Projection, city visibility.CityProjection, production rules.RuntimeProduction) bool {
	if !production.CantBuildInSea && !production.CoastalOnly {
		return true
	}
	cityTile, ok := projectedTileByID(projection, city.TileID)
	if !ok {
		return false
	}
	isSeaCity := projectedTerrainIsSea(projection, cityTile.Terrain)
	if production.CantBuildInSea && isSeaCity {
		return false
	}
	if !production.CoastalOnly || isSeaCity {
		return true
	}
	for _, tile := range projection.Tiles {
		if adjacent(cityTile, tile) && projectedTerrainIsSea(projection, tile.Terrain) {
			return true
		}
	}
	return false
}

func projectedTileByID(projection visibility.Projection, id string) (visibility.TileProjection, bool) {
	for _, tile := range projection.Tiles {
		if tile.ID == id {
			return tile, true
		}
	}
	return visibility.TileProjection{}, false
}

func projectedTerrainIsSea(projection visibility.Projection, terrainID string) bool {
	if terrainID == engine.TerrainOcean {
		return true
	}
	if projection.Rules == nil || projection.Rules.Movement == nil {
		return false
	}
	terrain, ok := projectedRuntimeTerrain(projection.Rules.Movement.Terrains, terrainID)
	return ok && containsString(terrain.Domains, engine.DomainSea) && !containsString(terrain.Domains, engine.DomainLand)
}

func projectedRuntimeTerrain(terrains []rules.RuntimeTerrain, id string) (rules.RuntimeTerrain, bool) {
	for _, terrain := range terrains {
		if terrain.ID == id {
			return terrain, true
		}
	}
	return rules.RuntimeTerrain{}, false
}

func runtimeProductionAvailable(production rules.RuntimeProduction, owned map[string]bool, built map[string]bool) bool {
	for _, prerequisite := range production.Prerequisites {
		if !owned[prerequisite] {
			return false
		}
	}
	for _, requiredBuilding := range production.RequiredBuildings {
		if !built[requiredBuilding] {
			return false
		}
	}
	for _, obsolete := range production.ObsoleteBy {
		if owned[obsolete] {
			return false
		}
	}
	return true
}

func projectedRuntimeProductionScore(production rules.RuntimeProduction) int {
	score := 0
	for _, effect := range production.Effects {
		score += effect.Value
	}
	if score == 0 {
		score = 1
	}
	return score
}

func productionItemFromRuntime(production rules.RuntimeProduction) engine.ProductionItem {
	item := engine.ProductionItem{ID: production.ID, Kind: production.Kind, Cost: production.Cost, Upkeep: production.Upkeep, Effects: runtimeProductionEffects(production)}
	if len(production.Prerequisites) > 0 {
		item.Prerequisite = production.Prerequisites[0]
	}
	if len(production.ObsoleteBy) > 0 {
		item.ObsoletedBy = production.ObsoleteBy[0]
	}
	return item
}

func runtimeProductionEffects(production rules.RuntimeProduction) []engine.Effect {
	effects := make([]engine.Effect, 0, len(production.Effects))
	for _, effect := range production.Effects {
		effects = append(effects, engine.Effect{ID: effect.ID, Kind: effect.Kind, Value: effect.Value})
	}
	return effects
}

func bestProjectedCombatUnitProduction(projection visibility.Projection) (engine.ProductionItem, bool) {
	return bestProjectedUnitProductionMatching(projection, func(unit rules.RuntimeUnitDefinition) bool {
		return !runtimeUnitHasEffect(unit, engine.EffectSettler) && (unit.Attack > 0 || unit.Defense > 0)
	})
}

func bestProjectedUnitProduction(projection visibility.Projection) (engine.ProductionItem, bool) {
	return bestProjectedUnitProductionMatching(projection, func(rules.RuntimeUnitDefinition) bool { return true })
}

func bestProjectedUnitProductionMatching(projection visibility.Projection, include func(rules.RuntimeUnitDefinition) bool) (engine.ProductionItem, bool) {
	if projection.Rules == nil || len(projection.Rules.Units) == 0 {
		return engine.ProductionItem{}, false
	}
	owned := ownedAdvances(projection)
	if item, ok := bestAIListedUnitProduction(projection, owned, include); ok {
		return item, true
	}
	var best rules.RuntimeUnitDefinition
	bestScore := -1
	for _, unit := range projection.Rules.Units {
		if unit.ID == "" || unit.ID == "UNIT_SYNTHETIC" || !include(unit) || !runtimeUnitAvailable(unit, owned) {
			continue
		}
		score := projectedUnitProductionScore(unit)
		if bestScore == -1 || score > bestScore || score == bestScore && productionCost(unit) < productionCost(best) || score == bestScore && productionCost(unit) == productionCost(best) && unit.ID < best.ID {
			best, bestScore = unit, score
		}
	}
	if bestScore < 0 {
		return engine.ProductionItem{}, false
	}
	return engine.ProductionItem{ID: best.ID, Kind: engine.ProductionUnit, Cost: productionCost(best), Upkeep: best.Upkeep, Effects: productionEffects(best)}, true
}

func runtimeUnitAvailable(unit rules.RuntimeUnitDefinition, owned map[string]bool) bool {
	for _, prerequisite := range unit.Prerequisites {
		if !owned[prerequisite] {
			return false
		}
	}
	for _, obsolete := range unit.ObsoleteBy {
		if owned[obsolete] {
			return false
		}
	}
	return true
}

func projectedUnitProductionScore(unit rules.RuntimeUnitDefinition) int {
	score := unit.Attack*4 + unit.Defense*3 + unit.Health + unit.TransportCapacity*2 + unit.VisibilityRange
	for _, effect := range unit.Effects {
		if effect.Kind == engine.EffectSettler {
			score += 50
		}
	}
	return score
}

func productionCost(unit rules.RuntimeUnitDefinition) int {
	if unit.ProductionCost > 0 {
		return unit.ProductionCost
	}
	return 1
}

func productionEffects(unit rules.RuntimeUnitDefinition) []engine.Effect {
	effects := make([]engine.Effect, 0, len(unit.Effects))
	for _, effect := range unit.Effects {
		effects = append(effects, engine.Effect{ID: effect.ID, Kind: effect.Kind, Value: effect.Value})
	}
	return effects
}

func projectionTiles(projection visibility.Projection) map[string]visibility.TileProjection {
	tiles := make(map[string]visibility.TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		tiles[tile.ID] = tile
	}
	return tiles
}

func projectionUnits(projection visibility.Projection) map[string]visibility.UnitProjection {
	units := make(map[string]visibility.UnitProjection, len(projection.Units))
	for _, unit := range projection.Units {
		units[unit.ID] = unit
	}
	return units
}

func cargoCount(projection visibility.Projection, transportID string) int {
	count := 0
	for _, unit := range projection.Units {
		if unit.TransportID == transportID {
			count++
		}
	}
	return count
}

func combatDomainsCompatible(attacker, defender visibility.UnitProjection) bool {
	return attacker.Domain == "" || defender.Domain == "" || attacker.Domain == defender.Domain || attacker.Domain == engine.DomainAmphibious || defender.Domain == engine.DomainAmphibious
}

func embassyCandidates(projection visibility.Projection, playerID string, visited map[string]bool) []Candidate {
	knownEmbassies := make(map[string]bool, len(projection.EmbassyPlayerIDs))
	for _, embassyPlayerID := range projection.EmbassyPlayerIDs {
		knownEmbassies[embassyPlayerID] = true
	}
	tiles := make(map[string]visibility.TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		tiles[tile.ID] = tile
	}
	candidates := make([]Candidate, 0)
	for _, unit := range projection.Units {
		if unit.OwnerID != playerID || unit.Movement < 1 || !hasEffect(unit.Effects, engine.EffectSpecialEmbassy) {
			continue
		}
		unitTile, ok := tiles[unit.TileID]
		if !ok {
			continue
		}
		for _, city := range projection.Cities {
			if city.Stale || city.OwnerID == playerID || knownEmbassies[city.OwnerID] {
				continue
			}
			cityTile, ok := tiles[city.TileID]
			if !ok {
				continue
			}
			if adjacent(unitTile, cityTile) {
				candidates = append(candidates, candidate(engine.CommandEstablishEmbassy, unit.ID, city.ID, 500, map[string]string{"unit_id": unit.ID, "city_id": city.ID}))
				continue
			}
			routes := make([]plannedRoute, 0)
			for _, target := range projection.Tiles {
				if target.Stale || !adjacent(target, cityTile) {
					continue
				}
				path, err := visibility.Plan(projection, unit.ID, target.ID)
				if err != nil || len(path) == 0 {
					continue
				}
				step, ok := projectionTile(projection, path[0])
				if !ok {
					continue
				}
				routes = append(routes, plannedRoute{unitID: unit.ID, cityID: city.ID, tile: step, path: path, stepID: path[0], frontier: visibleFrontier(projection, step)})
			}
			if route, ok := selectRoute(routes, visited); ok {
				candidates = append(candidates, candidate(engine.CommandMoveUnit, route.unitID, route.tile.ID, 400, map[string]any{"unit_id": route.unitID, "path": route.path[:1]}))
			}
		}
	}
	return candidates
}

func projectionTile(projection visibility.Projection, tileID string) (visibility.TileProjection, bool) {
	for _, tile := range projection.Tiles {
		if tile.ID == tileID {
			return tile, true
		}
	}
	return visibility.TileProjection{}, false
}

func explorationCandidates(projection visibility.Projection, playerID string, visited map[string]bool) []Candidate {
	pursuitCities := unresolvedRelationshipTiles(projection, playerID)
	resolvedCities := resolvedRelationshipTiles(projection, playerID)
	home, hasHomeCity := ownVisibleCityTile(projection, playerID)
	candidates := make([]Candidate, 0, 1)
	for _, unit := range projection.Units {
		if unit.OwnerID != playerID || unit.Movement < 1 {
			continue
		}
		homeY, hasHome := visibleCityY(projection, playerID)
		routes := make([]plannedRoute, 0)
		for _, tile := range projection.Tiles {
			if tile.Stale || tile.ID == unit.TileID {
				continue
			}
			path, err := visibility.Plan(projection, unit.ID, tile.ID)
			if err != nil || len(path) == 0 {
				continue
			}
			step, ok := projectionTile(projection, path[0])
			if !ok {
				continue
			}
			homeDistance := 0
			if hasHome {
				homeDistance = abs(step.Y - homeY)
			}
			coverage := pursuitCoverage(step, pursuitCities, home, hasHomeCity, resolvedCities)
			if len(pursuitCities) == 0 && explorationLane(playerID) < 0 {
				coverage = -coverage
			}
			routes = append(routes, plannedRoute{unitID: unit.ID, tile: step, path: path, stepID: path[0], frontier: len(pursuitCities) == 0 && visibleFrontier(projection, step), coverage: coverage, homeDistance: homeDistance})
		}
		if len(pursuitCities) > 0 {
			improving := make([]plannedRoute, 0, len(routes))
			for _, route := range routes {
				if firstHopImproves(projection, unit.TileID, route.path[0], pursuitCities) {
					improving = append(improving, route)
				}
			}
			if len(improving) > 0 {
				routes = improving
			}
		}
		if route, ok := selectRoute(routes, visited); ok {
			candidates = append(candidates, candidate(engine.CommandMoveUnit, route.unitID, route.tile.ID, 150, map[string]any{"unit_id": route.unitID, "path": route.path[:1]}))
		}
	}
	return candidates
}

func firstHopImproves(projection visibility.Projection, originID, nextID string, cities []visibility.TileProjection) bool {
	var origin, next visibility.TileProjection
	foundOrigin, foundNext := false, false
	for _, tile := range projection.Tiles {
		if tile.ID == originID {
			origin, foundOrigin = tile, true
		}
		if tile.ID == nextID {
			next, foundNext = tile, true
		}
	}
	if !foundOrigin || !foundNext {
		return false
	}
	return pursuitCoverage(next, cities, visibility.TileProjection{}, false, nil) > pursuitCoverage(origin, cities, visibility.TileProjection{}, false, nil)
}

func visibleFrontier(projection visibility.Projection, tile visibility.TileProjection) bool {
	visible := make(map[[2]int]bool, len(projection.Tiles))
	for _, candidate := range projection.Tiles {
		if !candidate.Stale {
			visible[[2]int{candidate.X, candidate.Y}] = true
		}
	}
	for _, offset := range [][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}} {
		if !visible[[2]int{tile.X + offset[0], tile.Y + offset[1]}] {
			return true
		}
	}
	return false
}

func visibleCityY(projection visibility.Projection, playerID string) (int, bool) {
	tiles := make(map[string]visibility.TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		if !tile.Stale {
			tiles[tile.ID] = tile
		}
	}
	cityID, cityY := "", 0
	for _, city := range projection.Cities {
		tile, visible := tiles[city.TileID]
		if city.OwnerID != playerID || city.Stale || !visible || (cityID != "" && city.ID >= cityID) {
			continue
		}
		cityID, cityY = city.ID, tile.Y
	}
	return cityY, cityID != ""
}

func resolvedRelationshipTiles(projection visibility.Projection, playerID string) []visibility.TileProjection {
	knownEmbassies := make(map[string]bool, len(projection.EmbassyPlayerIDs))
	for _, player := range projection.EmbassyPlayerIDs {
		knownEmbassies[player] = true
	}
	tiles := make(map[string]visibility.TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		tiles[tile.ID] = tile
	}
	result := make([]visibility.TileProjection, 0)
	for _, city := range projection.Cities {
		if city.OwnerID == playerID || !knownEmbassies[city.OwnerID] {
			continue
		}
		if tile, ok := tiles[city.TileID]; ok {
			result = append(result, tile)
		}
	}
	return result
}

func unresolvedRelationshipTiles(projection visibility.Projection, playerID string) []visibility.TileProjection {
	knownEmbassies := make(map[string]bool, len(projection.EmbassyPlayerIDs))
	for _, player := range projection.EmbassyPlayerIDs {
		knownEmbassies[player] = true
	}
	tiles := make(map[string]visibility.TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		tiles[tile.ID] = tile
	}
	result := make([]visibility.TileProjection, 0)
	for _, city := range projection.Cities {
		if city.OwnerID == playerID || knownEmbassies[city.OwnerID] {
			continue
		}
		if tile, ok := tiles[city.TileID]; ok {
			result = append(result, tile)
		}
	}
	return result
}

func resolvedCoverage(tile, home visibility.TileProjection, hasHome bool, resolvedCities []visibility.TileProjection) int {
	coverage := 0
	for _, city := range resolvedCities {
		if hasHome {
			coverage += (tile.X-home.X)*(home.X-city.X) + (tile.Y-home.Y)*(home.Y-city.Y)
			continue
		}
		coverage += abs(tile.X-city.X) + abs(tile.Y-city.Y)
	}
	return coverage
}

func pursuitCoverage(tile visibility.TileProjection, pursuitCities []visibility.TileProjection, home visibility.TileProjection, hasHome bool, resolvedCities []visibility.TileProjection) int {
	if len(pursuitCities) == 0 {
		return resolvedCoverage(tile, home, hasHome, resolvedCities)
	}
	distance := abs(tile.X-pursuitCities[0].X) + abs(tile.Y-pursuitCities[0].Y)
	for _, city := range pursuitCities[1:] {
		candidate := abs(tile.X-city.X) + abs(tile.Y-city.Y)
		if candidate < distance {
			distance = candidate
		}
	}
	return -distance
}

func ownVisibleCityTile(projection visibility.Projection, playerID string) (visibility.TileProjection, bool) {
	tiles := make(map[string]visibility.TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		tiles[tile.ID] = tile
	}
	cityID := ""
	var result visibility.TileProjection
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || (cityID != "" && city.ID >= cityID) {
			continue
		}
		if tile, ok := tiles[city.TileID]; ok {
			cityID, result = city.ID, tile
		}
	}
	return result, cityID != ""
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func explorationLane(playerID string) int {
	value := 0
	for _, character := range playerID {
		value += int(character)
	}
	if value%2 == 0 {
		return 1
	}
	return -1
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func hasEffect(effects []engine.Effect, kind string) bool {
	for _, effect := range effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func adjacent(left, right visibility.TileProjection) bool {
	dx, dy := left.X-right.X, left.Y-right.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx+dy == 1
}

func candidate(kind, entityID, targetID string, score int, payload any) Candidate {
	bytes, _ := json.Marshal(payload)
	return Candidate{Type: kind, EntityID: entityID, TargetID: targetID, Score: score, Payload: bytes}
}

// Compare returns the documented deterministic ranking: higher score first, then type/entity/target IDs.
func Compare(left, right Candidate) int {
	if left.Score != right.Score {
		return right.Score - left.Score
	}
	if left.Type != right.Type {
		if left.Type < right.Type {
			return -1
		}
		return 1
	}
	if left.EntityID != right.EntityID {
		if left.EntityID < right.EntityID {
			return -1
		}
		return 1
	}
	if left.TargetID < right.TargetID {
		return -1
	}
	if left.TargetID > right.TargetID {
		return 1
	}
	return 0
}

// Drive repeatedly projects and selects ordinary commands until the planner completes its turn or
// reaches its action budget. Stable receipt keys make post-commit scheduling and restart replay safe.
func (d *Driver) submitCandidate(ctx context.Context, gameID int64, playerID string, revision int64, candidate Candidate) (SubmissionResult, error) {
	request := Submission{GameID: gameID, ExpectedRevision: revision, ActorPlayerID: playerID, Type: candidate.Type, Payload: candidate.Payload}
	request.IdempotencyKey = receiptKey(gameID, playerID, revision, candidate)
	request.CommandID = "ai-" + request.IdempotencyKey
	submitted, err := d.gateway.Submit(ctx, request)
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("submit %s for %s/%s with payload %s: %w", candidate.Type, candidate.EntityID, candidate.TargetID, candidate.Payload, err)
	}
	d.observeSubmission(request)
	return submitted, nil
}

func planningCloseout(input []byte) (closeoutProjection, error) {
	var projection closeoutProjection
	if err := json.Unmarshal(input, &projection); err != nil {
		return closeoutProjection{}, err
	}
	return projection, nil
}

func (d *Driver) Drive(ctx context.Context, gameID int64, playerID string, revision int64) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if d == nil || d.projections == nil || d.gateway == nil {
		return Result{}, errors.New("AI driver is not configured")
	}
	d.mu.Lock()
	if d.active[gameID] {
		d.mu.Unlock()
		return Result{Revision: revision}, nil
	}
	d.active[gameID] = true
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		delete(d.active, gameID)
		d.mu.Unlock()
	}()

	result := Result{Revision: revision}
	for action := 0; action < d.budget; action++ {
		input, err := d.projections.Projection(ctx, gameID, playerID)
		if err != nil {
			return Result{}, err
		}
		projection, err := planningCloseout(input)
		if err != nil {
			return Result{}, err
		}
		if projection.Revision != 0 {
			result.Revision = projection.Revision
		}
		switch projection.PlanningStatus {
		case visibility.PlanningStatusClosed:
			return result, nil
		case visibility.PlanningStatusQueued:
			selected := Candidate{Type: engine.CommandRoundComplete, EntityID: playerID, Payload: json.RawMessage(`{}`)}
			submitted, err := d.submitCandidate(ctx, gameID, playerID, result.Revision, selected)
			if err != nil {
				return Result{}, err
			}
			result.Commands++
			result.Revision = submitted.Revision
			result.Replayed = result.Replayed || submitted.Replayed
			if submitted.Terminal {
				return result, ErrTerminal
			}
			return result, nil
		}
		var history []ExplorationStep
		if d.history != nil {
			history, err = d.history.ExplorationHistory(ctx, gameID, playerID)
			if err != nil {
				return Result{}, err
			}
		}
		candidates, err := PlanWithHistory(input, playerID, history)
		if err != nil {
			return Result{}, err
		}
		if len(candidates) == 0 {
			return Result{}, errors.New("AI planner returned no pass candidate")
		}
		selected := candidates[0]
		submitted, err := d.submitCandidate(ctx, gameID, playerID, result.Revision, selected)
		if err != nil {
			return Result{}, err
		}
		result.Commands++
		result.Revision = submitted.Revision
		result.Replayed = result.Replayed || submitted.Replayed
		if submitted.Terminal {
			return result, ErrTerminal
		}
		if selected.Type == engine.CommandPass || selected.Type == engine.CommandRoundComplete {
			break
		}
		if d.closeoutAfterAction {
			closeout, ok := closeoutCandidate(candidates)
			if !ok {
				return Result{}, errors.New("AI planner returned no closeout candidate")
			}
			submitted, err := d.submitCandidate(ctx, gameID, playerID, result.Revision, closeout)
			if err != nil {
				return Result{}, err
			}
			result.Commands++
			result.Revision = submitted.Revision
			result.Replayed = result.Replayed || submitted.Replayed
			if submitted.Terminal {
				return result, ErrTerminal
			}
			break
		}
	}
	return result, nil
}

func closeoutCandidate(candidates []Candidate) (Candidate, bool) {
	for _, candidate := range candidates {
		if candidate.Type == engine.CommandPass || candidate.Type == engine.CommandRoundComplete {
			return candidate, true
		}
	}
	return Candidate{}, false
}

// Pass submits the AI controller's ordinary round-complete command through the same gateway.
func (d *Driver) Pass(ctx context.Context, gameID int64, playerID string, revision int64) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if d == nil || d.gateway == nil {
		return Result{}, errors.New("AI driver is not configured")
	}
	candidate := Candidate{Type: engine.CommandRoundComplete, EntityID: playerID, Payload: json.RawMessage(`{}`)}
	submitted, err := d.submitCandidate(ctx, gameID, playerID, revision, candidate)
	if err != nil {
		return Result{}, err
	}
	return Result{Commands: 1, Revision: submitted.Revision, Replayed: submitted.Replayed}, nil
}

func receiptKey(gameID int64, playerID string, revision int64, candidate Candidate) string {
	bytes, _ := json.Marshal(struct {
		GameID    int64
		PlayerID  string
		Revision  int64
		Candidate Candidate
	}{gameID, playerID, revision, candidate})
	digest := sha256.Sum256(bytes)
	return "ai-" + hex.EncodeToString(digest[:])
}
