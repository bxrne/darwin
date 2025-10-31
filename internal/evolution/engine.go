package evolution

import (
	"context"
	"sort"
	"time"

	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/selection"
)

// EvolutionEngine manages the evolution process using channels
type EvolutionEngine struct {
	population  []individual.Evolvable
	selector    selection.Selector
	metricsChan chan<- metrics.GenerationMetrics
	cmdChan     <-chan EvolutionCommand
	done        chan struct{}
	currentGen  int
}

// NewEvolutionEngine creates a new evolution engine
func NewEvolutionEngine(
	population []individual.Evolvable,
	selector selection.Selector,
	metricsChan chan<- metrics.GenerationMetrics,
	cmdChan <-chan EvolutionCommand,
) *EvolutionEngine {
	return &EvolutionEngine{
		population:  population,
		selector:    selector,
		metricsChan: metricsChan,
		cmdChan:     cmdChan,
		done:        make(chan struct{}),
		currentGen:  0,
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

// processGeneration performs one generation of evolution
func (ee *EvolutionEngine) processGeneration(cmd EvolutionCommand) {
	start := time.Now()

	// Sort population by fitness (descending)
	ee.sortPopulation()

	// Create new population
	newPop := make([]individual.Evolvable, 0, len(ee.population))

	// Elitism: keep best individuals
	elitismCount := int(float64(len(ee.population)) * cmd.ElitismPct)
	for i := 0; i < elitismCount && i < len(ee.population); i++ {
		newPop = append(newPop, ee.population[i])
	}

	// Generate offspring
	for len(newPop) < cap(newPop) {
		parent1 := ee.selector.Select(ee.population)
		parent2 := ee.selector.Select(ee.population)

		// Perform crossover and mutation
		child1, child2 := parent1.MultiPointCrossover(parent2, cmd.CrossoverPoints)
		child1.Mutate(cmd.MutationPoints, cmd.MutationRate)
		child2.Mutate(cmd.MutationPoints, cmd.MutationRate)

		// Select the better child
		betterChild := child1.Max(child2)
		newPop = append(newPop, betterChild)
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

	return metrics.GenerationMetrics{
		Generation:     generation,
		Duration:       duration,
		BestFitness:    maxFitness,
		AvgFitness:     avgFitness,
		MinFitness:     minFitness,
		MaxFitness:     maxFitness,
		PopulationSize: len(ee.population),
		Timestamp:      time.Now(),
	}
}
