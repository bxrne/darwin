package garden

import (
	"fmt"
	"math/rand"
	"sort"
)

type Population []Evolvable

func NewPopulation(size int, genomeSize int) Population {
	pop := make(Population, size)
	for i := range pop {
		pop[i] = newBinaryIndividual(genomeSize)
	}
	return pop
}

func (p *Population) Roulette(input_amount int) Evolvable {
	rouletteTable := make([]Evolvable, 0, input_amount)
	total := 0.0
	for range input_amount {
		randIndex := rand.Intn(len(*p))
		rouletteTable = append(rouletteTable, (*p)[randIndex])
		total = total + (*p)[randIndex].GetFitness()
	}
	runningTotal := 0.0
	randomValue := rand.Float64() * total
	for i := range input_amount {
		runningTotal = rouletteTable[i].GetFitness()
		if runningTotal > randomValue {
			return rouletteTable[i]
		}
	}
	return rouletteTable[len(rouletteTable)-1]
}

func (p *Population) Tournament(inputAmount int) Evolvable {
	tournamentPop := make([]Evolvable, 0, inputAmount)
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
		return (*p)[i].GetFitness() > (*p)[j].GetFitness()
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
	totalFitness := 0.0
	for _, individual := range *p {
		totalFitness += individual.GetFitness()
	}
	avgFitness := float64(totalFitness) / float64(len(*p))
	maxFitness := 0.0
	for _, individual := range *p {
		if individual.GetFitness() > maxFitness {
			maxFitness = individual.GetFitness()
		}
	}
	minFitness := maxFitness
	for _, individual := range *p {
		if individual.GetFitness() < minFitness {
			minFitness = individual.GetFitness()
		}
	}
	return fmt.Sprintf("Population Summary: Size=%d, Avg Fitness=%.2f, Max Fitness=%.2f, Min Fitness=%.2f", len(*p), avgFitness, maxFitness, minFitness)

}
