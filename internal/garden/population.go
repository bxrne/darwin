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
		pop[i] = makeIndividual(genomeSize)
	}
	return pop
}

func (p *Population) Roulette(input_amount int) Individual {
	rouletteTable := make([]Individual, 0, input_amount)
	total := 0
	for range input_amount {
		randIndex := rand.Intn(len(*p))
		rouletteTable = append(rouletteTable, (*p)[randIndex])
		total = total + (*p)[randIndex].Fitness
	}
	runningTotal := 0
	randomValue := rand.Intn(total)
	for i := range input_amount {
		runningTotal = rouletteTable[i].Fitness
		if runningTotal > randomValue {
			return rouletteTable[i]
		}
	}
	return rouletteTable[len(rouletteTable)-1]
}

func (p *Population) Tournament(inputAmount int) Individual {
	tournamentPop := make([]Individual, 0, inputAmount)
	for range inputAmount {
		randIndex := rand.Intn(len(*p))
		tournamentPop = append(tournamentPop, (*p)[randIndex])
	}
	max := tournamentPop[0]
	for _, ind := range tournamentPop[1:] {
		max = ind.Max(max)
	}
	return max
}

func (p *Population) Sort() {
	sort.SliceStable(*p, func(i, j int) bool {
		return (*p)[i].Fitness > (*p)[j].Fitness
	})
}

func (p *Population) Step(crossoverPointCount int, mutationPoints []int, mutationRate float64, elistimPercentage float64) {
	newPop := make(Population, 0, len(*p))
	p.Sort()
	elitismAmount := len(*p) - int(float64(len(*p))*elistimPercentage)
	copy(newPop[:elitismAmount], (*p)[:elitismAmount])
	for len(newPop) < cap(newPop) {
		parent1 := p.Roulette(30)
		parent2 := p.Roulette(30)

		// Perform crossover and mutation
		child1, child2 := parent1.MultiPointCrossover(parent2, crossoverPointCount)
		child1.Mutate(mutationPoints, mutationRate)
		child2.Mutate(mutationPoints, mutationRate)
		newPop = append(newPop, child1.Max(child2))
	}
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
