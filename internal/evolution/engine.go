package evolution

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/population"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
	"go.uber.org/zap"
)

// EvolutionEngine manages the evolution process using channels
type EvolutionEngine struct {
	population           population.Population
	selector             selection.Selector
	metricsChan          chan<- metrics.GenerationMetrics
	cmdChan              <-chan EvolutionCommand
	done                 chan struct{}
	currentGen           int
	fitnessCalculator    fitness.FitnessCalculator
	crossoverInformation individual.CrossoverInformation
	mutateInformation    individual.MutateInformation
	logger               *zap.Logger
}

// NewEvolutionEngine creates a new evolution engine
func NewEvolutionEngine(
	population population.Population,
	selector selection.Selector,
	metricsChan chan<- metrics.GenerationMetrics,
	cmdChan <-chan EvolutionCommand,
	fitnessCalculator fitness.FitnessCalculator,
	crossoverInformation individual.CrossoverInformation,
	mutateInformation individual.MutateInformation,
	logger *zap.Logger,
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
		logger:               logger,
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
					// Log immediately when command is received (before processing starts)
					ee.processGeneration(cmd)
				case CmdStop:
					return
				}
			}
		}
	}()
}

// GetPopulation returns the current population
func (ee *EvolutionEngine) GetPopulation() []individual.Evolvable {
	return ee.population.GetPopulation()
}

// Wait blocks until the engine is done
func (ee *EvolutionEngine) Wait() {
	<-ee.done
}

func (ee *EvolutionEngine) generateOffspring(cmd EvolutionCommand, out chan<- individual.Evolvable) {
	parent1 := ee.selector.Select(ee.population.GetPopulation())
	parent2 := ee.selector.Select(ee.population.GetPopulation())
	// Perform crossover and mutation
	// Create copies of parents to avoid mutating the original population
	parentCopy1 := parent1.Clone()
	parentCopy2 := parent2.Clone()
	// Crossover with configured probability; otherwise mutate
	if rng.Float64() < cmd.CrossoverRate {
		child1, child2 := parentCopy1.MultiPointCrossover(parentCopy2, &ee.crossoverInformation)
		// Mutate children post-crossover
		child1.Mutate(cmd.MutationRate, &ee.mutateInformation)
		child2.Mutate(cmd.MutationRate, &ee.mutateInformation)
		out <- child1
		out <- child2
		return
	}

	parentCopy1.Mutate(cmd.MutationRate, &ee.mutateInformation)
	parentCopy2.Mutate(cmd.MutationRate, &ee.mutateInformation)
	out <- parentCopy1
	out <- parentCopy2
}

// processGeneration performs one generation of evolution
func (ee *EvolutionEngine) processGeneration(cmd EvolutionCommand) {
	start := time.Now()
	ee.logger.Info("Starting generation", zap.Int("generation", cmd.Generation))

	// For generation 1, calculate fitness for the initial population first
	// (initial population doesn't have fitness calculated yet)
	if cmd.Generation == 1 {
		ee.logger.Info("Calculating fitness for initial population (generation 1)", zap.Int("population_size", ee.population.Count()))
		ee.population.CalculateFitnesses(ee.fitnessCalculator)
		ee.logger.Info("Initial population fitness calculation complete")
	}
	// Sort population by fitness (descending)
	ee.sortPopulation()
	// Create new population
	newPop := make([]individual.Evolvable, 0, ee.population.Count())
	// Elitism: keep best individuals
	elitismCount := max(int(float64(ee.population.Count())*cmd.ElitismPct), 1)
	for i := 0; i < elitismCount && i < ee.population.Count(); i++ {
		newPop = append(newPop, ee.population.Get(i))
	}
	offspringNeeded := ee.population.Count() - len(newPop)
	offspringChanCount := (offspringNeeded + 1) / 2
	offspringChan := make(chan individual.Evolvable)
	var wg sync.WaitGroup
	// Generate offspring
	for i := 0; i < offspringChanCount; i++ {
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
	count := 0
	for ind := range offspringChan {
		if count < offspringNeeded {
			newPop = append(newPop, ind)
			count++
		} else {
			break
		}
	}
	ee.population.SetPopulation(newPop)
	ee.population.Update(cmd.Generation)
	ee.population.CalculateFitnesses(ee.fitnessCalculator)
	duration := time.Since(start)
	// Calculate and send metrics
	genMetrics := ee.calculateMetrics(cmd.Generation, duration)

	// Send metrics before logging completion to ensure proper ordering
	select {
	case ee.metricsChan <- genMetrics:
	default:
		// Skip if metrics channel is full (non-blocking)
	}

	// Log completion after metrics are sent to ensure ordering
	ee.logger.Info("Generation completed", zap.Int("generation", cmd.Generation), zap.Int64("duration_ms", duration.Milliseconds()))
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
	metricsMap := ee.population.Get(0).GetMetrics()
	totalMetricValues := make(map[string]float64, len(metricsMap))
	minMetricValues := make(map[string]float64, len(metricsMap))
	maxMetricValues := make(map[string]float64, len(metricsMap))

	for key := range metricsMap {
		totalMetricValues[key] = 0
		minMetricValues[key] = math.Inf(1)  // +∞
		maxMetricValues[key] = math.Inf(-1) // -∞
	}
	for i := range ee.population.Count() {
		currentMetricsMap := ee.population.Get(i).GetMetrics()
		for j, value := range currentMetricsMap {
			totalMetricValues[j] += value
			if value > maxMetricValues[j] {
				maxMetricValues[j] = value
			}
			if minMetricValues[j] > value {
				minMetricValues[j] = value
			}
		}
	}

	overallMetrics := make(map[string]float64, len(metricsMap)*3)

	for key, _ := range totalMetricValues {
		avgKey := fmt.Sprintf("avg_%s", key)
		minKey := fmt.Sprintf("min_%s", key)
		maxKey := fmt.Sprintf("max_%s", key)
		overallMetrics[avgKey] = totalMetricValues[key] / float64(ee.population.Count())
		overallMetrics[minKey] = minMetricValues[key]
		overallMetrics[maxKey] = maxMetricValues[key]
	}

	ee.sortPopulation()
	bestDescription := ee.population.Get(0).Describe()

	return metrics.GenerationMetrics{
		Generation:      generation,
		Duration:        duration,
		BestDescription: bestDescription,
		Metrics:         overallMetrics,
		PopulationSize:  ee.population.Count(),
		Timestamp:       time.Now(),
	}
}
