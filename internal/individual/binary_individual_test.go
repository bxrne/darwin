package individual_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestNewBinaryIndividual_GIVEN_genome_size_WHEN_create_THEN_random_genome_and_fitness_calculated(t *testing.T) {
	ind := individual.NewBinaryIndividual(5)

	assert.NotNil(t, ind)
	assert.Len(t, ind.Genome, 5)
	assert.GreaterOrEqual(t, ind.Fitness, 0.0)
	assert.LessOrEqual(t, ind.Fitness, 1.0)
}

func TestBinaryIndividual_Max_GIVEN_higher_fitness_WHEN_max_THEN_returns_higher(t *testing.T) {
	ind1 := &individual.BinaryIndividual{Fitness: 0.5}
	ind2 := &individual.BinaryIndividual{Fitness: 0.8}

	result := ind1.Max(ind2)

	assert.Equal(t, ind2, result)
}

func TestBinaryIndividual_Max_GIVEN_lower_fitness_WHEN_max_THEN_returns_self(t *testing.T) {
	ind1 := &individual.BinaryIndividual{Fitness: 0.8}
	ind2 := &individual.BinaryIndividual{Fitness: 0.5}

	result := ind1.Max(ind2)

	assert.Equal(t, ind1, result)
}

func TestBinaryIndividual_Max_GIVEN_equal_fitness_WHEN_max_THEN_returns_other(t *testing.T) {
	ind1 := &individual.BinaryIndividual{Fitness: 0.5}
	ind2 := &individual.BinaryIndividual{Fitness: 0.5}

	result := ind1.Max(ind2)

	assert.Equal(t, ind2, result)
}
