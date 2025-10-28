package main

import (
	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/garden"
	"github.com/bxrne/logmgr"
)

func main() {
	defer logmgr.Shutdown()
	logmgr.AddSink(logmgr.DefaultConsoleSink)

	cfg, err := cfg.LoadConfig("config/default.toml")
	if err != nil {
		logmgr.Fatal("Failed to load config", logmgr.Field("error", err))
	}

	if err := cfg.Validate(); err != nil {
		logmgr.Fatal("Config validation failed", logmgr.Field("error", err))
	}

	population := garden.NewPopulation(cfg.Evolution.PopulationSize, cfg.Evolution.GenomeSize)

	for range cfg.Evolution.Generations {
		population.Step(cfg.Evolution.CrossoverRate, cfg.Evolution.MutationPoints, cfg.Evolution.MutationRate)
	}

	logmgr.Info("Evolution complete", logmgr.Field("population_summary", population.Summary()))
}
