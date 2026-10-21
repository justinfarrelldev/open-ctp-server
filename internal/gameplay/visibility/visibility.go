// Package visibility derives deterministic player-visible state and persistent fog memory.
package visibility

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/events"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	defaultSightRadius = 2
	retailPartyCost    = 100
)

const (
	PlanningStatusOpen   = "open"
	PlanningStatusQueued = "queued"
	PlanningStatusClosed = "closed"
)

type TileProjection struct {
	ID           string                  `json:"id"`
	X            int                     `json:"x"`
	Y            int                     `json:"y"`
	Terrain      string                  `json:"terrain"`
	River        *bool                   `json:"river,omitempty"`
	GoodID       *string                 `json:"good_id,omitempty"`
	Ruin         *bool                   `json:"ruin,omitempty"`
	OwnerID      *string                 `json:"owner_id,omitempty"`
	Yields       *engine.TileYield       `json:"yields,omitempty"`
	YieldContext string                  `json:"yield_context,omitempty"`
	MovementCost *int                    `json:"movement_cost,omitempty"`
	Pollution    int                     `json:"pollution,omitempty"`
	Stale        bool                    `json:"stale"`
	Observed     int                     `json:"observed_turn"`
	Planning     *TilePlanning           `json:"planning,omitempty"`
	Overlays     []TileOverlayProjection `json:"overlays,omitempty"`
}

type TileOverlayProjection struct {
	Kind          string `json:"kind"`
	ImprovementID string `json:"improvement_id,omitempty"`
	Remaining     int    `json:"remaining,omitempty"`
}

type TilePlanning struct {
	Cost       int      `json:"cost"`
	Domains    []string `json:"domains"`
	Accessible bool     `json:"accessible"`
	EnemyZOC   bool     `json:"enemy_zoc"`
}

type UnitProjection struct {
	ID                 string          `json:"id"`
	TypeID             string          `json:"type_id,omitempty"`
	OwnerID            string          `json:"owner_id"`
	TileID             string          `json:"tile_id"`
	TransportID        string          `json:"transport_id,omitempty"`
	ArmyID             string          `json:"army_id,omitempty"`
	Fortified          bool            `json:"fortified"`
	Asleep             bool            `json:"asleep,omitempty"`
	Veteran            bool            `json:"veteran,omitempty"`
	Domain             string          `json:"domain,omitempty"`
	Movement           int             `json:"movement,omitempty"`
	MaxMovement        int             `json:"max_movement,omitempty"`
	CargoCapacity      int             `json:"cargo_capacity,omitempty"`
	CargoDomains       []string        `json:"cargo_domains,omitempty"`
	CargoClasses       []string        `json:"cargo_classes,omitempty"`
	Size               string          `json:"size,omitempty"`
	Upkeep             int             `json:"upkeep,omitempty"`
	ShieldHunger       int             `json:"shield_hunger,omitempty"`
	FoodHunger         int             `json:"food_hunger,omitempty"`
	NeedsNoSupport     bool            `json:"needs_no_support,omitempty"`
	IsProfessional     bool            `json:"is_professional,omitempty"`
	IsSpecialForces    bool            `json:"is_special_forces,omitempty"`
	ZOC                bool            `json:"zoc,omitempty"`
	Attack             int             `json:"attack,omitempty"`
	Defense            int             `json:"defense,omitempty"`
	Ranged             int             `json:"ranged,omitempty"`
	BombardRange       int             `json:"bombard_range,omitempty"`
	BombRounds         int             `json:"bomb_rounds,omitempty"`
	ActiveDefenseRange int             `json:"active_defense_range,omitempty"`
	Strength           int             `json:"strength,omitempty"`
	Armor              int             `json:"armor,omitempty"`
	Firepower          int             `json:"firepower,omitempty"`
	Health             int             `json:"health,omitempty"`
	MaxHealth          int             `json:"max_health,omitempty"`
	VisibilityRange    int             `json:"visibility_range,omitempty"`
	InSpace            bool            `json:"in_space,omitempty"`
	ReentryTileID      string          `json:"reentry_tile_id,omitempty"`
	ReentryTurns       int             `json:"reentry_turns,omitempty"`
	SettleCityTypeID   string          `json:"settle_city_type_id,omitempty"`
	Effects            []engine.Effect `json:"effects,omitempty"`
}

type CityProjection struct {
	ID                  string                  `json:"id"`
	Name                string                  `json:"name,omitempty"`
	TypeID              string                  `json:"type_id,omitempty"`
	Art                 *engine.CityArt         `json:"art,omitempty"`
	OwnerID             string                  `json:"owner_id"`
	TileID              string                  `json:"tile_id"`
	Population          int                     `json:"population,omitempty"`
	Happiness           int                     `json:"happiness,omitempty"`
	CrimePercent        int                     `json:"crime_percent,omitempty"`
	OvercrowdingLevel   int                     `json:"overcrowding_level,omitempty"`
	ConvertedToPlayerID string                  `json:"converted_to_player_id,omitempty"`
	ConvertedBy         string                  `json:"converted_by,omitempty"`
	Rioting             bool                    `json:"rioting,omitempty"`
	Specialists         map[string]int          `json:"specialists,omitempty"`
	Goods               []string                `json:"goods,omitempty"`
	Effects             []engine.Effect         `json:"effects,omitempty"`
	StoredFood          int                     `json:"stored_food,omitempty"`
	StoredProduction    int                     `json:"stored_production,omitempty"`
	Queue               []engine.ProductionItem `json:"queue,omitempty"`
	Buildings           []engine.Building       `json:"buildings,omitempty"`
	Yields              *engine.Yield           `json:"yields,omitempty"`
	RushBuyQuote        *engine.RushBuyQuote    `json:"rush_buy_quote,omitempty"`
	ProductionForecast  *engine.Forecast        `json:"production_forecast,omitempty"`
	Governor            *engine.GovernorState   `json:"governor,omitempty"`
	Stale               bool                    `json:"stale"`
	Observed            int                     `json:"observed_turn"`
}

type ResearchProjection struct {
	Advances  []string         `json:"advances,omitempty"`
	TargetID  string           `json:"target_id,omitempty"`
	GoalID    string           `json:"goal_id,omitempty"`
	Progress  int              `json:"progress,omitempty"`
	Available []string         `json:"available,omitempty"`
	Forecast  *engine.Forecast `json:"forecast,omitempty"`
}

type TradeRouteProjection struct {
	ID                string  `json:"id"`
	SourceCityID      string  `json:"source_city_id"`
	DestinationCityID string  `json:"destination_city_id"`
	CrossesWater      bool    `json:"crosses_water,omitempty"`
	PiratingArmyID    string  `json:"pirating_army_id,omitempty"`
	TransportCost     int     `json:"transport_cost,omitempty"`
	GoodID            *string `json:"good_id,omitempty"`
	Gold              *int    `json:"gold,omitempty"`
}

type DiplomacyRelationProjection struct {
	TargetPlayerID string `json:"target_player_id"`
	Treaty         string `json:"treaty,omitempty"`
	Trade          bool   `json:"trade,omitempty"`
	Regard         string `json:"regard,omitempty"`
}

type DiplomacyProposalProjection struct {
	ID           string `json:"id"`
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
	Kind         string `json:"kind"`
}

type DiplomacyProjection struct {
	GovernmentID             string                        `json:"government_id,omitempty"`
	PublicWorks              int                           `json:"public_works,omitempty"`
	Gold                     int                           `json:"gold,omitempty"`
	SciencePercent           *int                          `json:"science_percent,omitempty"`
	PublicWorksPercent       *int                          `json:"public_works_percent,omitempty"`
	Workday                  *int                          `json:"workday,omitempty"`
	Wages                    *int                          `json:"wages,omitempty"`
	Rations                  *int                          `json:"rations,omitempty"`
	Readiness                string                        `json:"readiness,omitempty"`
	TradeTransportPoints     int                           `json:"trade_transport_points,omitempty"`
	UsedTradeTransportPoints int                           `json:"used_trade_transport_points,omitempty"`
	Relations                []DiplomacyRelationProjection `json:"relations,omitempty"`
	Proposals                []DiplomacyProposalProjection `json:"proposals,omitempty"`
}

type RulesProjection struct {
	ContentHash         string                                      `json:"content_hash"`
	GovernorPriorities  []string                                    `json:"governor_priorities,omitempty"`
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
	Revision         int64                     `json:"revision"`
	Turn             int                       `json:"turn"`
	Tiles            []TileProjection          `json:"tiles"`
	Units            []UnitProjection          `json:"units"`
	Cities           []CityProjection          `json:"cities"`
	TradeRoutes      []TradeRouteProjection    `json:"trade_routes,omitempty"`
	Research         *ResearchProjection       `json:"research,omitempty"`
	Diplomacy        *DiplomacyProjection      `json:"diplomacy,omitempty"`
	Rules            *RulesProjection          `json:"rules,omitempty"`
	CityReports      []engine.CityReport       `json:"city_reports,omitempty"`
	Notifications    []engine.Notification     `json:"notifications,omitempty"`
	MovementOrders   []engine.MovementOrder    `json:"movement_orders,omitempty"`
	Statistics       []engine.StatisticsSample `json:"statistics,omitempty"`
	LegalActions     []ActionProjection        `json:"legal_actions,omitempty"`
	ActionIntents    []ActionProjection        `json:"action_intents,omitempty"`
	Contract         *ContractProjection       `json:"contract"`
	EmbassyPlayerIDs []string                  `json:"embassy_player_ids,omitempty"`
	PlanningStatus   string                    `json:"planning_status,omitempty"`
	World            *WorldProjection          `json:"world,omitempty"`
	Calendar         *engine.CalendarDisplay   `json:"calendar,omitempty"`
}

// Update records only what each player can currently observe. The returned state
// remains the single canonical authority and is safe to persist/reload unchanged.
func Update(state engine.State) (engine.State, error) {
	if err := engine.ValidateState(state); err != nil {
		return engine.State{}, err
	}
	inspector, err := engine.NewInspector(state)
	if err != nil {
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
			observation := engine.FogObservation{PlayerID: playerID, TileID: tile.ID, Turn: state.Phase.Turn, Terrain: engine.EffectiveTerrain(state, tile.ID), River: copyValue(&tile.River), GoodID: copyValue(&tile.GoodID), Ruin: ruinPresence(tile), OwnerID: copyValue(&tile.OwnerID), Inspection: inspector.Tile(tile.ID)}
			if city, ok := cities[tile.ID]; ok {
				observation.CityID = city.ID
				observation.CityOwnerID = city.OwnerID
				observation.CityName = city.Name
				observation.CityTypeID = city.TypeID
				observation.CityPopulation = copyValue(&city.Population)
				observation.CityArt = inspector.CityArt(city.ID)
				observation.CityEffects = append([]engine.Effect(nil), city.Effects...)
			}
			memory[playerID+"\x00"+tileID] = observation
		}
	}
	state.Fog = make([]engine.FogObservation, 0, len(memory))
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
	state, err := engine.WithQueueIdentities(state)
	if err != nil {
		return Projection{}, err
	}
	inspector, err := engine.NewInspector(state)
	if err != nil {
		return Projection{}, err
	}
	visible := visibleTiles(state)[playerID]
	memory := memoryByTile(state.Fog, playerID)
	planning, err := planningTiles(state, playerID, visible, inspector)
	if err != nil {
		return Projection{}, err
	}
	result := Projection{Revision: state.Revision, Turn: state.Phase.Turn, PlanningStatus: planningStatus(state, playerID), Rules: rulesProjection(state), World: worldProjection(state), Calendar: inspector.Calendar()}
	viewerNationID := nationForPlayer(state, playerID)
	for _, nation := range state.Nations {
		if nation.ID != viewerNationID {
			continue
		}
		result.Research = researchProjection(state, nation)
		if result.Research == nil {
			result.Research = &ResearchProjection{}
		}
		result.Research.Forecast = inspector.Research(playerID)
		result.Diplomacy = diplomacyProjection(state, playerID, nation)
		result.CityReports = cloneCityReports(nation.CityReports)
		result.Statistics = append([]engine.StatisticsSample(nil), nation.Statistics...)
		for _, embassyPlayer := range state.Players {
			if embassyPlayer.ID != playerID && engine.HasEffectiveEmbassyWith(state, playerID, embassyPlayer.ID) {
				result.EmbassyPlayerIDs = append(result.EmbassyPlayerIDs, embassyPlayer.ID)
			}
		}
	}
	for _, order := range state.MovementOrders {
		if order.PlayerID == playerID {
			order.UnitIDs = append([]string(nil), order.UnitIDs...)
			result.MovementOrders = append(result.MovementOrders, order)
		}
	}
	for _, notification := range state.Notifications {
		if notification.PlayerID == playerID {
			result.Notifications = append(result.Notifications, notification)
		}
	}
	for _, tile := range state.Tiles {
		if visible[tile.ID] {
			projected := TileProjection{ID: tile.ID, X: tile.X, Y: tile.Y, Terrain: engine.EffectiveTerrain(state, tile.ID), River: copyValue(&tile.River), GoodID: copyValue(&tile.GoodID), Ruin: ruinPresence(tile), OwnerID: copyValue(&tile.OwnerID), Observed: state.Phase.Turn, Planning: planning[tile.ID]}
			applyInspection(&projected, inspector.Tile(tile.ID))
			result.Tiles = append(result.Tiles, projected)
			continue
		}
		if observation, ok := memory[tile.ID]; ok {
			projected := TileProjection{ID: tile.ID, X: tile.X, Y: tile.Y, Terrain: observation.Terrain, River: copyValue(observation.River), GoodID: copyValue(observation.GoodID), Ruin: copyValue(observation.Ruin), OwnerID: copyValue(observation.OwnerID), Stale: true, Observed: observation.Turn}
			applyInspection(&projected, observation.Inspection)
			result.Tiles = append(result.Tiles, projected)
			if observation.CityID != "" {
				result.Cities = append(result.Cities, CityProjection{ID: observation.CityID, Name: observation.CityName, TypeID: observation.CityTypeID, Art: copyValue(observation.CityArt), Population: knownPopulation(observation.CityPopulation), OwnerID: observation.CityOwnerID, TileID: tile.ID, Effects: append([]engine.Effect(nil), observation.CityEffects...), Stale: true, Observed: observation.Turn})
			}
		}
	}
	for _, unit := range state.Units {
		if unitVisibleToPlayer(state, playerID, visible, unit) {
			projected := UnitProjection{ID: unit.ID, TypeID: unit.TypeID, OwnerID: unit.OwnerID, TileID: unit.TileID, TransportID: unit.TransportID, ArmyID: unit.ArmyID, Fortified: unit.Fortified}
			if unit.OwnerID != playerID {
				projected.Effects = publicTargetEffects(unit.Effects)
			}
			if unit.OwnerID == playerID {
				projected.Domain, projected.Movement, projected.MaxMovement = unit.Domain, unit.Movement, unit.MaxMovement
				projected.Veteran = unit.Veteran
				projected.Asleep = unit.Asleep
				projected.CargoCapacity = unit.CargoCapacity
				projected.CargoDomains = append([]string(nil), unit.CargoDomains...)
				projected.CargoClasses = append([]string(nil), unit.CargoClasses...)
				projected.Size = unit.Size
				projected.Upkeep = unit.Upkeep
				projected.ShieldHunger = unit.ShieldHunger
				projected.FoodHunger = unit.FoodHunger
				projected.NeedsNoSupport = unit.NeedsNoSupport
				projected.IsProfessional = unit.IsProfessional
				projected.IsSpecialForces = unit.IsSpecialForces
				projected.ZOC = unit.ZOC
				projected.Attack, projected.Defense, projected.Ranged, projected.BombardRange, projected.BombRounds, projected.ActiveDefenseRange, projected.Strength, projected.Armor, projected.Firepower = unit.Attack, unit.Defense, unit.Ranged, unit.BombardRange, unit.BombRounds, unit.ActiveDefenseRange, unit.Strength, unit.Armor, unit.Firepower
				projected.Health, projected.MaxHealth = unit.Health, unit.MaxHealth
				projected.VisibilityRange = unit.VisibilityRange
				projected.InSpace = unit.InSpace
				projected.ReentryTileID = unit.ReentryTileID
				projected.ReentryTurns = unit.ReentryTurns
				projected.SettleCityTypeID = unit.SettleCityTypeID
				projected.Effects = append([]engine.Effect(nil), unit.Effects...)
			}
			result.Units = append(result.Units, projected)
		}
	}
	for _, city := range state.Cities {
		if visible[city.TileID] {
			projected := projectCity(city, playerID, state.Phase.Turn)
			projected.Art = inspector.CityArt(city.ID)
			projected.RushBuyQuote = inspector.RushBuy(city.ID, playerID)
			projected.ProductionForecast = inspector.Production(city.ID, playerID)
			result.Cities = append(result.Cities, projected)
		}
	}
	result.TradeRoutes = tradeRouteProjection(state, playerID, visible)
	sort.Slice(result.Tiles, func(i, j int) bool { return result.Tiles[i].ID < result.Tiles[j].ID })
	sort.Slice(result.Units, func(i, j int) bool { return result.Units[i].ID < result.Units[j].ID })
	sort.Slice(result.Cities, func(i, j int) bool { return result.Cities[i].ID < result.Cities[j].ID })
	sort.Slice(result.TradeRoutes, func(i, j int) bool { return result.TradeRoutes[i].ID < result.TradeRoutes[j].ID })
	sort.Strings(result.EmbassyPlayerIDs)
	result.LegalActions = legalActions(state, result, playerID)
	result.ActionIntents = movementIntents(state, result, playerID)
	result.Contract = contractProjection(result)
	return result, nil
}

func tradeRouteProjection(state engine.State, playerID string, visible map[string]bool) []TradeRouteProjection {
	projection := make([]TradeRouteProjection, 0)
	for _, route := range state.TradeRoutes {
		source, sourceOK := cityByID(state.Cities, route.SourceCityID)
		destination, destinationOK := cityByID(state.Cities, route.DestinationCityID)
		if !sourceOK || !destinationOK || !visible[source.TileID] || !visible[destination.TileID] {
			continue
		}
		projected := TradeRouteProjection{ID: route.ID, SourceCityID: route.SourceCityID, DestinationCityID: route.DestinationCityID, CrossesWater: route.CrossesWater, PiratingArmyID: route.PiratingArmyID, TransportCost: route.TransportCost}
		// Sight of both cities reveals the route, not its private accounting.
		if source.OwnerID == playerID || destination.OwnerID == playerID {
			projected.GoodID = copyValue(&route.GoodID)
			projected.Gold = copyValue(&route.Gold)
		}
		projection = append(projection, projected)
	}
	return projection
}

func diplomacyProjection(state engine.State, playerID string, nation engine.Nation) *DiplomacyProjection {
	projection := &DiplomacyProjection{GovernmentID: nation.GovernmentID, PublicWorks: nation.PublicWorks, Gold: nation.Gold, Readiness: nation.Readiness, TradeTransportPoints: nation.TradeTransportPoints, UsedTradeTransportPoints: projectedUsedTradeTransportPoints(state, nation.ID)}
	if nation.SciencePercent != nil {
		science := *nation.SciencePercent
		projection.SciencePercent = &science
	}
	if nation.PublicWorksPercent != nil {
		publicWorks := *nation.PublicWorksPercent
		projection.PublicWorksPercent = &publicWorks
	}
	if nation.Workday != nil {
		workday := *nation.Workday
		projection.Workday = &workday
	}
	if nation.Wages != nil {
		wages := *nation.Wages
		projection.Wages = &wages
	}
	if nation.Rations != nil {
		rations := *nation.Rations
		projection.Rations = &rations
	}
	for _, relation := range state.Diplomacy {
		if relation.PlayerAID != playerID && relation.PlayerBID != playerID || !relation.Contact {
			continue
		}
		target := relation.PlayerAID
		if target == playerID {
			target = relation.PlayerBID
		}
		projection.Relations = append(projection.Relations, DiplomacyRelationProjection{TargetPlayerID: target, Treaty: relation.Treaty, Trade: relation.Trade, Regard: diplomaticRegard(state, playerID, target, relation)})
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
	freightModelEnabled := false
	if runtime, ok := runtimeRules(state); ok && runtime.Constants.CaravanCoefPercent > 0 {
		freightModelEnabled = true
	}
	if projection.GovernmentID == "" && projection.PublicWorks == 0 && projection.Gold == 0 && projection.TradeTransportPoints == 0 && projection.UsedTradeTransportPoints == 0 && !freightModelEnabled && projection.SciencePercent == nil && projection.PublicWorksPercent == nil && projection.Workday == nil && projection.Wages == nil && projection.Rations == nil && projection.Readiness == "" && len(projection.Relations) == 0 && len(projection.Proposals) == 0 {
		return nil
	}
	return projection
}

func diplomaticRegard(state engine.State, playerID, target string, relation engine.DiplomaticRelation) string {
	left, right := playerID, target
	if right < left {
		left, right = right, left
	}
	for _, war := range state.Policy.Wars {
		if war.A == left && war.B == right {
			return "hotwar"
		}
	}
	if relation.Treaty == "union" {
		return "love"
	}
	if directionalRegardBonus(playerID, target, relation) > 0 || engine.HasActiveWonderEffect(state, target, engine.EffectIncreaseRegard) {
		return "friendly"
	}
	return "neutral"
}

func directionalRegardBonus(playerID, target string, relation engine.DiplomaticRelation) int {
	if playerID == relation.PlayerAID && target == relation.PlayerBID {
		return relation.RegardBonusAToB
	}
	if playerID == relation.PlayerBID && target == relation.PlayerAID {
		return relation.RegardBonusBToA
	}
	return 0
}

func projectCity(city engine.City, playerID string, observed int) CityProjection {
	projection := CityProjection{ID: city.ID, Name: city.Name, TypeID: city.TypeID, Population: city.Population, OwnerID: city.OwnerID, TileID: city.TileID, Effects: append([]engine.Effect(nil), city.Effects...), Observed: observed}
	if city.OwnerID == playerID || city.ConvertedToPlayerID == playerID {
		projection.ConvertedToPlayerID = city.ConvertedToPlayerID
		projection.ConvertedBy = city.ConvertedBy
	}
	if city.OwnerID == playerID {
		projection.Governor = copyValue(city.Governor)
		projection.Population = city.Population
		projection.Happiness = city.Happiness
		projection.CrimePercent = city.CrimePercent
		projection.OvercrowdingLevel = city.OvercrowdingLevel
		projection.Rioting = city.Rioting
		if city.Specialists != nil {
			projection.Specialists = make(map[string]int, len(city.Specialists))
			for kind, count := range city.Specialists {
				projection.Specialists[kind] = count
			}
		}
		projection.Goods = append([]string(nil), city.Goods...)
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
	projection := &ResearchProjection{Advances: append([]string(nil), nation.Advances...), TargetID: nation.ResearchTargetID, GoalID: nation.ResearchGoalID, Progress: nation.ResearchProgress}
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
		GovernorPriorities:  engine.GovernorPriorities(&runtime),
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

func cloneCityReports(reports []engine.CityReport) []engine.CityReport {
	cloned := make([]engine.CityReport, 0, len(reports))
	for _, report := range reports {
		copy := report
		copy.Buildings = append([]string(nil), report.Buildings...)
		copy.Effects = append([]engine.Effect(nil), report.Effects...)
		cloned = append(cloned, copy)
	}
	return cloned
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
		Strategies:         cloneRuntimeAIStrategies(ai.Strategies),
	}
}

func cloneRuntimeAIStrategies(strategies []rules.RuntimeAIStrategy) []rules.RuntimeAIStrategy {
	cloned := make([]rules.RuntimeAIStrategy, 0, len(strategies))
	for _, strategy := range strategies {
		cloned = append(cloned, strategy)
	}
	return cloned
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
	return engine.CanResearchAdvance(nation, runtime, advanceID)
}

func legalActions(state engine.State, projection Projection, playerID string) []ActionProjection {
	if !viewerCanAct(state, projection, playerID) {
		return nil
	}
	actions := make([]ActionProjection, 0)
	if projection.PlanningStatus == PlanningStatusOpen || projection.PlanningStatus == PlanningStatusQueued {
		actions = append(actions, action(engine.CommandRoundComplete, "round-complete", playerID, "", map[string]any{}, nil))
		if projection.PlanningStatus == PlanningStatusQueued {
			return actions
		}
	} else {
		actions = append(actions, action(engine.CommandPass, "pass", playerID, "", map[string]any{}, nil))
	}
	for _, order := range projection.MovementOrders {
		actions = append(actions, action(engine.CommandCancelMovementOrder, "cancel-movement-order", order.UnitID, "", map[string]string{"unit_id": order.UnitID}, []string{"unit_id"}))
	}
	actions = append(actions, researchActions(projection, playerID)...)
	actions = append(actions, researchGoalActions(state, playerID)...)
	actions = append(actions, governmentActions(projection, playerID)...)
	actions = append(actions, diplomacyActions(projection, playerID)...)
	actions = append(actions, tradeRouteActions(state, projection, playerID)...)
	actions = append(actions, specialistActions(projection, playerID)...)
	actions = append(actions, productionActions(state, projection, playerID)...)
	actions = append(actions, governorActions(projection, playerID)...)
	actions = append(actions, publicWorksTerrainActions(projection, playerID)...)
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

// movementIntents authorizes input shape, not hidden legality or predicted
// outcomes. The client supplies a public coordinate and may animate provisionally;
// the normal engine revalidates the entire move and receipts remain authoritative.
func movementIntents(state engine.State, projection Projection, playerID string) []ActionProjection {
	if !viewerCanAct(state, projection, playerID) || projection.PlanningStatus == PlanningStatusQueued || projection.Rules == nil || projection.Rules.Movement == nil || !projection.Rules.Movement.Configured {
		return nil
	}
	intents := []ActionProjection{}
	for _, unit := range projection.Units {
		if unit.OwnerID != playerID || unit.TileID == "" || unit.TransportID != "" || unit.InSpace || unit.Domain == "" {
			continue
		}
		intents = append(intents, action(engine.CommandSetMovementOrder, "set-movement-order", unit.ID, "", map[string]string{"unit_id": unit.ID}, []string{"unit_id", "target_x", "target_y"}))
		if unit.Movement > 0 {
			intents = append(intents, action(engine.CommandMoveUnit, "move", unit.ID, "", map[string]string{"unit_id": unit.ID}, []string{"unit_id", "target_x", "target_y"}))
		}
	}
	return intents
}

func researchGoalActions(state engine.State, playerID string) []ActionProjection {
	runtime, ok := runtimeRules(state)
	if !ok {
		return nil
	}
	var nation engine.Nation
	found := false
	for _, candidate := range state.Nations {
		if candidate.ID == nationForPlayer(state, playerID) {
			nation = candidate
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	actions := []ActionProjection{}
	if nation.ResearchGoalID != "" {
		actions = append(actions, action(engine.CommandSetResearchGoal, "set-research-goal", playerID, "", map[string]string{"advance_id": ""}, []string{"advance_id"}))
	}
	for _, advance := range runtime.Advances {
		if advance.ID != nation.ResearchGoalID && engine.CanSetResearchGoal(nation, runtime, advance.ID) {
			actions = append(actions, action(engine.CommandSetResearchGoal, "set-research-goal", playerID, advance.ID, map[string]string{"advance_id": advance.ID}, []string{"advance_id"}))
		}
	}
	return actions
}

func viewerCanAct(state engine.State, projection Projection, playerID string) bool {
	if state.Terminal.Ended || state.Phase.ResolutionLocked || playerID == "" {
		return false
	}
	eligible := false
	for _, player := range state.Players {
		if player.ID == playerID && player.Lifecycle == engine.PlayerActive {
			eligible = true
			break
		}
	}
	if !eligible {
		return false
	}
	switch projection.PlanningStatus {
	case PlanningStatusOpen, PlanningStatusQueued:
		return true
	case PlanningStatusClosed:
		return false
	}
	if len(state.Phase.ActivePlayerOrder) == 0 || state.Phase.ActivePlayerIndex < 0 || state.Phase.ActivePlayerIndex >= len(state.Phase.ActivePlayerOrder) {
		return false
	}
	if state.Phase.Mode == turns.ModeHybrid && state.Policy.Stage == turns.StageConflict {
		index := state.Policy.ComponentIndex
		if index < 0 || index >= len(state.Policy.Components) || !containsString(state.Policy.Components[index], playerID) {
			return false
		}
	}
	return state.Phase.ActivePlayerOrder[state.Phase.ActivePlayerIndex] == playerID
}

func projectedAIStrategySciencePercent(projection Projection) (int, bool) {
	if projection.Rules == nil || projection.Rules.AI == nil {
		return 0, false
	}
	for _, strategy := range projection.Rules.AI.Strategies {
		if strategy.SciencePercent > 0 {
			return clampPercent(strategy.SciencePercent), true
		}
	}
	return 0, false
}

func projectedAIStrategyPublicWorksPercent(projection Projection) (int, bool) {
	if projection.Rules == nil || projection.Rules.AI == nil {
		return 0, false
	}
	for _, strategy := range projection.Rules.AI.Strategies {
		if strategy.PublicWorksPercent > 0 {
			return clampPercent(strategy.PublicWorksPercent), true
		}
	}
	return 0, false
}

func projectedAIStrategySocialPolicy(projection Projection, playerID string) (workday int, wages int, rations int, ok bool) {
	if projection.Rules == nil || projection.Rules.AI == nil {
		return 0, 0, 0, false
	}
	expectWorkday, expectWages, expectRations := projectedGovernmentExpectations(projection)
	for _, strategy := range projection.Rules.AI.Strategies {
		for _, element := range strategy.SliderElements {
			if element.Delta == 0 {
				continue
			}
			deltaProduction, deltaGold, deltaFood := 0, 0, 0
			switch element.Kind {
			case "Production":
				deltaProduction, deltaGold, deltaFood = element.Delta, -element.Delta, -element.Delta
			case "Gold":
				deltaProduction, deltaGold, deltaFood = -element.Delta, element.Delta, -element.Delta
			case "Food":
				deltaProduction, deltaGold, deltaFood = -element.Delta, -element.Delta, element.Delta
			default:
				continue
			}
			candidateWorkday := clampSocialPolicy(expectWorkday - deltaProduction)
			candidateWages := clampSocialPolicy(expectWages - deltaGold)
			candidateRations := clampSocialPolicy(expectRations - deltaFood)
			if !strategySocialPolicyFitsVisibleCities(projection, playerID, candidateWorkday, candidateWages, candidateRations) {
				continue
			}
			return candidateWorkday, candidateWages, candidateRations, true
		}
	}
	return 0, 0, 0, false
}

func strategySocialPolicyFitsVisibleCities(projection Projection, playerID string, workday int, wages int, rations int) bool {
	checked := false
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.Stale || city.Yields == nil {
			continue
		}
		checked = true
		population := city.Population
		if population < 1 {
			population = 1
		}
		production := city.Yields.Production
		if production > 0 {
			percent := 100 - (workday * 15)
			if percent < 0 {
				percent = 0
			}
			production = (production*percent + 99) / 100
		}
		if len(city.Queue) > 0 && production < 1 {
			return false
		}
		wagePerCitizen := 4 + wages
		if wagePerCitizen < 0 {
			wagePerCitizen = 0
		}
		if city.Yields.Gold-population*wagePerCitizen < 0 {
			return false
		}
		centsPerCitizen := 200 + (rations * 50)
		if centsPerCitizen < 0 {
			centsPerCitizen = 0
		}
		consumption := (population*centsPerCitizen + 99) / 100
		if city.StoredFood+city.Yields.Food-consumption < 0 {
			return false
		}
	}
	return checked
}

func projectedGovernmentExpectations(projection Projection) (workday int, wages int, rations int) {
	if projection.Rules == nil || projection.Diplomacy == nil {
		return 0, 0, 0
	}
	for _, government := range projection.Rules.Governments {
		if government.ID != projection.Diplomacy.GovernmentID {
			continue
		}
		return government.WorkdayExpectation, government.WagesExpectation, government.RationsExpectation
	}
	return 0, 0, 0
}

func appendUniqueInt(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func clampSocialPolicy(value int) int {
	if value < -2 {
		return -2
	}
	if value > 2 {
		return 2
	}
	return value
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
	currentSciencePercent := 100
	if projection.Diplomacy.SciencePercent != nil {
		currentSciencePercent = *projection.Diplomacy.SciencePercent
	}
	sciencePercents := []int{0, 50, 100}
	if strategyScience, ok := projectedAIStrategySciencePercent(projection); ok {
		sciencePercents = appendUniqueInt(sciencePercents, strategyScience)
		sort.Ints(sciencePercents)
	}
	for _, percent := range sciencePercents {
		if currentSciencePercent == percent {
			continue
		}
		actions = append(actions, action(engine.CommandSetScienceRate, "set-science-rate", playerID, strconv.Itoa(percent), map[string]int{"science_percent": percent}, []string{"science_percent"}))
	}
	currentPublicWorksPercent := 0
	if projection.Diplomacy.PublicWorksPercent != nil {
		currentPublicWorksPercent = *projection.Diplomacy.PublicWorksPercent
	}
	if strategyPublicWorks, ok := projectedAIStrategyPublicWorksPercent(projection); ok && currentPublicWorksPercent != strategyPublicWorks {
		actions = append(actions, action(engine.CommandSetPublicWorksRate, "set-public-works-rate", playerID, strconv.Itoa(strategyPublicWorks), map[string]int{"public_works_percent": strategyPublicWorks}, []string{"public_works_percent"}))
	}
	workday, wages, rations := 0, 0, 0
	if projection.Diplomacy.Workday != nil {
		workday = *projection.Diplomacy.Workday
	}
	if projection.Diplomacy.Wages != nil {
		wages = *projection.Diplomacy.Wages
	}
	if projection.Diplomacy.Rations != nil {
		rations = *projection.Diplomacy.Rations
	}
	if strategyWorkday, strategyWages, strategyRations, ok := projectedAIStrategySocialPolicy(projection, playerID); ok {
		workday, wages, rations = strategyWorkday, strategyWages, strategyRations
	}
	actions = append(actions, action(engine.CommandSetSocialPolicy, "set-social-policy", playerID, "social-policy", map[string]int{"workday": workday, "wages": wages, "rations": rations}, []string{"workday", "wages", "rations"}))
	currentReadiness := projection.Diplomacy.Readiness
	if currentReadiness == "" {
		currentReadiness = "war"
	}
	for _, readiness := range []string{"peace", "alert", "war"} {
		if readiness == currentReadiness {
			continue
		}
		actions = append(actions, action(engine.CommandSetReadiness, "set-readiness", playerID, readiness, map[string]string{"readiness": readiness}, []string{"readiness"}))
	}
	for _, government := range projection.Rules.Governments {
		if government.ID == "" || government.ID == projection.Diplomacy.GovernmentID || !advanceSatisfied(known, government.EnableAdvanceID) {
			continue
		}
		actions = append(actions, action(engine.CommandSetGovernment, "set-government", playerID, government.ID, map[string]string{"government_id": government.ID}, []string{"government_id"}))
	}
	return actions
}

func diplomacyActions(projection Projection, playerID string) []ActionProjection {
	contacted := map[string]bool{}
	if projection.Diplomacy != nil {
		for _, relation := range projection.Diplomacy.Relations {
			if relation.TargetPlayerID != "" {
				contacted[relation.TargetPlayerID] = true
			}
		}
	}
	actions := make([]ActionProjection, 0)
	for _, target := range visibleForeignOwners(projection, playerID) {
		if contacted[target] {
			continue
		}
		actions = append(actions, action(engine.CommandEstablishContact, "establish-contact", playerID, target, map[string]string{"target_player_id": target}, []string{"target_player_id"}))
	}
	if projection.Diplomacy == nil {
		return actions
	}
	pending := make(map[string]bool, len(projection.Diplomacy.Proposals))
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
		if relation.TargetPlayerID == "" {
			continue
		}
		if relation.Regard == "hotwar" {
			if !pending["peace\x00"+playerID+"\x00"+relation.TargetPlayerID] {
				proposalID := "client-peace-" + playerID + "-" + relation.TargetPlayerID
				actions = append(actions, action(engine.CommandProposeAgreement, "propose-agreement", playerID, relation.TargetPlayerID, map[string]string{"proposal_id": proposalID, "target_player_id": relation.TargetPlayerID, "kind": "peace"}, []string{"proposal_id", "target_player_id", "kind"}))
			}
			continue
		}
		if !relation.Trade && relation.Treaty != "trade" && relation.Treaty != "union" && !pending["trade\x00"+playerID+"\x00"+relation.TargetPlayerID] {
			proposalID := "client-trade-" + playerID + "-" + relation.TargetPlayerID
			actions = append(actions, action(engine.CommandProposeAgreement, "propose-agreement", playerID, relation.TargetPlayerID, map[string]string{"proposal_id": proposalID, "target_player_id": relation.TargetPlayerID, "kind": "trade"}, []string{"proposal_id", "target_player_id", "kind"}))
		}
		if relation.Treaty == "union" || pending["union\x00"+playerID+"\x00"+relation.TargetPlayerID] {
			continue
		}
		proposalID := "client-union-" + playerID + "-" + relation.TargetPlayerID
		actions = append(actions, action(engine.CommandProposeAgreement, "propose-agreement", playerID, relation.TargetPlayerID, map[string]string{"proposal_id": proposalID, "target_player_id": relation.TargetPlayerID, "kind": "union"}, []string{"proposal_id", "target_player_id", "kind"}))
	}
	return actions
}

func visibleForeignOwners(projection Projection, playerID string) []string {
	owners := map[string]bool{}
	for _, city := range projection.Cities {
		if city.OwnerID != "" && city.OwnerID != playerID {
			owners[city.OwnerID] = true
		}
	}
	for _, unit := range projection.Units {
		if unit.OwnerID != "" && unit.OwnerID != playerID {
			owners[unit.OwnerID] = true
		}
	}
	result := make([]string, 0, len(owners))
	for owner := range owners {
		result = append(result, owner)
	}
	sort.Strings(result)
	return result
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

func tradeRouteActions(state engine.State, projection Projection, playerID string) []ActionProjection {
	existing := make(map[string]bool, len(state.TradeRoutes))
	for _, route := range state.TradeRoutes {
		source, ok := cityByID(state.Cities, route.SourceCityID)
		if ok && source.OwnerID == playerID {
			existing[route.SourceCityID+"\x00"+route.DestinationCityID] = true
		}
	}
	owned := make([]CityProjection, 0)
	destinations := make([]CityProjection, 0)
	for _, city := range projection.Cities {
		if city.ID == "" || city.Stale {
			continue
		}
		if city.OwnerID == playerID {
			owned = append(owned, city)
		}
		destinations = append(destinations, city)
	}
	maxOutgoing, maxIncoming := projectedTradeRouteCaps(projection)
	actions := make([]ActionProjection, 0)
	for _, source := range owned {
		if maxOutgoing > 0 && projectedTradeRouteOutgoingCount(state, source.ID) >= maxOutgoing {
			continue
		}
		for _, destination := range destinations {
			if source.ID == destination.ID || existing[source.ID+"\x00"+destination.ID] || !projectedTradeAgreementAllows(projection, playerID, destination.OwnerID) {
				continue
			}
			if maxIncoming > 0 && projectedTradeRouteIncomingCount(state, destination.ID) >= maxIncoming {
				continue
			}
			if !projectedTradeTransportAllows(state, projection, source, destination) {
				continue
			}
			actions = append(actions, action(engine.CommandCreateTradeRoute, "create-trade-route", source.ID, destination.ID, map[string]string{"source_city_id": source.ID, "destination_city_id": destination.ID}, []string{"source_city_id", "destination_city_id"}))
		}
	}
	return actions
}

func projectedUsedTradeTransportPoints(state engine.State, nationID string) int {
	used := 0
	for _, route := range state.TradeRoutes {
		source, ok := cityByID(state.Cities, route.SourceCityID)
		if !ok || nationForPlayer(state, source.OwnerID) != nationID {
			continue
		}
		used += route.TransportCost
	}
	return used
}

func projectedTradeTransportAllows(state engine.State, projection Projection, source, destination CityProjection) bool {
	if projection.Diplomacy == nil {
		return true
	}
	cost := projectedTradeRouteTransportCost(state, source, destination)
	if cost <= 0 {
		return true
	}
	return projection.Diplomacy.UsedTradeTransportPoints+cost <= projection.Diplomacy.TradeTransportPoints
}

func projectedTradeRouteTransportCost(state engine.State, source, destination CityProjection) int {
	runtime, ok := runtimeRules(state)
	if !ok || runtime.Constants.CaravanCoefPercent <= 0 {
		return 0
	}
	sourceTile, sourceOK := tileByID(state.Tiles, source.TileID)
	destinationTile, destinationOK := tileByID(state.Tiles, destination.TileID)
	if !sourceOK || !destinationOK {
		return 0
	}
	distance := abs(sourceTile.X-destinationTile.X) + abs(sourceTile.Y-destinationTile.Y)
	if distance < 1 {
		distance = 1
	}
	return (distance*100*runtime.Constants.CaravanCoefPercent + 50) / 100
}

func nationForPlayer(state engine.State, playerID string) string {
	for _, player := range state.Players {
		if player.ID == playerID {
			return player.NationID
		}
	}
	return ""
}

func projectedTradeRouteCaps(projection Projection) (maxOutgoing, maxIncoming int) {
	if projection.Diplomacy == nil || projection.Rules == nil || projection.Diplomacy.GovernmentID == "" {
		return 0, 0
	}
	for _, government := range projection.Rules.Governments {
		if government.ID == projection.Diplomacy.GovernmentID {
			return government.MaxOutgoingTrade, government.MaxIncomingTrade
		}
	}
	return 0, 0
}

func projectedTradeRouteOutgoingCount(state engine.State, cityID string) int {
	count := 0
	for _, route := range state.TradeRoutes {
		if route.SourceCityID == cityID {
			count++
		}
	}
	return count
}

func projectedTradeRouteIncomingCount(state engine.State, cityID string) int {
	count := 0
	for _, route := range state.TradeRoutes {
		if route.DestinationCityID == cityID {
			count++
		}
	}
	return count
}

func projectedTradeAgreementAllows(projection Projection, playerID, destinationOwnerID string) bool {
	if destinationOwnerID == playerID {
		return true
	}
	if projection.Diplomacy == nil {
		return false
	}
	for _, relation := range projection.Diplomacy.Relations {
		if relation.TargetPlayerID == destinationOwnerID && relation.Regard != "hotwar" && (relation.Trade || relation.Treaty == "trade" || relation.Treaty == "union") {
			return true
		}
	}
	return false
}

func specialistActions(projection Projection, playerID string) []ActionProjection {
	actions := make([]ActionProjection, 0)
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.Population <= 0 {
			continue
		}
		specialists := make(map[string]int, len(city.Specialists))
		for kind, count := range city.Specialists {
			specialists[kind] = count
		}
		actions = append(actions, action(engine.CommandSetSpecialists, "set-specialists", city.ID, city.ID, map[string]any{"city_id": city.ID, "specialists": specialists}, []string{"city_id", "specialists"}))
	}
	return actions
}

func productionActions(state engine.State, projection Projection, playerID string) []ActionProjection {
	inspector, err := engine.NewInspector(state)
	if err != nil {
		return nil
	}
	actions := make([]ActionProjection, 0)
	known := []string(nil)
	if projection.Research != nil {
		known = projection.Research.Advances
	}
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.Stale {
			continue
		}
		name := city.Name
		if name == "" {
			name = "City " + city.ID
		}
		actions = append(actions, action(engine.CommandRenameCity, "rename-city", city.ID, "", engine.RenameCityPayload{CityID: city.ID, Name: name}, []string{"city_id", "name"}))
		if inspector.CanDisband(city.ID, playerID) {
			actions = append(actions, action(engine.CommandDisbandCity, "disband-city", city.ID, "", engine.CityIdentityPayload{CityID: city.ID}, []string{"city_id"}))
		}
		entries := make([]engine.QueueSelection, 0, len(city.Queue))
		for _, queued := range city.Queue {
			entries = append(entries, engine.QueueSelection{EntryID: queued.EntryID})
		}
		actions = append(actions, action(engine.CommandEditProduction, "edit-production", city.ID, "", engine.EditProductionPayload{CityID: city.ID, Entries: entries}, []string{"city_id", "entries"}))
		for index, queued := range city.Queue {
			actions = append(actions, action(engine.CommandCancelProduction, "cancel-production", city.ID, queued.EntryID, map[string]string{"city_id": city.ID, "entry_id": queued.EntryID}, []string{"city_id", "entry_id"}))
			rushCost := projectedRushBuyCost(projection, city, queued)
			if index == 0 && rushCost > 0 && projection.Diplomacy != nil && projection.Diplomacy.Gold >= rushCost {
				actions = append(actions, action(engine.CommandRushBuy, "rush-buy", city.ID, queued.ID, map[string]string{"city_id": city.ID}, []string{"city_id"}))
			}
		}
		if len(city.Queue) != 0 || projection.Rules == nil {
			continue
		}
		built := cityBuildingIDs(city)
		for _, production := range projection.Rules.Production {
			if production.ID == "" || production.Kind == engine.ProductionUnit || !allAdvancesSatisfied(known, production.Prerequisites) || !allBuildingsSatisfied(built, production.RequiredBuildings) || !cityProductionPlacementAllowed(projection, city, production) || anyAdvanceSatisfied(known, production.ObsoleteBy) {
				continue
			}
			if production.Kind == engine.ProductionBuilding && built[production.ID] {
				continue
			}
			item := productionActionItem(production)
			actions = append(actions, action(engine.CommandQueueProduction, "queue-production", city.ID, production.ID, map[string]any{"city_id": city.ID, "item": item}, []string{"city_id", "item"}))
		}
		for _, unit := range projection.Rules.Units {
			if unit.ID == "" || runtimeUnitHasEffect(unit, engine.EffectCantBuild) || unitProductionBlockedByGovernment(projection, unit) || unitProductionBlockedByCityGood(city, unit) || unitProductionBlockedByEliminateNukes(state, playerID, unit) || unitProductionBlockedByExclusiveSlavery(projection, playerID, unit) || !allAdvancesSatisfied(known, unit.Prerequisites) || unitRemovesPopulation(unit) && city.Population <= 1 || anyAdvanceSatisfied(known, unit.ObsoleteBy) {
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

func projectedRushBuyCost(projection Projection, city CityProjection, item engine.ProductionItem) int {
	if item.Kind == engine.ProductionCapitalization || item.Kind == engine.ProductionInfrastructure {
		return 0
	}
	remaining := item.Cost - city.StoredProduction
	if remaining <= 0 {
		return 0
	}
	if item.Cost < 1 {
		return remaining
	}
	typeCoefficient := projectedRushBuyTypeCoefficient(projection, item)
	percentRemaining := float64(remaining) / float64(item.Cost)
	return int(typeCoefficient * (1 + percentRemaining) * float64(remaining))
}

func projectedRushBuyTypeCoefficient(projection Projection, item engine.ProductionItem) float64 {
	if projection.Diplomacy == nil || projection.Rules == nil || projection.Diplomacy.GovernmentID == "" {
		return 1
	}
	for _, government := range projection.Rules.Governments {
		if government.ID != projection.Diplomacy.GovernmentID {
			continue
		}
		switch item.Kind {
		case engine.ProductionUnit:
			return positiveProjectedRushBuyCoefficient(government.UnitRushModifier)
		case engine.ProductionBuilding:
			return positiveProjectedRushBuyCoefficient(government.BuildingRushModifier)
		case engine.ProductionWonder:
			return positiveProjectedRushBuyCoefficient(government.WonderRushModifier)
		default:
			return positiveProjectedRushBuyCoefficient(government.EndGameRushModifier)
		}
	}
	return 1
}

func positiveProjectedRushBuyCoefficient(value float64) float64 {
	if value <= 0 {
		return 1
	}
	return value
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
		if !unit.InSpace && unit.TransportID == "" {
			if unit.Asleep {
				actions = append(actions, action(engine.CommandWake, "wake", unit.ID, "", map[string]string{"unit_id": unit.ID}, []string{"unit_id"}))
			} else {
				actions = append(actions, action(engine.CommandSleep, "sleep", unit.ID, "", map[string]string{"unit_id": unit.ID}, []string{"unit_id"}))
			}
		}
		if unit.TransportID != "" {
			actions = append(actions, disembarkActions(projection, tiles, units, unit)...)
			continue
		}
		if unit.ArmyID == "" && unit.TileID != "" {
			actions = append(actions, action(engine.CommandArmyCreate, "army-create", unit.ID, "", map[string]any{"army_id": "client-army-" + unit.ID, "unit_ids": []string{unit.ID}}, []string{"army_id", "unit_ids"}))
		}
		if unit.TileID != "" && !unit.Fortified && projectedUnitCanFortify(unit.Effects) {
			actions = append(actions, action(engine.CommandFortify, "fortify", unit.ID, unit.TileID, map[string]string{"unit_id": unit.ID}, []string{"unit_id"}))
		}
		if unit.ArmyID == "" && unit.TileID != "" && hasEffect(unit.Effects, engine.EffectSettler) && !cityOnTile(projection, unit.TileID) {
			actions = append(actions, action(engine.CommandFoundCity, "found-city", unit.ID, unit.TileID, map[string]string{"unit_id": unit.ID, "city_id": "client-city-" + unit.ID, "name": "City " + unit.ID}, []string{"unit_id", "city_id", "name"}))
		}
		actions = append(actions, upgradeActions(projection, unit)...)
		actions = append(actions, moveActions(projection, tiles, unit)...)
		actions = append(actions, embarkActions(projection, unit)...)
		actions = append(actions, combatActions(projection, tiles, unit)...)
		actions = append(actions, bombardActions(projection, tiles, unit)...)
		actions = append(actions, nuclearAttackActions(cities, unit)...)
		actions = append(actions, spaceLaunchActions(cities, unit)...)
		actions = append(actions, giftUnitActions(projection, unit, playerID)...)
		actions = append(actions, pirateTradeActions(projection, cities, unit)...)
		actions = append(actions, expelActions(projection, tiles, unit)...)
		actions = append(actions, sueUnitActions(projection, tiles, unit)...)
		actions = append(actions, specialActions(projection, tiles, cities, unit, playerID)...)
		actions = append(actions, terrainActions(projection, tiles, unit)...)
		actions = append(actions, createParkActions(projection, tiles, unit)...)
	}
	return actions
}

func publicWorksTerrainActions(projection Projection, playerID string) []ActionProjection {
	if projection.Diplomacy == nil || projection.Rules == nil || projection.Diplomacy.PublicWorks < 1 || len(projection.Rules.TerrainImprovements) == 0 {
		return nil
	}
	tiles := projectedTiles(projection)
	actions := make([]ActionProjection, 0)
	seen := make(map[string]bool)
	for _, city := range projection.Cities {
		if city.OwnerID != playerID || city.TileID == "" || city.Stale {
			continue
		}
		cityTile, ok := tiles[city.TileID]
		if !ok || cityTile.Stale {
			continue
		}
		for _, tile := range projection.Tiles {
			if tile.Stale || tile.ID == "" || (tile.ID != city.TileID && !adjacentProjection(cityTile, tile)) {
				continue
			}
			for _, improvement := range orderedTerrainImprovements(*projection.Rules) {
				if terrainImprovementBlockedByAdvance(projection, improvement) || terrainImprovementBlockedByTerrain(improvement, tile.Terrain) || terrainImprovementBlockedByOverlay(*projection.Rules, improvement, tile) {
					continue
				}
				effective := projectedEffectiveTerrainImprovement(*projection.Rules, improvement, tile.Terrain)
				if effective.PublicWorksCost < 1 || effective.BuildTurns < 1 || projection.Diplomacy.PublicWorks < effective.PublicWorksCost {
					continue
				}
				key := tile.ID + "\x00" + improvement.ID
				if seen[key] {
					break
				}
				seen[key] = true
				actions = append(actions, action(engine.CommandBuildImprovement, "build-improvement", city.ID, tile.ID, map[string]string{"tile_id": tile.ID, "improvement_id": improvement.ID}, []string{"tile_id", "improvement_id"}))
				break
			}
		}
	}
	return actions
}

func upgradeActions(projection Projection, unit UnitProjection) []ActionProjection {
	if projection.Rules == nil || unit.TypeID == "" || unit.TransportID != "" {
		return nil
	}
	known := []string(nil)
	if projection.Research != nil {
		known = projection.Research.Advances
	}
	definitions := make(map[string]rules.RuntimeUnitDefinition, len(projection.Rules.Units))
	for _, definition := range projection.Rules.Units {
		definitions[definition.ID] = definition
	}
	current, ok := definitions[unit.TypeID]
	if !ok || len(current.UpgradeTo) == 0 {
		return nil
	}
	actions := make([]ActionProjection, 0, len(current.UpgradeTo))
	for _, targetID := range current.UpgradeTo {
		target, ok := definitions[targetID]
		if !ok || unitProductionBlockedByGovernment(projection, target) || !allAdvancesSatisfied(known, target.Prerequisites) || anyAdvanceSatisfied(known, target.ObsoleteBy) {
			continue
		}
		actions = append(actions, action(engine.CommandUpgradeUnit, "upgrade-unit", unit.ID, targetID, map[string]string{"unit_id": unit.ID, "upgrade_to": targetID}, []string{"unit_id", "upgrade_to"}))
	}
	return actions
}

func publicTargetEffects(effects []engine.Effect) []engine.Effect {
	projected := make([]engine.Effect, 0, 2)
	for _, effect := range effects {
		switch effect.Kind {
		case engine.EffectCanBeSued, engine.EffectCanBeExpelled:
			projected = append(projected, engine.Effect{ID: effect.ID, Kind: effect.Kind})
		}
	}
	return projected
}

func nuclearAttackActions(cities map[string]CityProjection, unit UnitProjection) []ActionProjection {
	if unit.InSpace || !effectHasKind(unit.Effects, engine.EffectNuclearAttack) {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, city := range cities {
		if city.Stale || city.OwnerID == unit.OwnerID {
			continue
		}
		actions = append(actions, action(engine.CommandNuclearAttack, "nuclear-attack", unit.ID, city.TileID, map[string]string{"unit_id": unit.ID, "target_tile_id": city.TileID}, []string{"unit_id", "target_tile_id"}))
	}
	return actions
}

func spaceLaunchActions(cities map[string]CityProjection, unit UnitProjection) []ActionProjection {
	if unit.InSpace || !effectHasKind(unit.Effects, engine.EffectSpaceLaunchTurns) {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, city := range cities {
		if city.Stale || city.OwnerID != unit.OwnerID {
			continue
		}
		actions = append(actions, action(engine.CommandSpaceLaunch, "space-launch", unit.ID, city.TileID, map[string]string{"unit_id": unit.ID, "target_tile_id": city.TileID}, []string{"unit_id", "target_tile_id"}))
	}
	return actions
}

func giftUnitActions(projection Projection, unit UnitProjection, playerID string) []ActionProjection {
	if unit.InSpace || unit.TileID == "" || unit.TransportID != "" || !effectHasKind(unit.Effects, engine.EffectCanBeGifted) || projection.Diplomacy == nil {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, relation := range projection.Diplomacy.Relations {
		if relation.TargetPlayerID == "" || relation.TargetPlayerID == playerID {
			continue
		}
		actions = append(actions, action(engine.CommandGiftUnit, "gift-unit", unit.ID, relation.TargetPlayerID, map[string]string{"unit_id": unit.ID, "target_player_id": relation.TargetPlayerID}, []string{"unit_id", "target_player_id"}))
	}
	return actions
}

func pirateTradeActions(projection Projection, cities map[string]CityProjection, unit UnitProjection) []ActionProjection {
	if unit.InSpace || unit.TileID == "" || unit.TransportID != "" || unit.ArmyID == "" || unit.Movement < 1 || !effectHasKind(unit.Effects, engine.EffectCanPirate) {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, route := range projection.TradeRoutes {
		if route.ID == "" || route.PiratingArmyID == unit.ArmyID {
			continue
		}
		source, sourceOK := cities[route.SourceCityID]
		destination, destinationOK := cities[route.DestinationCityID]
		if !sourceOK || !destinationOK || source.Stale || destination.Stale || source.OwnerID == unit.OwnerID {
			continue
		}
		if source.TileID != unit.TileID && destination.TileID != unit.TileID {
			continue
		}
		actions = append(actions, action(engine.CommandPirateTrade, "pirate-trade", unit.ID, route.ID, map[string]string{"unit_id": unit.ID, "route_id": route.ID}, []string{"unit_id", "route_id"}))
	}
	return actions
}

func expelActions(projection Projection, tiles map[string]TileProjection, unit UnitProjection) []ActionProjection {
	if unit.TileID == "" || unit.Movement < 1 || !hasEffect(unit.Effects, engine.EffectSpecialExpel) {
		return nil
	}
	origin, ok := tiles[unit.TileID]
	if !ok {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, target := range projection.Units {
		if target.OwnerID == unit.OwnerID || target.TileID == "" || target.TransportID != "" || !hasEffect(target.Effects, engine.EffectCanBeExpelled) {
			continue
		}
		targetTile, ok := tiles[target.TileID]
		if !ok || targetTile.Stale || !adjacentProjection(origin, targetTile) {
			continue
		}
		actions = append(actions, action(engine.CommandExpel, "expel", unit.ID, target.ID, map[string]string{"unit_id": unit.ID, "target_unit_id": target.ID}, []string{"unit_id", "target_unit_id"}))
	}
	return actions
}

func sueUnitActions(projection Projection, tiles map[string]TileProjection, unit UnitProjection) []ActionProjection {
	if unit.TileID == "" || unit.Movement < 1 || !hasEffect(unit.Effects, engine.EffectSpecialSue) {
		return nil
	}
	origin, ok := tiles[unit.TileID]
	if !ok {
		return nil
	}
	actions := make([]ActionProjection, 0)
	seenTiles := map[string]bool{}
	for _, target := range projection.Units {
		if target.OwnerID == unit.OwnerID || target.TileID == "" || target.TransportID != "" || !hasEffect(target.Effects, engine.EffectCanBeSued) || seenTiles[target.TileID] {
			continue
		}
		targetTile, ok := tiles[target.TileID]
		if !ok || targetTile.Stale || !(target.TileID == unit.TileID || adjacentProjection(origin, targetTile)) {
			continue
		}
		seenTiles[target.TileID] = true
		actions = append(actions, action(engine.CommandSue, "sue", unit.ID, target.ID, map[string]string{"unit_id": unit.ID, "target_unit_id": target.ID}, []string{"unit_id", "target_unit_id"}))
	}
	return actions
}

func createParkActions(projection Projection, tiles map[string]TileProjection, unit UnitProjection) []ActionProjection {
	if unit.TileID == "" || unit.Movement < 1 || !hasEffect(unit.Effects, engine.EffectSpecialCreatePark) {
		return nil
	}
	origin, ok := tiles[unit.TileID]
	if !ok {
		return nil
	}
	actions := make([]ActionProjection, 0)
	for _, tile := range projection.Tiles {
		if tile.Stale || tile.Pollution <= 0 || projectedTileDistance(origin, tile) > 1 {
			continue
		}
		actions = append(actions, action(engine.CommandCreatePark, "create-park", unit.ID, tile.ID, map[string]string{"unit_id": unit.ID, "tile_id": tile.ID}, []string{"unit_id", "tile_id"}))
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
		if transport.OwnerID != unit.OwnerID || transport.ID == unit.ID || transport.TileID != unit.TileID || transport.TransportID != "" || transport.CargoCapacity <= cargoCount(projection, transport.ID) || !projectedTransportCanCarryUnit(transport, unit) {
			continue
		}
		actions = append(actions, action(engine.CommandEmbark, "embark", unit.ID, transport.ID, map[string]string{"unit_id": unit.ID, "transport_id": transport.ID}, []string{"unit_id", "transport_id"}))
	}
	return actions
}

func projectedTransportCanCarryUnit(transport UnitProjection, unit UnitProjection) bool {
	if len(transport.CargoClasses) > 0 {
		for _, cargoClass := range projectedUnitCargoClasses(unit) {
			if containsString(transport.CargoClasses, cargoClass) {
				return true
			}
		}
		return false
	}
	if containsString(transport.CargoDomains, unit.Domain) {
		return true
	}
	if unit.Domain == engine.DomainAmphibious {
		return containsString(transport.CargoDomains, engine.DomainLand) || containsString(transport.CargoDomains, engine.DomainSea)
	}
	return false
}

func projectedUnitCargoClasses(unit UnitProjection) []string {
	size := unit.Size
	if size == "" {
		size = "Small"
	}
	size = strings.TrimSpace(size)
	if strings.EqualFold(size, "Medium") {
		size = "Med"
	}
	switch strings.ToLower(unit.Domain) {
	case engine.DomainLand:
		return []string{size + "Land"}
	case engine.DomainAir:
		return []string{size + "Air"}
	case engine.DomainSea:
		return []string{size + "Water"}
	case engine.DomainAmphibious:
		return []string{size + "Land", size + "Water"}
	case engine.DomainSpace:
		return []string{size + "Space"}
	default:
		return []string{size + unit.Domain}
	}
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

func cityHasFranchise(city CityProjection, ownerID string) bool {
	for _, effect := range city.Effects {
		if effect.Kind == engine.EffectFranchise && effect.ID == "franchise:"+ownerID {
			return true
		}
	}
	return false
}

func cityHasExternalFranchise(city CityProjection, ownerID string) bool {
	for _, effect := range city.Effects {
		if effect.Kind == engine.EffectFranchise && effect.ID != "franchise:"+ownerID {
			return true
		}
	}
	return false
}

func bombardActions(projection Projection, tiles map[string]TileProjection, attacker UnitProjection) []ActionProjection {
	if attacker.TileID == "" || attacker.Movement < 1 || attacker.Ranged < 1 || attacker.BombardRange < 1 || attacker.BombRounds < 1 {
		return nil
	}
	origin, ok := tiles[attacker.TileID]
	if !ok {
		return nil
	}
	seenTiles := map[string]bool{}
	actions := make([]ActionProjection, 0)
	for _, defender := range projection.Units {
		if defender.OwnerID == attacker.OwnerID || defender.TileID == "" || defender.TransportID != "" || seenTiles[defender.TileID] {
			continue
		}
		target, ok := tiles[defender.TileID]
		if !ok || target.Stale || projectedTileDistance(origin, target) > attacker.BombardRange || !projectionDomainsCompatible(attacker, defender) {
			continue
		}
		seenTiles[defender.TileID] = true
		actions = append(actions, action(engine.CommandBombard, "bombard", attacker.ID, defender.TileID, map[string]string{"unit_id": attacker.ID, "target_tile_id": defender.TileID}, []string{"unit_id", "target_tile_id"}))
	}
	for _, city := range projection.Cities {
		if city.OwnerID == attacker.OwnerID || city.TileID == "" || city.Stale || seenTiles[city.TileID] {
			continue
		}
		target, ok := tiles[city.TileID]
		if !ok || target.Stale || projectedTileDistance(origin, target) > attacker.BombardRange || !projectionDomainsCompatible(attacker, UnitProjection{Domain: engine.DomainLand}) {
			continue
		}
		seenTiles[city.TileID] = true
		actions = append(actions, action(engine.CommandBombard, "bombard", attacker.ID, city.TileID, map[string]string{"unit_id": attacker.ID, "target_tile_id": city.TileID}, []string{"unit_id", "target_tile_id"}))
	}
	return actions
}

func projectedTileDistance(left, right TileProjection) int {
	return abs(left.X-right.X) + abs(left.Y-right.Y)
}

func unitVisibleToPlayer(state engine.State, playerID string, visible map[string]bool, unit engine.Unit) bool {
	if unit.OwnerID == playerID {
		return true
	}
	if !visible[unit.TileID] {
		return false
	}
	if effectHasKind(unit.Effects, engine.EffectInvisibleForGlobeSat) && playerHasGlobalRadar(state, playerID) && !visibleTilesWithoutGlobalRadar(state)[playerID][unit.TileID] {
		return false
	}
	submarine := effectHasKind(unit.Effects, engine.EffectSubmarine)
	stealth := effectHasKind(unit.Effects, engine.EffectStealth)
	if !submarine && !stealth {
		return true
	}
	if submarine && unitDetectedByPlayer(state, playerID, unit.TileID, engine.EffectCanAttackUnderwater) {
		return true
	}
	return stealth && unitDetectedByPlayer(state, playerID, unit.TileID, engine.EffectCanSeeStealth)
}

func unitDetectedByPlayer(state engine.State, playerID, tileID string, detectorEffect string) bool {
	target, ok := tileByID(state.Tiles, tileID)
	if !ok {
		return false
	}
	for _, unit := range state.Units {
		if unit.OwnerID != playerID || unit.TileID == "" || unit.TransportID != "" || !effectHasKind(unit.Effects, detectorEffect) {
			continue
		}
		origin, ok := tileByID(state.Tiles, unit.TileID)
		if !ok {
			continue
		}
		radius := defaultSightRadius
		if unit.VisibilityRange > radius {
			radius = unit.VisibilityRange
		}
		if abs(origin.X-target.X)+abs(origin.Y-target.Y) <= radius {
			return true
		}
	}
	for _, overlay := range state.Overlays {
		if overlay.OwnerID != playerID || overlay.TileID == "" || overlay.Remaining > 0 || overlay.Kind != "improvement" || !overlayCanSee(overlay, detectorEffect) {
			continue
		}
		origin, ok := tileByID(state.Tiles, overlay.TileID)
		if !ok {
			continue
		}
		radius := overlay.VisionRange
		if overlay.RadarRange > radius {
			radius = overlay.RadarRange
		}
		if radius < defaultSightRadius {
			radius = defaultSightRadius
		}
		if abs(origin.X-target.X)+abs(origin.Y-target.Y) <= radius {
			return true
		}
	}
	return false
}

func overlayCanSee(overlay engine.TileOverlay, detectorEffect string) bool {
	switch detectorEffect {
	case engine.EffectCanAttackUnderwater:
		return containsString(overlay.CanSee, "Underwater")
	case engine.EffectCanSeeStealth:
		return containsString(overlay.CanSee, "Stealth")
	default:
		return containsString(overlay.CanSee, "Standard")
	}
}

func effectHasKind(effects []engine.Effect, kind string) bool {
	for _, effect := range effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
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
		if !ok || !adjacentProjection(origin, cityTile) {
			continue
		}
		if city.OwnerID == playerID {
			if hasEffect(unit.Effects, engine.EffectSpecialSue) && cityHasExternalFranchise(city, playerID) {
				actions = append(actions, action(engine.CommandSue, "sue", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
			}
			continue
		}
		if hasEffect(unit.Effects, engine.EffectSpecialEmbassy) && !containsString(projection.EmbassyPlayerIDs, city.OwnerID) {
			actions = append(actions, action(engine.CommandEstablishEmbassy, "establish-embassy", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSoothsaying) {
			actions = append(actions, action(engine.CommandSoothsaying, "soothsaying", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectAdvertise) && hasEffect(unit.Effects, engine.EffectSpecialCauseUnhappiness) {
			actions = append(actions, action(engine.CommandAdvertise, "advertise", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		} else if hasEffect(unit.Effects, engine.EffectSpecialCauseUnhappiness) {
			actions = append(actions, action(engine.CommandCauseUnhappiness, "cause-unhappiness", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialInciteRevolution) {
			actions = append(actions, action(engine.CommandInciteRevolution, "incite-revolution", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialIndulgenceSales) {
			actions = append(actions, action(engine.CommandIndulgenceSales, "indulgence-sales", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSlaveUprising) {
			actions = append(actions, action(engine.CommandSlaveUprising, "slave-uprising", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialConvertCity) {
			actions = append(actions, action(engine.CommandConvertCity, "convert-city", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialCreateFranchise) && !cityHasFranchise(city, unit.OwnerID) {
			actions = append(actions, action(engine.CommandCreateFranchise, "create-franchise", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialInvestigateCity) {
			actions = append(actions, action(engine.CommandInvestigateCity, "investigate-city", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialPlantNuke) {
			actions = append(actions, action(engine.CommandPlantNuke, "plant-nuke", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if partyActionAllowed(projection, city.OwnerID) && hasEffect(unit.Effects, engine.EffectThrowParty) {
			actions = append(actions, action(engine.CommandThrowParty, "throw-party", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialAssassinateRuler) {
			actions = append(actions, action(engine.CommandAssassinateRuler, "assassinate-ruler", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialInjoin) && !hasEffect(city.Effects, engine.EffectInjunction) {
			actions = append(actions, action(engine.CommandInjoin, "injoin", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialReform) {
			actions = append(actions, action(engine.CommandReform, "reform", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialUndergroundRailway) {
			actions = append(actions, action(engine.CommandUndergroundRailway, "underground-railway", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSue) && !hasEffect(city.Effects, engine.EffectLawsuit) {
			actions = append(actions, action(engine.CommandSue, "sue", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialBioTerror) {
			actions = append(actions, action(engine.CommandBioTerror, "bio-terror", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialNanoTerror) {
			actions = append(actions, action(engine.CommandNanoTerror, "nano-terror", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialPlague) {
			actions = append(actions, action(engine.CommandPlague, "plague", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSlaveRaid) {
			actions = append(actions, action(engine.CommandSlaveRaid, "slave-raid", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSettlerSlaveRaid) {
			actions = append(actions, action(engine.CommandSettlerSlaveRaid, "settler-slave-raid", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
		if hasEffect(unit.Effects, engine.EffectSpecialSteal) || hasEffect(unit.Effects, engine.EffectSpecialCapture) {
			actions = append(actions, action(engine.CommandStealTechnology, "steal-technology", unit.ID, city.ID, map[string]string{"unit_id": unit.ID, "city_id": city.ID}, []string{"unit_id", "city_id"}))
		}
	}
	return actions
}

func partyActionAllowed(projection Projection, targetPlayerID string) bool {
	if projection.Diplomacy == nil || projection.Diplomacy.Gold < retailPartyCost {
		return false
	}
	for _, relation := range projection.Diplomacy.Relations {
		if relation.TargetPlayerID == targetPlayerID {
			return relation.Regard != "hotwar"
		}
	}
	return true
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
		if tile.Stale || !adjacentProjection(origin, tile) || tileHasPendingOverlay(tile) {
			continue
		}
		if tile.Terrain == engine.TerrainForest && terraformCost > 0 && terraformTurns > 0 && projection.Diplomacy.PublicWorks >= terraformCost {
			actions = append(actions, action(engine.CommandTerraform, "terraform", unit.ID, tile.ID, map[string]string{"unit_id": unit.ID, "tile_id": tile.ID, "target_terrain": engine.TerrainPlains}, []string{"unit_id", "tile_id", "target_terrain"}))
		}
		if projection.Rules != nil {
			for _, improvement := range orderedTerrainImprovements(*projection.Rules) {
				if terrainImprovementBlockedByAdvance(projection, improvement) || terrainImprovementBlockedByTerrain(improvement, tile.Terrain) || terrainImprovementBlockedByOverlay(*projection.Rules, improvement, tile) {
					continue
				}
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

func terrainImprovementBlockedByTerrain(improvement rules.RuntimeTerrainImprovementDefinition, terrain string) bool {
	return len(improvement.AllowedTerrains) > 0 && !containsString(improvement.AllowedTerrains, terrain)
}

func tileHasPendingOverlay(tile TileProjection) bool {
	for _, overlay := range tile.Overlays {
		if overlay.Remaining > 0 {
			return true
		}
	}
	return false
}

func terrainImprovementBlockedByOverlay(projection RulesProjection, improvement rules.RuntimeTerrainImprovementDefinition, tile TileProjection) bool {
	for _, overlay := range tile.Overlays {
		if overlay.Remaining > 0 {
			return true
		}
		if overlay.Kind != "improvement" {
			continue
		}
		if overlay.ImprovementID == improvement.ID || containsString(improvement.Excludes, overlay.ImprovementID) {
			return true
		}
		existing, ok := projectedTerrainImprovementByID(projection, overlay.ImprovementID)
		if !ok {
			continue
		}
		if existing.Class != "" && containsString(improvement.Excludes, existing.Class) || improvement.Class != "" && containsString(existing.Excludes, improvement.Class) || containsString(existing.Excludes, improvement.ID) {
			return true
		}
	}
	return false
}

func projectedTerrainImprovementByID(projection RulesProjection, id string) (rules.RuntimeTerrainImprovementDefinition, bool) {
	for _, improvement := range projection.TerrainImprovements {
		if improvement.ID == id {
			return improvement, true
		}
	}
	return rules.RuntimeTerrainImprovementDefinition{}, false
}

func terrainImprovementBlockedByAdvance(projection Projection, improvement rules.RuntimeTerrainImprovementDefinition) bool {
	if improvement.RequiredAdvanceID == "" {
		return false
	}
	if projection.Research == nil {
		return true
	}
	return !containsString(projection.Research.Advances, improvement.RequiredAdvanceID)
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

func unitRemovesPopulation(unit rules.RuntimeUnitDefinition) bool {
	return runtimeUnitHasEffect(unit, engine.EffectBuildingRemovesPopulation)
}

func unitProductionBlockedByCityGood(city CityProjection, unit rules.RuntimeUnitDefinition) bool {
	return unit.NeedsCityGoodID != "" && !containsString(city.Goods, unit.NeedsCityGoodID)
}

func unitProductionBlockedByEliminateNukes(state engine.State, playerID string, unit rules.RuntimeUnitDefinition) bool {
	return runtimeUnitHasEffect(unit, engine.EffectNuclearAttack) && engine.NuclearAttacksEliminated(state, playerID)
}

func stateEffectsContain(effects []engine.Effect, kind string) bool {
	for _, effect := range effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func unitProductionBlockedByGovernment(projection Projection, unit rules.RuntimeUnitDefinition) bool {
	requiredGovernmentIDs := unitRequiredGovernmentIDs(unit)
	if len(requiredGovernmentIDs) == 0 {
		return false
	}
	return projection.Diplomacy == nil || !containsString(requiredGovernmentIDs, projection.Diplomacy.GovernmentID)
}

func unitRequiredGovernmentIDs(unit rules.RuntimeUnitDefinition) []string {
	if len(unit.RequiredGovernmentIDs) > 0 {
		return unit.RequiredGovernmentIDs
	}
	if unit.RequiredGovernmentID != "" {
		return []string{unit.RequiredGovernmentID}
	}
	return nil
}

func unitProductionBlockedByExclusiveSlavery(projection Projection, playerID string, unit rules.RuntimeUnitDefinition) bool {
	if (runtimeUnitHasEffect(unit, engine.EffectSpecialSlaveRaid) || runtimeUnitHasEffect(unit, engine.EffectSpecialSettlerSlaveRaid)) && projectionPlayerHasUnitEffect(projection, playerID, engine.EffectNoSlaves) {
		return true
	}
	if runtimeUnitHasEffect(unit, engine.EffectNoSlaves) && projectionPlayerHasSlaverUnit(projection, playerID) {
		return true
	}
	return false
}

func projectionPlayerHasUnitEffect(projection Projection, playerID, effectKind string) bool {
	for _, unit := range projection.Units {
		if unit.OwnerID == playerID && unitProjectionHasEffect(unit, effectKind) {
			return true
		}
	}
	return false
}

func projectionPlayerHasSlaverUnit(projection Projection, playerID string) bool {
	for _, unit := range projection.Units {
		if unit.OwnerID == playerID && (unitProjectionHasEffect(unit, engine.EffectSpecialSlaveRaid) || unitProjectionHasEffect(unit, engine.EffectSpecialSettlerSlaveRaid)) {
			return true
		}
	}
	return false
}

func unitProjectionHasEffect(unit UnitProjection, kind string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func runtimeUnitHasEffect(unit rules.RuntimeUnitDefinition, kind string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func cityProductionPlacementAllowed(projection Projection, city CityProjection, production rules.RuntimeProduction) bool {
	if !production.CantBuildInSea && !production.CantBuildOnLand && !production.CoastalOnly {
		return true
	}
	tiles := projectedTiles(projection)
	cityTile, ok := tiles[city.TileID]
	if !ok {
		return false
	}
	isSeaCity := projectedTerrainIsSea(projection, cityTile.Terrain)
	if production.CantBuildInSea && isSeaCity {
		return false
	}
	if production.CantBuildOnLand && !isSeaCity {
		return false
	}
	if !production.CoastalOnly || isSeaCity {
		return true
	}
	for _, tile := range tiles {
		if adjacentProjection(cityTile, tile) && projectedTerrainIsSea(projection, tile.Terrain) {
			return true
		}
	}
	return false
}

func projectedTerrainIsSea(projection Projection, terrainID string) bool {
	if terrainID == engine.TerrainOcean {
		return true
	}
	if projection.Rules == nil || projection.Rules.Movement == nil {
		return false
	}
	terrain, ok := projectedRuntimeTerrain(projection.Rules.Movement.Terrains, terrainID)
	return ok && containsString(terrain.Domains, engine.DomainSea) && !containsString(terrain.Domains, engine.DomainLand)
}

func allBuildingsSatisfied(built map[string]bool, required []string) bool {
	for _, id := range required {
		if !built[id] {
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

func planningTiles(state engine.State, playerID string, visible map[string]bool, inspector *engine.Inspector) (map[string]*TilePlanning, error) {
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
		cost := inspector.MovementCost(tile.ID)
		if cost == nil {
			return nil, errors.New("visible movement cost is unavailable")
		}
		result[tile.ID] = &TilePlanning{
			Cost:       *cost,
			Domains:    append([]string(nil), terrain.Domains...),
			Accessible: tile.OwnerID == "" || tile.OwnerID == playerID || (runtime.Movement.OpenAccess && tile.Open),
		}
	}
	if !runtime.Movement.ZOCStopsPath {
		return result, nil
	}
	for _, unit := range state.Units {
		if unit.OwnerID == playerID || !unit.ZOC || unit.TransportID != "" || !unitVisibleToPlayer(state, playerID, visible, unit) {
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
		planning := TilePlanning{Cost: 1, Domains: []string{"amphibious", "land"}, Accessible: true}
		if tile.Terrain == "forest" {
			planning.Cost = 2
		}
		if tile.Terrain == "ocean" {
			planning.Domains = []string{"amphibious", "sea"}
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
	return visibleTilesFor(state, true)
}

func visibleTilesWithoutGlobalRadar(state engine.State) map[string]map[string]bool {
	return visibleTilesFor(state, false)
}

func visibleTilesFor(state engine.State, includeGlobalRadar bool) map[string]map[string]bool {
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
	for _, overlay := range state.Overlays {
		if overlay.OwnerID == "" || overlay.TileID == "" || overlay.Remaining > 0 || overlay.Kind != "improvement" {
			continue
		}
		radius := overlay.VisionRange
		if overlay.RadarRange > radius {
			radius = overlay.RadarRange
		}
		if overlay.ListeningPost && radius < defaultSightRadius {
			radius = defaultSightRadius
		}
		if radius > 0 {
			addSight(result[overlay.OwnerID], state, overlay.TileID, radius)
		}
	}
	for _, player := range state.Players {
		if includeGlobalRadar && playerHasGlobalRadar(state, player.ID) {
			for _, tile := range state.Tiles {
				result[player.ID][tile.ID] = true
			}
		}
	}
	return result
}

func playerHasGlobalRadar(state engine.State, playerID string) bool {
	for _, wonder := range state.Wonders {
		if !wonderActiveForPlayer(state, wonder, playerID) {
			continue
		}
		for _, effect := range wonder.Effects {
			if effect.Kind == engine.EffectGlobalRadar {
				return true
			}
		}
	}
	return false
}

func wonderActiveForPlayer(state engine.State, wonder engine.Wonder, playerID string) bool {
	if wonder.Destroyed || (wonder.OwnerID != playerID && wonder.Scope != "global") {
		return false
	}
	if wonder.ObsoletedBy == "" {
		return true
	}
	ownerNationID := ""
	for _, player := range state.Players {
		if player.ID == wonder.OwnerID {
			ownerNationID = player.NationID
			break
		}
	}
	if ownerNationID == "" {
		return true
	}
	for _, nation := range state.Nations {
		if nation.ID == ownerNationID {
			return !containsString(nation.Advances, wonder.ObsoletedBy)
		}
	}
	return true
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
func cityByID(cities []engine.City, id string) (engine.City, bool) {
	for _, city := range cities {
		if city.ID == id {
			return city, true
		}
	}
	return engine.City{}, false
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
