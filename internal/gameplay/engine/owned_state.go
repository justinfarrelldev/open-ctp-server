package engine

import (
	"encoding"
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// cloneOwnedState is the execution clone, equivalent to the former canonical
// JSON encode/decode round trip without serializing and reparsing the entire
// world. JSON omission, raw-message escaping, timestamps and custom codecs must
// remain identical; a plain shallow copy or JSON-agnostic deep clone is unsafe.
// Durable serialization still uses CanonicalStateJSON unchanged.
func cloneOwnedState(state State) (State, error) {
	if err := ValidateState(state); err != nil {
		return State{}, err
	}
	value, err := copyJSONValue(reflect.ValueOf(state))
	if err != nil {
		return State{}, err
	}
	next := value.Interface().(State)
	normalizeState(&next)
	return next, nil
}

type jsonCopyField struct {
	index   int
	omit    bool
	ignored bool
}

var jsonCopyPlans sync.Map // reflect.Type -> []jsonCopyField; fixed data-model types only
var jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
var jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
var textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func customJSONCodec(typ reflect.Type) bool {
	if typ.NumMethod() == 0 && reflect.PointerTo(typ).NumMethod() == 0 {
		return false
	}
	return typ.Implements(jsonMarshalerType) || reflect.PointerTo(typ).Implements(jsonUnmarshalerType) || typ.Implements(textMarshalerType) || reflect.PointerTo(typ).Implements(textUnmarshalerType)
}

// Bulk-copy scalar struct fields; inspect reference/string/float fields only.
// Plans contain types/indices, never game data, and are safe across goroutines.
func jsonCopyFields(typ reflect.Type) []jsonCopyField {
	if found, ok := jsonCopyPlans.Load(typ); ok {
		return found.([]jsonCopyField)
	}
	var fields []jsonCopyField
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		tag := field.Tag.Get("json")
		ignored := tag == "-"
		switch field.Type.Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if !ignored && !customJSONCodec(field.Type) {
				continue
			}
		}
		fields = append(fields, jsonCopyField{index: index, ignored: ignored, omit: strings.Contains(","+tag+",", ",omitempty,")})
	}
	actual, _ := jsonCopyPlans.LoadOrStore(typ, fields)
	return actual.([]jsonCopyField)
}

// Maps and custom JSON types (notably time.Time and json.RawMessage) retain their
// actual codec semantics. World arrays and ordinary structs use an owned copy.
// In particular, RawMessage nil becomes literal null unless omitted by its
// enclosing field, and time.Time loses monotonic/named-zone audit-only internals
// exactly as it did in the old round trip.
func jsonCodecCopy(value reflect.Value) (reflect.Value, error) {
	raw, err := json.Marshal(value.Interface())
	if err != nil {
		return reflect.Value{}, err
	}
	next := reflect.New(value.Type())
	if err := json.Unmarshal(raw, next.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return next.Elem(), nil
}

func copyJSONValue(value reflect.Value) (reflect.Value, error) {
	typ := value.Type()
	if typ == reflect.TypeOf(RulesPin{}) {
		pin := value.Interface().(RulesPin)
		definitions, err := rules.CompactDefinitions(pin.Definitions)
		if err != nil {
			return reflect.Value{}, err
		}
		pin.Definitions = definitions
		if !utf8.ValidString(pin.Version) {
			pin.Version = string([]rune(pin.Version))
		}
		if !utf8.ValidString(pin.ContentHash) {
			pin.ContentHash = string([]rune(pin.ContentHash))
		}
		return reflect.ValueOf(pin), nil
	}
	if customJSONCodec(typ) {
		return jsonCodecCopy(value)
	}
	switch value.Kind() {
	case reflect.Struct:
		next := reflect.New(typ).Elem()
		next.Set(value)
		for _, field := range jsonCopyFields(typ) {
			original := value.Field(field.index)
			if field.ignored || field.omit && jsonEmpty(original) {
				next.Field(field.index).Set(reflect.Zero(original.Type()))
				continue
			}
			copied, err := copyJSONValue(original)
			if err != nil {
				return reflect.Value{}, err
			}
			next.Field(field.index).Set(copied)
		}
		return next, nil
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(typ), nil
		}
		copied, err := copyJSONValue(value.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		next := reflect.New(typ.Elem())
		next.Elem().Set(copied)
		return next, nil
	case reflect.Slice:
		if customJSONCodec(typ.Elem()) {
			return jsonCodecCopy(value)
		}
		if value.IsNil() {
			return reflect.Zero(typ), nil
		}
		next := reflect.MakeSlice(typ, value.Len(), value.Len())
		switch typ.Elem().Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			reflect.Copy(next, value)
			return next, nil
		}
		for index := 0; index < value.Len(); index++ {
			copied, err := copyJSONValue(value.Index(index))
			if err != nil {
				return reflect.Value{}, err
			}
			next.Index(index).Set(copied)
		}
		return next, nil
	case reflect.Array:
		next := reflect.New(typ).Elem()
		for index := 0; index < value.Len(); index++ {
			copied, err := copyJSONValue(value.Index(index))
			if err != nil {
				return reflect.Value{}, err
			}
			next.Index(index).Set(copied)
		}
		return next, nil
	case reflect.Map, reflect.Interface:
		// The engine's maps are small specialist maps. JSON's key sorting/collision
		// and interface-number behavior are preserved rather than reimplemented.
		return jsonCodecCopy(value)
	case reflect.String:
		if utf8.ValidString(value.String()) {
			return value, nil
		}
		next := reflect.New(typ).Elem()
		next.SetString(string([]rune(value.String())))
		return next, nil
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(value.Float()) || math.IsInf(value.Float(), 0) {
			return jsonCodecCopy(value)
		}
		return value, nil
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value, nil
	default:
		return jsonCodecCopy(value) // Preserve fail-closed codec behavior for unsupported future fields.
	}
}

// Match encoding/json's omitempty test, not reflect.IsZero for structs/arrays.
func jsonEmpty(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Interface, reflect.Pointer:
		return value.IsZero()
	default:
		return false
	}
}
