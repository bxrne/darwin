package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/evolution"
	"github.com/bxrne/darwin/internal/fitness"
	"github.com/bxrne/darwin/internal/individual"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/population"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
	"go.uber.org/zap"
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
// It takes a context, config, optional metrics handler, and logger.
// Returns the final population, a completion channel, and an error.
func RunEvolution(ctx context.Context, config *cfg.Config, handler MetricsHandler, logger *zap.Logger) ([]individual.Evolvable, MetricsComplete, error) {
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

	metricsChan := make(chan metrics.GenerationMetrics, config.Evolution.Generations)
	cmdChan := make(chan evolution.EvolutionCommand, config.Evolution.Generations)
	metricsComplete := make(chan struct{})

	populationType := getGenomeType(config)
	grammar := individual.CreateGrammar(config.Tree.TerminalSet, config.Tree.VariableSet, config.Tree.OperandSet)

	popBuilder := population.NewPopulationBuilder()
	popinfo := population.NewPopulationInfo(config, populationType)

	// Atomic counter for ramped half-and-half initialization
	var treeCounter atomic.Int64

	// Helper function to create ramped half-and-half tree
	// Depth 0 is disallowed, so we distribute from depth 1 to initialDepth
	createRampedHalfAndHalfTree := func(initialDepth int, operandSet, variableSet, terminalSet []string) *individual.Tree {
		popSize := config.Evolution.PopulationSize
		index := int(treeCounter.Add(1) - 1) // Get current index (0-based)

		// Calculate depth group: divide population into initialDepth groups (depths 1 to initialDepth)
		// Depth 0 is disallowed
		depthGroups := initialDepth
		if depthGroups <= 0 {
			depthGroups = 1 // Ensure at least one group
		}
		groupSize := popSize / depthGroups
		remainder := popSize % depthGroups

		// Determine which depth group this individual belongs to (1 to initialDepth)
		depth := 1 // Start from depth 1
		groupIndex := index
		for d := 0; d < depthGroups; d++ {
			groupCount := groupSize
			if d < remainder {
				groupCount++ // Distribute remainder across first groups
			}
			if groupIndex < groupCount {
				depth = d + 1 // Depth is 1-indexed (1 to initialDepth)
				break
			}
			groupIndex -= groupCount
		}

		// Within the depth group, determine if we use grow (first half) or full (second half)
		// Recalculate group boundaries for this specific depth
		groupStart := 0
		for d := 0; d < (depth - 1); d++ { // depth - 1 because depth is 1-indexed
			groupCount := groupSize
			if d < remainder {
				groupCount++
			}
			groupStart += groupCount
		}
		groupCount := groupSize
		if (depth - 1) < remainder {
			groupCount++
		}
		localIndex := index - groupStart
		useGrow := localIndex < groupCount/2

		return individual.NewRampedHalfAndHalfTree(depth, useGrow, operandSet, variableSet, terminalSet)
	}

	population := popBuilder.BuildPopulation(&popinfo, func() individual.Evolvable {
		switch populationType {
		case individual.BitStringGenome:
			return individual.NewBinaryIndividual(config.BitString.GenomeSize)
		case individual.TreeGenome:
			return createRampedHalfAndHalfTree(config.Tree.InitalDepth, config.Tree.OperandSet, config.Tree.VariableSet, config.Tree.TerminalSet)
		case individual.GrammarTreeGenome:
			return individual.NewGrammarTree(config.GrammarTree.GenomeSize)
		case individual.ActionTreeGenome:
			// Create random trees for each action using ramped half-and-half
			// All trees in an individual use the same depth and method
			// Depth 0 is disallowed, so we distribute from depth 1 to initialDepth
			index := int(treeCounter.Add(1) - 1)
			popSize := config.Evolution.PopulationSize
			initialDepth := config.Tree.InitalDepth

			// Calculate depth and method for this individual
			// Depth 0 is disallowed, so we have initialDepth groups (depths 1 to initialDepth)
			depthGroups := initialDepth
			if depthGroups <= 0 {
				depthGroups = 1 // Ensure at least one group
			}
			groupSize := popSize / depthGroups
			remainder := popSize % depthGroups

			depth := 1 // Start from depth 1
			groupIndex := index
			for d := 0; d < depthGroups; d++ {
				groupCount := groupSize
				if d < remainder {
					groupCount++
				}
				if groupIndex < groupCount {
					depth = d + 1 // Depth is 1-indexed (1 to initialDepth)
					break
				}
				groupIndex -= groupCount
			}

			groupStart := 0
			for d := 0; d < (depth - 1); d++ { // depth - 1 because depth is 1-indexed
				groupCount := groupSize
				if d < remainder {
					groupCount++
				}
				groupStart += groupCount
			}
			groupCount := groupSize
			if (depth - 1) < remainder {
				groupCount++
			}
			localIndex := index - groupStart
			useGrow := localIndex < groupCount/2

			initialTrees := make(map[string]*individual.Tree)
			variableSet := make([]string, config.ActionTree.WeightsColumnCount)
			for i := range config.ActionTree.WeightsColumnCount {
				key := fmt.Sprintf("w%d", i)
				variableSet[i] = key
			}
			variableSet = append(variableSet, config.Tree.VariableSet...)
			for _, action := range config.ActionTree.Actions {
				tree := individual.NewRampedHalfAndHalfTree(depth, useGrow, config.Tree.OperandSet, variableSet, config.Tree.TerminalSet)
				initialTrees[action.Name] = tree
			}
			result := individual.NewActionTreeIndividual(config.ActionTree.Actions, initialTrees)
			return result
		default:
			fmt.Printf("Unknown genome type: %v\n", populationType)
			return nil
		}
	})

	fitnessInfo := fitness.GenerateFitnessInfoFromConfig(config, populationType, grammar, population.GetPopulations())
	fitnessCalculator := fitness.FitnessCalculatorFactoryWithConfig(fitnessInfo, config)

	logger.Info("Calculating initial population fitness", zap.Int("population_size", population.Count()))
	population.CalculateFitnesses(fitnessCalculator)
	logger.Info("Initial population fitness calculation complete")

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
	mutateInformation := individual.MutateInformation{OperandSet: config.Tree.OperandSet, TerminalSet: config.Tree.TerminalSet, VariableSet: config.Tree.VariableSet, MaxDepth: config.Tree.MaxDepth}
	evolutionEngine := evolution.NewEvolutionEngine(population, selector, metricsChan, cmdChan, fitnessCalculator, crossoverInformation, mutateInformation, logger)

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
			logger.Error("Failed to cleanup fitness calculator", zap.Error(err))
		}
	}

	finalPop := evolutionEngine.GetPopulation()
	return finalPop, metricsComplete, nil
}
