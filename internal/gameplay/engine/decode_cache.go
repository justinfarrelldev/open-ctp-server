package engine

import "sync"

// AI projection and command admission often read the same committed row. Cache
// only exact, fully validated input bytes, not a game/revision guess. A same-
// revision edit or corrupt pin must never borrow a previously decoded authority.
// Capacity bounds retained raw inputs AND entry count; each decoded graph is
// bounded by its input state. Cache contents never cross a viewer/API boundary.
var decodedStates = newDecodedStateCache(8, 8<<20)

type decodedStateCache struct {
	mu         sync.RWMutex
	misses     sync.Mutex // a cold world must not stall cached worlds
	entries    map[string]State
	order      []string
	bytes      int
	maxEntries int
	maxBytes   int
}

func newDecodedStateCache(maxEntries, maxBytes int) *decodedStateCache {
	return &decodedStateCache{entries: make(map[string]State), maxEntries: maxEntries, maxBytes: maxBytes}
}
func (c *decodedStateCache) decode(raw []byte) (State, error) {
	if len(raw) > c.maxBytes || c.maxEntries < 1 {
		return decodeCanonicalStateJSON(raw)
	}
	c.mu.RLock()
	state, ok := c.entries[string(raw)] // allocation-free full-byte map lookup
	c.mu.RUnlock()
	if ok {
		return cloneOwnedState(state)
	}
	c.misses.Lock()
	defer c.misses.Unlock()
	c.mu.RLock()
	state, ok = c.entries[string(raw)]
	c.mu.RUnlock()
	if !ok {
		var err error
		state, err = decodeCanonicalStateJSON(raw)
		if err != nil {
			return State{}, err
		}
		c.mu.Lock()
		for len(c.entries) >= c.maxEntries || c.bytes+len(raw) > c.maxBytes {
			oldest := c.order[0]
			delete(c.entries, oldest)
			c.order[0] = "" // do not retain evicted private inputs in the FIFO backing array
			c.order = c.order[1:]
			c.bytes -= len(oldest)
		}
		key := string(raw) // immutable owned input, retained only on a validated miss
		c.entries[key] = state
		c.order = append(c.order, key)
		c.bytes += len(raw)
		c.mu.Unlock()
	}
	// The cached graph never escapes, even on its first decode. This also retains
	// all invariant checks for warm callers; no mutable state is shared with cache.
	return cloneOwnedState(state)
}
