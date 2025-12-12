package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/individual"
	"go.uber.org/zap"
)

type benchmarkCase struct {
	name           string
	popSize        int
	sizeParam      int
	generations    int
	individualType string
}

var benchmarkCases = []benchmarkCase{
	{"BitString_Small", 100, 50, 10, "bitstring"},
	{"BitString_Medium", 500, 200, 50, "bitstring"},
	{"BitString_Large", 1000, 500, 100, "bitstring"},
	{"BitString_Huge", 500, 5000, 10, "bitstring"},
	{"Tree_Small", 100, 3, 10, "tree"},
	{"Tree_Medium", 500, 5, 50, "tree"},
	{"Tree_Large", 1000, 7, 100, "tree"},
	{"Tree_Huge", 500, 10, 10, "tree"},
	{"GrammarTree_Small", 100, 50, 10, "grammar_tree"},
	{"GrammarTree_Medium", 500, 100, 50, "grammar_tree"},
	{"GrammarTree_Large", 1000, 200, 100, "grammar_tree"},
	{"GrammarTree_Huge", 500, 500, 10, "grammar_tree"},
	{"ActionTree", 5, 2, 1, "action_tree"},
	{"Compare_BitString", 300, 150, 30, "bitstring"},
	{"Compare_Tree", 300, 4, 30, "tree"},
	{"Compare_GrammarTree", 300, 75, 30, "grammar_tree"},
}

func BenchmarkEvolution(b *testing.B) {
	for _, bc := range benchmarkCases {
		b.Run(bc.name, func(b *testing.B) {
			config := newBenchmarkConfig(bc.popSize, bc.sizeParam, bc.generations, bc.individualType)
			runBenchmark(b, config)
		})
	}
}

func newBenchmarkConfig(popSize, sizeParam, generations int, individualType string) *cfg.Config {
	// Base evolution config for all benchmarks
	mutationRate := 0.05
	crossoverRate := 0.9
	if individualType == "tree" || individualType == "grammar_tree" {
		// More conservative rates for symbolic regression
		mutationRate = 0.1
		crossoverRate = 0.8
	}
	evolutionConfig := cfg.EvolutionConfig{
		PopulationSize:      popSize,
		CrossoverPointCount: 1,
		MutationRate:        mutationRate,
		CrossoverRate:       crossoverRate,
		Generations:         generations,
		ElitismPercentage:   0.1,
		SelectionSize:       5,
		SelectionType:       "tournament",
		Seed:                42,
	}

	switch individualType {
	case "bitstring":
		return &cfg.Config{
			Evolution: evolutionConfig,
			BitString: cfg.BitStringIndividualConfig{
				Enabled:    true,
				GenomeSize: sizeParam,
			},
			Tree:        cfg.TreeIndividualConfig{Enabled: false},
			GrammarTree: cfg.GrammarTreeConfig{Enabled: false},
			ActionTree:  cfg.ActionTreeConfig{Enabled: false},
		}

	case "tree":
		return &cfg.Config{
			Evolution: evolutionConfig,
			Tree: cfg.TreeIndividualConfig{
				Enabled:     true,
				MaxDepth:    sizeParam,
				OperandSet:  []string{"+", "-", "*"},
				VariableSet: []string{"x", "y"},
				TerminalSet: []string{"0.0", "1.0", "2.0", "3.0", "4.0", "5.0"},
			},
			BitString:   cfg.BitStringIndividualConfig{Enabled: false},
			GrammarTree: cfg.GrammarTreeConfig{Enabled: false},
			ActionTree:  cfg.ActionTreeConfig{Enabled: false},
			Fitness: cfg.FitnessConfig{
				TestCaseCount:  20,
				TargetFunction: "x + y", // Simpler function for better convergence
			},
		}

	case "grammar_tree":
		return &cfg.Config{
			Evolution: evolutionConfig,
			GrammarTree: cfg.GrammarTreeConfig{
				Enabled:    true,
				GenomeSize: sizeParam,
			},
			Tree: cfg.TreeIndividualConfig{
				Enabled:     false, // Keep disabled but provide grammar config
				MaxDepth:    8,
				OperandSet:  []string{"+", "-", "*"},
				VariableSet: []string{"x", "y"},
				TerminalSet: []string{"0.0", "1.0", "2.0", "3.0", "4.0", "5.0"},
			},
			BitString:  cfg.BitStringIndividualConfig{Enabled: false},
			ActionTree: cfg.ActionTreeConfig{Enabled: false},
			Fitness: cfg.FitnessConfig{
				TestCaseCount:  20,
				TargetFunction: "x + y", // Simpler function for better convergence
			},
		}

	case "action_tree":
		actions := []individual.ActionTuple{
			{Name: "move_north", Value: 1},
			{Name: "move_south", Value: 2},
			{Name: "move_east", Value: 3},
			{Name: "move_west", Value: 4},
		}
		// Ensure SwitchTrainingTargetStep is at least 1 to avoid divide by zero
		switchStep := max(generations/2, 1)

		// For ActionTree: User wants trees 5, weights 2, test cases 2
		treePopulation := 5 // 5 trees
		weightsCount := 2   // 2 weights
		return &cfg.Config{
			Evolution: cfg.EvolutionConfig{
				PopulationSize:      treePopulation,
				CrossoverPointCount: 1,
				MutationRate:        0.05,
				CrossoverRate:       0.9,
				Generations:         generations,
				ElitismPercentage:   0.1,
				SelectionSize:       5,
				SelectionType:       "tournament",
				Seed:                42,
			},
			ActionTree: cfg.ActionTreeConfig{
				Enabled:                  true,
				Actions:                  actions,
				WeightsCount:             weightsCount,
				WeightsColumnCount:       4, // Fixed matrix dimensions for 4 actions
				ServerAddr:               "localhost:5000",
				OpponentType:             "random",
				MaxSteps:                 100,
				ConnectionPoolSize:       5,
				ConnectionTimeout:        "5s",
				HealthCheckTimeout:       "5s",
				SwitchTrainingTargetStep: switchStep,
				TrainWeightsFirst:        false,
			},
			BitString: cfg.BitStringIndividualConfig{Enabled: false},
			Tree: cfg.TreeIndividualConfig{
				Enabled:     false, // Keep disabled but provide grammar config
				MaxDepth:    5,
				OperandSet:  []string{"+", "-", "*", "/"},
				VariableSet: []string{"army_diff", "land_diff", "distance_to_enemy_general"},
				TerminalSet: []string{"0.1", "0.5", "1.0", "2.0"},
			},
			GrammarTree: cfg.GrammarTreeConfig{Enabled: false},
			Fitness: cfg.FitnessConfig{
				TestCaseCount:  2, // 2 test cases as requested
				TargetFunction: "",
			},
		}

	default:
		// Fallback to bitstring for unknown types
		return &cfg.Config{
			Evolution: evolutionConfig,
			BitString: cfg.BitStringIndividualConfig{
				Enabled:    true,
				GenomeSize: sizeParam,
			},
			Tree:        cfg.TreeIndividualConfig{Enabled: false},
			GrammarTree: cfg.GrammarTreeConfig{Enabled: false},
			ActionTree:  cfg.ActionTreeConfig{Enabled: false},
		}
	}
}

func runBenchmark(b *testing.B, config *cfg.Config) {
	// Check ActionTree prerequisites
	if config.ActionTree.Enabled {
		b.Logf("ActionTree benchmark requires game server at %s", config.ActionTree.ServerAddr)
		b.Logf("To start game server: cd game && uv venv && uv sync && uv run main.py")
	}

	// Report config
	var sizeDesc string
	var individualType string
	switch {
	case config.BitString.Enabled:
		sizeDesc = fmt.Sprintf("GenomeSize=%d", config.BitString.GenomeSize)
		individualType = "BitString"
	case config.Tree.Enabled:
		sizeDesc = fmt.Sprintf("MaxDepth=%d", config.Tree.MaxDepth)
		individualType = "Tree"
	case config.GrammarTree.Enabled:
		sizeDesc = fmt.Sprintf("GenomeSize=%d", config.GrammarTree.GenomeSize)
		individualType = "GrammarTree"
	case config.ActionTree.Enabled:
		sizeDesc = fmt.Sprintf("WeightsCount=%d, Actions=%d", config.ActionTree.WeightsCount, len(config.ActionTree.Actions))
		individualType = "ActionTree"
	default:
		sizeDesc = "Unknown"
		individualType = "Unknown"
	}

	if config.Fitness.TargetFunction != "" {
		sizeDesc += fmt.Sprintf(", TargetFunction=%s", config.Fitness.TargetFunction)
	}

	b.Logf("Config: Type=%s, Population=%d, %s, Generations=%d, Seed=%d, MutationRate=%.3f, Elitism=%.3f",
		individualType,
		config.Evolution.PopulationSize,
		sizeDesc,
		config.Evolution.Generations,
		config.Evolution.Seed,
		config.Evolution.MutationRate,
		config.Evolution.ElitismPercentage)

	// Memory stats before
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	b.ReportAllocs()

	startTime := time.Now()

	// Run evolution b.N times
	for i := 0; b.Loop(); i++ {
		logger := zap.NewNop() // Use no-op logger for benchmarks
		finalPop, _, err := RunEvolution(b.Context(), config, nil, logger)
		if err != nil {
			b.Fatalf("Evolution failed: %v", err.Error())
		}

		if len(finalPop) > 0 {
			bestFitness := 0.0
			totalFitness := 0.0
			minFitness := finalPop[0].GetFitness()

			for _, ind := range finalPop {
				fitness := ind.GetFitness()
				totalFitness += fitness
				if fitness > bestFitness {
					bestFitness = fitness
				}
				if fitness < minFitness {
					minFitness = fitness
				}
			}

			avgFitness := totalFitness / float64(len(finalPop))

			b.Logf("Run %d: Best=%.3f, Avg=%.3f, Min=%.3f", i+1, bestFitness, avgFitness, minFitness)
		}
	}

	totalTime := time.Since(startTime)

	// Memory stats after
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Report memory usage
	memUsed := memStatsAfter.Alloc - memStatsBefore.Alloc
	totalAlloc := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
	sysMem := memStatsAfter.Sys

	b.Logf("Memory: Used=%d bytes, TotalAlloc=%d bytes, Sys=%d bytes",
		memUsed, totalAlloc, sysMem)

	b.Logf("Total time for %d runs: %v (avg: %v per run)",
		b.N, totalTime, totalTime/time.Duration(b.N))
}
