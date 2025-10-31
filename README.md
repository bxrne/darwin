# darwin 

### Clone and Build

```bash
git clone https://github.com/bxrne/darwin.git
cd darwin
go mod tidy
go build ./cmd/darwin
```

## Usage

### Basic Run

```bash
./darwin
```

### Configuration

Create a custom config file:

```toml
[evolution]
population_size = 500
genome_size = 200
crossover_point_count = 1
mutation_rate = 0.05
mutation_points = [6]
generations = 50
elitism_percentage = 0.1
seed = 42
```



| Parameter | Description | Default |
|-----------|-------------|---------|
| `population_size` | Number of individuals in population | 500 |
| `genome_size` | Size of each individual's genome | 200 |
| `crossover_point_count` | Number of crossover points | 1 |
| `mutation_rate` | Probability of mutation (0.0-1.0) | 0.05 |
| `mutation_points` | Genome indices to mutate | [6] |
| `generations` | Number of evolution generations | 50 |
| `elitism_percentage` | Percentage of best individuals preserved | 0.1 |
| `seed` | Random seed for reproducibility | 42 |


### Predefined Configurations

The project includes several predefined configurations for different use cases:

- `config/small.toml`: Quick testing (100 pop, 10 gen)
- `config/medium.toml`: Balanced performance (500 pop, 50 gen)
- `config/large.toml`: Comprehensive evolution (1000 pop, 100 gen)
- `config/default.toml`: Standard configuration

## Benchmarking

Darwin includes comprehensive benchmarking capabilities for performance analysis.

### Running Benchmarks

```bash
# Run all evolution benchmarks
go test -bench=BenchmarkEvolution ./cmd/darwin -benchmem

# Run specific benchmark sizes
go test -bench=BenchmarkEvolution_Small ./cmd/darwin -benchmem
go test -bench=BenchmarkEvolution_Medium ./cmd/darwin -benchmem
go test -bench=BenchmarkEvolution_Large ./cmd/darwin -benchmem
```

### Benchmark Results

Example output:

```
BenchmarkEvolution_Small-16    477    2484647 ns/op    1086608 B/op    5698 allocs/op
Config: Population=100, GenomeSize=50, Generations=10, Seed=42
Run 1: Best=0.900, Avg=0.837
Memory: Used=1085744 bytes, TotalAlloc=1085744 bytes
```

### Performance Profiling

```bash
# CPU profiling
go test -bench=BenchmarkEvolution ./cmd/darwin -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -bench=BenchmarkEvolution ./cmd/darwin -memprofile=mem.prof
go tool pprof mem.prof
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Test Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

