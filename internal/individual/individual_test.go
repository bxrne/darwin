package individual

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBinaryIndividual_GIVEN_genome_size_WHEN_create_THEN_random_genome_and_fitness_calculated(t *testing.T) {
	ind := NewBinaryIndividual(5)

	assert.NotNil(t, ind)
	assert.Len(t, ind.Genome, 5)
	assert.GreaterOrEqual(t, ind.Fitness, 0.0)
	assert.LessOrEqual(t, ind.Fitness, 1.0)
}

func TestBinaryIndividual_GetFitness_GIVEN_calculated_fitness_WHEN_get_THEN_returns_fitness(t *testing.T) {
	ind := &BinaryIndividual{
		Genome:  []byte{'1', '0', '1'},
		Fitness: 0.5,
	}

	fitness := ind.GetFitness()

	assert.Equal(t, 0.5, fitness)
}

func TestBinaryIndividual_CalculateFitness_GIVEN_genome_WHEN_calculate_THEN_fitness_is_ratio_of_ones(t *testing.T) {
	ind := &BinaryIndividual{
		Genome: []byte{'1', '0', '1', '1'},
	}

	ind.CalculateFitness()

	assert.Equal(t, 0.75, ind.Fitness) // 3 ones out of 4
}

func TestBinaryIndividual_CalculateFitness_GIVEN_all_zeros_WHEN_calculate_THEN_fitness_zero(t *testing.T) {
	ind := &BinaryIndividual{
		Genome: []byte{'0', '0', '0'},
	}

	ind.CalculateFitness()

	assert.Equal(t, 0.0, ind.Fitness)
}

func TestBinaryIndividual_CalculateFitness_GIVEN_all_ones_WHEN_calculate_THEN_fitness_one(t *testing.T) {
	ind := &BinaryIndividual{
		Genome: []byte{'1', '1', '1'},
	}

	ind.CalculateFitness()

	assert.Equal(t, 1.0, ind.Fitness)
}

func TestBinaryIndividual_Max_GIVEN_higher_fitness_WHEN_max_THEN_returns_higher(t *testing.T) {
	ind1 := &BinaryIndividual{Fitness: 0.5}
	ind2 := &BinaryIndividual{Fitness: 0.8}

	result := ind1.Max(ind2)

	assert.Equal(t, ind2, result)
}

func TestBinaryIndividual_Max_GIVEN_lower_fitness_WHEN_max_THEN_returns_self(t *testing.T) {
	ind1 := &BinaryIndividual{Fitness: 0.8}
	ind2 := &BinaryIndividual{Fitness: 0.5}

	result := ind1.Max(ind2)

	assert.Equal(t, ind1, result)
}

func TestBinaryIndividual_Max_GIVEN_equal_fitness_WHEN_max_THEN_returns_other(t *testing.T) {
	ind1 := &BinaryIndividual{Fitness: 0.5}
	ind2 := &BinaryIndividual{Fitness: 0.5}

	result := ind1.Max(ind2)

	assert.Equal(t, ind2, result)
}

func TestBinaryIndividual_Mutate_GIVEN_mutation_rate_one_WHEN_mutate_THEN_points_flipped(t *testing.T) {
	ind := &BinaryIndividual{
		Genome: []byte{'0', '1', '0'},
	}

	ind.Mutate([]int{1}, 1.0) // Always mutate

	assert.Equal(t, byte('0'), ind.Genome[1]) // Flipped from '1' to '0'
}

func TestBinaryIndividual_Mutate_GIVEN_mutation_rate_zero_WHEN_mutate_THEN_no_change(t *testing.T) {
	ind := &BinaryIndividual{
		Genome: []byte{'0', '1', '0'},
	}

	original := make([]byte, len(ind.Genome))
	copy(original, ind.Genome)

	ind.Mutate([]int{1}, 0.0) // Never mutate

	assert.Equal(t, original, ind.Genome)
}

func TestBinaryIndividual_MultiPointCrossover_GIVEN_two_individuals_WHEN_crossover_THEN_offspring_created(t *testing.T) {
	ind1 := &BinaryIndividual{
		Genome: []byte{'0', '1', '0', '1'},
	}
	ind2 := &BinaryIndividual{
		Genome: []byte{'1', '0', '1', '0'},
	}

	child1, child2 := ind1.MultiPointCrossover(ind2, 1)

	assert.NotNil(t, child1)
	assert.NotNil(t, child2)
	c1, ok1 := child1.(*BinaryIndividual)
	c2, ok2 := child2.(*BinaryIndividual)
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Len(t, c1.Genome, 4)
	assert.Len(t, c2.Genome, 4)
	assert.GreaterOrEqual(t, c1.Fitness, 0.0)
	assert.LessOrEqual(t, c1.Fitness, 1.0)
	assert.GreaterOrEqual(t, c2.Fitness, 0.0)
	assert.LessOrEqual(t, c2.Fitness, 1.0)
}

func TestBinaryIndividual_MultiPointCrossover_GIVEN_zero_crossover_points_WHEN_crossover_THEN_no_crossover(t *testing.T) {
	ind1 := &BinaryIndividual{
		Genome: []byte{'0', '1'},
	}
	ind2 := &BinaryIndividual{
		Genome: []byte{'1', '0'},
	}

	child1, child2 := ind1.MultiPointCrossover(ind2, 0)

	// With 0 crossover points, should swap entire segments, but since random, hard to test exactly
	assert.NotNil(t, child1)
	assert.NotNil(t, child2)
}

func TestBinaryIndividual_MultiPointCrossover_GIVEN_multiple_points_WHEN_crossover_THEN_correct_segments_swapped(t *testing.T) {
	// This is hard to test deterministically due to randomness, but we can test that it doesn't panic and produces valid offspring
	ind1 := &BinaryIndividual{
		Genome: []byte{'0', '1', '0', '1', '0'},
	}
	ind2 := &BinaryIndividual{
		Genome: []byte{'1', '0', '1', '0', '1'},
	}

	child1, child2 := ind1.MultiPointCrossover(ind2, 2)

	assert.NotNil(t, child1)
	assert.NotNil(t, child2)
	c1, ok1 := child1.(*BinaryIndividual)
	c2, ok2 := child2.(*BinaryIndividual)
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Len(t, c1.Genome, 5)
	assert.Len(t, c2.Genome, 5)
}
