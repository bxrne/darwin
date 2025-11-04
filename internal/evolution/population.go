package evolution

import (
	"github.com/bxrne/darwin/internal/individual"
	"runtime"
	"sync"
)

// PopulationBuilder creates initial populations
type PopulationBuilder struct{}

// NewPopulationBuilder creates a new population builder
func NewPopulationBuilder() *PopulationBuilder {
	return &PopulationBuilder{}
}

// BuildPopulation creates a population of binary individuals
// func (pb *PopulationBuilder) BuildPopulation(size, genomeSize int) []individual.Evolvable {
func (pb *PopulationBuilder) BuildPopulation(size int, creator func() individual.Evolvable) []individual.Evolvable {
	population := make([]individual.Evolvable, size)
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()

	chunkSize := (size + numWorkers - 1) / numWorkers
	for i := range numWorkers {
		start := i * chunkSize
		end := start + chunkSize
		end = min(end, size)

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				population[j] = creator()
			}
		}(start, end) // chunk to use
	}

	wg.Wait()
	return population
}
