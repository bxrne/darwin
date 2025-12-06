package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/bxrne/darwin/internal/cfg"
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
	{"Tree_Small", 100, 3, 10, "tree"},
	{"BitString_Medium", 500, 200, 50, "bitstring"},
	{"Tree_Medium", 500, 5, 50, "tree"},
	{"BitString_Large", 1000, 500, 100, "bitstring"},
	{"Tree_Large", 1000, 7, 100, "tree"},
	{"BitString_Huge", 500, 5000, 10, "bitstring"},
	{"Tree_Huge", 500, 10, 10, "tree"},
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
	var bitString cfg.BitStringIndividualConfig
	var tree cfg.TreeIndividualConfig

	if individualType == "bitstring" {
		bitString = cfg.BitStringIndividualConfig{
			Enabled:    true,
			GenomeSize: sizeParam,
		}
		tree.Enabled = false
	} else if individualType == "tree" {
		tree = cfg.TreeIndividualConfig{
			Enabled:     true,
			MaxDepth:    sizeParam,
			OperandSet:  []string{"+", "-", "*", "/"},
			VariableSet: []string{"x"},
			TerminalSet: []string{"1.0", "2.0"},
		}
		bitString.Enabled = false
	}

	return &cfg.Config{
		Evolution: cfg.EvolutionConfig{
			PopulationSize:      popSize,
			CrossoverPointCount: 1,
			MutationRate:        0.05,
			CrossoverRate:       0.9,
			Generations:         generations,
			ElitismPercentage:   0.1,
			Seed:                42,
		},
		BitString: bitString,
		Tree:      tree,
	}
}

func runBenchmark(b *testing.B, config *cfg.Config) {
	// Report config
	sizeDesc := ""
	if config.BitString.Enabled {
		sizeDesc = fmt.Sprintf("GenomeSize=%d", config.BitString.GenomeSize)
	} else if config.Tree.Enabled {
		sizeDesc = fmt.Sprintf("MaxDepth=%d", config.Tree.MaxDepth)
	}
	b.Logf("Config: Population=%d, %s, Generations=%d, Seed=%d, MutationRate=%.3f, Elitism=%.3f",
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
		finalPop, _, err := RunEvolution(b.Context(), config, nil)
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
