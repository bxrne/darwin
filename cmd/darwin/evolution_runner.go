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
	"github.com/bxrne/darwin/internal/population"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
	"github.com/bxrne/logmgr"
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
	} else if config.ActionTree.Enabled {
		return individual.ActionTreeGenome
	}
	return -1 // or panic/error
}

// RunEvolution encapsulates the shared evolution logic.
// It takes a context, config, and optional metrics handler.
// Returns the final population, a completion channel, and an error.
func RunEvolution(ctx context.Context, config *cfg.Config, handler MetricsHandler) ([]individual.Evolvable, MetricsComplete, error) {
	// pre evolution srv heartbeat
	if config.ActionTree.Enabled {
		timeout := 5 * time.Second
		if parsedTimeout, err := time.ParseDuration(config.ActionTree.ConnectionTimeout); err == nil {
			timeout = parsedTimeout
		}

		healthChecker := fitness.NewServerHealthChecker(config.ActionTree.ServerAddr, timeout)
		if err := healthChecker.CheckServerHealthWithRetry(); err != nil {
			return nil, nil, fmt.Errorf("server health check failed: %w", err)
		}
	}

	rng.Seed(config.Evolution.Seed)

	metricsChan := make(chan metrics.GenerationMetrics, 100)
	cmdChan := make(chan evolution.EvolutionCommand, 10)
	metricsComplete := make(chan struct{})

	populationType := getGenomeType(config)
	fmt.Printf("Population type: %v\n", populationType)
	grammar := individual.CreateGrammar(config.Tree.TerminalSet, config.Tree.VariableSet, config.Tree.OperandSet)

	popBuilder := population.NewPopulationBuilder()
	popinfo := population.NewPopulationInfo(config, populationType)
	population := popBuilder.BuildPopulation(&popinfo, func() individual.Evolvable {
		switch populationType {
		case individual.BitStringGenome:
			return individual.NewBinaryIndividual(config.BitString.GenomeSize)
		case individual.TreeGenome:
			return individual.NewRandomTree(config.Tree.InitalDepth, config.Tree.OperandSet, config.Tree.VariableSet, config.Tree.TerminalSet)
		case individual.GrammarTreeGenome:
			return individual.NewGrammarTree(config.GrammarTree.GenomeSize)
		case individual.ActionTreeGenome:
			fmt.Printf("Creating ActionTree individual\n")
			// Create random trees for each action
			initialTrees := make(map[string]*individual.Tree)
			variableSet := make([]string, config.ActionTree.WeightsColumnCount)
			for i := range config.ActionTree.WeightsColumnCount {
				key := fmt.Sprintf("w%d", i)
				variableSet[i] = key
			}
			variableSet = append(variableSet, config.Tree.VariableSet...)
			for _, action := range config.ActionTree.Actions {
				tree := individual.NewRandomTree(
					config.Tree.InitalDepth,
					config.Tree.OperandSet,
					variableSet,
					config.Tree.TerminalSet,
				)
				initialTrees[action] = tree
			}
			result := individual.NewActionTreeIndividual(config.ActionTree.Actions, initialTrees)
			fmt.Printf("ActionTree individual created: %v\n", result != nil)
			return result
		default:
			fmt.Printf("Unknown genome type: %v\n", populationType)
			return nil
		}
	})

	fitnessInfo := fitness.GenerateFitnessInfoFromConfig(config, populationType, grammar, population.GetPopulations())
	fitnessCalculator := fitness.FitnessCalculatorFactoryWithConfig(fitnessInfo, config)

	population.CalculateFitnesses(fitnessCalculator)

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

	// Clean up fitness calculator resources
	if cleanupCalc, ok := fitnessCalculator.(interface{ Close() error }); ok {
		if err := cleanupCalc.Close(); err != nil {
			logmgr.Error("Failed to cleanup fitness calculator", logmgr.Field("error", err))
		}
	}

	finalPop := evolutionEngine.GetPopulation()
	return finalPop, metricsComplete, nil
}
