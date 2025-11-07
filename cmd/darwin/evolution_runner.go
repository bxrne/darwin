package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/evolution"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
)

type MetricsHandler func(metrics.GenerationMetrics)

func getGenomeType(config *cfg.Config) individual.GenomeType {
	if config.BitString.Enabled {
		return individual.BitStringGenome
	} else if config.Tree.Enabled {
		return individual.TreeGenome
	}
	return -1 // or panic/error
}

// runEvolution encapsulates the shared evolution logic.
// It takes a context, config, and optional metrics handler.
// Returns the final population or an error.
func runEvolution(ctx context.Context, config *cfg.Config, handler MetricsHandler) ([]individual.Evolvable, error) {
	// Seed the RNG for reproducible results
	rng.Seed(config.Evolution.Seed)

	metricsChan := make(chan metrics.GenerationMetrics, 100)
	cmdChan := make(chan evolution.EvolutionCommand, 10)

	populationType := getGenomeType(config)
	fitnessCalculator := individual.FitnessCalculatorFactory(individual.FitnessSetupInformation{GenomeType: populationType, EvalFunction: config.Tree.TargetFunction, TerminalSet: config.Tree.TerminalSet})

	popBuilder := evolution.NewPopulationBuilder()
	population := popBuilder.BuildPopulation(config.Evolution.PopulationSize, func() individual.Evolvable {
		switch populationType {
		case individual.BitStringGenome:
			return individual.NewBinaryIndividual(config.BitString.GenomeSize)
		case individual.TreeGenome:
			return individual.NewRandomTree(config.Tree.MaxDepth, config.Tree.PrimitiveSet, config.Tree.TerminalSet)
		default:
			return nil
		}
	}, fitnessCalculator)

	selector := selection.NewRouletteSelector(30)
	metricsStreamer := metrics.NewMetricsStreamer(metricsChan)
	var metricsSubscriber <-chan metrics.GenerationMetrics
	if handler != nil {
		metricsSubscriber = metricsStreamer.Subscribe()
	}

	evolutionEngine := evolution.NewEvolutionEngine(population, selector, metricsChan, cmdChan, fitnessCalculator)

	metricsStreamer.Start(ctx)
	evolutionEngine.Start(ctx)

	// Handle metrics if handler provided
	if handler != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case m, ok := <-metricsSubscriber:
					if !ok {
						return
					}
					handler(m)
				}
			}
		}()
	}

	// Send evolution commands
	for gen := 1; gen <= config.Evolution.Generations; gen++ {
		cmd := evolution.EvolutionCommand{
			Type:            evolution.CmdStartGeneration,
			Generation:      gen,
			CrossoverPoints: config.Evolution.CrossoverPointCount,
			CrossoverRate:   config.Evolution.CrossoverRate,
			MutationRate:    config.Evolution.MutationRate,
			ElitismPct:      config.Evolution.ElitismPercentage,
		}

		select {
		case cmdChan <- cmd:
		case <-time.After(5 * time.Second):
			return nil, fmt.Errorf("timeout sending evolution command for generation %d", gen)
		}
	}

	close(cmdChan)
	evolutionEngine.Wait()
	metricsStreamer.Stop()

	finalPop := evolutionEngine.GetPopulation()
	return finalPop, nil
}
