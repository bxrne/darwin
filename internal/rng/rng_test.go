package rng_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/rng"
	"github.com/stretchr/testify/assert"
)

func TestSeed_GIVEN_seed_WHEN_seed_THEN_rng_initialized(t *testing.T) {
	rng.Seed(123)
	// Should not panic
}

func TestIntn_GIVEN_n_WHEN_intn_THEN_returns_value_in_range(t *testing.T) {
	rng.Seed(42)
	val := rng.Intn(10)
	assert.GreaterOrEqual(t, val, 0)
	assert.Less(t, val, 10)
}

func TestFloat64_GIVEN_no_args_WHEN_float64_THEN_returns_value_in_range(t *testing.T) {
	rng.Seed(42)
	val := rng.Float64()
	assert.GreaterOrEqual(t, val, 0.0)
	assert.Less(t, val, 1.0)
}

func TestReproducibility_GIVEN_same_seed_WHEN_generate_values_THEN_same_sequence(t *testing.T) {
	rng.Seed(99)
	val1 := rng.Intn(100)
	float1 := rng.Float64()

	rng.Seed(99)
	val2 := rng.Intn(100)
	float2 := rng.Float64()

	assert.Equal(t, val1, val2)
	assert.Equal(t, float1, float2)
}
