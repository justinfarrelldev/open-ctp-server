package rules

import (
	"math"
	"sort"
	"strconv"
)

// ProjectRetailRuntime maps the source-preserving retail catalog into the
// server's client-safe runtime-definition shape. It does not claim the engine can
// execute every behavior-bearing retail flag yet; unsupported mechanics remain in
// RetailRules.RawFields while supported scalar values are projected for clients,
// AI, and future engine work.
func ProjectRetailRuntime(db RetailRules) RuntimeRules {
	runtime := RuntimeRules{
		CompatibilityVersion: CTP2CompatibilityVersion,
		ContentHash:          "retail-catalog-projection",
		Calendar:             RuntimeCalendar{FinalTurn: calendarFinalTurn},
		Movement:             RuntimeMovement{Configured: true, PolicyConfigured: true, OpenAccess: false, ZOCStopsPath: true},
		PollutionModel:       projectRetailPollutionModel(db.PollutionModel),
	}
	for _, advance := range db.Advances {
		runtime.Advances = append(runtime.Advances, RuntimeAdvanceDefinition{ID: advance.ID, Cost: advance.Cost, Prerequisites: append([]string(nil), advance.Prerequisites...)})
	}
	for _, unit := range db.Units {
		projected := projectRetailUnit(unit)
		runtime.Units = append(runtime.Units, projected)
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: unit.ID, Kind: "unit"})
		if unit.ID == "UNIT_SETTLER" {
			runtime.Starter = RuntimeStarter{Domain: projected.Domain, Movement: projected.Movement, Attack: projected.Attack, Defense: projected.Defense, Ranged: projected.Ranged, Strength: projected.Strength, Armor: projected.Armor, Health: projected.Health, MaxFuel: projected.MaxFuel, ZOC: projected.ZOC, TransportCapacity: projected.TransportCapacity, CargoDomains: append([]string(nil), projected.CargoDomains...), VisibilityRange: projected.VisibilityRange, Effects: append([]RuntimeStarterEffect(nil), projected.Effects...)}
		}
	}
	buildingEffects := make(map[string][]RuntimeStarterEffect, len(db.Buildings))
	for _, building := range db.Buildings {
		effects := retailBuildingRuntimeEffects(building)
		buildingEffects[building.ID] = effects
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: building.ID, Kind: "building", Cost: building.ProductionCost, Upkeep: building.Upkeep, Prerequisites: singleOrEmpty(building.EnableAdvance), ObsoleteBy: append([]string(nil), building.ObsoleteAdvance...), Effects: effects})
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: building.ID, Kind: "building"})
	}
	for _, wonder := range db.Wonders {
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: wonder.ID, Kind: "wonder", Cost: wonder.ProductionCost, Prerequisites: singleOrEmpty(wonder.EnableAdvance), ObsoleteBy: append([]string(nil), wonder.ObsoleteAdvance...), Effects: retailWonderRuntimeEffects(wonder, buildingEffects)})
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: wonder.ID, Kind: "wonder"})
		if containsRetailFlag(wonder.Flags, "StartGaiaController") {
			runtime.Gaia = RuntimeGaia{ActivationTurns: 10, BuildingID: "IMPROVE_GAIA_COMPUTER", WonderID: wonder.ID}
		}
	}
	for _, terrain := range db.Terrains {
		runtime.Movement.Terrains = append(runtime.Movement.Terrains, RuntimeTerrain{ID: terrain.ID, Cost: positiveOrOne(terrain.EnvBase.Movement), Domains: retailDomains(terrain.MovementTypes), Food: terrain.EnvBase.Food, Production: terrain.EnvBase.Shield, Gold: terrain.EnvBase.Gold, CityFood: terrain.EnvCity.Food, CityProduction: terrain.EnvCity.Shield, CityGold: terrain.EnvCity.Gold})
	}
	for _, improvement := range db.TerrainImprovements {
		runtime.TerrainImprovements = append(runtime.TerrainImprovements, projectRetailTerrainImprovement(improvement))
	}
	for _, government := range db.Governments {
		runtime.Governments = append(runtime.Governments, RuntimeGovernmentDefinition{ID: government.ID, EnableAdvanceID: government.EnableAdvance, ResearchPercent: signedCoefficientPercentPoints(government.FloatValues["KnowledgeCoef"]), FoodPercent: signedCoefficientPercentPoints(government.FloatValues["FoodCoef"]), ProductionPercent: signedCoefficientPercentPoints(government.FloatValues["ProductionCoef"]), GoldPercent: signedCoefficientPercentPoints(government.FloatValues["GoldCoef"]), PollutionPercent: signedCoefficientPercentPoints(government.PollutionCoef)})
	}
	for _, trigger := range db.PollutionTriggerLists {
		runtime.PollutionTriggers = append(runtime.PollutionTriggers, RuntimePollutionTriggerDefinition{ID: trigger.ID, TriggerLevels: append([]int64(nil), trigger.TriggerLevels...), DisasterTypes: append([]string(nil), trigger.DisasterTypes...)})
	}
	if ai := projectRetailAI(db.AI); ai != nil {
		runtime.AI = ai
	}
	sortRuntimeProjection(&runtime)
	return runtime
}

func projectRetailAI(ai RetailAILists) *RuntimeAI {
	projected := &RuntimeAI{
		UnitBuildLists:     projectRetailAILists(ai.UnitBuildLists),
		BuildingBuildLists: projectRetailAILists(ai.BuildingBuildLists),
		WonderBuildLists:   projectRetailAILists(ai.WonderBuildLists),
		AdvanceLists:       projectRetailAILists(ai.AdvanceLists),
		ImprovementLists:   projectRetailAILists(ai.ImprovementLists),
	}
	if len(projected.UnitBuildLists) == 0 && len(projected.BuildingBuildLists) == 0 && len(projected.WonderBuildLists) == 0 && len(projected.AdvanceLists) == 0 && len(projected.ImprovementLists) == 0 {
		return nil
	}
	return projected
}

func projectRetailAILists(lists []RetailAIList) []RuntimeAIList {
	projected := make([]RuntimeAIList, 0, len(lists))
	for _, list := range lists {
		if list.ID == "" {
			continue
		}
		projected = append(projected, RuntimeAIList{ID: list.ID, Items: append([]string(nil), list.Items...)})
	}
	return projected
}

func projectRetailPollutionModel(model RetailPollutionModel) *RuntimePollutionModel {
	if model == (RetailPollutionModel{}) {
		return nil
	}
	return &RuntimePollutionModel{PopulationStart: model.PopulationStart, PopulationRatio: model.PopulationRatio, ProductionStart: model.ProductionStart, ProductionRatio: model.ProductionRatio}
}

func retailBuildingRuntimeEffects(building RetailBuilding) []RuntimeStarterEffect {
	effects := make([]RuntimeStarterEffect, 0, 3)
	if value := percentPoints(building.FloatValues["ProductionPercent"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":production-percent", Kind: "yield_production_percent", Value: value})
	}
	if value := percentPoints(building.FloatValues["CommercePercent"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":commerce-percent", Kind: "yield_gold_percent", Value: value})
	}
	if value := percentPoints(building.FloatValues["FoodPercent"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":food-percent", Kind: "yield_food_percent", Value: value})
	}
	if value := percentPoints(building.FloatValues["SciencePercent"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":science-percent", Kind: "research_percent", Value: value})
	}
	if value := building.IntValues["GoldPerCitizen"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":gold-per-citizen", Kind: "yield_gold_per_population", Value: value})
	}
	if value := building.IntValues["RaiseMaxPopulation"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":max-population", Kind: "max_population", Value: value})
	}
	if value := building.IntValues["StarvationProtection"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":starvation-protection", Kind: "starvation_protection", Value: value})
	}
	if retailContainsString(building.Flags, "Airport") {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":airport", Kind: "airport"})
	}
	if value := retailPercentFieldPoints(building.FloatValues["DefendersPercent"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":defenders-percent", Kind: "combat_defense", Value: value})
	}
	if value := retailPercentFieldPoints(building.FloatValues["OffenseBonusLand"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":offense-land", Kind: "combat_attack_land", Value: value})
	}
	if value := retailPercentFieldPoints(building.FloatValues["OffenseBonusWater"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":offense-water", Kind: "combat_attack_sea", Value: value})
	}
	if value := retailPercentFieldPoints(building.FloatValues["OffenseBonusAir"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":offense-air", Kind: "combat_attack_air", Value: value})
	}
	if value, ok := signedPercentPoints(building.FloatValues["PopulationPollutionPercent"]); ok {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":population-pollution-percent", Kind: "pollution_population_percent", Value: value})
	}
	if value, ok := signedPercentPoints(building.FloatValues["ProductionPollutionPercent"]); ok {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":production-pollution-percent", Kind: "pollution_production_percent", Value: value})
	}
	sort.Slice(effects, func(i, j int) bool { return effects[i].Kind < effects[j].Kind })
	return effects
}

func containsRetailFlag(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func retailWonderRuntimeEffects(wonder RetailWonder, buildingEffects map[string][]RuntimeStarterEffect) []RuntimeStarterEffect {
	effects := make([]RuntimeStarterEffect, 0, 4)
	if value := wonder.IntValues["BonusGold"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":bonus-gold", Kind: "yield_gold", Value: value})
	}
	if value := wonder.IntValues["IncKnowledgePercent"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":knowledge-percent", Kind: "research_percent", Value: value})
	}
	if value := wonder.IntValues["IncreaseFoodAllCities"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":food-all-cities", Kind: "yield_food_percent", Value: value})
	}
	if value := wonder.IntValues["IncreaseProduction"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":production", Kind: "yield_production_percent", Value: value})
	}
	for _, buildingID := range append(retailFieldValues(wonder.RawFields, "BuildingEverywhere"), retailFieldValues(wonder.RawFields, "ActualBuildingEverywhere")...) {
		for _, effect := range buildingEffects[buildingID] {
			effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":" + effect.ID, Kind: effect.Kind, Value: effect.Value})
		}
	}
	sort.Slice(effects, func(i, j int) bool {
		if effects[i].Kind != effects[j].Kind {
			return effects[i].Kind < effects[j].Kind
		}
		return effects[i].ID < effects[j].ID
	})
	return effects
}

func signedPercentPoints(value float64) (int, bool) {
	if value == 0 {
		return 0, false
	}
	return int(math.Round(value * 100)), true
}

func retailPercentFieldPoints(value float64) int {
	if value <= 0 {
		return 0
	}
	if value <= 1 {
		return int(math.Round(value * 100))
	}
	return int(math.Round(value))
}

func percentPoints(value float64) int {
	if value <= 0 {
		return 0
	}
	return int(math.Round(value * 100))
}

func signedCoefficientPercentPoints(value float64) int {
	if value == 0 || value == 1 {
		return 0
	}
	return int(math.Round((value - 1) * 100))
}

func projectRetailUnit(unit RetailUnit) RuntimeUnitDefinition {
	effects := retailUnitEffects(unit)
	cargoDomains := retailCargoDomains(unit.CanCarry)
	return RuntimeUnitDefinition{ID: unit.ID, Role: unit.Category, Domain: retailUnitDomain(unit.MovementTypes), Movement: positiveOrOne(unit.MaxMovePoints), Attack: unit.Attack, Defense: unit.Defense, Ranged: unit.ZBRangeAttack, BombardRange: unit.BombardRange, BombRounds: unit.BombRounds, Strength: unit.Attack, Armor: unit.Armor, Health: positiveOrOne(unit.MaxHP), MaxFuel: unit.MaxFuel, ZOC: !retailContainsString(unit.Flags, "NoZoc"), TransportCapacity: unit.CargoMax, CargoDomains: cargoDomains, ProductionCost: unit.ShieldCost, Upkeep: unit.GoldHunger, Prerequisites: singleOrEmpty(unit.EnableAdvance), ObsoleteBy: append([]string(nil), unit.ObsoleteAdvance...), VisibilityRange: unit.VisionRange, Effects: effects}
}

func retailUnitEffects(unit RetailUnit) []RuntimeStarterEffect {
	effects := make([]RuntimeStarterEffect, 0)
	if unit.Category == "UNIT_CATEGORY_SETTLER" || len(retailFieldValues(unit.RawFields, "Settle")) > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("settler"), Kind: "settler"})
		if size := firstRetailUnitInt(unit, "SettleSize"); size > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("settle_size"), Kind: "settle_size", Value: size})
		}
		for _, buildingID := range retailFieldValues(unit.RawFields, "SettleBuilding") {
			effects = append(effects, RuntimeStarterEffect{ID: buildingID, Kind: "settle_building"})
		}
	}
	if retailContainsString(unit.Flags, "EstablishEmbassy") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_embassy"), Kind: "special_embassy"})
	}
	if len(unit.FieldsByKey("CanSoothsay")) > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_soothsaying"), Kind: "special_soothsaying"})
	}
	if fields := unit.FieldsByKey("StealTechnology"); len(fields) > 0 {
		chance := percentPoints(retailRawFieldFloat(fields[0], "RandomChance"))
		if chance <= 0 {
			chance = 100
		}
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_steal"), Kind: "special_steal", Value: chance})
		if deathChance := percentPoints(retailRawFieldFloat(fields[0], "DeathChance")); deathChance > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_capture"), Kind: "special_capture", Value: deathChance})
		}
	}
	sort.Slice(effects, func(i, j int) bool {
		if effects[i].Kind != effects[j].Kind {
			return effects[i].Kind < effects[j].Kind
		}
		return effects[i].ID < effects[j].ID
	})
	return effects
}

func retailRawFieldFloat(field RetailRawField, key string) float64 {
	value := field.FirstValue(key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func firstRetailUnitInt(unit RetailUnit, key string) int {
	values := retailFieldValues(unit.RawFields, key)
	if len(values) == 0 {
		return 0
	}
	value, err := strconv.Atoi(values[0])
	if err != nil {
		return 0
	}
	return value
}

func (u RetailUnit) FieldsByKey(key string) []RetailRawField {
	return retailFieldsByKey(u.RawFields, key)
}

func projectRetailTerrainImprovement(improvement RetailTerrainImprovement) RuntimeTerrainImprovementDefinition {
	definition := RuntimeTerrainImprovementDefinition{ID: improvement.ID, Class: improvement.Class, BuildTurns: 1, PublicWorksCost: 1, TargetTerrain: improvement.TerraformTerrain, CantBuildOn: append([]string(nil), improvement.CantBuildOn...), Excludes: append([]string(nil), improvement.Excludes...)}
	if len(improvement.Effects) == 0 {
		return definition
	}
	effect := improvement.Effects[0]
	definition.BuildTurns = positiveOrOne(effect.ProductionTime)
	definition.PublicWorksCost = positiveOrOne(effect.ProductionCost)
	definition.YieldFood = effect.BonusFood
	definition.YieldProduction = effect.BonusProduction
	definition.YieldGold = effect.BonusGold
	definition.MovementCostDelta = effect.MoveCost
	return definition
}

func retailUnitDomain(movementTypes []string) string {
	hasLand := retailContainsString(movementTypes, "Land") || retailContainsString(movementTypes, "Mountain")
	hasSea := retailContainsString(movementTypes, "Sea") || retailContainsString(movementTypes, "ShallowWater")
	switch {
	case retailContainsString(movementTypes, "Air"):
		return "air"
	case retailContainsString(movementTypes, "Space"):
		return "space"
	case hasLand && hasSea:
		return "amphibious"
	case hasSea:
		return "sea"
	default:
		return "land"
	}
}

func retailDomains(movementTypes []string) []string {
	domains := make([]string, 0, len(movementTypes))
	add := func(domain string) {
		if domain != "" && !retailContainsString(domains, domain) {
			domains = append(domains, domain)
		}
	}
	for _, movementType := range movementTypes {
		switch movementType {
		case "Land", "Mountain":
			add("land")
		case "Sea", "ShallowWater":
			add("sea")
		case "Air":
			add("air")
		case "Space":
			add("space")
		}
	}
	sort.Strings(domains)
	return domains
}

func retailCargoDomains(canCarry []string) []string {
	domains := make([]string, 0)
	for _, carry := range canCarry {
		switch carry {
		case "SmallLand", "MedLand", "LargeLand":
			if !retailContainsString(domains, "land") {
				domains = append(domains, "land")
			}
		case "SmallAir", "MedAir", "LargeAir":
			if !retailContainsString(domains, "air") {
				domains = append(domains, "air")
			}
		case "SmallWater", "MedWater", "LargeWater":
			if !retailContainsString(domains, "sea") {
				domains = append(domains, "sea")
			}
		case "SmallSpace", "MedSpace", "LargeSpace":
			if !retailContainsString(domains, "space") {
				domains = append(domains, "space")
			}
		}
	}
	sort.Strings(domains)
	return domains
}

func retailContainsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func singleOrEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func positiveOrOne(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

func sortRuntimeProjection(runtime *RuntimeRules) {
	sort.Slice(runtime.Advances, func(i, j int) bool { return runtime.Advances[i].ID < runtime.Advances[j].ID })
	sort.Slice(runtime.Units, func(i, j int) bool { return runtime.Units[i].ID < runtime.Units[j].ID })
	sort.Slice(runtime.Production, func(i, j int) bool { return runtime.Production[i].ID < runtime.Production[j].ID })
	sort.Slice(runtime.Effects, func(i, j int) bool { return runtime.Effects[i].ID < runtime.Effects[j].ID })
	sort.Slice(runtime.Movement.Terrains, func(i, j int) bool { return runtime.Movement.Terrains[i].ID < runtime.Movement.Terrains[j].ID })
	sort.Slice(runtime.TerrainImprovements, func(i, j int) bool { return runtime.TerrainImprovements[i].ID < runtime.TerrainImprovements[j].ID })
	sort.Slice(runtime.Governments, func(i, j int) bool { return runtime.Governments[i].ID < runtime.Governments[j].ID })
	sort.Slice(runtime.PollutionTriggers, func(i, j int) bool { return runtime.PollutionTriggers[i].ID < runtime.PollutionTriggers[j].ID })
}
