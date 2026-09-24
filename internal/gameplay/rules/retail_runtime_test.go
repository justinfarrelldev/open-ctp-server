package rules

import "testing"

func TestProjectRetailRuntimeIncludesCompleteReleasedCatalog(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	runtime := ProjectRetailRuntime(db)
	if got, want := len(runtime.Units), 74; got != want {
		t.Fatalf("runtime units = %d, want %d", got, want)
	}
	if got, want := len(runtime.Advances), 107; got != want {
		t.Fatalf("runtime advances = %d, want %d", got, want)
	}
	if got, want := len(runtime.Production), 156; got != want {
		t.Fatalf("runtime production = %d, want %d", got, want)
	}
	if got, want := len(runtime.Movement.Terrains), 26; got != want {
		t.Fatalf("runtime terrains = %d, want %d", got, want)
	}
	if got, want := len(runtime.TerrainImprovements), 40; got != want {
		t.Fatalf("runtime terrain improvements = %d, want %d", got, want)
	}
	if got, want := len(runtime.PollutionTriggers), 4; got != want {
		t.Fatalf("runtime pollution triggers = %d, want %d", got, want)
	}
	if runtime.PollutionModel == nil || runtime.PollutionModel.PopulationStart != 16 || runtime.PollutionModel.PopulationRatio != 15 || runtime.PollutionModel.ProductionStart != 300 || runtime.PollutionModel.ProductionRatio != 1 {
		t.Fatalf("runtime pollution model = %+v", runtime.PollutionModel)
	}
	if runtime.Gaia.ActivationTurns != 10 || runtime.Gaia.BuildingID != "IMPROVE_GAIA_COMPUTER" || runtime.Gaia.WonderID != "WONDER_THE_SOLARIS_PROJECT" {
		t.Fatalf("runtime Gaia = %+v", runtime.Gaia)
	}
	if runtime.AI == nil || len(runtime.AI.UnitBuildLists) != 27 || len(runtime.AI.BuildingBuildLists) != 13 || len(runtime.AI.WonderBuildLists) != 7 || len(runtime.AI.AdvanceLists) != 7 || len(runtime.AI.ImprovementLists) != 4 {
		t.Fatalf("runtime AI lists = %+v", runtime.AI)
	}
}

func TestProjectRetailRuntimeMapsSelectedRetailValues(t *testing.T) {
	root := localRetailGameDataRoot(t)
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatalf("LoadRetailRulesFromDirectory() error = %v", err)
	}
	runtime := ProjectRetailRuntime(db)

	if runtime.AI == nil || runtime.AI.UnitBuildLists[0].ID != "UNIT_BUILD_LIST_SPECIAL" || !sameStringSlices(runtime.AI.UnitBuildLists[0].Items[:3], []string{"UNIT_DIPLOMAT", "UNIT_SLAVER", "UNIT_ABOLITIONIST"}) {
		t.Fatalf("runtime AI unit list = %+v", runtime.AI)
	}
	if runtime.AI.BuildingBuildLists[1].ID != "BUILDING_BUILD_LIST_GROWTH" || !containsString(runtime.AI.BuildingBuildLists[1].Items, "IMPROVE_GRANARY") {
		t.Fatalf("runtime AI building lists = %+v", runtime.AI.BuildingBuildLists)
	}

	nuclearPower, ok := runtime.Advance("ADVANCE_NUCLEAR_POWER")
	if !ok || !nuclearPower.AllowsPlantNuke {
		t.Fatalf("nuclear power advance = %+v ok=%v", nuclearPower, ok)
	}

	settler, ok := runtime.Unit("UNIT_SETTLER")
	if !ok {
		t.Fatalf("UNIT_SETTLER missing")
	}
	if settler.Movement != 100 || settler.Attack != 0 || settler.Defense != 10 || settler.Health != 10 || settler.ProductionCost != 740 {
		t.Fatalf("settler runtime = %+v", settler)
	}
	if !runtimeUnitHasEffect(settler, "settler") || !runtimeUnitHasEffectValue(settler, "settle_size", 1) {
		t.Fatalf("settler effects missing: %+v", settler.Effects)
	}
	seaEngineer, ok := runtime.Unit("UNIT_SEA_ENGINEER")
	if !ok || !runtimeUnitHasEffectID(seaEngineer, "settle_building", "IMPROVE_GRANARY") {
		t.Fatalf("sea engineer runtime = %+v ok=%v", seaEngineer, ok)
	}
	cleric, ok := runtime.Unit("UNIT_CLERIC")
	if !ok || !runtimeUnitHasEffect(cleric, "special_soothsaying") || !runtimeUnitHasEffectValue(cleric, "special_convert_city", 75) {
		t.Fatalf("cleric runtime = %+v ok=%v", cleric, ok)
	}
	corporateBranch, ok := runtime.Unit("UNIT_CORPORATE_BRANCH")
	if !ok || !runtimeUnitHasEffectValue(corporateBranch, "special_create_franchise", 75) || !runtimeUnitHasEffect(corporateBranch, "cant_capture_city") || !runtimeUnitHasEffect(corporateBranch, "can_be_sued") || !runtimeUnitHasEffect(corporateBranch, "can_be_expelled") {
		t.Fatalf("corporate branch runtime = %+v ok=%v", corporateBranch, ok)
	}
	cyberNinja, ok := runtime.Unit("UNIT_CYBER_NINJA")
	if !ok || !runtimeUnitHasEffectValue(cyberNinja, "special_steal", 50) || !runtimeUnitHasEffectValue(cyberNinja, "special_capture", 25) || !runtimeUnitHasEffectValue(cyberNinja, "special_investigate_city_death_chance", 25) || !runtimeUnitHasEffectValue(cyberNinja, "special_investigate_city_elite_chance", 20) || !runtimeUnitHasEffectValue(cyberNinja, "special_plant_nuke_escape_chance", 50) || !runtimeUnitHasEffect(cyberNinja, "stealth") || !runtimeUnitHasEffect(cyberNinja, "can_see_stealth") {
		t.Fatalf("cyber ninja runtime = %+v ok=%v", cyberNinja, ok)
	}
	abolitionist, ok := runtime.Unit("UNIT_ABOLITIONIST")
	if !ok || !runtimeUnitHasEffectValue(abolitionist, "special_underground_railway", 75) || !runtimeUnitHasEffectValue(abolitionist, "special_underground_railway_death_chance", 50) {
		t.Fatalf("abolitionist runtime = %+v ok=%v", abolitionist, ok)
	}
	if runtime.Starter.Movement != 100 || runtime.Starter.Health != 10 || !runtimeStarterHasEffect(runtime.Starter, "settler") || !runtimeStarterHasEffect(runtime.Starter, "building_removes_population") {
		t.Fatalf("starter = %+v, want retail settler", runtime.Starter)
	}

	cavalry, ok := runtime.Unit("UNIT_CAVALRY")
	if !ok || !runtimeUnitHasEffect(cavalry, "mounted") {
		t.Fatalf("cavalry runtime = %+v ok=%v", cavalry, ok)
	}
	archer, ok := runtime.Unit("UNIT_ARCHER")
	if !ok || !runtimeUnitHasEffect(archer, "can_entrench") || !runtimeUnitHasEffect(archer, "can_pillage") {
		t.Fatalf("archer runtime = %+v ok=%v", archer, ok)
	}
	bomber, ok := runtime.Unit("UNIT_BOMBER")
	if !ok || !runtimeUnitHasEffect(bomber, "no_fuel_then_crash") || !runtimeUnitHasEffect(bomber, "ignore_zoc") || !runtimeUnitHasEffect(bomber, "can_attack_air") || !runtimeUnitHasEffect(bomber, "can_attack_land") || !runtimeUnitHasEffect(bomber, "can_attack_sea") {
		t.Fatalf("bomber runtime = %+v ok=%v", bomber, ok)
	}
	pikemen, ok := runtime.Unit("UNIT_PIKEMEN")
	if !ok || !runtimeUnitHasEffectValue(pikemen, "combat_defense_vs_mounted_percent", 50) {
		t.Fatalf("pikemen runtime = %+v ok=%v", pikemen, ok)
	}
	fireTrireme, ok := runtime.Unit("UNIT_FIRE_TRIREME")
	if !ok || !runtimeUnitHasEffect(fireTrireme, "wooden_ship") || !runtimeUnitHasEffectValue(fireTrireme, "combat_attack_vs_wooden_ship_percent", 25) {
		t.Fatalf("fire trireme runtime = %+v ok=%v", fireTrireme, ok)
	}

	carrier, ok := runtime.Unit("UNIT_AIRCRAFT_CARRIER")
	if !ok {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER missing")
	}
	if carrier.Attack != 25 || carrier.Defense != 30 || carrier.Ranged != 25 || carrier.BombardRange != 1 || carrier.BombRounds != 10 || carrier.Movement != 400 || carrier.TransportCapacity != 5 || carrier.ProductionCost != 3000 || carrier.Upkeep != 150 {
		t.Fatalf("carrier runtime = %+v", carrier)
	}
	if got, want := carrier.CargoDomains, []string{"air"}; !sameStringSlices(got, want) {
		t.Fatalf("carrier cargo domains = %#v, want %#v", got, want)
	}
	fighter, ok := runtime.Unit("UNIT_FIGHTER")
	if !ok || fighter.Domain != "air" || fighter.MaxFuel != 2400 {
		t.Fatalf("fighter runtime = %+v ok=%v", fighter, ok)
	}

	railroad, ok := runtime.Advance("ADVANCE_RAILROAD")
	if !ok || railroad.Cost != 10912 || !sameStringSlices(railroad.Prerequisites, []string{"ADVANCE_BANKING", "ADVANCE_MODERN_METALLURGY"}) {
		t.Fatalf("railroad runtime = %+v ok=%v", railroad, ok)
	}

	plains, ok := runtime.Movement.Terrain("TERRAIN_PLAINS")
	if !ok || plains.Cost != 100 || plains.Food != 10 || plains.Production != 10 || plains.Gold != 5 || plains.CityFood != 10 || plains.CityProduction != 15 || plains.CityGold != 15 || plains.TransformAddTime != 1 || plains.TransformAddMaterials != 200 || plains.TransformRemoveTime != 1 || plains.TransformRemoveMaterials != 50 || !sameStringSlices(plains.Domains, []string{"air", "land"}) {
		t.Fatalf("plains runtime terrain = %+v ok=%v", plains, ok)
	}

	farms, ok := runtime.TerrainImprovement("TILEIMP_FARMS")
	if !ok || farms.BuildTurns != 2 || farms.PublicWorksCost != 200 || farms.YieldFood != 10 || farms.Class != "Farm" || !retailContainsString(farms.CantBuildOn, "TERRAIN_WATER_SHELF") || !retailContainsString(farms.Excludes, "Mine") {
		t.Fatalf("farms runtime = %+v ok=%v", farms, ok)
	}
	terraformGrassland, ok := runtime.TerrainImprovement("TILEIMP_TERRAFORM_GRASSLAND")
	if !ok || terraformGrassland.TargetTerrain != "TERRAIN_GRASSLAND" || terraformGrassland.BuildTurns != 1 || terraformGrassland.PublicWorksCost != 1 {
		t.Fatalf("terraform grassland runtime = %+v ok=%v", terraformGrassland, ok)
	}

	granary, ok := runtime.ProductionDefinition("IMPROVE_GRANARY")
	if !ok || granary.Kind != "building" || granary.Cost != 300 || granary.Upkeep != 1 || !sameStringSlices(granary.Prerequisites, []string{"ADVANCE_AGRICULTURE"}) {
		t.Fatalf("granary production = %+v ok=%v", granary, ok)
	}
	democracy, ok := runtime.Government("GOVERNMENT_DEMOCRACY")
	if !ok || democracy.EnableAdvanceID != "ADVANCE_DEMOCRACY" || democracy.ResearchPercent != 100 || democracy.FoodPercent != 25 || democracy.ProductionPercent != 15 || democracy.GoldPercent != 25 || democracy.PollutionPercent != -20 || democracy.MaxSciencePercent == nil || *democracy.MaxSciencePercent != 80 {
		t.Fatalf("democracy government = %+v ok=%v", democracy, ok)
	}
	anarchy, ok := runtime.Government("GOVERNMENT_ANARCHY")
	if !ok || anarchy.ResearchPercent != -90 || anarchy.FoodPercent != -25 || anarchy.ProductionPercent != -25 || anarchy.GoldPercent != -25 {
		t.Fatalf("anarchy government = %+v ok=%v", anarchy, ok)
	}

	hagia, ok := runtime.ProductionDefinition("WONDER_HAGIA_SOPHIA")
	if !ok || hagia.Kind != "wonder" || hagia.Cost != 3240 || !sameStringSlices(hagia.Prerequisites, []string{"ADVANCE_THEOLOGY"}) {
		t.Fatalf("hagia production = %+v ok=%v", hagia, ok)
	}
	gutenberg, ok := runtime.ProductionDefinition("WONDER_GUTTENBERGS_BIBLE")
	if !ok || !runtimeProductionHasEffectValue(gutenberg, "happiness", 2) {
		t.Fatalf("Gutenberg effects = %+v ok=%v", gutenberg, ok)
	}
	zeroCrime, ok := runtime.ProductionDefinition("WONDER_ZERO_CRIME_BILL")
	if !ok || !runtimeProductionHasEffectValue(zeroCrime, "crime_percent", -60) {
		t.Fatalf("Zero Crime Bill effects = %+v ok=%v", zeroCrime, ok)
	}
	naniteDefuser, ok := runtime.ProductionDefinition("WONDER_NANITE_DEFUSER")
	if !ok || !runtimeProductionHasEffect(naniteDefuser, "eliminate_nukes") {
		t.Fatalf("Nanite Defuser effects = %+v ok=%v", naniteDefuser, ok)
	}
	agency, ok := runtime.ProductionDefinition("WONDER_THE_AGENCY")
	if !ok || !runtimeProductionHasEffect(agency, "spies_everywhere") {
		t.Fatalf("Agency effects = %+v ok=%v", agency, ok)
	}
	eastIndia, ok := runtime.ProductionDefinition("WONDER_EAST_INDIA_COMPANY")
	if !ok || !runtimeProductionHasEffectValue(eastIndia, "increase_boat_movement", 100) {
		t.Fatalf("East India effects = %+v ok=%v", eastIndia, ok)
	}
	factory, ok := runtime.ProductionDefinition("IMPROVE_FACTORY")
	if !ok || !runtimeProductionHasEffectValue(factory, "yield_production_percent", 15) || !runtimeProductionHasEffect(factory, "allow_grunts") {
		t.Fatalf("factory production effects = %+v ok=%v", factory, ok)
	}
	bazaar, ok := runtime.ProductionDefinition("IMPROVE_BAZAAR")
	if !ok || len(bazaar.Effects) != 1 || bazaar.Effects[0].Kind != "yield_gold_percent" || bazaar.Effects[0].Value != 10 {
		t.Fatalf("bazaar production effects = %+v ok=%v", bazaar, ok)
	}
	bank, ok := runtime.ProductionDefinition("IMPROVE_BANK")
	if !ok || !sameStringSlices(bank.RequiredBuildings, []string{"IMPROVE_BAZAAR"}) {
		t.Fatalf("bank production prerequisites = %+v ok=%v", bank, ok)
	}
	ecoTransit, ok := runtime.ProductionDefinition("IMPROVE_ECO_TRANSIT")
	if !ok || !runtimeProductionHasEffectValue(ecoTransit, "pollution_population_percent", -50) {
		t.Fatalf("eco transit production effects = %+v ok=%v", ecoTransit, ok)
	}
	airport, ok := runtime.ProductionDefinition("IMPROVE_AIRPORT")
	if !ok || !airport.CantBuildInSea || !runtimeProductionHasEffect(airport, "airport") {
		t.Fatalf("airport production effects = %+v ok=%v", airport, ok)
	}
	matter, ok := runtime.ProductionDefinition("IMPROVE_MATTER_DECOMPILER")
	if !ok || !runtimeProductionHasEffectValue(matter, "pollution_production_percent", -50) {
		t.Fatalf("matter decomposer production effects = %+v ok=%v", matter, ok)
	}

	television, ok := runtime.ProductionDefinition("IMPROVE_TELEVISION")
	if !ok || !runtimeProductionHasEffectValue(television, "yield_gold_per_population", 5) {
		t.Fatalf("television production effects = %+v ok=%v", television, ok)
	}

	academy, ok := runtime.ProductionDefinition("IMPROVE_ACADEMY")
	if !ok || len(academy.Effects) != 1 || academy.Effects[0].Kind != "research_percent" || academy.Effects[0].Value != 10 {
		t.Fatalf("academy production effects = %+v ok=%v", academy, ok)
	}
	granaryEffects, ok := runtime.ProductionDefinition("IMPROVE_GRANARY")
	if !ok || !runtimeProductionHasEffectValue(granaryEffects, "yield_food_percent", 10) || !runtimeProductionHasEffectValue(granaryEffects, "starvation_protection", 2) {
		t.Fatalf("granary production effects = %+v ok=%v", granaryEffects, ok)
	}
	foodSilo, ok := runtime.ProductionDefinition("IMPROVE_FOOD_SILO")
	if !ok || !runtimeProductionHasEffectValue(foodSilo, "yield_food_percent", 15) || !runtimeProductionHasEffectValue(foodSilo, "starvation_protection", 5) {
		t.Fatalf("food silo effects = %+v ok=%v", foodSilo, ok)
	}
	aqueduct, ok := runtime.ProductionDefinition("IMPROVE_AQUEDUCT")
	if !ok || !runtimeProductionHasEffectValue(aqueduct, "max_population", 14) || !runtimeProductionHasEffectValue(aqueduct, "overcrowding_level", 6) {
		t.Fatalf("aqueduct effects = %+v ok=%v", aqueduct, ok)
	}
	arena, ok := runtime.ProductionDefinition("IMPROVE_ARENA")
	if !ok || !runtimeProductionHasEffectValue(arena, "happiness", 2) {
		t.Fatalf("arena effects = %+v ok=%v", arena, ok)
	}
	courthouse, ok := runtime.ProductionDefinition("IMPROVE_COURTHOUSE")
	if !ok || !runtimeProductionHasEffectValue(courthouse, "crime_percent", -10) {
		t.Fatalf("courthouse effects = %+v ok=%v", courthouse, ok)
	}
	capitol, ok := runtime.ProductionDefinition("IMPROVE_CAPITOL")
	if !ok || !runtimeProductionHasEffect(capitol, "capitol") {
		t.Fatalf("capitol effects = %+v ok=%v", capitol, ok)
	}
	basilica, ok := runtime.ProductionDefinition("IMPROVE_BASCILICA")
	if !ok || !runtimeProductionHasEffect(basilica, "cathedral") || !runtimeProductionHasEffect(basilica, "is_religious") {
		t.Fatalf("basilica effects = %+v ok=%v", basilica, ok)
	}
	cornucopicVat, ok := runtime.ProductionDefinition("IMPROVE_CORNUCOPIC_VAT")
	if !ok || !runtimeProductionHasEffectValue(cornucopicVat, "food_vat_percent", 100) {
		t.Fatalf("cornucopic vat effects = %+v ok=%v", cornucopicVat, ok)
	}
	brokerage, ok := runtime.ProductionDefinition("IMPROVE_BROKERAGE")
	if !ok || !runtimeProductionHasEffect(brokerage, "brokerage") {
		t.Fatalf("brokerage effects = %+v ok=%v", brokerage, ok)
	}

	cityWalls, ok := runtime.ProductionDefinition("IMPROVE_CITY_WALLS")
	if !ok || !runtimeProductionHasEffectValue(cityWalls, "combat_defense", 15) || !runtimeProductionHasEffectValue(cityWalls, "city_walls", 15) || !runtimeProductionHasEffectValue(cityWalls, "prevent_conversion_percent", 50) || !runtimeProductionHasEffectValue(cityWalls, "prevent_slavery_percent", 50) {
		t.Fatalf("city walls effects = %+v ok=%v", cityWalls, ok)
	}
	missiles, ok := runtime.ProductionDefinition("IMPROVE_ANTI_BALLISTIC_MISSILES")
	if !ok || !runtimeProductionHasEffect(missiles, "protect_from_nukes") {
		t.Fatalf("missile protection effects = %+v ok=%v", missiles, ok)
	}
	forceField, ok := runtime.ProductionDefinition("IMPROVE_FORCEFIELD")
	if !ok || !runtimeProductionHasEffect(forceField, "force_field") {
		t.Fatalf("force field effects = %+v ok=%v", forceField, ok)
	}
	ballista, ok := runtime.ProductionDefinition("IMPROVE_BALLISTA_TOWERS")
	if !ok || !runtimeProductionHasEffectValue(ballista, "combat_attack_land", 20) {
		t.Fatalf("ballista effects = %+v ok=%v", ballista, ok)
	}
	battlements, ok := runtime.ProductionDefinition("IMPROVE_BATTLEMENTS")
	if !ok || !battlements.CoastalOnly || !runtimeProductionHasEffectValue(battlements, "combat_attack_sea", 25) {
		t.Fatalf("battlements effects = %+v ok=%v", battlements, ok)
	}
	flak, ok := runtime.ProductionDefinition("IMPROVE_FLAK_TOWERS")
	if !ok || !runtimeProductionHasEffectValue(flak, "combat_attack_air", 40) {
		t.Fatalf("flak tower effects = %+v ok=%v", flak, ok)
	}
	academyWonder, ok := runtime.ProductionDefinition("WONDER_ARISTOTLES_LYCEUM")
	if !ok || len(academyWonder.Effects) != 1 || academyWonder.Effects[0].Kind != "research_percent" || academyWonder.Effects[0].Value != 20 {
		t.Fatalf("Aristotle's Lyceum effects = %+v ok=%v", academyWonder, ok)
	}
	internet, ok := runtime.ProductionDefinition("WONDER_INTERNET")
	if !ok || !runtimeProductionHasEffectValue(internet, "research_percent", 20) {
		t.Fatalf("Internet actual-building effects = %+v ok=%v", internet, ok)
	}
	matterDecomp, ok := runtime.ProductionDefinition("WONDER_CENTRAL_MATTER_DECOMPILER")
	if !ok || !runtimeProductionHasEffectValue(matterDecomp, "pollution_production_percent", -50) {
		t.Fatalf("Central Matter Decompiler building-everywhere effects = %+v ok=%v", matterDecomp, ok)
	}
	pyramids, ok := runtime.ProductionDefinition("WONDER_PYRAMIDS")
	if !ok || len(pyramids.Effects) != 1 || pyramids.Effects[0].Kind != "yield_gold" || pyramids.Effects[0].Value != 100 {
		t.Fatalf("Pyramids effects = %+v ok=%v", pyramids, ok)
	}
}

func runtimeProductionHasEffect(production RuntimeProduction, kind string) bool {
	for _, effect := range production.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func runtimeProductionHasEffectValue(production RuntimeProduction, kind string, value int) bool {
	for _, effect := range production.Effects {
		if effect.Kind == kind && effect.Value == value {
			return true
		}
	}
	return false
}

func runtimeUnitHasEffect(unit RuntimeUnitDefinition, kind string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}

func runtimeUnitHasEffectValue(unit RuntimeUnitDefinition, kind string, value int) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind && effect.Value == value {
			return true
		}
	}
	return false
}

func runtimeUnitHasEffectID(unit RuntimeUnitDefinition, kind string, id string) bool {
	for _, effect := range unit.Effects {
		if effect.Kind == kind && effect.ID == id {
			return true
		}
	}
	return false
}

func runtimeStarterHasEffect(starter RuntimeStarter, kind string) bool {
	for _, effect := range starter.Effects {
		if effect.Kind == kind {
			return true
		}
	}
	return false
}
