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
		Victory:              []RuntimeVictory{{ID: "bloodlust_conquest"}, {ID: "diplomatic_world_union"}, {ID: "score_2300"}, {ID: "science_gaia_controller"}},
		Calendar:             projectRetailCalendar(db.Calendar),
		Constants:            RuntimeConstants{MaxDisbandSize: db.Constants.MaxDisbandSize, CityGrowthCoefficient: db.Constants.CityGrowthCoefficient, UnconvertedIndulgenceGold: db.Constants.UnconvertedIndulgenceGold, UnconvertedIndulgenceHappiness: db.Constants.UnconvertedIndulgenceHappiness, ConvertedIndulgenceGold: db.Constants.ConvertedIndulgenceGold, ConvertedIndulgenceHappiness: db.Constants.ConvertedIndulgenceHappiness, OtherFaithIndulgenceGold: db.Constants.OtherFaithIndulgenceGold, OtherFaithIndulgenceHappiness: db.Constants.OtherFaithIndulgenceHappiness, ClericConversionPercent: db.Constants.ClericConversionPercent, TelevangelistConversionPercent: db.Constants.TelevangelistConversionPercent, EmancipationUnhappinessTurns: db.Constants.EmancipationUnhappinessTurns, EmancipationUnhappinessAmount: db.Constants.EmancipationUnhappinessAmount, NukePopulationPercent: db.Constants.NukePopulationPercent, NukeKillTiles: db.Constants.NukeKillTiles, PollutionCausedByNuke: db.Constants.PollutionCausedByNuke, PiracyWastePercent: db.Constants.PiracyWastePercent, VeteranCoefPercent: db.Constants.VeteranCoefPercent, BombardDestroyBuildingPercent: db.Constants.BombardDestroyBuildingPercent, BombardKillPopulationPercent: db.Constants.BombardKillPopulationPercent, SpecialActionMoveCost: db.Constants.SpecialActionMoveCost, CaravanCoefPercent: db.Constants.CaravanCoefPercent},
		Movement:             RuntimeMovement{Configured: true, PolicyConfigured: true, OpenAccess: false, ZOCStopsPath: true},
		PollutionModel:       projectRetailPollutionModel(db.PollutionModel),
	}
	for _, age := range db.Ages {
		runtime.Ages = append(runtime.Ages, RuntimeAgeDefinition{ID: age.ID, Index: age.Index, Number: age.Number, NeedAdvances: age.NeedAdvances})
	}
	if len(runtime.Ages) > 0 {
		runtime.StartAgeID = runtime.Ages[0].ID
		runtime.EndAgeID = runtime.Ages[len(runtime.Ages)-1].ID
	}
	for _, advance := range db.Advances {
		projected := projectRetailAdvance(advance)
		runtime.Advances = append(runtime.Advances, projected)
		if projected.EnablesCapitalization {
			runtime.Production = append(runtime.Production, RuntimeProduction{ID: "CAPITALIZATION", Kind: "capitalization", Cost: 1, Prerequisites: []string{projected.ID}})
		}
		if projected.EnablesInfrastructure {
			runtime.Production = append(runtime.Production, RuntimeProduction{ID: "INFRASTRUCTURE", Kind: "infrastructure", Cost: 1, Prerequisites: []string{projected.ID}})
		}
	}
	for _, feat := range db.Feats {
		runtime.Feats = append(runtime.Feats, projectRetailFeat(feat))
	}
	for _, unit := range db.Units {
		projected := projectRetailUnit(unit)
		runtime.Units = append(runtime.Units, projected)
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: unit.ID, Kind: "unit"})
		if unit.ID == "UNIT_SETTLER" {
			runtime.Starter = RuntimeStarter{TypeID: projected.ID, Domain: projected.Domain, Movement: projected.Movement, Attack: projected.Attack, Defense: projected.Defense, Ranged: projected.Ranged, Strength: projected.Strength, Armor: projected.Armor, Firepower: projected.Firepower, Health: projected.Health, MaxFuel: projected.MaxFuel, ZOC: projected.ZOC, TransportCapacity: projected.TransportCapacity, CargoDomains: append([]string(nil), projected.CargoDomains...), CargoClasses: append([]string(nil), projected.CargoClasses...), Size: projected.Size, VisibilityRange: projected.VisibilityRange, Effects: append([]RuntimeStarterEffect(nil), projected.Effects...)}
		}
	}
	buildingEffects := make(map[string][]RuntimeStarterEffect, len(db.Buildings))
	for _, building := range db.Buildings {
		effects := retailBuildingRuntimeEffects(building)
		buildingEffects[building.ID] = effects
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: building.ID, Kind: "building", Cost: building.ProductionCost, Upkeep: building.Upkeep, Prerequisites: singleOrEmpty(building.EnableAdvance), RequiredBuildings: append([]string(nil), building.PrerequisiteBuilding...), CantBuildInSea: retailContainsString(building.Flags, "CantBuildInSea"), CantBuildOnLand: retailContainsString(building.Flags, "CantBuildOnLand"), CoastalOnly: retailContainsString(building.Flags, "CoastalBuilding"), ObsoleteBy: append([]string(nil), building.ObsoleteAdvance...), Effects: effects})
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: building.ID, Kind: "building"})
	}
	for _, wonder := range db.Wonders {
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: wonder.ID, Kind: "wonder", Cost: wonder.ProductionCost, Prerequisites: singleOrEmpty(wonder.EnableAdvance), ObsoleteBy: append([]string(nil), wonder.ObsoleteAdvance...), Effects: retailWonderRuntimeEffects(wonder, buildingEffects)})
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: wonder.ID, Kind: "wonder"})
		if containsRetailFlag(wonder.Flags, "StartGaiaController") {
			runtime.Gaia = RuntimeGaia{ActivationTurns: 10, BuildingID: "IMPROVE_GAIA_COMPUTER", MinBuildingCount: 10, SatelliteBuildingID: "IMPROVE_POWER_SATELLITE", MinSatelliteCount: 10, MaxSatelliteCount: 40, TowerImprovementID: "TILEIMP_PROCESSING_TOWER", MinTowerCount: 5, MinTowerCoveragePercent: 60, TowerMinRadius: 5, TowerMaxRadius: 16, WonderID: wonder.ID}
		}
	}
	for _, terrain := range db.Terrains {
		runtime.Movement.Terrains = append(runtime.Movement.Terrains, RuntimeTerrain{ID: terrain.ID, Cost: positiveOrOne(terrain.EnvBase.Movement), Domains: retailDomains(terrain.MovementTypes), Food: terrain.EnvBase.Food, Production: terrain.EnvBase.Shield, Gold: terrain.EnvBase.Gold, CityFood: terrain.EnvCity.Food, CityProduction: terrain.EnvCity.Shield, CityGold: terrain.EnvCity.Gold, CityCost: terrain.EnvCity.Movement, RiverFood: terrain.EnvRiver.Food, RiverProduction: terrain.EnvRiver.Shield, RiverGold: terrain.EnvRiver.Gold, RiverCost: terrain.EnvRiver.Movement, RiverDefenseBonusPercent: percentPoints(terrain.EnvRiver.Defense), DefenseBonusPercent: percentPoints(terrain.EnvBase.Defense), CanDie: terrain.CanDie, DeadFood: terrain.EnvBase.DeadFood, DeadProduction: terrain.EnvBase.DeadShield, DeadGold: terrain.EnvBase.DeadGold, TransformAddAdvanceID: terrain.AddAdvance, TransformAddTime: terrain.TransformAdd.Time, TransformAddMaterials: terrain.TransformAdd.Materials, TransformRemoveAdvanceID: terrain.RemoveAdvance, TransformRemoveTime: terrain.TransformRemove.Time, TransformRemoveMaterials: terrain.TransformRemove.Materials})
	}
	for _, improvement := range db.TerrainImprovements {
		runtime.TerrainImprovements = append(runtime.TerrainImprovements, projectRetailTerrainImprovement(improvement))
	}
	for _, government := range db.Governments {
		maxSciencePercent := percentPoints(government.MaxScienceRate)
		runtime.Governments = append(runtime.Governments, RuntimeGovernmentDefinition{
			ID:                        government.ID,
			EnableAdvanceID:           government.EnableAdvance,
			ResearchPercent:           signedCoefficientPercentPoints(government.FloatValues["KnowledgeCoef"]),
			InfrastructureCoefficient: government.FloatValues["InfrastructureCoefficient"],
			CapitalizationCoefficient: government.FloatValues["CapitalizationCoefficient"],
			FoodPercent:               signedCoefficientPercentPoints(government.FloatValues["FoodCoef"]),
			ProductionPercent:         signedCoefficientPercentPoints(government.FloatValues["ProductionCoef"]),
			GoldPercent:               signedCoefficientPercentPoints(government.FloatValues["GoldCoef"]),
			PollutionPercent:          signedCoefficientPercentPoints(government.PollutionCoef),
			PollutionUnhappyCoef:      government.FloatValues["PollutionUnhappyCoef"],
			CrimeCoef:                 government.FloatValues["CrimeCoef"],
			CrimeOffset:               government.IntValues["CrimeOffset"],
			WorkdayExpectation:        government.IntValues["WorkdayExpectation"],
			PositiveWorkdayCoef:       government.FloatValues["PositiveWorkdayCoef"],
			NegativeWorkdayCoef:       government.FloatValues["NegativeWorkdayCoef"],
			WagesExpectation:          government.IntValues["WagesExpectation"],
			PositiveWagesCoef:         government.FloatValues["PositiveWagesCoef"],
			NegativeWagesCoef:         government.FloatValues["NegativeWagesCoef"],
			RationsExpectation:        government.IntValues["RationsExpectation"],
			PositiveRationsCoef:       government.FloatValues["PositiveRationsCoef"],
			NegativeRationsCoef:       government.FloatValues["NegativeRationsCoef"],
			SupportCoef:               government.FloatValues["SupportCoef"],
			DefenseCoef:               government.FloatValues["DefenseCoef"],
			ProfessionalUnits:         government.IntValues["ProfessionalUnits"],
			TurnsToNewReadiness:       government.IntValues["TurnsToNewReadiness"],
			ReadyPeaceCoef:            government.FloatValues["ReadyPeaceCoef"],
			ReadyPeaceHP:              government.FloatValues["ReadyPeaceHP"],
			ReadyAlertCoef:            government.FloatValues["ReadyAlertCoef"],
			ReadyAlertHP:              government.FloatValues["ReadyAlertHP"],
			ReadyWarCoef:              government.FloatValues["ReadyWarCoef"],
			ReadyWarHP:                government.FloatValues["ReadyWarHP"],
			WarDiscontentMaxUnits:     government.IntValues["WarDiscontentMaxUnits"],
			WarDiscontentPerUnit:      government.FloatValues["WarDiscontentPerUnit"],
			ConquestDistress:          government.FloatValues["ConquestDistress"],
			ConquestDistressDecay:     government.FloatValues["ConquestDistressDecay"],
			AtHomeRadius:              government.IntValues["AtHomeRadius"],
			OverseasCoef:              government.FloatValues["OverseasCoef"],
			OverseasDefeatDecay:       government.FloatValues["OverseasDefeatDecay"],
			OverseasDefeatCoef:        government.FloatValues["OverseasDefeatCoef"],
			HomeDefeatDecay:           government.FloatValues["HomeDefeatDecay"],
			HomeDefeatCoef:            government.FloatValues["HomeDefeatCoef"],
			EmpireDistanceScale:       government.FloatValues["EmpireDistanceScale"],
			MinEmpireDistance:         government.IntValues["MinEmpireDistance"],
			MaxEmpireDistance:         government.IntValues["MaxEmpireDistance"],
			TooManyCitiesThreshold:    government.IntValues["TooManyCitiesThreshold"],
			TooManyCitiesCoefficient:  government.FloatValues["TooManyCitiesCoefficient"],
			MartialLawMaxUnits:        government.IntValues["MaxMartialLawUnits"],
			MartialLawEffect:          int(government.FloatValues["MartialLawEffect"]),
			MartialLawThreshold:       government.IntValues["MartialLawThreshold"],
			UnitRushModifier:          government.FloatValues["UnitRushModifier"],
			BuildingRushModifier:      government.FloatValues["BuildingRushModifier"],
			WonderRushModifier:        government.FloatValues["WonderRushModifier"],
			EndGameRushModifier:       government.FloatValues["EndGameRushModifier"],
			MaxIncomingTrade:          government.IntValues["MaxIncomingTrade"],
			MaxOutgoingTrade:          government.IntValues["MaxOutgoingTrade"],
			MaxSciencePercent:         &maxSciencePercent,
		})
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
		Strategies:         projectRetailAIStrategies(ai.Strategies),
	}
	if len(projected.UnitBuildLists) == 0 && len(projected.BuildingBuildLists) == 0 && len(projected.WonderBuildLists) == 0 && len(projected.AdvanceLists) == 0 && len(projected.ImprovementLists) == 0 && len(projected.Strategies) == 0 {
		return nil
	}
	return projected
}

func projectRetailAIStrategies(strategies []RetailAIStrategy) []RuntimeAIStrategy {
	projected := make([]RuntimeAIStrategy, 0, len(strategies))
	for _, strategy := range strategies {
		if strategy.ID == "" {
			continue
		}
		projected = append(projected, RuntimeAIStrategy{ID: strategy.ID, SciencePercent: strategy.SciencePercent, PublicWorksPercent: strategy.PublicWorksPercent, PublicWorksReserve: strategy.PublicWorksReserve, SliderElements: projectRetailAISliderElements(strategy.SliderElements)})
	}
	return projected
}

func projectRetailAISliderElements(elements []RetailAISliderElement) []RuntimeAISliderElement {
	projected := make([]RuntimeAISliderElement, 0, len(elements))
	for _, element := range elements {
		if element.Kind == "" {
			continue
		}
		projected = append(projected, RuntimeAISliderElement{Kind: element.Kind, Delta: element.Delta})
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

func retailRuntimeFinalTurn(calendar RetailCalendar) int {
	if calendar.FinalTurn > 0 {
		return calendar.FinalTurn
	}
	return calendarFinalTurn
}

func projectRetailPollutionModel(model RetailPollutionModel) *RuntimePollutionModel {
	if model == (RetailPollutionModel{}) {
		return nil
	}
	return &RuntimePollutionModel{PopulationStart: model.PopulationStart, PopulationRatio: model.PopulationRatio, ProductionStart: model.ProductionStart, ProductionRatio: model.ProductionRatio}
}

func projectRetailAdvance(advance RetailAdvance) RuntimeAdvanceDefinition {
	return RuntimeAdvanceDefinition{
		ID:                          advance.ID,
		Cost:                        advance.Cost,
		AgeID:                       advance.AgeID,
		Prerequisites:               append([]string(nil), advance.Prerequisites...),
		TriggerFeatID:               advance.TriggerFeatID,
		AllowsPlantNuke:             retailContainsString(advance.Flags, "PlantNuke"),
		EnablesCapitalization:       retailContainsString(advance.Flags, "Capitalization"),
		EnablesInfrastructure:       retailContainsString(advance.Flags, "Infrastructure"),
		EnablesTransform:            retailContainsString(advance.Flags, "Transform"),
		EnablesRoads:                retailContainsString(advance.Flags, "Roads"),
		EnablesRailroads:            retailContainsString(advance.Flags, "Railroads"),
		EnablesMagLev:               retailContainsString(advance.Flags, "MagLev"),
		EnablesTunnels:              retailContainsString(advance.Flags, "Tunnels"),
		EnablesDeepOcean:            retailContainsString(advance.Flags, "DeepOcean"),
		PollutionSizeModifier:       signedPercentPointsValue(advance.FloatValues["PollutionSizeModifier"]),
		PollutionProductionModifier: signedPercentPointsValue(advance.FloatValues["PollutionProductionModifier"]),
	}
}

func projectRetailFeat(feat RetailFeat) RuntimeFeatDefinition {
	return RuntimeFeatDefinition{ID: feat.ID, Duration: feat.Duration, BuildingID: feat.BuildingID, BuildingCount: feat.BuildingCount, BuildingPercent: feat.BuildingPercent, MinimumCivSize: feat.IntValues["MinimumSizeOfCiv"], Effects: retailFeatRuntimeEffects(feat)}
}

func retailFeatRuntimeEffects(feat RetailFeat) []RuntimeStarterEffect {
	effects := make([]RuntimeStarterEffect, 0)
	for _, mapping := range []struct{ field, kind string }{
		{"EffectIncreaseProduction", "yield_production_percent"},
		{"EffectIncreaseCommerce", "yield_gold_percent"},
		{"EffectIncreaseScience", "research_percent"},
		{"EffectIncreaseHappiness", "happiness"},
		{"EffectReduceCityWalls", "reduce_city_walls"},
		{"EffectBoatMovement", "increase_boat_movement"},
		{"EffectCityDefenseBonus", "combat_defense"},
		{"EffectIncreaseCityVision", "vision"},
		{"EffectIncreaseBoatVision", "vision"},
		{"EffectIncreaseHitPoints", "increase_hp"},
	} {
		if value, ok := feat.IntValues[mapping.field]; ok {
			effects = append(effects, RuntimeStarterEffect{ID: feat.ID + ":" + mapping.kind, Kind: mapping.kind, Value: value})
		}
	}
	for _, mapping := range []struct{ field, kind string }{
		{"EffectGiveMaps", "give_maps"},
		{"EffectEliminateDistancePenalty", "eliminate_distance_penalty"},
		{"EffectScriptedTurn", "scripted_turn"},
		{"EffectScriptedCity", "scripted_city"},
	} {
		if retailContainsString(feat.Flags, mapping.field) || len(retailFieldValues(feat.RawFields, mapping.field)) > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: feat.ID + ":" + mapping.kind, Kind: mapping.kind})
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
	if value := building.IntValues["HappyInc"]; value != 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":happy-inc", Kind: "happiness", Value: value})
	}
	if value, ok := signedPercentPoints(building.FloatValues["LowerCrime"]); ok {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":lower-crime", Kind: "crime_percent", Value: value})
	}
	if value := building.IntValues["RaiseMaxPopulation"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":max-population", Kind: "max_population", Value: value})
	}
	if value := building.IntValues["RaiseOvercrowdingLevel"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":overcrowding-level", Kind: "overcrowding_level", Value: value})
	}
	if value := building.IntValues["StarvationProtection"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":starvation-protection", Kind: "starvation_protection", Value: value})
	}
	if retailContainsString(building.Flags, "Airport") {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":airport", Kind: "airport"})
	}
	if value := percentPoints(building.FloatValues["FoodVat"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":food-vat", Kind: "food_vat_percent", Value: value})
	}
	for _, marker := range []struct{ flag, kind string }{
		{"AllowGrunts", "allow_grunts"},
		{"Brokerage", "brokerage"},
		{"Capitol", "capitol"},
		{"Cathedral", "cathedral"},
		{"DoubleTelevangelists", "double_televangelists"},
		{"IsReligious", "is_religious"},
		{"NuclearPlant", "nuclear_plant"},
		{"Television", "television"},
	} {
		if retailContainsString(building.Flags, marker.flag) {
			effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":" + marker.kind, Kind: marker.kind})
		}
	}
	if value, ok := signedPercentPoints(building.FloatValues["LowerPeaceMovement"]); ok {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":lower-peace-movement", Kind: "lower_peace_movement_percent", Value: value})
	}
	if retailContainsString(building.Flags, "ProtectFromNukes") {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":protect-from-nukes", Kind: "protect_from_nukes"})
	}
	if value := percentPoints(building.FloatValues["ProtectFromBioAgents"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":protect-from-bio-agents", Kind: "protect_from_bio_agents_percent", Value: value})
	}
	if value := percentPoints(building.FloatValues["ProtectFromNanoVirus"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":protect-from-nano-virus", Kind: "protect_from_nano_virus_percent", Value: value})
	}
	if value := percentPoints(building.FloatValues["PreventConversion"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":prevent-conversion", Kind: "prevent_conversion_percent", Value: value})
	}
	if value := percentPoints(building.FloatValues["PreventSlavery"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":prevent-slavery", Kind: "prevent_slavery_percent", Value: value})
	}
	if retailContainsString(building.Flags, "ForceField") {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":force-field", Kind: "force_field"})
	}
	if value := retailPercentFieldPoints(building.FloatValues["DefendersPercent"]); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":defenders-percent", Kind: "combat_defense", Value: value})
		if retailContainsString(building.Flags, "CityWalls") {
			effects = append(effects, RuntimeStarterEffect{ID: building.ID + ":city-walls", Kind: "city_walls", Value: value})
		}
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
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":bonus-gold", Kind: "bonus_gold", Value: value})
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
	if value := wonder.IntValues["IncHappinessEmpire"]; value != 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":happiness", Kind: "happiness", Value: value})
	}
	if value := wonder.IntValues["DecCrimePercent"]; value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":dec-crime-percent", Kind: "crime_percent", Value: -value})
	}
	for _, numeric := range []struct {
		field string
		kind  string
	}{
		{"IncreaseBoatMovement", "increase_boat_movement"},
		{"IncreaseHp", "increase_hp"},
		{"IncreaseRegard", "increase_regard"},
		{"IncreaseBrokerages", "increase_brokerages"},
		{"IncreaseScientists", "increase_scientists"},
		{"ReduceReadinessCost", "reduce_readiness_cost"},
		{"TemporaryFullHappiness", "temporary_full_happiness"},
		{"GoldPerWaterTradeRoute", "gold_per_water_trade_route"},
		{"GoldPerInternationalTradeRoute", "gold_per_international_trade_route"},
		{"GoldPerTelevision", "gold_per_television"},
		{"DecEmpireSize", "decrease_empire_size"},
	} {
		if value := wonder.IntValues[numeric.field]; value > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":" + numeric.kind, Kind: numeric.kind, Value: value})
		}
	}
	for _, marker := range []struct {
		field string
		kind  string
	}{
		{"CloseEmbassies", "close_embassies"},
		{"EliminateNukes", "eliminate_nukes"},
		{"EmbassiesEverywhereEvenAtWar", "embassies_everywhere"},
		{"EnableParkRangers", "enable_park_rangers"},
		{"FreeSlaves", "free_slaves"},
		{"GlobalRadar", "global_radar"},
		{"NoPollutionUnhappiness", "no_pollution_unhappiness"},
		{"PollutersToParks", "polluters_to_parks"},
		{"PreventConversion", "prevent_conversion_empire"},
		{"ProhibitSlavers", "prohibit_slavers"},
		{"ProtectFromBarbarians", "protect_from_barbarians"},
		{"ReformCities", "reform_cities"},
		{"RevoltingCitiesJoinPlayer", "revolting_cities_join_player"},
		{"SpiesEverywhere", "spies_everywhere"},
		{"StartGaiaController", "start_gaia_controller"},
	} {
		if retailContainsString(wonder.Flags, marker.field) || len(retailFieldValues(wonder.RawFields, marker.field)) > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: wonder.ID + ":" + marker.kind, Kind: marker.kind})
		}
	}
	for _, buildingID := range append(retailFieldValues(wonder.RawFields, "BuildingEverywhere"), retailFieldValues(wonder.RawFields, "ActualBuildingEverywhere")...) {
		effects = append(effects, RuntimeStarterEffect{ID: buildingID, Kind: "building_everywhere"})
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

func signedPercentPointsValue(value float64) int {
	result, _ := signedPercentPoints(value)
	return result
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
	cargoClasses := retailCargoClasses(unit.CanCarry)
	requiredGovernmentIDs := retailFieldValues(unit.RawFields, "GovernmentType")
	requiredGovernmentID := ""
	if len(requiredGovernmentIDs) > 0 {
		requiredGovernmentID = requiredGovernmentIDs[0]
	}
	needsCityGoodID := firstRetailUnitString(unit, "NeedsCityGood")
	settleCityTypeID := firstRetailUnitString(unit, "SettleCityType")
	cityGrowthCoefficientPercent := percentPoints(firstRetailUnitFloat(unit, "CityGrowthCoefficient"))
	return RuntimeUnitDefinition{ID: unit.ID, Role: unit.Category, Domain: retailUnitDomain(unit.MovementTypes), Movement: positiveOrOne(unit.MaxMovePoints), Attack: unit.Attack, Defense: unit.Defense, Ranged: unit.ZBRangeAttack, BombardRange: unit.BombardRange, BombRounds: unit.BombRounds, ActiveDefenseRange: unit.ActiveDefenseRange, Strength: unit.Attack, Armor: unit.Armor, Firepower: positiveOrOne(unit.Firepower), Health: positiveOrOne(unit.MaxHP), MaxFuel: unit.MaxFuel, ZOC: !retailContainsString(unit.Flags, "NoZoc"), TransportCapacity: unit.CargoMax, CargoDomains: cargoDomains, CargoClasses: cargoClasses, Size: unit.Size, ProductionCost: unit.ShieldCost, PowerPoints: unit.PowerPoints, Upkeep: unit.GoldHunger, ShieldHunger: unit.ShieldHunger, FoodHunger: unit.FoodHunger, NeedsNoSupport: retailContainsString(unit.Flags, "NeedsNoSupport"), IsSpecialForces: retailContainsString(unit.Flags, "IsSpecialForces"), Prerequisites: singleOrEmpty(unit.EnableAdvance), RequiredGovernmentID: requiredGovernmentID, RequiredGovernmentIDs: append([]string(nil), requiredGovernmentIDs...), NeedsCityGoodID: needsCityGoodID, SettleCityTypeID: settleCityTypeID, CityGrowthCoefficientPercent: cityGrowthCoefficientPercent, ObsoleteBy: append([]string(nil), unit.ObsoleteAdvance...), UpgradeTo: append([]string(nil), unit.UpgradeTo...), VisibilityRange: unit.VisionRange, Effects: effects}
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
	if retailContainsString(unit.Flags, "IsMounted") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("mounted"), Kind: "mounted"})
	}
	if retailContainsString(unit.Flags, "NoFuelThenCrash") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("no_fuel_then_crash"), Kind: "no_fuel_then_crash"})
	}
	if retailContainsString(unit.Flags, "LossMoveToDmgNone") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("loss_move_to_dmg_none"), Kind: "loss_move_to_dmg_none"})
	}
	if retailContainsString(unit.Flags, "LossMoveToDmgTwo") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("loss_move_to_dmg_two"), Kind: "loss_move_to_dmg_two"})
	}
	if retailContainsString(unit.Flags, "CantBuild") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("cant_build"), Kind: "cant_build"})
	}
	if retailContainsString(unit.Flags, "CanBeGifted") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_be_gifted"), Kind: "can_be_gifted"})
	}
	if retailContainsString(unit.Flags, "Advertise") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("advertise"), Kind: "advertise"})
	}
	if retailContainsString(unit.Flags, "GLHidden") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("gl_hidden"), Kind: "gl_hidden"})
	}
	if retailContainsString(unit.Flags, "HasPopAndCanBuild") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("has_pop_and_can_build"), Kind: "has_pop_and_can_build"})
	}
	if retailContainsString(unit.Flags, "ImmuneToHostileTerrain") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("immune_to_hostile_terrain"), Kind: "immune_to_hostile_terrain"})
	}
	if retailContainsString(unit.Flags, "NoIndex") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("no_index"), Kind: "no_index"})
	}
	if retailContainsString(unit.Flags, "OnlyBuildOne") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("only_build_one"), Kind: "only_build_one"})
	}
	if retailContainsString(unit.Flags, "ThrowParty") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("throw_party"), Kind: "throw_party"})
	}
	if retailContainsString(unit.Flags, "VictoryEnslavement") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("victory_enslavement"), Kind: "victory_enslavement"})
	}
	if retailContainsString(unit.Flags, "CantCaptureCity") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("cant_capture_city"), Kind: "cant_capture_city"})
	}
	if retailContainsString(unit.Flags, "CanBeSued") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_be_sued"), Kind: "can_be_sued"})
	}
	if retailContainsString(unit.Flags, "CanBeExpelled") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_be_expelled"), Kind: "can_be_expelled"})
	}
	if retailContainsString(unit.Flags, "CanEntrench") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_entrench"), Kind: "can_entrench"})
	}
	if retailContainsString(unit.Flags, "CanPillage") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_pillage"), Kind: "can_pillage"})
	}
	if retailContainsString(unit.Flags, "ExertsMartialLaw") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("exerts_martial_law"), Kind: "exerts_martial_law"})
	}
	if retailContainsString(unit.Flags, "BuildingRemovesAPop") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("building_removes_population"), Kind: "building_removes_population"})
	}
	if retailContainsString(unit.Flags, "IgnoreZOC") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("ignore_zoc"), Kind: "ignore_zoc"})
	}
	if retailContainsString(unit.Flags, "BonusAgainstMounted") {
		if value := percentPoints(firstRetailUnitFloat(unit, "MountedBonus")); value > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("combat_defense_vs_mounted"), Kind: "combat_defense_vs_mounted_percent", Value: value})
		}
	}
	if retailContainsString(unit.Flags, "BonusAirDefense") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("combat_defense_vs_air"), Kind: "combat_defense_vs_air_percent", Value: 100})
	}
	if retailContainsString(unit.Flags, "WoodenShip") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("wooden_ship"), Kind: "wooden_ship"})
	}
	if retailContainsString(unit.Flags, "IsSubmarine") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("submarine"), Kind: "submarine"})
	}
	if retailUnitFieldHasValue(unit.RawFields, "VisionClass", "Stealth") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("stealth"), Kind: "stealth"})
	}
	if retailUnitFieldHasValue(unit.RawFields, "CanSee", "Stealth") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_see_stealth"), Kind: "can_see_stealth"})
	}
	if retailContainsString(unit.Flags, "InvisibleForGlobeSat") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("invisible_for_globe_sat"), Kind: "invisible_for_globe_sat"})
	}
	for _, target := range retailUnitCanAttackEffects(unit.RawFields) {
		effects = append(effects, RuntimeStarterEffect{ID: effectID(target), Kind: target})
	}
	for _, target := range retailUnitCanBombardEffects(unit.RawFields) {
		effects = append(effects, RuntimeStarterEffect{ID: effectID(target), Kind: target})
	}
	for _, target := range retailUnitDefendEffects(unit.RawFields) {
		effects = append(effects, RuntimeStarterEffect{ID: effectID(target), Kind: target})
	}
	if retailContainsString(unit.Flags, "CanCounterBombard") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_counter_bombard"), Kind: "can_counter_bombard"})
	}
	if retailContainsString(unit.Flags, "IsFlanker") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("is_flanker"), Kind: "is_flanker"})
	}
	if retailContainsString(unit.Flags, "CanBeachAssault") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_beach_assault"), Kind: "can_beach_assault"})
	}
	if retailContainsString(unit.Flags, "SingleUse") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("single_use"), Kind: "single_use"})
	}
	if retailContainsString(unit.Flags, "DeathEffectsHappy") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("death_effects_happy"), Kind: "death_effects_happy"})
	}
	if retailContainsString(unit.Flags, "Explodes") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("explodes"), Kind: "explodes"})
	}
	if len(unit.FieldsByKey("NuclearAttack")) > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("nuclear_attack"), Kind: "nuclear_attack"})
	}
	if fields := unit.FieldsByKey("CauseUnhappiness"); len(fields) > 0 {
		if chance := percentPoints(retailRawFieldFloat(fields[0], "Chance")); chance > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_cause_unhappiness"), Kind: "special_cause_unhappiness", Value: chance})
		}
		if timer := int(retailRawFieldFloat(fields[0], "Timer")); timer > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_cause_unhappiness_timer"), Kind: "special_cause_unhappiness_timer", Value: timer})
		}
		if amount := int(retailRawFieldFloat(fields[0], "Amount")); amount > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_cause_unhappiness_amount"), Kind: "special_cause_unhappiness_amount", Value: amount})
		}
	}
	if len(unit.FieldsByKey("IndulgenceSales")) > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_indulgence_sales"), Kind: "special_indulgence_sales"})
	}
	if fields := unit.FieldsByKey("InciteRevolution"); len(fields) > 0 {
		if chance := percentPoints(retailRawFieldFloat(fields[0], "Chance")); chance > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_incite_revolution"), Kind: "special_incite_revolution", Value: chance})
		}
		if chance := percentPoints(retailRawFieldFloat(fields[0], "EliteChance")); chance > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_incite_revolution_elite_chance"), Kind: "special_incite_revolution_elite_chance", Value: chance})
		}
		if chance := percentPoints(retailRawFieldFloat(fields[0], "DeathChance")); chance > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_incite_revolution_death_chance"), Kind: "special_incite_revolution_death_chance", Value: chance})
		}
	}
	if len(unit.FieldsByKey("SlaveUprising")) > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_slave_uprising"), Kind: "special_slave_uprising"})
	}
	if retailContainsString(unit.Flags, "CanPirate") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("can_pirate"), Kind: "can_pirate"})
	}
	if retailContainsString(unit.Flags, "Civilian") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("civilian"), Kind: "civilian"})
	}
	if retailContainsString(unit.Flags, "IsTrader") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("is_trader"), Kind: "is_trader"})
	}
	if retailContainsString(unit.Flags, "NoSlaves") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("no_slaves"), Kind: "no_slaves"})
	}
	if retailContainsString(unit.Flags, "IsTelevangelist") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("is_televangelist"), Kind: "is_televangelist"})
	}
	if retailContainsString(unit.Flags, "Paratrooper") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("paratrooper"), Kind: "paratrooper"})
	}
	if retailContainsString(unit.Flags, "ParatrooperTransport") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("paratrooper_transport"), Kind: "paratrooper_transport"})
	}
	if turns := firstRetailUnitInt(unit, "SpaceLaunch"); turns > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("space_launch_turns"), Kind: "space_launch_turns", Value: turns})
	}
	if retailContainsString(unit.Flags, "ActiveDefenseOnlyWhenCarryingEnablers") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("active_defense_requires_enabler"), Kind: "active_defense_requires_enabler"})
	}
	if retailContainsString(unit.Flags, "EnableCarrierDefenses") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("enable_carrier_defenses"), Kind: "enable_carrier_defenses"})
	}
	if value := percentPoints(firstRetailUnitFloat(unit, "DefendAgainstSpies")); value > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("defend_against_spies"), Kind: "defend_against_spies", Value: value})
	}
	if retailContainsString(unit.Flags, "AttackWoodenShipBonus") {
		if value := percentPoints(firstRetailUnitFloat(unit, "WoodenShipBonus")); value > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("combat_attack_vs_wooden_ship"), Kind: "combat_attack_vs_wooden_ship_percent", Value: value})
		}
	}
	if retailContainsString(unit.Flags, "IgnoreCityWalls") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("ignore_city_walls"), Kind: "ignore_city_walls"})
	}
	if retailContainsString(unit.Flags, "EstablishEmbassy") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_embassy"), Kind: "special_embassy"})
	}
	if len(unit.FieldsByKey("CanSoothsay")) > 0 {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_soothsaying"), Kind: "special_soothsaying"})
	}
	for _, special := range []struct {
		field string
		kind  string
	}{
		{"ConvertCities", "special_convert_city"},
		{"CreateFranchise", "special_create_franchise"},
		{"InvestigateCity", "special_investigate_city"},
		{"PlantNuke", "special_plant_nuke"},
		{"BioTerror", "special_bio_terror"},
		{"NanoTerror", "special_nano_terror"},
		{"Plague", "special_plague"},
		{"AssasinateRuler", "special_assassinate_ruler"},
		{"SlaveRaids", "special_slave_raid"},
		{"SettlerSlaveRaids", "special_settler_slave_raid"},
		{"UndergroundRailway", "special_underground_railway"},
		{"CanInjoin", "special_injoin"},
		{"CanReform", "special_reform"},
		{"CreateParks", "special_create_park"},
	} {
		if fields := unit.FieldsByKey(special.field); len(fields) > 0 {
			chance := percentPoints(retailRawFieldFloat(fields[0], "Chance"))
			if chance <= 0 {
				chance = 100
			}
			effects = append(effects, RuntimeStarterEffect{ID: effectID(special.kind), Kind: special.kind, Value: chance})
			if eliteChance := percentPoints(retailRawFieldFloat(fields[0], "EliteChance")); eliteChance > 0 {
				eliteKind := special.kind + "_elite_chance"
				effects = append(effects, RuntimeStarterEffect{ID: effectID(eliteKind), Kind: eliteKind, Value: eliteChance})
			}
			if special.kind == "special_plant_nuke" {
				if escapeChance := percentPoints(retailRawFieldFloat(fields[0], "EscapeChance")); escapeChance > 0 {
					escapeKind := special.kind + "_escape_chance"
					effects = append(effects, RuntimeStarterEffect{ID: effectID(escapeKind), Kind: escapeKind, Value: escapeChance})
				}
			}
			if deathChance := percentPoints(retailRawFieldFloat(fields[0], "DeathChance")); deathChance > 0 {
				deathKind := special.kind + "_death_chance"
				effects = append(effects, RuntimeStarterEffect{ID: effectID(deathKind), Kind: deathKind, Value: deathChance})
			}
		}
	}
	if retailContainsString(unit.Flags, "CanExpel") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_expel"), Kind: "special_expel"})
	}
	if retailContainsString(unit.Flags, "CanSue") {
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_sue"), Kind: "special_sue"})
	}
	if fields := unit.FieldsByKey("StealTechnology"); len(fields) > 0 {
		chance := percentPoints(retailRawFieldFloat(fields[0], "RandomChance"))
		if chance <= 0 {
			chance = 100
		}
		effects = append(effects, RuntimeStarterEffect{ID: effectID("special_steal"), Kind: "special_steal", Value: chance})
		if specificChance := percentPoints(retailRawFieldFloat(fields[0], "SpecificChance")); specificChance > 0 {
			effects = append(effects, RuntimeStarterEffect{ID: effectID("special_steal_specific_chance"), Kind: "special_steal_specific_chance", Value: specificChance})
		}
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

func retailUnitFieldHasValue(fields []RetailRawField, key, value string) bool {
	for _, field := range fields {
		if field.Key != key {
			continue
		}
		for _, candidate := range field.Values {
			if candidate == value {
				return true
			}
		}
	}
	return false
}

func retailUnitCanBombardEffects(fields []RetailRawField) []string {
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, value := range retailFieldValues(fields, "CanBombard") {
		kind := ""
		switch value {
		case "Land", "Mountain":
			kind = "can_bombard_land"
		case "Sea", "Water", "ShallowWater":
			kind = "can_bombard_sea"
		case "Air":
			kind = "can_bombard_air"
		case "Space":
			kind = "can_bombard_space"
		case "Underwater":
			kind = "can_bombard_underwater"
		}
		if kind != "" && !seen[kind] {
			seen[kind] = true
			result = append(result, kind)
		}
	}
	return result
}

func retailUnitDefendEffects(fields []RetailRawField) []string {
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, value := range retailFieldValues(fields, "Defend") {
		kind := ""
		switch value {
		case "Land", "Mountain":
			kind = "defend_land"
		case "Sea", "Water", "ShallowWater":
			kind = "defend_sea"
		case "Air":
			kind = "defend_air"
		case "Space":
			kind = "defend_space"
		case "Underwater":
			kind = "defend_underwater"
		}
		if kind != "" && !seen[kind] {
			seen[kind] = true
			result = append(result, kind)
		}
	}
	return result
}

func retailUnitCanAttackEffects(fields []RetailRawField) []string {
	effects := make([]string, 0)
	add := func(kind string) {
		if kind != "" && !retailContainsString(effects, kind) {
			effects = append(effects, kind)
		}
	}
	for _, value := range retailFieldValues(fields, "CanAttack") {
		switch value {
		case "Land", "Mountain":
			add("can_attack_land")
		case "Sea", "ShallowWater":
			add("can_attack_sea")
		case "Air":
			add("can_attack_air")
		case "Space":
			add("can_attack_space")
		case "Underwater":
			add("can_attack_underwater")
		}
	}
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

func firstRetailUnitString(unit RetailUnit, key string) string {
	values := retailFieldValues(unit.RawFields, key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func firstRetailUnitFloat(unit RetailUnit, key string) float64 {
	values := retailFieldValues(unit.RawFields, key)
	if len(values) == 0 {
		return 0
	}
	value, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return 0
	}
	return value
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
	definition := RuntimeTerrainImprovementDefinition{ID: improvement.ID, Class: improvement.Class, Level: improvement.Level, ConstructionTiles: append([]int(nil), improvement.ConstructionTiles...), BuildTurns: 1, PublicWorksCost: 1, TargetTerrain: improvement.TerraformTerrain, CanSee: append([]string(nil), improvement.CanSee...), IntBorderRadius: improvement.IntBorderRadius, SquaredBorderRadius: improvement.SquaredBorderRadius, CantBuildOn: append([]string(nil), improvement.CantBuildOn...), Excludes: append([]string(nil), improvement.Excludes...)}
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
	definition.RequiredAdvanceID = effect.EnableAdvance
	definition.AllowedTerrains = append([]string(nil), effect.Terrains...)
	definition.DefenseBonusPercent = percentPoints(effect.DefenseBonus)
	definition.HealRatePercent = percentPoints(effect.HealRate)
	definition.VisionRange = effect.VisionRange
	definition.RadarRange = effect.RadarRange
	definition.Freight = effect.Freight
	definition.Airport = retailContainsString(effect.Flags, "Airport")
	definition.Radar = retailContainsString(effect.Flags, "Radar")
	definition.ListeningPost = retailContainsString(effect.Flags, "ListeningPost")
	definition.CanUpgrade = retailContainsString(effect.Flags, "CanUpgrade")
	definition.Endgame = retailContainsString(effect.Flags, "Endgame")
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

func retailCargoClasses(canCarry []string) []string {
	classes := make([]string, 0, len(canCarry))
	for _, carry := range canCarry {
		switch carry {
		case "SmallLand", "MedLand", "LargeLand", "SmallAir", "MedAir", "LargeAir", "SmallWater", "MedWater", "LargeWater", "SmallSpace", "MedSpace", "LargeSpace":
			if !retailContainsString(classes, carry) {
				classes = append(classes, carry)
			}
		}
	}
	sort.Strings(classes)
	return classes
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
	sort.Slice(runtime.Ages, func(i, j int) bool {
		if runtime.Ages[i].Index != runtime.Ages[j].Index {
			return runtime.Ages[i].Index < runtime.Ages[j].Index
		}
		return runtime.Ages[i].ID < runtime.Ages[j].ID
	})
	sort.Slice(runtime.Advances, func(i, j int) bool { return runtime.Advances[i].ID < runtime.Advances[j].ID })
	sort.Slice(runtime.Feats, func(i, j int) bool { return runtime.Feats[i].ID < runtime.Feats[j].ID })
	sort.Slice(runtime.Units, func(i, j int) bool { return runtime.Units[i].ID < runtime.Units[j].ID })
	sort.Slice(runtime.Production, func(i, j int) bool { return runtime.Production[i].ID < runtime.Production[j].ID })
	sort.Slice(runtime.Effects, func(i, j int) bool { return runtime.Effects[i].ID < runtime.Effects[j].ID })
	sort.Slice(runtime.Movement.Terrains, func(i, j int) bool { return runtime.Movement.Terrains[i].ID < runtime.Movement.Terrains[j].ID })
	sort.Slice(runtime.TerrainImprovements, func(i, j int) bool { return runtime.TerrainImprovements[i].ID < runtime.TerrainImprovements[j].ID })
	sort.Slice(runtime.Governments, func(i, j int) bool { return runtime.Governments[i].ID < runtime.Governments[j].ID })
	sort.Slice(runtime.PollutionTriggers, func(i, j int) bool { return runtime.PollutionTriggers[i].ID < runtime.PollutionTriggers[j].ID })
}
