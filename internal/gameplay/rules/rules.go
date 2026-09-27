// Package rules imports a bounded, source-provenanced subset of user-owned CTP2 data files.
package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	CTP2CompatibilityVersion = "ctp2-supported-v2"

	ErrorDuplicateID            = "duplicate_id"
	ErrorMalformedValue         = "malformed_value"
	ErrorUnsupportedRequiredTag = "unsupported_required_tag"
	ErrorUnsupportedMechanic    = "unsupported_mechanic"
	ErrorInvalidDirectory       = "invalid_directory"
	ErrorMissingRequiredTag     = "missing_required_tag"
	ErrorMissingVictoryVector   = "missing_victory_vector"
)

// CatalogFile defines a supported upstream CTP2 source filename and record prefix.
type CatalogFile struct {
	Filename string `json:"filename"`
	Kind     string `json:"kind"`
	Prefix   string `json:"prefix"`
}

// CatalogFiles is the data-driven production file map. Synthetic fixtures may use the
// same basename with a .ctp2synthetic suffix; retail filenames are never tracked.
var CatalogFiles = []CatalogFile{
	{Filename: "Units.txt", Kind: "unit", Prefix: "UNIT_"},
	{Filename: "Advance.txt", Kind: "advance", Prefix: "ADVANCE_"},
	{Filename: "buildings.txt", Kind: "building", Prefix: "IMPROVE_"},
	{Filename: "Wonder.txt", Kind: "wonder", Prefix: "WONDER_"},
	{Filename: "terrain.txt", Kind: "terrain", Prefix: "TERRAIN_"},
	{Filename: "TerrainImprovement.txt", Kind: "terrain_improvement", Prefix: "TERRAIN_IMPROVEMENT_"},
	{Filename: "PollutionSource.txt", Kind: "pollution_source", Prefix: "POLLUTION_SOURCE_"},
	{Filename: "Government.txt", Kind: "government", Prefix: "GOVERNMENT_"},
	{Filename: "Const.txt", Kind: "rules", Prefix: "CONST_"},
	{Filename: "EndGameObjects.txt", Kind: "endgame", Prefix: "ENDGAME_"},
	{Filename: "Civilisation.txt", Kind: "civilization", Prefix: "CIVILIZATION_"},
}

type ImportConfig struct {
	Directory string
	Manifest  SupportedRulesManifest
}

type ImportError struct {
	Code     string
	Filename string
	Line     int
	RecordID string
	Detail   string
}

func (e *ImportError) Error() string {
	return fmt.Sprintf("%s: %s:%d (%s): %s", e.Code, e.Filename, e.Line, e.RecordID, e.Detail)
}

type Record struct {
	Kind string            `json:"kind"`
	ID   string            `json:"id"`
	Tags map[string]string `json:"tags"`
}

type Ruleset struct {
	CompatibilityVersion string                 `json:"compatibility_version"`
	Manifest             SupportedRulesManifest `json:"manifest"`
	Records              []Record               `json:"records"`
	CanonicalJSON        []byte                 `json:"-"`
	ContentHash          string                 `json:"content_hash"`
}

type SourceProvenance struct {
	Repository string `json:"repository"`
	Commit     string `json:"commit"`
	Path       string `json:"path"`
	Symbol     string `json:"symbol"`
	Purpose    string `json:"purpose"`
}

type VictoryEvaluator struct {
	ID                string           `json:"id"`
	SyntheticVectorID string           `json:"synthetic_vector_id"`
	Source            SourceProvenance `json:"source"`
}

type VictoryVector struct {
	ID          string `json:"id"`
	EvaluatorID string `json:"evaluator_id"`
	Original    bool   `json:"original"`
}

type SpecialActionEvaluator struct {
	ID                string           `json:"id"`
	SyntheticVectorID string           `json:"synthetic_vector_id"`
	Source            SourceProvenance `json:"source"`
}

type SpecialActionVector struct {
	ID          string `json:"id"`
	EvaluatorID string `json:"evaluator_id"`
	Original    bool   `json:"original"`
}

type RejectedVictoryEvaluator struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

const (
	mechanicStatusIncludedNow        = "included-now"
	mechanicStatusDeferred           = "deferred"
	mechanicStatusExplicitlyExcluded = "explicitly-excluded"
)

type MechanicReadiness struct {
	Family     string `json:"family"`
	Status     string `json:"status"`
	ContractID string `json:"contract_id"`
	Importer   string `json:"importer,omitempty"`
	Runtime    string `json:"runtime,omitempty"`
	Engine     string `json:"engine,omitempty"`
	API        string `json:"api,omitempty"`
	AI         string `json:"ai,omitempty"`
	Notes      string `json:"notes"`
}

type SyntheticRecordContract struct {
	ID                string           `json:"id"`
	Family            string           `json:"family"`
	RecordKind        string           `json:"record_kind"`
	SyntheticIDPrefix string           `json:"synthetic_id_prefix"`
	RequiredTags      []string         `json:"required_tags"`
	Purpose           string           `json:"purpose"`
	Source            SourceProvenance `json:"source"`
}

// SupportedRulesManifest inventories the bounded CTP2-style mechanic slice. Only
// entries in VictoryEvaluators are enabled; rejected entries fail closed.
type SupportedRulesManifest struct {
	CompatibilityVersion      string                     `json:"compatibility_version"`
	Modes                     []string                   `json:"modes"`
	SupportedMechanics        []string                   `json:"supported_mechanics"`
	RejectedTags              []string                   `json:"rejected_tags"`
	OpaqueTags                []string                   `json:"opaque_tags"`
	VictoryEvaluators         []VictoryEvaluator         `json:"victory_evaluators"`
	VictoryVectors            []VictoryVector            `json:"victory_vectors"`
	RejectedVictoryEvaluators []RejectedVictoryEvaluator `json:"rejected_victory_evaluators"`
	SpecialActionEvaluators   []SpecialActionEvaluator   `json:"special_action_evaluators"`
	SpecialActionVectors      []SpecialActionVector      `json:"special_action_vectors"`
	MechanicMatrix            []MechanicReadiness        `json:"mechanic_matrix"`
	SyntheticContracts        []SyntheticRecordContract  `json:"synthetic_contracts"`
}

const upstreamRepository = "https://github.com/civctp2/civctp2"
const upstreamCommit = "0668dd199c2519ff8ab7d2871e7c7a89e21c8e52"

func SupportedManifest() SupportedRulesManifest {
	return SupportedRulesManifest{
		CompatibilityVersion: CTP2CompatibilityVersion,
		Modes:                []string{"sequential", "long_turn", "simultaneous", "hybrid"},
		SupportedMechanics: []string{
			"catalog_import", "endgame_objects", "unit_movement", "unit_zoc",
			"unit_disband", "unit_fortify", "unit_embark", "unit_disembark",
			"army_composition", "army_capacity", "transport_capacity", "combat_resolution",
			"city_founding", "city_worked_tiles", "production_queue", "production_upkeep",
			"building_city_effects", "wonder_nation_effects", "wonder_global_uniqueness",
			"effect_yield", "effect_combat", "effect_vision", "effect_production", "effect_victory",
			"special_embassy", "special_soothsaying", "special_steal_technology", "terrain_transformation",
		},
		RejectedTags: []string{"SLIC", "CargoData"},
		OpaqueTags:   []string{"DefaultIcon", "DefaultSprite", "SoundSelect1"},
		VictoryEvaluators: []VictoryEvaluator{
			{ID: "bloodlust_conquest", SyntheticVectorID: "bloodlust_conquest_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/Player.cpp", Symbol: "Player::GameOverCheck", Purpose: "conquest world-win evaluator"}},
			{ID: "diplomatic_world_union", SyntheticVectorID: "world_union_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/ai/diplomacy/diplomat.cpp", Symbol: "Diplomat::ComputeAllianceSize", Purpose: "World Union alliance-coverage evaluator"}},
			{ID: "score_2300", SyntheticVectorID: "score_2300_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/utility/TurnCnt.cpp", Symbol: "TurnCnt::GetYear", Purpose: "end-of-game-year highest-score evaluator"}},
			{ID: "science_gaia_controller", SyntheticVectorID: "gaia_controller_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/gaiacontroller.cpp", Symbol: "GaiaController::CanStartCountdown", Purpose: "Gaia endgame requirements evaluator"}},
		},
		VictoryVectors: []VictoryVector{
			{ID: "bloodlust_conquest_vector", EvaluatorID: "bloodlust_conquest", Original: true},
			{ID: "gaia_controller_vector", EvaluatorID: "science_gaia_controller", Original: true},
			{ID: "score_2300_vector", EvaluatorID: "score_2300", Original: true},
			{ID: "world_union_vector", EvaluatorID: "diplomatic_world_union", Original: true},
		},
		SpecialActionEvaluators: []SpecialActionEvaluator{
			{ID: "embassy", SyntheticVectorID: "embassy_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/ArmyData.cpp", Symbol: "ArmyData::EstablishEmbassy", Purpose: "adjacent foreign city embassy authorization"}},
			{ID: "soothsaying", SyntheticVectorID: "soothsaying_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/ArmyData.cpp", Symbol: "ArmyData::Soothsay", Purpose: "city special-action effect"}},
			{ID: "steal_technology", SyntheticVectorID: "steal_technology_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/UnitData.cpp", Symbol: "UnitData::StealTechnology", Purpose: "eligible advance selection, success and capture"}},
			{ID: "terraform", SyntheticVectorID: "terraform_vector", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/Player.cpp", Symbol: "Player::CanTerraform", Purpose: "terrain transformation prerequisites"}},
		},
		SpecialActionVectors: []SpecialActionVector{{ID: "embassy_vector", EvaluatorID: "embassy", Original: true}, {ID: "soothsaying_vector", EvaluatorID: "soothsaying", Original: true}, {ID: "steal_technology_vector", EvaluatorID: "steal_technology", Original: true}, {ID: "terraform_vector", EvaluatorID: "terraform", Original: true}},
		MechanicMatrix: []MechanicReadiness{
			{Family: "units", Status: mechanicStatusIncludedNow, ContractID: "contract-units", Importer: "CatalogFiles Units.txt -> parseCTP2File(unit)", Runtime: "compileRecord(kind=unit) and starter-unit presence gate", Engine: "production and unit command slice already consumes runtime production/unit records", API: "existing gameplay setup and projection contracts expose unit production and unit actions", AI: "basic legal-use slice exists through current deterministic command legality; richer tactics remain deferred", Notes: "Included now as a bounded synthetic production and action family, not retail-stat parity."},
			{Family: "advances", Status: mechanicStatusIncludedNow, ContractID: "contract-advances", Importer: "CatalogFiles Advance.txt -> parseCTP2File(advance)", Runtime: "compileRecord(kind=advance) builds sorted prerequisite/cost runtime definitions", Engine: "set-research and round completion accrue progress and grant advances deterministically", API: "public projection exposes owned advances, progress, and available research only", AI: "basic legal-use slice can select visible available research; strategy remains deferred", Notes: "Included now as bounded synthetic prerequisite/unlock semantics, not retail tree parity."},
			{Family: "buildings", Status: mechanicStatusIncludedNow, ContractID: "contract-buildings", Importer: "CatalogFiles buildings.txt -> parseCTP2File(building)", Runtime: "compileRecord(kind=building) produces production and effect inventory", Engine: "city production and upkeep slice can already bind generic building production/effect records", API: "existing gameplay projections already expose production choices and city-state consequences", AI: "basic legal-use slice can choose bounded synthetic production; optimization remains deferred", Notes: "Included now only for generic city production/effect wiring and bounded synthetic upkeep semantics."},
			{Family: "wonders", Status: mechanicStatusIncludedNow, ContractID: "contract-wonders", Importer: "CatalogFiles Wonder.txt -> parseCTP2File(wonder)", Runtime: "compileRecord(kind=wonder) produces production and effect inventory", Engine: "city production and wonder uniqueness hooks can already consume generic wonder/effect records", API: "existing gameplay projections already expose production choices and resulting wonder state", AI: "basic legal-use slice can pick bounded synthetic wonders; strategic wonder timing remains deferred", Notes: "Included now for bounded production/effect and uniqueness contracts only."},
			{Family: "terrain", Status: mechanicStatusIncludedNow, ContractID: "contract-terrain", Importer: "CatalogFiles terrain.txt -> parseCTP2File(terrain)", Runtime: "compileRecord(kind=terrain) and movement policy synthesize terrain movement tables", Engine: "movement, visibility, and world traversal already consume bounded terrain movement definitions", API: "existing world and gameplay projections expose terrain-backed movement legality", AI: "basic legal-use slice can issue commands against current movement legality; advanced path heuristics remain deferred", Notes: "Included now for bounded movement/yield anchors, not full terrain ecology parity."},
			{Family: "terrain-improvements", Status: mechanicStatusIncludedNow, ContractID: "contract-terrain-improvements", Importer: "CatalogFiles TerrainImprovement.txt -> parseCTP2File(terrain_improvement)", Runtime: "compileRecord(kind=terrain_improvement) builds public-works cost, build turns, movement, yield, and pollution deltas", Engine: "build-improvement queues deterministic overlays that complete through round recalculation", API: "gameplay commands and projections expose visible terrain/pollution state", AI: "basic legal-use can prefer projected public-works actions; optimization remains deferred", Notes: "Included for roads/farms/cleanup style synthetic overlays, not full retail improvement catalog."},
			{Family: "governments", Status: mechanicStatusIncludedNow, ContractID: "contract-governments", Importer: "CatalogFiles Government.txt -> parseCTP2File(government)", Runtime: "compileRecord(kind=government) validates bounded modifiers and advance unlocks", Engine: "set-government mutates nation government and bounded effects", API: "public projection exposes only viewer government", AI: "basic legal-use can pick available governments from public state", Notes: "Included for bounded synthetic switching/modifiers, not full retail tax/readiness parity."},
			{Family: "trade-routes-agreements", Status: mechanicStatusIncludedNow, ContractID: "contract-trade-routes-agreements", Importer: "agreement kind is command-level bounded rather than separate retail route catalog", Runtime: "relationship state has deterministic treaty/trade booleans", Engine: "contact/proposal/response commands mutate canonical relations", API: "public projection filters viewer relations/proposals", AI: "basic legal-use can respond/propose visible agreements", Notes: "Included for diplomatic agreements/trade flags, not route pathfinding parity."},
			{Family: "pollution-sources", Status: mechanicStatusIncludedNow, ContractID: "contract-pollution-sources", Importer: "CatalogFiles PollutionSource.txt -> parseCTP2File(pollution_source)", Runtime: "compileRecord(kind=pollution_source) validates bounded point/scope definitions", Engine: "terrain cleanup overlays mutate tile pollution deterministically", API: "visible tile pollution is projected; hidden pollution remains redacted", AI: "basic legal-use can avoid hidden pollution and choose cleanup only from projection", Notes: "Included for bounded tile pollution counters and cleanup, not global disaster parity."},
			{Family: "special-actions", Status: mechanicStatusIncludedNow, ContractID: "contract-special-actions", Importer: "SupportedManifest special-action evaluator/vector inventory pins bounded action families", Runtime: "manifest validation plus starter action effects provide executable bounded action hooks", Engine: "engine command slice already names embassy, steal, and terraform style actions as bounded mechanics", API: "existing gameplay command/projection contracts can expose bounded special-action legality and outcomes", AI: "basic legal-use slice can stay bounded to current command legality; richer targeting and sequencing remain deferred", Notes: "Included now for bounded action-vector contracts and starter hooks only."},
			{Family: "client-presentation-assets", Status: mechanicStatusExplicitlyExcluded, Notes: "Explicitly excluded from Task 4 and this server slice because graphics, audio, UI presentation, and localization stay on the client side and must not gain fake synthetic readiness layers or contracts."},
		},
		SyntheticContracts: []SyntheticRecordContract{
			{ID: "contract-units", Family: "units", RecordKind: "unit", SyntheticIDPrefix: "UNIT_SYNTHETIC_", RequiredTags: []string{"Description", "MaxMovePoints", "Role"}, Purpose: "Synthetic units must at minimum name their role and movement budget so runtime production and legal action slices stay deterministic.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/Player.cpp", Symbol: "Player::CreateUnit", Purpose: "unit creation branches on unit record properties and movement domain constraints"}},
			{ID: "contract-advances", Family: "advances", RecordKind: "advance", SyntheticIDPrefix: "ADVANCE_SYNTHETIC_", RequiredTags: []string{"Description", "Cost", "PrerequisiteAdvanceIDs"}, Purpose: "Synthetic advances must describe prerequisite and cost structure without copying retail trees so future unlock logic can fail closed.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/Advances.cpp", Symbol: "Advances::ResetCanResearch", Purpose: "advance availability is recalculated from prerequisites and gating conditions"}},
			{ID: "contract-buildings", Family: "buildings", RecordKind: "building", SyntheticIDPrefix: "IMPROVE_SYNTHETIC_", RequiredTags: []string{"Description", "Upkeep", "EffectVectorID"}, Purpose: "Synthetic buildings must declare upkeep and a bounded effect vector so city production semantics remain deterministic.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/buildingutil.cpp", Symbol: "buildingutil_GetTotalUpkeep", Purpose: "built improvements aggregate upkeep and effect-bearing record properties"}},
			{ID: "contract-wonders", Family: "wonders", RecordKind: "wonder", SyntheticIDPrefix: "WONDER_SYNTHETIC_", RequiredTags: []string{"Description", "EffectVectorID", "UniquenessScope"}, Purpose: "Synthetic wonders must declare bounded global uniqueness and effect semantics without copying retail names or movies.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/WonderTracker.cpp", Symbol: "WonderTracker::AddBuilt", Purpose: "wonder ownership is tracked globally and must stay unique"}},
			{ID: "contract-terrain", Family: "terrain", RecordKind: "terrain", SyntheticIDPrefix: "TERRAIN_SYNTHETIC_", RequiredTags: []string{"Description", "MovementCost", "MovementDomains"}, Purpose: "Synthetic terrain must declare movement semantics and a stable identifier so traversal rules and future yield work stay deterministic.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/world/WrlImprove.cpp", Symbol: "Terrain::GetFood", Purpose: "terrain records expose bounded per-tile yields alongside movement/environment semantics"}},
			{ID: "contract-terrain-improvements", Family: "terrain-improvements", RecordKind: "terrain_improvement", SyntheticIDPrefix: "TERRAIN_IMPROVEMENT_SYNTHETIC_", RequiredTags: []string{"Description", "BuildTurns", "PublicWorksCost"}, Purpose: "Synthetic terrain improvements must separate placement/build-time semantics from terrain base records for public-works and transformation tasks.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/world/WrlImprove.cpp", Symbol: "World::InsertImprovement", Purpose: "terrain improvements are inserted separately from base terrain and need their own contract"}},
			{ID: "contract-governments", Family: "governments", RecordKind: "government", SyntheticIDPrefix: "GOVERNMENT_SYNTHETIC_", RequiredTags: []string{"Description", "EnableAdvanceID", "TaxRateCap"}, Purpose: "Synthetic governments must bind to explicit unlocks and top-level policy caps so future government switching is deterministic.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/Player.cpp", Symbol: "Player::InitPlayer", Purpose: "government choice depends on unlocked advances and drives tax/readiness initialization"}},
			{ID: "contract-trade-routes-agreements", Family: "trade-routes-agreements", RecordKind: "trade_agreement", SyntheticIDPrefix: "TRADE_AGREEMENT_SYNTHETIC_", RequiredTags: []string{"Description", "RouteClass", "DurationTurns"}, Purpose: "Synthetic trade and agreement records must declare route/agreement class and bounded duration so pathing and diplomatic lifecycle can be validated later.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/TradeRouteData.cpp", Symbol: "TradeRouteData::GeneratePath", Purpose: "trade routes have dedicated path generation and cost state that require explicit synthetic contracts"}},
			{ID: "contract-pollution-sources", Family: "pollution-sources", RecordKind: "pollution_source", SyntheticIDPrefix: "POLLUTION_SOURCE_SYNTHETIC_", RequiredTags: []string{"Description", "PollutionPoints", "EffectScope"}, Purpose: "Synthetic pollution sources must declare bounded contribution and effect scope so city/world accumulation can be tested without retail disaster tables.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/Pollution.cpp", Symbol: "Pollution::GetGlobalPollutionLevel", Purpose: "global pollution aggregates player and wonder-driven deltas from multiple sources"}},
			{ID: "contract-special-actions", Family: "special-actions", RecordKind: "special_action", SyntheticIDPrefix: "SPECIAL_ACTION_SYNTHETIC_", RequiredTags: []string{"Description", "ActionType", "VectorID"}, Purpose: "Synthetic special actions must bind a stable action type to a bounded evaluator vector so legality and projections fail closed.", Source: SourceProvenance{Repository: upstreamRepository, Commit: upstreamCommit, Path: "ctp2_code/gs/gameobj/ArmyData.cpp", Symbol: "ArmyData::EstablishEmbassy", Purpose: "special actions depend on unit-side command execution and adjacency legality"}},
		},
	}
}

func (m SupportedRulesManifest) Validate() error {
	if m.CompatibilityVersion != CTP2CompatibilityVersion || !sameStrings(m.Modes, []string{"hybrid", "long_turn", "sequential", "simultaneous"}) {
		return errors.New("invalid supported rules manifest version or modes")
	}
	vectors := make(map[string]VictoryVector, len(m.VictoryVectors))
	for _, vector := range m.VictoryVectors {
		if vector.ID == "" || vector.EvaluatorID == "" || !vector.Original {
			return fmt.Errorf("%s: invalid synthetic vector", ErrorMissingVictoryVector)
		}
		if _, exists := vectors[vector.ID]; exists {
			return fmt.Errorf("%s: duplicate vector %s", ErrorMissingVictoryVector, vector.ID)
		}
		vectors[vector.ID] = vector
	}
	seen := map[string]bool{}
	for _, evaluator := range m.VictoryEvaluators {
		if evaluator.ID == "" || seen[evaluator.ID] || evaluator.Source.Repository != upstreamRepository || evaluator.Source.Commit != upstreamCommit || evaluator.Source.Path == "" || evaluator.Source.Symbol == "" {
			return fmt.Errorf("invalid evaluator provenance %q", evaluator.ID)
		}
		vector, ok := vectors[evaluator.SyntheticVectorID]
		if !ok || vector.EvaluatorID != evaluator.ID {
			return fmt.Errorf("%s: %s", ErrorMissingVictoryVector, evaluator.ID)
		}
		seen[evaluator.ID] = true
	}
	for _, evaluator := range m.RejectedVictoryEvaluators {
		if evaluator.ID == "" || evaluator.Reason == "" || seen[evaluator.ID] {
			return fmt.Errorf("invalid rejected victory inventory %q", evaluator.ID)
		}
		seen[evaluator.ID] = true
	}
	if !sameStrings(mapKeys(seen), []string{"bloodlust_conquest", "diplomatic_world_union", "score_2300", "science_gaia_controller"}) {
		return errors.New("baseline victory inventory is incomplete")
	}
	if err := validateSpecialActionVectors(m.SpecialActionEvaluators, m.SpecialActionVectors); err != nil {
		return err
	}
	contracts, err := validateSyntheticContracts(m.SyntheticContracts)
	if err != nil {
		return err
	}
	return validateMechanicMatrix(m.MechanicMatrix, contracts)
}

// Import reads only the explicitly supplied user-owned directory. It never scans installation paths.
func validateSpecialActionVectors(evaluators []SpecialActionEvaluator, vectors []SpecialActionVector) error {
	byID := make(map[string]SpecialActionVector, len(vectors))
	for _, vector := range vectors {
		if vector.ID == "" || vector.EvaluatorID == "" || !vector.Original {
			return fmt.Errorf("%s: invalid special-action vector", ErrorMissingVictoryVector)
		}
		if _, exists := byID[vector.ID]; exists {
			return fmt.Errorf("%s: duplicate special-action vector %s", ErrorMissingVictoryVector, vector.ID)
		}
		byID[vector.ID] = vector
	}
	seen := make(map[string]bool, len(evaluators))
	for _, evaluator := range evaluators {
		if evaluator.ID == "" || seen[evaluator.ID] || evaluator.Source.Repository != upstreamRepository || evaluator.Source.Commit != upstreamCommit || evaluator.Source.Path == "" || evaluator.Source.Symbol == "" {
			return fmt.Errorf("invalid special-action evaluator %q", evaluator.ID)
		}
		vector, ok := byID[evaluator.SyntheticVectorID]
		if !ok || vector.EvaluatorID != evaluator.ID {
			return fmt.Errorf("%s: %s", ErrorMissingVictoryVector, evaluator.ID)
		}
		seen[evaluator.ID] = true
	}
	if !sameStrings(mapKeys(seen), []string{"embassy", "soothsaying", "steal_technology", "terraform"}) {
		return errors.New("special-action inventory is incomplete")
	}
	return nil
}

func requiredMechanicFamilies() []string {
	return []string{
		"units",
		"advances",
		"buildings",
		"wonders",
		"terrain",
		"terrain-improvements",
		"governments",
		"trade-routes-agreements",
		"pollution-sources",
		"special-actions",
	}
}

func validMechanicStatus(status string) bool {
	switch status {
	case mechanicStatusIncludedNow, mechanicStatusDeferred, mechanicStatusExplicitlyExcluded:
		return true
	default:
		return false
	}
}

func validateSyntheticContracts(contracts []SyntheticRecordContract) (map[string]SyntheticRecordContract, error) {
	byID := make(map[string]SyntheticRecordContract, len(contracts))
	seenFamilies := make(map[string]bool, len(contracts))
	for _, contract := range contracts {
		if contract.ID == "" || contract.Family == "" || contract.RecordKind == "" || contract.SyntheticIDPrefix == "" || len(contract.RequiredTags) == 0 || contract.Purpose == "" {
			return nil, fmt.Errorf("invalid synthetic contract %q", contract.ID)
		}
		if contract.Source.Repository != upstreamRepository || contract.Source.Commit != upstreamCommit || contract.Source.Path == "" || contract.Source.Symbol == "" {
			return nil, fmt.Errorf("invalid synthetic contract provenance %q", contract.ID)
		}
		if _, exists := byID[contract.ID]; exists {
			return nil, fmt.Errorf("duplicate synthetic contract %q", contract.ID)
		}
		if seenFamilies[contract.Family] {
			return nil, fmt.Errorf("duplicate synthetic contract family %q", contract.Family)
		}
		byID[contract.ID] = contract
		seenFamilies[contract.Family] = true
	}
	if !sameStrings(mapKeys(seenFamilies), requiredMechanicFamilies()) {
		return nil, errors.New("synthetic contract inventory is incomplete")
	}
	return byID, nil
}

func validateMechanicMatrix(rows []MechanicReadiness, contracts map[string]SyntheticRecordContract) error {
	seenFamilies := make(map[string]bool, len(rows))
	requiredFamilies := make(map[string]bool, len(contracts))
	statusSeen := make(map[string]bool, 3)
	for _, contract := range contracts {
		requiredFamilies[contract.Family] = true
	}
	for _, row := range rows {
		if row.Family == "" || row.Status == "" || strings.TrimSpace(row.Notes) == "" {
			return fmt.Errorf("invalid mechanic matrix row %q", row.Family)
		}
		if !validMechanicStatus(row.Status) {
			return fmt.Errorf("unknown mechanic status %q for %s", row.Status, row.Family)
		}
		if seenFamilies[row.Family] {
			return fmt.Errorf("duplicate mechanic matrix family %q", row.Family)
		}
		switch row.Status {
		case mechanicStatusExplicitlyExcluded:
			if strings.TrimSpace(row.ContractID) != "" {
				return fmt.Errorf("explicitly excluded mechanic %q must not reference a synthetic contract", row.Family)
			}
			for layer, value := range map[string]string{
				"importer": row.Importer,
				"runtime":  row.Runtime,
				"engine":   row.Engine,
				"api":      row.API,
				"ai":       row.AI,
			} {
				if strings.TrimSpace(value) != "" {
					return fmt.Errorf("explicitly excluded mechanic %q must not declare %s readiness", row.Family, layer)
				}
			}
		default:
			if strings.TrimSpace(row.ContractID) == "" {
				return fmt.Errorf("mechanic %q is missing synthetic contract linkage", row.Family)
			}
			contract, ok := contracts[row.ContractID]
			if !ok {
				return fmt.Errorf("mechanic %q references unknown contract %q", row.Family, row.ContractID)
			}
			if contract.Family != row.Family {
				return fmt.Errorf("mechanic %q contract mismatch %q", row.Family, row.ContractID)
			}
			delete(requiredFamilies, row.Family)
			if row.Status == mechanicStatusIncludedNow {
				for layer, value := range map[string]string{
					"importer": row.Importer,
					"runtime":  row.Runtime,
					"engine":   row.Engine,
					"api":      row.API,
					"ai":       row.AI,
				} {
					if strings.TrimSpace(value) == "" {
						return fmt.Errorf("mechanic %q missing %s readiness", row.Family, layer)
					}
				}
			}
		}
		statusSeen[row.Status] = true
		seenFamilies[row.Family] = true
	}
	if len(requiredFamilies) != 0 {
		return errors.New("mechanic matrix inventory is incomplete")
	}
	for _, requiredStatus := range []string{mechanicStatusIncludedNow, mechanicStatusExplicitlyExcluded} {
		if !statusSeen[requiredStatus] {
			return errors.New("mechanic matrix status coverage is incomplete")
		}
	}
	return nil
}

func Import(config ImportConfig) (Ruleset, error) {
	if strings.TrimSpace(config.Directory) == "" {
		return Ruleset{}, importError(ErrorInvalidDirectory, "", 1, "", "directory must be provided")
	}
	manifest := config.Manifest
	if manifest.CompatibilityVersion == "" {
		manifest = SupportedManifest()
	}
	if err := manifest.Validate(); err != nil {
		return Ruleset{}, err
	}
	entries, err := os.ReadDir(config.Directory)
	if err != nil {
		return Ruleset{}, importError(ErrorInvalidDirectory, config.Directory, 1, "", err.Error())
	}
	var parsed []Record
	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(strings.ToLower(entry.Name()), ".txt") && !strings.HasSuffix(strings.ToLower(entry.Name()), ".ctp2synthetic")) {
			continue
		}
		catalog, ok := catalogFileFor(entry.Name())
		if !ok {
			return Ruleset{}, importError(ErrorUnsupportedMechanic, entry.Name(), 1, "", "catalog filename is outside the supported CTP2 subset")
		}
		records, err := parseCTP2File(filepath.Join(config.Directory, entry.Name()), catalog)
		if err != nil {
			return Ruleset{}, err
		}
		for _, record := range records {
			if seen[record.ID] {
				return Ruleset{}, importError(ErrorDuplicateID, entry.Name(), 1, record.ID, "record ID already exists")
			}
			seen[record.ID] = true
			parsed = append(parsed, record)
		}
	}
	if len(parsed) == 0 {
		return Ruleset{}, importError(ErrorMalformedValue, config.Directory, 1, "", "no supported CTP2 records")
	}
	sort.Slice(parsed, func(i, j int) bool { return parsed[i].ID < parsed[j].ID })
	canonical, err := json.Marshal(struct {
		CompatibilityVersion string                 `json:"compatibility_version"`
		Manifest             SupportedRulesManifest `json:"manifest"`
		Records              []Record               `json:"records"`
	}{CTP2CompatibilityVersion, normalizedManifest(manifest), parsed})
	if err != nil {
		return Ruleset{}, err
	}
	canonical, err = CanonicalDefinitions(canonical)
	if err != nil {
		return Ruleset{}, err
	}
	digest := sha256.Sum256(canonical)
	return Ruleset{CompatibilityVersion: CTP2CompatibilityVersion, Manifest: normalizedManifest(manifest), Records: parsed, CanonicalJSON: canonical, ContentHash: hex.EncodeToString(digest[:])}, nil
}

// PlayableCivilizationIDs returns stable imported IDs that are explicitly selectable.
// Missing or malformed metadata is rejected during import, so special and barbarian
// records remain unavailable unless all three metadata fields are present and false.
func (r Ruleset) PlayableCivilizationIDs() ([]string, error) {
	ids := make([]string, 0)
	for _, record := range r.Records {
		if record.Kind != "civilization" {
			continue
		}
		if record.Tags["Playable"] == "true" && record.Tags["Special"] == "false" && record.Tags["Barbarian"] == "false" {
			ids = append(ids, record.ID)
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("no selectable playable civilizations")
	}
	sort.Strings(ids)
	return ids, nil
}

func catalogFileFor(name string) (CatalogFile, bool) {
	isSynthetic := strings.HasSuffix(strings.ToLower(name), ".ctp2synthetic")
	base := strings.TrimSuffix(name, ".ctp2synthetic")
	for _, catalog := range CatalogFiles {
		if strings.EqualFold(base, catalog.Filename) || (isSynthetic && strings.EqualFold(base, strings.TrimSuffix(catalog.Filename, ".txt"))) {
			return catalog, true
		}
	}
	return CatalogFile{}, false
}

type token struct {
	value string
	line  int
}

func parseCTP2File(path string, catalog CatalogFile) ([]Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, importError(ErrorMalformedValue, path, 1, "", err.Error())
	}
	tokens, err := lexCTP2(string(data), path)
	if err != nil {
		return nil, err
	}
	var records []Record
	for index := 0; index < len(tokens); {
		id := tokens[index]
		if id.value == "{" || id.value == "}" || index+1 >= len(tokens) || tokens[index+1].value != "{" {
			return nil, importError(ErrorMalformedValue, path, id.line, "", "expected RECORD_ID followed by {")
		}
		if !strings.HasPrefix(id.value, catalog.Prefix) {
			return nil, importError(ErrorUnsupportedMechanic, path, id.line, id.value, "record prefix does not match catalog file")
		}
		if catalog.Kind == "rules" {
			return nil, importError(ErrorUnsupportedMechanic, path, id.line, id.value, "Const.txt semantics are not implemented")
		}
		index += 2
		tags := map[string]string{}
		for index < len(tokens) && tokens[index].value != "}" {
			key := tokens[index]
			if index+1 >= len(tokens) || tokens[index+1].value == "}" || tokens[index+1].value == "{" {
				return nil, importError(ErrorUnsupportedRequiredTag, path, key.line, id.value, "nested or valueless behavior is unsupported: "+key.value)
			}
			value := tokens[index+1]
			if index+2 < len(tokens) && tokens[index+2].value == "{" {
				return nil, importError(ErrorUnsupportedRequiredTag, path, key.line, id.value, "nested behavior is unsupported: "+key.value)
			}
			if _, exists := tags[key.value]; exists {
				return nil, importError(ErrorMalformedValue, path, key.line, id.value, "duplicate field "+key.value)
			}
			tags[key.value] = value.value
			index += 2
		}
		if index == len(tokens) {
			return nil, importError(ErrorMalformedValue, path, id.line, id.value, "missing closing brace")
		}
		if err := validateCTP2Tags(catalog.Kind, id, tags, path); err != nil {
			return nil, err
		}
		records = append(records, Record{Kind: catalog.Kind, ID: id.value, Tags: tags})
		index++
	}
	return records, nil
}

func validateCTP2Tags(kind string, id token, tags map[string]string, path string) error {
	allowed := map[string]bool{}
	required := []string{}
	switch kind {
	case "unit":
		allowed = map[string]bool{"Description": true, "MaxMovePoints": true, "Role": true, "Domain": true, "Attack": true, "Defense": true, "Strength": true, "Armor": true, "Health": true, "ZOC": true, "TransportCapacity": true, "CargoDomains": true, "ProductionCost": true, "Upkeep": true, "FoodHunger": true, "Prerequisites": true, "ObsoleteBy": true, "VisibilityRange": true, "Effects": true}
		required = []string{"Description", "MaxMovePoints", "Role"}
	case "advance":
		allowed = map[string]bool{"Description": true, "Cost": true, "PrerequisiteAdvanceIDs": true}
		required = []string{"Description", "Cost"}
	case "terrain_improvement":
		allowed = map[string]bool{"Description": true, "BuildTurns": true, "PublicWorksCost": true, "MovementCostDelta": true, "YieldFood": true, "YieldProduction": true, "YieldGold": true, "PollutionDelta": true, "TargetTerrain": true}
		required = []string{"Description", "BuildTurns", "PublicWorksCost"}
	case "pollution_source":
		allowed = map[string]bool{"Description": true, "PollutionPoints": true, "EffectScope": true}
		required = []string{"Description", "PollutionPoints", "EffectScope"}
	case "government":
		allowed = map[string]bool{"Description": true, "EnableAdvanceID": true, "ResearchBonus": true, "ResearchPercent": true, "PublicWorksBonus": true, "FoodPercent": true, "ProductionPercent": true, "GoldPercent": true, "PollutionPercent": true}
		required = []string{"Description"}
	case "building", "wonder", "terrain":
		allowed = map[string]bool{"Description": true}
		required = []string{"Description"}
	case "endgame":
		allowed = map[string]bool{"TurnsToActivate": true}
		required = []string{"TurnsToActivate"}
	case "civilization":
		allowed = map[string]bool{"Playable": true, "Special": true, "Barbarian": true}
		required = []string{"Playable", "Special", "Barbarian"}
	}
	for key, value := range tags {
		if !allowed[key] {
			return importError(ErrorUnsupportedRequiredTag, path, id.line, id.value, "unsupported behavior-bearing field "+key)
		}
		if key == "MaxMovePoints" || key == "TurnsToActivate" || key == "Cost" || key == "BuildTurns" || key == "PublicWorksCost" || key == "MovementCostDelta" || key == "YieldFood" || key == "YieldProduction" || key == "YieldGold" || key == "PollutionDelta" || key == "PollutionPoints" || key == "ResearchBonus" || key == "ResearchPercent" || key == "PublicWorksBonus" || key == "FoodPercent" || key == "ProductionPercent" || key == "GoldPercent" || key == "PollutionPercent" || key == "Attack" || key == "Defense" || key == "Strength" || key == "Armor" || key == "Health" || key == "TransportCapacity" || key == "ProductionCost" || key == "Upkeep" || key == "VisibilityRange" {
			if _, err := strconv.Atoi(value); err != nil {
				return importError(ErrorMalformedValue, path, id.line, id.value, key+" must be an integer")
			}
		}
		if key == "ZOC" && value != "true" && value != "false" {
			return importError(ErrorMalformedValue, path, id.line, id.value, key+" must be true or false")
		}
		if kind == "civilization" {
			if value != "true" && value != "false" {
				return importError(ErrorMalformedValue, path, id.line, id.value, key+" must be true or false")
			}
		}
	}
	for _, key := range required {
		if tags[key] == "" {
			return importError(ErrorMissingRequiredTag, path, id.line, id.value, "missing required field "+key)
		}
	}
	return nil
}

func lexCTP2(input, filename string) ([]token, error) {
	var tokens []token
	line := 1
	for index := 0; index < len(input); {
		char := rune(input[index])
		if char == '\n' {
			line++
			index++
			continue
		}
		if unicode.IsSpace(char) {
			index++
			continue
		}
		if char == '#' {
			for index < len(input) && input[index] != '\n' {
				index++
			}
			continue
		}
		if char == '{' || char == '}' {
			tokens = append(tokens, token{string(char), line})
			index++
			continue
		}
		if char == '"' {
			start := index
			index++
			for index < len(input) && input[index] != '"' {
				if input[index] == '\n' {
					line++
				}
				index++
			}
			if index == len(input) {
				return nil, importError(ErrorMalformedValue, filename, line, "", "unterminated quoted string")
			}
			tokens = append(tokens, token{input[start+1 : index], line})
			index++
			continue
		}
		if !isTokenCharacter(char) {
			return nil, importError(ErrorMalformedValue, filename, line, "", "invalid CTP2 token character "+string(char))
		}
		start := index
		for index < len(input) && isTokenCharacter(rune(input[index])) {
			index++
		}
		tokens = append(tokens, token{input[start:index], line})
	}
	return tokens, nil
}

func isTokenCharacter(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' || char == '.' || char == '+' || char == '-'
}
func importError(code, filename string, line int, id, detail string) *ImportError {
	return &ImportError{Code: code, Filename: filename, Line: line, RecordID: id, Detail: detail}
}

func normalizedManifest(manifest SupportedRulesManifest) SupportedRulesManifest {
	result := manifest
	result.Modes = sortedStrings(manifest.Modes)
	result.SupportedMechanics = sortedStrings(manifest.SupportedMechanics)
	result.RejectedTags = sortedStrings(manifest.RejectedTags)
	result.OpaqueTags = sortedStrings(manifest.OpaqueTags)
	result.VictoryEvaluators = append([]VictoryEvaluator(nil), manifest.VictoryEvaluators...)
	sort.Slice(result.VictoryEvaluators, func(i, j int) bool { return result.VictoryEvaluators[i].ID < result.VictoryEvaluators[j].ID })
	result.VictoryVectors = append([]VictoryVector(nil), manifest.VictoryVectors...)
	sort.Slice(result.VictoryVectors, func(i, j int) bool { return result.VictoryVectors[i].ID < result.VictoryVectors[j].ID })
	result.RejectedVictoryEvaluators = append([]RejectedVictoryEvaluator(nil), manifest.RejectedVictoryEvaluators...)
	sort.Slice(result.RejectedVictoryEvaluators, func(i, j int) bool {
		return result.RejectedVictoryEvaluators[i].ID < result.RejectedVictoryEvaluators[j].ID
	})
	result.SpecialActionEvaluators = append([]SpecialActionEvaluator(nil), manifest.SpecialActionEvaluators...)
	sort.Slice(result.SpecialActionEvaluators, func(i, j int) bool {
		return result.SpecialActionEvaluators[i].ID < result.SpecialActionEvaluators[j].ID
	})
	result.SpecialActionVectors = append([]SpecialActionVector(nil), manifest.SpecialActionVectors...)
	sort.Slice(result.SpecialActionVectors, func(i, j int) bool {
		return result.SpecialActionVectors[i].ID < result.SpecialActionVectors[j].ID
	})
	result.MechanicMatrix = append([]MechanicReadiness(nil), manifest.MechanicMatrix...)
	sort.Slice(result.MechanicMatrix, func(i, j int) bool {
		return result.MechanicMatrix[i].Family < result.MechanicMatrix[j].Family
	})
	result.SyntheticContracts = append([]SyntheticRecordContract(nil), manifest.SyntheticContracts...)
	for i := range result.SyntheticContracts {
		result.SyntheticContracts[i].RequiredTags = sortedStrings(result.SyntheticContracts[i].RequiredTags)
	}
	sort.Slice(result.SyntheticContracts, func(i, j int) bool {
		return result.SyntheticContracts[i].ID < result.SyntheticContracts[j].ID
	})
	return result
}
func sameStrings(got, want []string) bool {
	return strings.Join(sortedStrings(got), ",") == strings.Join(sortedStrings(want), ",")
}
func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
func mapKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

type AssetFile struct {
	Path    string
	Content []byte
}

func CheckProtectedAssets(files []AssetFile) error {
	marker := protectedAssetMarker()
	for _, file := range files {
		path := filepath.ToSlash(file.Path)
		if strings.Contains(string(file.Content), marker) {
			return fmt.Errorf("protected asset content: %s", path)
		}
		if strings.HasPrefix(path, "internal/gameplay/rules/testdata/synthetic-ctp2/") {
			continue
		}
		base := strings.ToLower(filepath.Base(path))
		if base == "units.txt" || base == "advance.txt" || base == "buildings.txt" || base == "wonder.txt" || base == "terrain.txt" || base == "terrainimprovement.txt" || base == "pollutionsource.txt" || base == "government.txt" || base == "endgameobjects.txt" || strings.Contains(strings.ToLower(path), "retail-snapshot") {
			return fmt.Errorf("protected asset path: %s", path)
		}
	}
	return nil
}
func CheckTrackedAssets(repositoryRoot string) error {
	output, err := exec.Command("git", "-C", repositoryRoot, "ls-files", "-z").Output()
	if err != nil {
		return err
	}
	var paths []string
	for _, path := range strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00") {
		if path != "" && !strings.HasSuffix(path, ".go") {
			paths = append(paths, path)
		}
	}
	return checkTrackedFiles(repositoryRoot, paths)
}

func checkTrackedFiles(repositoryRoot string, paths []string) error {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	files := make([]AssetFile, 0, len(sorted))
	for _, path := range sorted {
		content, err := os.ReadFile(filepath.Join(repositoryRoot, path))
		if err != nil {
			return fmt.Errorf("read tracked asset %q: %w", path, err)
		}
		files = append(files, AssetFile{Path: path, Content: content})
	}
	return CheckProtectedAssets(files)
}

func protectedAssetMarker() string {
	return strings.Join([]string{"CTP2", "RETAIL", "ASSET"}, "_")
}
