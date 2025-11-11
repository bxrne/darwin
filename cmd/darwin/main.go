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
	csvOutput := flag.String("csv-output", "", "Path to CSV file for metrics output")
	flag.Parse()

	cfg, err := cfg.LoadConfig(*configPath)
	if err != nil {
		logmgr.Fatal("Failed to load config", logmgr.Field("error", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logmgr.Info("Starting evolution", logmgr.Field("config", cfg.Evolution))

	// Create composite metrics handler
	var handler MetricsHandler

	// Always add logging handler
	logHandler := func(m metrics.GenerationMetrics) {
		if m.Generation%10 == 0 || m.Generation == cfg.Evolution.Generations {
			logmgr.Debug("",
				logmgr.Field("gen", m.Generation),
				logmgr.Field("ns", m.Duration.Nanoseconds()),
				logmgr.Field("best_fit", fmt.Sprintf("%.3f", m.BestFitness)),
				logmgr.Field("avg_fit", fmt.Sprintf("%.3f", m.AvgFitness)),
				logmgr.Field("min_fit", fmt.Sprintf("%.3f", m.MinFitness)),
				logmgr.Field("max_fit", fmt.Sprintf("%.3f", m.MaxFitness)),
				logmgr.Field("pop_size", m.PopulationSize),
				logmgr.Field("best_desc", m.BestDescription),
				logmgr.Field("min_depth", m.MinDepth),
				logmgr.Field("max_depth", m.MaxDepth),
				logmgr.Field("avg_depth", fmt.Sprintf("%.2f", m.AvgDepth)),
			)
		}
	}

	// Determine CSV output file (flag takes precedence over config)
	csvFile := *csvOutput
	if csvFile == "" && cfg.Metrics.CSVEnabled {
		csvFile = cfg.Metrics.CSVFile
	}

	// Add CSV handler if CSV output is enabled
	if csvFile != "" {
		csvHandler, err := metrics.CreateCSVHandler(csvFile)
		if err != nil {
			logmgr.Fatal("Failed to create CSV handler", logmgr.Field("error", err))
		}

		// Combine both handlers
		handler = func(m metrics.GenerationMetrics) {
			logHandler(m)
			csvHandler(m)
		}

		logmgr.Info("CSV output enabled", logmgr.Field("file", csvFile))
	} else {
		handler = logHandler
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
