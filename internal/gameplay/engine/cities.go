package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	ProductionUnit     = "unit"
	ProductionBuilding = "building"
	ProductionWonder   = "wonder"

	EffectSettler                         = "settler"
	EffectSettleSize                      = "settle_size"
	EffectSettleBuilding                  = "settle_building"
	EffectYieldFood                       = "yield_food"
	EffectYieldProduction                 = "yield_production"
	EffectYieldGold                       = "yield_gold"
	EffectYieldGoldPerPopulation          = "yield_gold_per_population"
	EffectHappiness                       = "happiness"
	EffectCrimePercent                    = "crime_percent"
	EffectMaxPopulation                   = "max_population"
	EffectOvercrowdingLevel               = "overcrowding_level"
	EffectStarvationProtection            = "starvation_protection"
	EffectAirport                         = "airport"
	EffectMounted                         = "mounted"
	EffectNoFuelThenCrash                 = "no_fuel_then_crash"
	EffectCantCaptureCity                 = "cant_capture_city"
	EffectCanEntrench                     = "can_entrench"
	EffectCanPillage                      = "can_pillage"
	EffectBuildingRemovesPopulation       = "building_removes_population"
	EffectIgnoreZOC                       = "ignore_zoc"
	EffectCombatDefenseVsMountedPercent   = "combat_defense_vs_mounted_percent"
	EffectCombatDefenseVsAirPercent       = "combat_defense_vs_air_percent"
	EffectWoodenShip                      = "wooden_ship"
	EffectSubmarine                       = "submarine"
	EffectCanAttackLand                   = "can_attack_land"
	EffectCanAttackSea                    = "can_attack_sea"
	EffectCanAttackAir                    = "can_attack_air"
	EffectCanAttackSpace                  = "can_attack_space"
	EffectCanAttackUnderwater             = "can_attack_underwater"
	EffectCombatAttackVsWoodenShipPercent = "combat_attack_vs_wooden_ship_percent"
	EffectCityWalls                       = "city_walls"
	EffectIgnoreCityWalls                 = "ignore_city_walls"
	EffectProtectFromNukes                = "protect_from_nukes"
	EffectProtectFromBioAgents            = "protect_from_bio_agents_percent"
	EffectProtectFromNanoVirus            = "protect_from_nano_virus_percent"
	EffectPreventConversion               = "prevent_conversion_percent"
	EffectPreventSlavery                  = "prevent_slavery_percent"
	EffectForceField                      = "force_field"
	EffectFoodVatPercent                  = "food_vat_percent"
	EffectAllowGrunts                     = "allow_grunts"
	EffectBrokerage                       = "brokerage"
	EffectCapitol                         = "capitol"
	EffectCathedral                       = "cathedral"
	EffectDoubleTelevangelists            = "double_televangelists"
	EffectIsReligious                     = "is_religious"
	EffectLowerPeaceMovementPercent       = "lower_peace_movement_percent"
	EffectNuclearPlant                    = "nuclear_plant"
	EffectTelevision                      = "television"
	EffectEliminateNukes                  = "eliminate_nukes"
	EffectCloseEmbassies                  = "close_embassies"
	EffectEmbassiesEverywhere             = "embassies_everywhere"
	EffectEnableParkRangers               = "enable_park_rangers"
	EffectFreeSlaves                      = "free_slaves"
	EffectGlobalRadar                     = "global_radar"
	EffectIncreaseBoatMovement            = "increase_boat_movement"
	EffectIncreaseHP                      = "increase_hp"
	EffectIncreaseBrokerages              = "increase_brokerages"
	EffectIncreaseRegard                  = "increase_regard"
	EffectIncreaseScientists              = "increase_scientists"
	EffectPreventConversionEmpire         = "prevent_conversion_empire"
	EffectProhibitSlavers                 = "prohibit_slavers"
	EffectProtectFromBarbarians           = "protect_from_barbarians"
	EffectReduceReadinessCost             = "reduce_readiness_cost"
	EffectReformCities                    = "reform_cities"
	EffectRevoltingCitiesJoinPlayer       = "revolting_cities_join_player"
	EffectSpiesEverywhere                 = "spies_everywhere"
	EffectStartGaiaController             = "start_gaia_controller"
	EffectTemporaryFullHappiness          = "temporary_full_happiness"
	EffectPollutersToParks                = "polluters_to_parks"
	EffectGoldPerWaterTradeRoute          = "gold_per_water_trade_route"
	EffectGoldPerInternationalTradeRoute  = "gold_per_international_trade_route"
	EffectGoldPerTelevision               = "gold_per_television"
	EffectDecreaseEmpireSize              = "decrease_empire_size"
	EffectPollutionPercent                = "pollution_percent"
	EffectPollutionPopulationPercent      = "pollution_population_percent"
	EffectPollutionProductionPercent      = "pollution_production_percent"
	EffectPollution                       = "pollution"
	EffectYieldFoodPercent                = "yield_food_percent"
	EffectYieldProductionPercent          = "yield_production_percent"
	EffectYieldGoldPercent                = "yield_gold_percent"
	EffectCombat                          = "combat"
	EffectVision                          = "vision"
	EffectProduction                      = "production"
	EffectVictory                         = "victory"
	EffectResearch                        = "research"
	EffectResearchPercent                 = "research_percent"
	EffectMaxSciencePercent               = "max_science_percent"
	EffectBonus                           = "bonus"
	EffectCombatAttack                    = "combat_attack"
	EffectCombatAttackLand                = "combat_attack_land"
	EffectCombatAttackSea                 = "combat_attack_sea"
	EffectCombatAttackAir                 = "combat_attack_air"
	EffectCombatDefense                   = "combat_defense"
)

type Yield struct {
	Food       int `json:"food,omitempty"`
	Production int `json:"production,omitempty"`
	Gold       int `json:"gold,omitempty"`
	Combat     int `json:"combat,omitempty"`
	Vision     int `json:"vision,omitempty"`
	Victory    int `json:"victory,omitempty"`
}

type ProductionItem struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	Cost         int      `json:"cost"`
	Upkeep       int      `json:"upkeep,omitempty"`
	Prerequisite string   `json:"prerequisite,omitempty"`
	ObsoletedBy  string   `json:"obsoleted_by,omitempty"`
	Scope        string   `json:"scope,omitempty"`
	Effects      []Effect `json:"effects,omitempty"`
}

type Building struct {
	ID          string   `json:"id"`
	Upkeep      int      `json:"upkeep,omitempty"`
	ObsoletedBy string   `json:"obsoleted_by,omitempty"`
	Effects     []Effect `json:"effects,omitempty"`
}

type Wonder struct {
	ID          string   `json:"id"`
	OwnerID     string   `json:"owner_id"`
	CityID      string   `json:"city_id"`
	Scope       string   `json:"scope,omitempty"`
	ObsoletedBy string   `json:"obsoleted_by,omitempty"`
	Destroyed   bool     `json:"destroyed,omitempty"`
	Effects     []Effect `json:"effects,omitempty"`
}

type FoundCityPayload struct {
	UnitID string `json:"unit_id"`
	CityID string `json:"city_id"`
	Name   string `json:"name"`
}

type QueueProductionPayload struct {
	CityID string         `json:"city_id"`
	Item   ProductionItem `json:"item"`
}

type CancelProductionPayload struct {
	CityID string `json:"city_id"`
	ItemID string `json:"item_id"`
}

func applyCityCommand(state *State, command Command, runtime *rules.RuntimeRules) (string, error) {
	if err := activeActor(*state, command.ActorPlayerID); err != nil {
		return "", err
	}
	switch command.Type {
	case CommandFoundCity:
		var payload FoundCityPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "city.founded", foundCity(state, command.ActorPlayerID, payload)
	case CommandQueueProduction:
		var payload QueueProductionPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "production.queued", queueProduction(state, command.ActorPlayerID, payload, runtime)
	case CommandCancelProduction:
		var payload CancelProductionPayload
		if err := decodePayload(command.Payload, &payload); err != nil {
			return "", err
		}
		return "production.cancelled", cancelProduction(state, command.ActorPlayerID, payload)
	default:
		return "", fmt.Errorf("unsupported command type %q", command.Type)
	}
}

func foundCity(state *State, actor string, payload FoundCityPayload) error {
	if payload.UnitID == "" || payload.CityID == "" || payload.Name == "" {
		return errors.New("found city unit, city id, and name are required")
	}
	if _, exists := cityByID(*state, payload.CityID); exists {
		return errors.New("city id already exists")
	}
	for _, city := range state.Cities {
		if city.Name == payload.Name {
			return errors.New("city name already exists")
		}
	}
	unit, err := ownedUnit(state, actor, payload.UnitID)
	if err != nil || unit.TransportID != "" || unit.ArmyID != "" || !hasEffect(unit.Effects, EffectSettler) {
		return errors.New("unit cannot found a city")
	}
	tile, ok := tileByID(*state, unit.TileID)
	if !ok || tile.OwnerID != "" && tile.OwnerID != actor {
		return errors.New("city site is not owned by founding player")
	}
	for _, city := range state.Cities {
		if city.TileID == unit.TileID {
			return errors.New("city site is occupied")
		}
	}
	tileID := unit.TileID
	population, buildings := foundedCitySeed(unit.Effects)
	removeUnitAndCargo(state, unit.ID)
	state.Cities = append(state.Cities, City{ID: payload.CityID, Name: payload.Name, OwnerID: actor, TileID: tileID, Population: population, Buildings: buildings})
	return nil
}

func foundedCitySeed(effects []Effect) (int, []Building) {
	population := 1
	buildings := make([]Building, 0)
	seenBuildings := make(map[string]bool)
	for _, effect := range effects {
		switch effect.Kind {
		case EffectSettleSize:
			if effect.Value > 0 {
				population = effect.Value
			}
		case EffectSettleBuilding:
			if effect.ID != "" && !seenBuildings[effect.ID] {
				buildings = append(buildings, Building{ID: effect.ID})
				seenBuildings[effect.ID] = true
			}
		}
	}
	return population, buildings
}

func queueProduction(state *State, actor string, payload QueueProductionPayload, runtime *rules.RuntimeRules) error {
	city, ok := cityByID(*state, payload.CityID)
	if !ok || city.OwnerID != actor {
		return errors.New("city is not owned by actor")
	}
	if err := validateProductionItem(payload.Item); err != nil {
		return err
	}
	payload.Item = applyRuntimeProductionDefaults(payload.Item, runtime)
	if err := validateProductionItem(payload.Item); err != nil {
		return err
	}
	if err := requireProductionUnlock(*state, *city, payload.Item, runtime); err != nil {
		return err
	}
	if productionRemovesPopulation(payload.Item, runtime) && city.Population <= 1 {
		return errors.New("city population is insufficient")
	}
	if payload.Item.Kind == ProductionUnit && HasActiveWonderEffect(*state, actor, EffectProhibitSlavers) && productionHasSlaverEffect(payload.Item, runtime) {
		return errors.New("production prerequisite is not met")
	}
	if payload.Item.Kind == ProductionBuilding && cityHasBuilding(*city, payload.Item.ID) {
		return errors.New("building already exists")
	}
	if payload.Item.Kind == ProductionWonder && wonderExists(*state, payload.Item.ID) {
		return errors.New("wonder already exists")
	}
	for _, item := range city.Queue {
		if item.ID == payload.Item.ID {
			return errors.New("production item already queued")
		}
	}
	city.Queue = append(city.Queue, payload.Item)
	return nil
}

func productionHasSlaverEffect(item ProductionItem, runtime *rules.RuntimeRules) bool {
	if item.Kind != ProductionUnit || runtime == nil {
		return effectsContainSlavery(item.Effects)
	}
	unit, ok := runtime.Unit(item.ID)
	if !ok {
		return effectsContainSlavery(item.Effects)
	}
	for _, effect := range unit.Effects {
		if effect.Kind == EffectSpecialSlaveRaid || effect.Kind == EffectSpecialSettlerSlaveRaid {
			return true
		}
	}
	return false
}

func effectsContainSlavery(effects []Effect) bool {
	for _, effect := range effects {
		if effect.Kind == EffectSpecialSlaveRaid || effect.Kind == EffectSpecialSettlerSlaveRaid {
			return true
		}
	}
	return false
}

func productionRemovesPopulation(item ProductionItem, runtime *rules.RuntimeRules) bool {
	if item.Kind != ProductionUnit || runtime == nil {
		return hasEffect(item.Effects, EffectBuildingRemovesPopulation)
	}
	unit, ok := runtime.Unit(item.ID)
	if !ok {
		return hasEffect(item.Effects, EffectBuildingRemovesPopulation)
	}
	for _, effect := range unit.Effects {
		if effect.Kind == EffectBuildingRemovesPopulation {
			return true
		}
	}
	return false
}

func applyRuntimeProductionDefaults(item ProductionItem, runtime *rules.RuntimeRules) ProductionItem {
	if runtime == nil {
		return item
	}
	if production, ok := runtime.ProductionDefinition(item.ID); ok {
		if production.Cost > 0 {
			item.Cost = production.Cost
		}
		if production.Upkeep > 0 {
			item.Upkeep = production.Upkeep
		}
		if item.Prerequisite == "" && len(production.Prerequisites) > 0 {
			item.Prerequisite = production.Prerequisites[0]
		}
		if item.ObsoletedBy == "" && len(production.ObsoleteBy) > 0 {
			item.ObsoletedBy = production.ObsoleteBy[0]
		}
		if len(production.Effects) > 0 {
			item.Effects = runtimeProductionEffects(production)
		}
	}
	if item.Kind != ProductionUnit {
		return item
	}
	unit, ok := runtime.Unit(item.ID)
	if !ok {
		return item
	}
	if unit.ProductionCost > 0 {
		item.Cost = unit.ProductionCost
	}
	if unit.Upkeep > 0 {
		item.Upkeep = unit.Upkeep
	}
	if item.Prerequisite == "" && len(unit.Prerequisites) > 0 {
		item.Prerequisite = unit.Prerequisites[0]
	}
	if item.ObsoletedBy == "" && len(unit.ObsoleteBy) > 0 {
		item.ObsoletedBy = unit.ObsoleteBy[0]
	}
	return item
}

func cancelProduction(state *State, actor string, payload CancelProductionPayload) error {
	city, ok := cityByID(*state, payload.CityID)
	if !ok || city.OwnerID != actor || payload.ItemID == "" {
		return errors.New("production item is not owned by actor")
	}
	for i, item := range city.Queue {
		if item.ID == payload.ItemID {
			city.Queue = append(city.Queue[:i], city.Queue[i+1:]...)
			city.StoredProduction = 0 // Supported default: cancelled production has no refund.
			return nil
		}
	}
	return errors.New("production item is not queued")
}

// RecalculateRound applies the city pipeline once for the current resolved turn.
// It has no external authority: callers persist the returned canonical state through the aggregate.
func RecalculateRound(state State) (State, error) {
	if err := ValidateState(state); err != nil {
		return State{}, err
	}
	if state.RecalculatedTurn == state.Phase.Turn {
		return state, nil
	}
	runtime, err := runtimeRulesForState(state)
	if err != nil {
		return State{}, err
	}
	next, err := canonicalClone(state)
	if err != nil {
		return State{}, err
	}
	advanceTerraform(&next)
	resetFranchiseProduction(&next)
	cityOrder := cityIDs(next.Cities)
	sort.Strings(cityOrder)
	for _, cityID := range cityOrder {
		city, ok := cityByID(next, cityID)
		if !ok {
			continue
		}
		if err := completeProduction(&next, city, runtime); err != nil {
			return State{}, err
		}
		city, ok = cityByID(next, cityID)
		if !ok {
			continue
		}
		var franchiseOwnerID string
		var franchiseLoss int
		city.Yields, franchiseOwnerID, franchiseLoss, err = cityYieldAndFranchiseLossWithRuntime(next, *city, runtime)
		if err != nil {
			return State{}, err
		}
		if franchiseLoss > 0 {
			addFranchiseProduction(&next, franchiseOwnerID, franchiseLoss)
		}
		city.Happiness = cityHappiness(next, *city)
		advanceTemporaryCityEffects(city)
		city.CrimePercent = cityCrimePercent(next, *city)
		city.OvercrowdingLevel = cityOvercrowdingLevel(next, *city)
		advanceCityFood(city, cityMaxPopulation(next, *city), cityStarvationProtection(next, *city), foodVatPercent(*city))
		if cityPollutionEnabled(runtime) {
			applyCityPollution(&next, *city, runtimePollutionModel(runtime))
		}
	}
	next.RecalculatedTurn = next.Phase.Turn
	return next, nil
}

func resetFranchiseProduction(state *State) {
	for index := range state.Nations {
		state.Nations[index].FranchiseProduction = 0
	}
}

func addFranchiseProduction(state *State, playerID string, amount int) {
	if amount <= 0 {
		return
	}
	nationID := nationForPlayer(*state, playerID)
	if nationID == "" {
		return
	}
	for index := range state.Nations {
		if state.Nations[index].ID == nationID {
			state.Nations[index].FranchiseProduction += amount
			return
		}
	}
}

func cityPollutionEnabled(runtime *rules.RuntimeRules) bool {
	return runtime != nil && len(runtime.PollutionTriggers) > 0
}

func runtimePollutionModel(runtime *rules.RuntimeRules) rules.RuntimePollutionModel {
	if runtime == nil || runtime.PollutionModel == nil {
		return rules.RuntimePollutionModel{}
	}
	return *runtime.PollutionModel
}

func applyCityPollution(state *State, city City, model rules.RuntimePollutionModel) {
	pollution := cityPollution(*state, city, model)
	for index := range state.Tiles {
		if state.Tiles[index].ID == city.TileID {
			state.Tiles[index].Pollution = pollution
			return
		}
	}
}

func cityPollution(state State, city City, model rules.RuntimePollutionModel) int {
	basePopulationPollution := cityPollutionBucket(positivePopulation(city.Population), model.PopulationStart, model.PopulationRatio)
	baseProductionPollution := cityPollutionBucket(city.Yields.Production, model.ProductionStart, model.ProductionRatio)
	pollutionPercent := 0
	populationPercent := 0
	productionPercent := 0
	flatPollution := 0
	for _, effect := range cityEffects(state, city) {
		switch effect.Kind {
		case EffectPollutionPercent:
			pollutionPercent += effect.Value
		case EffectPollutionPopulationPercent:
			populationPercent += effect.Value
		case EffectPollutionProductionPercent:
			productionPercent += effect.Value
		case EffectPollution:
			flatPollution += effect.Value
		}
	}
	basePopulationPollution += signedPercentDelta(basePopulationPollution, pollutionPercent)
	baseProductionPollution += signedPercentDelta(baseProductionPollution, pollutionPercent)
	if basePopulationPollution < 0 {
		basePopulationPollution = 0
	}
	if baseProductionPollution < 0 {
		baseProductionPollution = 0
	}
	populationPollution := basePopulationPollution + signedPercentDelta(basePopulationPollution, populationPercent)
	productionPollution := baseProductionPollution + signedPercentDelta(baseProductionPollution, productionPercent)
	if populationPollution < 0 {
		populationPollution = 0
	}
	if productionPollution < 0 {
		productionPollution = 0
	}
	total := populationPollution + productionPollution + flatPollution
	if total < 0 {
		return 0
	}
	return total
}

func cityPollutionBucket(value, start, ratio int) int {
	if start <= 0 && ratio <= 0 {
		return value
	}
	bucket := value - start
	if bucket < 0 {
		return 0
	}
	if ratio <= 0 {
		ratio = 1
	}
	return bucket * ratio
}

func signedPercentDelta(value, percent int) int {
	if value == 0 || percent == 0 {
		return 0
	}
	return value * percent / 100
}

func advanceCityFood(city *City, maxPopulation int, starvationProtection int, foodVatPercent int) {
	if city.Population < 1 {
		city.Population = 1
	}
	consumption := cityFoodConsumption(city.Population)
	if foodVatPercent > 0 {
		consumption -= consumption * foodVatPercent / 100
		if consumption < 0 {
			consumption = 0
		}
	}
	surplus := city.Yields.Food - consumption
	city.StoredFood += surplus
	if city.StoredFood < 0 {
		if starvationProtection > 0 {
			city.StarvationTurns++
			if city.StarvationTurns <= starvationProtection {
				city.StoredFood = 0
				return
			}
		}
		if city.Population > 1 {
			city.Population--
		}
		city.StoredFood = 0
		return
	}
	city.StarvationTurns = 0
	for city.StoredFood >= cityFoodGrowthThreshold(city.Population) {
		if maxPopulation > 0 && city.Population >= maxPopulation {
			return
		}
		city.StoredFood -= cityFoodGrowthThreshold(city.Population)
		city.Population++
	}
}

func foodVatPercent(city City) int {
	percent := 0
	for _, effect := range city.Effects {
		if effect.Kind == EffectFoodVatPercent && effect.Value > percent {
			percent = effect.Value
		}
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func cityHappiness(state State, city City) int {
	happiness := 0
	for _, effect := range cityEffects(state, city) {
		if effect.Kind == EffectHappiness {
			happiness += effect.Value
		}
	}
	if cityHasTemporaryFullHappiness(city) && happiness < 100 {
		return 100
	}
	return happiness
}

func cityHasTemporaryFullHappiness(city City) bool {
	for _, effect := range city.Effects {
		if effect.Kind == EffectTemporaryFullHappiness && effect.Value > 0 {
			return true
		}
	}
	return false
}

func advanceTemporaryCityEffects(city *City) {
	if len(city.Effects) == 0 {
		return
	}
	effects := city.Effects[:0]
	for _, effect := range city.Effects {
		if effect.Kind == EffectTemporaryFullHappiness {
			effect.Value--
			if effect.Value <= 0 {
				continue
			}
		}
		effects = append(effects, effect)
	}
	city.Effects = effects
}

func cityCrimePercent(state State, city City) int {
	crimePercent := 0
	for _, effect := range cityEffects(state, city) {
		if effect.Kind == EffectCrimePercent {
			crimePercent += effect.Value
		}
	}
	return crimePercent
}

func cityOvercrowdingLevel(state State, city City) int {
	level := 0
	for _, effect := range cityEffects(state, city) {
		if effect.Kind == EffectOvercrowdingLevel && effect.Value > 0 {
			level += effect.Value
		}
	}
	return level
}

func cityStarvationProtection(state State, city City) int {
	protection := 0
	for _, effect := range cityEffects(state, city) {
		if effect.Kind == EffectStarvationProtection && effect.Value > 0 {
			protection += effect.Value
		}
	}
	return protection
}

func cityMaxPopulation(state State, city City) int {
	maxPopulation := 0
	for _, effect := range cityEffects(state, city) {
		if effect.Kind == EffectMaxPopulation && effect.Value > maxPopulation {
			maxPopulation = effect.Value
		}
	}
	return maxPopulation
}

func cityFoodConsumption(population int) int {
	if population < 1 {
		population = 1
	}
	return population * 2
}

func cityFoodGrowthThreshold(population int) int {
	if population < 1 {
		population = 1
	}
	return population * 100
}

func runtimeRulesForState(state State) (*rules.RuntimeRules, error) {
	if !hasRuntimeDefinitions(state.Rules.Definitions) {
		return nil, nil
	}
	runtime, err := rules.CompileRuntime(state.Rules.Definitions, state.Rules.ContentHash)
	if err != nil {
		return nil, err
	}
	return &runtime, nil
}

func completeProduction(state *State, city *City, runtime *rules.RuntimeRules) error {
	if len(city.Queue) == 0 || hasEffect(city.Effects, EffectInjunction) {
		return nil
	}
	yield, err := cityYieldWithRuntime(*state, *city, runtime)
	if err != nil {
		return err
	}
	city.StoredProduction += yield.Production
	item := city.Queue[0]
	if city.StoredProduction < item.Cost {
		return nil
	}
	city.StoredProduction -= item.Cost
	city.Queue = city.Queue[1:]
	switch item.Kind {
	case ProductionUnit:
		return completeUnitProduction(state, city, item, runtime)
	case ProductionBuilding:
		city.Buildings = append(city.Buildings, Building{ID: item.ID, Upkeep: item.Upkeep, ObsoletedBy: item.ObsoletedBy, Effects: item.Effects})
		return nil
	case ProductionWonder:
		if wonderExists(*state, item.ID) {
			city.StoredProduction = 0
			return nil
		}
		state.Wonders = append(state.Wonders, Wonder{ID: item.ID, OwnerID: city.OwnerID, CityID: city.ID, Scope: item.Scope, ObsoletedBy: item.ObsoletedBy, Effects: item.Effects})
		applyWonderCompletionEffects(state, state.Wonders[len(state.Wonders)-1])
		return nil
	default:
		return errors.New("unsupported production kind")
	}
}

func ownerWonderEffectSum(state State, ownerID, kind string) int {
	total := 0
	for _, wonder := range state.Wonders {
		if !wonderAppliesToPlayer(state, wonder, ownerID) {
			continue
		}
		for _, effect := range wonder.Effects {
			if effect.Kind == kind {
				total += effect.Value
			}
		}
	}
	return total
}

// HasActiveWonderEffect reports whether a player is affected by an active wonder marker.
func HasActiveWonderEffect(state State, playerID, kind string) bool {
	for _, wonder := range state.Wonders {
		if !wonderAppliesToPlayer(state, wonder, playerID) {
			continue
		}
		for _, effect := range wonder.Effects {
			if effect.Kind == kind {
				return true
			}
		}
	}
	return false
}

func wonderAppliesToPlayer(state State, wonder Wonder, playerID string) bool {
	return !wonder.Destroyed && (wonder.OwnerID == playerID || wonder.Scope == "global") && !isObsoleteForPlayer(state, wonder.OwnerID, nil, wonder.ObsoletedBy)
}

// HasEffectiveEmbassyWith applies retail embassy wonder modifiers in addition to explicit embassies.
func HasEffectiveEmbassyWith(state State, actorPlayerID, targetPlayerID string) bool {
	if actorPlayerID == "" || targetPlayerID == "" || actorPlayerID == targetPlayerID || HasActiveWonderEffect(state, targetPlayerID, EffectCloseEmbassies) {
		return false
	}
	if HasActiveWonderEffect(state, actorPlayerID, EffectEmbassiesEverywhere) {
		return true
	}
	actorNationID := nationForPlayer(state, actorPlayerID)
	targetNationID := nationForPlayer(state, targetPlayerID)
	actorNation := nationForID(&state, actorNationID)
	return actorNation != nil && targetNationID != "" && containsString(actorNation.Embassies, targetNationID)
}

func applyWonderCompletionEffects(state *State, wonder Wonder) {
	if wonder.Destroyed {
		return
	}
	for _, effect := range wonder.Effects {
		switch effect.Kind {
		case EffectIncreaseHP:
			if effect.Value <= 0 {
				continue
			}
			for index := range state.Units {
				if state.Units[index].OwnerID != wonder.OwnerID {
					continue
				}
				state.Units[index].Health += effect.Value
				state.Units[index].MaxHealth += effect.Value
			}
		case EffectCloseEmbassies:
			closeForeignEmbassiesWithOwner(state, wonder.OwnerID)
			removeWarsWithPlayer(state, wonder.OwnerID)
		case EffectEliminateNukes:
			eliminateNukes(state)
		case EffectFreeSlaves:
			removeSlaverUnits(state)
		case EffectPollutersToParks:
			convertPollutedCitiesToParks(state, wonder.OwnerID, wonder.CityID, effect.Value)
		case EffectTemporaryFullHappiness:
			applyTemporaryFullHappiness(state, wonder.OwnerID, wonder.ID, effect.Value)
		case EffectReformCities:
			reformOwnerCities(state, wonder.OwnerID)
		case EffectEmbassiesEverywhere:
			removeWarsWithPlayer(state, wonder.OwnerID)
		}
	}
}

func eliminateNukes(state *State) {
	for index := range state.Nations {
		upsertNationEffect(&state.Nations[index], EffectEliminateNukes, 1)
	}
	removed := make([]string, 0)
	for _, unit := range state.Units {
		if hasEffect(unit.Effects, EffectSpecialPlantNuke) {
			removed = append(removed, unit.ID)
		}
	}
	for _, id := range removed {
		removeUnitAndCargo(state, id)
	}
}

func reformOwnerCities(state *State, ownerID string) {
	for index := range state.Cities {
		if state.Cities[index].OwnerID != ownerID {
			continue
		}
		state.Cities[index].CrimePercent = 0
		state.Cities[index].Effects = removeEffectsByKind(state.Cities[index].Effects, EffectInjunction)
	}
}

func removeEffectsByKind(effects []Effect, kind string) []Effect {
	filtered := effects[:0]
	for _, effect := range effects {
		if effect.Kind != kind {
			filtered = append(filtered, effect)
		}
	}
	return filtered
}

func applyTemporaryFullHappiness(state *State, ownerID, sourceID string, turns int) {
	if turns <= 0 {
		return
	}
	for index := range state.Cities {
		if state.Cities[index].OwnerID != ownerID {
			continue
		}
		upsertCityEffect(&state.Cities[index], Effect{ID: "temporary-full-happiness:" + sourceID, Kind: EffectTemporaryFullHappiness, Value: turns})
	}
}

func upsertCityEffect(city *City, effect Effect) {
	for index := range city.Effects {
		if city.Effects[index].Kind == effect.Kind && city.Effects[index].ID == effect.ID {
			if effect.Value > city.Effects[index].Value {
				city.Effects[index].Value = effect.Value
			}
			return
		}
	}
	city.Effects = append(city.Effects, effect)
}

func convertPollutedCitiesToParks(state *State, ownerID, excludedCityID string, limit int) {
	if limit <= 0 {
		return
	}
	tilePollution := make(map[string]int, len(state.Tiles))
	for _, tile := range state.Tiles {
		tilePollution[tile.ID] = tile.Pollution
	}
	type candidate struct {
		cityID    string
		tileID    string
		pollution int
	}
	candidates := make([]candidate, 0, len(state.Cities))
	for _, city := range state.Cities {
		if city.ID == excludedCityID || tilePollution[city.TileID] <= 0 {
			continue
		}
		candidates = append(candidates, candidate{cityID: city.ID, tileID: city.TileID, pollution: tilePollution[city.TileID]})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].pollution != candidates[j].pollution {
			return candidates[i].pollution > candidates[j].pollution
		}
		return candidates[i].cityID < candidates[j].cityID
	})
	if limit > len(candidates) {
		limit = len(candidates)
	}
	removedCities := make(map[string]bool, limit)
	parkedTiles := make(map[string]bool, limit)
	for _, candidate := range candidates[:limit] {
		removedCities[candidate.cityID] = true
		parkedTiles[candidate.tileID] = true
		for index := range state.Tiles {
			if state.Tiles[index].ID == candidate.tileID {
				state.Tiles[index].Pollution = 0
				break
			}
		}
	}
	if len(removedCities) == 0 {
		return
	}
	for _, unit := range append([]Unit(nil), state.Units...) {
		if parkedTiles[unit.TileID] {
			removeUnitAndCargo(state, unit.ID)
		}
	}
	for index := range state.Wonders {
		if removedCities[state.Wonders[index].CityID] {
			state.Wonders[index].Destroyed = true
		}
	}
	state.Overlays = filterOverlaysNotOnTiles(state.Overlays, parkedTiles)
	for _, candidate := range candidates[:limit] {
		state.Overlays = append(state.Overlays, TileOverlay{ID: "park:" + ownerID + ":" + candidate.tileID, TileID: candidate.tileID, Kind: "park", OwnerID: ownerID})
	}
	cities := state.Cities[:0]
	for _, city := range state.Cities {
		if !removedCities[city.ID] {
			cities = append(cities, city)
		}
	}
	state.Cities = cities
}

func filterOverlaysNotOnTiles(overlays []TileOverlay, removed map[string]bool) []TileOverlay {
	filtered := overlays[:0]
	for _, overlay := range overlays {
		if !removed[overlay.TileID] {
			filtered = append(filtered, overlay)
		}
	}
	return filtered
}

func removeSlaverUnits(state *State) {
	removed := make([]string, 0)
	for _, unit := range state.Units {
		if unitHasSlaveryEffect(unit) {
			removed = append(removed, unit.ID)
		}
	}
	for _, id := range removed {
		removeUnitAndCargo(state, id)
	}
}

func unitHasSlaveryEffect(unit Unit) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == EffectSpecialSlaveRaid || effect.Kind == EffectSpecialSettlerSlaveRaid {
			return true
		}
	}
	return false
}

func closeForeignEmbassiesWithOwner(state *State, ownerID string) {
	ownerNationID := nationForPlayer(*state, ownerID)
	if ownerNationID == "" {
		return
	}
	for index := range state.Nations {
		if state.Nations[index].ID == ownerNationID {
			continue
		}
		state.Nations[index].Embassies = removeString(state.Nations[index].Embassies, ownerNationID)
	}
}

func removeWarsWithPlayer(state *State, playerID string) {
	wars := state.Policy.Wars[:0]
	for _, war := range state.Policy.Wars {
		if war.A == playerID || war.B == playerID {
			continue
		}
		wars = append(wars, war)
	}
	state.Policy.Wars = wars
}

func completeUnitProduction(state *State, city *City, item ProductionItem, runtime *rules.RuntimeRules) error {
	if runtime == nil {
		return nil
	}
	definition, ok := runtime.Unit(item.ID)
	if !ok {
		return nil
	}
	health := definition.Health + ownerWonderEffectSum(*state, city.OwnerID, EffectIncreaseHP)
	if health < 1 {
		health = 1
	}
	unit := Unit{
		ID:                 producedUnitID(*state, city.ID, item.ID),
		OwnerID:            city.OwnerID,
		TileID:             city.TileID,
		Domain:             definition.Domain,
		Movement:           definition.Movement,
		MaxMovement:        definition.Movement,
		Fuel:               definition.MaxFuel,
		MaxFuel:            definition.MaxFuel,
		CargoCapacity:      definition.TransportCapacity,
		CargoDomains:       append([]string(nil), definition.CargoDomains...),
		Upkeep:             definition.Upkeep,
		ZOC:                definition.ZOC,
		Effects:            runtimeUnitEffects(definition),
		Attack:             definition.Attack,
		Defense:            definition.Defense,
		Ranged:             definition.Ranged,
		BombardRange:       definition.BombardRange,
		BombRounds:         definition.BombRounds,
		ActiveDefenseRange: definition.ActiveDefenseRange,
		Strength:           definition.Strength,
		Armor:              definition.Armor,
		Health:             health,
		MaxHealth:          health,
		VisibilityRange:    definition.VisibilityRange,
	}
	state.Units = append(state.Units, unit)
	if hasEffect(unit.Effects, EffectBuildingRemovesPopulation) && city.Population > 1 {
		city.Population--
	}
	return nil
}

func producedUnitID(state State, cityID, itemID string) string {
	base := fmt.Sprintf("unit:%s:%s:%d", cityID, itemID, state.Phase.Turn)
	if !unitIDExists(state, base) {
		return base
	}
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s:%d", base, suffix)
		if !unitIDExists(state, candidate) {
			return candidate
		}
	}
}

func unitIDExists(state State, id string) bool {
	for _, unit := range state.Units {
		if unit.ID == id {
			return true
		}
	}
	return false
}

func runtimeProductionEffects(definition rules.RuntimeProduction) []Effect {
	effects := make([]Effect, 0, len(definition.Effects))
	for _, effect := range definition.Effects {
		effects = append(effects, Effect{ID: effect.ID, Kind: effect.Kind, Value: effect.Value})
	}
	return effects
}

func runtimeUnitEffects(definition rules.RuntimeUnitDefinition) []Effect {
	effects := make([]Effect, 0, len(definition.Effects))
	for _, effect := range definition.Effects {
		effects = append(effects, Effect{ID: effect.ID, Kind: effect.Kind, Value: effect.Value})
	}
	return effects
}

func cityYield(state State, city City) (Yield, error) {
	return cityYieldWithRuntime(state, city, nil)
}

func cityYieldWithRuntime(state State, city City, runtime *rules.RuntimeRules) (Yield, error) {
	yield, _, _, err := cityYieldAndFranchiseLossWithRuntime(state, city, runtime)
	return yield, err
}

func cityYieldAndFranchiseLossWithRuntime(state State, city City, runtime *rules.RuntimeRules) (Yield, string, int, error) {
	yield, err := cityTileYieldWithRuntime(state, city.TileID, runtime)
	if err != nil {
		return Yield{}, "", 0, err
	}
	seen := map[string]bool{city.TileID: true}
	workedLimit := len(city.WorkedTileIDs)
	if city.Population > 0 && city.Population < workedLimit {
		workedLimit = city.Population
	}
	for index, tileID := range city.WorkedTileIDs {
		if seen[tileID] {
			return Yield{}, "", 0, errors.New("worked tile is duplicated")
		}
		seen[tileID] = true
		if index >= workedLimit {
			continue
		}
		worked, err := terrainYieldWithRuntime(state, tileID, runtime)
		if err != nil {
			return Yield{}, "", 0, err
		}
		yield = addYield(yield, worked)
	}
	effects := cityEffects(state, city)
	hasTelevision := false
	goldPerTelevision := 0
	for _, effect := range effects {
		switch effect.Kind {
		case EffectYieldFood:
			yield.Food += effect.Value
		case EffectYieldProduction, EffectProduction:
			yield.Production += effect.Value
		case EffectYieldGold:
			yield.Gold += effect.Value
		case EffectYieldGoldPerPopulation:
			yield.Gold += effect.Value * positivePopulation(city.Population)
		case EffectYieldFoodPercent:
			yield.Food += signedYieldPercentDelta(yield.Food, effect.Value)
		case EffectYieldProductionPercent:
			yield.Production += signedYieldPercentDelta(yield.Production, effect.Value)
		case EffectYieldGoldPercent:
			yield.Gold += signedYieldPercentDelta(yield.Gold, effect.Value)
		case EffectTelevision:
			hasTelevision = true
		case EffectGoldPerTelevision:
			goldPerTelevision += effect.Value
		case EffectCombat:
			yield.Combat += effect.Value
		case EffectVision:
			yield.Vision += effect.Value
		case EffectVictory:
			yield.Victory += effect.Value
		}
	}
	if brokeragePercent := brokerageGoldPercentBonus(state, city); brokeragePercent != 0 {
		yield.Gold += signedYieldPercentDelta(yield.Gold, brokeragePercent)
	}
	if hasTelevision && goldPerTelevision > 0 {
		yield.Gold += goldPerTelevision
	}
	yield.Gold += tradeRouteWonderGold(state, city)
	for _, building := range city.Buildings {
		if !isObsolete(city.Obsolete, building.ObsoletedBy) {
			yield.Gold -= building.Upkeep
		}
	}
	crimePercent := cityCrimeLossPercent(city)
	if crimePercent > 0 {
		yield.Production -= yield.Production * crimePercent / 100
		yield.Gold -= yield.Gold * crimePercent / 100
	}
	franchiseOwnerID := cityFranchiseOwnerID(city)
	franchiseLoss := 0
	if franchiseOwnerID != "" && franchiseOwnerID != city.OwnerID && yield.Production > 0 {
		franchiseLoss = (yield.Production*10 + 99) / 100
		yield.Production -= franchiseLoss
	}
	if yield.Production < 0 {
		yield.Production = 0
	}
	if yield.Gold < 0 {
		yield.Gold = 0
	}
	return yield, franchiseOwnerID, franchiseLoss, nil
}

func cityFranchiseOwnerID(city City) string {
	const prefix = "franchise:"
	for _, effect := range city.Effects {
		if effect.Kind != EffectFranchise || len(effect.ID) <= len(prefix) || effect.ID[:len(prefix)] != prefix {
			continue
		}
		return effect.ID[len(prefix):]
	}
	return ""
}

func brokerageGoldPercentBonus(state State, city City) int {
	increasePercent := ownerWonderEffectSum(state, city.OwnerID, EffectIncreaseBrokerages)
	if increasePercent == 0 {
		return 0
	}
	bonusPercent := 0
	for _, building := range city.Buildings {
		if isObsoleteForPlayer(state, city.OwnerID, city.Obsolete, building.ObsoletedBy) {
			continue
		}
		brokerage := false
		commercePercent := 0
		for _, effect := range building.Effects {
			switch effect.Kind {
			case EffectBrokerage:
				brokerage = true
			case EffectYieldGoldPercent:
				commercePercent += effect.Value
			}
		}
		if brokerage && commercePercent != 0 {
			bonusPercent += commercePercent * increasePercent / 100
		}
	}
	return bonusPercent
}

func tradeRouteWonderGold(state State, city City) int {
	waterGold := ownerWonderEffectSum(state, city.OwnerID, EffectGoldPerWaterTradeRoute)
	internationalGold := ownerWonderEffectSum(state, city.OwnerID, EffectGoldPerInternationalTradeRoute)
	if waterGold == 0 && internationalGold == 0 {
		return 0
	}
	gold := 0
	for _, route := range state.TradeRoutes {
		if route.SourceCityID != city.ID {
			continue
		}
		if route.CrossesWater {
			gold += waterGold
		}
		destination, ok := cityByID(state, route.DestinationCityID)
		if ok && destination.OwnerID != city.OwnerID {
			gold += internationalGold
		}
	}
	return gold
}

func cityCrimeLossPercent(city City) int {
	if city.CrimePercent <= 0 {
		return 0
	}
	if city.CrimePercent > 100 {
		return 100
	}
	return city.CrimePercent
}

func positivePopulation(population int) int {
	if population < 1 {
		return 1
	}
	return population
}

func signedYieldPercentDelta(value, percent int) int {
	if value <= 0 || percent == 0 {
		return 0
	}
	if percent > 0 {
		return (value*percent + 99) / 100
	}
	adjustedPercent := 100 + percent
	if adjustedPercent <= 0 {
		return -value
	}
	adjusted := (value*adjustedPercent + 99) / 100
	return adjusted - value
}

func cityEffects(state State, city City) []Effect {
	effects := cityLocalEffects(state, city)
	if nationID := nationForPlayer(state, city.OwnerID); nationID != "" {
		for _, nation := range state.Nations {
			if nation.ID == nationID {
				effects = append(effects, nation.Effects...)
			}
		}
	}
	normalizeEffects(effects)
	return effects
}

func cityLocalEffects(state State, city City) []Effect {
	effects := append([]Effect(nil), city.Effects...)
	for _, building := range city.Buildings {
		if !isObsoleteForPlayer(state, city.OwnerID, city.Obsolete, building.ObsoletedBy) {
			effects = append(effects, building.Effects...)
		}
	}
	for _, wonder := range state.Wonders {
		if !wonder.Destroyed && !isObsoleteForPlayer(state, wonder.OwnerID, city.Obsolete, wonder.ObsoletedBy) && (wonder.Scope == "global" || wonder.OwnerID == city.OwnerID) {
			effects = append(effects, wonder.Effects...)
		}
	}
	normalizeEffects(effects)
	return effects
}

func terrainYield(state State, tileID string) (Yield, error) {
	return terrainYieldWithRuntime(state, tileID, nil)
}

func cityTileYieldWithRuntime(state State, tileID string, runtime *rules.RuntimeRules) (Yield, error) {
	if runtime == nil {
		return terrainYieldWithRuntime(state, tileID, nil)
	}
	tile, ok := tileByID(state, tileID)
	if !ok {
		return Yield{}, errors.New("worked tile does not exist")
	}
	terrainID := EffectiveTerrain(state, tile.ID)
	if terrain, ok := runtime.Movement.Terrain(terrainID); ok && (terrain.CityFood != 0 || terrain.CityProduction != 0 || terrain.CityGold != 0) {
		return tileYieldWithOverlays(state, tile.ID, Yield{Food: terrain.CityFood, Production: terrain.CityProduction, Gold: terrain.CityGold}), nil
	}
	return terrainYieldWithRuntime(state, tileID, runtime)
}

func terrainYieldWithRuntime(state State, tileID string, runtime *rules.RuntimeRules) (Yield, error) {
	tile, ok := tileByID(state, tileID)
	if !ok {
		return Yield{}, errors.New("worked tile does not exist")
	}
	terrainID := EffectiveTerrain(state, tile.ID)
	if runtime != nil {
		if terrain, ok := runtime.Movement.Terrain(terrainID); ok && (terrain.Food != 0 || terrain.Production != 0 || terrain.Gold != 0) {
			return tileYieldWithOverlays(state, tile.ID, Yield{Food: terrain.Food, Production: terrain.Production, Gold: terrain.Gold}), nil
		}
	}
	var yield Yield
	switch terrainID {
	case TerrainPlains, "TERRAIN_PLAINS", "TERRAIN_GRASSLAND", "TERRAIN_TUNDRA", "TERRAIN_DESERT", "TERRAIN_SWAMP":
		yield = Yield{Food: 2, Production: 3, Gold: 1}
	case TerrainForest, "TERRAIN_FOREST", "TERRAIN_JUNGLE":
		yield = Yield{Food: 1, Production: 2}
	case TerrainOcean, "TERRAIN_WATER_SHALLOW", "TERRAIN_WATER_DEEP", "TERRAIN_WATER_SHELF", "TERRAIN_WATER_BEACH", "TERRAIN_WATER_KELP", "TERRAIN_WATER_REEF":
		yield = Yield{Food: 1, Gold: 2}
	default:
		return Yield{}, errors.New("unsupported terrain yield")
	}
	return tileYieldWithOverlays(state, tile.ID, yield), nil
}

func tileYieldWithOverlays(state State, tileID string, yield Yield) Yield {
	for _, overlay := range state.Overlays {
		if overlay.TileID == tileID && overlay.Kind == "improvement" && overlay.Remaining == 0 {
			yield.Food += overlay.YieldFood
			yield.Production += overlay.YieldProduction
			yield.Gold += overlay.YieldGold
		}
	}
	return yield
}

func addYield(a, b Yield) Yield {
	a.Food += b.Food
	a.Production += b.Production
	a.Gold += b.Gold
	a.Combat += b.Combat
	a.Vision += b.Vision
	a.Victory += b.Victory
	return a
}

func validateProductionItem(item ProductionItem) error {
	if item.ID == "" || item.Cost < 1 || item.Upkeep < 0 || (item.Scope != "" && item.Scope != "city" && item.Scope != "nation" && item.Scope != "global") || (item.Kind != ProductionUnit && item.Kind != ProductionBuilding && item.Kind != ProductionWonder) {
		return errors.New("production item is invalid")
	}
	return validateEffects(item.Effects)
}

func validateCityState(state State) error {
	if state.RecalculatedTurn < 0 || state.RecalculatedTurn > state.Phase.Turn {
		return errors.New("recalculated turn is invalid")
	}
	if err := uniqueIDs("wonder", wonderIDs(state.Wonders)); err != nil {
		return err
	}
	for _, city := range state.Cities {
		if city.ID == "" || city.OwnerID == "" || city.TileID == "" {
			return errors.New("city is invalid")
		}
		if len(state.Tiles) > 0 {
			if _, ok := tileByID(state, city.TileID); !ok {
				return errors.New("city tile does not exist")
			}
		}
		if err := validateEffects(city.Effects); err != nil {
			return err
		}
		for _, building := range city.Buildings {
			if building.ID == "" || building.Upkeep < 0 || validateEffects(building.Effects) != nil {
				return errors.New("building is invalid")
			}
		}
		for _, item := range city.Queue {
			if err := validateProductionItem(item); err != nil {
				return err
			}
		}
	}
	for _, wonder := range state.Wonders {
		if wonder.ID == "" || wonder.OwnerID == "" || wonder.CityID == "" || (wonder.Scope != "" && wonder.Scope != "nation" && wonder.Scope != "global") {
			return errors.New("wonder is invalid")
		}
		if err := validateEffects(wonder.Effects); err != nil {
			return err
		}
	}
	for _, nation := range state.Nations {
		if err := validateEffects(nation.Effects); err != nil {
			return err
		}
	}
	return nil
}

func validateEffects(effects []Effect) error {
	for _, effect := range effects {
		if effect.ID == "" {
			return errors.New("effect id is required")
		}
		switch effect.Kind {
		case EffectSettler, EffectSettleSize, EffectSettleBuilding, EffectYieldFood, EffectYieldProduction, EffectYieldGold, EffectYieldGoldPerPopulation, EffectHappiness, EffectCrimePercent, EffectMaxPopulation, EffectOvercrowdingLevel, EffectStarvationProtection, EffectAirport, EffectMounted, EffectNoFuelThenCrash, EffectCantCaptureCity, EffectCanEntrench, EffectCanPillage, EffectBuildingRemovesPopulation, EffectIgnoreZOC, EffectCombatDefenseVsMountedPercent, EffectCombatDefenseVsAirPercent, EffectWoodenShip, EffectSubmarine, EffectCanAttackLand, EffectCanAttackSea, EffectCanAttackAir, EffectCanAttackSpace, EffectCanAttackUnderwater, EffectCombatAttackVsWoodenShipPercent, EffectCityWalls, EffectIgnoreCityWalls, EffectProtectFromNukes, EffectProtectFromBioAgents, EffectProtectFromNanoVirus, EffectPreventConversion, EffectPreventSlavery, EffectForceField, EffectFoodVatPercent, EffectAllowGrunts, EffectBrokerage, EffectCapitol, EffectCathedral, EffectDoubleTelevangelists, EffectIsReligious, EffectLowerPeaceMovementPercent, EffectNuclearPlant, EffectTelevision, EffectEliminateNukes, EffectCloseEmbassies, EffectEmbassiesEverywhere, EffectEnableParkRangers, EffectFreeSlaves, EffectGlobalRadar, EffectIncreaseBoatMovement, EffectIncreaseHP, EffectIncreaseBrokerages, EffectIncreaseRegard, EffectIncreaseScientists, EffectPreventConversionEmpire, EffectProhibitSlavers, EffectProtectFromBarbarians, EffectReduceReadinessCost, EffectReformCities, EffectRevoltingCitiesJoinPlayer, EffectSpiesEverywhere, EffectStartGaiaController, EffectTemporaryFullHappiness, EffectPollutersToParks, EffectGoldPerWaterTradeRoute, EffectGoldPerInternationalTradeRoute, EffectGoldPerTelevision, EffectDecreaseEmpireSize, EffectPollutionPercent, EffectPollutionPopulationPercent, EffectPollutionProductionPercent, EffectPollution, EffectYieldFoodPercent, EffectYieldProductionPercent, EffectYieldGoldPercent, EffectCombat, EffectVision, EffectProduction, EffectVictory, EffectResearch, EffectResearchPercent, EffectMaxSciencePercent, EffectBonus, EffectCombatAttack, EffectCombatAttackLand, EffectCombatAttackSea, EffectCombatAttackAir, EffectCombatDefense, EffectSoothsaying, EffectFranchise, EffectRulerAssassinated, EffectFreedPopulation, EffectInjunction, EffectLawsuit, EffectSpecialEmbassy, EffectSpecialSoothsaying, EffectSpecialSteal, EffectSpecialCapture, EffectSpecialTerraform, EffectSpecialTerraformTurns, EffectSpecialConvertCity, EffectSpecialCreateFranchise, EffectSpecialInvestigateCity, EffectSpecialPlantNuke, EffectSpecialBioTerror, EffectSpecialNanoTerror, EffectSpecialPlague, EffectSpecialAssassinateRuler, EffectSpecialSlaveRaid, EffectSpecialSettlerSlaveRaid, EffectSpecialUndergroundRailway, EffectSpecialInjoin, EffectSpecialReform, EffectSpecialCreatePark, EffectSpecialExpel, EffectSpecialSue, EffectSpecialConvertCityDeathChance, EffectSpecialCreateFranchiseDeathChance, EffectSpecialInvestigateCityDeathChance, EffectSpecialPlantNukeDeathChance, EffectSpecialBioTerrorDeathChance, EffectSpecialNanoTerrorDeathChance, EffectSpecialPlagueDeathChance, EffectSpecialAssassinateRulerDeathChance, EffectSpecialSlaveRaidDeathChance, EffectSpecialSettlerSlaveRaidDeathChance, EffectSpecialUndergroundRailwayDeathChance:
		default:
			return fmt.Errorf("unsupported effect kind %q", effect.Kind)
		}
	}
	return nil
}

func cityByID(state State, id string) (*City, bool) {
	for i := range state.Cities {
		if state.Cities[i].ID == id {
			return &state.Cities[i], true
		}
	}
	return nil, false
}

func cityHasBuilding(city City, id string) bool {
	for _, building := range city.Buildings {
		if building.ID == id && !isObsolete(city.Obsolete, building.ObsoletedBy) {
			return true
		}
	}
	return false
}

func wonderExists(state State, id string) bool {
	for _, wonder := range state.Wonders {
		if wonder.ID == id && !wonder.Destroyed {
			return true
		}
	}
	return false
}

func requireProductionUnlock(state State, city City, item ProductionItem, runtime *rules.RuntimeRules) error {
	prerequisite := item.Prerequisite
	obsolete := item.ObsoletedBy
	if runtime != nil {
		if production, ok := runtime.ProductionDefinition(item.ID); ok {
			if production.CantBuildInSea && cityIsSea(state, city, runtime) || production.CoastalOnly && !cityIsCoastal(state, city, runtime) {
				return errors.New("production prerequisite is not met")
			}
			if prerequisite == "" && len(production.Prerequisites) > 0 {
				prerequisite = production.Prerequisites[0]
			} else if prerequisite != "" && len(production.Prerequisites) > 0 && !containsString(production.Prerequisites, prerequisite) {
				return errors.New("production prerequisite is not met")
			}
			for _, requiredBuilding := range production.RequiredBuildings {
				if !cityHasBuilding(city, requiredBuilding) {
					return errors.New("production prerequisite is not met")
				}
			}
			if obsolete == "" && len(production.ObsoleteBy) > 0 {
				obsolete = production.ObsoleteBy[0]
			}
		}
		if item.Kind == ProductionUnit {
			if unit, ok := runtime.Unit(item.ID); ok {
				if prerequisite == "" && len(unit.Prerequisites) > 0 {
					prerequisite = unit.Prerequisites[0]
				} else if prerequisite != "" && len(unit.Prerequisites) > 0 && !containsString(unit.Prerequisites, prerequisite) {
					return errors.New("production prerequisite is not met")
				}
				if obsolete == "" && len(unit.ObsoleteBy) > 0 {
					obsolete = unit.ObsoleteBy[0]
				}
			}
		}
	}
	if !hasPrerequisite(state, city, prerequisite) || isProductionObsolete(state, city, obsolete) {
		return errors.New("production prerequisite is not met")
	}
	return nil
}

func cityIsSea(state State, city City, runtime *rules.RuntimeRules) bool {
	tile, ok := tileByID(state, city.TileID)
	if !ok {
		return false
	}
	return terrainIsSea(runtime, tile.Terrain)
}

func cityIsCoastal(state State, city City, runtime *rules.RuntimeRules) bool {
	cityTile, ok := tileByID(state, city.TileID)
	if !ok {
		return false
	}
	if terrainIsSea(runtime, cityTile.Terrain) {
		return true
	}
	for _, tile := range state.Tiles {
		if adjacent(cityTile, tile) && terrainIsSea(runtime, tile.Terrain) {
			return true
		}
	}
	return false
}

func terrainIsSea(runtime *rules.RuntimeRules, terrainID string) bool {
	if terrainID == TerrainOcean {
		return true
	}
	if runtime == nil || !runtime.Movement.Configured {
		return false
	}
	terrain, ok := runtime.Movement.Terrain(terrainID)
	return ok && containsString(terrain.Domains, DomainSea) && !containsString(terrain.Domains, DomainLand)
}

func hasPrerequisite(state State, city City, prerequisite string) bool {
	if prerequisite == "" {
		return true
	}
	for _, building := range city.Buildings {
		if building.ID == prerequisite {
			return true
		}
	}
	if wonderExists(state, prerequisite) {
		return true
	}
	nationID := nationForPlayer(state, city.OwnerID)
	for _, nation := range state.Nations {
		if nation.ID == nationID && containsString(nation.Advances, prerequisite) {
			return true
		}
	}
	return false
}

func isProductionObsolete(state State, city City, obsoleteID string) bool {
	if obsoleteID == "" {
		return false
	}
	nationID := nationForPlayer(state, city.OwnerID)
	for _, nation := range state.Nations {
		if nation.ID == nationID && containsString(nation.Advances, obsoleteID) {
			return true
		}
	}
	return wonderExists(state, obsoleteID)
}

func nationForPlayer(state State, playerID string) string {
	for _, player := range state.Players {
		if player.ID == playerID {
			return player.NationID
		}
	}
	return ""
}

func hasEffect(effects []Effect, kind string) bool {
	for _, effect := range effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func isObsoleteForPlayer(state State, playerID string, obsolete []string, required string) bool {
	if isObsolete(obsolete, required) {
		return true
	}
	if required == "" {
		return false
	}
	nationID := nationForPlayer(state, playerID)
	for _, nation := range state.Nations {
		if nation.ID == nationID && containsString(nation.Advances, required) {
			return true
		}
	}
	return false
}

func isObsolete(obsolete []string, required string) bool {
	if required == "" {
		return false
	}
	for _, id := range obsolete {
		if id == required {
			return true
		}
	}
	return false
}

func wonderIDs(wonders []Wonder) []string {
	ids := make([]string, len(wonders))
	for i := range wonders {
		ids[i] = wonders[i].ID
	}
	return ids
}

func decodeCityPayload(raw json.RawMessage, target any) error { return json.Unmarshal(raw, target) }
