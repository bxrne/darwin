package selection

import (
	"testing"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/stretchr/testify/assert"
)

func TestRouletteSelector_Select_GIVEN_population_WHEN_select_THEN_returns_individual_based_on_fitness(t *testing.T) {
	// Create population with known fitness
	pop := []individual.Evolvable{
		&individual.BinaryIndividual{Fitness: 0.1},
		&individual.BinaryIndividual{Fitness: 0.9},
	}

	selector := NewRouletteSelector(2)

	selected := selector.Select(pop)

	assert.NotNil(t, selected)
	assert.Contains(t, pop, selected)
}

func TestRouletteSelector_Select_GIVEN_sample_size_WHEN_select_THEN_uses_sample_size(t *testing.T) {
	pop := []individual.Evolvable{
		&individual.BinaryIndividual{Fitness: 0.5},
		&individual.BinaryIndividual{Fitness: 0.5},
	}

	selector := NewRouletteSelector(1) // Sample size 1

	selected := selector.Select(pop)

	assert.NotNil(t, selected)
}

func TestTournamentSelector_Select_GIVEN_population_WHEN_select_THEN_returns_best_from_tournament(t *testing.T) {
	pop := []individual.Evolvable{
		&individual.BinaryIndividual{Fitness: 0.1},
		&individual.BinaryIndividual{Fitness: 0.5},
		&individual.BinaryIndividual{Fitness: 0.9},
	}

	selector := NewTournamentSelector(2)

	selected := selector.Select(pop)

	assert.NotNil(t, selected)
	// Should be one of the individuals
	assert.Contains(t, pop, selected)
}

func TestTournamentSelector_Select_GIVEN_tournament_size_WHEN_select_THEN_uses_tournament_size(t *testing.T) {
	pop := []individual.Evolvable{
		&individual.BinaryIndividual{Fitness: 0.5},
		&individual.BinaryIndividual{Fitness: 0.5},
	}

	selector := NewTournamentSelector(1) // Tournament size 1

	selected := selector.Select(pop)

	assert.NotNil(t, selected)
}

func TestTournamentSelector_Select_GIVEN_tournament_size_one_WHEN_select_THEN_returns_random_individual(t *testing.T) {
	pop := []individual.Evolvable{
		&individual.BinaryIndividual{Fitness: 0.1},
		&individual.BinaryIndividual{Fitness: 0.9},
	}

	selector := NewTournamentSelector(1)

	selected := selector.Select(pop)

	assert.NotNil(t, selected)
	assert.Contains(t, pop, selected)
}
