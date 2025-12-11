package metrics

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

// CSVWriter handles writing generation metrics to a CSV file
type CSVWriter struct {
	file         *os.File
	writer       *csv.Writer
	mu           sync.Mutex
	initialized  bool
	isFirstWrite bool
}

// NewCSVWriter creates a new CSV writer for the specified file
func NewCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file %s: %w", filename, err)
	}

	csvWriter := csv.NewWriter(file)

	csvw := &CSVWriter{
		file:         file,
		writer:       csvWriter,
		isFirstWrite: false,
	}

	csvw.initialized = true
	return csvw, nil
}

// WriteMetrics writes a single generation metrics row to the CSV
func (csvw *CSVWriter) WriteMetrics(metrics GenerationMetrics) error {
	if !csvw.initialized {
		return fmt.Errorf("CSV writer not properly initialized")
	}

	csvw.mu.Lock()
	defer csvw.mu.Unlock()

	// --- Write header if first generation ---
	if csvw.isFirstWrite {
		header := []string{"generation", "duration_ns", "population_size", "timestamp"}

		// Add all metric keys dynamically
		for key := range metrics.Metrics {
			header = append(header, key)
		}

		if err := csvw.writer.Write(header); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
		csvw.writer.Flush()
		if err := csvw.writer.Error(); err != nil {
			return fmt.Errorf("failed to flush CSV writer: %w", err)
		}
		csvw.isFirstWrite = false
	}

	// --- Build CSV row ---
	record := []string{
		fmt.Sprintf("%d", metrics.Generation),
		fmt.Sprintf("%d", metrics.Duration.Nanoseconds()),
		fmt.Sprintf("%d", metrics.PopulationSize),
		metrics.Timestamp.Format("2006-01-02T15:04:05.000Z"),
	}

	// Append metric values in same order as header
	for key := range metrics.Metrics {
		record = append(record, fmt.Sprintf("%f", metrics.Metrics[key]))
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
