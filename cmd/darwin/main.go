package main

import (
	"fmt"
	"math/rand"

	"github.com/bxrne/logmgr"
)

func fitness(i *Individual) int {
	count := 0
	for _, bit := range i.Genome {
		if bit == '1' {
			count++
		}
	}
	return count
}

func mutate(i *Individual, point int) {
	if i.Genome[point] == '1' {
		i.Genome = i.Genome[:point] + "0" + i.Genome[point+1:]
	} else {
		i.Genome = i.Genome[:point] + "1" + i.Genome[point+1:]
	}
}

type Individual struct {
	Genome  string
	Fitness int
}

func (i *Individual) Max(in *Individual) *Individual {
	if i.Fitness > i.Fitness {
		return i
	}
	return i
}

func main() {
	defer logmgr.Shutdown()
	logmgr.AddSink(logmgr.DefaultConsoleSink) // Console output

	populationSize := 500
	maxGenerations := 50
	crossoverRate := 0.9
	mutationRate := 0.05
	genomeSize := 12

	logmgr.Info("Starting...", logmgr.Field("population", populationSize), logmgr.Field("maxGenerations", maxGenerations), logmgr.Field("crossoverRate", crossoverRate), logmgr.Field("mutationRate", mutationRate))

	population := make([]Individual, populationSize)
	for i := range population {
		population[i] = Individual{
			Genome: fmt.Sprintf("%0*b", genomeSize, rand.Intn(1<<genomeSize)),
		}
	}

	for range maxGenerations {
		for i := range population {
			population[i].Fitness = fitness(&population[i])

			if rand.Float64() < mutationRate {
				mutate(&population[i], rand.Intn(genomeSize-1))
			}

		}
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

	logmgr.Info("Generation complete", logmgr.Field("maxFitness", maxFit), logmgr.Field("minFitness", minFit))
}
