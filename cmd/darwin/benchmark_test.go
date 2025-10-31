package main

import (
	"runtime"
	"testing"
	"time"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/evolution"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/rng"
	"github.com/bxrne/darwin/internal/selection"
)

func BenchmarkEvolution(b *testing.B) {
	// Use same config as medium benchmark (from default.toml)
	config := &cfg.Config{
		Evolution: cfg.EvolutionConfig{
			PopulationSize:      500,
			GenomeSize:          200,
			CrossoverPointCount: 1,
			MutationRate:        0.05,
			MutationPoints:      []int{6},
			Generations:         50,
			ElitismPercentage:   0.1,
			Seed:                42,
		},
	}

	// Report config
	b.Logf("Config: Population=%d, GenomeSize=%d, Generations=%d, Seed=%d, MutationRate=%.3f, Elitism=%.3f",
		config.Evolution.PopulationSize,
		config.Evolution.GenomeSize,
		config.Evolution.Generations,
		config.Evolution.Seed,
		config.Evolution.MutationRate,
		config.Evolution.ElitismPercentage)

	// Memory stats before
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	startTime := time.Now()

	// Run evolution b.N times
	for i := 0; i < b.N; i++ {
		// Seed RNG for each run
		rng.Seed(config.Evolution.Seed)

		// Setup components for each run
		metricsChan := make(chan metrics.GenerationMetrics, 100)
		cmdChan := make(chan evolution.EvolutionCommand, 10)

		popBuilder := evolution.NewPopulationBuilder()
		population := popBuilder.BuildBinaryPopulation(config.Evolution.PopulationSize, config.Evolution.GenomeSize)

		selector := selection.NewRouletteSelector(30)

		metricsStreamer := metrics.NewMetricsStreamer(metricsChan)
		evolutionEngine := evolution.NewEvolutionEngine(population, selector, metricsChan, cmdChan)

		// Start metrics and evolution
		metricsStreamer.Start(b.Context())
		evolutionEngine.Start(b.Context())

		// Send evolution commands
		for gen := 1; gen <= config.Evolution.Generations; gen++ {
			cmd := evolution.EvolutionCommand{
				Type:            evolution.CmdStartGeneration,
				Generation:      gen,
				CrossoverPoints: config.Evolution.CrossoverPointCount,
				MutationPoints:  config.Evolution.MutationPoints,
				MutationRate:    config.Evolution.MutationRate,
				ElitismPct:      config.Evolution.ElitismPercentage,
			}

			select {
			case cmdChan <- cmd:
			case <-time.After(5 * time.Second):
				b.Fatalf("Timeout sending evolution command generation %d", gen)
			}
		}

		close(cmdChan)
		evolutionEngine.Wait()
		metricsStreamer.Stop()

		// Get final results
		finalPop := evolutionEngine.GetPopulation()
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

func BenchmarkEvolution_Small(b *testing.B) {
	benchmarkWithConfig(b, 100, 50, 10) // pop, genome, gens
}

func BenchmarkEvolution_Medium(b *testing.B) {
	benchmarkWithConfig(b, 500, 200, 50)
}

func BenchmarkEvolution_Large(b *testing.B) {
	benchmarkWithConfig(b, 1000, 500, 100)
}

func benchmarkWithConfig(b *testing.B, popSize, genomeSize, generations int) {
	// Create config programmatically
	config := &cfg.Config{
		Evolution: cfg.EvolutionConfig{
			PopulationSize:      popSize,
			GenomeSize:          genomeSize,
			CrossoverPointCount: 1,
			MutationRate:        0.05,
			MutationPoints:      []int{genomeSize / 10}, // 10% of genome
			Generations:         generations,
			ElitismPercentage:   0.1,
			Seed:                42,
		},
	}

	// Config is hardcoded, assume valid

	// Report config
	b.Logf("Config: Population=%d, GenomeSize=%d, Generations=%d, Seed=%d",
		config.Evolution.PopulationSize,
		config.Evolution.GenomeSize,
		config.Evolution.Generations,
		config.Evolution.Seed)

	// Memory stats
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	b.ResetTimer()
	b.ReportAllocs()

	startTime := time.Now()

	for i := 0; i < b.N; i++ {
		// Seed RNG for each run to ensure reproducibility
		rng.Seed(config.Evolution.Seed)

		// Setup components for each run
		metricsChan := make(chan metrics.GenerationMetrics, 100)
		cmdChan := make(chan evolution.EvolutionCommand, 10)

		popBuilder := evolution.NewPopulationBuilder()
		population := popBuilder.BuildBinaryPopulation(config.Evolution.PopulationSize, config.Evolution.GenomeSize)

		selector := selection.NewRouletteSelector(30)

		metricsStreamer := metrics.NewMetricsStreamer(metricsChan)
		evolutionEngine := evolution.NewEvolutionEngine(population, selector, metricsChan, cmdChan)

		// Start
		metricsStreamer.Start(b.Context())
		evolutionEngine.Start(b.Context())

		// Commands
		for gen := 1; gen <= config.Evolution.Generations; gen++ {
			cmd := evolution.EvolutionCommand{
				Type:            evolution.CmdStartGeneration,
				Generation:      gen,
				CrossoverPoints: config.Evolution.CrossoverPointCount,
				MutationPoints:  config.Evolution.MutationPoints,
				MutationRate:    config.Evolution.MutationRate,
				ElitismPct:      config.Evolution.ElitismPercentage,
			}

			select {
			case cmdChan <- cmd:
			case <-time.After(5 * time.Second):
				b.Fatalf("Timeout in generation %d", gen)
			}
		}

		close(cmdChan)
		evolutionEngine.Wait()
		metricsStreamer.Stop()

		// Results
		finalPop := evolutionEngine.GetPopulation()
		if len(finalPop) > 0 {
			bestFitness := 0.0
			totalFitness := 0.0

			for _, ind := range finalPop {
				fitness := ind.GetFitness()
				totalFitness += fitness
				if fitness > bestFitness {
					bestFitness = fitness
				}
			}

			avgFitness := totalFitness / float64(len(finalPop))
			b.Logf("Run %d: Best=%.3f, Avg=%.3f", i+1, bestFitness, avgFitness)
		}
	}

	totalTime := time.Since(startTime)

	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	memUsed := memStatsAfter.Alloc - memStatsBefore.Alloc
	totalAlloc := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc

	b.Logf("Memory: Used=%d bytes, TotalAlloc=%d bytes", memUsed, totalAlloc)
	b.Logf("Total time: %v (avg: %v per run)", totalTime, totalTime/time.Duration(b.N))
}
