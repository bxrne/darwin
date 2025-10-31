package main

import (
	"fmt"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/garden"
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

	populationManager := garden.NewPopulationManager(cfg.Evolution.PopulationSize, cfg.Evolution.GenomeSize)
	logmgr.Info("Starting evolution", logmgr.Field("config", cfg.Evolution))

	for gen := 1; gen <= cfg.Evolution.Generations; gen++ {
		populationManager.Step(gen, cfg.Evolution.CrossoverPointCount, cfg.Evolution.MutationPoints, cfg.Evolution.MutationRate, cfg.Evolution.ElitismPercentage)

		if gen%10 == 0 || gen == cfg.Evolution.Generations {
			if latest := populationManager.GetLatestMetrics(); latest != nil {
				logmgr.Debug("Generation completed",
					logmgr.Field("generation", latest.Generation),
					logmgr.Field("duration_ms", latest.Duration.Milliseconds()),
					logmgr.Field("best_fitness", fmt.Sprintf("%.3f", latest.BestFitness)),
					logmgr.Field("avg_fitness", fmt.Sprintf("%.3f", latest.AvgFitness)),
				)
			}
		}
	}

	logmgr.Info("Evolution complete", logmgr.Field("population_summary", populationManager.Summary()))
}
