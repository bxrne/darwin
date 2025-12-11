package metrics

import "time"

// GenerationMetrics contains metrics for a single generation
type GenerationMetrics struct {
	Generation      int
	Duration        time.Duration
	BestDescription string
	PopulationSize  int
	Metrics         map[string]float64
	Timestamp       time.Time
}
