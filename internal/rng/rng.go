package rng

import (
	"math/rand"
	"sync"
)

var (
	mu  sync.Mutex
	rng *rand.Rand
)

func init() {
	// Default seed for reproducibility
	Seed(42)
}

// Seed sets the random seed for reproducible results
func Seed(seed int64) {
	mu.Lock()
	defer mu.Unlock()
	rng = rand.New(rand.NewSource(seed))
}

// Intn returns a random int in [0,n)
func Intn(n int) int {
	mu.Lock()
	defer mu.Unlock()
	return rng.Intn(n)
}

// Float64 returns a random float64 in [0.0,1.0)
func Float64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return rng.Float64()
}
