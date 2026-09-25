package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

type fixedClock struct{ value time.Time }

func (c fixedClock) Now() time.Time { return c.value }

type sequenceIDs struct{ next int }

func (s *sequenceIDs) Next() string {
	s.next++
	return "event-" + string(rune('0'+s.next))
}

type deterministicRNG struct{}

func (deterministicRNG) Next(domain string, state uint64) (uint64, uint64) {
	if domain == "combat" {
		return state + 1, state + 11
	}
	if domain == "special-action" {
		return state + 2, state + 22
	}
	return state + 3, state + 33
}

func fixtureState() State {
	return State{
		SchemaVersion: StateSchemaV1,
		Rules: RulesPin{
			Version:     "ctp2-supported-v2",
			ContentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
		Phase: PhaseState{Mode: "sequential", ModeVersion: "turn-policy-v1", Turn: 1, ActivePlayerIndex: 0, ActivePlayerOrder: []string{"player-a", "player-b"}},
		Players: []Player{
			{ID: "player-b", NationID: "nation-b", Lifecycle: PlayerActive, Controller: ControllerHuman},
			{ID: "player-a", NationID: "nation-a", Lifecycle: PlayerActive, Controller: ControllerHuman},
		},
		RNG: []RNGDomainState{
			{Domain: RNGDomainAI, State: 3},
			{Domain: RNGDomainCombat, State: 1},
			{Domain: RNGDomainRiot, State: 4},
			{Domain: RNGDomainSpecialAction, State: 2},
		},
	}
}

func fixtureSources() Sources {
	return Sources{
		Clock: fixedClock{value: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)},
		IDs:   &sequenceIDs{},
		RNG:   deterministicRNG{},
	}
}

func fixtureCommand() Command {
	return Command{SchemaVersion: CommandSchemaV1, ID: "command-1", ActorPlayerID: "player-a", Type: CommandNoop}
}

func artifactHash(t *testing.T, state State, events []Event) string {
	t.Helper()
	stateJSON, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	eventJSON, err := CanonicalEventsJSON(events)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(append(stateJSON, eventJSON...))
	return hex.EncodeToString(digest[:])
}

func TestCanonicalReplayHasOneArtifactHash(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 1000; i++ {
		next, events, _, err := Apply(fixtureState(), fixtureCommand(), fixtureState().Rules, fixtureSources())
		if err != nil {
			t.Fatal(err)
		}
		if err := CheckInvariants(next); err != nil {
			t.Fatal(err)
		}
		seen[artifactHash(t, next, events)] = struct{}{}
	}
	if len(seen) != 1 {
		t.Fatalf("identical replays produced %d artifacts", len(seen))
	}
}

func TestCanonicalSortingIsIndependentOfInputOrder(t *testing.T) {
	first := fixtureState()
	second := fixtureState()
	second.Players[0], second.Players[1] = second.Players[1], second.Players[0]
	second.RNG[0], second.RNG[2] = second.RNG[2], second.RNG[0]

	firstNext, firstEvents, _, err := Apply(first, fixtureCommand(), first.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	secondNext, secondEvents, _, err := Apply(second, fixtureCommand(), second.Rules, fixtureSources())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := artifactHash(t, secondNext, secondEvents), artifactHash(t, firstNext, firstEvents); got != want {
		t.Fatalf("permuted inputs differ: got %s want %s", got, want)
	}
}

func TestSchemaAndInvariantRejections(t *testing.T) {
	tests := []struct {
		name  string
		state State
	}{
		{name: "unknown schema", state: func() State { s := richState(); s.SchemaVersion = 99; return s }()},
		{name: "duplicate player", state: func() State { s := richState(); s.Players = append(s.Players, s.Players[0]); return s }()},
		{name: "invalid active player", state: func() State { s := richState(); s.Phase.ActivePlayerIndex = 2; return s }()},
		{name: "missing rng domain", state: func() State { s := richState(); s.RNG = s.RNG[:2]; return s }()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, _, err := Apply(tt.state, fixtureCommand(), tt.state.Rules, fixtureSources()); err == nil {
				t.Fatal("Apply accepted invalid state")
			}
		})
	}
}

func TestPhaseModeContractAndDeadlineReferenceValidation(t *testing.T) {
	state := fixtureState()
	state.Phase.Mode = "unknown"
	if err := ValidateState(state); err == nil {
		t.Fatal("unknown phase mode accepted")
	}
	state = fixtureState()
	state.Phase.ModeVersion = "turn-policy-v0"
	if err := ValidateState(state); err == nil {
		t.Fatal("unknown phase mode version accepted")
	}
	state = fixtureState()
	state.Phase.BaseRevision = 1
	if err := ValidateState(state); err == nil {
		t.Fatal("future phase base revision accepted")
	}
	state = fixtureState()
	state.Phase.DeadlineID = "deadline-1"
	if err := ValidateState(state); err == nil {
		t.Fatal("missing phase deadline accepted")
	}
	state.Deadlines = []Deadline{{ID: "deadline-1", Phase: "planning", PlayerID: "player-a", AtUnix: 1}}
	if err := ValidateState(state); err != nil {
		t.Fatalf("valid phase deadline rejected: %v", err)
	}
}

func TestApplyRejectsRulesPinAndCommandSchemaWithoutMutation(t *testing.T) {
	state := fixtureState()
	wrongRules := state.Rules
	wrongRules.ContentHash = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if _, events, _, err := Apply(state, fixtureCommand(), wrongRules, fixtureSources()); err == nil || events != nil {
		t.Fatal("mismatched rules pin was accepted or emitted events")
	}
	command := fixtureCommand()
	command.SchemaVersion++
	if _, events, _, err := Apply(state, command, state.Rules, fixtureSources()); err == nil || events != nil {
		t.Fatal("unknown command schema was accepted or emitted events")
	}
}

func TestDecodeCanonicalStateRejectsCorruptAndUnknownBytes(t *testing.T) {
	if _, err := DecodeCanonicalStateJSON([]byte(`{`)); err == nil {
		t.Fatal("corrupt JSON was accepted")
	}
	if _, err := DecodeCanonicalStateJSON([]byte(`{"schema_version":99}`)); err == nil {
		t.Fatal("unknown state schema was accepted")
	}
	canonical, err := CanonicalStateJSON(fixtureState())
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCanonicalStateJSON(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := artifactHash(t, decoded, nil), artifactHash(t, fixtureState(), nil); got != want {
		t.Fatalf("decoded state differs: got %s want %s", got, want)
	}
}

func TestInjectedClockAndIDAreAuditOnly(t *testing.T) {
	firstSources := fixtureSources()
	secondSources := fixtureSources()
	first, firstEvents, _, err := Apply(fixtureState(), fixtureCommand(), fixtureState().Rules, firstSources)
	if err != nil {
		t.Fatal(err)
	}
	second, secondEvents, _, err := Apply(fixtureState(), fixtureCommand(), fixtureState().Rules, secondSources)
	if err != nil {
		t.Fatal(err)
	}
	if firstEvents[0].ID != "event-1" || firstEvents[0].AuditAt.IsZero() {
		t.Fatal("injected ID or clock was not used")
	}
	if got, want := artifactHash(t, first, firstEvents), artifactHash(t, second, secondEvents); got != want {
		t.Fatalf("injected deterministic sources differ: got %s want %s", got, want)
	}
}

func TestGameplayHashesExcludeAuditButStorageRetainsIt(t *testing.T) {
	firstState := fixtureState()
	firstState.Audit.UpdatedAt = time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	secondState := firstState
	secondState.Audit.UpdatedAt = firstState.Audit.UpdatedAt.Add(time.Hour)
	firstEvents := []Event{{SchemaVersion: EventSchemaV1, ID: "event-1", Sequence: 1, Type: "command.accepted", ActorPlayerID: "player-a", CommandID: "command-1", AuditAt: firstState.Audit.UpdatedAt}}
	secondEvents := append([]Event(nil), firstEvents...)
	secondEvents[0].AuditAt = secondState.Audit.UpdatedAt

	firstStorage, err := CanonicalStateJSON(firstState)
	if err != nil {
		t.Fatal(err)
	}
	secondStorage, err := CanonicalStateJSON(secondState)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstStorage) == string(secondStorage) {
		t.Fatal("storage serialization omitted distinct audit timestamps")
	}
	decoded, err := DecodeCanonicalStateJSON(secondStorage)
	if err != nil || !decoded.Audit.UpdatedAt.Equal(secondState.Audit.UpdatedAt) {
		t.Fatalf("storage audit timestamp was not retained: %v, %v", decoded.Audit.UpdatedAt, err)
	}

	firstEventStorage, err := CanonicalEventsJSON(firstEvents)
	if err != nil {
		t.Fatal(err)
	}
	secondEventStorage, err := CanonicalEventsJSON(secondEvents)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstEventStorage) == string(secondEventStorage) {
		t.Fatal("event storage serialization omitted distinct audit timestamps")
	}
	firstStateHash, err := CanonicalStateHash(firstState)
	if err != nil {
		t.Fatal(err)
	}
	secondStateHash, err := CanonicalStateHash(secondState)
	if err != nil {
		t.Fatal(err)
	}
	if firstStateHash != secondStateHash {
		t.Fatal("audit-only change altered gameplay state hash")
	}
	firstEventHash, err := CanonicalEventsHash(firstEvents)
	if err != nil {
		t.Fatal(err)
	}
	secondEventHash, err := CanonicalEventsHash(secondEvents)
	if err != nil {
		t.Fatal(err)
	}
	if firstEventHash != secondEventHash {
		t.Fatal("audit-only change altered gameplay event hash")
	}
	firstHash, err := GameplayArtifactHash(firstState, firstEvents)
	if err != nil {
		t.Fatal(err)
	}
	secondHash, err := GameplayArtifactHash(secondState, secondEvents)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != secondHash {
		t.Fatalf("audit-only change altered gameplay artifact: %s != %s", firstHash, secondHash)
	}
}

func TestGameplayHashesIncludeAuthoritativeFields(t *testing.T) {
	state := fixtureState()
	events := []Event{{SchemaVersion: EventSchemaV1, ID: "event-1", Sequence: 1, Type: "command.accepted", ActorPlayerID: "player-a", CommandID: "command-1"}}
	baseline, err := GameplayArtifactHash(state, events)
	if err != nil {
		t.Fatal(err)
	}
	changedState := state
	changedState.Revision++
	changedStateHash, err := GameplayArtifactHash(changedState, events)
	if err != nil {
		t.Fatal(err)
	}
	if baseline == changedStateHash {
		t.Fatal("authoritative state revision did not alter gameplay artifact")
	}
	changedEvents := append([]Event(nil), events...)
	changedEvents[0].Type = "other.event"
	changedEventHash, err := GameplayArtifactHash(state, changedEvents)
	if err != nil {
		t.Fatal(err)
	}
	if baseline == changedEventHash {
		t.Fatal("authoritative event type did not alter gameplay artifact")
	}
}

func TestRNGDomainsAreNamedAndSerializable(t *testing.T) {
	state := fixtureState()
	canonical, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) == "" {
		t.Fatal("empty canonical state")
	}
	for _, domain := range []string{RNGDomainCombat, RNGDomainSpecialAction, RNGDomainAI, RNGDomainRiot} {
		if _, ok := state.RNGState(domain); !ok {
			t.Fatalf("missing named RNG domain %q", domain)
		}
	}
	if _, ok := state.RNGState("unknown"); ok {
		t.Fatal("unknown RNG domain was found")
	}
}

func richState() State {
	state := fixtureState()
	state.Overlays = []TileOverlay{{ID: "overlay-b", TileID: "tile-b", Kind: "improvement", ImprovementID: "road", OwnerID: "player-b"}, {ID: "overlay-a", TileID: "tile-a", Kind: "improvement", ImprovementID: "farm", OwnerID: "player-a"}}
	state.Units = []Unit{{ID: "unit-b", OwnerID: "player-b", TileID: "tile-b", Effects: []Effect{{ID: "effect-b", Kind: "bonus"}, {ID: "effect-a", Kind: "vision"}}}, {ID: "unit-a", OwnerID: "player-a", TileID: "tile-a"}}
	state.Cities = []City{{ID: "city-b", OwnerID: "player-b", TileID: "tile-b", Effects: []Effect{{ID: "city-effect", Kind: "production"}}}}
	state.Armies = []Army{{ID: "army-b", OwnerID: "player-b", UnitIDs: []string{"unit-b", "unit-a"}}}
	state.Nations = []Nation{{ID: "nation-b", Effects: []Effect{{ID: "nation-effect", Kind: "research"}}}}
	state.Fog = []FogObservation{{PlayerID: "player-b", TileID: "tile-b", Turn: 1}, {PlayerID: "player-a", TileID: "tile-a", Turn: 1}}
	state.Deadlines = []Deadline{{ID: "deadline-b", Phase: "turn", AtUnix: 2}, {ID: "deadline-a", Phase: "turn", AtUnix: 1}}
	return state
}

func TestCanonicalStorageSortsAllUnorderedCollections(t *testing.T) {
	state := richState()
	canonical, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCanonicalStateJSON(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Overlays[0].ID != "overlay-a" || decoded.Units[0].ID != "unit-a" || decoded.Cities[0].ID != "city-b" || decoded.Armies[0].UnitIDs[0] != "unit-a" || decoded.Nations[0].ID != "nation-b" || decoded.Fog[0].PlayerID != "player-a" || decoded.Deadlines[0].ID != "deadline-a" {
		t.Fatalf("unordered collections were not normalized: %+v", decoded)
	}
	if decoded.Units[1].Effects[0].ID != "effect-a" {
		t.Fatal("nested effects were not normalized")
	}
}

func TestValidateStateRejectsEveryAuthorityViolation(t *testing.T) {
	tests := []struct {
		name  string
		state State
	}{
		{"empty rules version", func() State { s := richState(); s.Rules.Version = ""; return s }()},
		{"short rules hash", func() State { s := richState(); s.Rules.ContentHash = "bad"; return s }()},
		{"nonhex rules hash", func() State {
			s := richState()
			s.Rules.ContentHash = "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg"
			return s
		}()},
		{"negative revision", func() State { s := richState(); s.Revision = -1; return s }()},
		{"negative sequence", func() State { s := richState(); s.EventSequence = -1; return s }()},
		{"negative turn", func() State { s := richState(); s.Phase.Turn = -1; return s }()},
		{"negative Gaia completion turn", func() State { s := richState(); s.Nations[0].GaiaCompletionTurn = -1; return s }()},
		{"missing active order", func() State { s := richState(); s.Phase.ActivePlayerOrder = nil; return s }()},
		{"negative active index", func() State { s := richState(); s.Phase.ActivePlayerIndex = -1; return s }()},
		{"duplicate overlay", func() State { s := richState(); s.Overlays = append(s.Overlays, s.Overlays[0]); return s }()},
		{"empty unit id", func() State { s := richState(); s.Units[0].ID = ""; return s }()},
		{"duplicate city", func() State { s := richState(); s.Cities = append(s.Cities, s.Cities[0]); return s }()},
		{"duplicate army", func() State { s := richState(); s.Armies = append(s.Armies, s.Armies[0]); return s }()},
		{"duplicate nation", func() State { s := richState(); s.Nations = append(s.Nations, s.Nations[0]); return s }()},
		{"duplicate deadline", func() State { s := richState(); s.Deadlines = append(s.Deadlines, s.Deadlines[0]); return s }()},
		{"duplicate active player", func() State { s := richState(); s.Phase.ActivePlayerOrder = []string{"player-a", "player-a"}; return s }()},
		{"unknown active player", func() State {
			s := richState()
			s.Phase.ActivePlayerOrder = []string{"player-x"}
			s.Phase.ActivePlayerIndex = 0
			return s
		}()},
		{"invalid lifecycle", func() State { s := richState(); s.Players[0].Lifecycle = "unknown"; return s }()},
		{"invalid controller", func() State { s := richState(); s.Players[0].Controller = "unknown"; return s }()},
		{"wrong rng count", func() State { s := richState(); s.RNG = s.RNG[:2]; return s }()},
		{"unknown rng", func() State { s := richState(); s.RNG[0].Domain = "unknown"; return s }()},
		{"duplicate rng", func() State { s := richState(); s.RNG[1].Domain = RNGDomainAI; return s }()},
		{"incomplete terminal", func() State { s := richState(); s.Terminal.Ended = true; return s }()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateState(tt.state); err == nil {
				t.Fatal("invalid state was accepted")
			}
		})
	}
}

func TestApplyRejectsInvalidCommandsAndSources(t *testing.T) {
	state := fixtureState()
	tests := []struct {
		name    string
		command Command
		rules   RulesPin
		sources Sources
	}{
		{"invalid rules", fixtureCommand(), RulesPin{}, fixtureSources()},
		{"missing command fields", Command{SchemaVersion: CommandSchemaV1}, state.Rules, fixtureSources()},
		{"invalid payload", Command{SchemaVersion: CommandSchemaV1, ID: "command-1", ActorPlayerID: "player-a", Type: CommandNoop, Payload: []byte("{")}, state.Rules, fixtureSources()},
		{"missing sources", fixtureCommand(), state.Rules, Sources{}},
		{"unsupported command", Command{SchemaVersion: CommandSchemaV1, ID: "command-1", ActorPlayerID: "player-a", Type: "future"}, state.Rules, fixtureSources()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if next, events, _, err := Apply(state, tt.command, tt.rules, tt.sources); err == nil || events != nil || next.SchemaVersion != 0 {
				t.Fatal("rejected command produced a transition")
			}
		})
	}
}

func TestEventSerializationValidationAndHashes(t *testing.T) {
	valid := []Event{{SchemaVersion: EventSchemaV1, ID: "event-1", Sequence: 1, Type: "accepted", ActorPlayerID: "player-a", CommandID: "command-1", Payload: []byte(`{"z":1,"a":2}`)}}
	storage, err := CanonicalEventsJSON(valid)
	if err != nil {
		t.Fatal(err)
	}
	if string(storage) == "" || string(storage) == string(valid[0].Payload) {
		t.Fatal("event storage was not serialized")
	}
	if _, err := CanonicalEventsHash(valid); err != nil {
		t.Fatal(err)
	}
	for _, events := range [][]Event{
		{{SchemaVersion: 99, ID: "event-1", Sequence: 1}},
		{{SchemaVersion: EventSchemaV1, Sequence: 1}},
		{{SchemaVersion: EventSchemaV1, ID: "event-1"}},
		{{SchemaVersion: EventSchemaV1, ID: "event-1", Sequence: 1, Ordinal: -1}},
		{{SchemaVersion: EventSchemaV1, ID: "event-2", Sequence: 2}, {SchemaVersion: EventSchemaV1, ID: "event-1", Sequence: 1}},
	} {
		if _, err := CanonicalEventsJSON(events); err == nil {
			t.Fatal("invalid events were accepted")
		}
	}
}
