# darwin

Darwin is a flexible evolutionary computation framework supporting both Genetic Algorithms (GA) and Genetic Programming (GP). It features an extensible architecture with the Evolvable interface, channel-based evolution engine, and async metrics streaming. 

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
crossover_point_count = 1
crossover_rate = 0.9
mutation_rate = 0.05
generations = 50
elitism_percentage = 0.1
seed = 42

[bitstring_individual]
enabled = true
genome_size = 200

[tree_individual]
enabled = false
max_depth = 1
min_depth = 0
function_set = ["add"]
terminal_set = ["x"]
```



| Parameter | Description | Default |
|-----------|-------------|---------|
| `population_size` | Number of individuals in population | 500 |
| `crossover_point_count` | Number of crossover points | 1 |
| `crossover_rate` | Probability of crossover (0.0-1.0) | 0.9 |
| `mutation_rate` | Probability of mutation (0.0-1.0) | 0.05 |
| `generations` | Number of evolution generations | 50 |
| `elitism_percentage` | Percentage of best individuals preserved | 0.1 |
| `seed` | Random seed for reproducibility | 42 |

#### Individual Types

Darwin supports different individual representations:

**Bitstring Individuals** (`[bitstring_individual]`)
- `enabled`: Enable bitstring genome evolution
- `genome_size`: Length of binary genome

**Tree Individuals** (`[tree_individual]`)
- `enabled`: Enable tree-based genetic programming
- `max_depth`: Maximum tree depth
- `min_depth`: Minimum tree depth
- `function_set`: Available functions (e.g., ["add", "subtract", "multiply", "divide"])
- `terminal_set`: Terminal values/variables

### Predefined Configurations

The project includes several predefined configurations for different use cases:

- `config/small.toml`: Quick testing with bitstring individuals (100 pop, 10 gen)
- `config/medium.toml`: Balanced performance with bitstring individuals (500 pop, 50 gen)
- `config/large.toml`: Comprehensive evolution with bitstring individuals (2000 pop, 200 gen)
- `config/default.toml`: Genetic programming with tree individuals

## Features

### Genetic Programming Support

Darwin includes support for Genetic Programming (GP) with tree-based individuals. Configure `[tree_individual]` section to enable GP for problems like symbolic regression:

```toml
[tree_individual]
enabled = true
max_depth = 3
function_set = ["add", "subtract", "multiply", "divide"]
terminal_set = ["x", "y", "1.0", "2.0"]
```

### Selection Methods

- **Roulette Selection**: Fitness-proportional selection (default)
- **Tournament Selection**: Tournament-based selection available

### Extensible Architecture

Implement the `Evolvable` interface to create custom individual types:

```go
type Evolvable interface {
    CalculateFitness()
    Mutate(rate float64)
    GetFitness() float64
    Max(i2 Evolvable) Evolvable
    MultiPointCrossover(i2 Evolvable, crossoverPointCount int) (Evolvable, Evolvable)
}
```

### Async Metrics Streaming

Evolution runs with channel-based communication and provides real-time metrics streaming for monitoring progress.

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
Config: Population=100, GenomeSize=64, Generations=10, Seed=42
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

## Architecture

Darwin uses a channel-based evolution engine for concurrent processing and thread-safe random number generation. The async metrics streaming allows real-time monitoring of evolution progress.

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

