package garden

import (
	"fmt"
	"math/rand"
	"sort"
)

type Population []Individual

func NewPopulation(size int, genomeSize int) Population {
	pop := make(Population, size)
	for i := range pop {
		pop[i] = Individual{
			Genome: fmt.Sprintf("%0*b", genomeSize, rand.Intn(1<<genomeSize)),
		}
	}
	return pop
}

func (p *Population) Sort() {
	sort.SliceStable(*p, func(i, j int) bool {
		return (*p)[i].Fitness > (*p)[j].Fitness
	})
}

func (p *Population) Step(crossoverRate float64, mutationPoints []int, mutationRate float64) {
	for i := range *p {
		(*p)[i].CalculateFitness()
		if rand.Float64() < mutationRate {
			(*p)[i].Mutate(mutationPoints)
		}
	}

	newPop := NewPopulation(len(*p), len((*p)[0].Genome))
	p.Sort()
	copy(newPop[:10], (*p)[:10])
	*p = newPop
}

func (p *Population) Summary() string {
	totalFitness := 0
	for _, individual := range *p {
		totalFitness += individual.Fitness
	}
	avgFitness := float64(totalFitness) / float64(len(*p))
	maxFitness := 0
	for _, individual := range *p {
		if individual.Fitness > maxFitness {
			maxFitness = individual.Fitness
		}
	}
	minFitness := maxFitness
	for _, individual := range *p {
		if individual.Fitness < minFitness {
			minFitness = individual.Fitness
		}
	}
	return fmt.Sprintf("Population Summary: Size=%d, Avg Fitness=%.2f, Max Fitness=%d, Min Fitness=%d", len(*p), avgFitness, maxFitness, minFitness)

}
