package engine

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
)

func TestDecodedStateCacheMatchesCodecAndDoesNotShareAuthority(t *testing.T) {
	cache := newDecodedStateCache(2, 1<<20)
	state := richState()
	population := 3
	state.Fog[0].CityPopulation = &population
	state.Fog[0].CityEffects = []Effect{{ID: "effect", Kind: "bonus"}}
	raw, err := CanonicalStateJSON(state)
	if err != nil {
		t.Fatal(err)
	}
	want, err := decodeCanonicalStateJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		got, err := cache.decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("cached decoding differs from codec")
		}
		for index := range got.Fog {
			if got.Fog[index].CityPopulation != nil {
				*got.Fog[index].CityPopulation = 99
				got.Fog[index].CityEffects[0].ID = "mutated"
			}
		}
	}
	changed := append([]byte(nil), raw...)
	changed[0] = 'x'
	if _, err := cache.decode(changed); err == nil {
		t.Fatal("corrupted bytes borrowed cached revision authority")
	}
	if len(cache.entries) != 1 {
		t.Fatal("invalid state was cached")
	}
}

func TestDecodedStateCacheBoundsAndConcurrentIsolation(t *testing.T) {
	cache := newDecodedStateCache(2, 1<<20)
	var raws [][]byte
	for revision := int64(0); revision < 3; revision++ {
		state := fixtureState()
		state.Revision = revision
		raw, err := json.Marshal(state)
		if err != nil {
			t.Fatal(err)
		}
		raws = append(raws, raw)
		if _, err = cache.decode(raw); err != nil {
			t.Fatal(err)
		}
	}
	if len(cache.entries) != 2 || cache.bytes > 1<<20 {
		t.Fatal("decoded states exceed cache bounds")
	}
	small := newDecodedStateCache(2, 1)
	if _, err := small.decode(raws[0]); err != nil {
		t.Fatal(err)
	}
	if len(small.entries) != 0 {
		t.Fatal("oversized state entered cache")
	}
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 8 {
				got, err := cache.decode(raws[2])
				if err != nil {
					t.Error(err)
					return
				}
				got.Players[0].ID = "mutated"
			}
		}()
	}
	workers.Wait()
	got, err := cache.decode(raws[2])
	if err != nil {
		t.Fatal(err)
	}
	if got.Players[0].ID == "mutated" {
		t.Fatal("concurrent caller poisoned state authority")
	}
}
