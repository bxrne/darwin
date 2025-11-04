package rng

import (
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

var (
	mu   sync.Mutex
	seed int64 = 42 // Default seed for reproducibility
	pool       = sync.Pool{
		New: func() any {
			return rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
		},
	}
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
	// Reset pool to get new RNGs with the new seed
	pool = sync.Pool{
		New: func() any {
			return rand.New(rand.NewSource(uint64(seed)))
		},
	}
}

// Intn returns a random int in [0,n)
func Intn(n int) int {
	rng := pool.Get().(*rand.Rand)
	defer pool.Put(rng)
	return rng.Intn(n)
}

// Float64 returns a random float64 in [0.0,1.0)
func Float64() float64 {
	rng := pool.Get().(*rand.Rand)
	defer pool.Put(rng)
	return rng.Float64()
}
