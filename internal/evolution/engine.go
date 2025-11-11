package evolution

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
)

// EvolutionEngine manages the evolution process using channels
type EvolutionEngine struct {
	population        []individual.Evolvable
	selector          selection.Selector
	metricsChan       chan<- metrics.GenerationMetrics
	cmdChan           <-chan EvolutionCommand
	done              chan struct{}
	currentGen        int
	fitnessCalculator individual.FitnessCalculator
	primitiveSet      []string
	terminalSet       []string
}

// NewEvolutionEngine creates a new evolution engine
func NewEvolutionEngine(
	population []individual.Evolvable,
	selector selection.Selector,
	metricsChan chan<- metrics.GenerationMetrics,
	cmdChan <-chan EvolutionCommand,
	fitnessCalculator individual.FitnessCalculator,
	primitiveSet []string,
	terminalSet []string,
) *EvolutionEngine {
	return &EvolutionEngine{
		population:        population,
		selector:          selector,
		metricsChan:       metricsChan,
		cmdChan:           cmdChan,
		done:              make(chan struct{}),
		currentGen:        0,
		fitnessCalculator: fitnessCalculator,
		primitiveSet:      primitiveSet,
		terminalSet:       terminalSet,
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
func (ee *EvolutionEngine) GetPopulation() []individual.Evolvable {
	return ee.population
}

func (ee *EvolutionEngine) generateOffspring(cmd EvolutionCommand, out chan<- individual.Evolvable) {
	parent1 := ee.selector.Select(ee.population)
	parent2 := ee.selector.Select(ee.population)
	// Perform crossover and mutation
	// Create copies of parents to avoid mutating the original population
	parentCopy1 := parent1.Clone()
	parentCopy2 := parent2.Clone()
	if 1 > rng.Float64() {
		child1, child2 := parentCopy1.MultiPointCrossover(parentCopy2, cmd.CrossoverPoints)
		ee.fitnessCalculator.CalculateFitness(&child1)
		ee.fitnessCalculator.CalculateFitness(&child2)
		out <- child1.Max(child2)
		return
	}

	// Handle mutation based on individual type
	if tree1, ok := parentCopy1.(*individual.Tree); ok {
		tree1.MutateWithSets(cmd.MutationRate, ee.primitiveSet, ee.terminalSet)
	} else {
		parentCopy1.Mutate(cmd.MutationRate)
	}
	ee.fitnessCalculator.CalculateFitness(&parentCopy1)

	if tree2, ok := parentCopy2.(*individual.Tree); ok {
		tree2.MutateWithSets(cmd.MutationRate, ee.primitiveSet, ee.terminalSet)
	} else {
		parentCopy2.Mutate(cmd.MutationRate)
	}
	ee.fitnessCalculator.CalculateFitness(&parentCopy2)
	out <- parentCopy1.Max(parentCopy2)
}

// processGeneration performs one generation of evolution
func (ee *EvolutionEngine) processGeneration(cmd EvolutionCommand) {
	start := time.Now()

	// Sort population by fitness (descending)
	ee.sortPopulation()
	//o, ok := ee.population[0].(*individual.Tree)
	//if ok {
	//
	//	individual.PrintTreeJSON(o)
	//}
	// Create new population
	newPop := make([]individual.Evolvable, 0, len(ee.population))
	// Elitism: keep best individuals
	elitismCount := int(float64(len(ee.population)) * cmd.ElitismPct)
	for i := 0; i < elitismCount && i < len(ee.population); i++ {
		newPop = append(newPop, ee.population[i])
	}
	offspringNeeded := len(ee.population) - len(newPop)
	offspringChan := make(chan individual.Evolvable, len(ee.population)-elitismCount+1)
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
	ee.population = newPop
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
	sort.SliceStable(ee.population, func(i, j int) bool {
		return ee.population[i].GetFitness() > ee.population[j].GetFitness()
	})
}

// calculateMetrics computes generation metrics
func (ee *EvolutionEngine) calculateMetrics(generation int, duration time.Duration) metrics.GenerationMetrics {
	if len(ee.population) == 0 {
		return metrics.GenerationMetrics{
			Generation:     generation,
			Duration:       duration,
			PopulationSize: 0,
			Timestamp:      time.Now(),
		}
	}

	totalFitness := 0.0
	maxFitness := ee.population[0].GetFitness()
	minFitness := ee.population[0].GetFitness()

	for _, individual := range ee.population {
		fitness := individual.GetFitness()
		totalFitness += fitness
		if fitness > maxFitness {
			maxFitness = fitness
		}
		if fitness < minFitness {
			minFitness = fitness
		}
	}

	avgFitness := totalFitness / float64(len(ee.population))
	bestDescription := ee.population[0].Describe()
	minDepth := -1
	maxDepth := -1
	totalDepth := 0.0
	for _, citizen := range ee.population {
		if tree, ok := citizen.(*individual.Tree); ok {
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
		avgDepth = totalDepth / float64(len(ee.population))
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
		PopulationSize:  len(ee.population),
		Timestamp:       time.Now(),
	}
}
