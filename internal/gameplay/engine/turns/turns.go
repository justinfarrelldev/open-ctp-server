// Package turns contains the deterministic, persistence-friendly state machines for gameplay turns.
package turns

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"
)

const ContractVersion = "turn-policy-v1"

const (
	ModeSequential   = "sequential"
	ModeLongTurn     = "long_turn"
	ModeSimultaneous = "simultaneous"
	ModeHybrid       = "hybrid"
	StagePlanning    = "planning"
	StageConflict    = "conflict"
	CommandAction    = "action"
	CommandPass      = "pass"
	CommandEarlyEnd  = "early_end"
)

type Player struct {
	ID       string
	Eligible bool
}
type War struct {
	A string
	B string
}
type Deadline struct {
	PlayerID string
	AtUnix   int64
}
type Command struct {
	Kind            string
	ActorID         string
	ActionClass     string
	UnitOrArmyID    string
	ReceiptSequence int64
	Hostile         bool
	TargetPlayerID  string
	CommandID       string
	IdempotencyKey  string
	Payload         json.RawMessage
}
type State struct {
	Mode             string
	Version          string
	Turn             int
	ActiveOrder      []string
	ActiveIndex      int
	Players          []Player
	Deadline         *Deadline
	ResolutionLocked bool
	Queue            []Command
	EarlyEnds        []string
	Wars             []War
	Stage            string
	Components       [][]string
	ComponentIndex   int
	Recalculations   int
}
type Result struct {
	Immediate        bool
	Resolved         bool
	SystemPass       bool
	ResolvedCommands []Command
}
type TurnPolicy interface {
	Mode() string
	Start(State, time.Time) (State, Result, error)
	Apply(State, Command, time.Time) (State, Result, error)
	Expire(State, time.Time) (State, Result, error)
}
type policy struct {
	mode     string
	deadline time.Duration
}
type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string { return e.Message }
func Status(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode
	}
	return 0
}

func New(mode string, deadline time.Duration) (TurnPolicy, error) {
	if !SupportedMode(mode) {
		return nil, fmt.Errorf("unsupported turn mode %q", mode)
	}
	if mode == ModeLongTurn || mode == ModeSimultaneous || mode == ModeHybrid {
		if deadline <= 0 {
			return nil, errors.New("turn deadline must be positive")
		}
	}
	return policy{mode: mode, deadline: deadline}, nil
}
func SupportedMode(mode string) bool {
	return mode == ModeSequential || mode == ModeLongTurn || mode == ModeSimultaneous || mode == ModeHybrid
}
func (p policy) Mode() string { return p.mode }
func (s State) ActivePlayer() string {
	if s.ActiveIndex < 0 || s.ActiveIndex >= len(s.ActiveOrder) {
		return ""
	}
	return s.ActiveOrder[s.ActiveIndex]
}
func (s State) HasEarlyEnd(id string) bool {
	for _, v := range s.EarlyEnds {
		if v == id {
			return true
		}
	}
	return false
}

func (p policy) Start(s State, now time.Time) (State, Result, error) {
	if err := p.validate(s); err != nil {
		return State{}, Result{}, err
	}
	if s.Stage == "" {
		s.Stage = StagePlanning
	}
	if p.needsDeadline() && s.Deadline == nil {
		p.setDeadline(&s, now)
	}
	return s, Result{}, nil
}
func (p policy) Apply(s State, c Command, now time.Time) (State, Result, error) {
	if err := p.validate(s); err != nil {
		return State{}, Result{}, err
	}
	if s.ResolutionLocked {
		return State{}, Result{}, &Error{409, "phase resolution is locked"}
	}
	if !s.eligible(c.ActorID) {
		return State{}, Result{}, &Error{422, "player is not eligible for this phase"}
	}
	switch p.mode {
	case ModeSequential, ModeLongTurn:
		return p.applySequential(s, c, now)
	case ModeSimultaneous:
		return p.applyPlanning(s, c, now, false)
	case ModeHybrid:
		if s.Stage == StageConflict {
			return p.applyConflict(s, c, now)
		}
		return p.applyPlanning(s, c, now, true)
	}
	return State{}, Result{}, fmt.Errorf("unsupported turn mode %q", p.mode)
}
func (p policy) Expire(s State, now time.Time) (State, Result, error) {
	if err := p.validate(s); err != nil {
		return State{}, Result{}, err
	}
	if s.Deadline == nil || now.Unix() < s.Deadline.AtUnix {
		return s, Result{}, nil
	}
	if p.mode == ModeLongTurn {
		c := Command{Kind: CommandPass, ActorID: s.Deadline.PlayerID}
		next, result, err := p.applySequential(s, c, now)
		result.SystemPass = err == nil
		return next, result, err
	}
	if p.mode == ModeSimultaneous || (p.mode == ModeHybrid && s.Stage == StagePlanning) {
		return p.resolvePlanning(s, now, p.mode == ModeHybrid)
	}
	return s, Result{}, nil
}
func (p policy) applySequential(s State, c Command, now time.Time) (State, Result, error) {
	if c.ActorID != s.ActivePlayer() {
		return State{}, Result{}, &Error{422, "only the active player may act"}
	}
	switch c.Kind {
	case CommandAction:
		return s, Result{Immediate: true}, nil
	case CommandPass:
		resolved := p.advance(&s, now)
		return s, Result{Resolved: resolved}, nil
	default:
		return State{}, Result{}, &Error{422, "command is not valid in a sequential turn"}
	}
}
func (p policy) applyPlanning(s State, c Command, now time.Time, hybrid bool) (State, Result, error) {
	if c.Kind == CommandEarlyEnd {
		if !s.HasEarlyEnd(c.ActorID) {
			s.EarlyEnds = append(s.EarlyEnds, c.ActorID)
			sort.Strings(s.EarlyEnds)
		}
		if s.allEligibleEarlyEnded() {
			return p.resolvePlanning(s, now, hybrid)
		}
		return s, Result{}, nil
	}
	if c.Kind != CommandAction {
		return State{}, Result{}, &Error{422, "only planning actions or early end are valid"}
	}
	if hybrid && c.Hostile {
		return State{}, Result{}, &Error{422, "hostile actions are valid only in conflict subturns"}
	}
	s.clearEarlyEnd(c.ActorID)
	s.Queue = append(s.Queue, c)
	return s, Result{}, nil
}
func (p policy) resolvePlanning(s State, now time.Time, hybrid bool) (State, Result, error) {
	if s.ResolutionLocked {
		return s, Result{}, nil
	}
	s.ResolutionLocked = true
	commands := append([]Command(nil), s.Queue...)
	sortCommands(commands)
	s.Queue = nil
	s.EarlyEnds = nil
	if hybrid {
		s.Components = warComponents(s)
		s.ComponentIndex = 0
		if len(s.Components) > 0 {
			s.Stage = StageConflict
			s.ActiveIndex = indexOf(s.ActiveOrder, s.Components[0][0])
			s.ResolutionLocked = false
			return s, Result{Resolved: true, ResolvedCommands: commands}, nil
		}
	}
	s.finishRound(now)
	return s, Result{Resolved: true, ResolvedCommands: commands}, nil
}
func (p policy) applyConflict(s State, c Command, now time.Time) (State, Result, error) {
	component := s.currentComponent()
	if len(component) == 0 {
		return State{}, Result{}, &Error{422, "no active conflict component"}
	}
	if c.ActorID != s.ActivePlayer() || !contains(component, c.ActorID) {
		return State{}, Result{}, &Error{422, "only the active component player may act"}
	}
	if c.Kind == CommandAction {
		if c.Hostile && !contains(component, c.TargetPlayerID) {
			return State{}, Result{}, &Error{422, "hostile action crosses conflict components"}
		}
		return s, Result{Immediate: true}, nil
	}
	if c.Kind != CommandPass {
		return State{}, Result{}, &Error{422, "command is not valid in a conflict subturn"}
	}
	next := nextIn(component, s.ActiveOrder, c.ActorID)
	if next != "" {
		s.ActiveIndex = indexOf(s.ActiveOrder, next)
		return s, Result{}, nil
	}
	s.ComponentIndex++
	if s.ComponentIndex < len(s.Components) {
		s.ActiveIndex = indexOf(s.ActiveOrder, s.Components[s.ComponentIndex][0])
		return s, Result{}, nil
	}
	s.finishRound(now)
	return s, Result{Resolved: true}, nil
}
func (p policy) validate(s State) error {
	if s.Mode != p.mode || s.Version != ContractVersion {
		return errors.New("immutable turn mode contract mismatch")
	}
	if s.Turn < 1 || len(s.ActiveOrder) == 0 || s.ActiveIndex < 0 || s.ActiveIndex >= len(s.ActiveOrder) {
		return errors.New("invalid turn state")
	}
	for _, id := range s.ActiveOrder {
		if !s.eligible(id) {
			continue
		}
	}
	return nil
}
func (p policy) needsDeadline() bool { return p.mode != ModeSequential }
func (p policy) setDeadline(s *State, now time.Time) {
	s.Deadline = &Deadline{PlayerID: s.ActivePlayer(), AtUnix: now.UTC().Add(p.deadline).Unix()}
}
func (p policy) advance(s *State, now time.Time) bool {
	for i := 1; i <= len(s.ActiveOrder); i++ {
		idx := (s.ActiveIndex + i) % len(s.ActiveOrder)
		if s.eligible(s.ActiveOrder[idx]) {
			s.ActiveIndex = idx
			if idx == 0 {
				s.finishRound(now)
				return true
			}
			if p.mode == ModeLongTurn {
				p.setDeadline(s, now)
			}
			return false
		}
	}
	s.finishRound(now)
	return true
}
func (s *State) finishRound(_ time.Time) {
	s.Turn++
	s.Recalculations++
	s.ActiveIndex = firstEligibleIndex(*s)
	s.Stage = StagePlanning
	s.Components = nil
	s.ComponentIndex = 0
	s.Deadline = nil
	s.ResolutionLocked = false
}
func (s State) eligible(id string) bool {
	for _, p := range s.Players {
		if p.ID == id {
			return p.Eligible
		}
	}
	return false
}
func (s *State) clearEarlyEnd(id string) {
	out := s.EarlyEnds[:0]
	for _, value := range s.EarlyEnds {
		if value != id {
			out = append(out, value)
		}
	}
	s.EarlyEnds = out
}
func (s State) allEligibleEarlyEnded() bool {
	for _, p := range s.Players {
		if p.Eligible && !s.HasEarlyEnd(p.ID) {
			return false
		}
	}
	return true
}
func (s State) currentComponent() []string {
	if s.ComponentIndex < 0 || s.ComponentIndex >= len(s.Components) {
		return nil
	}
	return s.Components[s.ComponentIndex]
}
func sortCommands(commands []Command) {
	sort.SliceStable(commands, func(i, j int) bool {
		a, b := commands[i], commands[j]
		if a.ActionClass != b.ActionClass {
			return actionPriority(a.ActionClass) < actionPriority(b.ActionClass)
		}
		if a.ActorID != b.ActorID {
			return a.ActorID < b.ActorID
		}
		if a.UnitOrArmyID != b.UnitOrArmyID {
			return a.UnitOrArmyID < b.UnitOrArmyID
		}
		return a.ReceiptSequence < b.ReceiptSequence
	})
}
func actionPriority(class string) int {
	switch class {
	case "economic":
		return 0
	case "non_hostile":
		return 1
	case "hostile":
		return 2
	default:
		return 3
	}
}
func warComponents(s State) [][]string {
	adj := map[string][]string{}
	for _, w := range s.Wars {
		if s.eligible(w.A) && s.eligible(w.B) {
			adj[w.A] = append(adj[w.A], w.B)
			adj[w.B] = append(adj[w.B], w.A)
		}
	}
	seen := map[string]bool{}
	var components [][]string
	for _, root := range s.ActiveOrder {
		if seen[root] || len(adj[root]) == 0 {
			continue
		}
		stack := []string{root}
		seen[root] = true
		var component []string
		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			component = append(component, n)
			for _, to := range adj[n] {
				if !seen[to] {
					seen[to] = true
					stack = append(stack, to)
				}
			}
		}
		sort.SliceStable(component, func(i, j int) bool {
			return indexOf(s.ActiveOrder, component[i]) < indexOf(s.ActiveOrder, component[j])
		})
		components = append(components, component)
	}
	sort.SliceStable(components, func(i, j int) bool { return smallest(components[i]) < smallest(components[j]) })
	return components
}
func nextIn(component, order []string, current string) string {
	idx := indexOf(component, current)
	for i := idx + 1; i < len(component); i++ {
		if contains(order, component[i]) {
			return component[i]
		}
	}
	return ""
}
func firstEligibleIndex(s State) int {
	for i, id := range s.ActiveOrder {
		if s.eligible(id) {
			return i
		}
	}
	return 0
}
func indexOf(values []string, target string) int {
	for i, v := range values {
		if v == target {
			return i
		}
	}
	return -1
}
func contains(values []string, target string) bool { return indexOf(values, target) >= 0 }
func smallest(values []string) string {
	if len(values) == 0 {
		return ""
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}
