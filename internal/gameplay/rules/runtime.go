package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const calendarFinalTurn = 48

type RuntimeRules struct {
	CompatibilityVersion string                                `json:"compatibility_version"`
	ContentHash          string                                `json:"content_hash"`
	Victory              []RuntimeVictory                      `json:"victory"`
	Calendar             RuntimeCalendar                       `json:"calendar"`
	Constants            RuntimeConstants                      `json:"constants,omitempty"`
	Gaia                 RuntimeGaia                           `json:"gaia"`
	Feats                []RuntimeFeatDefinition               `json:"feats,omitempty"`
	Production           []RuntimeProduction                   `json:"production"`
	Effects              []RuntimeEffect                       `json:"effects"`
	Starter              RuntimeStarter                        `json:"starter"`
	Movement             RuntimeMovement                       `json:"movement"`
	Advances             []RuntimeAdvanceDefinition            `json:"advances,omitempty"`
	TerrainImprovements  []RuntimeTerrainImprovementDefinition `json:"terrain_improvements,omitempty"`
	PollutionSources     []RuntimePollutionSourceDefinition    `json:"pollution_sources,omitempty"`
	PollutionTriggers    []RuntimePollutionTriggerDefinition   `json:"pollution_triggers,omitempty"`
	PollutionModel       *RuntimePollutionModel                `json:"pollution_model,omitempty"`
	Governments          []RuntimeGovernmentDefinition         `json:"governments,omitempty"`
	Units                []RuntimeUnitDefinition               `json:"units"`
	AI                   *RuntimeAI                            `json:"ai,omitempty"`
}

// RuntimeAdvanceDefinition is the bounded synthetic technology model. It keeps
// CTP2's prerequisite/cost shape without copying the retail tree.
type RuntimeAdvanceDefinition struct {
	ID                          string   `json:"id"`
	Cost                        int      `json:"cost"`
	Prerequisites               []string `json:"prerequisites,omitempty"`
	TriggerFeatID               string   `json:"trigger_feat_id,omitempty"`
	AllowsPlantNuke             bool     `json:"allows_plant_nuke,omitempty"`
	EnablesCapitalization       bool     `json:"enables_capitalization,omitempty"`
	EnablesInfrastructure       bool     `json:"enables_infrastructure,omitempty"`
	EnablesTransform            bool     `json:"enables_transform,omitempty"`
	EnablesRoads                bool     `json:"enables_roads,omitempty"`
	EnablesRailroads            bool     `json:"enables_railroads,omitempty"`
	EnablesMagLev               bool     `json:"enables_maglev,omitempty"`
	EnablesTunnels              bool     `json:"enables_tunnels,omitempty"`
	EnablesDeepOcean            bool     `json:"enables_deep_ocean,omitempty"`
	PollutionSizeModifier       int      `json:"pollution_size_modifier,omitempty"`
	PollutionProductionModifier int      `json:"pollution_production_modifier,omitempty"`
}

// RuntimeUnitDefinition is the bounded, client-safe unit capability model compiled
// from synthetic CTP2 unit records. It intentionally models only mechanics the
// server can execute deterministically; unsupported source tags still fail closed
// during import/runtime compilation.
type RuntimeFeatDefinition struct {
	ID       string                 `json:"id"`
	Duration int                    `json:"duration,omitempty"`
	Effects  []RuntimeStarterEffect `json:"effects,omitempty"`
}

type RuntimeUnitDefinition struct {
	ID                           string                 `json:"id"`
	Role                         string                 `json:"role,omitempty"`
	Domain                       string                 `json:"domain"`
	Movement                     int                    `json:"movement"`
	Attack                       int                    `json:"attack"`
	Defense                      int                    `json:"defense"`
	Ranged                       int                    `json:"ranged,omitempty"`
	BombardRange                 int                    `json:"bombard_range,omitempty"`
	BombRounds                   int                    `json:"bomb_rounds,omitempty"`
	ActiveDefenseRange           int                    `json:"active_defense_range,omitempty"`
	Strength                     int                    `json:"strength"`
	Armor                        int                    `json:"armor"`
	Firepower                    int                    `json:"firepower,omitempty"`
	Health                       int                    `json:"health"`
	MaxFuel                      int                    `json:"max_fuel,omitempty"`
	ZOC                          bool                   `json:"zoc"`
	TransportCapacity            int                    `json:"transport_capacity,omitempty"`
	CargoDomains                 []string               `json:"cargo_domains,omitempty"`
	ProductionCost               int                    `json:"production_cost,omitempty"`
	PowerPoints                  int                    `json:"power_points,omitempty"`
	Upkeep                       int                    `json:"upkeep,omitempty"`
	ShieldHunger                 int                    `json:"shield_hunger,omitempty"`
	FoodHunger                   int                    `json:"food_hunger,omitempty"`
	NeedsNoSupport               bool                   `json:"needs_no_support,omitempty"`
	IsSpecialForces              bool                   `json:"is_special_forces,omitempty"`
	Prerequisites                []string               `json:"prerequisites,omitempty"`
	RequiredGovernmentID         string                 `json:"required_government_id,omitempty"`
	NeedsCityGoodID              string                 `json:"needs_city_good_id,omitempty"`
	SettleCityTypeID             string                 `json:"settle_city_type_id,omitempty"`
	CityGrowthCoefficientPercent int                    `json:"city_growth_coefficient_percent,omitempty"`
	ObsoleteBy                   []string               `json:"obsolete_by,omitempty"`
	UpgradeTo                    []string               `json:"upgrade_to,omitempty"`
	VisibilityRange              int                    `json:"visibility_range,omitempty"`
	Effects                      []RuntimeStarterEffect `json:"effects,omitempty"`
}

type RuntimeTerrainImprovementDefinition struct {
	ID                  string   `json:"id"`
	Class               string   `json:"class,omitempty"`
	BuildTurns          int      `json:"build_turns"`
	PublicWorksCost     int      `json:"public_works_cost"`
	MovementCostDelta   int      `json:"movement_cost_delta,omitempty"`
	YieldFood           int      `json:"yield_food,omitempty"`
	YieldProduction     int      `json:"yield_production,omitempty"`
	YieldGold           int      `json:"yield_gold,omitempty"`
	PollutionDelta      int      `json:"pollution_delta,omitempty"`
	TargetTerrain       string   `json:"target_terrain,omitempty"`
	RequiredAdvanceID   string   `json:"required_advance_id,omitempty"`
	AllowedTerrains     []string `json:"allowed_terrains,omitempty"`
	DefenseBonusPercent int      `json:"defense_bonus_percent,omitempty"`
	HealRatePercent     int      `json:"heal_rate_percent,omitempty"`
	VisionRange         int      `json:"vision_range,omitempty"`
	RadarRange          int      `json:"radar_range,omitempty"`
	Airport             bool     `json:"airport,omitempty"`
	Radar               bool     `json:"radar,omitempty"`
	ListeningPost       bool     `json:"listening_post,omitempty"`
	CanUpgrade          bool     `json:"can_upgrade,omitempty"`
	Endgame             bool     `json:"endgame,omitempty"`
	CantBuildOn         []string `json:"cant_build_on,omitempty"`
	Excludes            []string `json:"excludes,omitempty"`
}

type RuntimePollutionSourceDefinition struct {
	ID              string `json:"id"`
	PollutionPoints int    `json:"pollution_points"`
	EffectScope     string `json:"effect_scope"`
}

type RuntimePollutionTriggerDefinition struct {
	ID            string   `json:"id"`
	TriggerLevels []int64  `json:"trigger_levels"`
	DisasterTypes []string `json:"disaster_types"`
}

type RuntimePollutionModel struct {
	PopulationStart int `json:"population_start,omitempty"`
	PopulationRatio int `json:"population_ratio,omitempty"`
	ProductionStart int `json:"production_start,omitempty"`
	ProductionRatio int `json:"production_ratio,omitempty"`
}

type RuntimeGovernmentDefinition struct {
	ID                        string  `json:"id"`
	EnableAdvanceID           string  `json:"enable_advance_id,omitempty"`
	ResearchBonus             int     `json:"research_bonus,omitempty"`
	ResearchPercent           int     `json:"research_percent,omitempty"`
	PublicWorksBonus          int     `json:"public_works_bonus,omitempty"`
	InfrastructureCoefficient float64 `json:"infrastructure_coefficient,omitempty"`
	CapitalizationCoefficient float64 `json:"capitalization_coefficient,omitempty"`
	FoodPercent               int     `json:"food_percent,omitempty"`
	ProductionPercent         int     `json:"production_percent,omitempty"`
	GoldPercent               int     `json:"gold_percent,omitempty"`
	PollutionPercent          int     `json:"pollution_percent,omitempty"`
	PollutionUnhappyCoef      float64 `json:"pollution_unhappy_coef,omitempty"`
	CrimeCoef                 float64 `json:"crime_coef,omitempty"`
	CrimeOffset               int     `json:"crime_offset,omitempty"`
	WorkdayExpectation        int     `json:"workday_expectation,omitempty"`
	PositiveWorkdayCoef       float64 `json:"positive_workday_coef,omitempty"`
	NegativeWorkdayCoef       float64 `json:"negative_workday_coef,omitempty"`
	WagesExpectation          int     `json:"wages_expectation,omitempty"`
	PositiveWagesCoef         float64 `json:"positive_wages_coef,omitempty"`
	NegativeWagesCoef         float64 `json:"negative_wages_coef,omitempty"`
	RationsExpectation        int     `json:"rations_expectation,omitempty"`
	PositiveRationsCoef       float64 `json:"positive_rations_coef,omitempty"`
	NegativeRationsCoef       float64 `json:"negative_rations_coef,omitempty"`
	SupportCoef               float64 `json:"support_coef,omitempty"`
	DefenseCoef               float64 `json:"defense_coef,omitempty"`
	ProfessionalUnits         int     `json:"professional_units,omitempty"`
	TurnsToNewReadiness       int     `json:"turns_to_new_readiness,omitempty"`
	ReadyPeaceCoef            float64 `json:"ready_peace_coef,omitempty"`
	ReadyPeaceHP              float64 `json:"ready_peace_hp,omitempty"`
	ReadyAlertCoef            float64 `json:"ready_alert_coef,omitempty"`
	ReadyAlertHP              float64 `json:"ready_alert_hp,omitempty"`
	ReadyWarCoef              float64 `json:"ready_war_coef,omitempty"`
	ReadyWarHP                float64 `json:"ready_war_hp,omitempty"`
	WarDiscontentMaxUnits     int     `json:"war_discontent_max_units,omitempty"`
	WarDiscontentPerUnit      float64 `json:"war_discontent_per_unit,omitempty"`
	ConquestDistress          float64 `json:"conquest_distress,omitempty"`
	ConquestDistressDecay     float64 `json:"conquest_distress_decay,omitempty"`
	AtHomeRadius              int     `json:"at_home_radius,omitempty"`
	OverseasCoef              float64 `json:"overseas_coef,omitempty"`
	OverseasDefeatDecay       float64 `json:"overseas_defeat_decay,omitempty"`
	OverseasDefeatCoef        float64 `json:"overseas_defeat_coef,omitempty"`
	HomeDefeatDecay           float64 `json:"home_defeat_decay,omitempty"`
	HomeDefeatCoef            float64 `json:"home_defeat_coef,omitempty"`
	EmpireDistanceScale       float64 `json:"empire_distance_scale,omitempty"`
	MinEmpireDistance         int     `json:"min_empire_distance,omitempty"`
	MaxEmpireDistance         int     `json:"max_empire_distance,omitempty"`
	TooManyCitiesThreshold    int     `json:"too_many_cities_threshold,omitempty"`
	TooManyCitiesCoefficient  float64 `json:"too_many_cities_coefficient,omitempty"`
	MartialLawMaxUnits        int     `json:"martial_law_max_units,omitempty"`
	MartialLawEffect          int     `json:"martial_law_effect,omitempty"`
	MartialLawThreshold       int     `json:"martial_law_threshold,omitempty"`
	UnitRushModifier          float64 `json:"unit_rush_modifier,omitempty"`
	BuildingRushModifier      float64 `json:"building_rush_modifier,omitempty"`
	WonderRushModifier        float64 `json:"wonder_rush_modifier,omitempty"`
	EndGameRushModifier       float64 `json:"endgame_rush_modifier,omitempty"`
	MaxIncomingTrade          int     `json:"max_incoming_trade,omitempty"`
	MaxOutgoingTrade          int     `json:"max_outgoing_trade,omitempty"`
	MaxSciencePercent         *int    `json:"max_science_percent,omitempty"`
}

type RuntimeMovement struct {
	Configured       bool             `json:"configured"`
	PolicyConfigured bool             `json:"-"`
	OpenAccess       bool             `json:"open_access"`
	ZOCStopsPath     bool             `json:"zoc_stops_path"`
	Terrains         []RuntimeTerrain `json:"terrains"`
}

type RuntimeTerrain struct {
	ID                       string   `json:"id"`
	Cost                     int      `json:"cost"`
	Domains                  []string `json:"domains"`
	Food                     int      `json:"food,omitempty"`
	Production               int      `json:"production,omitempty"`
	Gold                     int      `json:"gold,omitempty"`
	CityFood                 int      `json:"city_food,omitempty"`
	CityProduction           int      `json:"city_production,omitempty"`
	CityGold                 int      `json:"city_gold,omitempty"`
	DefenseBonusPercent      int      `json:"defense_bonus_percent,omitempty"`
	TransformAddTime         int      `json:"transform_add_time,omitempty"`
	TransformAddMaterials    int      `json:"transform_add_materials,omitempty"`
	TransformRemoveTime      int      `json:"transform_remove_time,omitempty"`
	TransformRemoveMaterials int      `json:"transform_remove_materials,omitempty"`
}

func (movement RuntimeMovement) Terrain(id string) (RuntimeTerrain, bool) {
	index := sort.Search(len(movement.Terrains), func(index int) bool { return movement.Terrains[index].ID >= id })
	if index == len(movement.Terrains) || movement.Terrains[index].ID != id {
		return RuntimeTerrain{}, false
	}
	return movement.Terrains[index], true
}

func (terrain RuntimeTerrain) Allows(domain string) bool {
	index := sort.SearchStrings(terrain.Domains, domain)
	return index < len(terrain.Domains) && terrain.Domains[index] == domain
}

func (runtime RuntimeRules) Feat(id string) (RuntimeFeatDefinition, bool) {
	index := sort.Search(len(runtime.Feats), func(index int) bool { return runtime.Feats[index].ID >= id })
	if index == len(runtime.Feats) || runtime.Feats[index].ID != id {
		return RuntimeFeatDefinition{}, false
	}
	return runtime.Feats[index], true
}

func (runtime RuntimeRules) Advance(id string) (RuntimeAdvanceDefinition, bool) {
	index := sort.Search(len(runtime.Advances), func(index int) bool { return runtime.Advances[index].ID >= id })
	if index == len(runtime.Advances) || runtime.Advances[index].ID != id {
		return RuntimeAdvanceDefinition{}, false
	}
	return runtime.Advances[index], true
}

func (runtime RuntimeRules) Government(id string) (RuntimeGovernmentDefinition, bool) {
	index := sort.Search(len(runtime.Governments), func(index int) bool { return runtime.Governments[index].ID >= id })
	if index == len(runtime.Governments) || runtime.Governments[index].ID != id {
		return RuntimeGovernmentDefinition{}, false
	}
	return runtime.Governments[index], true
}

func (runtime RuntimeRules) TerrainImprovement(id string) (RuntimeTerrainImprovementDefinition, bool) {
	index := sort.Search(len(runtime.TerrainImprovements), func(index int) bool { return runtime.TerrainImprovements[index].ID >= id })
	if index == len(runtime.TerrainImprovements) || runtime.TerrainImprovements[index].ID != id {
		return RuntimeTerrainImprovementDefinition{}, false
	}
	return runtime.TerrainImprovements[index], true
}

func (runtime RuntimeRules) Unit(id string) (RuntimeUnitDefinition, bool) {
	index := sort.Search(len(runtime.Units), func(index int) bool { return runtime.Units[index].ID >= id })
	if index == len(runtime.Units) || runtime.Units[index].ID != id {
		return RuntimeUnitDefinition{}, false
	}
	return runtime.Units[index], true
}

type RuntimeStarter struct {
	TypeID            string                 `json:"type_id,omitempty"`
	Domain            string                 `json:"domain"`
	Movement          int                    `json:"movement"`
	Attack            int                    `json:"attack"`
	Defense           int                    `json:"defense"`
	Ranged            int                    `json:"ranged,omitempty"`
	Strength          int                    `json:"strength"`
	Armor             int                    `json:"armor"`
	Firepower         int                    `json:"firepower,omitempty"`
	Health            int                    `json:"health"`
	MaxFuel           int                    `json:"max_fuel,omitempty"`
	ZOC               bool                   `json:"zoc"`
	TransportCapacity int                    `json:"transport_capacity,omitempty"`
	CargoDomains      []string               `json:"cargo_domains,omitempty"`
	VisibilityRange   int                    `json:"visibility_range,omitempty"`
	Effects           []RuntimeStarterEffect `json:"effects"`
}

type RuntimeStarterEffect struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}

type RuntimeVictory struct {
	ID string `json:"id"`
}

type RuntimeCalendar struct {
	FinalTurn int `json:"final_turn"`
}

type RuntimeConstants struct {
	CityGrowthCoefficient          int `json:"city_growth_coefficient,omitempty"`
	UnconvertedIndulgenceGold      int `json:"unconverted_indulgence_gold,omitempty"`
	UnconvertedIndulgenceHappiness int `json:"unconverted_indulgence_happiness,omitempty"`
	ConvertedIndulgenceGold        int `json:"converted_indulgence_gold,omitempty"`
	ConvertedIndulgenceHappiness   int `json:"converted_indulgence_happiness,omitempty"`
	OtherFaithIndulgenceGold       int `json:"other_faith_indulgence_gold,omitempty"`
	OtherFaithIndulgenceHappiness  int `json:"other_faith_indulgence_happiness,omitempty"`
	ClericConversionPercent        int `json:"cleric_conversion_percent,omitempty"`
	TelevangelistConversionPercent int `json:"televangelist_conversion_percent,omitempty"`
	EmancipationUnhappinessTurns   int `json:"emancipation_unhappiness_turns,omitempty"`
	EmancipationUnhappinessAmount  int `json:"emancipation_unhappiness_amount,omitempty"`
	NukePopulationPercent          int `json:"nuke_population_percent,omitempty"`
	NukeKillTiles                  int `json:"nuke_kill_tiles,omitempty"`
	PollutionCausedByNuke          int `json:"pollution_caused_by_nuke,omitempty"`
}

type RuntimeGaia struct {
	ActivationTurns int    `json:"activation_turns"`
	BuildingID      string `json:"building_id,omitempty"`
	WonderID        string `json:"wonder_id,omitempty"`
}

type RuntimeProduction struct {
	ID                string                 `json:"id"`
	Kind              string                 `json:"kind"`
	Cost              int                    `json:"cost,omitempty"`
	Upkeep            int                    `json:"upkeep,omitempty"`
	Prerequisites     []string               `json:"prerequisites,omitempty"`
	RequiredBuildings []string               `json:"required_buildings,omitempty"`
	CantBuildInSea    bool                   `json:"cant_build_in_sea,omitempty"`
	CoastalOnly       bool                   `json:"coastal_only,omitempty"`
	ObsoleteBy        []string               `json:"obsolete_by,omitempty"`
	Effects           []RuntimeStarterEffect `json:"effects,omitempty"`
}

type RuntimeAI struct {
	UnitBuildLists     []RuntimeAIList     `json:"unit_build_lists,omitempty"`
	BuildingBuildLists []RuntimeAIList     `json:"building_build_lists,omitempty"`
	WonderBuildLists   []RuntimeAIList     `json:"wonder_build_lists,omitempty"`
	AdvanceLists       []RuntimeAIList     `json:"advance_lists,omitempty"`
	ImprovementLists   []RuntimeAIList     `json:"improvement_lists,omitempty"`
	Strategies         []RuntimeAIStrategy `json:"strategies,omitempty"`
}

// RuntimeAIStrategy carries retail AI economy/social slider preferences that can be
// used from projections without exposing hidden world state.
type RuntimeAIStrategy struct {
	ID                 string                   `json:"id"`
	SciencePercent     int                      `json:"science_percent,omitempty"`
	PublicWorksPercent int                      `json:"public_works_percent,omitempty"`
	PublicWorksReserve int                      `json:"public_works_reserve,omitempty"`
	SliderElements     []RuntimeAISliderElement `json:"slider_elements,omitempty"`
}

type RuntimeAISliderElement struct {
	Kind  string `json:"kind"`
	Delta int    `json:"delta,omitempty"`
}

type RuntimeAIList struct {
	ID    string   `json:"id"`
	Items []string `json:"items,omitempty"`
}

type RuntimeEffect struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

func (runtime RuntimeRules) VictoryEvaluatorIDs() []string {
	ids := make([]string, len(runtime.Victory))
	for i, evaluator := range runtime.Victory {
		ids[i] = evaluator.ID
	}
	return ids
}

func containsVictoryEvaluator(evaluators []VictoryEvaluator, id string) bool {
	for _, evaluator := range evaluators {
		if evaluator.ID == id {
			return true
		}
	}
	return false
}

func (runtime RuntimeRules) ProductionDefinition(id string) (RuntimeProduction, bool) {
	index := sort.Search(len(runtime.Production), func(index int) bool { return runtime.Production[index].ID >= id })
	if index == len(runtime.Production) || runtime.Production[index].ID != id {
		return RuntimeProduction{}, false
	}
	return runtime.Production[index], true
}

func (runtime RuntimeRules) ProductionIDs() string {
	ids := make([]string, len(runtime.Production))
	for i, production := range runtime.Production {
		ids[i] = production.ID
	}
	return strings.Join(ids, ",")
}

func (runtime RuntimeRules) EffectIDs() string {
	ids := make([]string, len(runtime.Effects))
	for i, effect := range runtime.Effects {
		ids[i] = effect.ID
	}
	return strings.Join(ids, ",")
}

func CanonicalDefinitions(definitions []byte) ([]byte, error) {
	if len(definitions) == 0 || !json.Valid(definitions) {
		return nil, errors.New("rules definitions are malformed")
	}
	var value any
	if err := json.Unmarshal(definitions, &value); err != nil {
		return nil, errors.New("rules definitions are malformed")
	}
	return json.Marshal(value)
}

func CanonicalRuntimeDefinitions(runtime RuntimeRules) ([]byte, string, error) {
	runtime.ContentHash = ""
	sortRuntimeProjection(&runtime)
	encoded, err := json.Marshal(runtime)
	if err != nil {
		return nil, "", err
	}
	canonical, err := CanonicalDefinitions(encoded)
	if err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(canonical)
	return canonical, hex.EncodeToString(digest[:]), nil
}

func CompileRuntime(definitions []byte, contentHash string) (RuntimeRules, error) {
	canonical, err := CanonicalDefinitions(definitions)
	if err != nil {
		return RuntimeRules{}, err
	}
	digest := sha256.Sum256(canonical)
	if hex.EncodeToString(digest[:]) != contentHash {
		return RuntimeRules{}, errors.New("rules content hash mismatch")
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(definitions, &top); err != nil {
		return RuntimeRules{}, errors.New("rules definitions are malformed")
	}
	if _, hasRecords := top["records"]; !hasRecords {
		return compileRuntimeProjectionDefinitions(definitions, contentHash)
	}
	var definition struct {
		CompatibilityVersion       string                 `json:"compatibility_version"`
		Manifest                   SupportedRulesManifest `json:"manifest"`
		Records                    []Record               `json:"records"`
		EnabledVictoryEvaluatorIDs []string               `json:"enabled_victory_evaluator_ids,omitempty"`
	}
	if err := json.Unmarshal(definitions, &definition); err != nil {
		return RuntimeRules{}, errors.New("rules definitions are malformed")
	}
	if definition.CompatibilityVersion != CTP2CompatibilityVersion || definition.Manifest.CompatibilityVersion != CTP2CompatibilityVersion {
		return RuntimeRules{}, errors.New("rules compatibility version is unsupported")
	}
	if err := definition.Manifest.Validate(); err != nil {
		return RuntimeRules{}, fmt.Errorf("rules manifest is invalid: %w", err)
	}
	if len(definition.Records) == 0 {
		return RuntimeRules{}, errors.New("rules definitions contain no records")
	}

	runtime := RuntimeRules{
		CompatibilityVersion: definition.CompatibilityVersion,
		ContentHash:          contentHash,
		Calendar:             RuntimeCalendar{FinalTurn: calendarFinalTurn},
	}
	enabledVictories := make(map[string]struct{}, len(definition.Manifest.VictoryEvaluators))
	for _, evaluator := range definition.Manifest.VictoryEvaluators {
		enabledVictories[evaluator.ID] = struct{}{}
	}
	if len(definition.EnabledVictoryEvaluatorIDs) > 0 {
		enabledVictories = make(map[string]struct{}, len(definition.EnabledVictoryEvaluatorIDs))
		for _, evaluatorID := range definition.EnabledVictoryEvaluatorIDs {
			if _, selected := enabledVictories[evaluatorID]; selected {
				return RuntimeRules{}, fmt.Errorf("victory evaluator %q is duplicated", evaluatorID)
			}
			if !containsVictoryEvaluator(definition.Manifest.VictoryEvaluators, evaluatorID) {
				return RuntimeRules{}, fmt.Errorf("victory evaluator %q is unsupported", evaluatorID)
			}
			enabledVictories[evaluatorID] = struct{}{}
		}
	}
	for evaluatorID := range enabledVictories {
		runtime.Victory = append(runtime.Victory, RuntimeVictory{ID: evaluatorID})
	}

	seen := make(map[string]struct{}, len(definition.Records))
	for _, record := range definition.Records {
		if record.ID == "" || record.Kind == "" || record.Tags == nil {
			return RuntimeRules{}, errors.New("rules definition record is malformed")
		}
		if _, ok := seen[record.ID]; ok {
			return RuntimeRules{}, fmt.Errorf("rules definition record %q is duplicated", record.ID)
		}
		seen[record.ID] = struct{}{}
		if err := compileRecord(&runtime, record); err != nil {
			return RuntimeRules{}, err
		}
	}
	if runtime.Gaia.ActivationTurns < 1 {
		return RuntimeRules{}, errors.New("rules definitions require one Gaia endgame record")
	}
	if !containsRuntimeProduction(runtime.Production, "UNIT_SYNTHETIC") {
		return RuntimeRules{}, errors.New("rules definitions require the synthetic starter unit")
	}
	if runtime.Movement.Configured && (!runtime.Movement.PolicyConfigured || len(runtime.Movement.Terrains) == 0) {
		return RuntimeRules{}, errors.New("rules definitions require complete synthetic movement definitions")
	}
	sort.Slice(runtime.Advances, func(i, j int) bool { return runtime.Advances[i].ID < runtime.Advances[j].ID })
	sort.Slice(runtime.Feats, func(i, j int) bool { return runtime.Feats[i].ID < runtime.Feats[j].ID })
	if err := validateRuntimeAdvances(runtime.Advances); err != nil {
		return RuntimeRules{}, err
	}
	for _, government := range runtime.Governments {
		if government.EnableAdvanceID != "" {
			if _, ok := runtime.Advance(government.EnableAdvanceID); !ok {
				return RuntimeRules{}, fmt.Errorf("rules definition government %q references missing advance %q", government.ID, government.EnableAdvanceID)
			}
		}
	}
	sort.Slice(runtime.TerrainImprovements, func(i, j int) bool { return runtime.TerrainImprovements[i].ID < runtime.TerrainImprovements[j].ID })
	for _, improvement := range runtime.TerrainImprovements {
		if improvement.RequiredAdvanceID != "" {
			if _, ok := runtime.Advance(improvement.RequiredAdvanceID); !ok {
				return RuntimeRules{}, fmt.Errorf("rules definition terrain improvement %q references missing advance %q", improvement.ID, improvement.RequiredAdvanceID)
			}
		}
	}
	sort.Slice(runtime.PollutionSources, func(i, j int) bool { return runtime.PollutionSources[i].ID < runtime.PollutionSources[j].ID })
	sort.Slice(runtime.PollutionTriggers, func(i, j int) bool { return runtime.PollutionTriggers[i].ID < runtime.PollutionTriggers[j].ID })
	sort.Slice(runtime.Governments, func(i, j int) bool { return runtime.Governments[i].ID < runtime.Governments[j].ID })
	sort.Slice(runtime.Units, func(i, j int) bool { return runtime.Units[i].ID < runtime.Units[j].ID })
	starterUnit, ok := runtime.Unit("UNIT_SYNTHETIC")
	if !ok {
		return RuntimeRules{}, errors.New("rules definitions require the synthetic starter unit")
	}
	runtime.Starter = RuntimeStarter{
		TypeID: starterUnit.ID, Domain: starterUnit.Domain, Movement: starterUnit.Movement, Attack: starterUnit.Attack, Defense: starterUnit.Defense, Ranged: starterUnit.Ranged, Strength: starterUnit.Strength, Armor: starterUnit.Armor, Firepower: starterUnit.Firepower, Health: starterUnit.Health, MaxFuel: starterUnit.MaxFuel, ZOC: starterUnit.ZOC, TransportCapacity: starterUnit.TransportCapacity, CargoDomains: append([]string(nil), starterUnit.CargoDomains...), VisibilityRange: starterUnit.VisibilityRange,
		Effects: append([]RuntimeStarterEffect(nil), starterUnit.Effects...),
	}
	sort.Slice(runtime.Victory, func(i, j int) bool { return runtime.Victory[i].ID < runtime.Victory[j].ID })
	sort.Slice(runtime.Production, func(i, j int) bool { return runtime.Production[i].ID < runtime.Production[j].ID })
	sort.Slice(runtime.Effects, func(i, j int) bool { return runtime.Effects[i].ID < runtime.Effects[j].ID })
	sort.Slice(runtime.Movement.Terrains, func(i, j int) bool { return runtime.Movement.Terrains[i].ID < runtime.Movement.Terrains[j].ID })
	return runtime, nil
}

func compileRuntimeProjectionDefinitions(definitions []byte, contentHash string) (RuntimeRules, error) {
	var runtime RuntimeRules
	if err := json.Unmarshal(definitions, &runtime); err != nil {
		return RuntimeRules{}, errors.New("rules definitions are malformed")
	}
	if runtime.CompatibilityVersion != CTP2CompatibilityVersion {
		return RuntimeRules{}, errors.New("rules compatibility version is unsupported")
	}
	runtime.ContentHash = contentHash
	sortRuntimeProjection(&runtime)
	if len(runtime.Production) == 0 {
		return RuntimeRules{}, errors.New("runtime definitions contain no production records")
	}
	if len(runtime.Units) == 0 {
		return RuntimeRules{}, errors.New("runtime definitions contain no unit records")
	}
	if runtime.Movement.Configured {
		runtime.Movement.PolicyConfigured = len(runtime.Movement.Terrains) > 0
		if !runtime.Movement.PolicyConfigured {
			return RuntimeRules{}, errors.New("runtime definitions require movement terrains")
		}
	}
	if runtime.Starter.Domain == "" || runtime.Starter.Movement < 1 || runtime.Starter.Health < 1 {
		starter, ok := runtime.Unit("UNIT_SETTLER")
		if !ok {
			return RuntimeRules{}, errors.New("runtime definitions require a starter unit")
		}
		runtime.Starter = RuntimeStarter{TypeID: starter.ID, Domain: starter.Domain, Movement: starter.Movement, Attack: starter.Attack, Defense: starter.Defense, Ranged: starter.Ranged, Strength: starter.Strength, Armor: starter.Armor, Firepower: starter.Firepower, Health: starter.Health, MaxFuel: starter.MaxFuel, ZOC: starter.ZOC, TransportCapacity: starter.TransportCapacity, CargoDomains: append([]string(nil), starter.CargoDomains...), VisibilityRange: starter.VisibilityRange, Effects: append([]RuntimeStarterEffect(nil), starter.Effects...)}
	}
	if err := validateRuntimeProjection(runtime); err != nil {
		return RuntimeRules{}, err
	}
	return runtime, nil
}

func validateRuntimeProjection(runtime RuntimeRules) error {
	if err := validateRuntimeProjectionVictory(runtime.Victory); err != nil {
		return err
	}
	seenProduction := make(map[string]struct{}, len(runtime.Production))
	for _, production := range runtime.Production {
		if production.ID == "" || production.Kind == "" {
			return errors.New("runtime production record is malformed")
		}
		if _, exists := seenProduction[production.ID]; exists {
			return fmt.Errorf("runtime production record %q is duplicated", production.ID)
		}
		seenProduction[production.ID] = struct{}{}
	}
	seenUnits := make(map[string]struct{}, len(runtime.Units))
	for _, unit := range runtime.Units {
		if unit.ID == "" || unit.Domain == "" || unit.Movement < 1 || unit.Health < 1 {
			return fmt.Errorf("runtime unit %q is malformed", unit.ID)
		}
		if _, exists := seenUnits[unit.ID]; exists {
			return fmt.Errorf("runtime unit %q is duplicated", unit.ID)
		}
		seenUnits[unit.ID] = struct{}{}
	}
	if err := validateRuntimeAdvanceProjection(runtime.Advances); err != nil {
		return err
	}
	for _, government := range runtime.Governments {
		if government.EnableAdvanceID != "" {
			if _, ok := runtime.Advance(government.EnableAdvanceID); !ok {
				return fmt.Errorf("runtime government %q references missing advance %q", government.ID, government.EnableAdvanceID)
			}
		}
	}
	for _, improvement := range runtime.TerrainImprovements {
		if improvement.RequiredAdvanceID != "" {
			if _, ok := runtime.Advance(improvement.RequiredAdvanceID); !ok {
				return fmt.Errorf("runtime terrain improvement %q references missing advance %q", improvement.ID, improvement.RequiredAdvanceID)
			}
		}
	}
	return nil
}

func validateRuntimeProjectionVictory(victories []RuntimeVictory) error {
	supported := make(map[string]struct{})
	for _, evaluator := range SupportedManifest().VictoryEvaluators {
		supported[evaluator.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(victories))
	for _, victory := range victories {
		if victory.ID == "" {
			return errors.New("runtime victory evaluator is malformed")
		}
		if _, ok := supported[victory.ID]; !ok {
			return fmt.Errorf("runtime victory evaluator %q is unsupported", victory.ID)
		}
		if _, exists := seen[victory.ID]; exists {
			return fmt.Errorf("runtime victory evaluator %q is duplicated", victory.ID)
		}
		seen[victory.ID] = struct{}{}
	}
	return nil
}

func validateRuntimeAdvanceProjection(advances []RuntimeAdvanceDefinition) error {
	seen := make(map[string]struct{}, len(advances))
	for _, advance := range advances {
		if advance.ID == "" {
			return errors.New("runtime advance record is malformed")
		}
		if _, exists := seen[advance.ID]; exists {
			return fmt.Errorf("runtime advance %q is duplicated", advance.ID)
		}
		seen[advance.ID] = struct{}{}
	}
	for _, advance := range advances {
		for _, prerequisite := range advance.Prerequisites {
			if prerequisite == advance.ID {
				continue
			}
			if _, ok := seen[prerequisite]; !ok {
				return fmt.Errorf("runtime advance %q references missing prerequisite %q", advance.ID, prerequisite)
			}
		}
	}
	return nil
}

// ServerOwnedSyntheticRuleset is the fixed original synthetic catalog used when
// a server starts a gameplay aggregate. It is intentionally independent of
// client-visible setup options and contains no retail data.
func ServerOwnedSyntheticRuleset(victoryEvaluatorIDs ...string) (Ruleset, error) {
	manifest := SupportedManifest()
	records := []Record{
		{Kind: "endgame", ID: "ENDGAME_GAIA_VECTOR", Tags: map[string]string{"TurnsToActivate": "1"}},
		{Kind: "endgame", ID: "ENDGAME_UNION_VECTOR", Tags: map[string]string{"TurnsToActivate": "0"}},
		{Kind: "building", ID: "IMPROVE_SYNTHETIC", Tags: map[string]string{"Description": "IMPROVE_SYNTHETIC"}},
		{Kind: "terrain", ID: "forest", Tags: map[string]string{"MovementCost": "2", "MovementDomains": "amphibious,land"}},
		{Kind: "terrain", ID: "ocean", Tags: map[string]string{"MovementCost": "1", "MovementDomains": "amphibious,sea"}},
		{Kind: "terrain", ID: "plains", Tags: map[string]string{"MovementCost": "1", "MovementDomains": "amphibious,land"}},
		{Kind: "movement", ID: "MOVEMENT_SYNTHETIC", Tags: map[string]string{"OpenAccess": "true", "ZOCStopsPath": "true"}},
		{Kind: "terrain_improvement", ID: "TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD", Tags: map[string]string{"Description": "TERRAIN_IMPROVEMENT_SYNTHETIC_ROAD", "BuildTurns": "1", "PublicWorksCost": "1", "MovementCostDelta": "-1"}},
		{Kind: "terrain_improvement", ID: "TERRAIN_IMPROVEMENT_SYNTHETIC_FARM", Tags: map[string]string{"Description": "TERRAIN_IMPROVEMENT_SYNTHETIC_FARM", "BuildTurns": "1", "PublicWorksCost": "1", "YieldFood": "1"}},
		{Kind: "terrain_improvement", ID: "TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP", Tags: map[string]string{"Description": "TERRAIN_IMPROVEMENT_SYNTHETIC_CLEANUP", "BuildTurns": "1", "PublicWorksCost": "1", "PollutionDelta": "-2"}},
		{Kind: "pollution_source", ID: "POLLUTION_SOURCE_SYNTHETIC_FACTORY", Tags: map[string]string{"Description": "POLLUTION_SOURCE_SYNTHETIC_FACTORY", "PollutionPoints": "2", "EffectScope": "city"}},
		{Kind: "government", ID: "GOVERNMENT_SYNTHETIC_REPUBLIC", Tags: map[string]string{"Description": "GOVERNMENT_SYNTHETIC_REPUBLIC", "EnableAdvanceID": "ADVANCE_SYNTHETIC_ROOT", "ResearchBonus": "1"}},
		{Kind: "advance", ID: "ADVANCE_SYNTHETIC_ROOT", Tags: map[string]string{"Description": "ADVANCE_SYNTHETIC_ROOT", "Cost": "2"}},
		{Kind: "advance", ID: "ADVANCE_SYNTHETIC_UNLOCK", Tags: map[string]string{"Description": "ADVANCE_SYNTHETIC_UNLOCK", "Cost": "2", "PrerequisiteAdvanceIDs": "ADVANCE_SYNTHETIC_ROOT"}},
		{Kind: "unit", ID: "UNIT_SYNTHETIC", Tags: map[string]string{"Description": "UNIT_SYNTHETIC", "Role": "starter", "Domain": "land", "MaxMovePoints": "2", "Attack": "4", "Defense": "4", "Strength": "4", "Armor": "2", "Health": "10", "ZOC": "true", "VisibilityRange": "2", "Effects": "settler,special_embassy,special_steal:100,special_capture:0,special_terraform:1,special_terraform_turns:1"}},
		{Kind: "wonder", ID: "WONDER_SYNTHETIC", Tags: map[string]string{"Description": "WONDER_SYNTHETIC"}},
	}
	enabled := append([]string(nil), victoryEvaluatorIDs...)
	sort.Strings(enabled)
	canonical, err := json.Marshal(struct {
		CompatibilityVersion       string                 `json:"compatibility_version"`
		Manifest                   SupportedRulesManifest `json:"manifest"`
		Records                    []Record               `json:"records"`
		EnabledVictoryEvaluatorIDs []string               `json:"enabled_victory_evaluator_ids,omitempty"`
	}{CompatibilityVersion: CTP2CompatibilityVersion, Manifest: manifest, Records: records, EnabledVictoryEvaluatorIDs: enabled})
	if err != nil {
		return Ruleset{}, err
	}
	canonical, err = CanonicalDefinitions(canonical)
	if err != nil {
		return Ruleset{}, err
	}
	digest := sha256.Sum256(canonical)
	ruleset := Ruleset{CompatibilityVersion: CTP2CompatibilityVersion, Manifest: manifest, Records: records, CanonicalJSON: canonical, ContentHash: hex.EncodeToString(digest[:])}
	if _, err := CompileRuntime(ruleset.CanonicalJSON, ruleset.ContentHash); err != nil {
		return Ruleset{}, err
	}
	return ruleset, nil
}

func compileRecord(runtime *RuntimeRules, record Record) error {
	switch record.Kind {
	case "advance":
		advance, err := compileAdvanceDefinition(record)
		if err != nil {
			return err
		}
		runtime.Advances = append(runtime.Advances, advance)
	case "unit":
		unit, err := compileUnitDefinition(record)
		if err != nil {
			return err
		}
		runtime.Units = append(runtime.Units, unit)
		runtime.Production = append(runtime.Production, RuntimeProduction{ID: record.ID, Kind: "unit", Cost: unit.ProductionCost, Upkeep: unit.Upkeep, Prerequisites: append([]string(nil), unit.Prerequisites...), ObsoleteBy: append([]string(nil), unit.ObsoleteBy...)})
	case "building":
		production, err := compileProductionDefinition(record, "building")
		if err != nil {
			return err
		}
		runtime.Production = append(runtime.Production, production)
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: record.ID, Kind: "building"})
	case "wonder":
		production, err := compileProductionDefinition(record, "wonder")
		if err != nil {
			return err
		}
		runtime.Production = append(runtime.Production, production)
		runtime.Effects = append(runtime.Effects, RuntimeEffect{ID: record.ID, Kind: "wonder"})
	case "terrain_improvement":
		improvement, err := compileTerrainImprovementDefinition(record)
		if err != nil {
			return err
		}
		runtime.TerrainImprovements = append(runtime.TerrainImprovements, improvement)
	case "pollution_source":
		source, err := compilePollutionSourceDefinition(record)
		if err != nil {
			return err
		}
		runtime.PollutionSources = append(runtime.PollutionSources, source)
	case "government":
		government, err := compileGovernmentDefinition(record)
		if err != nil {
			return err
		}
		runtime.Governments = append(runtime.Governments, government)
	case "terrain":
		if _, ok := record.Tags["MovementCost"]; !ok {
			return nil
		}
		cost, err := positiveInteger(record, "MovementCost")
		if err != nil {
			return err
		}
		domains, err := movementDomains(record)
		if err != nil {
			return err
		}
		defenseBonusPercent, err := optionalInteger(record, "DefenseBonusPercent", 0)
		if err != nil {
			return err
		}
		food, err := optionalInteger(record, "Food", 0)
		if err != nil {
			return err
		}
		production, err := optionalInteger(record, "Production", 0)
		if err != nil {
			return err
		}
		gold, err := optionalInteger(record, "Gold", 0)
		if err != nil {
			return err
		}
		runtime.Movement.Configured = true
		runtime.Movement.Terrains = append(runtime.Movement.Terrains, RuntimeTerrain{ID: record.ID, Cost: cost, Domains: domains, Food: food, Production: production, Gold: gold, DefenseBonusPercent: defenseBonusPercent})
	case "movement":
		if record.ID != "MOVEMENT_SYNTHETIC" || runtime.Movement.PolicyConfigured {
			return fmt.Errorf("rules definition record %q has invalid movement policy", record.ID)
		}
		openAccess, err := boolean(record, "OpenAccess")
		if err != nil {
			return err
		}
		zocStopsPath, err := boolean(record, "ZOCStopsPath")
		if err != nil {
			return err
		}
		runtime.Movement.Configured = true
		runtime.Movement.PolicyConfigured = true
		runtime.Movement.OpenAccess = openAccess
		runtime.Movement.ZOCStopsPath = zocStopsPath
	case "endgame":
		turns, err := nonNegativeInteger(record, "TurnsToActivate")
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(record.ID), "gaia") {
			if runtime.Gaia.ActivationTurns != 0 {
				return errors.New("rules definitions contain multiple Gaia endgame records")
			}
			if turns < 1 {
				return fmt.Errorf("rules definition record %q has invalid TurnsToActivate", record.ID)
			}
			runtime.Gaia = RuntimeGaia{ActivationTurns: turns}
		}
	default:
		return fmt.Errorf("rules definition record %q has unknown kind %q", record.ID, record.Kind)
	}
	return nil
}

func containsRuntimeProduction(production []RuntimeProduction, id string) bool {
	for _, item := range production {
		if item.ID == id {
			return true
		}
	}
	return false
}

func compileProductionDefinition(record Record, kind string) (RuntimeProduction, error) {
	cost, err := optionalNonNegativeInteger(record, "ProductionCost", 0)
	if err != nil {
		return RuntimeProduction{}, err
	}
	upkeep, err := optionalNonNegativeInteger(record, "Upkeep", 0)
	if err != nil {
		return RuntimeProduction{}, err
	}
	prerequisites, err := optionalIDs(record, "EnableAdvanceID")
	if err != nil {
		return RuntimeProduction{}, err
	}
	if len(prerequisites) == 0 {
		prerequisites, err = optionalIDs(record, "Prerequisite")
		if err != nil {
			return RuntimeProduction{}, err
		}
	}
	requiredBuildings, err := optionalIDs(record, "RequiredBuildings")
	if err != nil {
		return RuntimeProduction{}, err
	}
	cantBuildInSea := record.Tags["CantBuildInSea"] == "true"
	coastalOnly := record.Tags["CoastalOnly"] == "true"
	obsoleteBy, err := optionalIDs(record, "ObsoleteBy")
	if err != nil {
		return RuntimeProduction{}, err
	}
	return RuntimeProduction{ID: record.ID, Kind: kind, Cost: cost, Upkeep: upkeep, Prerequisites: prerequisites, RequiredBuildings: requiredBuildings, CantBuildInSea: cantBuildInSea, CoastalOnly: coastalOnly, ObsoleteBy: obsoleteBy}, nil
}

func compileAdvanceDefinition(record Record) (RuntimeAdvanceDefinition, error) {
	cost, err := positiveInteger(record, "Cost")
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	prerequisites, err := optionalIDs(record, "PrerequisiteAdvanceIDs")
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	triggerFeatID := record.Tags["TriggerFeatID"]
	allowsPlantNuke, err := optionalBool(record, "AllowsPlantNuke", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	capitalization, err := optionalBool(record, "EnablesCapitalization", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	infrastructure, err := optionalBool(record, "EnablesInfrastructure", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	transform, err := optionalBool(record, "EnablesTransform", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	roads, err := optionalBool(record, "EnablesRoads", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	railroads, err := optionalBool(record, "EnablesRailroads", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	maglev, err := optionalBool(record, "EnablesMagLev", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	tunnels, err := optionalBool(record, "EnablesTunnels", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	deepOcean, err := optionalBool(record, "EnablesDeepOcean", false)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	pollutionSizeModifier, err := optionalInteger(record, "PollutionSizeModifier", 0)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	pollutionProductionModifier, err := optionalInteger(record, "PollutionProductionModifier", 0)
	if err != nil {
		return RuntimeAdvanceDefinition{}, err
	}
	return RuntimeAdvanceDefinition{ID: record.ID, Cost: cost, Prerequisites: prerequisites, TriggerFeatID: triggerFeatID, AllowsPlantNuke: allowsPlantNuke, EnablesCapitalization: capitalization, EnablesInfrastructure: infrastructure, EnablesTransform: transform, EnablesRoads: roads, EnablesRailroads: railroads, EnablesMagLev: maglev, EnablesTunnels: tunnels, EnablesDeepOcean: deepOcean, PollutionSizeModifier: pollutionSizeModifier, PollutionProductionModifier: pollutionProductionModifier}, nil
}

func compileTerrainImprovementDefinition(record Record) (RuntimeTerrainImprovementDefinition, error) {
	buildTurns, err := positiveInteger(record, "BuildTurns")
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	publicWorksCost, err := positiveInteger(record, "PublicWorksCost")
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	movementCostDelta, err := optionalInteger(record, "MovementCostDelta", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	yieldFood, err := optionalInteger(record, "YieldFood", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	yieldProduction, err := optionalInteger(record, "YieldProduction", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	yieldGold, err := optionalInteger(record, "YieldGold", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	pollutionDelta, err := optionalInteger(record, "PollutionDelta", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	defenseBonusPercent, err := optionalInteger(record, "DefenseBonusPercent", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	healRatePercent, err := optionalInteger(record, "HealRatePercent", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	visionRange, err := optionalInteger(record, "VisionRange", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	radarRange, err := optionalInteger(record, "RadarRange", 0)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	airport, err := optionalBool(record, "Airport", false)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	radar, err := optionalBool(record, "Radar", false)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	listeningPost, err := optionalBool(record, "ListeningPost", false)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	canUpgrade, err := optionalBool(record, "CanUpgrade", false)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	endgame, err := optionalBool(record, "Endgame", false)
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	targetTerrain := record.Tags["TargetTerrain"]
	allowedTerrains, err := optionalIDs(record, "AllowedTerrains")
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	cantBuildOn, err := optionalIDs(record, "CantBuildOn")
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	excludes, err := optionalIDs(record, "Excludes")
	if err != nil {
		return RuntimeTerrainImprovementDefinition{}, err
	}
	if movementCostDelta == 0 && yieldFood == 0 && yieldProduction == 0 && yieldGold == 0 && pollutionDelta == 0 && targetTerrain == "" && defenseBonusPercent == 0 && healRatePercent == 0 && visionRange == 0 && radarRange == 0 && !airport && !radar && !listeningPost && !canUpgrade && !endgame {
		return RuntimeTerrainImprovementDefinition{}, fmt.Errorf("rules definition record %q has no terrain improvement effect", record.ID)
	}
	return RuntimeTerrainImprovementDefinition{ID: record.ID, Class: record.Tags["Class"], BuildTurns: buildTurns, PublicWorksCost: publicWorksCost, MovementCostDelta: movementCostDelta, YieldFood: yieldFood, YieldProduction: yieldProduction, YieldGold: yieldGold, PollutionDelta: pollutionDelta, TargetTerrain: targetTerrain, RequiredAdvanceID: record.Tags["RequiredAdvanceID"], AllowedTerrains: allowedTerrains, DefenseBonusPercent: defenseBonusPercent, HealRatePercent: healRatePercent, VisionRange: visionRange, RadarRange: radarRange, Airport: airport, Radar: radar, ListeningPost: listeningPost, CanUpgrade: canUpgrade, Endgame: endgame, CantBuildOn: cantBuildOn, Excludes: excludes}, nil
}

func compilePollutionSourceDefinition(record Record) (RuntimePollutionSourceDefinition, error) {
	points, err := positiveInteger(record, "PollutionPoints")
	if err != nil {
		return RuntimePollutionSourceDefinition{}, err
	}
	scope := record.Tags["EffectScope"]
	if scope != "city" && scope != "nation" && scope != "global" {
		return RuntimePollutionSourceDefinition{}, fmt.Errorf("rules definition record %q has invalid EffectScope", record.ID)
	}
	return RuntimePollutionSourceDefinition{ID: record.ID, PollutionPoints: points, EffectScope: scope}, nil
}

func compileGovernmentDefinition(record Record) (RuntimeGovernmentDefinition, error) {
	researchBonus, err := optionalInteger(record, "ResearchBonus", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	publicWorksBonus, err := optionalInteger(record, "PublicWorksBonus", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	infrastructureCoefficient, err := optionalFloat(record, "InfrastructureCoefficient", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	capitalizationCoefficient, err := optionalFloat(record, "CapitalizationCoefficient", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	researchPercent, err := optionalInteger(record, "ResearchPercent", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	foodPercent, err := optionalInteger(record, "FoodPercent", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	productionPercent, err := optionalInteger(record, "ProductionPercent", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	goldPercent, err := optionalInteger(record, "GoldPercent", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	pollutionPercent, err := optionalInteger(record, "PollutionPercent", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	pollutionUnhappyCoef, err := optionalFloat(record, "PollutionUnhappyCoef", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	crimeCoef, err := optionalFloat(record, "CrimeCoef", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	crimeOffset, err := optionalInteger(record, "CrimeOffset", 0)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	definition, err := compileGovernmentSocialFields(record)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	readiness, err := compileGovernmentReadinessFields(record)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	definition.SupportCoef = readiness.SupportCoef
	definition.DefenseCoef = readiness.DefenseCoef
	definition.ProfessionalUnits = readiness.ProfessionalUnits
	definition.TurnsToNewReadiness = readiness.TurnsToNewReadiness
	definition.ReadyPeaceCoef = readiness.ReadyPeaceCoef
	definition.ReadyPeaceHP = readiness.ReadyPeaceHP
	definition.ReadyAlertCoef = readiness.ReadyAlertCoef
	definition.ReadyAlertHP = readiness.ReadyAlertHP
	definition.ReadyWarCoef = readiness.ReadyWarCoef
	definition.ReadyWarHP = readiness.ReadyWarHP
	empire, err := compileGovernmentEmpireFields(record)
	if err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	definition.WarDiscontentMaxUnits = empire.WarDiscontentMaxUnits
	definition.WarDiscontentPerUnit = empire.WarDiscontentPerUnit
	definition.ConquestDistress = empire.ConquestDistress
	definition.ConquestDistressDecay = empire.ConquestDistressDecay
	definition.AtHomeRadius = empire.AtHomeRadius
	definition.OverseasCoef = empire.OverseasCoef
	definition.OverseasDefeatDecay = empire.OverseasDefeatDecay
	definition.OverseasDefeatCoef = empire.OverseasDefeatCoef
	definition.HomeDefeatDecay = empire.HomeDefeatDecay
	definition.HomeDefeatCoef = empire.HomeDefeatCoef
	definition.EmpireDistanceScale = empire.EmpireDistanceScale
	definition.MinEmpireDistance = empire.MinEmpireDistance
	definition.MaxEmpireDistance = empire.MaxEmpireDistance
	definition.TooManyCitiesThreshold = empire.TooManyCitiesThreshold
	definition.TooManyCitiesCoefficient = empire.TooManyCitiesCoefficient
	if researchBonus < 0 || publicWorksBonus < 0 {
		return RuntimeGovernmentDefinition{}, fmt.Errorf("rules definition record %q has invalid government bonus", record.ID)
	}
	definition.ID = record.ID
	definition.EnableAdvanceID = record.Tags["EnableAdvanceID"]
	definition.ResearchBonus = researchBonus
	definition.ResearchPercent = researchPercent
	definition.PublicWorksBonus = publicWorksBonus
	definition.InfrastructureCoefficient = infrastructureCoefficient
	definition.CapitalizationCoefficient = capitalizationCoefficient
	definition.FoodPercent = foodPercent
	definition.ProductionPercent = productionPercent
	definition.GoldPercent = goldPercent
	definition.PollutionPercent = pollutionPercent
	definition.PollutionUnhappyCoef = pollutionUnhappyCoef
	definition.CrimeCoef = crimeCoef
	definition.CrimeOffset = crimeOffset
	return definition, nil
}

func compileGovernmentSocialFields(record Record) (RuntimeGovernmentDefinition, error) {
	var definition RuntimeGovernmentDefinition
	var err error
	if definition.WorkdayExpectation, err = optionalInteger(record, "WorkdayExpectation", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.PositiveWorkdayCoef, err = optionalFloat(record, "PositiveWorkdayCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.NegativeWorkdayCoef, err = optionalFloat(record, "NegativeWorkdayCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.WagesExpectation, err = optionalInteger(record, "WagesExpectation", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.PositiveWagesCoef, err = optionalFloat(record, "PositiveWagesCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.NegativeWagesCoef, err = optionalFloat(record, "NegativeWagesCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.RationsExpectation, err = optionalInteger(record, "RationsExpectation", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.PositiveRationsCoef, err = optionalFloat(record, "PositiveRationsCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.NegativeRationsCoef, err = optionalFloat(record, "NegativeRationsCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	return definition, nil
}

func compileGovernmentReadinessFields(record Record) (RuntimeGovernmentDefinition, error) {
	var definition RuntimeGovernmentDefinition
	var err error
	if definition.SupportCoef, err = optionalFloat(record, "SupportCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.DefenseCoef, err = optionalFloat(record, "DefenseCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ProfessionalUnits, err = optionalInteger(record, "ProfessionalUnits", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.TurnsToNewReadiness, err = optionalInteger(record, "TurnsToNewReadiness", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ReadyPeaceCoef, err = optionalFloat(record, "ReadyPeaceCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ReadyPeaceHP, err = optionalFloat(record, "ReadyPeaceHP", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ReadyAlertCoef, err = optionalFloat(record, "ReadyAlertCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ReadyAlertHP, err = optionalFloat(record, "ReadyAlertHP", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ReadyWarCoef, err = optionalFloat(record, "ReadyWarCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ReadyWarHP, err = optionalFloat(record, "ReadyWarHP", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	return definition, nil
}

func compileGovernmentEmpireFields(record Record) (RuntimeGovernmentDefinition, error) {
	var definition RuntimeGovernmentDefinition
	var err error
	if definition.WarDiscontentMaxUnits, err = optionalInteger(record, "WarDiscontentMaxUnits", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.WarDiscontentPerUnit, err = optionalFloat(record, "WarDiscontentPerUnit", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ConquestDistress, err = optionalFloat(record, "ConquestDistress", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.ConquestDistressDecay, err = optionalFloat(record, "ConquestDistressDecay", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.AtHomeRadius, err = optionalInteger(record, "AtHomeRadius", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.OverseasCoef, err = optionalFloat(record, "OverseasCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.OverseasDefeatDecay, err = optionalFloat(record, "OverseasDefeatDecay", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.OverseasDefeatCoef, err = optionalFloat(record, "OverseasDefeatCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.HomeDefeatDecay, err = optionalFloat(record, "HomeDefeatDecay", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.HomeDefeatCoef, err = optionalFloat(record, "HomeDefeatCoef", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.EmpireDistanceScale, err = optionalFloat(record, "EmpireDistanceScale", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.MinEmpireDistance, err = optionalInteger(record, "MinEmpireDistance", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.MaxEmpireDistance, err = optionalInteger(record, "MaxEmpireDistance", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.TooManyCitiesThreshold, err = optionalInteger(record, "TooManyCitiesThreshold", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	if definition.TooManyCitiesCoefficient, err = optionalFloat(record, "TooManyCitiesCoefficient", 0); err != nil {
		return RuntimeGovernmentDefinition{}, err
	}
	return definition, nil
}

func validateRuntimeAdvances(advances []RuntimeAdvanceDefinition) error {
	byID := make(map[string]RuntimeAdvanceDefinition, len(advances))
	for _, advance := range advances {
		if advance.ID == "" || advance.Cost < 1 {
			return fmt.Errorf("rules definition advance %q is malformed", advance.ID)
		}
		if _, duplicate := byID[advance.ID]; duplicate {
			return fmt.Errorf("rules definition advance %q is duplicated", advance.ID)
		}
		byID[advance.ID] = advance
	}
	for _, advance := range advances {
		for _, prerequisite := range advance.Prerequisites {
			if prerequisite == advance.ID {
				return fmt.Errorf("rules definition advance %q has cyclic prerequisites", advance.ID)
			}
			if _, ok := byID[prerequisite]; !ok {
				return fmt.Errorf("rules definition advance %q references missing prerequisite %q", advance.ID, prerequisite)
			}
		}
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("rules definition advance %q has cyclic prerequisites", id)
		}
		visiting[id] = true
		advance := byID[id]
		for _, prerequisite := range advance.Prerequisites {
			if err := visit(prerequisite); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for _, advance := range advances {
		if err := visit(advance.ID); err != nil {
			return err
		}
	}
	return nil
}

func compileUnitDefinition(record Record) (RuntimeUnitDefinition, error) {
	movement, err := positiveInteger(record, "MaxMovePoints")
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	domain := optionalString(record, "Domain", "land")
	if domain != "land" && domain != "sea" && domain != "amphibious" && domain != "air" && domain != "space" {
		return RuntimeUnitDefinition{}, fmt.Errorf("rules definition record %q has invalid Domain", record.ID)
	}
	attack, err := optionalPositiveInteger(record, "Attack", 4)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	defense, err := optionalPositiveInteger(record, "Defense", 4)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	ranged, err := optionalNonNegativeInteger(record, "Ranged", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	if ranged == 0 {
		ranged, err = optionalNonNegativeInteger(record, "ZBRangeAttack", 0)
		if err != nil {
			return RuntimeUnitDefinition{}, err
		}
	}
	bombardRange, err := optionalNonNegativeInteger(record, "BombardRange", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	bombRounds, err := optionalNonNegativeInteger(record, "BombRounds", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	strength, err := optionalPositiveInteger(record, "Strength", attack)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	armor, err := optionalPositiveInteger(record, "Armor", 2)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	firepower, err := optionalPositiveInteger(record, "Firepower", 1)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	health, err := optionalPositiveInteger(record, "Health", 10)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	maxFuel, err := optionalNonNegativeInteger(record, "MaxFuel", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	transportCapacity, err := optionalNonNegativeInteger(record, "TransportCapacity", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	productionCost, err := optionalNonNegativeInteger(record, "ProductionCost", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	upkeep, err := optionalNonNegativeInteger(record, "Upkeep", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	visibilityRange, err := optionalNonNegativeInteger(record, "VisibilityRange", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	zoc, err := optionalBool(record, "ZOC", true)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	cargoDomains, err := optionalCSV(record, "CargoDomains", []string{"land", "sea", "amphibious"})
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	if transportCapacity == 0 && len(cargoDomains) > 0 {
		return RuntimeUnitDefinition{}, fmt.Errorf("rules definition record %q has CargoDomains without TransportCapacity", record.ID)
	}
	prerequisites, err := optionalIDs(record, "Prerequisites")
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	obsoleteBy, err := optionalIDs(record, "ObsoleteBy")
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	requiredGovernmentID := record.Tags["RequiredGovernmentID"]
	needsCityGoodID := record.Tags["NeedsCityGoodID"]
	settleCityTypeID := record.Tags["SettleCityTypeID"]
	cityGrowthCoefficientPercent, err := optionalNonNegativeInteger(record, "CityGrowthCoefficientPercent", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	effects, err := unitEffects(record)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	shieldHunger, err := optionalNonNegativeInteger(record, "ShieldHunger", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	foodHunger, err := optionalNonNegativeInteger(record, "FoodHunger", 0)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	needsNoSupport, err := optionalBool(record, "NeedsNoSupport", false)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	isSpecialForces, err := optionalBool(record, "IsSpecialForces", false)
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	upgradeTo, err := optionalIDs(record, "UpgradeTo")
	if err != nil {
		return RuntimeUnitDefinition{}, err
	}
	return RuntimeUnitDefinition{ID: record.ID, Role: record.Tags["Role"], Domain: domain, Movement: movement, Attack: attack, Defense: defense, Ranged: ranged, BombardRange: bombardRange, BombRounds: bombRounds, Strength: strength, Armor: armor, Firepower: firepower, Health: health, MaxFuel: maxFuel, ZOC: zoc, TransportCapacity: transportCapacity, CargoDomains: cargoDomains, ProductionCost: productionCost, Upkeep: upkeep, ShieldHunger: shieldHunger, FoodHunger: foodHunger, NeedsNoSupport: needsNoSupport, IsSpecialForces: isSpecialForces, Prerequisites: prerequisites, RequiredGovernmentID: requiredGovernmentID, NeedsCityGoodID: needsCityGoodID, SettleCityTypeID: settleCityTypeID, CityGrowthCoefficientPercent: cityGrowthCoefficientPercent, ObsoleteBy: obsoleteBy, UpgradeTo: upgradeTo, VisibilityRange: visibilityRange, Effects: effects}, nil
}

func unitEffects(record Record) ([]RuntimeStarterEffect, error) {
	value := record.Tags["Effects"]
	if value == "" {
		return nil, nil
	}
	items := strings.Split(value, ",")
	effects := make([]RuntimeStarterEffect, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("rules definition record %q has invalid Effects", record.ID)
		}
		kind, rawValue, hasValue := strings.Cut(item, ":")
		kind = strings.TrimSpace(kind)
		if !validUnitEffectKind(kind) {
			return nil, fmt.Errorf("rules definition record %q has invalid Effects", record.ID)
		}
		if _, duplicate := seen[kind]; duplicate {
			return nil, fmt.Errorf("rules definition record %q has duplicate Effects", record.ID)
		}
		seen[kind] = struct{}{}
		effect := RuntimeStarterEffect{ID: effectID(kind), Kind: kind}
		if hasValue {
			parsed, err := strconv.Atoi(strings.TrimSpace(rawValue))
			if err != nil || parsed < 0 {
				return nil, fmt.Errorf("rules definition record %q has invalid Effects", record.ID)
			}
			effect.Value = parsed
		}
		effects = append(effects, effect)
	}
	sort.Slice(effects, func(i, j int) bool { return effects[i].Kind < effects[j].Kind })
	return effects, nil
}

func validUnitEffectKind(kind string) bool {
	if validSpecialEliteChanceRuntimeKind(kind) {
		return true
	}
	switch kind {
	case "settler", "special_embassy", "special_soothsaying", "special_steal", "special_capture", "special_terraform", "special_terraform_turns", "can_be_sued", "can_be_expelled", "stealth", "can_see_stealth", "exerts_martial_law", "max_science_percent", "special_convert_city_death_chance", "special_create_franchise_death_chance", "special_investigate_city_death_chance", "special_plant_nuke_death_chance", "special_plant_nuke_escape_chance", "special_bio_terror_death_chance", "special_nano_terror_death_chance", "special_plague_death_chance", "special_assassinate_ruler_death_chance", "special_slave_raid_death_chance", "special_settler_slave_raid_death_chance", "special_underground_railway_death_chance":
		return true
	default:
		return false
	}
}

func validSpecialEliteChanceRuntimeKind(kind string) bool {
	if !strings.HasSuffix(kind, "_elite_chance") {
		return false
	}
	switch strings.TrimSuffix(kind, "_elite_chance") {
	case "special_convert_city", "special_create_franchise", "special_investigate_city", "special_plant_nuke", "special_bio_terror", "special_nano_terror", "special_plague", "special_assassinate_ruler", "special_slave_raid", "special_settler_slave_raid", "special_underground_railway":
		return true
	default:
		return false
	}
}

func effectID(kind string) string {
	return strings.ReplaceAll(strings.TrimPrefix(kind, "special_"), "_", "-")
}

func requirePositiveInteger(record Record, tag string) error {
	_, err := positiveInteger(record, tag)
	return err
}

func positiveInteger(record Record, tag string) (int, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return 0, fmt.Errorf("rules definition record %q is missing %s", record.ID, tag)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}

func nonNegativeInteger(record Record, tag string) (int, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return 0, fmt.Errorf("rules definition record %q is missing %s", record.ID, tag)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}

func movementDomains(record Record) ([]string, error) {
	value, ok := record.Tags["MovementDomains"]
	if !ok || value == "" {
		return nil, fmt.Errorf("rules definition record %q is missing MovementDomains", record.ID)
	}
	domains := strings.Split(value, ",")
	for _, domain := range domains {
		if domain != "amphibious" && domain != "land" && domain != "sea" {
			return nil, fmt.Errorf("rules definition record %q has invalid MovementDomains", record.ID)
		}
	}
	sort.Strings(domains)
	for index := 1; index < len(domains); index++ {
		if domains[index] == domains[index-1] {
			return nil, fmt.Errorf("rules definition record %q has duplicate MovementDomains", record.ID)
		}
	}
	return domains, nil
}

func boolean(record Record, tag string) (bool, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return false, fmt.Errorf("rules definition record %q is missing %s", record.ID, tag)
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}

func optionalString(record Record, tag, fallback string) string {
	if value := record.Tags[tag]; value != "" {
		return value
	}
	return fallback
}

func optionalPositiveInteger(record Record, tag string, fallback int) (int, error) {
	if _, ok := record.Tags[tag]; !ok {
		return fallback, nil
	}
	return positiveInteger(record, tag)
}

func optionalNonNegativeInteger(record Record, tag string, fallback int) (int, error) {
	if _, ok := record.Tags[tag]; !ok {
		return fallback, nil
	}
	return nonNegativeInteger(record, tag)
}

func optionalInteger(record Record, tag string, fallback int) (int, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}

func optionalFloat(record Record, tag string, fallback float64) (float64, error) {
	value, ok := record.Tags[tag]
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
	}
	return parsed, nil
}

func optionalBool(record Record, tag string, fallback bool) (bool, error) {
	if _, ok := record.Tags[tag]; !ok {
		return fallback, nil
	}
	return boolean(record, tag)
}

func optionalCSV(record Record, tag string, allowed []string) ([]string, error) {
	value := record.Tags[tag]
	if value == "" {
		return nil, nil
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, item := range allowed {
		allowedSet[item] = struct{}{}
	}
	items := strings.Split(value, ",")
	for index, item := range items {
		items[index] = strings.TrimSpace(item)
		if items[index] == "" {
			return nil, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
		}
		if _, ok := allowedSet[items[index]]; !ok {
			return nil, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
		}
	}
	sort.Strings(items)
	for index := 1; index < len(items); index++ {
		if items[index] == items[index-1] {
			return nil, fmt.Errorf("rules definition record %q has duplicate %s", record.ID, tag)
		}
	}
	return items, nil
}

func optionalIDs(record Record, tag string) ([]string, error) {
	value := record.Tags[tag]
	if value == "" || value == "none" {
		return nil, nil
	}
	ids := strings.Split(value, ",")
	for index, id := range ids {
		ids[index] = strings.TrimSpace(id)
		if ids[index] == "" {
			return nil, fmt.Errorf("rules definition record %q has invalid %s", record.ID, tag)
		}
	}
	sort.Strings(ids)
	for index := 1; index < len(ids); index++ {
		if ids[index] == ids[index-1] {
			return nil, fmt.Errorf("rules definition record %q has duplicate %s", record.ID, tag)
		}
	}
	return ids, nil
}
