package garden

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type GenerationMetrics struct {
	Generation     int
	Duration       time.Duration
	BestFitness    float64
	AvgFitness     float64
	MinFitness     float64
	MaxFitness     float64
	PopulationSize int
}

type Population []Evolvable

type PopulationManager struct {
	Population Population
	Metrics    []GenerationMetrics
}

func NewPopulation(size int, genomeSize int) Population {
	pop := make(Population, size)
	for i := range pop {
		pop[i] = newBinaryIndividual(genomeSize)
	}
	return pop
}

func NewPopulationManager(size int, genomeSize int) *PopulationManager {
	return &PopulationManager{
		Population: NewPopulation(size, genomeSize),
		Metrics:    make([]GenerationMetrics, 0),
	}
}

func (pm *PopulationManager) Roulette(input_amount int) Evolvable {
	rouletteTable := make([]Evolvable, 0, input_amount)
	total := 0.0
	for range input_amount {
		randIndex := rand.Intn(len(pm.Population))
		rouletteTable = append(rouletteTable, pm.Population[randIndex])
		total = total + pm.Population[randIndex].GetFitness()
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

func (pm *PopulationManager) Tournament(inputAmount int) Evolvable {
	tournamentPop := make([]Evolvable, 0, inputAmount)
	for range inputAmount {
		randIndex := rand.Intn(len(pm.Population))
		tournamentPop = append(tournamentPop, pm.Population[randIndex])
	}
	max := tournamentPop[0]
	for _, ind := range tournamentPop[1:] {
		max = ind.Max(max)
	}
	return max
}

func (pm *PopulationManager) Sort() {
	sort.SliceStable(pm.Population, func(i, j int) bool {
		return pm.Population[i].GetFitness() > pm.Population[j].GetFitness()
	})
}

func (pm *PopulationManager) Step(generation int, crossoverPointCount int, mutationPoints []int, mutationRate float64, elistimPercentage float64) {
	start := time.Now()

	newPop := make(Population, 0, len(pm.Population))
	pm.Sort()
	elitismAmount := len(pm.Population) - int(float64(len(pm.Population))*elistimPercentage)
	copy(newPop[:elitismAmount], pm.Population[:elitismAmount])
	for len(newPop) < cap(newPop) {
		parent1 := pm.Roulette(30)
		parent2 := pm.Roulette(30)

		// Perform crossover and mutation
		child1, child2 := parent1.MultiPointCrossover(parent2, crossoverPointCount)
		child1.Mutate(mutationPoints, mutationRate)
		child2.Mutate(mutationPoints, mutationRate)
		newPop = append(newPop, child1.Max(child2))
	}
	pm.Population = newPop

	duration := time.Since(start)

	// Collect metrics
	metrics := pm.calculateMetrics(generation, duration)
	pm.Metrics = append(pm.Metrics, metrics)
}

func (pm *PopulationManager) calculateMetrics(generation int, duration time.Duration) GenerationMetrics {
	totalFitness := 0.0
	maxFitness := 0.0
	minFitness := pm.Population[0].GetFitness()

	for _, individual := range pm.Population {
		fitness := individual.GetFitness()
		totalFitness += fitness
		if fitness > maxFitness {
			maxFitness = fitness
		}
		if fitness < minFitness {
			minFitness = fitness
		}
	}

	avgFitness := totalFitness / float64(len(pm.Population))

	return GenerationMetrics{
		Generation:     generation,
		Duration:       duration,
		BestFitness:    maxFitness,
		AvgFitness:     avgFitness,
		MinFitness:     minFitness,
		MaxFitness:     maxFitness,
		PopulationSize: len(pm.Population),
	}
}

func (pm *PopulationManager) Summary() string {
	if len(pm.Metrics) == 0 {
		return "No metrics available"
	}
	latest := pm.Metrics[len(pm.Metrics)-1]
	return fmt.Sprintf("Population Summary: Size=%d, Avg Fitness=%.2f, Max Fitness=%.2f, Min Fitness=%.2f",
		latest.PopulationSize, latest.AvgFitness, latest.MaxFitness, latest.MinFitness)
}

func (pm *PopulationManager) GetLatestMetrics() *GenerationMetrics {
	if len(pm.Metrics) == 0 {
		return nil
	}
	return &pm.Metrics[len(pm.Metrics)-1]
}

func (pm *PopulationManager) GetAllMetrics() []GenerationMetrics {
	return pm.Metrics
}
