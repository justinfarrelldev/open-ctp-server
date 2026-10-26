package engine

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func legacyStateClone(state State) (State, error) {
	raw, err := CanonicalStateJSON(state)
	if err != nil {
		return State{}, err
	}
	var result State
	err = json.Unmarshal(raw, &result)
	return result, err
}

// State deliberately uses ordinary exported JSON fields. If future schema work
// introduces embedding/private fields, audit codec promotion/omission before the
// execution copy is allowed to silently diverge from encoding/json.
func TestExecutionCopyCoversTheCompleteStateSchema(t *testing.T) {
	seen := make(map[reflect.Type]bool)
	var inspect func(reflect.Type)
	inspect = func(typ reflect.Type) {
		if seen[typ] {
			return
		}
		seen[typ] = true
		if typ.Implements(jsonMarshalerType) || reflect.PointerTo(typ).Implements(jsonUnmarshalerType) {
			return
		}
		switch typ.Kind() {
		case reflect.Struct:
			for index := 0; index < typ.NumField(); index++ {
				field := typ.Field(index)
				if field.Anonymous || field.PkgPath != "" {
					t.Fatalf("audit execution JSON-copy semantics for %s.%s", typ, field.Name)
				}
				inspect(field.Type)
			}
		case reflect.Pointer, reflect.Slice, reflect.Array:
			inspect(typ.Elem())
		case reflect.Map:
			inspect(typ.Key())
			inspect(typ.Elem())
		case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		default:
			t.Fatalf("audit new execution data shape %s", typ)
		}
	}
	inspect(reflect.TypeOf(State{}))
}

func TestOwnedStateCloneIsExactlyTheCanonicalJSONRoundTrip(t *testing.T) {
	for _, fixture := range []State{fixtureState(), richState()} {
		// Include JSON omission/null, raw JSON compaction, invalid UTF-8 and audit
		// normalization: these are easy to change accidentally in a fast deep copy.
		fixture.Notifications = []Notification{}
		fixture.Policy.Queue = []TurnPolicyCommand{}
		fixture.Rules.Definitions = json.RawMessage("null")
		fixture.Audit.UpdatedAt = time.Now()
		fixture.Cities = append(fixture.Cities, City{ID: "clone-city", OwnerID: "player-a", TileID: "tile-a", Name: "invalid\xff\xff<name>", Specialists: map[string]int{}})
		want, err := legacyStateClone(fixture)
		if err != nil {
			t.Fatal(err)
		}
		got, err := cloneOwnedState(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("fast clone changed canonical shape\ngot=%+v\nwant=%+v", got, want)
		}
		rawWant, err := CanonicalStateJSON(want)
		if err != nil {
			t.Fatal(err)
		}
		rawGot, err := CanonicalStateJSON(got)
		if err != nil {
			t.Fatal(err)
		}
		if string(rawGot) != string(rawWant) {
			t.Fatal("fast clone changed persisted bytes")
		}
	}
}

func TestOwnedStateCloneOwnsNestedFogAndRuleBytes(t *testing.T) {
	state := richState()
	state.Fog[0].CityEffects = []Effect{{ID: "b", Kind: "vision"}, {ID: "a", Kind: "bonus"}}
	population := 3
	state.Fog[0].CityPopulation = &population
	state.Fog[0].Inspection = &TileInspection{MovementCost: &population, Yields: &TileYield{Food: 2}}
	state.Rules.Definitions = json.RawMessage("null")
	before, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	next, err := cloneOwnedState(state)
	if err != nil {
		t.Fatal(err)
	}
	// Find the owned fog row after deterministic sorting.
	for index := range next.Fog {
		if next.Fog[index].Inspection != nil {
			*next.Fog[index].CityPopulation = 99
			*next.Fog[index].Inspection.MovementCost = 99
			next.Fog[index].Inspection.Yields.Food = 99
			next.Fog[index].CityEffects[0].ID = "changed"
		}
	}
	next.Rules.Definitions[0] = 'x'
	after, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("cloning or mutating the result altered caller-owned fog/rules")
	}
}

func TestCanonicalEncodingDoesNotSortCallerFogEffects(t *testing.T) {
	state := richState()
	state.Fog[0].CityEffects = []Effect{{ID: "b", Kind: "vision"}, {ID: "a", Kind: "bonus"}}
	before, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CanonicalStateJSON(state); err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("canonical serialization reordered caller-owned fog memory")
	}
}

type ownedCustomNumber int

func (ownedCustomNumber) MarshalJSON() ([]byte, error)       { return []byte(`7`), nil }
func (number *ownedCustomNumber) UnmarshalJSON([]byte) error { *number = 21; return nil }

type ownedCustomText string

func (text ownedCustomText) MarshalText() ([]byte, error) { return []byte("wire-" + string(text)), nil }
func (text *ownedCustomText) UnmarshalText(raw []byte) error {
	*text = ownedCustomText("decoded-" + string(raw))
	return nil
}

func TestOwnedCopyPreservesCustomScalarCodecs(t *testing.T) {
	type data struct {
		Number  ownedCustomNumber
		Text    ownedCustomText
		Numbers []ownedCustomNumber
	}
	input := data{Number: 3, Text: "value", Numbers: []ownedCustomNumber{3}}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var want data
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	got, err := copyJSONValue(reflect.ValueOf(input))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Interface(), want) {
		t.Fatalf("custom codec differs: %+v != %+v", got.Interface(), want)
	}
}

func TestOwnedJSONCopyPreservesRawMessageAndRejectsInvalidValues(t *testing.T) {
	type example struct {
		Raw     json.RawMessage `json:"raw"`
		Omitted []int           `json:"omitted,omitempty"`
		Empty   []int           `json:"empty"`
		Ignored []int           `json:"-"`
	}
	input := example{Raw: json.RawMessage(` { "text" : "<x>", "z":1 } `), Omitted: []int{}, Empty: []int{}, Ignored: []int{3}}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var want example
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	got, err := copyJSONValue(reflect.ValueOf(input))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Interface(), want) {
		t.Fatalf("JSON copy differs: %+v != %+v", got.Interface(), want)
	}
	if _, err := copyJSONValue(reflect.ValueOf(example{Raw: json.RawMessage(`bad-json`)})); err == nil {
		t.Fatal("invalid raw JSON accepted")
	}
}
