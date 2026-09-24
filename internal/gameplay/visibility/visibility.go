// Package visibility derives deterministic player-visible state and persistent fog memory.
package visibility

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const defaultSightRadius = 2

const (
	PlanningStatusOpen   = "open"
	PlanningStatusQueued = "queued"
	PlanningStatusClosed = "closed"
)

type TileProjection struct {
	ID        string        `json:"id"`
	X         int           `json:"x"`
	Y         int           `json:"y"`
	Terrain   string        `json:"terrain"`
	Pollution int           `json:"pollution,omitempty"`
	Stale     bool          `json:"stale"`
	Observed  int           `json:"observed_turn"`
	Planning  *TilePlanning `json:"planning,omitempty"`
}

type TilePlanning struct {
	Cost       int      `json:"cost"`
	Domains    []string `json:"domains"`
	Accessible bool     `json:"accessible"`
	EnemyZOC   bool     `json:"enemy_zoc"`
}

type UnitProjection struct {
	ID              string          `json:"id"`
	OwnerID         string          `json:"owner_id"`
	TileID          string          `json:"tile_id"`
	TransportID     string          `json:"transport_id,omitempty"`
	ArmyID          string          `json:"army_id,omitempty"`
	Fortified       bool            `json:"fortified"`
	Domain          string          `json:"domain,omitempty"`
	Movement        int             `json:"movement,omitempty"`
	CargoCapacity   int             `json:"cargo_capacity,omitempty"`
	CargoDomains    []string        `json:"cargo_domains,omitempty"`
	ZOC             bool            `json:"zoc,omitempty"`
	Attack          int             `json:"attack,omitempty"`
	Defense         int             `json:"defense,omitempty"`
	Strength        int             `json:"strength,omitempty"`
	Armor           int             `json:"armor,omitempty"`
	Health          int             `json:"health,omitempty"`
	MaxHealth       int             `json:"max_health,omitempty"`
	VisibilityRange int             `json:"visibility_range,omitempty"`
	Effects         []engine.Effect `json:"effects,omitempty"`
}

type CityProjection struct {
	ID               string                  `json:"id"`
	OwnerID          string                  `json:"owner_id"`
	TileID           string                  `json:"tile_id"`
	Population       int                     `json:"population,omitempty"`
	Happiness        int                     `json:"happiness,omitempty"`
	Effects          []engine.Effect         `json:"effects,omitempty"`
	StoredFood       int                     `json:"stored_food,omitempty"`
	StoredProduction int                     `json:"stored_production,omitempty"`
	Queue            []engine.ProductionItem `json:"queue,omitempty"`
	Buildings        []engine.Building       `json:"buildings,omitempty"`
	Yields           *engine.Yield           `json:"yields,omitempty"`
	Stale            bool                    `json:"stale"`
	Observed         int                     `json:"observed_turn"`
}

type ResearchProjection struct {
	Advances  []string `json:"advances,omitempty"`
	TargetID  string   `json:"target_id,omitempty"`
	Progress  int      `json:"progress,omitempty"`
	Available []string `json:"available,omitempty"`
}

type DiplomacyRelationProjection struct {
	TargetPlayerID string `json:"target_player_id"`
	Treaty         string `json:"treaty,omitempty"`
	Trade          bool   `json:"trade,omitempty"`
}

type DiplomacyProposalProjection struct {
	ID           string `json:"id"`
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
	Kind         string `json:"kind"`
}

type DiplomacyProjection struct {
	GovernmentID string                        `json:"government_id,omitempty"`
	PublicWorks  int                           `json:"public_works,omitempty"`
	Relations    []DiplomacyRelationProjection `json:"relations,omitempty"`
	Proposals    []DiplomacyProposalProjection `json:"proposals,omitempty"`
}

type RulesProjection struct {
	ContentHash         string                                      `json:"content_hash"`
	Units               []rules.RuntimeUnitDefinition               `json:"units,omitempty"`
	Production          []rules.RuntimeProduction                   `json:"production,omitempty"`
	Advances            []rules.RuntimeAdvanceDefinition            `json:"advances,omitempty"`
	Governments         []rules.RuntimeGovernmentDefinition         `json:"governments,omitempty"`
	TerrainImprovements []rules.RuntimeTerrainImprovementDefinition `json:"terrain_improvements,omitempty"`
	PollutionModel      *rules.RuntimePollutionModel                `json:"pollution_model,omitempty"`
	Movement            *rules.RuntimeMovement                      `json:"movement,omitempty"`
	AI                  *rules.RuntimeAI                            `json:"ai,omitempty"`
}

type ActionProjection struct {
	Command  string          `json:"command"`
	Route    string          `json:"route"`
	EntityID string          `json:"entity_id,omitempty"`
	TargetID string          `json:"target_id,omitempty"`
	Fields   []string        `json:"fields,omitempty"`
	Payload  json.RawMessage `json:"payload,omitempty"`
}

type Projection struct {
	Revision         int64                `json:"revision"`
	Turn             int                  `json:"turn"`
	Tiles            []TileProjection     `json:"tiles"`
	Units            []UnitProjection     `json:"units"`
	Cities           []CityProjection     `json:"cities"`
	Research         *ResearchProjection  `json:"research,omitempty"`
	Diplomacy        *DiplomacyProjection `json:"diplomacy,omitempty"`
	Rules            *RulesProjection     `json:"rules,omitempty"`
	LegalActions     []ActionProjection   `json:"legal_actions,omitempty"`
	EmbassyPlayerIDs []string             `json:"embassy_player_ids,omitempty"`
	PlanningStatus   string               `json:"planning_status,omitempty"`
}

// Update records only what each player can currently observe. The returned state
// remains the single canonical authority and is safe to persist/reload unchanged.
func Update(state engine.State) (engine.State, error) {
	if err := engine.ValidateState(state); err != nil {
		return engine.State{}, err
	}
	visible := visibleTiles(state)
	memory := make(map[string]engine.FogObservation, len(state.Fog))
	for _, observation := range state.Fog {
		memory[observation.PlayerID+"\x00"+observation.TileID] = observation
	}
	cities := citiesByTile(state.Cities)
	for playerID, tiles := range visible {
		for tileID := range tiles {
			tile, ok := tileByID(state.Tiles, tileID)
			if !ok {
				continue
			}
			observation := engine.FogObservation{PlayerID: playerID, TileID: tile.ID, Turn: state.Phase.Turn, Terrain: tile.Terrain}
			if city, ok := cities[tile.ID]; ok {
				observation.CityID = city.ID
				observation.CityOwnerID = city.OwnerID
				observation.CityEffects = append([]engine.Effect(nil), city.Effects...)
			}
			memory[playerID+"\x00"+tileID] = observation
		}
	}
	state.Fog = state.Fog[:0]
	for _, observation := range memory {
		state.Fog = append(state.Fog, observation)
	}
	sort.Slice(state.Fog, func(i, j int) bool {
		if state.Fog[i].PlayerID == state.Fog[j].PlayerID {
			return state.Fog[i].TileID < state.Fog[j].TileID
		}
		return state.Fog[i].PlayerID < state.Fog[j].PlayerID
	})
	return state, nil
}

func Project(state engine.State, playerID string) (Projection, error) {
	if playerID == "" {
		return Projection{}, errors.New("viewer player is required")
	}
	if err := engine.ValidateState(state); err != nil {
		return Projection{}, err
	}
	visible := visibleTiles(state)[playerID]
	memory := memoryByTile(state.Fog, playerID)
	planning, err := planningTiles(state, playerID, visible)
	if err != nil {
		return Projection{}, err
	}
	result := Projection{Revision: state.Revision, Turn: state.Phase.Turn, PlanningStatus: planningStatus(state, playerID), Rules: rulesProjection(state)}
	for _, player := range state.Players {
		if player.ID != playerID {
			continue
		}
		for _, nation := range state.Nations {
			if nation.ID != player.NationID {
				continue
			}
			result.Research = researchProjection(state, nation)
			result.Diplomacy = diplomacyProjection(state, playerID, nation)
			for _, embassyNationID := range nation.Embassies {
				for _, embassyPlayer := range state.Players {
					if embassyPlayer.NationID == embassyNationID {
						result.EmbassyPlayerIDs = append(result.EmbassyPlayerIDs, embassyPlayer.ID)
					}
				}
			}
		}
	}
	for _, tile := range state.Tiles {
		if visible[tile.ID] {
			result.Tiles = append(result.Tiles, TileProjection{ID: tile.ID, X: tile.X, Y: tile.Y, Terrain: tile.Terrain, Pollution: tile.Pollution, Observed: state.Phase.Turn, Planning: planning[tile.ID]})
			continue
		}
		if observation, ok := memory[tile.ID]; ok {
			result.Tiles = append(result.Tiles, TileProjection{ID: tile.ID, X: tile.X, Y: tile.Y, Terrain: observation.Terrain, Stale: true, Observed: observation.Turn})
			if observation.CityID != "" {
				result.Cities = append(result.Cities, CityProjection{ID: observation.CityID, OwnerID: observation.CityOwnerID, TileID: tile.ID, Effects: append([]engine.Effect(nil), observation.CityEffects...), Stale: true, Observed: observation.Turn})
			}
		}
	}
	for _, unit := range state.Units {
		if unit.OwnerID == playerID || visible[unit.TileID] {
			projected := UnitProjection{ID: unit.ID, OwnerID: unit.OwnerID, TileID: unit.TileID, TransportID: unit.TransportID, ArmyID: unit.ArmyID, Fortified: unit.Fortified}
			if unit.OwnerID == playerID {
				projected.Domain, projected.Movement = unit.Domain, unit.Movement
				projected.CargoCapacity = unit.CargoCapacity
				projected.CargoDomains = append([]string(nil), unit.CargoDomains...)
				projected.ZOC = unit.ZOC
				projected.Attack, projected.Defense, projected.Strength, projected.Armor = unit.Attack, unit.Defense, unit.Strength, unit.Armor
				projected.Health, projected.MaxHealth = unit.Health, unit.MaxHealth
				projected.VisibilityRange = unit.VisibilityRange
				projected.Effects = append([]engine.Effect(nil), unit.Effects...)
			}
			result.Units = append(result.Units, projected)
		}
	}
	for _, city := range state.Cities {
		if visible[city.TileID] {
			result.Cities = append(result.Cities, projectCity(city, playerID, state.Phase.Turn))
		}
	}
	sort.Slice(result.Tiles, func(i, j int) bool { return result.Tiles[i].ID < result.Tiles[j].ID })
	sort.Slice(result.Units, func(i, j int) bool { return result.Units[i].ID < result.Units[j].ID })
	sort.Slice(result.Cities, func(i, j int) bool { return result.Cities[i].ID < result.Cities[j].ID })
	sort.Strings(result.EmbassyPlayerIDs)
	result.LegalActions = legalActions(state, result, playerID)
	return result, nil
}

func diplomacyProjection(state engine.State, playerID string, nation engine.Nation) *DiplomacyProjection {
	projection := &DiplomacyProjection{GovernmentID: nation.GovernmentID, PublicWorks: nation.PublicWorks}
	for _, relation := range state.Diplomacy {
		if relation.PlayerAID != playerID && relation.PlayerBID != playerID || !relation.Contact {
			continue
		}
		target := relation.PlayerAID
		if target == playerID {
			target = relation.PlayerBID
		}
		projection.Relations = append(projection.Relations, DiplomacyRelationProjection{TargetPlayerID: target, Treaty: relation.Treaty, Trade: relation.Trade})
	}
	for _, proposal := range state.Proposals {
		if proposal.FromPlayerID == playerID || proposal.ToPlayerID == playerID {
			projection.Proposals = append(projection.Proposals, DiplomacyProposalProjection{ID: proposal.ID, FromPlayerID: proposal.FromPlayerID, ToPlayerID: proposal.ToPlayerID, Kind: proposal.Kind})
		}
	}
	sort.Slice(projection.Relations, func(i, j int) bool {
		return projection.Relations[i].TargetPlayerID < projection.Relations[j].TargetPlayerID
	})
	sort.Slice(projection.Proposals, func(i, j int) bool { return projection.Proposals[i].ID < projection.Proposals[j].ID })
	if projection.GovernmentID == "" && projection.PublicWorks == 0 && len(projection.Relations) == 0 && len(projection.Proposals) == 0 {
		return nil
	}
	return projection
}

func projectCity(city engine.City, playerID string, observed int) CityProjection {
	projection := CityProjection{ID: city.ID, OwnerID: city.OwnerID, TileID: city.TileID, Effects: append([]engine.Effect(nil), city.Effects...), Observed: observed}
	if city.OwnerID == playerID {
		projection.Population = city.Population
		projection.Happiness = city.Happiness
		projection.StoredFood = city.StoredFood
		projection.StoredProduction = city.StoredProduction
		projection.Queue = append([]engine.ProductionItem(nil), city.Queue...)
		projection.Buildings = append([]engine.Building(nil), city.Buildings...)
		yields := city.Yields
		projection.Yields = &yields
	}
	return projection
}

func researchProjection(state engine.State, nation engine.Nation) *ResearchProjection {
	projection := &ResearchProjection{Advances: append([]string(nil), nation.Advances...), TargetID: nation.ResearchTargetID, Progress: nation.ResearchProgress}
	if runtime, ok := runtimeRules(state); ok {
		for _, advance := range runtime.Advances {
			if projectionAdvanceAvailable(nation, runtime, advance.ID) {
				projection.Available = append(projection.Available, advance.ID)
			}
		}
	}
	sort.Strings(projection.Advances)
	sort.Strings(projection.Available)
	if len(projection.Advances) == 0 && projection.TargetID == "" && projection.Progress == 0 && len(projection.Available) == 0 {
		return nil
	}
	return projection
}

func rulesProjection(state engine.State) *RulesProjection {
	runtime, ok := runtimeRules(state)
	if !ok {
		return nil
	}
	movement := runtime.Movement
	ai := cloneRuntimeAI(runtime.AI)
	return &RulesProjection{
		ContentHash:         runtime.ContentHash,
		Units:               append([]rules.RuntimeUnitDefinition(nil), runtime.Units...),
		Production:          append([]rules.RuntimeProduction(nil), runtime.Production...),
		Advances:            append([]rules.RuntimeAdvanceDefinition(nil), runtime.Advances...),
		Governments:         append([]rules.RuntimeGovernmentDefinition(nil), runtime.Governments...),
		TerrainImprovements: append([]rules.RuntimeTerrainImprovementDefinition(nil), runtime.TerrainImprovements...),
		PollutionModel:      cloneRuntimePollutionModel(runtime.PollutionModel),
		Movement:            &movement,
		AI:                  ai,
	}
}

func cloneRuntimePollutionModel(model *rules.RuntimePollutionModel) *rules.RuntimePollutionModel {
	if model == nil {
		return nil
	}
	copy := *model
	return &copy
}

func cloneRuntimeAI(ai *rules.RuntimeAI) *rules.RuntimeAI {
	if ai == nil {
		return nil
	}
	return &rules.RuntimeAI{
		UnitBuildLists:     cloneRuntimeAILists(ai.UnitBuildLists),
		BuildingBuildLists: cloneRuntimeAILists(ai.BuildingBuildLists),
		WonderBuildLists:   cloneRuntimeAILists(ai.WonderBuildLists),
		AdvanceLists:       cloneRuntimeAILists(ai.AdvanceLists),
		ImprovementLists:   cloneRuntimeAILists(ai.ImprovementLists),
	}
}

func cloneRuntimeAILists(lists []rules.RuntimeAIList) []rules.RuntimeAIList {
	cloned := make([]rules.RuntimeAIList, 0, len(lists))
	for _, list := range lists {
		cloned = append(cloned, rules.RuntimeAIList{ID: list.ID, Items: append([]string(nil), list.Items...)})
	}
	return cloned
}

func runtimeRules(state engine.State) (rules.RuntimeRules, bool) {
	if len(state.Rules.Definitions) == 0 || string(state.Rules.Definitions) == "null" {
		return rules.RuntimeRules{}, false
	}
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	return runtime, err == nil
}

func projectionAdvanceAvailable(nation engine.Nation, runtime rules.RuntimeRules, advanceID string) bool {
	advance, ok := runtime.Advance(advanceID)
	if !ok || containsString(nation.Advances, advanceID) {
		return false
	}
	for _, prerequisite := range advance.Prerequisites {
		if !containsString(nation.Advances, prerequisite) {
			return false
		}
	}
	return true
}

func legalActions(state engine.State, projection Projection, playerID string) []ActionProjection {
	if !viewerCanAct(state, projection, playerID) {
		return nil
	}
	actions := make([]ActionProjection, 0)
	if projection.PlanningStatus == PlanningStatusOpen {
		actions = append(actions, action(engine.CommandRoundComplete, "round-complete", playerID, "", map[string]any{}, nil))
	} else {
		actions = append(actions, action(engine.CommandPass, "pass", playerID, "", map[string]any{}, nil))
	}
	actions = append(actions, researchActions(projection, playerID)...)
	actions = append(actions, governmentActions(projection, playerID)...)
	actions = append(actions, diplomacyActions(projection, playerID)...)
	actions = append(actions, productionActions(projection, playerID)...)
	actions = append(actions, unitActions(projection, playerID)...)
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].Command == actions[j].Command {
			if actions[i].EntityID == actions[j].EntityID {
				if actions[i].TargetID == actions[j].TargetID {
					return string(actions[i].Payload) < string(actions[j].Payload)
				}
				return actions[i].TargetID < actions[j].TargetID
			}
			return actions[i].EntityID < actions[j].EntityID
		}
		return actions[i].Command < actions[j].Command
	})
	return actions
}

func viewerCanAct(state engine.State, projection Projection, playerID string) bool {
	if state.Terminal.Ended || playerID == "" {
		return false
	}
	if projection.PlanningStatus == PlanningStatusOpen {
		return true
	}
	if state.Phase.ResolutionLocked || len(state.Phase.ActivePlayerOrder) == 0 || state.Phase.ActivePlayerIndex < 0 || state.Phase.ActivePlayerIndex >= len(state.Phase.ActivePlayerOrder) {
		return false
	}
	return state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex] == playerID
}

func researchActions(projection Projection, playerID string) []ActionProjection {
	if projection.Research == nil || projection.Research.TargetID != "" {
		return nil
	}
	actions := make([]ActionProjection, 0, len(projection.Research.Available))
	for _, advanceID := range projection.Research.Available {
		actions = append(actions, action(engine.CommandSetResearch, "set-research", playerID, advanceID, map[string]string{"advance_id": advanceID}, []string{"advance_id"}))
	}
	return actions
}

func governmentActions(projection Projection, playerID string) []ActionProjection {
	if projection.Rules == nil || projection.Diplomacy == nil {
		return nil
	}
	known := []string(nil)
	if projection.Research != nil {
		known = projection.Research.Advances
	}
	actions := make([]ActionProjection, 0)
	for _, government := range projection.Rules.Governments {
		if government.ID == "" || government.ID == projection.Diplomacy.GovernmentID || !advanceSatisfied(known, government.EnableAdvanceID) {
			continue
		}
		actions = append(actions, action(engine.CommandSetGovernment, "set-government", playerID, government.ID, map[string]string{"government_id": government.ID}, []string{"government_id"}))
	}
	return actions
}

func diplomacyActions(projection Projection, playerID string) []ActionProjection {
	if projection.Diplomacy == nil {
		return nil
	}
	pending := make(map[string]bool, len(projection.Diplomacy.Proposals))
	actions := make([]ActionProjection, 0)
	for _, proposal := range projection.Diplomacy.Proposals {
		if proposal.Kind != "" && proposal.FromPlayerID != "" && proposal.ToPlayerID != "" {
			pending[proposal.Kind+"\x00"+proposal.FromPlayerID+"\x00"+proposal.ToPlayerID] = true
			pending[proposal.Kind+"\x00"+proposal.ToPlayerID+"\x00"+proposal.FromPlayerID] = true
		}
		if proposal.ToPlayerID == playerID {
			actions = append(actions, action(engine.CommandRespondAgreement, "respond-agreement", proposal.ID, proposal.FromPlayerID, map[string]any{"proposal_id": proposal.ID, "accept": true}, []string{"proposal_id", "accept"}))
		}
	}
	for _, relation := range projection.Diplomacy.Relations {
		if relation.TargetPlayerID == "" || relation.Treaty == "union" || pending["union\x00"+playerID+"\x00"+relation.TargetPlayerID] {
			continue
		}
		proposalID := "client-union-" + playerID + "-" + relation.TargetPlayerID
		actions = append(actions, action(engine.CommandProposeAgreement, "propose-agreement", playerID, relation.TargetPlayerID, map[string]string{"proposal_id": proposalID, "target_player_id": relation.TargetPlayerID, "kind": "union"}, []string{"proposal_id", "target_player_id", "kind"}))
	}
	return actions
}

func productionActions(projection Projection, playerID string) []ActionProjection {
	actions := make([]ActionProjection, 0)
	known := []string(nil)
	if projection.Research != nil {
		known = projection.Research.Advances
	}
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.Stale {
			continue
		}
		for _, queued := range city.Queue {
			actions = append(actions, action(engine.CommandCancelProduction, "cancel-production", city.ID, queued.ID, map[string]string{"city_id": city.ID, "item_id": queued.ID}, []string{"city_id", "item_id"}))
		}
		if len(city.Queue) != 0 || projection.Rules == nil {
			continue
		}
		built := cityBuildingIDs(city)
		for _, production := range projection.Rules.Production {
			if production.ID == "" || production.Kind == engine.ProductionUnit || !allAdvancesSatisfied(known, production.Prerequisites) || anyAdvanceSatisfied(known, production.ObsoleteBy) {
				continue
			}
			if production.Kind == engine.ProductionBuilding && built[production.ID] {
				continue
			}
			item := productionActionItem(production)
			actions = append(actions, action(engine.CommandQueueProduction, "queue-production", city.ID, production.ID, map[string]any{"city_id": city.ID, "item": item}, []string{"city_id", "item"}))
		}
		for _, unit := range projection.Rules.Units {
			if unit.ID == "" || !allAdvancesSatisfied(known, unit.Prerequisites) || anyAdvanceSatisfied(known, unit.ObsoleteBy) {
				continue
			}
			cost := unit.ProductionCost
			if cost < 1 {
				cost = 1
			}
			item := engine.ProductionItem{ID: unit.ID, Kind: engine.ProductionUnit, Cost: cost, Upkeep: unit.Upkeep}
			if len(unit.Prerequisites) > 0 {
				item.Prerequisite = unit.Prerequisites[0]
			}
			if len(unit.ObsoleteBy) > 0 {
				item.ObsoletedBy = unit.ObsoleteBy[0]
			}
			actions = append(actions, action(engine.CommandQueueProduction, "queue-production", city.ID, unit.ID, map[string]any{"city_id": city.ID, "item": item}, []string{"city_id", "item"}))
			break
		}
	}
	return actions
}

func cityBuildingIDs(city CityProjection) map[string]bool {
	built := make(map[string]bool, len(city.Buildings))
	for _, building := range city.Buildings {
		built[building.ID] = true
	}
	return built
}

func productionActionItem(production rules.RuntimeProduction) engine.ProductionItem {
	cost := production.Cost
	if cost < 1 {
		cost = 1
	}
	item := engine.ProductionItem{ID: production.ID, Kind: production.Kind, Cost: cost, Upkeep: production.Upkeep}
	if len(production.Prerequisites) > 0 {
		item.Prerequisite = production.Prerequisites[0]
	}
	if len(production.ObsoleteBy) > 0 {
		item.ObsoletedBy = production.ObsoleteBy[0]
	}
	if len(production.Effects) > 0 {
		item.Effects = make([]engine.Effect, 0, len(production.Effects))
		for _, effect := range production.Effects {
			item.Effects = append(item.Effects, engine.Effect{ID: effect.ID, Kind: effect.Kind, Value: effect.Value})
		}
	}
	return item
}

func unitActions(projection Projection, playerID string) []ActionProjection {
	tiles := projectedTiles(projection)
	units := projectedUnits(projection)
	cities := projectedCities(projection)
	actions := make([]ActionProjection, 0)
	for _, unit := range projection.Units {
		if unit.OwnerID != playerID {
			continue
		}
		if unit.TransportID != "" {
			actions = append(actions, disembarkActions(projection, tiles, units, unit)...)
			continue
		}
		if unit.ArmyID == "" && unit.TileID != "" {
			actions = append(actions, action(engine.CommandArmyCreate, "army-create", unit.ID, "", map[string]any{"army_id": "client-army-" + unit.ID, "unit_ids": []string{unit.ID}}, []string{"army_id", "unit_ids"}))
		}
		if unit.TileID != "" && !unit.Fortified {
			actions = append(actions, action(engine.CommandFortify, "fortify", unit.ID, unit.TileID, map[string]string{"unit_id": unit.ID}, []string{"unit_id"}))
		}
		if unit.ArmyID == "" && unit.TileID != "" && hasEffect(unit.Effects, engine.EffectSettler) && !cityOnTile(projection, unit.TileID) {
			actions = append(actions, action(engine.CommandFoundCity, "found-city", unit.ID, unit.TileID, map[string]string{"unit_id": unit.ID, "city_id": "client-city-" + unit.ID, "name": "City " + unit.ID}, []string{"unit_id", "city_id", "name"}))
		}
		actions = append(actions, moveActions(projection, tiles, unit)...)
		actions = append(actions, embarkActions(projection, unit)...)
		actions = append(actions, combatActions(projection, tiles, unit)...)
		actions = append(actions, specialActions(projection, tiles, cities, unit, playerID)...)
		actions = append(actions, terrainActions(projection, tiles, unit)...)
	}
	return actions
}

func moveActions(projection Projection, tiles map[string]TileProjection, unit UnitProjection) []ActionProjection {
	if unit.TileID == "" || unit.Movement < 1 || unit.Domain == "" {
		return nil
	}
	origin, ok := tiles[unit.TileID]
	if !ok {
		return nil
	}
	occupied := make(map[string]bool)
	for _, other := range projection.Units {
		if other.OwnerID != unit.OwnerID && other.TileID != "" {
			occupied[other.TileID] = true
		}
	}
	actions := make([]ActionProjection, 0)
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.ID == unit.TileID || tile.Planning == nil || !tile.Planning.Accessible || tile.Planning.Cost > unit.Movement || !containsString(tile.Planning.Domains, unit.Domain) || occupied[tile.ID] || !adjacentProjection(origin, tile) {
			continue
		}
		actions = append(actions, action(engine.CommandMoveUnit, "move", unit.ID, tile.ID, map[string]any{"unit_id": unit.ID, "path": []string{tile.ID}}, []string{"unit_id", "path"}))
	}
	return actions
}

func embarkActions(projection Projection, unit UnitProjection) []ActionProjection {
	if unit.TileID == "" || unit.ArmyID != "" {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, transport := range projection.Units {
		if transport.OwnerID != unit.OwnerID || transport.ID == unit.ID || transport.TileID != unit.TileID || transport.TransportID != "" || transport.CargoCapacity <= cargoCount(projection, transport.ID) || !containsString(transport.CargoDomains, unit.Domain) {
			continue
		}
		actions = append(actions, action(engine.CommandEmbark, "embark", unit.ID, transport.ID, map[string]string{"unit_id": unit.ID, "transport_id": transport.ID}, []string{"unit_id", "transport_id"}))
	}
	return actions
}

func disembarkActions(projection Projection, tiles map[string]TileProjection, units map[string]UnitProjection, unit UnitProjection) []ActionProjection {
	transport, ok := units[unit.TransportID]
	origin, tileOK := tiles[transport.TileID]
	if !ok || !tileOK || unit.Movement < 1 || unit.Domain == "" {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || !tile.Planning.Accessible || tile.Planning.EnemyZOC || !containsString(tile.Planning.Domains, unit.Domain) || !adjacentProjection(origin, tile) {
			continue
		}
		actions = append(actions, action(engine.CommandDisembark, "disembark", unit.ID, tile.ID, map[string]string{"unit_id": unit.ID, "tile_id": tile.ID}, []string{"unit_id", "tile_id"}))
	}
	return actions
}

func combatActions(projection Projection, tiles map[string]TileProjection, attacker UnitProjection) []ActionProjection {
	if attacker.ArmyID == "" || attacker.Attack < 1 || attacker.TileID == "" {
		return nil
	}
	origin, ok := tiles[attacker.TileID]
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	actions := make([]ActionProjection, 0)
	for _, defender := range projection.Units {
		if defender.OwnerID == attacker.OwnerID || defender.ArmyID == "" || defender.TileID == "" || defender.TransportID != "" || seen[defender.ArmyID] {
			continue
		}
		target, ok := tiles[defender.TileID]
		if !ok || !adjacentProjection(origin, target) || !projectionDomainsCompatible(attacker, defender) {
			continue
		}
		seen[defender.ArmyID] = true
		actions = append(actions, action(engine.CommandCombat, "combat", attacker.ArmyID, defender.ArmyID, map[string]string{"attacker_army_id": attacker.ArmyID, "defender_army_id": defender.ArmyID}, []string{"attacker_army_id", "defender_army_id"}))
	}
	return actions
}

func specialActions(projection Projection, tiles map[string]TileProjection, cities map[string]CityProjection, unit UnitProjection, playerID string) []ActionProjection {
	if unit.TileID == "" || unit.Movement < 1 {
		return nil
	}
	origin, ok := tiles[unit.TileID]
	if !ok {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, city := range cities {
		cityTile, ok := tiles[city.TileID]
		if !ok || city.OwnerID == playerID || !adjacentProjection(origin, cityTile) {
			continue
		}
		if hasEffect(unit.Effects, engine.EffectSpecialEmbassy) && !containsString(projection.EmbassyPlayerIDs, city.OwnerID) {
			actions = append(actions, action(engine.CommandEstablishEmbassy, "establish-embassy", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSoothsaying) {
			actions = append(actions, action(engine.CommandSoothsaying, "soothsaying", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSteal) || hasEffect(unit.Effects, engine.EffectSpecialCapture) {
			actions = append(actions, action(engine.CommandStealTechnology, "steal-technology", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
	}
	return actions
}

func terrainActions(projection Projection, tiles map[string]TileProjection, unit UnitProjection) []ActionProjection {
	if unit.TileID == "" || unit.Movement < 1 || projection.Diplomacy == nil || !hasEffect(unit.Effects, engine.EffectSpecialTerraform) {
		return nil
	}
	origin, ok := tiles[unit.TileID]
	if !ok {
		return nil
	}
	actions := make([]ActionProjection, 0)
	terraformCost, terraformTurns := effectValue(unit.Effects, engine.EffectSpecialTerraform), effectValue(unit.Effects, engine.EffectSpecialTerraformTurns)
	for _, tile := range projection.Tiles {
		if tile.Stale || !adjacentProjection(origin, tile) {
			continue
		}
		if tile.Terrain == engine.TerrainForest && terraformCost > 0 && terraformTurns > 0 && projection.Diplomacy.PublicWorks >= terraformCost {
			actions = append(actions, action(engine.CommandTerraform, "terraform", unit.ID, tile.ID, map[string]string{"unit_id": unit.ID, "tile_id": tile.ID, "target_terrain": engine.TerrainPlains}, []string{"unit_id", "tile_id", "target_terrain"}))
		}
		if projection.Rules != nil {
			for _, improvement := range orderedTerrainImprovements(*projection.Rules) {
				effective := projectedEffectiveTerrainImprovement(*projection.Rules, improvement, tile.Terrain)
				if effective.PublicWorksCost < 1 || effective.BuildTurns < 1 || projection.Diplomacy.PublicWorks < effective.PublicWorksCost {
					continue
				}
				actions = append(actions, action(engine.CommandBuildImprovement, "build-improvement", unit.ID, tile.ID, map[string]string{"unit_id": unit.ID, "tile_id": tile.ID, "improvement_id": improvement.ID}, []string{"unit_id", "tile_id", "improvement_id"}))
				break
			}
		}
	}
	return actions
}

func projectedEffectiveTerrainImprovement(projection RulesProjection, improvement rules.RuntimeTerrainImprovementDefinition, sourceTerrain string) rules.RuntimeTerrainImprovementDefinition {
	if improvement.TargetTerrain == "" || projection.Movement == nil {
		return improvement
	}
	from, fromOK := projectedRuntimeTerrain(projection.Movement.Terrains, sourceTerrain)
	to, toOK := projectedRuntimeTerrain(projection.Movement.Terrains, improvement.TargetTerrain)
	if !fromOK || !toOK || from.TransformRemoveTime < 1 || from.TransformRemoveMaterials < 1 || to.TransformAddTime < 1 || to.TransformAddMaterials < 1 {
		return improvement
	}
	effective := improvement
	effective.BuildTurns = from.TransformRemoveTime + to.TransformAddTime
	effective.PublicWorksCost = from.TransformRemoveMaterials + to.TransformAddMaterials
	return effective
}

func projectedRuntimeTerrain(terrains []rules.RuntimeTerrain, id string) (rules.RuntimeTerrain, bool) {
	for _, terrain := range terrains {
		if terrain.ID == id {
			return terrain, true
		}
	}
	return rules.RuntimeTerrain{}, false
}

func orderedTerrainImprovements(projection RulesProjection) []rules.RuntimeTerrainImprovementDefinition {
	if len(projection.TerrainImprovements) == 0 || projection.AI == nil || len(projection.AI.ImprovementLists) == 0 {
		return append([]rules.RuntimeTerrainImprovementDefinition(nil), projection.TerrainImprovements...)
	}
	byID := make(map[string]rules.RuntimeTerrainImprovementDefinition, len(projection.TerrainImprovements))
	used := make(map[string]bool, len(projection.TerrainImprovements))
	for _, improvement := range projection.TerrainImprovements {
		byID[improvement.ID] = improvement
	}
	ordered := make([]rules.RuntimeTerrainImprovementDefinition, 0, len(projection.TerrainImprovements))
	for _, list := range projection.AI.ImprovementLists {
		for _, id := range list.Items {
			if improvement, ok := byID[id]; ok && !used[id] {
				ordered = append(ordered, improvement)
				used[id] = true
			}
		}
	}
	for _, improvement := range projection.TerrainImprovements {
		if !used[improvement.ID] {
			ordered = append(ordered, improvement)
		}
	}
	return ordered
}

func action(command, route, entityID, targetID string, payload any, fields []string) ActionProjection {
	encoded, _ := json.Marshal(payload)
	return ActionProjection{Command: command, Route: route, EntityID: entityID, TargetID: targetID, Fields: append([]string(nil), fields...), Payload: encoded}
}

func projectedTiles(projection Projection) map[string]TileProjection {
	result := make(map[string]TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		if !tile.Stale {
			result[tile.ID] = tile
		}
	}
	return result
}

func projectedUnits(projection Projection) map[string]UnitProjection {
	result := make(map[string]UnitProjection, len(projection.Units))
	for _, unit := range projection.Units {
		result[unit.ID] = unit
	}
	return result
}

func projectedCities(projection Projection) map[string]CityProjection {
	result := make(map[string]CityProjection, len(projection.Cities))
	for _, city := range projection.Cities {
		if !city.Stale {
			result[city.ID] = city
		}
	}
	return result
}

func cityOnTile(projection Projection, tileID string) bool {
	for _, city := range projection.Cities {
		if !city.Stale && city.TileID == tileID {
			return true
		}
	}
	return false
}

func cargoCount(projection Projection, transportID string) int {
	count := 0
	for _, unit := range projection.Units {
		if unit.TransportID == transportID {
			count++
		}
	}
	return count
}

func projectionDomainsCompatible(attacker, defender UnitProjection) bool {
	return attacker.Domain == "" || defender.Domain == "" || attacker.Domain == defender.Domain || attacker.Domain == engine.DomainAmphibious || defender.Domain == engine.DomainAmphibious
}

func advanceSatisfied(known []string, advanceID string) bool {
	return advanceID == "" || containsString(known, advanceID)
}

func allAdvancesSatisfied(known, required []string) bool {
	for _, advanceID := range required {
		if !containsString(known, advanceID) {
			return false
		}
	}
	return true
}

func anyAdvanceSatisfied(known, blocked []string) bool {
	for _, advanceID := range blocked {
		if containsString(known, advanceID) {
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

func effectValue(effects []engine.Effect, kind string) int {
	for _, effect := range effects {
		if effect.Kind == kind {
			return effect.Value
		}
	}
	return 0
}

func planningStatus(state engine.State, playerID string) string {
	if state.Terminal.Ended {
		return ""
	}
	if state.Phase.Mode != turns.ModeSimultaneous && state.Phase.Mode != turns.ModeHybrid {
		return ""
	}
	stage := state.Policy.Stage
	if stage == "" {
		stage = turns.StagePlanning
	}
	if stage != turns.StagePlanning {
		return ""
	}
	eligible := false
	for _, player := range state.Players {
		if player.ID == playerID && player.Lifecycle == engine.PlayerActive {
			eligible = true
			break
		}
	}
	if !eligible {
		return ""
	}
	for _, ended := range state.Policy.EarlyEnds {
		if ended == playerID {
			return PlanningStatusClosed
		}
	}
	for _, queued := range state.Policy.Queue {
		if queued.ActorID == playerID {
			return PlanningStatusQueued
		}
	}
	return PlanningStatusOpen
}

func planningTiles(state engine.State, playerID string, visible map[string]bool) (map[string]*TilePlanning, error) {
	result := make(map[string]*TilePlanning)
	if len(state.Rules.Definitions) == 0 || string(state.Rules.Definitions) == "null" {
		return result, nil
	}
	var definition struct {
		CompatibilityVersion string `json:"compatibility_version"`
	}
	if err := json.Unmarshal(state.Rules.Definitions, &definition); err != nil {
		return nil, err
	}
	if definition.CompatibilityVersion != rules.CTP2CompatibilityVersion {
		return result, nil
	}
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		return nil, err
	}
	if !runtime.Movement.Configured {
		return result, nil
	}
	for _, tile := range state.Tiles {
		if !visible[tile.ID] {
			continue
		}
		terrain, ok := runtime.Movement.Terrain(engine.EffectiveTerrain(state, tile.ID))
		if !ok {
			return nil, errors.New("visible tile terrain is unavailable to the runtime movement model")
		}
		result[tile.ID] = &TilePlanning{
			Cost:       projectedTerrainCost(state, tile.ID, terrain.Cost),
			Domains:    append([]string(nil), terrain.Domains...),
			Accessible: tile.OwnerID == "" || tile.OwnerID == playerID || (runtime.Movement.OpenAccess && tile.Open),
		}
	}
	if !runtime.Movement.ZOCStopsPath {
		return result, nil
	}
	for _, unit := range state.Units {
		if unit.OwnerID == playerID || !unit.ZOC || unit.TransportID == "" && !visible[unit.TileID] {
			continue
		}
		for tileID, planning := range result {
			tile, _ := tileByID(state.Tiles, tileID)
			source, ok := tileByID(state.Tiles, unit.TileID)
			if ok && adjacentEngineTiles(source, tile) {
				planning.EnemyZOC = true
			}
		}
	}
	return result, nil
}

func projectedTerrainCost(state engine.State, tileID string, base int) int {
	cost := base
	for _, overlay := range state.Overlays {
		if overlay.TileID == tileID && overlay.Kind == "improvement" && overlay.Remaining == 0 {
			cost += overlay.MovementCostDelta
		}
	}
	if cost < 1 {
		return 1
	}
	return cost
}

func Plan(projection Projection, unitID, targetTileID string) ([]string, error) {
	var mover UnitProjection
	foundMover := false
	occupied := make(map[string]bool)
	for _, unit := range projection.Units {
		if unit.ID == unitID {
			mover, foundMover = unit, true
		}
	}
	if !foundMover || mover.Domain == "" || mover.TileID == "" || mover.Movement < 0 || targetTileID == "" {
		return nil, errors.New("planner inputs are unavailable")
	}
	for _, unit := range projection.Units {
		if unit.OwnerID != mover.OwnerID && unit.ID != unitID {
			occupied[unit.TileID] = true
		}
	}
	tiles := make(map[string]TileProjection, len(projection.Tiles))
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Planning == nil || tile.Planning.Cost < 1 || !contains(tile.Planning.Domains, mover.Domain) {
			continue
		}
		tiles[tile.ID] = tile
	}
	if _, ok := tiles[mover.TileID]; !ok {
		return nil, errors.New("planner origin is not currently visible")
	}
	if _, ok := tiles[targetTileID]; !ok {
		return nil, errors.New("planner target is not currently visible")
	}
	type node struct {
		tileID string
		spent  int
		path   []string
	}
	queue := []node{{tileID: mover.TileID}}
	best := map[string]int{mover.TileID: 0}
	for len(queue) > 0 {
		sort.Slice(queue, func(i, j int) bool {
			if queue[i].spent == queue[j].spent {
				return queue[i].tileID < queue[j].tileID
			}
			return queue[i].spent < queue[j].spent
		})
		current := queue[0]
		queue = queue[1:]
		if current.tileID == targetTileID {
			return current.path, nil
		}
		from := tiles[current.tileID]
		for _, to := range tiles {
			if !adjacentProjection(from, to) || !to.Planning.Accessible || occupied[to.ID] {
				continue
			}
			spent := current.spent + to.Planning.Cost
			if spent > mover.Movement || (to.Planning.EnemyZOC && to.ID != targetTileID) {
				continue
			}
			if previous, seen := best[to.ID]; seen && previous <= spent {
				continue
			}
			best[to.ID] = spent
			path := append(append([]string(nil), current.path...), to.ID)
			queue = append(queue, node{tileID: to.ID, spent: spent, path: path})
		}
	}
	return nil, errors.New("no legal visible route")
}

// PlanKnown permits routing through tiles already retained in a player's projection.
// It derives only coarse navigation metadata from the remembered public terrain.
func PlanKnown(projection Projection, unitID, targetTileID string) ([]string, error) {
	known := projection
	known.Tiles = append([]TileProjection(nil), projection.Tiles...)
	for index := range known.Tiles {
		tile := &known.Tiles[index]
		if !tile.Stale || tile.Planning != nil {
			continue
		}
		planning := TilePlanning{Cost: 1, Domains: []string{"land", "amphibious"}, Accessible: true}
		if tile.Terrain == "forest" {
			planning.Cost = 2
		}
		if tile.Terrain == "ocean" {
			planning.Domains = []string{"sea", "amphibious"}
		}
		tile.Stale = false
		tile.Planning = &planning
	}
	return Plan(known, unitID, targetTileID)
}

func contains(items []string, value string) bool {
	index := sort.SearchStrings(items, value)
	return index < len(items) && items[index] == value
}

func adjacentProjection(left, right TileProjection) bool {
	dx, dy := abs(left.X-right.X), abs(left.Y-right.Y)
	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}

func adjacentEngineTiles(left, right engine.Tile) bool {
	dx, dy := abs(left.X-right.X), abs(left.Y-right.Y)
	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}

// AIInput is deliberately byte-identical to the controller's HTTP projection.
func AIInput(state engine.State, playerID string) ([]byte, error) {
	projection, err := Project(state, playerID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(projection)
}

// SafeError prevents existence-specific details from crossing a viewer boundary.
func SafeError(error) []byte { return []byte(`{"error":"gameplay command unavailable"}`) }

type Projector struct{ PlayerID string }

func NewProjector(playerID ...string) Projector {
	if len(playerID) == 0 {
		return Projector{}
	}
	return Projector{PlayerID: playerID[0]}
}
func (p Projector) ProjectSnapshot(viewer events.Viewer, canonical []byte) ([]byte, error) {
	state, err := engine.DecodeCanonicalStateJSON(canonical)
	if err != nil {
		return nil, err
	}
	projection, err := Project(state, p.viewerPlayerID(viewer))
	if err != nil {
		return nil, err
	}
	if !state.Terminal.Ended {
		return json.Marshal(projection)
	}
	return json.Marshal(struct {
		Projection
		Terminal struct {
			WinningPlayerID string `json:"winning_player_id"`
			Reason          string `json:"reason"`
		} `json:"terminal"`
	}{Projection: projection, Terminal: struct {
		WinningPlayerID string `json:"winning_player_id"`
		Reason          string `json:"reason"`
	}{WinningPlayerID: state.Terminal.WinningPlayerID, Reason: state.Terminal.Reason}})
}
func (p Projector) ProjectEvent(_ events.Viewer, event events.Event) ([]byte, error) {
	return json.Marshal(struct {
		Type     string `json:"type"`
		Sequence int64  `json:"sequence"`
		Revision int64  `json:"revision"`
		Ordinal  int    `json:"ordinal"`
	}{Type: event.Type, Sequence: event.Sequence, Revision: event.Revision, Ordinal: event.Ordinal})
}
func (p Projector) viewerPlayerID(viewer events.Viewer) string {
	if p.PlayerID != "" {
		return p.PlayerID
	}
	if viewer.AccountID > 0 {
		return "account-" + strconv.FormatInt(viewer.AccountID, 10)
	}
	return strconv.FormatInt(viewer.PlayerID, 10)
}

func visibleTiles(state engine.State) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	for _, player := range state.Players {
		result[player.ID] = make(map[string]bool)
	}
	for _, unit := range state.Units {
		radius := defaultSightRadius
		if unit.VisibilityRange > radius {
			radius = unit.VisibilityRange
		}
		addSight(result[unit.OwnerID], state, unit.TileID, radius)
	}
	for _, city := range state.Cities {
		addSight(result[city.OwnerID], state, city.TileID, defaultSightRadius)
	}
	return result
}
func addSight(visible map[string]bool, state engine.State, origin string, radius int) {
	from, ok := tileByID(state.Tiles, origin)
	if !ok {
		return
	}
	for _, tile := range state.Tiles {
		distance := abs(tile.X-from.X) + abs(tile.Y-from.Y)
		if distance > radius || !hasLOS(state, from, tile) {
			continue
		}
		visible[tile.ID] = true
	}
}
func hasLOS(state engine.State, from, to engine.Tile) bool {
	x, y := from.X, from.Y
	dx, dy := sign(to.X-from.X), sign(to.Y-from.Y)
	for x != to.X || y != to.Y {
		x += dx
		y += dy
		if x == to.X && y == to.Y {
			return true
		}
		tile, ok := tileAt(state.Tiles, x, y)
		if !ok || engine.EffectiveTerrain(state, tile.ID) == engine.TerrainForest {
			return false
		}
	}
	return true
}
func tileByID(tiles []engine.Tile, id string) (engine.Tile, bool) {
	for _, tile := range tiles {
		if tile.ID == id {
			return tile, true
		}
	}
	return engine.Tile{}, false
}
func tileAt(tiles []engine.Tile, x, y int) (engine.Tile, bool) {
	for _, tile := range tiles {
		if tile.X == x && tile.Y == y {
			return tile, true
		}
	}
	return engine.Tile{}, false
}
func citiesByTile(cities []engine.City) map[string]engine.City {
	result := make(map[string]engine.City, len(cities))
	for _, city := range cities {
		result[city.TileID] = city
	}
	return result
}
func memoryByTile(fog []engine.FogObservation, playerID string) map[string]engine.FogObservation {
	result := make(map[string]engine.FogObservation)
	for _, observation := range fog {
		if observation.PlayerID == playerID {
			result[observation.TileID] = observation
		}
	}
	return result
}
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
