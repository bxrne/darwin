package metrics

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

// CSVWriter handles writing generation metrics to a CSV file
type CSVWriter struct {
	file        *os.File
	writer      *csv.Writer
	mu          sync.Mutex
	initialized bool
}

// NewCSVWriter creates a new CSV writer for the specified file
func NewCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file %s: %w", filename, err)
	}

	csvWriter := csv.NewWriter(file)

	csvw := &CSVWriter{
		file:   file,
		writer: csvWriter,
	}

	// Write header
	if err := csvw.writeHeader(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	csvw.initialized = true
	return csvw, nil
}

// writeHeader writes the CSV header row
func (csvw *CSVWriter) writeHeader() error {
	header := []string{
		"generation",
		"duration_ns",
		"best_fitness",
		"avg_fitness",
		"min_fitness",
		"max_fitness",
		"best_description",
		"min_depth",
		"max_depth",
		"avg_depth",
		"population_size",
		"timestamp",
	}

	return csvw.writer.Write(header)
}

// WriteMetrics writes a single generation metrics row to the CSV
func (csvw *CSVWriter) WriteMetrics(metrics GenerationMetrics) error {
	if !csvw.initialized {
		return fmt.Errorf("CSV writer not properly initialized")
	}

	csvw.mu.Lock()
	defer csvw.mu.Unlock()

	record := []string{
		fmt.Sprintf("%d", metrics.Generation),
		fmt.Sprintf("%d", metrics.Duration.Nanoseconds()),
		fmt.Sprintf("%.6f", metrics.BestFitness),
		fmt.Sprintf("%.6f", metrics.AvgFitness),
		fmt.Sprintf("%.6f", metrics.MinFitness),
		fmt.Sprintf("%.6f", metrics.MaxFitness),
		metrics.BestDescription,
		fmt.Sprintf("%d", metrics.MinDepth),
		fmt.Sprintf("%d", metrics.MaxDepth),
		fmt.Sprintf("%.2f", metrics.AvgDepth),
		fmt.Sprintf("%d", metrics.PopulationSize),
		metrics.Timestamp.Format("2006-01-02T15:04:05.000Z"),
	}

	if err := csvw.writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}

	// Flush to ensure data is written to file
	csvw.writer.Flush()
	if err := csvw.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return nil
}

// Close closes the CSV file and flushes any remaining data
func (csvw *CSVWriter) Close() error {
	csvw.mu.Lock()
	defer csvw.mu.Unlock()

	if !csvw.initialized {
		return nil
	}

	// Flush any remaining data
	csvw.writer.Flush()
	if err := csvw.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer on close: %w", err)
	}

	// Close the file
	if err := csvw.file.Close(); err != nil {
		return fmt.Errorf("failed to close CSV file: %w", err)
	}

	csvw.initialized = false
	return nil
}

// MetricsHandler represents a function that handles generation metrics
type MetricsHandler func(GenerationMetrics)

// CreateCSVHandler creates a metrics handler function that writes to CSV
func CreateCSVHandler(filename string) (MetricsHandler, error) {
	csvWriter, err := NewCSVWriter(filename)
	if err != nil {
		return nil, err
	}

	return func(metrics GenerationMetrics) {
		if err := csvWriter.WriteMetrics(metrics); err != nil {
			// Log error but don't crash the evolution
			fmt.Printf("Warning: failed to write metrics to CSV: %v\n", err)
		}
	}, nil
}
