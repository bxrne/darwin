package evolution

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
)

// EvolutionEngine manages the evolution process using channels
type EvolutionEngine struct {
	population           Population
	selector             selection.Selector
	metricsChan          chan<- metrics.GenerationMetrics
	cmdChan              <-chan EvolutionCommand
	done                 chan struct{}
	currentGen           int
	fitnessCalculator    fitness.FitnessCalculator
	crossoverInformation individual.CrossoverInformation
	mutateInformation    individual.MutateInformation
}

// NewEvolutionEngine creates a new evolution engine
func NewEvolutionEngine(
	population Population,
	selector selection.Selector,
	metricsChan chan<- metrics.GenerationMetrics,
	cmdChan <-chan EvolutionCommand,
	fitnessCalculator fitness.FitnessCalculator,
	crossoverInformation individual.CrossoverInformation,
	mutateInformation individual.MutateInformation,
) *EvolutionEngine {
	return &EvolutionEngine{
		population:           population,
		selector:             selector,
		metricsChan:          metricsChan,
		cmdChan:              cmdChan,
		done:                 make(chan struct{}),
		currentGen:           0,
		fitnessCalculator:    fitnessCalculator,
		crossoverInformation: crossoverInformation,
		mutateInformation:    mutateInformation,
	}
}

// Start begins processing evolution commands
func (ee *EvolutionEngine) Start(ctx context.Context) {
	go func() {
		defer close(ee.done)

		for {
			select {
			case <-ctx.Done():
				return
			case cmd, ok := <-ee.cmdChan:
				if !ok {
					return
				}

				switch cmd.Type {
				case CmdStartGeneration:
					ee.processGeneration(cmd)
				case CmdStop:
					return
				}
			}
		}
	}()
}

// Wait blocks until the engine is done
func (ee *EvolutionEngine) Wait() {
	<-ee.done
}

// GetPopulation returns the current population
func (ee *EvolutionEngine) GetPopulation() Population {
	return ee.population
}

func (ee *EvolutionEngine) generateOffspring(cmd EvolutionCommand, out chan<- individual.Evolvable) {
	parent1 := ee.selector.Select(ee.population.GetPopulation())
	parent2 := ee.selector.Select(ee.population.GetPopulation())
	// Perform crossover and mutation
	// Create copies of parents to avoid mutating the original population
	parentCopy1 := parent1.Clone()
	parentCopy2 := parent2.Clone()
	if 1 > rng.Float64() {

		child1, child2 := parentCopy1.MultiPointCrossover(parentCopy2, &ee.crossoverInformation)

		ee.fitnessCalculator.CalculateFitness(&child1)
		ee.fitnessCalculator.CalculateFitness(&child2)
		out <- child1.Max(child2)
		return
	}

	parentCopy1.Mutate(cmd.MutationRate, &ee.mutateInformation)
	ee.fitnessCalculator.CalculateFitness(&parentCopy1)

	parentCopy2.Mutate(cmd.MutationRate, &ee.mutateInformation)
	ee.fitnessCalculator.CalculateFitness(&parentCopy2)

	out <- parentCopy1.Max(parentCopy2)
}

// processGeneration performs one generation of evolution
func (ee *EvolutionEngine) processGeneration(cmd EvolutionCommand) {
	start := time.Now()

	// Sort population by fitness (descending)
	ee.sortPopulation()
	// Create new population
	newPop := make([]individual.Evolvable, 0, ee.population.Count())
	// Elitism: keep best individuals
	elitismCount := int(float64(ee.population.Count()) * cmd.ElitismPct)
	for i := 0; i < elitismCount && i < ee.population.Count(); i++ {
		newPop = append(newPop, ee.population.Get(i))
	}
	offspringNeeded := ee.population.Count() - len(newPop)
	offspringChan := make(chan individual.Evolvable, ee.population.Count()-elitismCount+1)
	var wg sync.WaitGroup
	// Generate offspring
	for i := 0; i < offspringNeeded; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ee.generateOffspring(cmd, offspringChan)
		}()
	}
	go func() {
		wg.Wait()
		close(offspringChan)
	}()
	for ind := range offspringChan {
		newPop = append(newPop, ind)
	}
	ee.population.SetPopulation(newPop)
	duration := time.Since(start)
	// Calculate and send metrics
	genMetrics := ee.calculateMetrics(cmd.Generation, duration)
	select {
	case ee.metricsChan <- genMetrics:
	default:
		// Skip if metrics channel is full (non-blocking)
	}
}

// sortPopulation sorts the population by fitness (descending)
func (ee *EvolutionEngine) sortPopulation() {
	sort.SliceStable(ee.population.GetPopulation(), func(i, j int) bool {
		return ee.population.Get(i).GetFitness() > ee.population.Get(j).GetFitness()
	})
}

// calculateMetrics computes generation metrics
func (ee *EvolutionEngine) calculateMetrics(generation int, duration time.Duration) metrics.GenerationMetrics {
	if ee.population.Count() == 0 {
		return metrics.GenerationMetrics{
			Generation:     generation,
			Duration:       duration,
			PopulationSize: 0,
			Timestamp:      time.Now(),
		}
	}

	totalFitness := 0.0
	maxFitness := ee.population.Get(0).GetFitness()
	minFitness := ee.population.Get(0).GetFitness()

	for i := range ee.population.Count() {
		fitness := ee.population.Get(i).GetFitness()
		totalFitness += fitness
		if fitness > maxFitness {
			maxFitness = fitness
		}
		if fitness < minFitness {
			minFitness = fitness
		}
	}

	avgFitness := totalFitness / float64(ee.population.Count())
	bestDescription := ee.population.Get(0).Describe()
	minDepth := -1
	maxDepth := -1
	totalDepth := 0.0
	for citizenIndex := range ee.population.Count() {
		if tree, ok := ee.population.Get(citizenIndex).(*individual.Tree); ok {
			depth := tree.GetDepth()
			if minDepth == -1 || depth < minDepth {
				minDepth = depth
			}
			if maxDepth == -1 || depth > maxDepth {
				maxDepth = depth
			}
			totalDepth += float64(depth)
		}
	}
	avgDepth := 0.0
	if minDepth != -1 && maxDepth != -1 {
		avgDepth = totalDepth / float64(ee.population.Count())
	}

	return metrics.GenerationMetrics{
		Generation:      generation,
		Duration:        duration,
		BestFitness:     maxFitness,
		AvgFitness:      avgFitness,
		MinFitness:      minFitness,
		MaxFitness:      maxFitness,
		BestDescription: bestDescription,
		MinDepth:        minDepth,
		MaxDepth:        maxDepth,
		AvgDepth:        avgDepth,
		PopulationSize:  ee.population.Count(),
		Timestamp:       time.Now(),
	}
}
