package neat

import "math/rand"

// RNG abstracts randomness for deterministic runs.
type RNG interface {
	Float64() float64
	Intn(n int) int
}

// NewRand returns a standard math/rand RNG with the provided seed.
func NewRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func randBool(rng RNG, p float64) bool {
	if p <= 0 {
		return false
	}
	if p >= 1 {
		return true
	}
	return rng.Float64() < p
}

func randRange(rng RNG, min, max float64) float64 {
	if max <= min {
		return min
	}
	return min + (max-min)*rng.Float64()
}
