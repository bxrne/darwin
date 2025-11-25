package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/evolution"
	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
)

type MetricsHandler func(metrics.GenerationMetrics)
type MetricsComplete chan struct{}

func getGenomeType(config *cfg.Config) individual.GenomeType {
	if config.BitString.Enabled {
		return individual.BitStringGenome
	} else if config.Tree.Enabled {
		return individual.TreeGenome
	} else if config.GrammarTree.Enabled {
		return individual.GrammarTreeGenome
	}
	return -1 // or panic/error
}

// runEvolution encapsulates the shared evolution logic.
// It takes a context, config, and optional metrics handler.
// Returns the final population, a completion channel, and an error.
func runEvolution(ctx context.Context, config *cfg.Config, handler MetricsHandler) ([]individual.Evolvable, MetricsComplete, error) {
	// Seed the RNG for reproducible results
	rng.Seed(config.Evolution.Seed)

	metricsChan := make(chan metrics.GenerationMetrics, 100)
	cmdChan := make(chan evolution.EvolutionCommand, 10)
	metricsComplete := make(chan struct{})

	populationType := getGenomeType(config)
	grammar := individual.CreateGrammar(config.Tree.TerminalSet, config.Tree.VariableSet, config.Tree.OperandSet)
	fitnessInfo := fitness.GenerateFitnessInfoFromConfig(config, populationType, grammar)
	fitnessCalculator := fitness.FitnessCalculatorFactory(fitnessInfo)

	popBuilder := evolution.NewPopulationBuilder()
	population := popBuilder.BuildPopulation(config.Evolution.PopulationSize, func() individual.Evolvable {
		switch populationType {
		case individual.BitStringGenome:
			return individual.NewBinaryIndividual(config.BitString.GenomeSize)
		case individual.TreeGenome:
			return individual.NewRandomTree(config.Tree.InitalDepth, config.Tree.OperandSet, config.Tree.VariableSet, config.Tree.TerminalSet)
		case individual.GrammarTreeGenome:
			return individual.NewGrammarTree(config.GrammarTree.GenomeSize)
		case individual.ActionTreeGenome:
			return individual.NewActionTreeIndividual(config.ActionTree.Actions, config.ActionTree.NumInputs, map[string]*individual.Tree)
		default:
			return nil
		}
	}, fitnessCalculator)
	var selector selection.Selector
	switch config.Evolution.SelectionType {
	case "tournament":
		selector = selection.NewTournamentSelector(config.Evolution.SelectionSize)
	case "roulette":
		selector = selection.NewRouletteSelector(config.Evolution.SelectionSize)

	}
	metricsStreamer := metrics.NewMetricsStreamer(metricsChan)
	var metricsSubscriber <-chan metrics.GenerationMetrics
	if handler != nil {
		metricsSubscriber = metricsStreamer.Subscribe()
	}
	crossoverInformation := individual.CrossoverInformation{CrossoverPoints: config.Evolution.CrossoverPointCount, MaxDepth: config.Tree.MaxDepth}
	mutateInformation := individual.MutateInformation{OperandSet: config.Tree.OperandSet, TerminalSet: config.Tree.TerminalSet, VariableSet: config.Tree.VariableSet}
	evolutionEngine := evolution.NewEvolutionEngine(population, selector, metricsChan, cmdChan, fitnessCalculator, crossoverInformation, mutateInformation)

	metricsStreamer.Start(ctx)
	evolutionEngine.Start(ctx)

	// Handle metrics if handler provided
	if handler != nil {
		go func() {
			defer close(metricsComplete)
			for {
				select {
				case <-ctx.Done():
					return
				case m, ok := <-metricsSubscriber:
					if !ok {
						return
					}
					handler(m)
					// Signal completion when we've processed the last generation
					if m.Generation == config.Evolution.Generations {
						return
					}
				}
			}
		}()
	} else {
		close(metricsComplete) // Close immediately if no handler
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
			return nil, metricsComplete, fmt.Errorf("timeout sending evolution command for generation %d", gen)
		}
	}

	close(cmdChan)
	evolutionEngine.Wait()
	metricsStreamer.Stop()

	finalPop := evolutionEngine.GetPopulation()
	return finalPop, metricsComplete, nil
}
