package evolution

import (
	"runtime"
	"sync"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
)

type Population interface {
	Get(index int) individual.Evolvable
	Count() int
	Update(generation int)
	SetPopulation(Population []individual.Evolvable)
	GetPopulation() []individual.Evolvable
}

// PopulationBuilder creates initial populations
type PopulationBuilder struct{}

// NewPopulationBuilder creates a new population builder
func NewPopulationBuilder() *PopulationBuilder {
	return &PopulationBuilder{}
}

// BuildPopulation creates a population of binary individuals
func (pb *PopulationBuilder) BuildPopulation(size int, genomeType individual.GenomeType, creator func() individual.Evolvable, fitnessCalc fitness.FitnessCalculator) Population {
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()

	chunkSize := (size + numWorkers - 1) / numWorkers
	switch genomeType {
	case individual.ActionTreeGenome:
		return newActionTreeAndWeightsPopulation(size)
	default:
		population := make([]individual.Evolvable, size)
		for i := range numWorkers {
			start := i * chunkSize
			end := start + chunkSize
			end = min(end, size)

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for j := start; j < end; j++ {
					population[j] = creator()
					fitnessCalc.CalculateFitness(&population[j])
				}
			}(start, end) // chunk to use
		}

		wg.Wait()
		newPop := newGenericPopulation(size)
		newPop.SetPopulation(population)
		return newPop
	}
}
