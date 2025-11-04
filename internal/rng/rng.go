package rng

import (
	"math/rand/v2"
	"sync"
)

var (
	mu   sync.Mutex
	rng  *rand.Rand
	seed int64 = 42 // Default seed for reproducibility
)

func init() {
	// Default seed for reproducibility
	Seed(42)
}

// Seed sets the random seed for reproducible results
func Seed(s int64) {
	mu.Lock()
	defer mu.Unlock()
	seed = s
	rng = rand.New(rand.NewPCG(uint64(seed), uint64(seed)))
}

// Intn returns a random int in [0,n)
func Intn(n int) int {
	mu.Lock()
	defer mu.Unlock()
	return rng.IntN(n)
}

// Float64 returns a random float64 in [0.0,1.0)
func Float64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return rng.Float64()
}
