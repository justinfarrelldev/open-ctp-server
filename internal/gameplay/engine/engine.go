// Package engine defines the pure, deterministic gameplay transition boundary.
package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

const (
	StateSchemaV1   = 1
	CommandSchemaV1 = 1
	EventSchemaV1   = 1

	RNGDomainCombat        = "combat"
	RNGDomainSpecialAction = "special-action"
	RNGDomainAI            = "ai"

	CommandNoop               = "noop"
	CommandSetWar             = "set-war"
	CommandEstablishContact   = "establish-contact"
	CommandProposeAgreement   = "propose-agreement"
	CommandRespondAgreement   = "respond-agreement"
	CommandSetGovernment      = "set-government"
	CommandSetScienceRate     = "set-science-rate"
	CommandPass               = "pass"
	CommandMoveUnit           = "move-unit"
	CommandDisband            = "disband-unit"
	CommandFortify            = "fortify-unit"
	CommandEmbark             = "embark-unit"
	CommandDisembark          = "disembark-unit"
	CommandArmyCreate         = "create-army"
	CommandArmyJoin           = "join-army"
	CommandArmyLeave          = "leave-army"
	CommandFoundCity          = "found-city"
	CommandQueueProduction    = "queue-production"
	CommandCancelProduction   = "cancel-production"
	CommandSetResearch        = "set-research"
	CommandCombat             = "combat"
	CommandBombard            = "bombard"
	CommandEstablishEmbassy   = "establish-embassy"
	CommandSoothsaying        = "soothsaying"
	CommandStealTechnology    = "steal-technology"
	CommandConvertCity        = "convert-city"
	CommandBioTerror          = "bio-terror"
	CommandNanoTerror         = "nano-terror"
	CommandPlague             = "plague"
	CommandSlaveRaid          = "slave-raid"
	CommandSettlerSlaveRaid   = "settler-slave-raid"
	CommandCreateFranchise    = "create-franchise"
	CommandInvestigateCity    = "investigate-city"
	CommandPlantNuke          = "plant-nuke"
	CommandCreatePark         = "create-park"
	CommandAssassinateRuler   = "assassinate-ruler"
	CommandExpel              = "expel"
	CommandInjoin             = "injoin"
	CommandReform             = "reform"
	CommandUndergroundRailway = "underground-railway"
	CommandSue                = "sue"
	CommandPillage            = "pillage"
	CommandTerraform          = "terraform"
	CommandBuildImprovement   = "build-improvement"
	CommandCancelTerraform    = "cancel-terraform"
	CommandRoundComplete      = "round-complete"
	CommandEndGame            = "end-game"
	CommandLeaveGame          = "leave-game"
	CommandTakeoverAI         = "takeover-ai"

	EventWarStateChanged          = "war.state_changed"
	EventContactEstablished       = "diplomacy.contact_established"
	EventAgreementProposed        = "diplomacy.agreement_proposed"
	EventAgreementResponded       = "diplomacy.agreement_responded"
	EventGovernmentChanged        = "government.changed"
	EventScienceRateChanged       = "government.science_rate_changed"
	EventUnitMoved                = "unit.moved"
	EventCombatResolved           = "combat.resolved"
	EventBombarded                = "bombard.resolved"
	EventUnitDisbanded            = "unit.disbanded"
	EventEmbassyEstablished       = "embassy.established"
	EventSoothsayingApplied       = "soothsaying.applied"
	EventTechnologyStolen         = "technology.stolen"
	EventTechnologyFailed         = "technology.failed"
	EventCityConverted            = "city.converted"
	EventCityConversionFailed     = "city.conversion_failed"
	EventBioTerrorApplied         = "city.bio_terror"
	EventNanoTerrorApplied        = "city.nano_terror"
	EventPlagueApplied            = "city.plague"
	EventTerrorFailed             = "city.terror_failed"
	EventSlaveRaidApplied         = "city.slave_raid"
	EventSlaveRaidFailed          = "city.slave_raid_failed"
	EventFranchiseCreated         = "city.franchise_created"
	EventFranchiseFailed          = "city.franchise_failed"
	EventCityInvestigated         = "city.investigated"
	EventCityInvestigationFailed  = "city.investigation_failed"
	EventNukePlanted              = "city.nuked"
	EventNukeFailed               = "city.nuke_failed"
	EventParkCreated              = "tile.park_created"
	EventRulerAssassinated        = "nation.ruler_assassinated"
	EventRulerAssassinationFailed = "nation.ruler_assassination_failed"
	EventUnitExpelled             = "unit.expelled"
	EventCityInjoined             = "city.injoined"
	EventCityReformed             = "city.reformed"
	EventUndergroundRailway       = "city.underground_railway"
	EventUndergroundRailwayFailed = "city.underground_railway_failed"
	EventCitySued                 = "city.sued"
	EventUnitsSued                = "unit.sued"
	EventTechnologyCapture        = "technology.capture"
	EventPillaged                 = "pillage.applied"
	EventTerraformQueued          = "terraform.queued"
	EventImprovementQueued        = "improvement.queued"
	EventTerraformCancelled       = "terraform.cancelled"
	EventTerminal                 = "terminal"
	EventPlayerDisconnected       = "player.disconnected"
	EventControllerChanged        = "player.controller_changed"

	DomainLand       = "land"
	DomainSea        = "sea"
	DomainAmphibious = "amphibious"
	DomainAir        = "air"
	DomainSpace      = "space"

	TerrainPlains = "plains"
	TerrainForest = "forest"
	TerrainOcean  = "ocean"

	PlayerActive    = "active"
	PlayerLeft      = "left"
	ControllerHuman = "human"
	ControllerAI    = "ai"
)

var (
	ErrUnknownStateSchema   = errors.New("unknown gameplay state schema")
	ErrUnknownCommandSchema = errors.New("unknown gameplay command schema")
	ErrRulesPinMismatch     = errors.New("rules pin mismatch")
)

// Clock, IDSource, and RNGSource make all nondeterministic inputs explicit.
type Clock interface{ Now() time.Time }
type IDSource interface{ Next() string }
type RNGSource interface {
	Next(domain string, state uint64) (nextState uint64, value uint64)
}

type Sources struct {
	Clock Clock
	IDs   IDSource
	RNG   RNGSource
}

type RulesPin struct {
	Version     string          `json:"version"`
	ContentHash string          `json:"content_hash"`
	Definitions json.RawMessage `json:"definitions"`
}

type TileOverlay struct {
	ID                string `json:"id"`
	TileID            string `json:"tile_id"`
	Kind              string `json:"kind"`
	OwnerID           string `json:"owner_id,omitempty"`
	Remaining         int    `json:"remaining,omitempty"`
	TargetTerrain     string `json:"target_terrain,omitempty"`
	ImprovementID     string `json:"improvement_id,omitempty"`
	MovementCostDelta int    `json:"movement_cost_delta,omitempty"`
	YieldFood         int    `json:"yield_food,omitempty"`
	YieldProduction   int    `json:"yield_production,omitempty"`
	YieldGold         int    `json:"yield_gold,omitempty"`
	PollutionDelta    int    `json:"pollution_delta,omitempty"`
}

// Tile is canonical server terrain/access input for unit transitions. Coordinates
// are used only to validate submitted adjacent path steps.
type Tile struct {
	ID        string `json:"id"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Terrain   string `json:"terrain"`
	OwnerID   string `json:"owner_id,omitempty"`
	Open      bool   `json:"open"`
	Pollution int    `json:"pollution,omitempty"`
}

type Unit struct {
	ID                 string   `json:"id"`
	OwnerID            string   `json:"owner_id"`
	TileID             string   `json:"tile_id"`
	ArmyID             string   `json:"army_id,omitempty"`
	TransportID        string   `json:"transport_id,omitempty"`
	Domain             string   `json:"domain,omitempty"`
	Movement           int      `json:"movement,omitempty"`
	MaxMovement        int      `json:"max_movement,omitempty"`
	Fuel               int      `json:"fuel,omitempty"`
	MaxFuel            int      `json:"max_fuel,omitempty"`
	CargoCapacity      int      `json:"cargo_capacity,omitempty"`
	CargoDomains       []string `json:"cargo_domains,omitempty"`
	Upkeep             int      `json:"upkeep,omitempty"`
	ZOC                bool     `json:"zoc,omitempty"`
	Fortified          bool     `json:"fortified,omitempty"`
	Veteran            bool     `json:"veteran,omitempty"`
	Effects            []Effect `json:"effects,omitempty"`
	Attack             int      `json:"attack,omitempty"`
	Defense            int      `json:"defense,omitempty"`
	Ranged             int      `json:"ranged,omitempty"`
	BombardRange       int      `json:"bombard_range,omitempty"`
	BombRounds         int      `json:"bomb_rounds,omitempty"`
	ActiveDefenseRange int      `json:"active_defense_range,omitempty"`
	Strength           int      `json:"strength,omitempty"`
	Armor              int      `json:"armor,omitempty"`
	Health             int      `json:"health,omitempty"`
	MaxHealth          int      `json:"max_health,omitempty"`
	VisibilityRange    int      `json:"visibility_range,omitempty"`
}

type City struct {
	ID                string           `json:"id"`
	Name              string           `json:"name,omitempty"`
	OwnerID           string           `json:"owner_id"`
	TileID            string           `json:"tile_id"`
	Population        int              `json:"population,omitempty"`
	Happiness         int              `json:"happiness,omitempty"`
	CrimePercent      int              `json:"crime_percent,omitempty"`
	OvercrowdingLevel int              `json:"overcrowding_level,omitempty"`
	WorkedTileIDs     []string         `json:"worked_tile_ids,omitempty"`
	StoredFood        int              `json:"stored_food,omitempty"`
	StarvationTurns   int              `json:"starvation_turns,omitempty"`
	StoredProduction  int              `json:"stored_production,omitempty"`
	Queue             []ProductionItem `json:"queue,omitempty"`
	Buildings         []Building       `json:"buildings,omitempty"`
	Obsolete          []string         `json:"obsolete,omitempty"`
	Yields            Yield            `json:"yields,omitempty"`
	Effects           []Effect         `json:"effects,omitempty"`
}

type Army struct {
	ID      string   `json:"id"`
	OwnerID string   `json:"owner_id"`
	UnitIDs []string `json:"unit_ids"`
}

type Nation struct {
	ID                  string       `json:"id"`
	Effects             []Effect     `json:"effects,omitempty"`
	Advances            []string     `json:"advances,omitempty"`
	Embassies           []string     `json:"embassies,omitempty"`
	CityReports         []CityReport `json:"city_reports,omitempty"`
	PublicWorks         int          `json:"public_works,omitempty"`
	FranchiseProduction int          `json:"franchise_production,omitempty"`
	SciencePercent      *int         `json:"science_percent,omitempty"`
	GovernmentID        string       `json:"government_id,omitempty"`
	ResearchTargetID    string       `json:"research_target_id,omitempty"`
	ResearchProgress    int          `json:"research_progress,omitempty"`
	GaiaCompletionTurn  int          `json:"gaia_completion_turn,omitempty"`
}

type CityReport struct {
	CityID     string   `json:"city_id"`
	OwnerID    string   `json:"owner_id"`
	TileID     string   `json:"tile_id"`
	Turn       int      `json:"turn"`
	Population int      `json:"population,omitempty"`
	Buildings  []string `json:"buildings,omitempty"`
	Effects    []Effect `json:"effects,omitempty"`
}

type Effect struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Value int    `json:"value,omitempty"`
}

type DiplomaticRelation struct {
	ID            string `json:"id"`
	PlayerAID     string `json:"player_a_id"`
	PlayerBID     string `json:"player_b_id"`
	Contact       bool   `json:"contact"`
	Treaty        string `json:"treaty,omitempty"`
	Trade         bool   `json:"trade,omitempty"`
	EstablishedAt int    `json:"established_at,omitempty"`
}

type DiplomaticProposal struct {
	ID           string `json:"id"`
	FromPlayerID string `json:"from_player_id"`
	ToPlayerID   string `json:"to_player_id"`
	Kind         string `json:"kind"`
	CreatedTurn  int    `json:"created_turn"`
}

type TradeRoute struct {
	ID                string `json:"id"`
	SourceCityID      string `json:"source_city_id"`
	DestinationCityID string `json:"destination_city_id"`
	CrossesWater      bool   `json:"crosses_water,omitempty"`
}

type Player struct {
	ID              string `json:"id"`
	NationID        string `json:"nation_id"`
	Lifecycle       string `json:"lifecycle"`
	Controller      string `json:"controller"`
	TakeoverPending bool   `json:"takeover_pending,omitempty"`
}

type PhaseState struct {
	Mode              string   `json:"mode"`
	ModeVersion       string   `json:"mode_version"`
	Turn              int      `json:"turn"`
	ActivePlayerIndex int      `json:"active_player_index"`
	ActivePlayerOrder []string `json:"active_player_order"`
	BaseRevision      int64    `json:"base_revision"`
	DeadlineID        string   `json:"deadline_id,omitempty"`
	ResolutionLocked  bool     `json:"resolution_locked"`
}

type TurnPolicyCommand struct {
	Kind            string          `json:"kind"`
	ActorID         string          `json:"actor_id"`
	ActionClass     string          `json:"action_class,omitempty"`
	UnitOrArmyID    string          `json:"unit_or_army_id,omitempty"`
	ReceiptSequence int64           `json:"receipt_sequence,omitempty"`
	Hostile         bool            `json:"hostile,omitempty"`
	TargetPlayerID  string          `json:"target_player_id,omitempty"`
	CommandID       string          `json:"command_id,omitempty"`
	IdempotencyKey  string          `json:"idempotency_key,omitempty"`
	Payload         json.RawMessage `json:"payload,omitempty"`
}

type TurnPolicyWar struct {
	A string `json:"a"`
	B string `json:"b"`
}

type TurnPolicyDeadline struct {
	PlayerID string `json:"player_id"`
	AtUnix   int64  `json:"at_unix"`
}

type TurnPolicyState struct {
	Stage          string              `json:"stage,omitempty"`
	Queue          []TurnPolicyCommand `json:"queue,omitempty"`
	EarlyEnds      []string            `json:"early_ends,omitempty"`
	Deadline       *TurnPolicyDeadline `json:"deadline,omitempty"`
	Components     [][]string          `json:"components,omitempty"`
	Wars           []TurnPolicyWar     `json:"wars,omitempty"`
	ComponentIndex int                 `json:"component_index,omitempty"`
	Recalculations int                 `json:"recalculations,omitempty"`
}

type FogObservation struct {
	PlayerID    string   `json:"player_id"`
	TileID      string   `json:"tile_id"`
	Turn        int      `json:"turn"`
	Terrain     string   `json:"terrain"`
	CityID      string   `json:"city_id,omitempty"`
	CityOwnerID string   `json:"city_owner_id,omitempty"`
	CityEffects []Effect `json:"city_effects,omitempty"`
}

type RNGDomainState struct {
	Domain string `json:"domain"`
	State  uint64 `json:"state"`
}

type Deadline struct {
	ID       string `json:"id"`
	Phase    string `json:"phase"`
	PlayerID string `json:"player_id,omitempty"`
	AtUnix   int64  `json:"at_unix"`
}

type PollutionState struct {
	Phase              int     `json:"phase,omitempty"`
	GWPhase            int     `json:"gw_phase,omitempty"`
	History            []int64 `json:"history,omitempty"`
	Trend              string  `json:"trend,omitempty"`
	LastDisaster       string  `json:"last_disaster,omitempty"`
	LastTriggerLevel   int64   `json:"last_trigger_level,omitempty"`
	LastPollutionLevel int64   `json:"last_pollution_level,omitempty"`
}

type VictoryState struct {
	WinnerNationID string `json:"winner_nation_id,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type TerminalState struct {
	Ended           bool   `json:"ended"`
	Reason          string `json:"reason,omitempty"`
	WinningPlayerID string `json:"winning_player_id,omitempty"`
}

type Audit struct {
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// State is the complete versioned dynamic authority. Storage bytes retain audit timestamps; gameplay hashes exclude them.
type State struct {
	SchemaVersion    int                  `json:"schema_version"`
	Rules            RulesPin             `json:"rules"`
	Revision         int64                `json:"revision"`
	EventSequence    int64                `json:"event_sequence"`
	RecalculatedTurn int                  `json:"recalculated_turn,omitempty"`
	Overlays         []TileOverlay        `json:"overlays"`
	Tiles            []Tile               `json:"tiles"`
	Units            []Unit               `json:"units"`
	Cities           []City               `json:"cities"`
	Armies           []Army               `json:"armies"`
	Nations          []Nation             `json:"nations"`
	Wonders          []Wonder             `json:"wonders,omitempty"`
	Pollution        *PollutionState      `json:"pollution,omitempty"`
	Diplomacy        []DiplomaticRelation `json:"diplomacy,omitempty"`
	TradeRoutes      []TradeRoute         `json:"trade_routes,omitempty"`
	Proposals        []DiplomaticProposal `json:"diplomatic_proposals,omitempty"`
	Players          []Player             `json:"players"`
	Phase            PhaseState           `json:"phase"`
	Policy           TurnPolicyState      `json:"policy"`
	Fog              []FogObservation     `json:"fog"`
	RNG              []RNGDomainState     `json:"rng"`
	Deadlines        []Deadline           `json:"deadlines"`
	Victory          VictoryState         `json:"victory"`
	Terminal         TerminalState        `json:"terminal"`
	Audit            Audit                `json:"audit,omitempty"`
}

type Command struct {
	SchemaVersion  int             `json:"schema_version"`
	ID             string          `json:"id"`
	ActorPlayerID  string          `json:"actor_player_id"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload,omitempty"`
}

type Event struct {
	SchemaVersion int             `json:"schema_version"`
	ID            string          `json:"id"`
	Sequence      int64           `json:"sequence"`
	Ordinal       int             `json:"ordinal"`
	Type          string          `json:"type"`
	ActorPlayerID string          `json:"actor_player_id"`
	CommandID     string          `json:"command_id"`
	Payload       json.RawMessage `json:"payload,omitempty"`
	AuditAt       time.Time       `json:"audit_at,omitempty"`
}

type Result struct {
	Revision      int64         `json:"revision"`
	EventSequence int64         `json:"event_sequence"`
	CommandID     string        `json:"command_id"`
	Combat        *CombatResult `json:"combat,omitempty"`
}

func (s State) RNGState(domain string) (RNGDomainState, bool) {
	for _, entry := range s.RNG {
		if entry.Domain == domain {
			return entry, true
		}
	}
	return RNGDomainState{}, false
}

// Apply has no IO or ambient authority. Future mechanics extend its command switch.
func Apply(state State, command Command, pin RulesPin, sources Sources) (State, []Event, Result, error) {
	runtime, err := prepareApply(state, command, pin, sources)
	if err != nil {
		return State{}, nil, Result{}, err
	}
	next, err := canonicalClone(state)
	if err != nil {
		return State{}, nil, Result{}, err
	}
	if command.Type != CommandSetWar && command.Type != CommandTakeoverAI && (command.Type == CommandPass || next.Phase.Mode == turns.ModeLongTurn || next.Phase.Mode == turns.ModeSimultaneous || next.Phase.Mode == turns.ModeHybrid) {
		return applyPolicyCommand(next, command, runtime, sources)
	}
	eventType, eventPayload, combatResult, err := applyValidatedMutation(&next, command, runtime, sources)
	if err != nil {
		return State{}, nil, Result{}, err
	}
	return finalizeApply(next, command, eventType, eventPayload, combatResult, sources)
}

type preparedRuntime struct{ rules *rules.RuntimeRules }

// prepareApply centralizes the shared validation phase for normal commands and
// persisted policy replays. It never mutates the caller's state.
func prepareApply(state State, command Command, pin RulesPin, sources Sources) (*rules.RuntimeRules, error) {
	if err := ValidateState(state); err != nil {
		return nil, err
	}
	if err := validateRulesPin(pin); err != nil {
		return nil, err
	}
	if !sameRulesPin(state.Rules, pin) {
		return nil, ErrRulesPinMismatch
	}
	if err := validateCommand(command); err != nil {
		return nil, err
	}
	if err := validateSources(sources); err != nil {
		return nil, err
	}
	if state.Terminal.Ended {
		return nil, errors.New("game is already terminal")
	}
	if hasRuntimeDefinitions(pin.Definitions) {
		compiled, err := rules.CompileRuntime(pin.Definitions, pin.ContentHash)
		if err != nil {
			return nil, err
		}
		return &compiled, nil
	}
	return nil, nil
}

// applyValidatedMutation is the sole mechanics dispatch phase. Callers must use
// prepareApply first; policy replay uses this phase rather than mutating state directly.
func applyValidatedMutation(state *State, command Command, runtime *rules.RuntimeRules, sources Sources) (string, json.RawMessage, *CombatResult, error) {
	eventType := "command.accepted"
	eventPayload := canonicalPayload(command.Payload)
	var combatResult *CombatResult
	if command.Type == CommandNoop {
		return eventType, eventPayload, combatResult, nil
	}
	var err error
	switch command.Type {
	case CommandSetWar, CommandEstablishContact, CommandProposeAgreement, CommandRespondAgreement:
		eventType, eventPayload, err = applyDiplomacyCommand(state, command)
	case CommandSetGovernment, CommandSetScienceRate:
		eventType, eventPayload, err = applyGovernmentCommand(state, command, runtime)
	case CommandSetResearch, CommandRoundComplete, CommandEndGame, CommandLeaveGame, CommandTakeoverAI:
		eventType, eventPayload, err = applyLifecycleCommand(state, command, runtime)
	case CommandCombat:
		combatResult, eventPayload, err = applyCombat(state, command, sources)
		eventType = EventCombatResolved
	case CommandBombard:
		eventType, eventPayload, err = applyBombard(state, command, sources)
	case CommandFoundCity, CommandQueueProduction, CommandCancelProduction:
		eventType, err = applyCityCommand(state, command, runtime)
	default:
		if isSpecialCommand(command.Type) {
			eventType, eventPayload, err = applySpecialCommand(state, command, sources, runtime)
		} else {
			eventType, err = applyUnitCommand(state, command, runtime)
		}
	}
	return eventType, eventPayload, combatResult, err
}

func finalizeApply(next State, command Command, eventType string, eventPayload json.RawMessage, combatResult *CombatResult, sources Sources) (State, []Event, Result, error) {
	next.Revision++
	next.EventSequence++
	next.Audit.UpdatedAt = sources.Clock.Now().UTC()
	event := Event{SchemaVersion: EventSchemaV1, ID: sources.IDs.Next(), Sequence: next.EventSequence, Ordinal: 0, Type: eventType, ActorPlayerID: command.ActorPlayerID, CommandID: command.ID, Payload: eventPayload, AuditAt: next.Audit.UpdatedAt}
	if err := CheckInvariants(next); err != nil {
		return State{}, nil, Result{}, err
	}
	return next, []Event{event}, Result{Revision: next.Revision, EventSequence: next.EventSequence, CommandID: command.ID, Combat: combatResult}, nil
}

func applyPolicyCommand(next State, command Command, runtime *rules.RuntimeRules, sources Sources) (State, []Event, Result, error) {
	policy, err := turns.New(next.Phase.Mode, time.Second)
	if err != nil {
		return State{}, nil, Result{}, err
	}
	resolverState := next.AsTurnPolicyState()
	now := sources.Clock.Now().UTC()
	if resolverState.Deadline == nil {
		resolverState, _, err = policy.Start(resolverState, now)
		if err != nil {
			return State{}, nil, Result{}, err
		}
	}
	resolverCommand := policyEnvelope(command, next.EventSequence+1)
	if next.Phase.Mode == turns.ModeLongTurn && command.Type == CommandRoundComplete {
		resolverCommand.Kind = turns.CommandPass
	}
	resolvedPolicy, resolved, err := policy.Apply(resolverState, resolverCommand, now)
	if err != nil {
		return State{}, nil, Result{}, err
	}
	if (next.Phase.Mode == turns.ModeSimultaneous || next.Phase.Mode == turns.ModeHybrid) && resolverState.Stage == turns.StagePlanning && command.Type == CommandLeaveGame && !resolved.Immediate {
		resolvedPolicy, resolved, err = policy.Apply(resolvedPolicy, turns.Command{Kind: turns.CommandEarlyEnd, ActorID: command.ActorPlayerID}, now)
		if err != nil {
			return State{}, nil, Result{}, err
		}
	}
	next = next.WithTurnPolicyState(resolvedPolicy)
	if resolved.Resolved {
		next.Phase.ResolutionLocked = true
	}
	if resolved.Immediate {
		eventType, eventPayload, combatResult, err := applyValidatedMutation(&next, command, runtime, sources)
		if err != nil {
			return State{}, nil, Result{}, err
		}
		return finalizeApply(next, command, eventType, eventPayload, combatResult, sources)
	}
	var replayBaseline State
	if resolved.Resolved {
		replayBaseline, err = canonicalClone(next)
		if err != nil {
			return State{}, nil, Result{}, err
		}
	}
	var resolvedDisconnectPayload json.RawMessage
	for _, queued := range resolved.ResolvedCommands {
		replayed, err := decodeQueuedCommand(queued)
		if err != nil {
			return State{}, nil, Result{}, err
		}
		replayedEventType, replayedEventPayload, _, err := applyValidatedMutation(&next, replayed, runtime, sources)
		if err != nil {
			suppressed, suppressErr := suppressDeterministicPlanningMoveCollision(replayBaseline, replayed, runtime, sources, err)
			if suppressErr != nil {
				return State{}, nil, Result{}, suppressErr
			}
			if suppressed {
				continue
			}
			return State{}, nil, Result{}, err
		}
		if replayedEventType == EventPlayerDisconnected {
			resolvedDisconnectPayload = append(json.RawMessage(nil), replayedEventPayload...)
		}
	}
	if resolved.Resolved {
		next.Phase.ResolutionLocked = false
	}
	if resolved.Resolved && next.Policy.Stage == turns.StagePlanning {
		if err := completePolicyRound(&next, runtime); err != nil {
			return State{}, nil, Result{}, err
		}
	}
	eventType := "command.queued"
	eventPayload := canonicalPayload(command.Payload)
	if resolved.Resolved {
		eventType = "turn.resolved"
		if len(resolvedDisconnectPayload) > 0 {
			eventType = EventPlayerDisconnected
			eventPayload = resolvedDisconnectPayload
		}
	}
	return finalizeApply(next, command, eventType, eventPayload, nil, sources)
}

func policyEnvelope(command Command, receiptSequence int64) turns.Command {
	if command.Type == CommandPass {
		return turns.Command{Kind: turns.CommandPass, ActorID: command.ActorPlayerID}
	}
	if command.Type == CommandRoundComplete {
		return turns.Command{Kind: turns.CommandEarlyEnd, ActorID: command.ActorPlayerID}
	}
	persisted := command
	persisted.Payload = canonicalPayload(command.Payload)
	if persisted.IdempotencyKey == "" {
		persisted.IdempotencyKey = persisted.ID
	}
	payload, _ := json.Marshal(persisted)
	var metadata struct {
		UnitID         string `json:"unit_id"`
		ArmyID         string `json:"army_id"`
		TargetPlayerID string `json:"target_player_id"`
	}
	_ = json.Unmarshal(command.Payload, &metadata)
	unitOrArmyID := metadata.UnitID
	if unitOrArmyID == "" {
		unitOrArmyID = metadata.ArmyID
	}
	class := "non_hostile"
	if command.Type == CommandCombat {
		class = "hostile"
	} else if command.Type == CommandFoundCity || command.Type == CommandQueueProduction || command.Type == CommandCancelProduction {
		class = "economic"
	}
	return turns.Command{Kind: turns.CommandAction, ActorID: command.ActorPlayerID, ActionClass: class, UnitOrArmyID: unitOrArmyID, ReceiptSequence: receiptSequence, Hostile: class == "hostile", TargetPlayerID: metadata.TargetPlayerID, CommandID: persisted.ID, IdempotencyKey: persisted.IdempotencyKey, Payload: payload}
}

func decodeQueuedCommand(queued turns.Command) (Command, error) {
	var command Command
	if err := json.Unmarshal(queued.Payload, &command); err != nil {
		return Command{}, fmt.Errorf("decode queued command %q: %w", queued.CommandID, err)
	}
	if err := validateCommand(command); err != nil {
		return Command{}, fmt.Errorf("validate queued command %q: %w", queued.CommandID, err)
	}
	if command.ID != queued.CommandID || command.ActorPlayerID != queued.ActorID {
		return Command{}, errors.New("queued command envelope does not match stored ordering metadata")
	}
	return command, nil
}

func suppressDeterministicPlanningMoveCollision(replayBaseline State, command Command, runtime *rules.RuntimeRules, sources Sources, replayErr error) (bool, error) {
	if command.Type != CommandMoveUnit || !errors.Is(replayErr, ErrObstructed) {
		return false, nil
	}
	baseline, err := canonicalClone(replayBaseline)
	if err != nil {
		return false, err
	}
	_, _, _, err = applyValidatedMutation(&baseline, command, runtime, sources)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func completePolicyRound(state *State, runtime *rules.RuntimeRules) error {
	next, err := RecalculateRound(*state)
	if err != nil {
		return err
	}
	*state = next
	updateGaiaCountdown(state, runtime, state.Phase.Turn-1)
	advanceResearch(state, runtime)
	refreshStarterMovement(state, runtime)
	if winner, reason := evaluateVictory(*state, runtime); winner != "" {
		endGame(state, winner, reason)
	}
	return nil
}

func refreshStarterMovement(state *State, runtime *rules.RuntimeRules) {
	if runtime == nil || runtime.Starter.Movement < 1 {
		return
	}
	for index := range state.Units {
		if state.Units[index].TransportID == "" {
			state.Units[index].Movement = runtime.Starter.Movement
		}
	}
}

func ValidateState(state State) error {
	if state.SchemaVersion != StateSchemaV1 {
		return fmt.Errorf("%w: %d", ErrUnknownStateSchema, state.SchemaVersion)
	}
	if err := validateRulesPin(state.Rules); err != nil {
		return err
	}
	if state.Revision < 0 || state.EventSequence < 0 || state.Phase.Turn < 0 {
		return errors.New("revision, event sequence, and turn must be nonnegative")
	}
	if !turns.SupportedMode(state.Phase.Mode) {
		return fmt.Errorf("unsupported turn mode %q", state.Phase.Mode)
	}
	if state.Phase.ModeVersion != turns.ContractVersion {
		return fmt.Errorf("unsupported turn mode contract version %q", state.Phase.ModeVersion)
	}
	if state.Phase.BaseRevision < 0 || state.Phase.BaseRevision > state.Revision {
		return errors.New("phase base revision is invalid")
	}
	if state.Phase.DeadlineID != "" {
		found := false
		for _, deadline := range state.Deadlines {
			if deadline.ID == state.Phase.DeadlineID {
				found = true
				break
			}
		}
		if !found {
			return errors.New("phase deadline does not exist")
		}
	}
	if len(state.Phase.ActivePlayerOrder) == 0 {
		return errors.New("active player order is required")
	}
	if state.Phase.ActivePlayerIndex < 0 || state.Phase.ActivePlayerIndex >= len(state.Phase.ActivePlayerOrder) {
		return errors.New("active player index is out of range")
	}
	if err := uniqueIDs("overlay", overlayIDs(state.Overlays)); err != nil {
		return err
	}
	if err := uniqueIDs("unit", unitIDs(state.Units)); err != nil {
		return err
	}
	if err := uniqueIDs("city", cityIDs(state.Cities)); err != nil {
		return err
	}
	if err := uniqueIDs("army", armyIDs(state.Armies)); err != nil {
		return err
	}
	if err := uniqueIDs("nation", nationIDs(state.Nations)); err != nil {
		return err
	}
	if err := uniqueIDs("player", playerIDs(state.Players)); err != nil {
		return err
	}
	if err := uniqueIDs("diplomatic relation", relationIDs(state.Diplomacy)); err != nil {
		return err
	}
	if err := uniqueIDs("trade route", tradeRouteIDs(state.TradeRoutes)); err != nil {
		return err
	}
	if err := uniqueIDs("diplomatic proposal", proposalIDs(state.Proposals)); err != nil {
		return err
	}
	if err := validateDiplomacyState(state); err != nil {
		return err
	}
	if err := validateTradeRoutes(state); err != nil {
		return err
	}
	if err := uniqueIDs("deadline", deadlineIDs(state.Deadlines)); err != nil {
		return err
	}
	if err := uniqueIDs("active player", state.Phase.ActivePlayerOrder); err != nil {
		return err
	}
	for _, tile := range state.Tiles {
		if tile.Pollution < 0 {
			return errors.New("tile pollution must be nonnegative")
		}
	}
	if state.Pollution != nil {
		for _, value := range state.Pollution.History {
			if value < 0 {
				return errors.New("pollution history must be nonnegative")
			}
		}
	}
	players := makeIDSet(playerIDs(state.Players))
	for _, id := range state.Phase.ActivePlayerOrder {
		if !players[id] {
			return fmt.Errorf("active player %q does not exist", id)
		}
	}
	if err := validateTurnPolicyState(state.Policy, players); err != nil {
		return err
	}
	for _, player := range state.Players {
		if player.Lifecycle != PlayerActive && player.Lifecycle != PlayerLeft {
			return fmt.Errorf("unknown player lifecycle %q", player.Lifecycle)
		}
		if player.Controller != ControllerHuman && player.Controller != ControllerAI {
			return fmt.Errorf("unknown player controller %q", player.Controller)
		}
	}
	for _, nation := range state.Nations {
		if nation.GaiaCompletionTurn < 0 || nation.ResearchProgress < 0 {
			return errors.New("nation counters must be nonnegative")
		}
	}
	if err := validateUnitState(state); err != nil {
		return err
	}
	if err := validateCityState(state); err != nil {
		return err
	}
	if err := validateSpecialState(state); err != nil {
		return err
	}
	if err := validateRNG(state.RNG); err != nil {
		return err
	}
	if state.Terminal.Ended != (state.Terminal.Reason != "" && state.Terminal.WinningPlayerID != "") {
		return errors.New("terminal state must have complete metadata exactly when ended")
	}
	return nil
}

func validateTradeRoutes(state State) error {
	cities := makeIDSet(cityIDs(state.Cities))
	for _, route := range state.TradeRoutes {
		if route.ID == "" || route.SourceCityID == "" || route.DestinationCityID == "" || route.SourceCityID == route.DestinationCityID {
			return errors.New("trade route is invalid")
		}
		if len(state.Cities) > 0 && (!cities[route.SourceCityID] || !cities[route.DestinationCityID]) {
			return errors.New("trade route city does not exist")
		}
	}
	return nil
}

func validateTurnPolicyState(policy TurnPolicyState, players map[string]bool) error {
	if policy.Stage != "" && policy.Stage != turns.StagePlanning && policy.Stage != turns.StageConflict {
		return fmt.Errorf("unsupported turn policy stage %q", policy.Stage)
	}
	if policy.ComponentIndex < 0 || policy.Recalculations < 0 {
		return errors.New("turn policy counters must be nonnegative")
	}
	if len(policy.Components) == 0 && policy.ComponentIndex != 0 {
		return errors.New("turn policy component index requires components")
	}
	if len(policy.Components) > 0 && policy.ComponentIndex >= len(policy.Components) {
		return errors.New("turn policy component index is out of range")
	}
	if err := uniqueIDs("turn policy early end", policy.EarlyEnds); err != nil {
		return err
	}
	for _, playerID := range policy.EarlyEnds {
		if !players[playerID] {
			return fmt.Errorf("turn policy early end player %q does not exist", playerID)
		}
	}
	for _, command := range policy.Queue {
		if command.Kind == "" || command.CommandID == "" || command.IdempotencyKey == "" || !json.Valid(command.Payload) || !players[command.ActorID] || command.ReceiptSequence < 0 {
			return errors.New("turn policy queue command is invalid")
		}
		if command.TargetPlayerID != "" && !players[command.TargetPlayerID] {
			return fmt.Errorf("turn policy target player %q does not exist", command.TargetPlayerID)
		}
	}
	if policy.Deadline != nil {
		if policy.Deadline.PlayerID == "" || !players[policy.Deadline.PlayerID] || policy.Deadline.AtUnix < 1 {
			return errors.New("turn policy deadline is invalid")
		}
	}
	for _, war := range policy.Wars {
		if war.A == "" || war.B == "" || war.A == war.B || !players[war.A] || !players[war.B] {
			return errors.New("turn policy war is invalid")
		}
	}
	for _, component := range policy.Components {
		if len(component) == 0 || uniqueIDs("turn policy component", component) != nil {
			return errors.New("turn policy conflict component is invalid")
		}
		for _, playerID := range component {
			if !players[playerID] {
				return fmt.Errorf("turn policy component player %q does not exist", playerID)
			}
		}
	}
	return nil
}

func CheckInvariants(state State) error { return ValidateState(state) }

func CanonicalStateJSON(state State) ([]byte, error) {
	if err := ValidateState(state); err != nil {
		return nil, err
	}
	state = cloneStateForNormalization(state)
	normalizeState(&state)
	return json.Marshal(state)
}

// CanonicalStateHash hashes gameplay authority only; storage audit timestamps are excluded.
func CanonicalStateHash(state State) (string, error) {
	canonical, err := GameplayStateJSON(state)
	if err != nil {
		return "", err
	}
	return hashBytes(canonical), nil
}

// GameplayStateJSON is the stable hash projection of State. It retains every
// authoritative field while explicitly excluding audit-only timestamps.
func GameplayStateJSON(state State) ([]byte, error) {
	if err := ValidateState(state); err != nil {
		return nil, err
	}
	state = cloneStateForNormalization(state)
	normalizeState(&state)
	state.Audit = Audit{}
	return json.Marshal(state)
}

// GameplayEventsJSON is the stable hash projection of ordered internal events.
func GameplayEventsJSON(events []Event) ([]byte, error) {
	canonical, err := CanonicalEventsJSON(events)
	if err != nil {
		return nil, err
	}
	var projected []Event
	if err := json.Unmarshal(canonical, &projected); err != nil {
		return nil, fmt.Errorf("decode canonical events: %w", err)
	}
	for index := range projected {
		projected[index].AuditAt = time.Time{}
	}
	return json.Marshal(projected)
}

// CanonicalEventsHash hashes event authority only; audit timestamps are excluded.
func CanonicalEventsHash(events []Event) (string, error) {
	canonical, err := GameplayEventsJSON(events)
	if err != nil {
		return "", err
	}
	return hashBytes(canonical), nil
}

// GameplayArtifactHash hashes the complete authoritative state and ordered events.
func GameplayArtifactHash(state State, events []Event) (string, error) {
	stateJSON, err := GameplayStateJSON(state)
	if err != nil {
		return "", err
	}
	eventJSON, err := GameplayEventsJSON(events)
	if err != nil {
		return "", err
	}
	return hashBytes(append(append(stateJSON, '\n'), eventJSON...)), nil
}

func hashBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// DecodeCanonicalStateJSON rejects malformed, unknown, and invariant-violating state bytes.
func DecodeCanonicalStateJSON(data []byte) (State, error) {
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode canonical state: %w", err)
	}
	if err := ValidateState(state); err != nil {
		return State{}, err
	}
	return canonicalClone(state)
}

func CanonicalEventsJSON(events []Event) ([]byte, error) {
	copied := append([]Event(nil), events...)
	for index := range copied {
		if copied[index].SchemaVersion != EventSchemaV1 || copied[index].ID == "" || copied[index].Sequence <= 0 || copied[index].Ordinal < 0 {
			return nil, errors.New("invalid event")
		}
		copied[index].Payload = canonicalPayload(copied[index].Payload)
		if index > 0 && (copied[index-1].Sequence > copied[index].Sequence || (copied[index-1].Sequence == copied[index].Sequence && copied[index-1].Ordinal >= copied[index].Ordinal)) {
			return nil, errors.New("events must be in stable sequence order")
		}
	}
	return json.Marshal(copied)
}

func canonicalClone(state State) (State, error) {
	canonical, err := CanonicalStateJSON(state)
	if err != nil {
		return State{}, err
	}
	var cloned State
	if err := json.Unmarshal(canonical, &cloned); err != nil {
		return State{}, err
	}
	return cloned, nil
}

func cloneStateForNormalization(state State) State {
	state.Overlays = append([]TileOverlay(nil), state.Overlays...)
	state.Tiles = append([]Tile(nil), state.Tiles...)
	state.Units = append([]Unit(nil), state.Units...)
	for index := range state.Units {
		state.Units[index].CargoDomains = append([]string(nil), state.Units[index].CargoDomains...)
		state.Units[index].Effects = append([]Effect(nil), state.Units[index].Effects...)
	}
	state.Cities = append([]City(nil), state.Cities...)
	for index := range state.Cities {
		state.Cities[index].WorkedTileIDs = append([]string(nil), state.Cities[index].WorkedTileIDs...)
		state.Cities[index].Queue = append([]ProductionItem(nil), state.Cities[index].Queue...)
		state.Cities[index].Buildings = append([]Building(nil), state.Cities[index].Buildings...)
		state.Cities[index].Obsolete = append([]string(nil), state.Cities[index].Obsolete...)
		state.Cities[index].Effects = append([]Effect(nil), state.Cities[index].Effects...)
		for item := range state.Cities[index].Queue {
			state.Cities[index].Queue[item].Effects = append([]Effect(nil), state.Cities[index].Queue[item].Effects...)
		}
		for building := range state.Cities[index].Buildings {
			state.Cities[index].Buildings[building].Effects = append([]Effect(nil), state.Cities[index].Buildings[building].Effects...)
		}
	}
	state.Armies = append([]Army(nil), state.Armies...)
	for index := range state.Armies {
		state.Armies[index].UnitIDs = append([]string(nil), state.Armies[index].UnitIDs...)
	}
	state.Nations = append([]Nation(nil), state.Nations...)
	for index := range state.Nations {
		state.Nations[index].Effects = append([]Effect(nil), state.Nations[index].Effects...)
		state.Nations[index].Advances = append([]string(nil), state.Nations[index].Advances...)
		state.Nations[index].Embassies = append([]string(nil), state.Nations[index].Embassies...)
		state.Nations[index].CityReports = append([]CityReport(nil), state.Nations[index].CityReports...)
		for report := range state.Nations[index].CityReports {
			state.Nations[index].CityReports[report].Buildings = append([]string(nil), state.Nations[index].CityReports[report].Buildings...)
			state.Nations[index].CityReports[report].Effects = append([]Effect(nil), state.Nations[index].CityReports[report].Effects...)
		}
	}
	state.Wonders = append([]Wonder(nil), state.Wonders...)
	for index := range state.Wonders {
		state.Wonders[index].Effects = append([]Effect(nil), state.Wonders[index].Effects...)
	}
	state.Diplomacy = append([]DiplomaticRelation(nil), state.Diplomacy...)
	state.TradeRoutes = append([]TradeRoute(nil), state.TradeRoutes...)
	state.Proposals = append([]DiplomaticProposal(nil), state.Proposals...)
	state.Players = append([]Player(nil), state.Players...)
	state.Phase.ActivePlayerOrder = append([]string(nil), state.Phase.ActivePlayerOrder...)
	state.Policy.Queue = append([]TurnPolicyCommand(nil), state.Policy.Queue...)
	for index := range state.Policy.Queue {
		state.Policy.Queue[index].Payload = append(json.RawMessage(nil), state.Policy.Queue[index].Payload...)
	}
	state.Policy.Wars = append([]TurnPolicyWar(nil), state.Policy.Wars...)
	state.Policy.EarlyEnds = append([]string(nil), state.Policy.EarlyEnds...)
	state.Policy.Components = append([][]string(nil), state.Policy.Components...)
	for index := range state.Policy.Components {
		state.Policy.Components[index] = append([]string(nil), state.Policy.Components[index]...)
	}
	if state.Policy.Deadline != nil {
		deadline := *state.Policy.Deadline
		state.Policy.Deadline = &deadline
	}
	state.Fog = append([]FogObservation(nil), state.Fog...)
	state.RNG = append([]RNGDomainState(nil), state.RNG...)
	state.Deadlines = append([]Deadline(nil), state.Deadlines...)
	return state
}

func normalizeState(state *State) {
	sort.Slice(state.Overlays, func(i, j int) bool { return state.Overlays[i].ID < state.Overlays[j].ID })
	sort.Slice(state.Tiles, func(i, j int) bool { return state.Tiles[i].ID < state.Tiles[j].ID })
	sort.Slice(state.Units, func(i, j int) bool { return state.Units[i].ID < state.Units[j].ID })
	sort.Slice(state.Cities, func(i, j int) bool { return state.Cities[i].ID < state.Cities[j].ID })
	sort.Slice(state.Armies, func(i, j int) bool { return state.Armies[i].ID < state.Armies[j].ID })
	sort.Slice(state.Nations, func(i, j int) bool { return state.Nations[i].ID < state.Nations[j].ID })
	sort.Slice(state.Wonders, func(i, j int) bool { return state.Wonders[i].ID < state.Wonders[j].ID })
	sort.Slice(state.Diplomacy, func(i, j int) bool { return state.Diplomacy[i].ID < state.Diplomacy[j].ID })
	sort.Slice(state.TradeRoutes, func(i, j int) bool { return state.TradeRoutes[i].ID < state.TradeRoutes[j].ID })
	sort.Slice(state.Proposals, func(i, j int) bool { return state.Proposals[i].ID < state.Proposals[j].ID })
	sort.Slice(state.Players, func(i, j int) bool { return state.Players[i].ID < state.Players[j].ID })
	sort.Slice(state.Fog, func(i, j int) bool {
		if state.Fog[i].PlayerID == state.Fog[j].PlayerID {
			return state.Fog[i].TileID < state.Fog[j].TileID
		}
		return state.Fog[i].PlayerID < state.Fog[j].PlayerID
	})
	sort.Slice(state.RNG, func(i, j int) bool { return state.RNG[i].Domain < state.RNG[j].Domain })
	sort.Slice(state.Deadlines, func(i, j int) bool { return state.Deadlines[i].ID < state.Deadlines[j].ID })
	sort.Strings(state.Policy.EarlyEnds)
	sort.Slice(state.Policy.Queue, func(i, j int) bool {
		left, right := state.Policy.Queue[i], state.Policy.Queue[j]
		if left.ReceiptSequence != right.ReceiptSequence {
			return left.ReceiptSequence < right.ReceiptSequence
		}
		if left.ActorID != right.ActorID {
			return left.ActorID < right.ActorID
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		return left.UnitOrArmyID < right.UnitOrArmyID
	})
	for index := range state.Policy.Components {
		sort.Strings(state.Policy.Components[index])
	}
	for index := range state.Policy.Wars {
		if state.Policy.Wars[index].B < state.Policy.Wars[index].A {
			state.Policy.Wars[index].A, state.Policy.Wars[index].B = state.Policy.Wars[index].B, state.Policy.Wars[index].A
		}
	}
	sort.Slice(state.Policy.Wars, func(i, j int) bool {
		if state.Policy.Wars[i].A != state.Policy.Wars[j].A {
			return state.Policy.Wars[i].A < state.Policy.Wars[j].A
		}
		return state.Policy.Wars[i].B < state.Policy.Wars[j].B
	})
	for i := range state.Units {
		normalizeEffects(state.Units[i].Effects)
		sort.Strings(state.Units[i].CargoDomains)
	}
	for i := range state.Cities {
		normalizeEffects(state.Cities[i].Effects)
		sort.Strings(state.Cities[i].WorkedTileIDs)
		sort.Strings(state.Cities[i].Obsolete)
		for item := range state.Cities[i].Queue {
			normalizeEffects(state.Cities[i].Queue[item].Effects)
		}
		for building := range state.Cities[i].Buildings {
			normalizeEffects(state.Cities[i].Buildings[building].Effects)
		}
	}
	for i := range state.Wonders {
		normalizeEffects(state.Wonders[i].Effects)
	}
	for i := range state.Fog {
		normalizeEffects(state.Fog[i].CityEffects)
	}
	for i := range state.Nations {
		normalizeEffects(state.Nations[i].Effects)
		sort.Strings(state.Nations[i].Advances)
		sort.Strings(state.Nations[i].Embassies)
		for report := range state.Nations[i].CityReports {
			normalizeEffects(state.Nations[i].CityReports[report].Effects)
			sort.Strings(state.Nations[i].CityReports[report].Buildings)
		}
		sort.Slice(state.Nations[i].CityReports, func(left, right int) bool {
			return state.Nations[i].CityReports[left].CityID < state.Nations[i].CityReports[right].CityID
		})
	}
	for i := range state.Armies {
		sort.Strings(state.Armies[i].UnitIDs)
	}
}

func normalizeEffects(effects []Effect) {
	sort.Slice(effects, func(i, j int) bool { return effects[i].ID < effects[j].ID })
}
func validateRulesPin(pin RulesPin) error {
	if strings.TrimSpace(pin.Version) == "" || len(pin.ContentHash) != sha256.Size*2 {
		return errors.New("rules pin requires version and SHA-256 content hash")
	}
	if _, err := hex.DecodeString(pin.ContentHash); err != nil {
		return errors.New("rules pin content hash must be hexadecimal")
	}
	if len(pin.Definitions) > 0 && string(pin.Definitions) != "null" {
		canonical, err := rules.CanonicalDefinitions(pin.Definitions)
		if err != nil {
			return errors.New("rules pin definitions must be canonical JSON")
		}
		digest := sha256.Sum256(canonical)
		if hex.EncodeToString(digest[:]) != pin.ContentHash {
			return errors.New("rules pin definitions do not match content hash")
		}
	}
	return nil
}

func sameRulesPin(left, right RulesPin) bool {
	if left.Version != right.Version || left.ContentHash != right.ContentHash {
		return false
	}
	if normalizedDefinitions(left.Definitions) == "" && normalizedDefinitions(right.Definitions) == "" {
		return true
	}
	leftDefinitions, leftErr := rules.CanonicalDefinitions([]byte(normalizedDefinitions(left.Definitions)))
	rightDefinitions, rightErr := rules.CanonicalDefinitions([]byte(normalizedDefinitions(right.Definitions)))
	return leftErr == nil && rightErr == nil && string(leftDefinitions) == string(rightDefinitions)
}

func hasRuntimeDefinitions(definitions json.RawMessage) bool {
	if len(definitions) == 0 || string(definitions) == "null" {
		return false
	}
	var header struct {
		CompatibilityVersion string `json:"compatibility_version"`
	}
	return json.Unmarshal(definitions, &header) == nil && header.CompatibilityVersion != ""
}

func normalizedDefinitions(definitions json.RawMessage) string {
	if len(definitions) == 0 || string(definitions) == "null" {
		return ""
	}
	return string(definitions)
}
func validateCommand(command Command) error {
	if command.SchemaVersion != CommandSchemaV1 {
		return fmt.Errorf("%w: %d", ErrUnknownCommandSchema, command.SchemaVersion)
	}
	if command.ID == "" || command.ActorPlayerID == "" || command.Type == "" {
		return errors.New("command id, actor player id, and type are required")
	}
	if len(command.Payload) > 0 && !json.Valid(command.Payload) {
		return errors.New("command payload must be valid JSON")
	}
	return nil
}
func validateSources(sources Sources) error {
	if sources.Clock == nil || sources.IDs == nil || sources.RNG == nil {
		return errors.New("clock, ID, and RNG sources are required")
	}
	return nil
}
func validateRNG(entries []RNGDomainState) error {
	if len(entries) != 3 {
		return errors.New("all named RNG domains are required")
	}
	seen := makeIDSet(nil)
	for _, entry := range entries {
		if entry.Domain != RNGDomainCombat && entry.Domain != RNGDomainSpecialAction && entry.Domain != RNGDomainAI {
			return fmt.Errorf("unknown RNG domain %q", entry.Domain)
		}
		if seen[entry.Domain] {
			return fmt.Errorf("duplicate RNG domain %q", entry.Domain)
		}
		seen[entry.Domain] = true
	}
	for _, domain := range []string{RNGDomainCombat, RNGDomainSpecialAction, RNGDomainAI} {
		if !seen[domain] {
			return fmt.Errorf("missing RNG domain %q", domain)
		}
	}
	return nil
}
func canonicalPayload(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return nil
	}
	var value any
	if json.Unmarshal(payload, &value) != nil {
		return append(json.RawMessage(nil), payload...)
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return append(json.RawMessage(nil), payload...)
	}
	return canonical
}
func uniqueIDs(kind string, ids []string) error {
	seen := makeIDSet(nil)
	for _, id := range ids {
		if id == "" {
			return fmt.Errorf("%s id is required", kind)
		}
		if seen[id] {
			return fmt.Errorf("duplicate %s id %q", kind, id)
		}
		seen[id] = true
	}
	return nil
}
func makeIDSet(ids []string) map[string]bool {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}
func overlayIDs(items []TileOverlay) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
func unitIDs(items []Unit) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
func cityIDs(items []City) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
func armyIDs(items []Army) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
func nationIDs(items []Nation) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
func playerIDs(items []Player) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
func relationIDs(items []DiplomaticRelation) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}

func tradeRouteIDs(items []TradeRoute) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}

func proposalIDs(items []DiplomaticProposal) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}

func deadlineIDs(items []Deadline) []string {
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
