package world

import "math"

const splitMix64GoldenRatio = 0x9e3779b97f4a7c15

// DeterministicRNG is the pure RNG abstraction for future generation work.
type DeterministicRNG interface {
	Seed() int64
	Next() (DeterministicRNG, uint64)
	Intn(n int) (DeterministicRNG, int)
	Float64() (DeterministicRNG, float64)
}

// SplitMix64RNG is a tiny immutable deterministic RNG state value.
type SplitMix64RNG struct {
	seed  uint64
	state uint64
}

// NewDeterministicRNG returns a deterministic SplitMix64 state value.
func NewDeterministicRNG(seed int64) SplitMix64RNG {
	state := uint64(seed)
	return SplitMix64RNG{seed: state, state: state}
}

// Seed returns the original seed used to initialize the RNG state.
func (r SplitMix64RNG) Seed() int64 {
	return int64(r.seed)
}

// Next advances the RNG and returns the next state plus a random value.
func (r SplitMix64RNG) Next() (DeterministicRNG, uint64) {
	r.state += splitMix64GoldenRatio
	return r, mix64(r.state)
}

// Intn returns a deterministic integer in [0, n).
func (r SplitMix64RNG) Intn(n int) (DeterministicRNG, int) {
	if n <= 0 {
		panic("Intn requires n > 0")
	}
	next, value := r.Next()
	return next, int(value % uint64(n))
}

// Float64 returns a deterministic floating-point value in [0.0, 1.0).
func (r SplitMix64RNG) Float64() (DeterministicRNG, float64) {
	next, value := r.Next()
	return next, float64(value>>11) * (1.0 / (1 << 53))
}

func mix64(value uint64) uint64 {
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

// Split returns a child RNG seeded from the current state and a label.
func (r SplitMix64RNG) Split(label string) SplitMix64RNG {
	seed := r.seed
	for i := 0; i < len(label); i++ {
		seed ^= uint64(label[i]) + splitMix64GoldenRatio + (seed << 6) + (seed >> 2)
	}
	return NewDeterministicRNG(int64(mix64(seed)))
}

// ClampFloat64 bounds a floating-point value to [0, 1) without adding randomness.
func ClampFloat64(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value >= 1 {
		return math.Nextafter(1, 0)
	}
	return value
}
