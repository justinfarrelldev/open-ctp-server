package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RetailRules is a typed, source-preserving projection of the original CTP2
// retail database files. It is intentionally separate from RuntimeRules: it
// records the released-game catalog and exact source values before server
// gameplay support is implemented for each behavior-bearing field.
type RetailRules struct {
	Units                 []RetailUnit                 `json:"units"`
	Advances              []RetailAdvance              `json:"advances"`
	Wonders               []RetailWonder               `json:"wonders"`
	Buildings             []RetailBuilding             `json:"buildings"`
	TerrainImprovements   []RetailTerrainImprovement   `json:"terrain_improvements"`
	Terrains              []RetailTerrain              `json:"terrains"`
	Governments           []RetailGovernment           `json:"governments"`
	PollutionTriggerLists []RetailPollutionTriggerList `json:"pollution_trigger_lists"`
	PollutionModel        RetailPollutionModel         `json:"pollution_model"`
	AI                    RetailAILists                `json:"ai"`
}

type RetailUnit struct {
	ID                 string           `json:"id"`
	Description        string           `json:"description,omitempty"`
	Category           string           `json:"category,omitempty"`
	EnableAdvance      string           `json:"enable_advance,omitempty"`
	ObsoleteAdvance    []string         `json:"obsolete_advance,omitempty"`
	Attack             int              `json:"attack"`
	Defense            int              `json:"defense"`
	ZBRangeAttack      int              `json:"zb_range_attack"`
	Firepower          int              `json:"firepower"`
	Armor              int              `json:"armor"`
	MaxHP              int              `json:"max_hp"`
	ShieldCost         int              `json:"shield_cost"`
	PowerPoints        int              `json:"power_points"`
	ShieldHunger       int              `json:"shield_hunger"`
	GoldHunger         int              `json:"gold_hunger"`
	FoodHunger         int              `json:"food_hunger"`
	MaxMovePoints      int              `json:"max_move_points"`
	VisionRange        int              `json:"vision_range"`
	ActiveDefenseRange int              `json:"active_defense_range"`
	MaxFuel            int              `json:"max_fuel"`
	BombRounds         int              `json:"bomb_rounds"`
	BombardRange       int              `json:"bombard_range"`
	MovementTypes      []string         `json:"movement_types,omitempty"`
	CanAttack          []string         `json:"can_attack,omitempty"`
	CanCarry           []string         `json:"can_carry,omitempty"`
	CanSee             []string         `json:"can_see,omitempty"`
	Defend             []string         `json:"defend,omitempty"`
	Size               string           `json:"size,omitempty"`
	VisionClass        string           `json:"vision_class,omitempty"`
	CargoMax           int              `json:"cargo_max,omitempty"`
	Flags              []string         `json:"flags,omitempty"`
	RawFields          []RetailRawField `json:"raw_fields"`
}

type RetailAdvance struct {
	ID            string           `json:"id"`
	Cost          int              `json:"cost"`
	Prerequisites []string         `json:"prerequisites,omitempty"`
	Flags         []string         `json:"flags,omitempty"`
	RawFields     []RetailRawField `json:"raw_fields"`
}

type RetailWonder struct {
	ID              string             `json:"id"`
	Description     string             `json:"description,omitempty"`
	EnableAdvance   string             `json:"enable_advance,omitempty"`
	ObsoleteAdvance []string           `json:"obsolete_advance,omitempty"`
	ProductionCost  int                `json:"production_cost"`
	IntValues       map[string]int     `json:"int_values,omitempty"`
	FloatValues     map[string]float64 `json:"float_values,omitempty"`
	Flags           []string           `json:"flags,omitempty"`
	RawFields       []RetailRawField   `json:"raw_fields"`
}

type RetailBuilding struct {
	ID                   string             `json:"id"`
	Description          string             `json:"description,omitempty"`
	EnableAdvance        string             `json:"enable_advance,omitempty"`
	ObsoleteAdvance      []string           `json:"obsolete_advance,omitempty"`
	PrerequisiteBuilding []string           `json:"prerequisite_building,omitempty"`
	ProductionCost       int                `json:"production_cost"`
	Upkeep               int                `json:"upkeep"`
	IntValues            map[string]int     `json:"int_values,omitempty"`
	FloatValues          map[string]float64 `json:"float_values,omitempty"`
	Flags                []string           `json:"flags,omitempty"`
	RawFields            []RetailRawField   `json:"raw_fields"`
}

type RetailTerrainImprovement struct {
	ID                string                           `json:"id"`
	Class             string                           `json:"class,omitempty"`
	Level             int                              `json:"level"`
	ConstructionTiles []int                            `json:"construction_tiles,omitempty"`
	CantBuildOn       []string                         `json:"cant_build_on,omitempty"`
	Excludes          []string                         `json:"excludes,omitempty"`
	Effects           []RetailTerrainImprovementEffect `json:"effects,omitempty"`
	Flags             []string                         `json:"flags,omitempty"`
	RawFields         []RetailRawField                 `json:"raw_fields"`
}

type RetailTerrainImprovementEffect struct {
	Terrains        []string `json:"terrains,omitempty"`
	BonusFood       int      `json:"bonus_food,omitempty"`
	BonusProduction int      `json:"bonus_production,omitempty"`
	BonusGold       int      `json:"bonus_gold,omitempty"`
	MoveCost        int      `json:"move_cost,omitempty"`
	EnableAdvance   string   `json:"enable_advance,omitempty"`
	ProductionCost  int      `json:"production_cost,omitempty"`
	ProductionTime  int      `json:"production_time,omitempty"`
	TilesetIndex    int      `json:"tileset_index,omitempty"`
	Flags           []string `json:"flags,omitempty"`
}

type RetailTerrain struct {
	ID            string                 `json:"id"`
	InternalType  string                 `json:"internal_type,omitempty"`
	MovementTypes []string               `json:"movement_types,omitempty"`
	EnvBase       RetailTerrainModifiers `json:"env_base"`
	EnvCity       RetailTerrainModifiers `json:"env_city"`
	RawFields     []RetailRawField       `json:"raw_fields"`
}

type RetailTerrainModifiers struct {
	Score    int     `json:"score"`
	Food     int     `json:"food"`
	Shield   int     `json:"shield"`
	Gold     int     `json:"gold"`
	Defense  float64 `json:"defense"`
	Movement int     `json:"movement"`
	Freight  int     `json:"freight"`
}

type RetailGovernment struct {
	ID             string             `json:"id"`
	EnableAdvance  string             `json:"enable_advance,omitempty"`
	Rank           int                `json:"rank"`
	MaxScienceRate float64            `json:"max_science_rate"`
	PollutionCoef  float64            `json:"pollution_coef"`
	FloatValues    map[string]float64 `json:"float_values,omitempty"`
	IntValues      map[string]int     `json:"int_values,omitempty"`
	Flags          []string           `json:"flags,omitempty"`
	RawFields      []RetailRawField   `json:"raw_fields"`
}

type RetailPollutionTriggerList struct {
	ID            string           `json:"id"`
	TriggerLevels []int64          `json:"trigger_levels"`
	DisasterTypes []string         `json:"disaster_types"`
	RawFields     []RetailRawField `json:"raw_fields"`
}

type RetailPollutionModel struct {
	PopulationStart int `json:"population_start"`
	PopulationRatio int `json:"population_ratio"`
	ProductionStart int `json:"production_start"`
	ProductionRatio int `json:"production_ratio"`
}

type RetailAILists struct {
	UnitBuildLists     []RetailAIList `json:"unit_build_lists,omitempty"`
	BuildingBuildLists []RetailAIList `json:"building_build_lists,omitempty"`
	WonderBuildLists   []RetailAIList `json:"wonder_build_lists,omitempty"`
	AdvanceLists       []RetailAIList `json:"advance_lists,omitempty"`
	ImprovementLists   []RetailAIList `json:"improvement_lists,omitempty"`
}

type RetailAIList struct {
	ID    string   `json:"id"`
	Items []string `json:"items,omitempty"`
}

func RetailGameCatalogs() []RetailCatalog {
	return []RetailCatalog{
		{Filename: "Units.txt", Kind: "unit", Prefixes: []string{"UNIT_"}},
		{Filename: "Advance.txt", Kind: "advance", Prefixes: []string{"ADVANCE_"}},
		{Filename: "Wonder.txt", Kind: "wonder", Prefixes: []string{"WONDER_"}},
		{Filename: "buildings.txt", Kind: "building", Prefixes: []string{"IMPROVE_"}},
		{Filename: "tileimp.txt", Kind: "terrain_improvement", Prefixes: []string{"TILEIMP_"}},
		{Filename: "terrain.txt", Kind: "terrain", Prefixes: []string{"TERRAIN_"}},
		{Filename: "govern.txt", Kind: "government", Prefixes: []string{"GOVERNMENT_"}},
		{Filename: "pollution.txt", Kind: "pollution", Prefixes: []string{"POLLUTION_TRIGGER_LIST_"}},
	}
}

func LoadRetailRulesFromDirectory(directory string) (RetailRules, error) {
	if strings.TrimSpace(directory) == "" {
		return RetailRules{}, importError(ErrorInvalidDirectory, "", 1, "", "directory must be provided")
	}
	var result RetailRules
	pollutionModel, err := loadRetailPollutionModel(filepath.Join(directory, "DiffDB.txt"))
	if err != nil {
		return RetailRules{}, err
	}
	result.PollutionModel = pollutionModel
	for _, catalog := range RetailGameCatalogs() {
		path := filepath.Join(directory, catalog.Filename)
		data, err := os.ReadFile(path)
		if err != nil {
			return RetailRules{}, importError(ErrorInvalidDirectory, path, 1, "", err.Error())
		}
		records, err := ParseRetailRawRecords(catalog.Filename, data, catalog)
		if err != nil {
			return RetailRules{}, err
		}
		for _, record := range records {
			switch catalog.Kind {
			case "unit":
				compiled, err := compileRetailUnit(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.Units = append(result.Units, compiled)
			case "advance":
				compiled, err := compileRetailAdvance(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.Advances = append(result.Advances, compiled)
			case "wonder":
				compiled, err := compileRetailWonder(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.Wonders = append(result.Wonders, compiled)
			case "building":
				compiled, err := compileRetailBuilding(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.Buildings = append(result.Buildings, compiled)
			case "terrain_improvement":
				compiled, err := compileRetailTerrainImprovement(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.TerrainImprovements = append(result.TerrainImprovements, compiled)
			case "terrain":
				compiled, err := compileRetailTerrain(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.Terrains = append(result.Terrains, compiled)
			case "government":
				compiled, err := compileRetailGovernment(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.Governments = append(result.Governments, compiled)
			case "pollution":
				compiled, err := compileRetailPollutionTriggerList(record)
				if err != nil {
					return RetailRules{}, err
				}
				result.PollutionTriggerLists = append(result.PollutionTriggerLists, compiled)
			}
		}
	}
	ai, err := loadRetailAILists(filepath.Join(filepath.Dir(directory), "aidata"))
	if err != nil {
		return RetailRules{}, err
	}
	result.AI = ai
	return result, nil
}

func loadRetailPollutionModel(path string) (RetailPollutionModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RetailPollutionModel{}, importError(ErrorInvalidDirectory, path, 1, "", err.Error())
	}
	tokens, err := lexRetailCTP2(string(data), filepath.Base(path))
	if err != nil {
		return RetailPollutionModel{}, err
	}
	for index := 0; index < len(tokens); index++ {
		if tokens[index].value != "{" {
			continue
		}
		fields, _, err := parseRetailFields(tokens, index+1, filepath.Base(path), "DIFFICULTY")
		if err != nil {
			return RetailPollutionModel{}, err
		}
		model, err := compileRetailPollutionModel(fields)
		if err != nil {
			return RetailPollutionModel{}, err
		}
		if model != (RetailPollutionModel{}) {
			return model, nil
		}
	}
	return RetailPollutionModel{}, importError(ErrorMalformedValue, path, 1, "DIFFICULTY", "pollution model fields missing")
}

func compileRetailPollutionModel(fields []RetailRawField) (RetailPollutionModel, error) {
	value := func(key string) (int, error) {
		for _, field := range fields {
			if field.Key == key && len(field.Values) > 0 {
				return retailParseInt("DIFFICULTY", key, field.Values[0])
			}
		}
		return 0, nil
	}
	populationStart, err := value("POLLUTION_START_POPULATION_LEVEL")
	if err != nil {
		return RetailPollutionModel{}, err
	}
	populationRatio, err := value("POLLUTION_POPULATION_RATIO")
	if err != nil {
		return RetailPollutionModel{}, err
	}
	productionStart, err := value("POLLUTION_START_PRODUCTION_LEVEL")
	if err != nil {
		return RetailPollutionModel{}, err
	}
	productionRatio, err := value("POLLUTION_PRODUCTION_RATIO")
	if err != nil {
		return RetailPollutionModel{}, err
	}
	return RetailPollutionModel{PopulationStart: populationStart, PopulationRatio: populationRatio, ProductionStart: productionStart, ProductionRatio: productionRatio}, nil
}

func loadRetailAILists(directory string) (RetailAILists, error) {
	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return RetailAILists{}, nil
		}
		return RetailAILists{}, importError(ErrorInvalidDirectory, directory, 1, "", err.Error())
	}
	var result RetailAILists
	var err error
	if result.UnitBuildLists, err = parseRetailAIListFile(filepath.Join(directory, "UnitBuildLists.txt"), "Unit"); err != nil {
		return RetailAILists{}, err
	}
	if result.BuildingBuildLists, err = parseRetailAIListFile(filepath.Join(directory, "BuildingBuildLists.txt"), "Building"); err != nil {
		return RetailAILists{}, err
	}
	if result.WonderBuildLists, err = parseRetailAIListFile(filepath.Join(directory, "WonderBuildLists.txt"), "Wonder"); err != nil {
		return RetailAILists{}, err
	}
	if result.AdvanceLists, err = parseRetailAIListFile(filepath.Join(directory, "AdvanceLists.txt"), "Advance"); err != nil {
		return RetailAILists{}, err
	}
	if result.ImprovementLists, err = parseRetailAIListFile(filepath.Join(directory, "ImprovementLists.txt"), "Improvement"); err != nil {
		return RetailAILists{}, err
	}
	return result, nil
}

func parseRetailAIListFile(path string, itemKey string) ([]RetailAIList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, importError(ErrorInvalidDirectory, path, 1, "", err.Error())
	}
	lines := strings.Split(string(data), "\n")
	var lists []RetailAIList
	var current *RetailAIList
	var pendingID string
	for index, line := range lines {
		lineNumber := index + 1
		if comment := strings.Index(line, "#"); comment >= 0 {
			line = line[:comment]
		}
		if comment := strings.Index(line, "//"); comment >= 0 {
			line = line[:comment]
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if current == nil {
			if pendingID != "" {
				if len(fields) == 1 && fields[0] == "{" {
					current = &RetailAIList{ID: pendingID}
					pendingID = ""
					continue
				}
				return nil, importError(ErrorMalformedValue, path, lineNumber, trimmed, "expected AI list opening brace")
			}
			if len(fields) == 1 {
				pendingID = fields[0]
				continue
			}
			if len(fields) >= 2 && fields[1] == "{" {
				current = &RetailAIList{ID: fields[0]}
				continue
			}
			return nil, importError(ErrorMalformedValue, path, lineNumber, trimmed, "expected AI list header")
		}
		if fields[0] == "}" {
			lists = append(lists, *current)
			current = nil
			continue
		}
		if fields[0] == itemKey && len(fields) >= 2 {
			current.Items = append(current.Items, fields[1])
			continue
		}
		return nil, importError(ErrorMalformedValue, path, lineNumber, trimmed, "expected AI list item")
	}
	if pendingID != "" {
		return nil, importError(ErrorMalformedValue, path, len(lines), pendingID, "unterminated AI list header")
	}
	if current != nil {
		return nil, importError(ErrorMalformedValue, path, len(lines), current.ID, "unterminated AI list")
	}
	return lists, nil
}

func compileRetailUnit(record RetailRawRecord) (RetailUnit, error) {
	unit := RetailUnit{ID: record.ID, RawFields: cloneRetailFields(record.Fields)}
	unit.Description = record.FirstValue("Description")
	unit.Category = record.FirstValue("Category")
	unit.EnableAdvance = record.FirstValue("EnableAdvance")
	unit.ObsoleteAdvance = record.Values("ObsoleteAdvance")
	unit.MovementTypes = record.Values("MovementType")
	unit.CanAttack = record.Values("CanAttack")
	unit.CanCarry = record.Values("CanCarry")
	unit.CanSee = record.Values("CanSee")
	unit.Defend = record.Values("Defend")
	unit.Size = record.FirstValue("Size")
	unit.VisionClass = record.FirstValue("VisionClass")
	unit.Flags = retailFlags(record.Fields)
	var err error
	for key, target := range map[string]*int{
		"Attack": &unit.Attack, "Defense": &unit.Defense, "ZBRangeAttack": &unit.ZBRangeAttack, "Firepower": &unit.Firepower, "Armor": &unit.Armor, "MaxHP": &unit.MaxHP, "ShieldCost": &unit.ShieldCost, "PowerPoints": &unit.PowerPoints, "ShieldHunger": &unit.ShieldHunger, "GoldHunger": &unit.GoldHunger, "FoodHunger": &unit.FoodHunger, "MaxMovePoints": &unit.MaxMovePoints, "VisionRange": &unit.VisionRange, "ActiveDefenseRange": &unit.ActiveDefenseRange, "MaxFuel": &unit.MaxFuel, "BombRounds": &unit.BombRounds, "BombardRange": &unit.BombardRange,
	} {
		*target, err = retailRecordInt(record, key)
		if err != nil {
			return RetailUnit{}, err
		}
	}
	if cargo, ok := record.FirstBlock("CargoData"); ok {
		unit.CargoMax, err = retailFieldInt(record.ID, cargo, "MaxCargo")
		if err != nil {
			return RetailUnit{}, err
		}
	}
	return unit, nil
}

func compileRetailAdvance(record RetailRawRecord) (RetailAdvance, error) {
	cost, err := retailRecordInt(record, "Cost")
	if err != nil {
		return RetailAdvance{}, err
	}
	return RetailAdvance{ID: record.ID, Cost: cost, Prerequisites: record.Values("Prerequisites"), Flags: retailFlags(record.Fields), RawFields: cloneRetailFields(record.Fields)}, nil
}

func compileRetailWonder(record RetailRawRecord) (RetailWonder, error) {
	cost, err := retailRecordInt(record, "ProductionCost")
	if err != nil {
		return RetailWonder{}, err
	}
	return RetailWonder{ID: record.ID, Description: record.FirstValue("Description"), EnableAdvance: record.FirstValue("EnableAdvance"), ObsoleteAdvance: record.Values("ObsoleteAdvance"), ProductionCost: cost, IntValues: retailIntValues(record.Fields), FloatValues: retailFloatValues(record.Fields), Flags: retailFlags(record.Fields), RawFields: cloneRetailFields(record.Fields)}, nil
}

func compileRetailBuilding(record RetailRawRecord) (RetailBuilding, error) {
	cost, err := retailRecordInt(record, "ProductionCost")
	if err != nil {
		return RetailBuilding{}, err
	}
	upkeep, err := retailRecordInt(record, "Upkeep")
	if err != nil {
		return RetailBuilding{}, err
	}
	return RetailBuilding{ID: record.ID, Description: record.FirstValue("Description"), EnableAdvance: record.FirstValue("EnableAdvance"), ObsoleteAdvance: record.Values("ObsoleteAdvance"), PrerequisiteBuilding: record.Values("PrerequisiteBuilding"), ProductionCost: cost, Upkeep: upkeep, IntValues: retailIntValues(record.Fields), FloatValues: retailFloatValues(record.Fields), Flags: retailFlags(record.Fields), RawFields: cloneRetailFields(record.Fields)}, nil
}

func compileRetailTerrainImprovement(record RetailRawRecord) (RetailTerrainImprovement, error) {
	level, err := retailRecordInt(record, "Level")
	if err != nil {
		return RetailTerrainImprovement{}, err
	}
	constructionTiles, err := retailRecordInts(record, "ConstructionTiles")
	if err != nil {
		return RetailTerrainImprovement{}, err
	}
	improvement := RetailTerrainImprovement{ID: record.ID, Class: record.FirstValue("Class"), Level: level, ConstructionTiles: constructionTiles, CantBuildOn: record.Values("CantBuildOn"), Excludes: record.Values("Excludes"), Flags: retailFlags(record.Fields), RawFields: cloneRetailFields(record.Fields)}
	for _, field := range record.FieldsByKey("TerrainEffect") {
		effect := RetailTerrainImprovementEffect{Terrains: field.ChildValues("Terrain"), EnableAdvance: field.FirstValue("EnableAdvance"), Flags: retailFlags(field.Fields)}
		for key, target := range map[string]*int{"BonusFood": &effect.BonusFood, "BonusProduction": &effect.BonusProduction, "BonusGold": &effect.BonusGold, "MoveCost": &effect.MoveCost, "ProductionCost": &effect.ProductionCost, "ProductionTime": &effect.ProductionTime, "TilesetIndex": &effect.TilesetIndex} {
			*target, err = retailFieldInt(record.ID, field, key)
			if err != nil {
				return RetailTerrainImprovement{}, err
			}
		}
		improvement.Effects = append(improvement.Effects, effect)
	}
	return improvement, nil
}

func compileRetailTerrain(record RetailRawRecord) (RetailTerrain, error) {
	terrain := RetailTerrain{ID: record.ID, InternalType: record.FirstValue("InternalType"), MovementTypes: record.Values("MovementType"), RawFields: cloneRetailFields(record.Fields)}
	if envBase, ok := record.FirstBlock("EnvBase"); ok {
		modifiers, err := retailTerrainModifiers(record.ID, envBase)
		if err != nil {
			return RetailTerrain{}, err
		}
		terrain.EnvBase = modifiers
	}
	if envCity, ok := record.FirstBlock("EnvCity"); ok {
		modifiers, err := retailTerrainModifiers(record.ID, envCity)
		if err != nil {
			return RetailTerrain{}, err
		}
		terrain.EnvCity = modifiers
	}
	return terrain, nil
}

func compileRetailGovernment(record RetailRawRecord) (RetailGovernment, error) {
	rank, err := retailRecordInt(record, "Rank")
	if err != nil {
		return RetailGovernment{}, err
	}
	maxScience, err := retailRecordFloat(record, "MaxScienceRate")
	if err != nil {
		return RetailGovernment{}, err
	}
	pollutionCoef, err := retailRecordFloat(record, "PollutionCoef")
	if err != nil {
		return RetailGovernment{}, err
	}
	return RetailGovernment{ID: record.ID, EnableAdvance: record.FirstValue("EnableAdvance"), Rank: rank, MaxScienceRate: maxScience, PollutionCoef: pollutionCoef, FloatValues: retailFloatValues(record.Fields), IntValues: retailIntValues(record.Fields), Flags: retailFlags(record.Fields), RawFields: cloneRetailFields(record.Fields)}, nil
}

func compileRetailPollutionTriggerList(record RetailRawRecord) (RetailPollutionTriggerList, error) {
	list := RetailPollutionTriggerList{ID: record.ID, RawFields: cloneRetailFields(record.Fields)}
	for _, field := range record.FieldsByKey("POLLUTION_TRIGGER") {
		if len(field.Values) < 2 {
			return RetailPollutionTriggerList{}, fmt.Errorf("retail record %q has malformed POLLUTION_TRIGGER", record.ID)
		}
		level, err := strconv.ParseInt(field.Values[0], 10, 64)
		if err != nil {
			return RetailPollutionTriggerList{}, fmt.Errorf("retail record %q field POLLUTION_TRIGGER: %w", record.ID, err)
		}
		list.TriggerLevels = append(list.TriggerLevels, level)
		list.DisasterTypes = append(list.DisasterTypes, field.Values[1])
	}
	return list, nil
}

func (r RetailRules) Unit(id string) (RetailUnit, bool) {
	for _, value := range r.Units {
		if value.ID == id {
			return value, true
		}
	}
	return RetailUnit{}, false
}

func (r RetailRules) Advance(id string) (RetailAdvance, bool) {
	for _, value := range r.Advances {
		if value.ID == id {
			return value, true
		}
	}
	return RetailAdvance{}, false
}

func (r RetailRules) Wonder(id string) (RetailWonder, bool) {
	for _, value := range r.Wonders {
		if value.ID == id {
			return value, true
		}
	}
	return RetailWonder{}, false
}

func (r RetailRules) Building(id string) (RetailBuilding, bool) {
	for _, value := range r.Buildings {
		if value.ID == id {
			return value, true
		}
	}
	return RetailBuilding{}, false
}

func (r RetailRules) TerrainImprovement(id string) (RetailTerrainImprovement, bool) {
	for _, value := range r.TerrainImprovements {
		if value.ID == id {
			return value, true
		}
	}
	return RetailTerrainImprovement{}, false
}

func (r RetailRules) Terrain(id string) (RetailTerrain, bool) {
	for _, value := range r.Terrains {
		if value.ID == id {
			return value, true
		}
	}
	return RetailTerrain{}, false
}

func (r RetailRules) Government(id string) (RetailGovernment, bool) {
	for _, value := range r.Governments {
		if value.ID == id {
			return value, true
		}
	}
	return RetailGovernment{}, false
}

func (r RetailRules) PollutionTriggerList(id string) (RetailPollutionTriggerList, bool) {
	for _, value := range r.PollutionTriggerLists {
		if value.ID == id {
			return value, true
		}
	}
	return RetailPollutionTriggerList{}, false
}

func (r RetailRawRecord) FieldsByKey(key string) []RetailRawField {
	return retailFieldsByKey(r.Fields, key)
}

func retailFieldsByKey(fields []RetailRawField, key string) []RetailRawField {
	matches := make([]RetailRawField, 0)
	for _, field := range fields {
		if field.Key == key {
			matches = append(matches, field)
		}
	}
	return matches
}

func retailRecordInt(record RetailRawRecord, key string) (int, error) {
	return retailParseInt(record.ID, key, record.FirstValue(key))
}

func retailRecordInts(record RetailRawRecord, key string) ([]int, error) {
	values := record.Values(key)
	result := make([]int, 0, len(values))
	for _, value := range values {
		parsed, err := retailParseInt(record.ID, key, value)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

func retailFieldInt(recordID string, field RetailRawField, key string) (int, error) {
	return retailParseInt(recordID, key, field.FirstValue(key))
}

func retailRecordFloat(record RetailRawRecord, key string) (float64, error) {
	return retailParseFloat(record.ID, key, record.FirstValue(key))
}

func retailParseInt(recordID, key, value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("retail record %q field %s: %w", recordID, key, err)
	}
	return parsed, nil
}

func retailParseFloat(recordID, key, value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("retail record %q field %s: %w", recordID, key, err)
	}
	return parsed, nil
}

func retailTerrainModifiers(recordID string, field RetailRawField) (RetailTerrainModifiers, error) {
	var modifiers RetailTerrainModifiers
	var err error
	for key, target := range map[string]*int{"Score": &modifiers.Score, "Food": &modifiers.Food, "Shield": &modifiers.Shield, "Gold": &modifiers.Gold, "Movement": &modifiers.Movement, "Freight": &modifiers.Freight} {
		*target, err = retailFieldInt(recordID, field, key)
		if err != nil {
			return RetailTerrainModifiers{}, err
		}
	}
	modifiers.Defense, err = retailParseFloat(recordID, "Defense", field.FirstValue("Defense"))
	if err != nil {
		return RetailTerrainModifiers{}, err
	}
	return modifiers, nil
}

func retailFlags(fields []RetailRawField) []string {
	flags := make([]string, 0)
	for _, field := range fields {
		if len(field.Values) == 0 && len(field.Fields) == 0 {
			flags = append(flags, field.Key)
		}
	}
	return flags
}

func retailIntValues(fields []RetailRawField) map[string]int {
	values := map[string]int{}
	for _, field := range fields {
		if len(field.Values) != 1 {
			continue
		}
		parsed, err := strconv.Atoi(field.Values[0])
		if err == nil {
			values[field.Key] = parsed
		}
	}
	return values
}

func retailFloatValues(fields []RetailRawField) map[string]float64 {
	values := map[string]float64{}
	for _, field := range fields {
		if len(field.Values) != 1 {
			continue
		}
		parsed, err := strconv.ParseFloat(field.Values[0], 64)
		if err == nil {
			values[field.Key] = parsed
		}
	}
	return values
}

func cloneRetailFields(fields []RetailRawField) []RetailRawField {
	if len(fields) == 0 {
		return nil
	}
	out := make([]RetailRawField, len(fields))
	for i, field := range fields {
		out[i] = RetailRawField{Key: field.Key, Values: append([]string(nil), field.Values...), Fields: cloneRetailFields(field.Fields)}
	}
	return out
}
