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

// buildActionTreePopulationWithRampedHalfAndHalf creates a population using ramped half-and-half initialization
// This ensures diversity by distributing trees across different depths and using both grow and full methods
func buildActionTreePopulationWithRampedHalfAndHalf(popInfo *population.PopulationInfo, config *cfg.Config) population.Population {
	// Use the existing population builder but with a custom creator that uses ramped half-and-half
	popBuilder := population.NewPopulationBuilder()
	
	// Track which individual we're creating to determine depth and method
	individualCounter := 0
	
	// Ramped half-and-half: distribute across depths from 1 to initialDepth
	minDepth := 1
	maxDepth := config.Tree.InitalDepth
	if maxDepth < minDepth {
		maxDepth = minDepth
	}
	
	// Calculate how many individuals per depth
	depthRange := maxDepth - minDepth + 1
	individualsPerDepth := popInfo.Size / depthRange
	remaining := popInfo.Size % depthRange
	
	// Helper to get depth and method for current individual
	getDepthAndMethod := func(index int) (depth int, useGrow bool) {
		// Find which depth this individual belongs to
		currentCount := 0
		for d := minDepth; d <= maxDepth; d++ {
			count := individualsPerDepth
			if d-minDepth < remaining {
				count++
			}
			// Half grow, half full for each depth
			growCount := count / 2
			
			if index < currentCount+growCount {
				return d, true // grow
			}
			if index < currentCount+count {
				return d, false // full
			}
			currentCount += count
		}
		// Fallback
		return maxDepth, rng.Float64() < 0.5
	}
	
	creator := func() individual.Evolvable {
		depth, useGrow := getDepthAndMethod(individualCounter)
		individualCounter++
		
		// Prepare variable set
		variableSet := make([]string, config.ActionTree.WeightsColumnCount)
		for i := range config.ActionTree.WeightsColumnCount {
			key := fmt.Sprintf("w%d", i)
			variableSet[i] = key
		}
		variableSet = append(variableSet, config.Tree.VariableSet...)
		
		// Create trees for each action using ramped half-and-half
		initialTrees := make(map[string]*individual.Tree)
		for _, action := range config.ActionTree.Actions {
			tree := individual.NewRampedHalfAndHalfTree(
				depth,
				useGrow,
				config.Tree.OperandSet,
				variableSet,
				config.Tree.TerminalSet,
			)
			initialTrees[action.Name] = tree
		}
		return individual.NewActionTreeIndividual(config.ActionTree.Actions, initialTrees)
	}
	
	return popBuilder.BuildPopulation(popInfo, creator)
}

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
	grammar := individual.CreateGrammar(config.Tree.TerminalSet, config.Tree.VariableSet, config.Tree.OperandSet)

	popBuilder := population.NewPopulationBuilder()
	popinfo := population.NewPopulationInfo(config, populationType)
	
	// For ActionTree, use ramped half-and-half initialization
	var population population.Population
	if populationType == individual.ActionTreeGenome {
		population = buildActionTreePopulationWithRampedHalfAndHalf(&popinfo, config)
	} else {
		population = popBuilder.BuildPopulation(&popinfo, func() individual.Evolvable {
			switch populationType {
			case individual.BitStringGenome:
				return individual.NewBinaryIndividual(config.BitString.GenomeSize)
			case individual.TreeGenome:
				return individual.NewRandomTree(config.Tree.InitalDepth, config.Tree.OperandSet, config.Tree.VariableSet, config.Tree.TerminalSet)
			case individual.GrammarTreeGenome:
				return individual.NewGrammarTree(config.GrammarTree.GenomeSize)
			default:
				fmt.Printf("Unknown genome type: %v\n", populationType)
				return nil
			}
		})
	}

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
