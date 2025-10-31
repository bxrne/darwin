package metrics

import "time"

// GenerationMetrics contains metrics for a single generation
type GenerationMetrics struct {
	Generation     int
	Duration       time.Duration
	BestFitness    float64
	AvgFitness     float64
	MinFitness     float64
	MaxFitness     float64
	PopulationSize int
	Timestamp      time.Time
}
