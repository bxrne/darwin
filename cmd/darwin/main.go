package main

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/bxrne/logmgr"
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

type Individual struct {
	Genome  string
	Fitness int
}

func (i *Individual) CalculateFitness() {
	count := 0
	for _, gene := range i.Genome {
		if gene == '1' {
			count++
		}
	}
	i.Fitness = count
}

func (i *Individual) Mutate(points []int) {
	for _, point := range points {
		if i.Genome[point] == '1' {
			i.Genome = i.Genome[:point] + "0" + i.Genome[point+1:]
		} else {
			i.Genome = i.Genome[:point] + "1" + i.Genome[point+1:]
		}
	}
}

func main() {
	defer logmgr.Shutdown()
	logmgr.AddSink(logmgr.DefaultConsoleSink) // Console output

	populationSize := 500
	maxGenerations := 50
	crossoverRate := 0.9
	mutationPoints := rand.Perm(12)[:2] // Mutate 2 random points
	mutationRate := 0.05
	genomeSize := 12

	logmgr.Info("Starting...", logmgr.Field("population", populationSize), logmgr.Field("max generations", maxGenerations), logmgr.Field("crossover rate", crossoverRate), logmgr.Field("mutation rate", mutationRate))

	population := NewPopulation(populationSize, genomeSize)

	for range maxGenerations {
		population.Step(crossoverRate, mutationPoints, mutationRate)
	}

	maxFit, minFit := 0, genomeSize
	for _, individual := range population {
		if individual.Fitness > maxFit {
			maxFit = individual.Fitness
		}
		if individual.Fitness < minFit {
			minFit = individual.Fitness
		}
	}

	logmgr.Info("Generation complete", logmgr.Field("max fitness", maxFit), logmgr.Field("min fitness", minFit))
}
