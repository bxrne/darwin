package evolution

import (
	"github.com/bxrne/darwin/internal/individual"
)

// PopulationBuilder creates initial populations
type PopulationBuilder struct{}

// NewPopulationBuilder creates a new population builder
func NewPopulationBuilder() *PopulationBuilder {
	return &PopulationBuilder{}
}

// BuildBinaryPopulation creates a population of binary individuals
func (pb *PopulationBuilder) BuildBinaryPopulation(size, genomeSize int) []individual.Evolvable {
	population := make([]individual.Evolvable, size)
	for i := range population {
		population[i] = individual.NewBinaryIndividual(genomeSize)
	}
	return population
}
