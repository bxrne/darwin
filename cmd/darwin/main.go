package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/metrics"
	"go.uber.org/zap"
)

func main() {
	// Initialize zap logger - use development config for better readability
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer func() {
		_ = logger.Sync() // Ignore sync errors on exit
	}()

	// Set global logger so zap.L() works throughout the codebase
	zap.ReplaceGlobals(logger)

	// Use sugared logger for convenience
	sugar := logger.Sugar()

	configPath := flag.String("config", "config/default.toml", "Path to config file")
	csvOutput := flag.String("csv-output", "", "Path to CSV file for metrics output")
	flag.Parse()

	cfg, err := cfg.LoadConfig(*configPath)
	if err != nil {
		sugar.Fatalw("Failed to load config", "error", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sugar.Infow("Starting evolution", "config", cfg.Evolution)

	// Create composite metrics handler
	var handler MetricsHandler

	// Always add logging handler
	// Use logger directly (not sugar) for better performance and ordering
	logHandler := func(m metrics.GenerationMetrics) {
		logInterval := cfg.Evolution.Generations / 10
		if logInterval == 0 {
			logInterval = 1
		}
		if m.Generation%logInterval == 0 || m.Generation == 1 || m.Generation == cfg.Evolution.Generations {
			// Use structured logging with proper field types for better ordering
			logger.Info("Generation metrics",
				zap.Int("gen", m.Generation),
				zap.Int64("ns", m.Duration.Nanoseconds()),
				zap.String("best_fit", fmt.Sprintf("%.3f", m.BestFitness)),
				zap.String("avg_fit", fmt.Sprintf("%.3f", m.AvgFitness)),
				zap.String("min_fit", fmt.Sprintf("%.3f", m.MinFitness)),
				zap.String("max_fit", fmt.Sprintf("%.3f", m.MaxFitness)),
				zap.Int("pop_size", m.PopulationSize),
				zap.String("best_desc", m.BestDescription),
				zap.Int("min_depth", m.MinDepth),
				zap.Int("max_depth", m.MaxDepth),
				zap.String("avg_depth", fmt.Sprintf("%.2f", m.AvgDepth)),
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
			sugar.Fatalw("Failed to create CSV handler", "error", err)
		}

		// Combine both handlers
		handler = func(m metrics.GenerationMetrics) {
			logHandler(m)
			csvHandler(m)
		}

		sugar.Infow("CSV output enabled", "file", csvFile)
	} else {
		handler = logHandler
	}

	finalPop, metricsComplete, err := RunEvolution(ctx, cfg, handler, logger)
	if err != nil {
		sugar.Fatalw("Evolution failed", "error", err.Error())
	}

	// Wait for metrics to finish processing before calculating final stats
	<-metricsComplete

	if len(finalPop) > 0 {
		bestFitness := finalPop[0].GetFitness()
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
		bestIndividual := finalPop[0]
		for _, ind := range finalPop {
			if ind.GetFitness() == bestFitness {
				bestIndividual = ind
				break
			}
		}

		sugar.Infow("Evolution complete",
			"population_size", len(finalPop),
			"best_fitness", fmt.Sprintf("%.3f", bestFitness),
			"avg_fitness", fmt.Sprintf("%.3f", avgFitness),
			"min_fitness", fmt.Sprintf("%.3f", minFitness),
			"best_individual", bestIndividual.Describe(),
		)
	}

	sugar.Info("Evolution finished successfully")
}
