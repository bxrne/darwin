package main

import (
	"fmt"

	"github.com/bxrne/darwin/internal/cfg"
	"go.uber.org/zap"
)

// InitializeLogger creates and configures a zap logger based on the config's log level
func InitializeLogger(config *cfg.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var buildErr error

	switch config.Logging.Level {
	case "debug":
		logger, buildErr = zap.NewDevelopment()
	case "info":
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logger, buildErr = zapConfig.Build()
	case "warn":
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		logger, buildErr = zapConfig.Build()
	case "error":
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		logger, buildErr = zapConfig.Build()
	default:
		// Default to info if invalid level
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logger, buildErr = zapConfig.Build()
	}

	if buildErr != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", buildErr)
	}

	return logger, nil
}
