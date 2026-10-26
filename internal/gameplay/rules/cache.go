package rules

import (
	"bytes"
	"reflect"
	"sync"
)

// Rules are immutable game pins, but callers historically own mutable runtime
// slices and canonical bytes. Cache ONLY verified results and return owned copies.
// Key the exact input as well as the claimed pin: a valid hash alone must never
// authorize subsequently corrupted definitions. Whitespace variants may occupy
// separate entries; eviction changes cost, never validation or gameplay results.
const rulesCacheEntries = 16
const rulesCacheInputBytes = 8 << 20

var canonicalRulesCache = newRulesCache[[]byte](rulesCacheEntries, rulesCacheInputBytes, bytes.Clone)
var compiledRulesCache = newRulesCache[RuntimeRules](rulesCacheEntries, rulesCacheInputBytes, cloneRuntime)

type rulesCacheNamespace struct {
	pin      string
	nilInput bool // RawMessage nil encodes null; a non-nil empty message is invalid.
}
type rulesCacheKey struct {
	input     string
	namespace rulesCacheNamespace
}
type rulesCacheEntry[T any] struct {
	value      T
	inputBytes int
}

// rulesCache bounds both entry count and retained input bytes. No invalid
// results or per-game state are retained. Each entry's derived value is
// bounded by its input definitions; oversized definitions still compile normally
// without entering the cache. FIFO avoids a second mutable LRU graph on hot hits.
type rulesCache[T any] struct {
	mu         sync.RWMutex
	misses     sync.Mutex // serialize cold loaders without blocking warm readers
	entries    map[rulesCacheNamespace]map[string]rulesCacheEntry[T]
	order      []rulesCacheKey
	bytes      int
	maxEntries int
	maxBytes   int
	clone      func(T) T
}

func newRulesCache[T any](maxEntries, maxBytes int, clone func(T) T) *rulesCache[T] {
	return &rulesCache[T]{entries: make(map[rulesCacheNamespace]map[string]rulesCacheEntry[T]), maxEntries: maxEntries, maxBytes: maxBytes, clone: clone}
}

func (c *rulesCache[T]) get(input []byte, pin string, load func() (T, error)) (T, error) {
	if len(input) > c.maxBytes || c.maxEntries < 1 {
		return load()
	}
	namespace := rulesCacheNamespace{pin: pin, nilInput: input == nil}
	c.mu.RLock()
	// Go's map lookup compares the complete input bytes, not a claimed digest
	// or a collision-prone abbreviated key. Direct string(input) LOOKUPS do not
	// allocate; copy the immutable key only on insertion. Cryptographic pin
	// verification still occurs in the cold compiler before anything is cached.
	entry, exists := c.entries[namespace][string(input)]
	c.mu.RUnlock()
	if exists {
		return c.clone(entry.value), nil
	}
	c.misses.Lock()
	defer c.misses.Unlock()
	c.mu.RLock()
	entry, exists = c.entries[namespace][string(input)]
	c.mu.RUnlock()
	if !exists {
		// At most one cold loader runs per cache, but warm game requests may
		// continue throughout decoding/compilation. No map lock spans the loader.
		value, err := load()
		if err != nil {
			var zero T
			return zero, err
		}
		c.mu.Lock()
		for len(c.order) >= c.maxEntries || c.bytes+len(input) > c.maxBytes {
			oldest := c.order[0]
			bucket := c.entries[oldest.namespace]
			c.bytes -= bucket[oldest.input].inputBytes
			delete(bucket, oldest.input)
			if len(bucket) == 0 {
				delete(c.entries, oldest.namespace)
			}
			c.order[0] = rulesCacheKey{} // release evicted input even while the backing array survives
			c.order = c.order[1:]
		}
		key := rulesCacheKey{input: string(input), namespace: namespace}
		entry = rulesCacheEntry[T]{value: value, inputBytes: len(input)}
		if c.entries[namespace] == nil {
			c.entries[namespace] = make(map[string]rulesCacheEntry[T])
		}
		c.entries[namespace][key.input] = entry
		c.order = append(c.order, key)
		c.bytes += len(input)
		c.mu.Unlock()
	}
	// Stored graphs never escape. Clone after unlocking: unrelated warm callers
	// need not wait for an owned copy, and evicting an entry cannot invalidate it.
	return c.clone(entry.value), nil
}

// cloneRuntime includes every nested field, including fields added in future.
// A JSON round trip is both substantially slower and incorrect for json:"-"
// fields such as Movement.PolicyConfigured. RuntimeRules is a closed, acyclic,
// exported data model; exhaustive shape/ownership tests guard that contract.
func cloneRuntime(value RuntimeRules) RuntimeRules {
	return cloneRulesValue(reflect.ValueOf(value)).Interface().(RuntimeRules)
}

// Precompute reference-bearing struct fields once. Bulk-copy scalar fields and
// scalar slices: reflecting thousands of individual numeric rule fields would
// otherwise make defensive ownership itself a hot-path cost.
var rulesCloneFields = referenceFields(reflect.TypeOf(RuntimeRules{}))

func referenceFields(root reflect.Type) map[reflect.Type][]int {
	result := make(map[reflect.Type][]int)
	var inspect func(reflect.Type) bool
	inspect = func(typ reflect.Type) bool {
		switch typ.Kind() {
		case reflect.Pointer, reflect.Slice:
			inspect(typ.Elem())
			return true
		case reflect.Struct:
			if fields, ok := result[typ]; ok {
				return len(fields) > 0
			}
			var fields []int
			for i := 0; i < typ.NumField(); i++ {
				if inspect(typ.Field(i).Type) {
					fields = append(fields, i)
				}
			}
			result[typ] = fields
			return len(fields) > 0
		default:
			return false
		}
	}
	inspect(root)
	return result
}

func cloneRulesValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		next := reflect.New(value.Type().Elem())
		next.Elem().Set(cloneRulesValue(value.Elem()))
		return next
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		next := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		if value.Type().Elem().Kind() == reflect.String || len(rulesCloneFields[value.Type().Elem()]) == 0 && value.Type().Elem().Kind() == reflect.Struct {
			reflect.Copy(next, value)
		} else {
			for i := 0; i < value.Len(); i++ {
				next.Index(i).Set(cloneRulesValue(value.Index(i)))
			}
		}
		return next
	case reflect.Struct:
		fields := rulesCloneFields[value.Type()]
		if len(fields) == 0 {
			return value
		}
		next := reflect.New(value.Type()).Elem()
		next.Set(value)
		for _, index := range fields {
			next.Field(index).Set(cloneRulesValue(value.Field(index)))
		}
		return next
	default:
		return value // Current leaves are immutable scalar values, not references.
	}
}
