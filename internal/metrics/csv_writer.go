package metrics

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"sync"
)

// CSVWriter supports dynamic schema expansion and rewriting
type CSVWriter struct {
	file        *os.File
	mu          sync.Mutex
	initialized bool

	header []string   // dynamic metric keys
	rows   [][]string // stored rows
	path   string     // path to CSV file
}

// NewCSVWriter creates a new dynamic CSV writer
func NewCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file %s: %w", filename, err)
	}

	return &CSVWriter{
		file:        file,
		path:        filename,
		header:      nil,
		rows:        make([][]string, 0),
		initialized: true,
	}, nil
}

// WriteMetrics writes a generation row and updates schema if needed
func (csvw *CSVWriter) WriteMetrics(metrics GenerationMetrics) error {
	if !csvw.initialized {
		return fmt.Errorf("CSV writer not initialized")
	}

	csvw.mu.Lock()
	defer csvw.mu.Unlock()

	// ---- Detect new metric keys ----
	newKeysFound := false
	for key := range metrics.Metrics {
		if !contains(csvw.header, key) {
			csvw.header = append(csvw.header, key)
			newKeysFound = true
		}
	}

	// Always sort for stable ordering
	sort.Strings(csvw.header)

	// ---- If schema expanded, rewrite whole file ----
	if newKeysFound {
		if err := csvw.rewriteCSV(); err != nil {
			return err
		}
	}

	// ---- Build new row ----
	row := []string{
		fmt.Sprintf("%d", metrics.Generation),
		fmt.Sprintf("%d", metrics.Duration.Nanoseconds()),
		fmt.Sprintf("%d", metrics.PopulationSize),
		metrics.Timestamp.Format("2006-01-02T15:04:05.000Z"),
	}

	// Add metric values in header order
	for _, key := range csvw.header {
		if v, ok := metrics.Metrics[key]; ok {
			row = append(row, fmt.Sprintf("%f", v))
		} else {
			row = append(row, "0") // zero-fill missing metrics
		}
	}

	// Save row
	csvw.rows = append(csvw.rows, row)

	// Rewrite CSV every time (cheap for <100k rows)
	return csvw.rewriteCSV()
}

// rewriteCSV rebuilds the file from header and rows
func (csvw *CSVWriter) rewriteCSV() error {
	// Close current file
	csvw.file.Close()

	// Recreate fresh file
	f, err := os.Create(csvw.path)
	if err != nil {
		return fmt.Errorf("failed to recreate CSV file: %w", err)
	}
	csvw.file = f

	writer := csv.NewWriter(csvw.file)

	// Build full header
	fullHeader := append([]string{
		"generation",
		"duration_ns",
		"population_size",
		"timestamp",
	}, csvw.header...)

	// Write header
	if err := writer.Write(fullHeader); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, row := range csvw.rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	return writer.Error()
}

// Close the file
func (csvw *CSVWriter) Close() error {
	csvw.mu.Lock()
	defer csvw.mu.Unlock()

	if !csvw.initialized {
		return nil
	}

	if err := csvw.file.Close(); err != nil {
		return err
	}

	csvw.initialized = false
	return nil
}

// MetricsHandler is a callback that writes metrics
type MetricsHandler func(GenerationMetrics)

// CreateCSVHandler creates a callback using this CSV writer
func CreateCSVHandler(filename string) (MetricsHandler, error) {
	csvWriter, err := NewCSVWriter(filename)
	if err != nil {
		return nil, err
	}

	return func(metrics GenerationMetrics) {
		if err := csvWriter.WriteMetrics(metrics); err != nil {
			fmt.Printf("Warning: failed to write metrics: %v\n", err)
		}
	}, nil
}

// ---- Helper ----

func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

