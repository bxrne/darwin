package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/metrics"
	"github.com/bxrne/logmgr"
)

func main() {
	defer logmgr.Shutdown()
	logmgr.AddSink(logmgr.DefaultConsoleSink)
	logmgr.SetLevel(logmgr.DebugLevel)

	configPath := flag.String("config", "config/default.toml", "Path to config file")
	flag.Parse()

	cfg, err := cfg.LoadConfig(*configPath)
	if err != nil {
		logmgr.Fatal("Failed to load config", logmgr.Field("error", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logmgr.Info("Starting evolution", logmgr.Field("config", cfg.Evolution))

	// Define metrics handler for logging
	handler := func(m metrics.GenerationMetrics) {
		if m.Generation%10 == 0 || m.Generation == cfg.Evolution.Generations {
			logmgr.Debug("Generation completed",
				logmgr.Field("generation", m.Generation),
				logmgr.Field("duration_ms", m.Duration.Milliseconds()),
				logmgr.Field("best_fitness", fmt.Sprintf("%.3f", m.BestFitness)),
				logmgr.Field("avg_fitness", fmt.Sprintf("%.3f", m.AvgFitness)),
			)
		}
	}

	finalPop, metricsComplete, err := runEvolution(ctx, cfg, handler)
	if err != nil {
		logmgr.Fatal("Evolution failed", logmgr.Field("error", err))
	}

	// Wait for metrics to finish processing before calculating final stats
	<-metricsComplete

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
