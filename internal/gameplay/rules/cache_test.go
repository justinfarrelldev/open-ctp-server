package rules

import (
	"bytes"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestRulesCacheReusesVerifiedResultsWithOwnedReturns(t *testing.T) {
	cache := newRulesCache[[]byte](2, 32, bytes.Clone)
	calls := 0
	load := func() ([]byte, error) { calls++; return []byte("verified"), nil }
	first, err := cache.get([]byte("a"), "pin", load)
	if err != nil {
		t.Fatal(err)
	}
	first[0] = 'X'
	second, err := cache.get([]byte("a"), "pin", load)
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != "verified" || calls != 1 {
		t.Fatalf("cache ownership/reuse: %q calls=%d", second, calls)
	}
	for _, key := range []string{"b", "c", "a"} {
		if _, err := cache.get([]byte(key), "pin", load); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 4 || len(cache.order) != 2 || cache.bytes > 32 {
		t.Fatalf("FIFO bounds: calls=%d entries=%d bytes=%d", calls, len(cache.order), cache.bytes)
	}
}

func TestRulesCacheSeparatesRawContentAndPinAndNeverStoresFailuresOrOversize(t *testing.T) {
	cache := newRulesCache[[]byte](8, 3, bytes.Clone)
	calls := 0
	load := func() ([]byte, error) { calls++; return []byte("ok"), nil }
	for _, request := range [][2]string{{"a", "one"}, {"a", "two"}, {"b", "one"}, {"oversize", "one"}, {"oversize", "one"}} {
		if _, err := cache.get([]byte(request[0]), request[1], load); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 5 || cache.bytes > 3 {
		t.Fatalf("key/size bounds: calls=%d bytes=%d", calls, cache.bytes)
	}
	fail := func() ([]byte, error) { calls++; return nil, errors.New("invalid") }
	for range 2 {
		if _, err := cache.get([]byte("x"), "one", fail); err == nil {
			t.Fatal("verification failure accepted")
		}
	}
	if calls != 7 {
		t.Fatal("failure was cached")
	}
}

func TestCompiledRulesCacheMatchesUncachedAndRejectsPoisoning(t *testing.T) {
	pin, err := ServerOwnedSyntheticRuleset("score_2300")
	if err != nil {
		t.Fatal(err)
	}
	want, err := compileRuntime(pin.CanonicalJSON, pin.ContentHash)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		got, err := CompileRuntime(pin.CanonicalJSON, pin.ContentHash)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("cache changed compiled runtime")
		}
		got.Starter.Effects = append(got.Starter.Effects, RuntimeStarterEffect{ID: "private"})
		if len(got.Units) > 0 {
			got.Units[0].ID = "mutated"
		}
	}
	tampered := bytes.Replace(pin.CanonicalJSON, []byte("score_2300"), []byte("score_9999"), 1)
	if _, err := CompileRuntime(tampered, pin.ContentHash); err == nil {
		t.Fatal("same claimed hash accepted tampered bytes")
	}
	if _, err := CompileRuntime(pin.CanonicalJSON, "wrong"); err == nil {
		t.Fatal("same raw bytes accepted wrong hash")
	}
	canonical, err := CanonicalDefinitions(pin.CanonicalJSON)
	if err != nil {
		t.Fatal(err)
	}
	canonical[0] = 'x'
	again, err := CanonicalDefinitions(pin.CanonicalJSON)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, pin.CanonicalJSON) {
		t.Fatal("caller poisoned canonical cache")
	}
}

// Populate every reference-bearing field, including future nested fields, then
// mutate a clone. This catches accidental sharing beyond the current fixtures.
func filledRulesValue(typ reflect.Type, marker string) reflect.Value {
	result := reflect.New(typ).Elem()
	switch typ.Kind() {
	case reflect.Struct:
		for i := 0; i < result.NumField(); i++ {
			result.Field(i).Set(filledRulesValue(typ.Field(i).Type, marker))
		}
	case reflect.Pointer:
		result.Set(reflect.New(typ.Elem()))
		result.Elem().Set(filledRulesValue(typ.Elem(), marker))
	case reflect.Slice:
		result.Set(reflect.MakeSlice(typ, 1, 1))
		result.Index(0).Set(filledRulesValue(typ.Elem(), marker))
	case reflect.Map:
		result.Set(reflect.MakeMap(typ))
		result.SetMapIndex(filledRulesValue(typ.Key(), marker), filledRulesValue(typ.Elem(), marker))
	case reflect.String:
		result.SetString(marker)
	case reflect.Bool:
		result.SetBool(marker == "original")
	case reflect.Int, reflect.Int64:
		if marker == "original" {
			result.SetInt(1)
		} else {
			result.SetInt(2)
		}
	case reflect.Float64:
		if marker == "original" {
			result.SetFloat(1)
		} else {
			result.SetFloat(2)
		}
	default:
		panic("extend exhaustive rules ownership test for " + typ.String())
	}
	return result
}

func TestCloneRuntimeOwnsEveryNestedFieldAndPreservesNil(t *testing.T) {
	original := filledRulesValue(reflect.TypeOf(RuntimeRules{}), "original").Interface().(RuntimeRules)
	copied := cloneRuntime(original)
	if !reflect.DeepEqual(original, copied) {
		t.Fatal("clone changed values")
	}
	// Recursively mutate scalar leaves in the cloned graph, not replace pointers.
	mutateRulesLeaves(reflect.ValueOf(&copied).Elem())
	expected := filledRulesValue(reflect.TypeOf(RuntimeRules{}), "original").Interface().(RuntimeRules)
	if !reflect.DeepEqual(original, expected) {
		t.Fatal("nested runtime storage was shared")
	}
	if !reflect.DeepEqual(cloneRuntime(RuntimeRules{}), RuntimeRules{}) {
		t.Fatal("nil slices/pointers changed")
	}
}
func mutateRulesLeaves(value reflect.Value) {
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			mutateRulesLeaves(value.Field(i))
		}
	case reflect.Pointer:
		mutateRulesLeaves(value.Elem())
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			mutateRulesLeaves(value.Index(i))
		}
	case reflect.String:
		value.SetString("changed")
	case reflect.Bool:
		value.SetBool(!value.Bool())
	case reflect.Int, reflect.Int64:
		value.SetInt(9)
	case reflect.Float64:
		value.SetFloat(9)
	default:
		panic("extend exhaustive rules mutation test")
	}
}

func TestWarmRulesCacheHitDoesNotWaitForUnrelatedColdCompilation(t *testing.T) {
	cache := newRulesCache[[]byte](2, 1024, bytes.Clone)
	if _, err := cache.get([]byte("warm"), "pin", func() ([]byte, error) { return []byte("verified"), nil }); err != nil {
		t.Fatal(err)
	}
	entered, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(finished)
		_, _ = cache.get([]byte("cold"), "pin", func() ([]byte, error) { close(entered); <-release; return []byte("cold"), nil })
	}()
	<-entered
	warm := make(chan []byte, 1)
	go func() {
		value, _ := cache.get([]byte("warm"), "pin", func() ([]byte, error) { return []byte("wrong"), nil })
		warm <- value
	}()
	immediate := false
	select {
	case value := <-warm:
		immediate = string(value) == "verified"
	case <-time.After(time.Second):
	}
	close(release)
	<-finished
	if !immediate {
		t.Fatal("warm gameplay rules waited behind another game's cold compiler")
	}
}

func TestRulesCacheConcurrentCallersAreIsolated(t *testing.T) {
	cache := newRulesCache[RuntimeRules](4, 1024, cloneRuntime)
	original := filledRulesValue(reflect.TypeOf(RuntimeRules{}), "original").Interface().(RuntimeRules)
	var workers sync.WaitGroup
	for range 16 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 10 {
				result, err := cache.get([]byte("rules"), "pin", func() (RuntimeRules, error) { return original, nil })
				if err != nil {
					t.Error(err)
					return
				}
				if !reflect.DeepEqual(result, original) {
					t.Error("concurrent caller poisoned rules")
					return
				}
				mutateRulesLeaves(reflect.ValueOf(&result).Elem())
			}
		}()
	}
	workers.Wait()
}
