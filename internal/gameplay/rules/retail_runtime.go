package rules

import (
	"math"
	"sort"
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
	for _, building := range db.Buildings {
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: building.ID, Kind: "building", Cost: building.ProductionCost, Upkeep: building.Upkeep, Prerequisites: singleOrEmpty(building.EnableAdvance), ObsoleteBy: append([]string(nil), building.ObsoleteAdvance...), Effects: retailBuildingRuntimeEffects(building)})
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: building.ID, Kind: "building"})
	}
	for _, wonder := range db.Wonders {
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: wonder.ID, Kind: "wonder", Cost: wonder.ProductionCost, Prerequisites: singleOrEmpty(wonder.EnableAdvance), ObsoleteBy: append([]string(nil), wonder.ObsoleteAdvance...), Effects: retailWonderRuntimeEffects(wonder)})
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
		runtime.Governments = append(runtime.Governments, RuntimeGovernmentDefinition{ID: government.ID, EnableAdvanceID: government.EnableAdvance, ResearchBonus: coefficientFlatBonus(government.FloatValues["KnowledgeCoef"]), ProductionPercent: coefficientBonusPoints(government.FloatValues["ProductionCoef"]), GoldPercent: coefficientBonusPoints(government.FloatValues["GoldCoef"])})
	}
	for _, trigger := range db.PollutionTriggerLists {
		runtime.PollutionTriggers = append(runtime.PollutionTriggers, RuntimePollutionTriggerDefinition{ID: trigger.ID, TriggerLevels: append([]int64(nil), trigger.TriggerLevels...), DisasterTypes: append([]string(nil), trigger.DisasterTypes...)})
	}
	sortRuntimeProjection(&runtime)
	return runtime
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
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":science-percent", Kind: "research", Value: value})
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

func retailWonderRuntimeEffects(wonder RetailWonder) []RuntimeStarterEffect {
	effects := make([]RuntimeStarterEffect, 0, 4)
	if value := wonder.IntValues["BonusGold"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":bonus-gold", Kind: "yield_gold", Value: value})
	}
	if value := wonder.IntValues["IncKnowledgePercent"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":knowledge-percent", Kind: "research", Value: value})
	}
	if value := wonder.IntValues["IncreaseFoodAllCities"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":food-all-cities", Kind: "yield_food_percent", Value: value})
	}
	if value := wonder.IntValues["IncreaseProduction"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":production", Kind: "yield_production_percent", Value: value})
	}
	sort.Slice(effects, func(i, j int) bool { return effects[i].Kind < effects[j].Kind })
	return effects
}

func percentPoints(value float64) int {
	if value <= 0 {
		return 0
	}
	return int(math.Round(value * 100))
}

func coefficientBonusPoints(value float64) int {
	if value <= 1 {
		return 0
	}
	return int(math.Round((value - 1) * 100))
}

func coefficientFlatBonus(value float64) int {
	if value <= 1 {
		return 0
	}
	return int(math.Round(value - 1))
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
	}
	if retailContainsString(unit.Flags, "EstablishEmbassy") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_embassy"), Kind: "special_embassy"})
	}
	if len(unit.FieldsByKey("StealTechnology")) > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_steal"), Kind: "special_steal", Value: 100})
	}
	sort.Slice(effects, func(i, j int) bool { return effects[i].Kind < effects[j].Kind })
	return effects
}

func (u RetailUnit) FieldsByKey(key string) []RetailRawField {
	return retailFieldsByKey(u.RawFields, key)
}

func projectRetailTerrainImprovement(improvement RetailTerrainImprovement) RuntimeTerrainImprovementDefinition {
	definition := RuntimeTerrainImprovementDefinition{ID: improvement.ID, Class: improvement.Class, BuildTurns: 1, PublicWorksCost: 1, CantBuildOn: append([]string(nil), improvement.CantBuildOn...), Excludes: append([]string(nil), improvement.Excludes...)}
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
