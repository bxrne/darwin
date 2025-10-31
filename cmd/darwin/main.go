package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/evolution"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/darwin/internal/selection"
	"github.com/bxrne/logmgr"
)

func main() {
	defer logmgr.Shutdown()
	logmgr.AddSink(logmgr.DefaultConsoleSink)
	logmgr.SetLevel(logmgr.DebugLevel)

	cfg, err := cfg.LoadConfig("config/default.toml")
	if err != nil {
		logmgr.Fatal("Failed to load config", logmgr.Field("error", err))
	}

	metricsChan := make(chan metrics.GenerationMetrics, 100)
	cmdChan := make(chan evolution.EvolutionCommand, 10)

	popBuilder := evolution.NewPopulationBuilder()
	population := popBuilder.BuildBinaryPopulation(cfg.Evolution.PopulationSize, cfg.Evolution.GenomeSize)

	selector := selection.NewRouletteSelector(30)

	metricsStreamer := metrics.NewMetricsStreamer(metricsChan)
	metricsSubscriber := metricsStreamer.Subscribe()

	evolutionEngine := evolution.NewEvolutionEngine(population, selector, metricsChan, cmdChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logmgr.Info("Starting evolution", logmgr.Field("config", cfg.Evolution))

	metricsStreamer.Start(ctx)
	evolutionEngine.Start(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-metricsSubscriber:
				if !ok {
					return
				}
				if m.Generation%10 == 0 || m.Generation == cfg.Evolution.Generations {
					logmgr.Debug("Generation completed",
						logmgr.Field("generation", m.Generation),
						logmgr.Field("duration_ms", m.Duration.Milliseconds()),
						logmgr.Field("best_fitness", fmt.Sprintf("%.3f", m.BestFitness)),
						logmgr.Field("avg_fitness", fmt.Sprintf("%.3f", m.AvgFitness)),
					)
				}
			}
		}
	}()

	// Send evolution commands
	for gen := 1; gen <= cfg.Evolution.Generations; gen++ {
		cmd := evolution.EvolutionCommand{
			Type:            evolution.CmdStartGeneration,
			Generation:      gen,
			CrossoverPoints: cfg.Evolution.CrossoverPointCount,
			MutationPoints:  cfg.Evolution.MutationPoints,
			MutationRate:    cfg.Evolution.MutationRate,
			ElitismPct:      cfg.Evolution.ElitismPercentage,
		}

		select {
		case cmdChan <- cmd:
		case <-time.After(5 * time.Second):
			logmgr.Error("Timeout sending evolution command", logmgr.Field("generation", gen))
			cancel()
			return
		}
	}

	close(cmdChan)
	evolutionEngine.Wait()
	metricsStreamer.Stop()
	wg.Wait()

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

		logmgr.Info("Evolution complete",
			logmgr.Field("population_size", len(finalPop)),
			logmgr.Field("best_fitness", fmt.Sprintf("%.3f", bestFitness)),
			logmgr.Field("avg_fitness", fmt.Sprintf("%.3f", avgFitness)),
			logmgr.Field("min_fitness", fmt.Sprintf("%.3f", minFitness)),
		)
	}

	logmgr.Info("Evolution finished successfully")
}
