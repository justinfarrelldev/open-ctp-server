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

	EffectSettler                = "settler"
	EffectYieldFood              = "yield_food"
	EffectYieldProduction        = "yield_production"
	EffectYieldGold              = "yield_gold"
	EffectYieldProductionPercent = "yield_production_percent"
	EffectYieldGoldPercent       = "yield_gold_percent"
	EffectCombat                 = "combat"
	EffectVision                 = "vision"
	EffectProduction             = "production"
	EffectVictory                = "victory"
	EffectResearch               = "research"
	EffectBonus                  = "bonus"
	EffectCombatAttack           = "combat_attack"
	EffectCombatDefense          = "combat_defense"
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
	removeUnitAndCargo(state, unit.ID)
	state.Cities = append(state.Cities, City{ID: payload.CityID, Name: payload.Name, OwnerID: actor, TileID: tileID})
	return nil
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
	sort.Slice(next.Cities, func(i, j int) bool { return next.Cities[i].ID < next.Cities[j].ID })
	for i := range next.Cities {
		city := &next.Cities[i]
		if err := completeProduction(&next, city, runtime); err != nil {
			return State{}, err
		}
		city.Yields, err = cityYield(next, *city)
		if err != nil {
			return State{}, err
		}
	}
	next.RecalculatedTurn = next.Phase.Turn
	return next, nil
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
	if len(city.Queue) == 0 {
		return nil
	}
	base, err := terrainYield(*state, city.TileID)
	if err != nil {
		return err
	}
	city.StoredProduction += base.Production
	item := city.Queue[0]
	if city.StoredProduction < item.Cost {
		return nil
	}
	city.StoredProduction -= item.Cost
	city.Queue = city.Queue[1:]
	switch item.Kind {
	case ProductionUnit:
		return completeUnitProduction(state, *city, item, runtime)
	case ProductionBuilding:
		city.Buildings = append(city.Buildings, Building{ID: item.ID, Upkeep: item.Upkeep, ObsoletedBy: item.ObsoletedBy, Effects: item.Effects})
		return nil
	case ProductionWonder:
		if wonderExists(*state, item.ID) {
			city.StoredProduction = 0
			return nil
		}
		state.Wonders = append(state.Wonders, Wonder{ID: item.ID, OwnerID: city.OwnerID, CityID: city.ID, Scope: item.Scope, ObsoletedBy: item.ObsoletedBy, Effects: item.Effects})
		return nil
	default:
		return errors.New("unsupported production kind")
	}
}

func completeUnitProduction(state *State, city City, item ProductionItem, runtime *rules.RuntimeRules) error {
	if runtime == nil {
		return nil
	}
	definition, ok := runtime.Unit(item.ID)
	if !ok {
		return nil
	}
	unit := Unit{
		ID:              producedUnitID(*state, city.ID, item.ID),
		OwnerID:         city.OwnerID,
		TileID:          city.TileID,
		Domain:          definition.Domain,
		Movement:        definition.Movement,
		CargoCapacity:   definition.TransportCapacity,
		CargoDomains:    append([]string(nil), definition.CargoDomains...),
		ZOC:             definition.ZOC,
		Effects:         runtimeUnitEffects(definition),
		Attack:          definition.Attack,
		Defense:         definition.Defense,
		Strength:        definition.Strength,
		Armor:           definition.Armor,
		Health:          definition.Health,
		MaxHealth:       definition.Health,
		VisibilityRange: definition.VisibilityRange,
	}
	state.Units = append(state.Units, unit)
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
	yield, err := terrainYield(state, city.TileID)
	if err != nil {
		return Yield{}, err
	}
	seen := map[string]bool{city.TileID: true}
	for _, tileID := range city.WorkedTileIDs {
		if seen[tileID] {
			return Yield{}, errors.New("worked tile is duplicated")
		}
		worked, err := terrainYield(state, tileID)
		if err != nil {
			return Yield{}, err
		}
		seen[tileID] = true
		yield = addYield(yield, worked)
	}
	effects := cityEffects(state, city)
	for _, effect := range effects {
		switch effect.Kind {
		case EffectYieldFood:
			yield.Food += effect.Value
		case EffectYieldProduction, EffectProduction:
			yield.Production += effect.Value
		case EffectYieldGold:
			yield.Gold += effect.Value
		case EffectYieldProductionPercent:
			yield.Production += percentBonus(yield.Production, effect.Value)
		case EffectYieldGoldPercent:
			yield.Gold += percentBonus(yield.Gold, effect.Value)
		case EffectCombat:
			yield.Combat += effect.Value
		case EffectVision:
			yield.Vision += effect.Value
		case EffectVictory:
			yield.Victory += effect.Value
		}
	}
	for _, building := range city.Buildings {
		if !isObsolete(city.Obsolete, building.ObsoletedBy) {
			yield.Production -= building.Upkeep
		}
	}
	if yield.Production < 0 {
		yield.Production = 0
	}
	return yield, nil
}

func percentBonus(value, percent int) int {
	if value <= 0 || percent <= 0 {
		return 0
	}
	return (value*percent + 99) / 100
}

func cityEffects(state State, city City) []Effect {
	effects := append([]Effect(nil), city.Effects...)
	if nationID := nationForPlayer(state, city.OwnerID); nationID != "" {
		for _, nation := range state.Nations {
			if nation.ID == nationID {
				effects = append(effects, nation.Effects...)
			}
		}
	}
	for _, building := range city.Buildings {
		if !isObsolete(city.Obsolete, building.ObsoletedBy) {
			effects = append(effects, building.Effects...)
		}
	}
	for _, wonder := range state.Wonders {
		if !wonder.Destroyed && !isObsolete(city.Obsolete, wonder.ObsoletedBy) && (wonder.Scope == "global" || wonder.OwnerID == city.OwnerID) {
			effects = append(effects, wonder.Effects...)
		}
	}
	normalizeEffects(effects)
	return effects
}

func terrainYield(state State, tileID string) (Yield, error) {
	tile, ok := tileByID(state, tileID)
	if !ok {
		return Yield{}, errors.New("worked tile does not exist")
	}
	var yield Yield
	switch EffectiveTerrain(state, tile.ID) {
	case TerrainPlains:
		yield = Yield{Food: 2, Production: 3, Gold: 1}
	case TerrainForest:
		yield = Yield{Food: 1, Production: 2}
	case TerrainOcean:
		yield = Yield{Food: 1, Gold: 2}
	default:
		return Yield{}, errors.New("unsupported terrain yield")
	}
	for _, overlay := range state.Overlays {
		if overlay.TileID == tile.ID && overlay.Kind == "improvement" && overlay.Remaining == 0 {
			yield.Food += overlay.YieldFood
			yield.Production += overlay.YieldProduction
			yield.Gold += overlay.YieldGold
		}
	}
	return yield, nil
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
		case EffectSettler, EffectYieldFood, EffectYieldProduction, EffectYieldGold, EffectYieldProductionPercent, EffectYieldGoldPercent, EffectCombat, EffectVision, EffectProduction, EffectVictory, EffectResearch, EffectBonus, EffectCombatAttack, EffectCombatDefense, EffectSoothsaying, EffectSpecialEmbassy, EffectSpecialSoothsaying, EffectSpecialSteal, EffectSpecialCapture, EffectSpecialTerraform, EffectSpecialTerraformTurns:
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
	if runtime != nil && item.Kind == ProductionUnit {
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
	if !hasPrerequisite(state, city, prerequisite) || isProductionObsolete(state, city, obsolete) {
		return errors.New("production prerequisite is not met")
	}
	return nil
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
