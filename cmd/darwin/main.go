package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/bxrne/darwin/internal/cfg"
	"github.com/bxrne/darwin/internal/metrics"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config/default.toml", "Path to config file")
	csvOutput := flag.String("csv-output", "", "Path to CSV file for metrics output")
	flag.Parse()

	// Load config first (needed for logger level)
	cfg, err := cfg.LoadConfig(*configPath)
	if err != nil {
		// Can't use logger yet, use fmt for error
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize zap logger based on config
	logger, err := InitializeLogger(cfg)
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
			fields := []zap.Field{
				zap.Int("gen", m.Generation),
				zap.Int64("ns", m.Duration.Nanoseconds()),
				zap.String("best_desc", m.BestDescription),
			}

			// Add all metrics dynamically
			for k, v := range m.Metrics {
				fields = append(fields, zap.Float64(k, v))
			}
			logger.Info("Generation metrics", fields...)
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

	_, metricsComplete, err := RunEvolution(ctx, cfg, handler, logger)
	if err != nil {
		sugar.Fatalw("Evolution failed", "error", err.Error())
	}

	// Wait for metrics to finish processing before calculating final stats
	<-metricsComplete

	sugar.Info("Evolution finished successfully")
}
