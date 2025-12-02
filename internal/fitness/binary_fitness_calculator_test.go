package fitness_test

import (
	"testing"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestCalculateFitness_GIVEN_valid_binary_individual_WHEN_calculate_fitness_THEN_sets_correct_fitness(t *testing.T) {
	tests := []struct {
		name            string
		genome          []byte
		expectedFitness float64
	}{
		{name: "simple mixed bits", genome: []byte("10101"), expectedFitness: 0.6},
		{name: "all ones", genome: []byte("1111"), expectedFitness: 1.0},
		{name: "all zeros", genome: []byte("0000"), expectedFitness: 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := &fitness.BinaryFitnessCalculator{}
			ind := &individual.BinaryIndividual{Genome: tt.genome}

			calculator.CalculateFitness(ind)

			assert.Equal(t, tt.expectedFitness, ind.Fitness)
		})
	}
}
